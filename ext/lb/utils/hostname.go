package utils

import (
	ctypes "github.com/docker/docker/api/types/container"
	"github.com/ehazlett/interlock/ext"
)

func Hostname(config *ctypes.Config) string {
	hostname := config.Hostname

	if v, ok := config.Labels[ext.InterlockHostnameLabel]; ok {
		hostname = v
	}

	return hostname
}
