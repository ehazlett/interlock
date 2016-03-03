package haproxy

import (
	"bytes"
	"os"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
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

func (p *HAProxyLoadBalancer) HandleEvent(event *dockerclient.Event) error {
	return nil
}

func (p *HAProxyLoadBalancer) Update() error {
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

func (p *HAProxyLoadBalancer) Reload() error {
	// drop SYN to allow for restarts
	if err := p.dropSYN(); err != nil {
		log().Warnf("error signaling clients to resend; you will notice dropped packets: %s", err)
	}

	if err := p.reloadProxyContainers(); err != nil {
		return err
	}

	if err := p.resumeSYN(); err != nil {
		log().Warnf("error signaling clients to resume; you will notice dropped packets: %s", err)
	}

	return nil
}

func (p *HAProxyLoadBalancer) reloadProxyContainers() error {
	// restart all interlock managed haproxy containers
	containers, err := p.client.ListContainers(false, false, "")
	if err != nil {
		return err
	}

	// find interlock haproxy containers
	for _, cnt := range containers {
		if v, ok := cnt.Labels[ext.InterlockExtNameLabel]; ok && v == pluginName {
			// restart
			if err := p.client.RestartContainer(cnt.Id, 1); err != nil {
				log().Errorf("error restarting container: id=%s err=%s", cnt.Id[:12], err)
				continue
			}

			log().Infof("restarted proxy container: id=%s name=%s", cnt.Id[:12], cnt.Names[0])
		}
	}

	return nil
}

func (p *HAProxyLoadBalancer) saveConfig(config *Config) error {
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

	t := template.New("haproxy")
	tmpl, err := t.Parse(haproxyConfTemplate)
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
