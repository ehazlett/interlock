package utils

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext"
)

func TestContextRoot(t *testing.T) {
	testContext := "/context"

	cfg := types.Container{
		Labels: map[string]string{
			ext.InterlockContextRootLabel: testContext,
		},
	}

	context := ContextRoot(cfg.Labels)

	if context != testContext {
		t.Fatalf("expected %s; received %s", testContext, context)
	}
}

func TestContextRootRewrite(t *testing.T) {
	cfg := types.Container{
		Labels: map[string]string{
			ext.InterlockContextRootRewriteLabel: "true",
		},
	}

	rewrite := ContextRootRewrite(cfg.Labels)

	if !rewrite {
		t.Fatal("expected context root rewrite")
	}
}
