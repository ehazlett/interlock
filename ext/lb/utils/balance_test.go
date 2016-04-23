package utils

import (
	"testing"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func TestBalanceAlgorithm(t *testing.T) {
	testAlgo := "foo"

	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			ext.InterlockBalanceAlgorithmLabel: testAlgo,
		},
	}

	algo := BalanceAlgorithm(cfg)

	if algo != testAlgo {
		t.Fatalf("expected %s; received %s", testAlgo, algo)
	}
}

func TestBalanceAlgorithmDefault(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{},
	}

	algo := BalanceAlgorithm(cfg)

	if algo != DefaultBalanceAlgorithm {
		t.Fatalf("expected %s; received %s", DefaultBalanceAlgorithm, algo)
	}
}
