package utils

import (
	"testing"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func TestUseOverlay(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			ext.InterlockNetworkLabel: "foo",
		},
	}

	if _, ok := OverlayEnabled(cfg); !ok {
		t.Fatal("expected to use overlay networking")
	}
}

func TestUseOverlayNoLabel(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{},
	}

	if _, ok := OverlayEnabled(cfg); ok {
		t.Fatal("expected to use bridge networking")
	}
}

func TestBackendOverlayAddress(t *testing.T) {
	containerID := "a5bb3cae92fb660eba775831d7fec0227d980ce04c138b0b8e5d69885a82d75f"

	network := &dockerclient.NetworkResource{
		Name: "testNetwork",
		ID:   "testNetwork",
		Containers: map[string]dockerclient.EndpointResource{
			containerID: dockerclient.EndpointResource{
				IPv4Address: "1.2.3.4/32",
			},
		},
	}

	containerInfo := &dockerclient.ContainerInfo{
		Id: containerID,
		Config: &dockerclient.ContainerConfig{
			Labels: map[string]string{},
		},
		NetworkSettings: struct {
			IPAddress   string `json:"IpAddress"`
			IPPrefixLen int    `json:"IpPrefixLen"`
			Gateway     string
			Bridge      string
			Ports       map[string][]dockerclient.PortBinding
			Networks    map[string]*dockerclient.EndpointSettings
		}{
			IPAddress:   "",
			IPPrefixLen: 0,
			Gateway:     "",
			Bridge:      "",
			Ports: map[string][]dockerclient.PortBinding{
				"80/tcp": nil,
			},
			Networks: map[string]*dockerclient.EndpointSettings{
				"": nil,
			},
		},
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
	containerInfo := &dockerclient.ContainerInfo{
		Config: &dockerclient.ContainerConfig{
			Labels: map[string]string{},
		},
		NetworkSettings: struct {
			IPAddress   string `json:"IpAddress"`
			IPPrefixLen int    `json:"IpPrefixLen"`
			Gateway     string
			Bridge      string
			Ports       map[string][]dockerclient.PortBinding
			Networks    map[string]*dockerclient.EndpointSettings
		}{
			IPAddress:   "",
			IPPrefixLen: 0,
			Gateway:     "",
			Bridge:      "",
			Ports: map[string][]dockerclient.PortBinding{
				"80/tcp": {
					dockerclient.PortBinding{
						HostIp:   "0.0.0.0",
						HostPort: "80",
					},
				},
			},
			Networks: map[string]*dockerclient.EndpointSettings{
				"": nil,
			},
		},
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
	containerInfo := &dockerclient.ContainerInfo{
		Config: &dockerclient.ContainerConfig{
			Labels: map[string]string{},
		},
		NetworkSettings: struct {
			IPAddress   string `json:"IpAddress"`
			IPPrefixLen int    `json:"IpPrefixLen"`
			Gateway     string
			Bridge      string
			Ports       map[string][]dockerclient.PortBinding
			Networks    map[string]*dockerclient.EndpointSettings
		}{
			IPAddress:   "",
			IPPrefixLen: 0,
			Gateway:     "",
			Bridge:      "",
			Ports: map[string][]dockerclient.PortBinding{
				"80/tcp": {
					dockerclient.PortBinding{
						HostIp:   "0.0.0.0",
						HostPort: "80",
					},
				},
			},
			Networks: map[string]*dockerclient.EndpointSettings{
				"": nil,
			},
		},
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
