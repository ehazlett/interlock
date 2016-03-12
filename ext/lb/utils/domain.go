package utils

import (
	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func Domain(config *dockerclient.ContainerConfig) string {
	domain := config.Domainname

	if v, ok := config.Labels[ext.InterlockDomainLabel]; ok {
		domain = v
	}

	return domain
}
