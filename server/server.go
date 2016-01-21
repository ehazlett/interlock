package server

import (
	"net/http"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/events"
	"github.com/ehazlett/interlock/ext"
	"github.com/ehazlett/interlock/ext/haproxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samalba/dockerclient"
)

type Server struct {
	cfg        *config.Config
	client     *dockerclient.DockerClient
	eventChan  chan (*dockerclient.Event)
	extensions []ext.LoadBalancer
	lock       *sync.Mutex
}

var (
	errChan      chan (error)
	lbUpdateChan chan (bool)
)

func NewServer(cfg *config.Config) (*Server, error) {
	s := &Server{
		cfg:  cfg,
		lock: &sync.Mutex{},
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
				log.Debugf("evt: %s", e)

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
					// ignore containers without exposed ports
					image := c.Config.Image
					log.Debugf("container start: id=%s image=%s", e.ID, image)
					lbUpdateChan <- true
				case "kill", "die", "stop":
					log.Debugf("container %s: id=%s", e.Status, e.ID)

					lbUpdateChan <- true
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
