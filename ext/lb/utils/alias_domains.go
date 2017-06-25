package utils

import (
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext"
)

func AliasDomains(config types.Container) []string {
	aliasDomains := []string{}

	for l, v := range config.Labels {
		// this is for labels like interlock.alias_domain.1=foo.local
		if strings.Index(l, ext.InterlockAliasDomainLabel) > -1 {
			aliasDomains = append(aliasDomains, v)
		}
	}

	return aliasDomains
}
