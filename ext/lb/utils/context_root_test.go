package utils

import (
	"testing"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func TestContextRoot(t *testing.T) {
	testContext := "/context"

	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			ext.InterlockContextRootLabel: testContext,
		},
	}

	context := ContextRoot(cfg)

	if context != testContext {
		t.Fatalf("expected %s; received %s", testContext, context)
	}
}
