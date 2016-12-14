package utils

import (
	"strings"

	"github.com/ehazlett/interlock/ext"
)

func WebsocketEndpoints(labels map[string]string) []string {
	websocketEndpoints := []string{}

	for l, v := range labels {
		// this is for labels like interlock.websocket_endpoint.1=foo
		if strings.Index(l, ext.InterlockWebsocketEndpointLabel) > -1 {
			websocketEndpoints = append(websocketEndpoints, v)
		}
	}

	return websocketEndpoints
}
