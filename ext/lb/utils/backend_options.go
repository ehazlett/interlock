package utils

import (
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext"
)

func BackendOptions(config types.Container) []string {
	options := []string{}

	for l, v := range config.Labels {
		// this is for labels like interlock.backend_option.1=foo
		if strings.Index(l, ext.InterlockBackendOptionLabel) > -1 {
			options = append(options, v)
		}
	}

	return options
}
