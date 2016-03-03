package nginx

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/ext"
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

func (p *NginxLoadBalancer) HandleEvent(event *dockerclient.Event) error {
	return nil
}

func (p *NginxLoadBalancer) Update() error {
	c, err := p.GenerateProxyConfig()
	if err != nil {
		return err
	}

	if err := p.saveConfig(c); err != nil {
		return err
	}

	log().Info("configuration updated")

	return nil
}

func (p *NginxLoadBalancer) Reload() error {
	// restart all interlock managed haproxy containers
	containers, err := p.client.ListContainers(false, false, "")
	if err != nil {
		return err
	}

	// find interlock nginx containers
	for _, cnt := range containers {
		if v, ok := cnt.Labels[ext.InterlockExtNameLabel]; ok && v == pluginName {
			// restart
			if err := p.client.KillContainer(cnt.Id, "HUP"); err != nil {
				log().Errorf("error reloading container: id=%s err=%s", cnt.Id[:12], err)
				continue
			}

			log().Infof("restarted proxy container: id=%s name=%s", cnt.Id[:12], cnt.Names[0])
		}
	}

	return nil
}

func (p *NginxLoadBalancer) saveConfig(config *Config) error {
	f, err := os.OpenFile(p.cfg.ConfigPath, os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		ff, fErr := os.Create(p.cfg.ConfigPath)
		defer ff.Close()
		if fErr != nil {
			return fErr
		}
		f = ff
	}
	defer f.Close()

	t := template.New("nginx")
	confTmpl := nginxConfTemplate

	if p.cfg.NginxPlusEnabled {
		confTmpl = nginxPlusConfTemplate
	}
	tmpl, err := t.Parse(confTmpl)
	if err != nil {
		return err
	}

	var c bytes.Buffer

	if err := tmpl.Execute(&c, config); err != nil {
		return err
	}

	if _, err := f.Write(c.Bytes()); err != nil {
		return err
	}

	f.Sync()

	return nil
}
