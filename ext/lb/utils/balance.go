package utils

import (
	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

const (
	DefaultBalanceAlgorithm = "roundrobin"
)

func BalanceAlgorithm(config *dockerclient.ContainerConfig) string {
	algo := DefaultBalanceAlgorithm

	if v, ok := config.Labels[ext.InterlockBalanceAlgorithmLabel]; ok {
		algo = v
	}

	return algo
}
