// +build integration
package nginx

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

func getDockerClient() (*client.Client, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func TestNginxBasic(t *testing.T) {
	cli, err := getDockerClient()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image: "busybox",
		Cmd:   []string{"sh"},
		Tty:   true,
		Labels: map[string]string{
			"interlock.hostname": "test.local",
		},
	}, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}

	id := resp.ID
	defer func() {
		if err := cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			RemoveLinks:   false,
			Force:         true,
		}); err != nil {
			t.Error(err)
		}
	}()
}
