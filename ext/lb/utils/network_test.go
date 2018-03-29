package utils

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext"
)

func TestUseOverlay(t *testing.T) {
	cfg := types.Container{
		Labels: map[string]string{
			ext.InterlockNetworkLabel: "foo",
		},
	}

	if _, ok := OverlayEnabled(cfg.Labels); !ok {
		t.Fatal("expected to use overlay networking")
	}
}

func TestUseOverlayNoLabel(t *testing.T) {
	cfg := types.Container{
		Labels: map[string]string{},
	}

	if _, ok := OverlayEnabled(cfg.Labels); ok {
		t.Fatal("expected to use bridge networking")
	}
}

func TestBackendOverlayAddress(t *testing.T) {
	containerID := "a5bb3cae92fb660eba775831d7fec0227d980ce04c138b0b8e5d69885a82d75f"

	network := types.NetworkResource{
		Name: "testNetwork",
		ID:   "testNetwork",
		Containers: map[string]types.EndpointResource{
			containerID: types.EndpointResource{
				IPv4Address: "1.2.3.4/32",
			},
		},
	}

	cnt := types.Container{
		ID:     containerID,
		Labels: map[string]string{},
		Ports: []types.Port{
			{
				IP:          "0.0.0.0",
				PrivatePort: 80,
				PublicPort:  32768,
			},
		},
	}

	addr, err := BackendOverlayAddress(network, cnt)
	if err != nil {
		t.Fatal(err)
	}

	expected := "1.2.3.4:32768"
	if addr != expected {
		t.Fatalf("expected %s; received %s", expected, addr)
	}
}

func TestBackendAddress(t *testing.T) {
	containerID := "a5bb3cae92fb660eba775831d7fec0227d980ce04c138b0b8e5d69885a82d75f"

	cnt := types.Container{
		ID:     containerID,
		Labels: map[string]string{},
		Ports: []types.Port{
			{
				IP:          "0.0.0.0",
				PrivatePort: 80,
				PublicPort:  32768,
			},
		},
	}

	addr, err := BackendAddress(cnt, "")
	if err != nil {
		t.Fatal(err)
	}

	expected := "0.0.0.0:32768"
	if addr != expected {
		t.Fatalf("expected %s; received %s", expected, addr)
	}
}

func TestBackendAddressOverride(t *testing.T) {
	containerID := "a5bb3cae92fb660eba775831d7fec0227d980ce04c138b0b8e5d69885a82d75f"

	cnt := types.Container{
		ID:     containerID,
		Labels: map[string]string{},
		Ports: []types.Port{
			{
				IP:          "0.0.0.0",
				PrivatePort: 80,
				PublicPort:  32768,
			},
		},
	}

	addr, err := BackendAddress(cnt, "1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}

	expected := "1.1.1.1:32768"
	if addr != expected {
		t.Fatalf("expected %s; received %s", expected, addr)
	}
}
