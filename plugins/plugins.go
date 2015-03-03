package plugins

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock"
	"github.com/samalba/dockerclient"
)

var (
	plugins map[string]*RegisteredPlugin
)

func init() {
	plugins = make(map[string]*RegisteredPlugin)
}

type RegisteredPlugin struct {
	New  func() (interlock.Plugin, error)
	Info func() *interlock.PluginInfo
}

func DispatchEvent(event *dockerclient.Event, errorChan chan error) {
	for _, plugin := range plugins {
		log.Debugf("dispatching event to plugin: name=%s version=%s",
			plugin.Info().Name, plugin.Info().Version)
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
	//allPlugins := []*interlock.Plugin{}
	//for _, p := range plugins {
	//	log.Debug(p.Info())
	//	plugin, err := p.New()
	//	if err != nil {
	//		log.Errorf("error loading plugin: name=%s version=%s error: %s",
	//			p.Info().Name, p.Info().Version, err)
	//		continue
	//	}
	//	allPlugins = append(allPlugins, &plugin)
	//}
	//log.Debug(allPlugins)
	//return allPlugins
	return plugins
}

func NewDriver(name string, info *interlock.PluginInfo) (interlock.Plugin, error) {
	plugin, exists := plugins[name]
	if !exists {
		return nil, fmt.Errorf("unknown plugin: %s", name)
	}
	return plugin.New()
}
