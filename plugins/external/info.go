package external

import (
	"github.com/ehazlett/interlock"
)

const (
	pluginName        = "external"
	pluginVersion     = "0.1"
	pluginDescription = "external integration plugin"
	pluginUrl         = "https://github.com/ehazlett/interlock/tree/master/plugins/external"
)

var (
	pluginInfo = &interlock.PluginInfo{
		Name:        pluginName,
		Version:     pluginVersion,
		Description: pluginDescription,
		Url:         pluginUrl,
	}
)
