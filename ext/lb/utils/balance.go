package utils

import (
	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext"
)

const (
	DefaultBalanceAlgorithm = "roundrobin"
)

func BalanceAlgorithm(config types.Container) string {
	algo := DefaultBalanceAlgorithm

	if v, ok := config.Labels[ext.InterlockBalanceAlgorithmLabel]; ok {
		algo = v
	}

	return algo
}
