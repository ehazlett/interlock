package utils

import (
	"testing"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func TestProtocol(t *testing.T) {
	testProtocol := "proxy"

	cfg := &dockerclient.ContainerConfig{
		Protocol: testProtocol,
	}

	protocol := Protocol(cfg)

	if protocol != testProtocol {
		t.Fatalf("expected %s; received %s", testProtocol, protocol)
	}
}

func TestProtocolLabel(t *testing.T) {
	testProtocol := "foo"
	testLabelProtocol := "proxy"

	cfg := &dockerclient.ContainerConfig{
		Protocol: testProtocol,
		Labels: map[string]string{
			ext.InterlockProtocolLabel: testLabelProtocol,
		},
	}

	protocol := Protocol(cfg)

	if protocol != testLabelProtocol {
		t.Fatalf("expected %s; received %s", testLabelProtocol, protocol)
	}
}
