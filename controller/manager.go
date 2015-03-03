package main

import (
	"crypto/tls"
	"net"
	"net/url"
	"os/exec"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock"
	"github.com/ehazlett/interlock/plugins"
	"github.com/samalba/dockerclient"
)

var (
	eventsErrChan = make(chan error)
)

type (
	Manager struct {
		Config    *interlock.Config
		Client    *dockerclient.DockerClient
		mux       sync.Mutex
		tlsConfig *tls.Config
		proxyCmd  *exec.Cmd
	}
)

func NewManager(cfg *interlock.Config, tlsConfig *tls.Config) *Manager {
	m := &Manager{
		Config:    cfg,
		tlsConfig: tlsConfig,
	}

	return m
}

func (m *Manager) connect() error {
	log.Debugf("connecting to swarm on %s", m.Config.SwarmUrl)
	c, err := dockerclient.NewDockerClient(m.Config.SwarmUrl, m.tlsConfig)
	if err != nil {
		log.Warn(err)
		return err
	}
	m.Client = c
	go m.startEventListener()
	go m.reconnectOnFail()
	return nil
}

func (m *Manager) startEventListener() {
	evt := NewEventHandler(m)
	m.Client.StartMonitorEvents(evt.Handle, eventsErrChan)
}

func waitForTCP(addr string) error {
	log.Debugf("waiting for swarm to become available on %s", addr)
	for {
		conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
		if err != nil {
			continue
		}
		conn.Close()
		break
	}
	return nil
}

func (m *Manager) reconnectOnFail() {
	<-eventsErrChan
	for {
		log.Warnf("error receiving events; attempting to reconnect")
		u, err := url.Parse(m.Config.SwarmUrl)
		if err != nil {
			log.Warnf("unable to parse Swarm URL: %s", err)
			continue
		}

		if err := waitForTCP(u.Host); err != nil {
			log.Warnf("error connecting to Swarm: %s", err)
			continue
		}

		if err := m.connect(); err == nil {
			log.Debugf("re-connected to Swarm: %s", u.Host)
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func (m *Manager) Run() error {
	if err := m.connect(); err != nil {
		return err
	}

	go func() {
		for {
			err := <-eventsErrChan
			log.Error(err)
		}
	}()

	// plugins
	for _, p := range plugins.GetPlugins() {
		log.Infof("loaded plugin name=%s version=%s",
			p.Info().Name,
			p.Info().Version)
	}

	return nil
}
