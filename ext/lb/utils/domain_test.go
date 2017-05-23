package utils

import (
	"testing"

	ctypes "github.com/docker/docker/api/types/container"
)

func TestDomain(t *testing.T) {
	testDomain := "foo.local"

	cfg := &ctypes.Config{
		Domainname: testDomain,
	}

	domain := Domain(cfg)

	if domain != testDomain {
		t.Fatalf("expected %s; received %s", testDomain, domain)
	}
}
