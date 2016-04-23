package utils

import (
	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func IpHash(config *dockerclient.ContainerConfig) bool {
	ipHash := false

	if _, ok := config.Labels[ext.InterlockIpHash]; ok {
		ipHash = true
	}

	return ipHash
}
