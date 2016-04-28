package utils

import (
	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func IPHash(config *dockerclient.ContainerConfig) bool {
	ipHash := false

	if _, ok := config.Labels[ext.InterlockIPHashLabel]; ok {
		ipHash = true
	}

	return ipHash
}
