package lb

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/ext"
	"github.com/ehazlett/interlock/ext/lb/haproxy"
	"github.com/ehazlett/interlock/ext/lb/nginx"
	"github.com/ehazlett/ttlcache"
	"github.com/samalba/dockerclient"
)

const (
	pluginName      = "lb"
	ReloadThreshold = time.Millisecond * 500
)

var (
	errChan      chan (error)
	restartChan  = make(chan bool)
	lbUpdateChan chan (bool)
)

type LoadBalancerBackend interface {
	Name() string
	ConfigPath() string
	GenerateProxyConfig(c []dockerclient.Container) (interface{}, error)
	Template() string
	Reload(proxyContainers []dockerclient.Container) error
}

type LoadBalancer struct {
	cfg     *config.ExtensionConfig
	client  *dockerclient.DockerClient
	cache   *ttlcache.TTLCache
	lock    *sync.Mutex
	backend LoadBalancerBackend
}

func log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"ext": pluginName,
	})
}

type eventArgs struct {
	Image string
}

func NewLoadBalancer(c *config.ExtensionConfig, client *dockerclient.DockerClient) (*LoadBalancer, error) {
	// parse config base dir
	c.ConfigBasePath = filepath.Dir(c.ConfigPath)

	errChan = make(chan error)
	go func() {
		for err := range errChan {
			log().Error(err)
		}
	}()

	lbUpdateChan = make(chan bool)

	cache, err := ttlcache.NewTTLCache(ReloadThreshold)
	if err != nil {
		return nil, err
	}

	cache.SetCallback(func(k string, v interface{}) {
		log().Debugf("triggering reload from cache")
		lbUpdateChan <- true
	})

	extension := &LoadBalancer{
		cfg:    c,
		client: client,
		cache:  cache,
		lock:   &sync.Mutex{},
	}

	// select backend
	switch c.Name {
	case "haproxy":
		p, err := haproxy.NewHAProxyLoadBalancer(c, client)
		if err != nil {
			return nil, fmt.Errorf("error setting backend: %s", err)
		}
		extension.backend = p
	case "nginx":
		p, err := nginx.NewNginxLoadBalancer(c, client)
		if err != nil {
			return nil, fmt.Errorf("error setting backend: %s", err)
		}
		extension.backend = p
	default:
		return nil, fmt.Errorf("unknown load balancer backend: %s", c.Name)
	}

	// lbUpdateChan handler
	go func() {
		for range lbUpdateChan {
			log().Debug("checking to reload")
			if v := extension.cache.Get("reload"); v != nil {
				log().Debug("skipping reload: too many requests")
				continue
			}

			start := time.Now()

			log().Debug("updating load balancers")

			containers, err := client.ListContainers(false, false, "")
			if err != nil {
				errChan <- err
				continue
			}

			// generate proxy config
			log().Debug("generating proxy config")
			cfg, err := extension.backend.GenerateProxyConfig(containers)
			if err != nil {
				errChan <- err
				continue
			}

			// save proxy config
			configPath := extension.backend.ConfigPath()

			log().Debug("saving proxy config")
			if err := extension.SaveConfig(configPath, cfg); err != nil {
				errChan <- err
				continue
			}

			proxyNetworks := map[string]string{}

			proxyContainers, err := extension.ProxyContainers(extension.backend.Name())
			if err != nil {
				errChan <- err
				continue
			}

			// connect to networks
			switch extension.backend.Name() {
			case "nginx":
				proxyConfig := cfg.(*nginx.Config)
				proxyNetworks = proxyConfig.Networks

			case "haproxy":
				proxyConfig := cfg.(*nginx.Config)
				proxyNetworks = proxyConfig.Networks
			default:
				errChan <- fmt.Errorf("unable to connect to networks; unknown backend: %s", extension.backend.Name())
				continue
			}

			for _, cnt := range proxyContainers {
				for net, _ := range proxyNetworks {
					if _, ok := cnt.NetworkSettings.Networks[net]; !ok {
						log().Debugf("connecting proxy container %s to network %s", cnt.Id, net)

						// connect
						if err := client.ConnectNetwork(net, cnt.Id); err != nil {
							log().Warnf("unable to connect container %s to network %s: %s", cnt.Id, net, err)
							continue
						}
					}
				}
			}

			// trigger reload
			log().Debug("reloading")
			if err := extension.backend.Reload(proxyContainers); err != nil {
				errChan <- err
				continue
			}

			d := time.Since(start)
			duration := float64(d.Seconds() * float64(1000))

			log().Debugf("reload duration: %0.2fms", duration)
		}
	}()

	return extension, nil
}

func (l *LoadBalancer) Name() string {
	return pluginName
}

func (l *LoadBalancer) ProxyContainers(name string) ([]dockerclient.Container, error) {
	// TODO:
	containers, err := l.client.ListContainers(false, false, "")
	if err != nil {
		return nil, err
	}

	proxyContainers := []dockerclient.Container{}

	// find interlock proxy containers
	for _, cnt := range containers {
		if v, ok := cnt.Labels[ext.InterlockExtNameLabel]; ok && v == l.backend.Name() {
			proxyContainers = append(proxyContainers, cnt)
		}
	}

	return proxyContainers, nil
}

func (l *LoadBalancer) SaveConfig(configPath string, cfg interface{}) error {
	f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		ff, fErr := os.Create(configPath)
		defer ff.Close()
		if fErr != nil {
			return fErr
		}
		f = ff
	}
	defer f.Close()

	t := template.New("lb")
	confTmpl := l.backend.Template()

	var tErr error
	var c bytes.Buffer

	tmpl, err := t.Parse(confTmpl)
	if err != nil {
		return err
	}

	// cast to config type
	switch l.backend.Name() {
	case "nginx":
		config := cfg.(*nginx.Config)
		if err := tmpl.Execute(&c, config); err != nil {
			return tErr
		}
	case "haproxy":
		config := cfg.(*haproxy.Config)
		if err := tmpl.Execute(&c, config); err != nil {
			return tErr
		}
	default:
		return fmt.Errorf("unknown backend type: %s", l.backend.Name())
	}

	if _, err := f.Write(c.Bytes()); err != nil {
		return err
	}

	f.Sync()

	return nil
}

func (l *LoadBalancer) HandleEvent(event *dockerclient.Event) error {
	reload := false

	// container event
	switch event.Status {
	case "start":
		reload = l.isExposedContainer(event.ID)
	case "stop":
		reload = l.isExposedContainer(event.ID)

		// wait for container to stop
		time.Sleep(time.Millisecond * 250)
	case "interlock-start", "destroy":
		// force reload
		reload = true
	}

	// network event
	switch event.Action {
	case "connect", "disconnect":
		// since event.ID is blank on an action we must get the proper ID
		id, ok := event.Actor.Attributes["container"]
		if !ok {
			return fmt.Errorf("unable to detect container id for network event")
		}

		reload = l.isExposedContainer(id)
	}

	if reload {
		log().Debug("triggering reload")
		l.cache.Set("reload", true)
	}

	return nil
}

// TODO: update for overlay?
func (l *LoadBalancer) isExposedContainer(id string) bool {
	log().Debugf("inspecting container: id=%s", id)
	c, err := l.client.InspectContainer(id)
	if err != nil {
		// ignore inspect errors
		log().Errorf("error: id=%s err=%s", id, err)
		return false
	}

	log().Debugf("checking container labels: id=%s", id)
	// ignore proxy containers
	if _, ok := c.Config.Labels[ext.InterlockExtNameLabel]; ok {
		log().Debugf("ignoring proxy container: id=%s", id)
		return false
	}

	log().Debugf("checking container ports: id=%s", id)
	// ignore containers without exposed ports
	if len(c.Config.ExposedPorts) == 0 {
		log().Debugf("no ports exposed; ignoring: id=%s", id)
		return false
	}

	log().Debugf("container is monitored; triggering reload: id=%s", id)
	return true
}
