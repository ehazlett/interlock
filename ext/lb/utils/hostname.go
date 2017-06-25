package utils

import (
	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext"
)

func Hostname(config types.Container) string {
	hostname := "unknown"

	if v, ok := config.Labels[ext.InterlockHostnameLabel]; ok {
		hostname = v
	}

	return hostname
}
