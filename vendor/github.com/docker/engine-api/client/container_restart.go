package client

import (
	"net/url"
	"time"

	timetypes "github.com/docker/engine-api/types/time"
	"golang.org/x/net/context"
)

// ContainerRestart stops and starts a container again.
// It makes the daemon to wait for the container to be up again for
// a specific amount of time, given the timeout.
<<<<<<< HEAD
func (cli *Client) ContainerRestart(ctx context.Context, containerID string, timeout *time.Duration) error {
	query := url.Values{}
	if timeout != nil {
		query.Set("t", timetypes.DurationToSecondsString(*timeout))
	}
=======
func (cli *Client) ContainerRestart(ctx context.Context, containerID string, timeout time.Duration) error {
	query := url.Values{}
	query.Set("t", timetypes.DurationToSecondsString(timeout))
>>>>>>> c73b1ae... switch to engine-api; update beacon to be more efficient
	resp, err := cli.post(ctx, "/containers/"+containerID+"/restart", query, nil, nil)
	ensureReaderClosed(resp)
	return err
}
