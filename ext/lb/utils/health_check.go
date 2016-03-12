package utils

import (
	"strconv"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

const (
	DefaultHealthCheckInterval = 5000
)

func HealthCheck(config *dockerclient.ContainerConfig) string {
	if v, ok := config.Labels[ext.InterlockHealthCheckLabel]; ok {
		return v
	}

	return ""
}

func HealthCheckInterval(config *dockerclient.ContainerConfig) (int, error) {
	checkInterval := DefaultHealthCheckInterval

	if v, ok := config.Labels[ext.InterlockHealthCheckIntervalLabel]; ok && v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return -1, err
		}
		if i != 0 {
			checkInterval = i
		}
	}

	return checkInterval, nil
}
