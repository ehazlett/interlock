package utils

import (
	"fmt"
	"net"
	"strconv"

	"github.com/docker/docker/api/types"
	ctypes "github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/ehazlett/interlock/ext"
)

func OverlayEnabled(config ctypes.Container) (string, bool) {
	if v, ok := config.Labels[ext.InterlockNetworkLabel]; ok {
		return v, true
	}

	return "", false
}

func BackendOverlayAddress(network types.NetworkResource, cnt types.Container) (string, error) {
	c, exists := network.Containers[cnt.ID]
	if !exists {
		return "", fmt.Errorf("container %s is not connected to network %s", cnt.ID, network.Name)
	}

	ip, _, err := net.ParseCIDR(c.IPv4Address)
	if err != nil {
		return "", err
	}

	ports := cnt.Ports
	portDef := nat.PortBinding{}
	addr := ""

	portDef.HostIP = ip.String()

	// parse the port
	for _, k := range ports {
		if k.PublicPort != 0 {
			portDef.HostPort = fmt.Sprintf("%d", k.PublicPort)
			break
		}
	}

	// check for custom port
	if v, ok := cnt.Labels[ext.InterlockPortLabel]; ok {
		portDef.HostPort = v
	}

	if portDef.HostPort == "" {
		return "", fmt.Errorf("unable to find exposed port")
	}

	addr = fmt.Sprintf("%s:%s", portDef.HostIP, portDef.HostPort)

	return addr, nil
}

func BackendAddress(cnt types.Container, backendOverrideAddress string) (string, error) {
	ports := cnt.Ports
	portDef := nat.PortBinding{}
	addr := ""

	// parse the published port
	for _, port := range ports {
		portDef.HostIP = port.IP
		portDef.HostPort = fmt.Sprintf("%d", port.PublicPort)
		break
	}

	if backendOverrideAddress != "" {
		portDef.HostIP = backendOverrideAddress
	}

	// check for custom port
	if v, ok := cnt.Labels[ext.InterlockPortLabel]; ok {
		interlockPort, err := strconv.Atoi(v)
		if err != nil {
			return "", err
		}
		for _, port := range ports {
			if port.PrivatePort == uint16(interlockPort) {
				portDef.HostIP = port.IP
				portDef.HostPort = fmt.Sprintf("%d", port.PublicPort)
				break
			}
		}
	}

	if portDef.HostIP == "" || portDef.HostPort == "" {
		return "", fmt.Errorf("unable to find ip or exposed port: %+v", portDef)
	}

	addr = fmt.Sprintf("%s:%s", portDef.HostIP, portDef.HostPort)
	return addr, nil
}
