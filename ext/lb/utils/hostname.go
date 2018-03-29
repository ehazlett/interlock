package utils

import (
	"github.com/ehazlett/interlock/ext"
)

func Hostname(labels map[string]string) string {
	hostname := "unknown"

	if v, ok := labels[ext.InterlockHostnameLabel]; ok {
		hostname = v
	}

	return hostname
}
