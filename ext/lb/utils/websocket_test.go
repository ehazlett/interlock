package utils

import (
	"testing"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func TestWebsocketEndpoints(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			ext.InterlockWebsocketEndpointLabel + ".0": "foo.bar",
			ext.InterlockWebsocketEndpointLabel + ".1": "bar.baz",
		},
	}

	ep := WebsocketEndpoints(cfg)

	if len(ep) != 2 {
		t.Fatalf("expected %d endpoints; received %d", len(cfg.Labels), len(ep))
	}
}

func TestWebsocketEndpointsNoLabels(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{},
	}

	ep := WebsocketEndpoints(cfg)

	if len(ep) != 0 {
		t.Fatalf("expected no endpoints; received %s", ep)
	}
}
