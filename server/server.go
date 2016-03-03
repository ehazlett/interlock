package server

import (
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/events"
	"github.com/ehazlett/interlock/ext"
	"github.com/ehazlett/interlock/ext/beacon"
	"github.com/ehazlett/interlock/ext/haproxy"
	"github.com/ehazlett/interlock/ext/nginx"
	"github.com/ehazlett/ttlcache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samalba/dockerclient"
)

type Server struct {
	cfg        *config.Config
	client     *dockerclient.DockerClient
	extensions []ext.Extension
	lock       *sync.Mutex
	cache      *ttlcache.TTLCache
	metrics    *Metrics
}

const (
	ReloadThreshold = time.Millisecond * 500
)

var (
	errChan      chan (error)
	eventChan    (chan *dockerclient.Event)
	eventErrChan chan (error)
	handler      *events.EventHandler
	restartChan  chan (bool)
	recoverChan  chan (bool)
	lbUpdateChan chan (bool)
)

func NewServer(cfg *config.Config) (*Server, error) {
	cache, err := ttlcache.NewTTLCache(ReloadThreshold)
	if err != nil {
		return nil, err
	}

	cache.SetCallback(func(k string, v interface{}) {
		log.Debugf("triggering reload from cache")
		lbUpdateChan <- true
	})

	s := &Server{
		cfg:     cfg,
		lock:    &sync.Mutex{},
		cache:   cache,
		metrics: NewMetrics(),
	}

	client, err := s.getDockerClient()
	if err != nil {
		return nil, err
	}

	// channel setup
	errChan = make(chan error)
	eventErrChan = make(chan error)
	restartChan = make(chan bool)
	recoverChan = make(chan bool)
	eventChan = make(chan *dockerclient.Event)
	lbUpdateChan = make(chan bool)

	s.client = client

	// eventErrChan handler
	// this handles event stream errors
	go func() {
		for range eventErrChan {
			// error from swarm event stream; attempt to restart
			log.Error("event stream fail; attempting to reconnect")

			s.waitForSwarm()

			restartChan <- true
		}
	}()

	// errChan handler
	// this is a general error handling channel
	go func() {
		for err := range errChan {
			log.Error(err)
			// HACK: check for errors from swarm and restart
			// events.  an example is "No primary manager elected"
			// before the event handler is created and thus
			// won't send the error there
			if strings.Index(err.Error(), "500 Internal Server Error") > -1 {
				log.Debug("swarm error detected")

				s.waitForSwarm()

				restartChan <- true
			}
		}
	}()

	// restartChan handler
	go func() {
		for range restartChan {
			log.Debug("starting event handling")

			// monitor events
			// event handler
			h, err := events.NewEventHandler(eventChan)
			if err != nil {
				errChan <- err
				return
			}

			handler = h

			s.client.StartMonitorEvents(handler.Handle, eventErrChan)

			// trigger initial load
			lbUpdateChan <- true
		}
	}()

	// load extensions
	s.loadExtensions(client)

	// lbUpdateChan handler
	go func() {
		for range lbUpdateChan {
			log.Debug("checking to reload")
			if v := s.cache.Get("reload"); v != nil {
				log.Debug("skipping reload: too many requests")
				continue
			}

			log.Debug("reloading")
			go func() {
				start := time.Now()

				log.Debug("updating load balancers")
				s.lock.Lock()
				defer s.lock.Unlock()

				for _, ext := range s.extensions {
					if err := ext.Update(); err != nil {
						errChan <- err
						continue
					}

					// trigger reload
					if err := ext.Reload(); err != nil {
						errChan <- err
						continue
					}
				}

				d := time.Since(start)
				duration := float64(d.Seconds() * float64(1000))

				s.metrics.LastReloadDuration.Set(duration)

				log.Debugf("reload duration: %0.2fms", duration)

			}()
		}
	}()

	go func() {
		for e := range eventChan {
			log.Debugf("event received: type=%s id=%s", e.Status, e.ID)

			if e.ID == "" {
				continue
			}

			// send the raw event for extension handling
			for _, ext := range s.extensions {
				if err := ext.HandleEvent(e); err != nil {
					errChan <- err
					continue
				}
			}

			reload := false

			switch e.Status {
			case "start":
				reload = s.isExposedContainer(e.ID)
			case "stop":
				reload = s.isExposedContainer(e.ID)

				// wait for container to stop
				time.Sleep(time.Millisecond * 250)
			case "destroy":
				// force reload to handle container removal
				reload = true
			}

			if reload {
				log.Debug("triggering reload")
				s.cache.Set("reload", true)
			}

			// counter
			s.metrics.EventsProcessed.Inc()
		}
	}()

	// uptime ticker
	t := time.NewTicker(time.Second * 1)
	go func() {
		for range t.C {
			s.metrics.Uptime.Inc()
		}
	}()

	// start event handler
	restartChan <- true

	return s, nil
}

func (s *Server) isExposedContainer(id string) bool {
	log.Debugf("inspecting container: id=%s", id)
	c, err := s.client.InspectContainer(id)
	if err != nil {
		// ignore inspect errors
		log.Errorf("error: id=%s err=%s", id, err)
		return false
	}

	log.Debugf("checking container labels: id=%s", id)
	// ignore proxy containers
	if _, ok := c.Config.Labels[ext.InterlockExtNameLabel]; ok {
		log.Debugf("ignoring proxy container: id=%s", id)
		return false
	}

	log.Debugf("checking container ports: id=%s", id)
	// ignore containetrs without exposed ports
	if len(c.Config.ExposedPorts) == 0 {
		log.Debugf("no ports exposed; ignoring: id=%s", id)
		return false
	}

	log.Debugf("container is monitored; triggering reload: id=%s", id)
	return true
}

func (s *Server) waitForSwarm() {
	log.Info("waiting for event stream to become ready")

	for {

		if _, err := s.client.ListContainers(false, false, ""); err == nil {
			log.Info("event stream appears to have recovered; restarting handler")
			return
		}

		log.Debug("event stream not yet ready; retrying")

		time.Sleep(time.Second * 1)
	}
}

func (s *Server) loadExtensions(client *dockerclient.DockerClient) {
	for _, x := range s.cfg.Extensions {
		log.Debugf("loading extension: name=%s configpath=%s", x.Name, x.ConfigPath)
		switch strings.ToLower(x.Name) {
		case "haproxy":
			p, err := haproxy.NewHAProxyLoadBalancer(x, client)
			if err != nil {
				log.Errorf("error loading haproxy extension: %s", err)
				continue
			}
			s.extensions = append(s.extensions, p)
		case "nginx":
			p, err := nginx.NewNginxLoadBalancer(x, client)
			if err != nil {
				log.Errorf("error loading nginx extension: %s", err)
				continue
			}
			s.extensions = append(s.extensions, p)
		case "beacon":
			p, err := beacon.NewBeacon(x, client)
			if err != nil {
				log.Errorf("error loading beacon extension: %s", err)
				continue
			}
			s.extensions = append(s.extensions, p)
		default:
			log.Errorf("unsupported extension: name=%s", x.Name)
		}
	}
}

func (s *Server) Run() error {
	if s.cfg.EnableMetrics {
		// start prometheus listener
		http.Handle("/metrics", prometheus.Handler())
	}

	if err := http.ListenAndServe(s.cfg.ListenAddr, nil); err != nil {
		return err
	}

	return nil
}
