package haproxy

import (
	"io/ioutil"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/events"
	"golang.org/x/net/context"
)

const (
	pluginName = "haproxy"
)

type HAProxyLoadBalancer struct {
	cfg    *config.ExtensionConfig
	client *client.Client
}

func log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"ext": pluginName,
	})
}

func NewHAProxyLoadBalancer(c *config.ExtensionConfig, cl *client.Client) (*HAProxyLoadBalancer, error) {
	lb := &HAProxyLoadBalancer{
		cfg:    c,
		client: cl,
	}

	return lb, nil
}

func (p *HAProxyLoadBalancer) Name() string {
	return pluginName
}

func (p *HAProxyLoadBalancer) HandleEvent(event *events.Message) error {
	return nil
}

func (p *HAProxyLoadBalancer) ConfigPath() string {
	return p.cfg.ConfigPath
}

func (p *HAProxyLoadBalancer) Template() string {
	if p.cfg.TemplatePath != "" {
		d, err := ioutil.ReadFile(p.cfg.TemplatePath)

		if err == nil {
			return string(d)
		} else {
			return err.Error()
		}
	} else {
		return haproxyConfTemplate
	}

}

func (p *HAProxyLoadBalancer) Reload(proxyContainers []types.Container) error {
	// drop SYN to allow for restarts
	if err := p.dropSYN(); err != nil {
		log().Warnf("error signaling clients to resend; you will notice dropped packets: %s", err)
	}

	for _, cnt := range proxyContainers {
		// restart
		log().Debugf("restarting proxy container: id=%s", cnt.ID)
		d := time.Millisecond * 1000
		if err := p.client.ContainerRestart(context.Background(), cnt.ID, &d); err != nil {
			log().Errorf("error restarting container: id=%s err=%s", cnt.ID[:12], err)
			continue
		}

		log().Infof("restarted proxy container: id=%s name=%s", cnt.ID[:12], cnt.Names[0])
	}

	if err := p.resumeSYN(); err != nil {
		log().Warnf("error signaling clients to resume; you will notice dropped packets: %s", err)
	}

	return nil
}
