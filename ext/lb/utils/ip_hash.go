package utils

import (
	"github.com/ehazlett/interlock/ext"
)

func IPHash(labels map[string]string) bool {
	ipHash := false

	if _, ok := labels[ext.InterlockIPHashLabel]; ok {
		ipHash = true
	}

	return ipHash
}
