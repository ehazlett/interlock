package client

import (
	"net/url"
	"strconv"

<<<<<<< HEAD
=======
	"github.com/docker/engine-api/types"
>>>>>>> 12a5469... start on swarm services; move to glade
	"github.com/docker/engine-api/types/swarm"
	"golang.org/x/net/context"
)

// ServiceUpdate updates a Service.
<<<<<<< HEAD
func (cli *Client) ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec) error {
	query := url.Values{}
	query.Set("version", strconv.FormatUint(version.Index, 10))
	resp, err := cli.post(ctx, "/services/"+serviceID+"/update", query, service, nil)
=======
func (cli *Client) ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) error {
	var (
		headers map[string][]string
		query   = url.Values{}
	)

	if options.EncodedRegistryAuth != "" {
		headers = map[string][]string{
			"X-Registry-Auth": []string{options.EncodedRegistryAuth},
		}
	}

	query.Set("version", strconv.FormatUint(version.Index, 10))

	resp, err := cli.post(ctx, "/services/"+serviceID+"/update", query, service, headers)
>>>>>>> 12a5469... start on swarm services; move to glade
	ensureReaderClosed(resp)
	return err
}
