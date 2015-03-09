package stats

import (
	"github.com/ehazlett/interlock"
)

const (
	pluginName        = "stats"
	pluginVersion     = "0.1"
	pluginDescription = "cluster stats to graphite"
	pluginUrl         = "https://github.com/ehazlett/interlock/tree/master/plugins/stats"
)

var (
	pluginInfo = &interlock.PluginInfo{
		Name:        pluginName,
		Version:     pluginVersion,
		Description: pluginDescription,
		Url:         pluginUrl,
	}
)
