package utils

import (
	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func ContextRoot(config *dockerclient.ContainerConfig) string {
	if v, ok := config.Labels[ext.InterlockContextRootLabel]; ok {
		return v
	}

	return ""
}
