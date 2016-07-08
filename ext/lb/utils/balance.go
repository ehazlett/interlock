package utils

import "github.com/ehazlett/interlock/ext"

const (
	DefaultBalanceAlgorithm = "roundrobin"
)

func BalanceAlgorithm(labels map[string]string) string {
	algo := DefaultBalanceAlgorithm

	if v, ok := labels[ext.InterlockBalanceAlgorithmLabel]; ok {
		algo = v
	}

	return algo
}
