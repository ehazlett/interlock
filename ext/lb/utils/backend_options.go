package utils

import (
	"strings"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func BackendOptions(config *dockerclient.ContainerConfig) []string {
	options := []string{}

	for l, v := range config.Labels {
		// this is for labels like interlock.backend_option.1=foo
		if strings.Index(l, ext.InterlockBackendOptionLabel) > -1 {
			options = append(options, v)
		}
	}

	return options
}
