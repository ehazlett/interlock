package utils

import (
	"testing"

	ctypes "github.com/docker/engine-api/types/container"
	"github.com/ehazlett/interlock/ext"
)

func TestBalanceAlgorithm(t *testing.T) {
	testAlgo := "foo"

	cfg := &ctypes.Config{
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
	cfg := &ctypes.Config{
		Labels: map[string]string{},
	}

	algo := BalanceAlgorithm(cfg.Labels)

	if algo != DefaultBalanceAlgorithm {
		t.Fatalf("expected %s; received %s", DefaultBalanceAlgorithm, algo)
	}
}
