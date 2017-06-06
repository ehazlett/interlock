package nginx

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/events"
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

func (p *NginxLoadBalancer) HandleEvent(event *events.Message) error {
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
		resp, err := p.client.ContainerExecCreate(context.Background(), cnt.ID, types.ExecConfig{
			User: "root",
			Cmd: []string{
				"nginx",
				"-s",
				"reload",
			},
			Detach:       false,
			AttachStdout: true,
			AttachStderr: true,
		})
		if err != nil {
			log().Errorf("error reloading container (exec create): id=%s err=%s", cnt.ID[:12], err)
			continue
		}

		aResp, err := p.client.ContainerExecAttach(context.Background(), resp.ID, types.ExecConfig{
			AttachStdout: true,
			AttachStderr: true,
		})
		if err != nil {
			log().Errorf("error reloading container (exec attach): id=%s err=%s", cnt.ID[:12], err)
			continue
		}
		defer aResp.Conn.Close()

		// wait for exec to finish
		var res types.ContainerExecInspect
		for {
			r, err := p.client.ContainerExecInspect(context.Background(), resp.ID)
			if err != nil {
				log().Errorf("error reloading container (exec inspect): id=%s err=%s", cnt.ID[:12], err)
				continue
			}

			if !r.Running {
				res = r
				break
			}
		}

		if res.ExitCode != 0 {
			out, err := aResp.Reader.ReadString('\n')
			if err != nil {
				log().Error("error reloading container: unable to read output from exec")
				continue
			}
			log().Errorf("error reloading container, invalid proxy configuration: %s", strings.TrimSpace(out))
			continue
		}

		log().Infof("restarted proxy container: id=%s name=%s", cnt.ID[:12], cnt.Names[0])
	}

	return nil
}
