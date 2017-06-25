package utils

import (
	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext"
)

func IPHash(config types.Container) bool {
	ipHash := false

	if _, ok := config.Labels[ext.InterlockIPHashLabel]; ok {
		ipHash = true
	}

	return ipHash
}
