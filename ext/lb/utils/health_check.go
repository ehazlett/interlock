package utils

import (
	"strconv"

	"github.com/ehazlett/interlock/ext"
)

const (
	DefaultHealthCheckInterval = 5000
)

func HealthCheck(labels map[string]string) string {
	if v, ok := labels[ext.InterlockHealthCheckLabel]; ok {
		return v
	}

	return ""
}

func HealthCheckInterval(labels map[string]string) (int, error) {
	checkInterval := DefaultHealthCheckInterval

	if v, ok := labels[ext.InterlockHealthCheckIntervalLabel]; ok && v != "" {
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
