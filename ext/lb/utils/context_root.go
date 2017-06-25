package utils

import (
	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext"
)

func ContextRoot(config types.Container) string {
	if v, ok := config.Labels[ext.InterlockContextRootLabel]; ok {
		return v
	}

	return ""
}

func ContextRootRewrite(config types.Container) bool {
	if _, ok := config.Labels[ext.InterlockContextRootRewriteLabel]; ok {
		return true
	}

	return false
}
