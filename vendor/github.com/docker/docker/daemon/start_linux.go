package daemon

import (
	"fmt"

	"github.com/docker/docker/container"
	"github.com/docker/docker/libcontainerd"
<<<<<<< HEAD
	"github.com/docker/engine-api/types"
=======
>>>>>>> 12a5469... start on swarm services; move to glade
)

func (daemon *Daemon) getLibcontainerdCreateOptions(container *container.Container) (*[]libcontainerd.CreateOption, error) {
	createOptions := []libcontainerd.CreateOption{}

	// Ensure a runtime has been assigned to this container
	if container.HostConfig.Runtime == "" {
<<<<<<< HEAD
		container.HostConfig.Runtime = types.DefaultRuntimeName
=======
		container.HostConfig.Runtime = stockRuntimeName
>>>>>>> 12a5469... start on swarm services; move to glade
		container.ToDisk()
	}

	rt := daemon.configStore.GetRuntime(container.HostConfig.Runtime)
	if rt == nil {
		return nil, fmt.Errorf("no such runtime '%s'", container.HostConfig.Runtime)
	}
	createOptions = append(createOptions, libcontainerd.WithRuntime(rt.Path, rt.Args))

	return &createOptions, nil
}
