package server

import (
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/events"
	"github.com/ehazlett/interlock/ext"
	"github.com/ehazlett/interlock/ext/beacon"
	"github.com/ehazlett/interlock/ext/lb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samalba/dockerclient"
)

type Server struct {
	cfg        *config.Config
	client     *dockerclient.DockerClient
	extensions []ext.Extension
	metrics    *Metrics
}

var (
	errChan      chan (error)
	eventChan    (chan *dockerclient.Event)
	eventErrChan chan (error)
	handler      *events.EventHandler
	restartChan  chan (bool)
	recoverChan  chan (bool)
)

func NewServer(cfg *config.Config) (*Server, error) {
	s := &Server{
		cfg:     cfg,
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
			eventChan <- &dockerclient.Event{
				ID:     "0",
				Status: "interlock-start",
			}
		}
	}()

	// load extensions
	s.loadExtensions(client)

	go func() {
		for e := range eventChan {
			log.Debugf("event received: type=%s id=%s", e.Status, e.ID)

			if e.ID == "" {
				continue
			}

			// send the raw event for extension handling
			for _, ext := range s.extensions {
				log.Debugf("notifying extension: %s", ext.Name())
				if err := ext.HandleEvent(e); err != nil {
					errChan <- err
					continue
				}
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
		log.Debugf("loading extension: name=%s", x.Name)
		switch strings.ToLower(x.Name) {
		case "haproxy", "nginx":
			p, err := lb.NewLoadBalancer(x, client)
			if err != nil {
				log.Errorf("error loading load balancer extension: %s", err)
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
