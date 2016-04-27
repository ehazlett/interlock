package lb

import (
<<<<<<< HEAD
<<<<<<< HEAD
	"io"
=======
>>>>>>> e2cf8e1... calculate proxy container restart among interlock instances
=======
>>>>>>> fe1739529c0f1908291b79455cf132ed000d2e42
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
<<<<<<< HEAD
		if err == io.EOF {
			t.Skipf("skipping GetNodeID; does not look like i am in a normal container")
		}

=======
>>>>>>> e2cf8e1... calculate proxy container restart among interlock instances
=======
>>>>>>> fe1739529c0f1908291b79455cf132ed000d2e42
		t.Fatal(err)
	}

	if id == "" {
		t.Fatalf("expected ID")
	}
}
