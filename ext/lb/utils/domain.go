package utils

import (
	ctypes "github.com/docker/docker/api/types/container"
	"github.com/ehazlett/interlock/ext"
)

func Domain(config *ctypes.Config) string {
	domain := config.Domainname

	if v, ok := config.Labels[ext.InterlockDomainLabel]; ok {
		domain = v
	}

	return domain
}
