// TODO: (lu) to be generated

package baz

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uber-go/zap"
	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	zanzibar "github.com/uber/zanzibar/runtime"

	clientTypeBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/clients/baz/baz"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/endpoints/baz/baz"
)

// HandleCallRequest handles "/baz/call-path".
func HandleCallRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
	clients *clients.Clients,
) {
	// Handle request headers.
	//h := http.Header{}
	var clientReqHeaders map[string]string

	// Handle request body.
	var body CallHTTPRequest
	if ok := req.ReadAndUnmarshalBody(&body); !ok {
		return
	}
	clientRequest := convertToCallClientRequest(&body)

	// TODO: (lu) deal response headers
	_, clientResp, err := clients.Baz.Call(ctx, clientReqHeaders, clientRequest)
	if err != nil {
		switch err.(type) {
		default:
			req.Logger.Error("Client request returned unexpected error",
				zap.String("error", err.Error()),
			)
			res.SendError(500, errors.Wrap(err, "could not make client request:"))
			return
		}
	}

	response := convertCallClientResponse(clientResp)

	// TODO: (lu) decide response status code
	statusCode := 200
	res.WriteJSON(statusCode, response)
}

func convertToCallClientRequest(body *CallHTTPRequest) *clientTypeBaz.SimpleService_Call_Args {
	clientRequest := &clientTypeBaz.SimpleService_Call_Args{}
	clientRequest.Arg = (*clientTypeBaz.BazRequest)(body.Arg)

	return clientRequest
}
func convertCallClientResponse(body *clientTypeBaz.BazResponse) *baz.BazResponse {
	response := (*baz.BazResponse)(body)
	return response
}
