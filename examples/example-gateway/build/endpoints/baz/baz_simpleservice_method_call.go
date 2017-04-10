// Code generated by zanzibar
// @generated

package baz

import (
	"context"

	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"

	githubComUberZanzibarEndpointsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/endpoints/baz/baz"
	customBaz "github.com/uber/zanzibar/examples/example-gateway/endpoints/baz"
)

// HandleCallRequest handles "/baz/call-path".
func HandleCallRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
	clients *clients.Clients,
) {
	var requestBody githubComUberZanzibarEndpointsBazBaz.BazRequest
	if ok := req.ReadAndUnmarshalBody(&requestBody); !ok {
		return
	}

	headers := map[string]string{}

	workflow := customBaz.CallEndpoint{
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
