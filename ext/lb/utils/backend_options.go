package utils

import (
	"strings"

	"github.com/ehazlett/interlock/ext"
)

func BackendOptions(labels map[string]string) []string {
	options := []string{}

	for l, v := range labels {
		// this is for labels like interlock.backend_option.1=foo
		if strings.Index(l, ext.InterlockBackendOptionLabel) > -1 {
			options = append(options, v)
		}
	}

	return options
}
