// Code generated by zanzibar
// @generated

package bar

import (
	"context"

	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"

	githubComUberZanzibarClientsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/clients/bar/bar"
	githubComUberZanzibarClientsFooFoo "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/clients/foo/foo"
	githubComUberZanzibarEndpointsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/endpoints/bar/bar"
	"github.com/uber/zanzibar/examples/example-gateway/build/github.com/uber/zanzibar/clients/bar"
)

// HandleTooManyArgsRequest handles "/bar/too-many-args-path".
func HandleTooManyArgsRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
	clients *clients.Clients,
) {
	if !req.CheckHeaders([]string{"x-uuid", "x-token"}) {
		return
	}
	var requestBody TooManyArgsHTTPRequest
	if ok := req.ReadAndUnmarshalBody(&requestBody); !ok {
		return
	}

	headers := map[string]string{}

	workflow := TooManyArgsEndpoint{
		Clients: clients,
		Logger:  req.Logger,
		Request: req,
	}

	response, _, err := workflow.Handle(ctx, headers, &requestBody)
	if err != nil {
		req.Logger.Warn("Workflow for endpoint returned error",
			zap.String("error", err.Error()),
		)
		res.SendErrorString(500, "Unexpected server error")
		return
	}

	res.WriteJSON(200, response)
}

// TooManyArgsEndpoint calls thrift client Bar.TooManyArgs
type TooManyArgsEndpoint struct {
	Clients *clients.Clients
	Logger  *zap.Logger
	Request *zanzibar.ServerHTTPRequest
}

// Handle calls thrift client.
func (w TooManyArgsEndpoint) Handle(
	ctx context.Context,
	headers map[string]string,
	r *TooManyArgsHTTPRequest,
) (*githubComUberZanzibarEndpointsBarBar.BarResponse, map[string]string, error) {
	clientRequest := convertToTooManyArgsClientRequest(r)

	clientRespBody, _, err := w.Clients.Bar.TooManyArgs(
		ctx, nil, clientRequest,
	)
	if err != nil {
		w.Logger.Warn("Could not make client request",
			zap.String("error", err.Error()),
		)
		return nil, nil, err
	}

	response := convertTooManyArgsClientResponse(clientRespBody)
	return response, nil, nil
}

func convertToTooManyArgsClientRequest(body *TooManyArgsHTTPRequest) *barClient.TooManyArgsHTTPRequest {
	clientRequest := &barClient.TooManyArgsHTTPRequest{}

	clientRequest.Foo = (*githubComUberZanzibarClientsFooFoo.FooStruct)(body.Foo)
	clientRequest.Request = (*githubComUberZanzibarClientsBarBar.BarRequest)(body.Request)

	return clientRequest
}
func convertTooManyArgsClientResponse(body *githubComUberZanzibarClientsBarBar.BarResponse) *githubComUberZanzibarEndpointsBarBar.BarResponse {
	// TODO: Add response fields mapping here.
	downstreamResponse := &githubComUberZanzibarEndpointsBarBar.BarResponse{}
	return downstreamResponse
}
