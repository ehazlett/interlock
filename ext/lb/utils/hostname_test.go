package utils

import (
	"testing"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func TestHostname(t *testing.T) {
	testHostname := "foo"

	cfg := &dockerclient.ContainerConfig{
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

	cfg := &dockerclient.ContainerConfig{
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
