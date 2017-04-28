package utils

import (
	"testing"

	ctypes "github.com/docker/docker/api/types/container"
	"github.com/ehazlett/interlock/ext"
)

func TestBackendOptions(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{
			ext.InterlockBackendOptionLabel + ".0": "foo=bar",
			ext.InterlockBackendOptionLabel + ".1": "bar=baz",
		},
	}

	opts := BackendOptions(cfg)

	if len(opts) != 2 {
		t.Fatalf("expected %d options; received %d", len(cfg.Labels), len(opts))
	}
}

func TestBackendOptionsNoLabels(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{},
	}

	opts := BackendOptions(cfg)

	if len(opts) != 0 {
		t.Fatalf("expected no options; received %s", opts)
	}
}
