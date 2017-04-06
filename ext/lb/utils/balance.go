package utils

import (
	ctypes "github.com/docker/docker/api/types/container"
	"github.com/ehazlett/interlock/ext"
)

const (
	DefaultBalanceAlgorithm = "roundrobin"
)

func BalanceAlgorithm(config *ctypes.Config) string {
	algo := DefaultBalanceAlgorithm

	if v, ok := config.Labels[ext.InterlockBalanceAlgorithmLabel]; ok {
		algo = v
	}

	return algo
}
