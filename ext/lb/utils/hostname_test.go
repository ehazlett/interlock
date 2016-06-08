package utils

import (
	"testing"

	ctypes "github.com/docker/engine-api/types/container"
	"github.com/ehazlett/interlock/ext"
)

func TestHostname(t *testing.T) {
	testHostname := "foo"

	cfg := &ctypes.Config{
		Hostname: testHostname,
	}

	hostname := Hostname(cfg)

	if hostname != testHostname {
		t.Fatalf("expected %s; received %s", testHostname, hostname)
	}
}

func TestHostnameLabel(t *testing.T) {
	testHostname := "foo"
	testLabelHostname := "bar"

	cfg := &ctypes.Config{
		Hostname: testHostname,
		Labels: map[string]string{
			ext.InterlockHostnameLabel: testLabelHostname,
		},
	}

	hostname := Hostname(cfg)

	if hostname != testLabelHostname {
		t.Fatalf("expected %s; received %s", testLabelHostname, hostname)
	}
}
