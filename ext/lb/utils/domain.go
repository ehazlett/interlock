package utils

import (
	"github.com/ehazlett/interlock/ext"
)

func Domain(labels map[string]string) string {
	domain := "local"

	if v, ok := labels[ext.InterlockDomainLabel]; ok {
		domain = v
	}

	return domain
}
