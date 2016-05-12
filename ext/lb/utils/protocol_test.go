package utils

import (
	"testing"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func TestProtocolLabel(t *testing.T) {
	testLabelProtocol := "proxy"

	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			ext.InterlockProtocolLabel: testLabelProtocol,
		},
	}

	protocol := Protocol(cfg)

	if protocol != testLabelProtocol {
		t.Fatalf("expected %s; received %s", testLabelProtocol, protocol)
	}
}
