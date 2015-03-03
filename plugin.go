package interlock

import (
	"github.com/samalba/dockerclient"
)

type PluginInfo struct {
	Name        string
	Version     string
	Description string
	Url         string
}

type Plugin interface {
	Info() *PluginInfo
	HandleEvent(event *dockerclient.Event) error
}
