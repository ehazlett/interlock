package server

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/events"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samalba/dockerclient"
)

type Server struct {
	cfg       *config.Config
	client    *dockerclient.DockerClient
	eventChan chan (*dockerclient.Event)
}

type eventArgs struct {
	Image string
}

var (
	errChan chan (error)
)

func NewServer(cfg *config.Config) (*Server, error) {
	s := &Server{
		cfg: cfg,
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

	s.eventChan = make(chan *dockerclient.Event)

	// event handler
	h, err := events.NewEventHandler(s.eventChan)
	if err != nil {
		return nil, err
	}

	// monitor events
	client.StartMonitorEvents(h.Handle, errChan)

	go func() {
		for e := range s.eventChan {
			log.Debugf("event: %v", e)

			switch e.Status {
			case "start":
				go func() {
					// get container info for event
					c, err := client.InspectContainer(e.ID)
					if err != nil {
						errChan <- err
						return
					}

					image := c.Config.Image
					log.Debugf("container start: id=%s image=%s", e.ID, image)
				}()
			case "kill", "die", "stop", "destroy":
				go func() {
					log.Debugf("container %s: id=%s", e.Status, e.ID)
				}()
			}
		}
	}()

	return s, nil
}

func (s *Server) Run() error {
	// start prometheus listener
	http.Handle("/metrics", prometheus.Handler())

	if err := http.ListenAndServe(s.cfg.ListenAddr, nil); err != nil {
		return err
	}

	return nil
}
