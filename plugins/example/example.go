package example

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock"
	"github.com/ehazlett/interlock/plugins"
	"github.com/samalba/dockerclient"
)

type ExamplePlugin struct{}

func init() {
	plugins.Register(
		pluginInfo.Name,
		&plugins.RegisteredPlugin{
			New: NewPlugin,
			Info: func() *interlock.PluginInfo {
				return pluginInfo
			},
		})
}

func NewPlugin() (interlock.Plugin, error) {
	return ExamplePlugin{}, nil
}

func (p ExamplePlugin) Info() *interlock.PluginInfo {
	return &interlock.PluginInfo{
		Name:        name,
		Version:     version,
		Description: description,
		Url:         url,
	}
}

func (p ExamplePlugin) HandleEvent(event *dockerclient.Event) error {
	log.Debugf("name=%s action=received event=%s time=%d",
		pluginInfo.Name, event.Id, event.Time,
	)
	return nil
}
