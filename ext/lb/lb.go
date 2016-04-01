package lb

import (
	"archive/tar"
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
	"encoding/json"
	"net/url"

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
	nodeID  string
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

	// load nodeID
	nodeID, err := getNodeID()
	if err != nil {
		return nil, err
	}

	log().Infof("interlock node: id=%s", nodeID)

	extension := &LoadBalancer{
		cfg:    c,
		client: client,
		cache:  cache,
		lock:   &sync.Mutex{},
		nodeID: nodeID,
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

			containers, err := extension.ListContainers()
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
			log().Debugf("proxy config path: %s", configPath)

			proxyNetworks := map[string]string{}

			proxyContainers, err := extension.ProxyContainers(extension.backend.Name())
			if err != nil {
				errChan <- err
				continue
			}

			// save config
			log().Debug("saving proxy config")
			if err := extension.SaveConfig(configPath, cfg, proxyContainers); err != nil {
				errChan <- err
				continue
			}

			// connect to networks
			switch extension.backend.Name() {
			case "nginx":
				proxyConfig := cfg.(*nginx.Config)
				proxyNetworks = proxyConfig.Networks

			case "haproxy":
				proxyConfig := cfg.(*haproxy.Config)
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

			// get interlock nodes
			interlockNodes := []dockerclient.Container{}

			for _, cnt := range containers {
				// always include self container
				if cnt.Id == nodeID {
					interlockNodes = append(interlockNodes, cnt)
					continue
				}

				cInfo, err := client.InspectContainer(cnt.Id)
				if err != nil {
					log().Errorf("unable to inspect interlock container: %s", err)
					continue
				}

				if strings.Index(cInfo.Config.Image, "interlock") > 0 {
					if _, ok := cInfo.Config.Labels[ext.InterlockAppLabel]; ok {
						interlockNodes = append(interlockNodes, cnt)
					}
				}
			}

			proxyContainersToRestart := extension.proxyContainersToRestart(interlockNodes, proxyContainers)

			// trigger reload
			log().Debug("signaling reload")
			if err := extension.backend.Reload(proxyContainersToRestart); err != nil {
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


func (l *LoadBalancer) ListContainers() ([]dockerclient.Container, error) {
	var containers []dockerclient.Container
	var err error

	if len(l.cfg.ServiceName) > 0 {
		labelFilter := fmt.Sprintf("%s=%s", ext.InterlockExtServiceNameLabel, l.cfg.ServiceName)
		filterMap := map[string][]string{ "label":[]string{labelFilter} }
		filterJSON, _ := json.Marshal(filterMap)
		filterStr := string(filterJSON)
		filter := url.QueryEscape(filterStr)

		containers, err = l.client.ListContainers(false, false, filter)
	} else {
		containers, err = l.client.ListContainers(false, false, "")
	}

	if err != nil {
		return nil, err
	}

	return containers, nil
}

func (l *LoadBalancer) ProxyContainers(name string) ([]dockerclient.Container, error) {
	containers, err := l.ListContainers()

	if err != nil {
		return nil, err
	}

	proxyContainers := []dockerclient.Container{}

	log().Debugf("backend name: %s", l.backend.Name())

	// find interlock proxy containers
	for _, cnt := range containers {
		log().Debugf("container: names=%s labels=%s", cnt.Names, cnt.Labels[ext.InterlockExtNameLabel])
		if v, ok := cnt.Labels[ext.InterlockExtNameLabel]; ok && v == l.backend.Name() {
			proxyContainers = append(proxyContainers, cnt)
		}
	}

	log().Debugf("proxy containers: %s", proxyContainers)

	return proxyContainers, nil
}

func (l *LoadBalancer) SaveConfig(configPath string, cfg interface{}, proxyContainers []dockerclient.Container) error {
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

	fName := path.Base(l.backend.ConfigPath())
	proxyConfigPath := path.Dir(l.backend.ConfigPath())

	data := c.Bytes()

	// copy to proxy nodes
	for _, cnt := range proxyContainers {
		log().Debugf("updating proxy config: id=%s", cnt.Id)
		// create tar stream to copy
		buf := new(bytes.Buffer)
		tw := tar.NewWriter(buf)
		hdr := &tar.Header{
			Name: fName,
			Mode: 0644,
			Size: int64(len(data)),
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("error writing proxy config header: %s", err)
		}

		if _, err := tw.Write(data); err != nil {
			return fmt.Errorf("error writing proxy config: %s", err)
		}

		if err := tw.Close(); err != nil {
			return fmt.Errorf("error closing tar writer: %s", err)
		}

		if err := l.client.CopyToContainer(cnt.Id, proxyConfigPath, buf); err != nil {
			log().Errorf("error copying proxy config: %s", err)
			continue
		}
	}

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

// proxyContainersToRestart returns a slice of proxy containers to restart
// based upon this instance's hash
func (l *LoadBalancer) proxyContainersToRestart(nodes []dockerclient.Container, proxyContainers []dockerclient.Container) []dockerclient.Container {
	numNodes := len(nodes)
	if numNodes == 0 {
		return nil
	}

	log().Debugf("calculating restart across interlock nodes: num=%d", numNodes)

	sub := len(proxyContainers) / numNodes

	work := map[string][]dockerclient.Container{}

	for i := 0; i < len(nodes)-1; i++ {
		p, n := proxyContainers[:len(proxyContainers)-sub], proxyContainers[len(proxyContainers)-sub:]
		proxyContainers = p
		work[nodes[i].Id] = n
	}

	work[nodes[len(nodes)-1].Id] = proxyContainers

	containersToRestart := work[l.nodeID]

	ids := []string{}
	for _, c := range containersToRestart {
		ids = append(ids, c.Id[:8])
	}

	log().Debugf("proxy containers to restart: num=%d containers=%s", len(containersToRestart), strings.Join(ids, ","))

	return containersToRestart
}

func (l *LoadBalancer) isExposedContainer(id string) bool {
	log().Debugf("inspecting container: id=%s", id)
	c, err := l.client.InspectContainer(id)
	if err != nil {
		// ignore inspect errors
		log().Errorf("error: id=%s err=%s", id, err)
		return false
	}

	log().Debugf("checking container labels: name=%s", c.Name)
	// ignore proxy containers
	if _, ok := c.Config.Labels[ext.InterlockExtNameLabel]; ok {
		log().Debugf("ignoring proxy container: name=%s", c.Name)
		return false
	}

	if len(l.cfg.ServiceName) > 0 {
		log().Infof("l.cfg.ServiceName = %s", l.cfg.ServiceName)
		serviceName, ok := c.Config.Labels[ext.InterlockExtServiceNameLabel]
		log().Infof("ok = %s serviceName = %s", ok, serviceName)
		if !ok || serviceName != l.cfg.ServiceName {
			log().Debugf("ignoring service container: name=%s labels=%s", c.Name, c.Config.Labels)
			return false
		}
	}

	if _, ok := c.Config.Labels[ext.InterlockAppLabel]; ok {
		log().Debugf("ignoring interlock container: name=%s", c.Name)
		return false
	}

	log().Debugf("checking container ports: name=%s", c.Name)
	// ignore containers without exposed ports
	if len(c.Config.ExposedPorts) == 0 {
		log().Debugf("no ports exposed; ignoring: name=%s", c.Name)
		return false
	}

	log().Debugf("container is monitored; triggering reload: name=%s", c.Name)
	return true
}
