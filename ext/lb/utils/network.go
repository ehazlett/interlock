package utils

import (
	"fmt"
	"net"
	"strings"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func OverlayEnabled(config *dockerclient.ContainerConfig) (string, bool) {
	if v, ok := config.Labels[ext.InterlockNetworkLabel]; ok {
		return v, true
	}

	return "", false
}

func BackendOverlayAddress(network *dockerclient.NetworkResource, containerInfo *dockerclient.ContainerInfo) (string, error) {
	c, exists := network.Containers[containerInfo.Id]
	if !exists {
		return "", fmt.Errorf("container %s is not connected to network %s", containerInfo.Id, network.Name)
	}

	ip, _, err := net.ParseCIDR(c.IPv4Address)
	if err != nil {
		return "", err
	}

	ports := containerInfo.NetworkSettings.Ports
	portDef := dockerclient.PortBinding{}
	addr := ""

	portDef.HostIp = ip.String()

	// parse the port
	for k, _ := range ports {
		if k != "" {
			portParts := strings.Split(k, "/")
			portDef.HostPort = portParts[0]
			break
		}
	}

	// check for custom port
	if v, ok := containerInfo.Config.Labels[ext.InterlockPortLabel]; ok {
		portDef.HostPort = v
	}

	addr = fmt.Sprintf("%s:%s", portDef.HostIp, portDef.HostPort)

	return addr, nil
}

func BackendAddress(containerInfo *dockerclient.ContainerInfo, backendOverrideAddress string) (string, error) {
	ports := containerInfo.NetworkSettings.Ports
	portDef := dockerclient.PortBinding{}
	addr := ""

	// parse the published port
	for _, v := range ports {
		if len(v) > 0 {
			portDef.HostIp = v[0].HostIp
			portDef.HostPort = v[0].HostPort
			break
		}
	}

	if backendOverrideAddress != "" {
		portDef.HostIp = backendOverrideAddress
	}

	// check for custom port
	if v, ok := containerInfo.Config.Labels[ext.InterlockPortLabel]; ok {
		interlockPort := v
		for k, x := range ports {
			parts := strings.Split(k, "/")
			if parts[0] == interlockPort {
				port := x[0]
				portDef.HostPort = port.HostPort
				break
			}
		}
	}

	addr = fmt.Sprintf("%s:%s", portDef.HostIp, portDef.HostPort)
	return addr, nil
}
