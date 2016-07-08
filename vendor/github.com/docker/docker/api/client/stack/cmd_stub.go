// +build !experimental

package stack

import (
	"github.com/docker/docker/api/client"
	"github.com/spf13/cobra"
)

<<<<<<< HEAD
// NewStackCommand returns nocommand
=======
// NewStackCommand returns no command
>>>>>>> 12a5469... start on swarm services; move to glade
func NewStackCommand(dockerCli *client.DockerCli) *cobra.Command {
	return &cobra.Command{}
}

<<<<<<< HEAD
// NewTopLevelDeployCommand return no command
=======
// NewTopLevelDeployCommand returns no command
>>>>>>> 12a5469... start on swarm services; move to glade
func NewTopLevelDeployCommand(dockerCli *client.DockerCli) *cobra.Command {
	return &cobra.Command{}
}
