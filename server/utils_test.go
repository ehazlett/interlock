package server

import (
	"testing"
)

func TestParseSwarmNodes(t *testing.T) {
	driverStatus := [][2]string{
		[2]string{"\u0008Strategy", "spread"},
		[2]string{"\u0008Filters", "affinity, health, constraint"},
		[2]string{"\u0008Nodes", "1"},
		[2]string{"localhost", "127.0.0.1:2375"},
		[2]string{" └ Containers", "10"},
		[2]string{" └ Reserved CPUs", "1 / 4"},
		[2]string{" └ Reserved Memory", "2 / 8.083GiB"},
		[2]string{" └ Labels", "executiondriver=native-0.2, kernelversion=3.16.0-4-amd64, operatingsystem=Debian GNU/Linux 8 (jessie), storagedriver=btrfs"},
		[2]string{"remote", "1.2.3.4:2375"},
		[2]string{" └ Containers", "3"},
		[2]string{" └ Reserved CPUs", "0 / 4"},
		[2]string{" └ Reserved Memory", "0 / 8.083GiB"},
		[2]string{" └ Labels", "executiondriver=native-0.2, kernelversion=3.16.0-4-amd64, operatingsystem=Debian GNU/Linux 8 (jessie), storagedriver=aufs"},
	}

	nodes, err := parseSwarmNodes(driverStatus)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes; received %d", len(nodes))
	}
}
