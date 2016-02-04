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
	"github.com/ehazlett/interlock/ext/haproxy"
	"github.com/ehazlett/interlock/ext/nginx"
	"github.com/ehazlett/interlock/pkg/ttlcache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samalba/dockerclient"
)

type Server struct {
	cfg        *config.Config
	client     *dockerclient.DockerClient
	extensions []ext.LoadBalancer
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

				for _, lb := range s.extensions {
					if err := lb.Update(); err != nil {
						errChan <- err
						continue
					}

					// trigger reload
					if err := lb.Reload(); err != nil {
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

			switch e.Status {
			case "start":
				log.Debugf("inspecting container: id=%s", e.ID)
				c, err := client.InspectContainer(e.ID)
				if err != nil {
					// ignore inspect errors
					log.Errorf("error: id=%s type=%s: %s", e.ID, e.Status, err)
					//return
					continue
				}

				log.Debugf("checking container labels: id=%s", e.ID)
				// ignore proxy containers
				if _, ok := c.Config.Labels[ext.InterlockExtNameLabel]; ok {
					log.Debugf("ignoring proxy container: id=%s", c.Id)
					//return
					continue
				}

				log.Debugf("checking container ports: id=%s", e.ID)
				// ignore containetrs without exposed ports
				if len(c.Config.ExposedPorts) == 0 {
					log.Debugf("no ports exposed; ignoring: id=%s", e.ID)
					//return
					continue
				}

				log.Debugf("ports found; checking event type for trigger: %q", e.Status)

				image := c.Config.Image
				log.Debugf("container start: id=%s image=%s", e.ID, image)

				s.cache.Set("reload", true)
			case "kill", "die", "stop":
				log.Debugf("container %s: id=%s", e.Status, e.ID)

				// wait for container to stop
				time.Sleep(time.Millisecond * 250)

				s.cache.Set("reload", true)
			default:
				log.Debugf("unhandled event type: id=%s %s", e.ID, e.Status)
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
