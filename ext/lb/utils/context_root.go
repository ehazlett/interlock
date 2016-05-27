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

func ContextRootRewrite(config *dockerclient.ContainerConfig) bool {
	if _, ok := config.Labels[ext.InterlockContextRootRewriteLabel]; ok {
		return true
	}

	return false
}
