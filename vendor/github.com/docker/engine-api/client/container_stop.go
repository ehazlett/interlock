package client

import (
	"net/url"
	"time"

	timetypes "github.com/docker/engine-api/types/time"
	"golang.org/x/net/context"
)

// ContainerStop stops a container without terminating the process.
// The process is blocked until the container stops or the timeout expires.
<<<<<<< HEAD
<<<<<<< HEAD
func (cli *Client) ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error {
	query := url.Values{}
	if timeout != nil {
		query.Set("t", timetypes.DurationToSecondsString(*timeout))
	}
=======
func (cli *Client) ContainerStop(ctx context.Context, containerID string, timeout time.Duration) error {
	query := url.Values{}
	query.Set("t", timetypes.DurationToSecondsString(timeout))
>>>>>>> c73b1ae... switch to engine-api; update beacon to be more efficient
=======
func (cli *Client) ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error {
	query := url.Values{}
	if timeout != nil {
		query.Set("t", timetypes.DurationToSecondsString(*timeout))
	}
>>>>>>> 12a5469... start on swarm services; move to glade
	resp, err := cli.post(ctx, "/containers/"+containerID+"/stop", query, nil, nil)
	ensureReaderClosed(resp)
	return err
}
