// Code generated by zanzibar
// @generated

package bar

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uber-go/zap"
	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	zanzibar "github.com/uber/zanzibar/runtime"

	"github.com/uber/zanzibar/examples/example-gateway/build/clients/bar"
	endpointsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/endpoints/bar/bar"

	clientsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/clients/bar/bar"
)

// HandleNormalRequest handles "/bar/bar-path".
func HandleNormalRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
	clients *clients.Clients,
) {

	var body NormalHTTPRequest
	if ok := req.ReadAndUnmarshalBody(&body); !ok {
		return
	}
	clientRequest := convertToNormalClientRequest(&body)

	clientRespBody, _, err := clients.Bar.Normal(
		ctx, nil, clientRequest,
	)

	if err != nil {
		req.Logger.Warn("Could not make client request",
			zap.String("error", err.Error()),
		)
		res.SendError(500, errors.Wrap(err, "could not make client request:"))
		return
	}

	response := convertNormalClientResponse(clientRespBody)
	res.WriteJSON(200, response)
}

func convertToNormalClientRequest(body *NormalHTTPRequest) *barClient.NormalHTTPRequest {
	clientRequest := &barClient.NormalHTTPRequest{}

	clientRequest.Request = (*clientsBarBar.BarRequest)(body.Request)

	return clientRequest
}
func convertNormalClientResponse(body *clientsBarBar.BarResponse) *endpointsBarBar.BarResponse {
	// TODO: Add response fields mapping here.
	downstreamResponse := &endpointsBarBar.BarResponse{}
	return downstreamResponse
}
