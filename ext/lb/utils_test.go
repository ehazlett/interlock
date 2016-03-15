package lb

import (
<<<<<<< HEAD
	"io"
=======
>>>>>>> e2cf8e1... calculate proxy container restart among interlock instances
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
<<<<<<< HEAD
		if err == io.EOF {
			t.Skipf("skipping GetNodeID; does not look like i am in a normal container")
		}

=======
>>>>>>> e2cf8e1... calculate proxy container restart among interlock instances
		t.Fatal(err)
	}

	if id == "" {
		t.Fatalf("expected ID")
	}
}
