package haproxy

import (
	"github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock/config"
	"github.com/samalba/dockerclient"
	"io/ioutil"
)

const (
	pluginName = "haproxy"
)

type HAProxyLoadBalancer struct {
	cfg    *config.ExtensionConfig
	client *dockerclient.DockerClient
}

func log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"ext": pluginName,
	})
}

func NewHAProxyLoadBalancer(c *config.ExtensionConfig, client *dockerclient.DockerClient) (*HAProxyLoadBalancer, error) {
	lb := &HAProxyLoadBalancer{
		cfg:    c,
		client: client,
	}

	return lb, nil
}

func (p *HAProxyLoadBalancer) Name() string {
	return pluginName
}

func (p *HAProxyLoadBalancer) HandleEvent(event *dockerclient.Event) error {
	return nil
}

func (p *HAProxyLoadBalancer) ConfigPath() string {
	return p.cfg.ConfigPath
}

func (p *HAProxyLoadBalancer) Template() string {
	d, err := ioutil.ReadFile(p.cfg.TemplatePath)

	if err == nil {
		return string(d)
	} else {
		return err.Error()
	}
}

func (p *HAProxyLoadBalancer) Reload(proxyContainers []dockerclient.Container) error {
	// drop SYN to allow for restarts
	if err := p.dropSYN(); err != nil {
		log().Warnf("error signaling clients to resend; you will notice dropped packets: %s", err)
	}

	for _, cnt := range proxyContainers {
		// restart
		if err := p.client.RestartContainer(cnt.Id, 1); err != nil {
			log().Errorf("error restarting container: id=%s err=%s", cnt.Id[:12], err)
			continue
		}

		log().Infof("restarted proxy container: id=%s name=%s", cnt.Id[:12], cnt.Names[0])
	}

	if err := p.resumeSYN(); err != nil {
		log().Warnf("error signaling clients to resume; you will notice dropped packets: %s", err)
	}

	return nil
}
