package nginx

import (
	"io/ioutil"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	etypes "github.com/docker/engine-api/types/events"
	"github.com/ehazlett/interlock/config"
	"golang.org/x/net/context"
)

const (
	pluginName = "nginx"
)

type NginxLoadBalancer struct {
	cfg    *config.ExtensionConfig
	client *client.Client
}

func log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"ext": pluginName,
	})
}

func NewNginxLoadBalancer(c *config.ExtensionConfig, cl *client.Client) (*NginxLoadBalancer, error) {
	// parse config base dir
	c.ConfigBasePath = filepath.Dir(c.ConfigPath)

	lb := &NginxLoadBalancer{
		cfg:    c,
		client: cl,
	}

	return lb, nil
}

func (p *NginxLoadBalancer) Name() string {
	return pluginName
}

func (p *NginxLoadBalancer) HandleEvent(event *etypes.Message) error {
	return nil
}

func (p *NginxLoadBalancer) Template() string {
	if p.cfg.TemplatePath != "" {
		d, err := ioutil.ReadFile(p.cfg.TemplatePath)

		if err == nil {
			return string(d)
		} else {
			return err.Error()
		}
	} else {
		if p.cfg.NginxPlusEnabled {
			return nginxPlusConfTemplate
		}

		return nginxConfTemplate
	}
}

func (p *NginxLoadBalancer) ConfigPath() string {
	return p.cfg.ConfigPath
}

func (p *NginxLoadBalancer) Reload(proxyContainers []types.Container) error {
	// restart all interlock managed nginx containers
	for _, cnt := range proxyContainers {
		// restart
		log().Debugf("reloading proxy container: id=%s", cnt.ID)
		if err := p.client.ContainerKill(context.Background(), cnt.ID, "HUP"); err != nil {
			log().Errorf("error reloading container: id=%s err=%s", cnt.ID[:12], err)
			continue
		}

		log().Infof("restarted proxy container: id=%s", cnt.ID[:12])
	}

	return nil
}
