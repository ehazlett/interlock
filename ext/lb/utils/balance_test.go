package utils

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext"
)

func TestBalanceAlgorithm(t *testing.T) {
	testAlgo := "foo"

	cfg := types.Container{
		Labels: map[string]string{
			ext.InterlockBalanceAlgorithmLabel: testAlgo,
		},
	}

	algo := BalanceAlgorithm(cfg.Labels)

	if algo != testAlgo {
		t.Fatalf("expected %s; received %s", testAlgo, algo)
	}
}

func TestBalanceAlgorithmDefault(t *testing.T) {
	cfg := types.Container{
		Labels: map[string]string{},
	}

	algo := BalanceAlgorithm(cfg.Labels)

	if algo != DefaultBalanceAlgorithm {
		t.Fatalf("expected %s; received %s", DefaultBalanceAlgorithm, algo)
	}
}
