package nginx

import (
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock/config"
	"github.com/samalba/dockerclient"
)

const (
	pluginName = "nginx"
)

type NginxLoadBalancer struct {
	cfg    *config.ExtensionConfig
	client *dockerclient.DockerClient
}

func log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"ext": pluginName,
	})
}

func NewNginxLoadBalancer(c *config.ExtensionConfig, client *dockerclient.DockerClient) (*NginxLoadBalancer, error) {
	// parse config base dir
	c.ConfigBasePath = filepath.Dir(c.ConfigPath)

	lb := &NginxLoadBalancer{
		cfg:    c,
		client: client,
	}

	return lb, nil
}

func (p *NginxLoadBalancer) Name() string {
	return pluginName
}

func (p *NginxLoadBalancer) HandleEvent(event *dockerclient.Event) error {
	return nil
}

func (p *NginxLoadBalancer) Template() string {
	if p.cfg.NginxPlusEnabled {
		return nginxPlusConfTemplate
	}

	return nginxConfTemplate
}

func (p *NginxLoadBalancer) ConfigPath() string {
	return p.cfg.ConfigPath
}

func (p *NginxLoadBalancer) Reload(proxyContainers []dockerclient.Container) error {
	// restart all interlock managed haproxy containers
	for _, cnt := range proxyContainers {
		// restart
		if err := p.client.KillContainer(cnt.Id, "HUP"); err != nil {
			log().Errorf("error reloading container: id=%s err=%s", cnt.Id[:12], err)
			continue
		}

		log().Infof("restarted proxy container: id=%s name=%s", cnt.Id[:12], cnt.Names[0])
	}

	return nil
}
