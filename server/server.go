package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	etypes "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/events"
	"github.com/ehazlett/interlock/ext"
	"github.com/ehazlett/interlock/ext/beacon"
	"github.com/ehazlett/interlock/ext/lb"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
)

const (
	defaultPollInterval = time.Millisecond * 2000
)

type Server struct {
	cfg           *config.Config
	client        *client.Client
	extensions    []ext.Extension
	metrics       *Metrics
	containerHash string
}

var (
	errChan      chan (error)
	eventChan    chan *events.Message
	eventErrChan chan (error)
	handler      *events.EventHandler
	restartChan  chan (bool)
	recoverChan  chan (bool)
)

func NewServer(cfg *config.Config) (*Server, error) {
	s := &Server{
		cfg:           cfg,
		metrics:       NewMetrics(),
		containerHash: "",
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
	eventChan = make(chan *events.Message)

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

			if s.cfg.PollInterval != "" {
				log.Infof("using polling for container updates: interval=%s", s.cfg.PollInterval)
			} else {
				log.Info("using event stream")
				ctx, cancel := context.WithCancel(context.Background())
				evtChan, evtErrChan := client.Events(ctx, types.EventsOptions{})
				defer cancel()

				go func(ch <-chan error) {
					for {
						err := <-ch
						eventErrChan <- err
					}
				}(evtErrChan)

				// since the event stream channel is receive
				// only we wrap it to be able to send
				// interlock events on the interlock chan
				go func(ch <-chan etypes.Message) {
					for {
						msg := <-ch
						m := &events.Message{
							msg,
						}

						eventChan <- m
					}
				}(evtChan)

				// monitor events
				// event handler
				h, err := events.NewEventHandler(eventChan)
				if err != nil {
					errChan <- err
					continue
				}

				handler = h
			}

			//go func(e io.ReadCloser) {
			//	s := bufio.NewScanner(e)
			//	for s.Scan() {
			//		if err != nil {
			//			errChan <- err
			//			continue
			//		}

			//		var msg etypes.Message
			//		if err := json.Unmarshal([]byte(s.Text()), &msg); err != nil {
			//			errChan <- err
			//			continue
			//		}

			//		eventChan <- msg
			//	}
			//}(e)

			// trigger initial load
			eventChan <- &events.Message{
				etypes.Message{
					ID:     "0",
					Status: "interlock-start",
				},
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
		options := types.ContainerListOptions{All: true}
		if _, err := s.client.ContainerList(context.Background(), options); err == nil {
			log.Info("event stream appears to have recovered; restarting handler")
			return
		}

		log.Debug("event stream not yet ready; retrying")

		time.Sleep(time.Second * 1)
	}
}

func (s *Server) loadExtensions(client *client.Client) {
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
			if !s.cfg.EnableMetrics {
				log.Errorf("unable to load beacon: metrics are disabled")
				continue
			}
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

func (s *Server) runPoller(d time.Duration) {
	t := time.NewTicker(d)
	go func() {
		for range t.C {
			opts := types.ContainerListOptions{
				All:  false,
				Size: false,
			}
			containers, err := s.client.ContainerList(context.Background(), opts)
			if err != nil {
				log.Warnf("unable to get containers: %s", err)
				continue
			}

			containerIDs := []string{}
			ports := []int{}

			for _, c := range containers {
				containerIDs = append(containerIDs, c.ID)
				for _, p := range c.Ports {
					ports = append(ports, int(p.PublicPort))
				}
			}

			sort.Strings(containerIDs)
			sort.Ints(ports)

			cData, err := json.Marshal(containerIDs)
			if err != nil {
				log.Errorf("unable to marshal containers: %s", err)
				continue
			}

			pData, err := json.Marshal(ports)
			if err != nil {
				log.Errorf("unable to marshal ports: %s", err)
				continue
			}

			h := sha256.New()
			h.Write(cData)
			h.Write(pData)
			sum := hex.EncodeToString(h.Sum(nil))

			if sum != s.containerHash {
				log.Debug("detected new containers; triggering reload")
				s.containerHash = sum
				// trigger update
				eventChan <- &events.Message{
					etypes.Message{
						ID:     fmt.Sprintf("%d", time.Now().UnixNano()),
						Status: "interlock-restart",
					},
				}
			}
		}
	}()
}

func (s *Server) Run() error {
	if s.cfg.EnableMetrics {
		// start prometheus listener
		http.Handle("/metrics", prometheus.Handler())
	}

	if s.cfg.PollInterval != "" {
		// run background poller
		d, err := time.ParseDuration(s.cfg.PollInterval)
		if err != nil {
			return err
		}

		if d < defaultPollInterval {
			log.Warnf("poll interval too quick; defaulting to %v", defaultPollInterval)
			s.cfg.PollInterval = "2s"
			d = defaultPollInterval
		}

		s.runPoller(d)
	}

	if err := http.ListenAndServe(s.cfg.ListenAddr, nil); err != nil {
		return err
	}

	return nil
}
