package haproxy

import (
	"io/ioutil"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	etypes "github.com/docker/engine-api/types/events"
	"github.com/ehazlett/interlock/config"
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

func (p *HAProxyLoadBalancer) HandleEvent(event *etypes.Message) error {
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
	// reload all interlock managed haproxy containers
	for _, cnt := range proxyContainers {
		// update the proxy container status
		cInfo, err := p.client.ContainerInspect(context.Background(), cnt.ID)
		if err != nil {
			log().Errorf("unable to inspect proxy container: %s", err)
			continue
		}
		switch cInfo.State.Status {
		case "exited":
			log().Infof("restarting proxy container: id=%s", cnt.ID)
			d := time.Millisecond * 1000
			if err := p.client.ContainerRestart(context.Background(), cnt.ID, &d); err != nil {
				log().Errorf("error restarting container: id=%s err=%s", cnt.ID[:12], err)
				continue
			}
		case "running":
			log().Debugf("reloading proxy container: id=%s", cnt.ID)
			if err := p.client.ContainerKill(context.Background(), cnt.ID, "HUP"); err != nil {
				log().Errorf("error reloading container: id=%s err=%s", cnt.ID[:12], err)
				continue
			}
		default:
			log().Infof("haproxy container id=%s name=%s in state %s", cnt.ID[:12], cnt.Names[0], cInfo.State.Status)
			continue
		}

		log().Infof("reloaded proxy container: id=%s name=%s", cnt.ID[:12], cnt.Names[0])
	}

	return nil
}
