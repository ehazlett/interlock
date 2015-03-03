package example

import (
	"github.com/ehazlett/interlock"
)

const (
	pluginName        = "example"
	pluginVersion     = "0.1"
	pluginDescription = "example plugin"
	pluginUrl         = "https://github.com/ehazlett/interlock"
)

var (
	pluginInfo = &interlock.PluginInfo{
		Name:        pluginName,
		Version:     pluginVersion,
		Description: pluginDescription,
		Url:         pluginUrl,
	}
)
