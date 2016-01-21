package server

import (
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/diegobernardes/ttlcache"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/events"
	"github.com/ehazlett/interlock/ext"
	"github.com/ehazlett/interlock/ext/haproxy"
	"github.com/ehazlett/interlock/ext/nginx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samalba/dockerclient"
)

type Server struct {
	cfg        *config.Config
	client     *dockerclient.DockerClient
	eventChan  chan (*dockerclient.Event)
	extensions []ext.LoadBalancer
	lock       *sync.Mutex
	cache      *ttlcache.Cache
}

const (
	ReloadThreshold = time.Millisecond * 500
)

var (
	errChan      chan (error)
	lbUpdateChan chan (bool)
)

func NewServer(cfg *config.Config) (*Server, error) {
	reloadCallback := func(key string, value interface{}) {
		lbUpdateChan <- true
	}

	cache := ttlcache.NewCache()
	cache.SetTTL(ReloadThreshold)
	cache.SetExpirationCallback(reloadCallback)

	s := &Server{
		cfg:   cfg,
		lock:  &sync.Mutex{},
		cache: cache,
	}

	client, err := s.getDockerClient()
	if err != nil {
		return nil, err
	}

	errChan = make(chan error)
	go func() {
		err := <-errChan
		log.Error(err)
	}()

	s.client = client

	// load extensions
	s.loadExtensions(client)

	s.eventChan = make(chan *dockerclient.Event)

	// event handler
	h, err := events.NewEventHandler(s.eventChan)
	if err != nil {
		return nil, err
	}

	lbUpdateChan = make(chan bool)
	go func() {
		for range lbUpdateChan {
			if _, exists := s.cache.Get("reload"); exists {
				log.Debugf("skipping reload: too many requests")
				continue
			}

			go func() {
				log.Debugf("updating load balancers")
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

			}()
		}
	}()

	// monitor events
	client.StartMonitorEvents(h.Handle, errChan)

	go func() {
		for e := range s.eventChan {
			go func() {
				c, err := client.InspectContainer(e.ID)
				if err != nil {
					// ignore inspect errors
					return
				}

				// ignore proxy containers
				if _, ok := c.Config.Labels[ext.InterlockExtNameLabel]; ok {
					return
				}

				if len(c.Config.ExposedPorts) == 0 {
					log.Debugf("no ports exposed; ignoring: id=%s", e.ID)
					return
				}

				switch e.Status {
				case "start":
					// ignore containetrs without exposed ports
					image := c.Config.Image
					log.Debugf("container start: id=%s image=%s", e.ID, image)

					s.cache.Set("reload", true)
				case "kill", "die", "stop":
					log.Debugf("container %s: id=%s", e.Status, e.ID)

					// wait for container to stop
					time.Sleep(time.Millisecond * 250)

					s.cache.Set("reload", true)
				}
			}()
		}
	}()

	// trigger initial load
	lbUpdateChan <- true

	return s, nil
}

func (s *Server) loadExtensions(client *dockerclient.DockerClient) {
	for _, x := range s.cfg.Extensions {
		log.Debugf("loading extension: name=%s configpath=%s", x.Name, x.ConfigPath)
		switch strings.ToLower(x.Name) {
		case "haproxy":
			p, err := haproxy.NewHAProxyLoadBalancer(&x, client)
			if err != nil {
				log.Errorf("error loading haproxy extension: %s", err)
				continue
			}
			s.extensions = append(s.extensions, p)
		case "nginx":
			p, err := nginx.NewNginxLoadBalancer(&x, client)
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
	// start prometheus listener
	http.Handle("/metrics", prometheus.Handler())

	if err := http.ListenAndServe(s.cfg.ListenAddr, nil); err != nil {
		return err
	}

	return nil
}
