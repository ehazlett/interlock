package utils

import (
	"testing"

	"github.com/docker/engine-api/types"
	ctypes "github.com/docker/engine-api/types/container"
	ntypes "github.com/docker/engine-api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/ehazlett/interlock/ext"
)

func TestUseOverlay(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{
			ext.InterlockNetworkLabel: "foo",
		},
	}

	if _, ok := OverlayEnabled(cfg.Labels); !ok {
		t.Fatal("expected to use overlay networking")
	}
}

func TestUseOverlayNoLabel(t *testing.T) {
	cfg := &ctypes.Config{
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

	b := &types.ContainerJSONBase{
		ID: containerID,
	}
	mounts := []types.MountPoint{}
	cfg := &ctypes.Config{
		Labels: map[string]string{},
	}
	netSettings := &types.NetworkSettings{}
	netSettings.IPAddress = ""
	netSettings.IPPrefixLen = 0
	netSettings.Gateway = ""
	netSettings.Bridge = ""
	netSettings.Ports = nat.PortMap{
		"80/tcp": []nat.PortBinding{
			{
				HostIP:   "",
				HostPort: "80",
			},
		},
	}
	netSettings.Networks = map[string]*ntypes.EndpointSettings{
		"": nil,
	}

	containerInfo := types.ContainerJSON{
		b,
		mounts,
		cfg,
		netSettings,
	}

	addr, err := BackendOverlayAddress(network, containerInfo)
	if err != nil {
		t.Fatal(err)
	}

	expected := "1.2.3.4:80"
	if addr != expected {
		t.Fatalf("expected %s; received %s", expected, addr)
	}
}

func TestBackendAddress(t *testing.T) {
	containerID := "a5bb3cae92fb660eba775831d7fec0227d980ce04c138b0b8e5d69885a82d75f"

	b := &types.ContainerJSONBase{
		ID: containerID,
	}
	mounts := []types.MountPoint{}
	cfg := &ctypes.Config{
		Labels: map[string]string{},
	}
	netSettings := &types.NetworkSettings{}
	netSettings.IPAddress = ""
	netSettings.IPPrefixLen = 0
	netSettings.Gateway = ""
	netSettings.Bridge = ""
	netSettings.Ports = nat.PortMap{
		"80/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "80",
			},
		},
	}
	netSettings.Networks = map[string]*ntypes.EndpointSettings{
		"": nil,
	}

	containerInfo := types.ContainerJSON{
		b,
		mounts,
		cfg,
		netSettings,
	}

	addr, err := BackendAddress(containerInfo, "")
	if err != nil {
		t.Fatal(err)
	}

	expected := "0.0.0.0:80"
	if addr != expected {
		t.Fatalf("expected %s; received %s", expected, addr)
	}
}

func TestBackendAddressOverride(t *testing.T) {
	containerID := "a5bb3cae92fb660eba775831d7fec0227d980ce04c138b0b8e5d69885a82d75f"

	b := &types.ContainerJSONBase{
		ID: containerID,
	}
	mounts := []types.MountPoint{}
	cfg := &ctypes.Config{
		Labels: map[string]string{},
	}
	netSettings := &types.NetworkSettings{}
	netSettings.IPAddress = ""
	netSettings.IPPrefixLen = 0
	netSettings.Gateway = ""
	netSettings.Bridge = ""
	netSettings.Ports = nat.PortMap{
		"80/tcp": []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "80",
			},
		},
	}
	netSettings.Networks = map[string]*ntypes.EndpointSettings{
		"": nil,
	}

	containerInfo := types.ContainerJSON{
		b,
		mounts,
		cfg,
		netSettings,
	}

	addr, err := BackendAddress(containerInfo, "1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}

	expected := "1.1.1.1:80"
	if addr != expected {
		t.Fatalf("expected %s; received %s", expected, addr)
	}
}
