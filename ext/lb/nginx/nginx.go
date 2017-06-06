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
		res, err := p.waitForExec(resp.ID)
		if err != nil {
			log().Errorf("error reloading container (exec attach): id=%s err=%s", cnt.ID[:12], err)
			continue
		}

		if res.ExitCode != 0 {
			out, err := aResp.Reader.ReadString('\n')
			if err != nil {
				log().Error("error reloading container: unable to read output from exec")
				continue
			}
			// restore
			log().Warn("restoring proxy config")
			if err := p.restoreConfig(cnt.ID); err != nil {
				log().Errorf("error reloading container: error restoring config: %s", err)
				continue
			}
			log().Errorf("error reloading container, invalid proxy configuration: %s", strings.TrimSpace(out))
			continue
		}

		// backup config
		if err := p.backupConfig(cnt.ID); err != nil {
			log().WithFields(logrus.Fields{
				"id": cnt.ID,
			}).Errorf("error backing up config: skipping update")
			continue
		}

		log().Infof("restarted proxy container: id=%s name=%s", cnt.ID[:12], cnt.Names[0])
	}

	return nil
}

func (p *NginxLoadBalancer) backupConfig(id string) error {
	resp, err := p.client.ContainerExecCreate(context.Background(), id, types.ExecConfig{
		User: "root",
		Cmd: []string{
			"cp",
			"-f",
			p.cfg.ConfigPath,
			p.cfg.ConfigPath + ".interlock",
		},
		Detach:       false,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return err
	}

	if err := p.client.ContainerExecStart(context.Background(), resp.ID, types.ExecStartCheck{}); err != nil {
		return err
	}

	if _, err := p.waitForExec(resp.ID); err != nil {
		return err
	}

	return nil
}

func (p *NginxLoadBalancer) restoreConfig(id string) error {
	resp, err := p.client.ContainerExecCreate(context.Background(), id, types.ExecConfig{
		User: "root",
		Cmd: []string{
			"cp",
			"-f",
			p.cfg.ConfigPath + ".interlock",
			p.cfg.ConfigPath,
		},
		Detach:       false,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return err
	}

	if err := p.client.ContainerExecStart(context.Background(), resp.ID, types.ExecStartCheck{}); err != nil {
		return err
	}

	if _, err := p.waitForExec(resp.ID); err != nil {
		return err
	}

	return nil
}

func (p *NginxLoadBalancer) waitForExec(execID string) (types.ContainerExecInspect, error) {
	var res types.ContainerExecInspect
	for {
		r, err := p.client.ContainerExecInspect(context.Background(), execID)
		if err != nil {
			return res, err
		}

		if !r.Running {
			res = r
			break
		}
	}

	return res, nil
}
