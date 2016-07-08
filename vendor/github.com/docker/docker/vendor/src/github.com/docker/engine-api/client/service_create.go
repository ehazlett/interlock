package client

import (
	"encoding/json"

	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/swarm"
	"golang.org/x/net/context"
)

// ServiceCreate creates a new Service.
<<<<<<< HEAD
func (cli *Client) ServiceCreate(ctx context.Context, service swarm.ServiceSpec) (types.ServiceCreateResponse, error) {
	var response types.ServiceCreateResponse
	resp, err := cli.post(ctx, "/services/create", nil, service, nil)
=======
func (cli *Client) ServiceCreate(ctx context.Context, service swarm.ServiceSpec, options types.ServiceCreateOptions) (types.ServiceCreateResponse, error) {
	var headers map[string][]string

	if options.EncodedRegistryAuth != "" {
		headers = map[string][]string{
			"X-Registry-Auth": []string{options.EncodedRegistryAuth},
		}
	}

	var response types.ServiceCreateResponse
	resp, err := cli.post(ctx, "/services/create", nil, service, headers)
>>>>>>> 12a5469... start on swarm services; move to glade
	if err != nil {
		return response, err
	}

	err = json.NewDecoder(resp.body).Decode(&response)
	ensureReaderClosed(resp)
	return response, err
}
