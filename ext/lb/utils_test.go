package lb

import (
	"io"
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
		if err == io.EOF {
			t.Skipf("skipping GetNodeID; does not look like i am in a normal container")
		}

		t.Fatal(err)
	}

	if id == "" {
		t.Fatalf("expected ID")
	}
}
