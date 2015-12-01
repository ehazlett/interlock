package manager

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

// Manager listens on events from the connected Docker client and dispatches them
// to registered plugins
type (
	Manager struct {
		Config    *interlock.Config
		Client    *dockerclient.DockerClient
		mux       sync.Mutex
		tlsConfig *tls.Config
		proxyCmd  *exec.Cmd
	}
)

// NewManager create a new Manager
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

// Run starts up the manager, loads plugins, and dispatches a Docker event with
// status "interlock-start"
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
	allPlugins := plugins.GetPlugins()
	if len(allPlugins) == 0 || len(m.Config.EnabledPlugins) == 0 {
		log.Warnf("no plugins enabled")
	}

	enabledPlugins := make(map[string]*plugins.RegisteredPlugin)

	for _, v := range m.Config.EnabledPlugins {
		if p, ok := allPlugins[v]; ok {
			log.Infof("loading plugin name=%s version=%s",
				p.Info().Name,
				p.Info().Version)
			enabledPlugins[v] = p
		}
	}

	plugins.SetEnabledPlugins(enabledPlugins)

	// custom event to signal startup
	evt := &dockerclient.Event{
		Id:     "",
		Status: "interlock-start",
		From:   "interlock",
		Time:   time.Now().UnixNano(),
	}
	plugins.DispatchEvent(m.Config, m.Client, evt, eventsErrChan)
	return nil
}

// Stop emits a Docker event with status "interlock-stop"
func (m *Manager) Stop() error {
	// custom event to signal shutdown
	evt := &dockerclient.Event{
		Id:     "",
		Status: "interlock-stop",
		From:   "interlock",
		Time:   time.Now().UnixNano(),
	}
	plugins.DispatchEvent(m.Config, m.Client, evt, eventsErrChan)
	return nil
}
