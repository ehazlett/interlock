package utils

import (
	"strings"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func AliasDomains(config *dockerclient.ContainerConfig) []string {
	aliasDomains := []string{}

	for l, v := range config.Labels {
		// this is for labels like interlock.alias_domain.1=foo.local
		if strings.Index(l, ext.InterlockAliasDomainLabel) > -1 {
			aliasDomains = append(aliasDomains, v)
		}
	}

	return aliasDomains
}
