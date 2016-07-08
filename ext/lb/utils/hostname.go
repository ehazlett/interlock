package utils

import "github.com/ehazlett/interlock/ext"

func Hostname(labels map[string]string) string {
	if v, ok := labels[ext.InterlockHostnameLabel]; ok {
		return v
	}

	return ""
}
