package utils

import (
	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func Protocol(config *dockerclient.ContainerConfig) string {

	if v, ok := config.Labels[ext.InterlockProtocolLabel]; ok {
		return v
	}

	return ""
}
