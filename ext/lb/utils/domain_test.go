package utils

import (
	"testing"

	"github.com/samalba/dockerclient"
)

func TestDomain(t *testing.T) {
	testDomain := "foo.local"

	cfg := &dockerclient.ContainerConfig{
		Domainname: testDomain,
	}

	domain := Domain(cfg)

	if domain != testDomain {
		t.Fatalf("expected %s; received %s", testDomain, domain)
	}
}
