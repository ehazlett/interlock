package utils

import (
	"testing"

	"github.com/docker/docker/api/types"
)

func TestDomain(t *testing.T) {
	testDomain := "foo.local"

	cfg := types.Container{
		Labels: map[string]string{
			"interlock.domain": testDomain,
		},
	}

	domain := Domain(cfg.Labels)

	if domain != testDomain {
		t.Fatalf("expected %s; received %s", testDomain, domain)
	}
}
