package utils

import (
	ctypes "github.com/docker/docker/api/types/container"
	"github.com/ehazlett/interlock/ext"
)

func IPHash(config *ctypes.Config) bool {
	ipHash := false

	if _, ok := config.Labels[ext.InterlockIPHashLabel]; ok {
		ipHash = true
	}

	return ipHash
}
