package utils

import (
	"strings"

	ctypes "github.com/docker/docker/api/types/container"
	"github.com/ehazlett/interlock/ext"
)

func AliasDomains(config *ctypes.Config) []string {
	aliasDomains := []string{}

	for l, v := range config.Labels {
		// this is for labels like interlock.alias_domain.1=foo.local
		if strings.Index(l, ext.InterlockAliasDomainLabel) > -1 {
			aliasDomains = append(aliasDomains, v)
		}
	}

	return aliasDomains
}
