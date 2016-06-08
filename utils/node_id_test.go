package utils

import (
	"os"
	"testing"
)

func TestGetNodeID(t *testing.T) {
	if _, err := os.Stat("/proc/self/cgroup"); err != nil {
		if os.IsNotExist(err) {
			t.Skipf("skipping GetNodeID; does not look like i am in a container")
		}
	}

	id, err := getNodeID()
	if err != nil {
		t.Skipf("skipping GetNodeID; does not look like i am in a normal container")
	}

	if id == "" {
		t.Fatalf("expected ID")
	}
}
