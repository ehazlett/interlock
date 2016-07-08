package client

import (
<<<<<<< HEAD
	"encoding/json"
=======
	"bytes"
	"encoding/json"
	"io/ioutil"
>>>>>>> 12a5469... start on swarm services; move to glade
	"net/http"

	"github.com/docker/engine-api/types/swarm"
	"golang.org/x/net/context"
)

<<<<<<< HEAD
// NodeInspect returns the node information.
func (cli *Client) NodeInspect(ctx context.Context, nodeID string) (swarm.Node, error) {
	serverResp, err := cli.get(ctx, "/nodes/"+nodeID, nil, nil)
	if err != nil {
		if serverResp.statusCode == http.StatusNotFound {
			return swarm.Node{}, nodeNotFoundError{nodeID}
		}
		return swarm.Node{}, err
	}

	var response swarm.Node
	err = json.NewDecoder(serverResp.body).Decode(&response)
	ensureReaderClosed(serverResp)
	return response, err
=======
// NodeInspectWithRaw returns the node information.
func (cli *Client) NodeInspectWithRaw(ctx context.Context, nodeID string) (swarm.Node, []byte, error) {
	serverResp, err := cli.get(ctx, "/nodes/"+nodeID, nil, nil)
	if err != nil {
		if serverResp.statusCode == http.StatusNotFound {
			return swarm.Node{}, nil, nodeNotFoundError{nodeID}
		}
		return swarm.Node{}, nil, err
	}
	defer ensureReaderClosed(serverResp)

	body, err := ioutil.ReadAll(serverResp.body)
	if err != nil {
		return swarm.Node{}, nil, err
	}

	var response swarm.Node
	rdr := bytes.NewReader(body)
	err = json.NewDecoder(rdr).Decode(&response)
	return response, body, err
>>>>>>> 12a5469... start on swarm services; move to glade
}
