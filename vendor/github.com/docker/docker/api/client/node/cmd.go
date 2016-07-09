package node

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/spf13/cobra"

	"github.com/docker/docker/api/client"
	"github.com/docker/docker/cli"
	apiclient "github.com/docker/engine-api/client"
)

// NewNodeCommand returns a cobra command for `node` subcommands
func NewNodeCommand(dockerCli *client.DockerCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Manage Docker Swarm nodes",
		Args:  cli.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(dockerCli.Err(), "\n"+cmd.UsageString())
		},
	}
	cmd.AddCommand(
		newAcceptCommand(dockerCli),
		newDemoteCommand(dockerCli),
		newInspectCommand(dockerCli),
		newListCommand(dockerCli),
		newPromoteCommand(dockerCli),
		newRemoveCommand(dockerCli),
		newTasksCommand(dockerCli),
		newUpdateCommand(dockerCli),
	)
	return cmd
}

<<<<<<< HEAD
func nodeReference(client apiclient.APIClient, ctx context.Context, ref string) (string, error) {
	// The special value "self" for a node reference is mapped to the current
	// node, hence the node ID is retrieved using the `/info` endpoint.
=======
// Reference return the reference of a node. The special value "self" for a node
// reference is mapped to the current node, hence the node ID is retrieved using
// the `/info` endpoint.
func Reference(client apiclient.APIClient, ctx context.Context, ref string) (string, error) {
>>>>>>> 12a5469... start on swarm services; move to glade
	if ref == "self" {
		info, err := client.Info(ctx)
		if err != nil {
			return "", err
		}
		return info.Swarm.NodeID, nil
	}
	return ref, nil
}