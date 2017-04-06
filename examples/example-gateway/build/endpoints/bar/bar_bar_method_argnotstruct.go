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
)

// HandleArgNotStructRequest handles "/bar/arg-not-struct-path".
func HandleArgNotStructRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
	clients *clients.Clients,
) {

	var body ArgNotStructHTTPRequest
	if ok := req.ReadAndUnmarshalBody(&body); !ok {
		return
	}
	clientRequest := convertToArgNotStructClientRequest(&body)

	_, err := clients.Bar.ArgNotStruct(
		ctx, nil, clientRequest,
	)

	if err != nil {
		req.Logger.Warn("Could not make client request",
			zap.String("error", err.Error()),
		)
		res.SendError(500, errors.Wrap(err, "could not make client request:"))
		return
	}

	res.WriteJSONBytes(200, nil)
}

func convertToArgNotStructClientRequest(body *ArgNotStructHTTPRequest) *barClient.ArgNotStructHTTPRequest {
	clientRequest := &barClient.ArgNotStructHTTPRequest{}

	clientRequest.Request = string(body.Request)

	return clientRequest
}
