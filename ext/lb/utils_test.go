package lb

import (
	"io"
	"os"
	"testing"
)

func TestGetContainerID(t *testing.T) {
	if _, err := os.Stat("/proc/self/cgroup"); err != nil {
		if os.IsNotExist(err) {
			t.Skipf("skipping GetContainerID; does not look like I am in a container")
		}
	}

	id, err := getContainerID()
	if err != nil {
		if err == io.EOF {
			t.Skipf("skipping GetContainerID; does not look like I am in a normal container")
		}

		t.Fatal(err)
	}

	if id == "" {
		t.Fatalf("expected ID")
	}
}
