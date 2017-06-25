package utils

import (
	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext"
)

func Domain(config types.Container) string {
	domain := "local"

	if v, ok := config.Labels[ext.InterlockDomainLabel]; ok {
		domain = v
	}

	return domain
}
