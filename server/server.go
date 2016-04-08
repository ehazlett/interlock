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
	"github.com/ehazlett/interlock/utils"
	"github.com/ehazlett/ttlcache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samalba/dockerclient"
)

const (
	heartbeatKey          = "interlock.heartbeat"
	heartbeatContainerKey = "interlock.heartbeat.container"
	heartbeatLabel        = heartbeatKey
)

type Server struct {
	cfg            *config.Config
	client         *dockerclient.DockerClient
	extensions     []ext.Extension
	metrics        *Metrics
	heartbeatCache *ttlcache.TTLCache
	heartbeatImage string
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

	if cfg.HeartbeatInterval == 0 {
		log.Warn("HeartbeatInterval too low; using default of 60s")
		cfg.HeartbeatInterval = 60
	}

	c, err := ttlcache.NewTTLCache(time.Second * time.Duration(cfg.HeartbeatInterval))
	if err != nil {
		return nil, err
	}
	s.heartbeatCache = c

	client, err := s.getDockerClient()
	if err != nil {
		return nil, err
	}

	// check for current image and set as heartbeat
	cID, err := utils.GetContainerID()
	if err != nil {
		return nil, err
	}

	// inspect current container to get image
	currentContainer, err := client.InspectContainer(cID)
	if err != nil {
		return nil, err
	}
	s.heartbeatImage = currentContainer.Config.Image

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
			log.Debugf("event received: status=%s id=%s type=%s action=%s", e.Status, e.ID, e.Type, e.Action)

			if e.ID == "" && e.Type == "" {
				continue
			}

			// check for heartbeat event
			if e.Action == "create" {
				c, err := s.client.InspectContainer(e.ID)
				if err != nil {
					log.Error(err)
					continue
				}

				// TODO: handle proper failure where event is never seen
				if _, exists := c.Config.Labels[heartbeatLabel]; exists {
					if err := s.heartbeatCache.Set(heartbeatKey, "alive"); err != nil {
						log.Errorf("unable to update heartbeat cache: %s", err)
						continue
					}

					if err := s.heartbeatCache.Set(heartbeatContainerKey, e.ID); err != nil {
						log.Errorf("unable to update heartbeat container info in cache: %s", err)
						continue
					}

					go func() {
						// allow time for cluster refresh
						time.Sleep(time.Millisecond * 1000)

						// remove heartbeat container
						if err := s.client.RemoveContainer(e.ID, true, true); err != nil {
							log.Warnf("error removing heartbeat container: %s", err)
						}
					}()

					continue
				}
			}

			// ignore heartbeat container
			heartbeatID := s.heartbeatCache.Get(heartbeatContainerKey)
			if e.ID == heartbeatID {
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

	// heartbeat ticker
	hbt := time.NewTicker(time.Second * time.Duration(s.cfg.HeartbeatInterval))
	go func() {
		for range hbt.C {
			log.Debug("heartbeat: checking system health")
			// run health container and watch for run event from this container and update TTL
			cfg := &dockerclient.ContainerConfig{
				Image: s.heartbeatImage,
				Cmd: []string{
					"sh",
				},
				Labels: map[string]string{
					heartbeatLabel: "alive",
				},
			}

			if _, err := s.client.CreateContainer(cfg, "", nil); err != nil {
				log.Errorf("error creating heartbeat container: %s", err)
			}

			// check for the heartbeat key; if not exists, warn and attempt reconnect
			if v := s.heartbeatCache.Get(heartbeatKey); v == nil {
				log.Warn("heartbeat failed; attempting reconnect")
				restartChan <- true
			}
		}
	}()

	// start event handler
	restartChan <- true

	// set initial heartbeat key
	if err := s.heartbeatCache.Set(heartbeatKey, "alive"); err != nil {
		log.Errorf("unable to update heartbeat cache: %s", err)
	}

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
