package utils

import (
	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func Hostname(config *dockerclient.ContainerConfig) string {
	hostname := config.Hostname

	if v, ok := config.Labels[ext.InterlockHostnameLabel]; ok {
		hostname = v
	}

	return hostname
}
