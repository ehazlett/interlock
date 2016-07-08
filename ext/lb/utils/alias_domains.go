package utils

import (
	"strings"

	"github.com/ehazlett/interlock/ext"
)

func AliasDomains(labels map[string]string) []string {
	aliasDomains := []string{}

	for l, v := range labels {
		// this is for labels like interlock.alias_domain.1=foo.local
		if strings.Index(l, ext.InterlockAliasDomainLabel) > -1 {
			aliasDomains = append(aliasDomains, v)
		}
	}

	return aliasDomains
}
