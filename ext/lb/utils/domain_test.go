package utils

import (
	"testing"

	ctypes "github.com/docker/engine-api/types/container"
	"github.com/ehazlett/interlock/ext"
)

func TestDomain(t *testing.T) {
	testDomain := "foo.local"

	cfg := &ctypes.Config{
		Labels: map[string]string{
			ext.InterlockDomainLabel: testDomain,
		},
	}

	domain := Domain(cfg.Labels)

	if domain != testDomain {
		t.Fatalf("expected %s; received %s", testDomain, domain)
	}
}
