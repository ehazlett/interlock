package lb

import (
	"fmt"
	"path/filepath"
	"sync"
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
	Reload() error
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

			log().Debug("reloading")
			start := time.Now()

			log().Debug("updating load balancers")
			//extension.lock.Lock()
			//defer extension.lock.Unlock()

			// trigger reload
			log().Debug("reloading")
			if err := extension.backend.Reload(); err != nil {
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

func (l *LoadBalancer) HandleEvent(event *dockerclient.Event) error {
	reload := false

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

	if reload {
		log().Debug("triggering reload")
		l.cache.Set("reload", true)
	}

	return nil
}

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
	// ignore containetrs without exposed ports
	if len(c.Config.ExposedPorts) == 0 {
		log().Debugf("no ports exposed; ignoring: id=%s", id)
		return false
	}

	log().Debugf("container is monitored; triggering reload: id=%s", id)
	return true
}
