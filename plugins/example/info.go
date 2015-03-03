package example

import (
	"github.com/ehazlett/interlock"
)

const (
	name        = "example"
	version     = "0.1"
	description = "example plugin"
	url         = "https://github.com/ehazlett/interlock"
)

var (
	pluginInfo = &interlock.PluginInfo{
		Name:        name,
		Version:     version,
		Description: description,
		Url:         url,
	}
)
