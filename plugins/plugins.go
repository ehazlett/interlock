package plugins

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/interlock"
	"github.com/samalba/dockerclient"
)

var (
	plugins map[string]*RegisteredPlugin
)

func init() {
	plugins = make(map[string]*RegisteredPlugin)
}

type (
	RegisteredPlugin struct {
		New  func(config *interlock.Config, client *dockerclient.DockerClient) (interlock.Plugin, error)
		Info func() *interlock.PluginInfo
	}
)

func DispatchEvent(config *interlock.Config, client *dockerclient.DockerClient, event *dockerclient.Event, errorChan chan error) {
	for _, plugin := range plugins {
		p, err := plugin.New(config, client)
		if err != nil {
			errorChan <- err
			continue
		}
		log.Infof("dispatching event to plugin: name=%s version=%s",
			p.Info().Name, p.Info().Version)
		if err := p.HandleEvent(event); err != nil {
			errorChan <- err
			continue
		}
	}
}

func Register(name string, registeredPlugin *RegisteredPlugin) error {
	if _, exists := plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}
	plugins[name] = registeredPlugin
	return nil
}

func GetPlugins() map[string]*RegisteredPlugin {
	return plugins
}

func GetCommands() []cli.Command {
	return nil
}

func NewPlugin(name string, config *interlock.Config, client *dockerclient.DockerClient) (interlock.Plugin, error) {
	plugin, exists := plugins[name]
	if !exists {
		return nil, fmt.Errorf("unknown plugin: %s", name)
	}
	return plugin.New(config, client)
}
