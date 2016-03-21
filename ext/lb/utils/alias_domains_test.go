package utils

import (
	"testing"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func TestAliasDomains(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			ext.InterlockAliasDomainLabel + ".0": "foo.bar",
			ext.InterlockAliasDomainLabel + ".1": "bar.baz",
		},
	}

	ep := AliasDomains(cfg)

	if len(ep) != 2 {
		t.Fatalf("expected %d alias domains; received %d", len(cfg.Labels), len(ep))
	}
}

func TestAliasDomainsNoLabels(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{},
	}

	ep := AliasDomains(cfg)

	if len(ep) != 0 {
		t.Fatalf("expected no alias domains; received %s", ep)
	}
}
