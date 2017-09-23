package utils

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext"
)

func TestHostnameLabel(t *testing.T) {
	testLabelHostname := "bar"

	cfg := types.Container{
		Labels: map[string]string{
			ext.InterlockHostnameLabel: testLabelHostname,
		},
	}

	hostname := Hostname(cfg.Labels)

	if hostname != testLabelHostname {
		t.Fatalf("expected %s; received %s", testLabelHostname, hostname)
	}
}
