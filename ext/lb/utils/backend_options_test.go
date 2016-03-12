package utils

import (
	"testing"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func TestBackendOptions(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
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
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{},
	}

	opts := BackendOptions(cfg)

	if len(opts) != 0 {
		t.Fatalf("expected no options; received %s", opts)
	}
}
