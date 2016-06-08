package utils

import (
	"fmt"
	"net"
	"strings"

	"github.com/docker/engine-api/types"
	ctypes "github.com/docker/engine-api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/ehazlett/interlock/ext"
)

func OverlayEnabled(config *ctypes.Config) (string, bool) {
	if v, ok := config.Labels[ext.InterlockNetworkLabel]; ok {
		return v, true
	}

	return "", false
}

func BackendOverlayAddress(network types.NetworkResource, containerInfo types.ContainerJSON) (string, error) {
	c, exists := network.Containers[containerInfo.ID]
	if !exists {
		return "", fmt.Errorf("container %s is not connected to network %s", containerInfo.ID, network.Name)
	}

	ip, _, err := net.ParseCIDR(c.IPv4Address)
	if err != nil {
		return "", err
	}

	ports := containerInfo.NetworkSettings.Ports
	portDef := nat.PortBinding{}
	addr := ""

	portDef.HostIP = ip.String()

	// parse the port
	for k, _ := range ports {
		if k != "" {
			portParts := strings.Split(string(k), "/")
			portDef.HostPort = portParts[0]
			break
		}
	}

	// check for custom port
	if v, ok := containerInfo.Config.Labels[ext.InterlockPortLabel]; ok {
		portDef.HostPort = v
	}

	addr = fmt.Sprintf("%s:%s", portDef.HostIP, portDef.HostPort)

	return addr, nil
}

func BackendAddress(containerInfo types.ContainerJSON, backendOverrideAddress string) (string, error) {
	ports := containerInfo.NetworkSettings.Ports
	portDef := nat.PortBinding{}
	addr := ""

	// parse the published port
	for _, v := range ports {
		if len(v) > 0 {
			portDef.HostIP = v[0].HostIP
			portDef.HostPort = v[0].HostPort
			break
		}
	}

	if backendOverrideAddress != "" {
		portDef.HostIP = backendOverrideAddress
	}

	// check for custom port
	if v, ok := containerInfo.Config.Labels[ext.InterlockPortLabel]; ok {
		interlockPort := v
		for k, x := range ports {
			parts := strings.Split(string(k), "/")
			if parts[0] == interlockPort {
				port := x[0]
				portDef.HostPort = port.HostPort
				break
			}
		}
	}

	addr = fmt.Sprintf("%s:%s", portDef.HostIP, portDef.HostPort)
	return addr, nil
}
