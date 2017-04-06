// TODO: (lu) to be generated

package baz

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uber-go/zap"
	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	zanzibar "github.com/uber/zanzibar/runtime"

	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/clients/baz/baz"
)

// HandleSimpleRequest handles "/baz/simple-path".
func HandleSimpleRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
	clients *clients.Clients,
) {
	// Handle request headers.
	//h := http.Header{}
	var clientReqHeaders map[string]string

	// TODO: (lu) deal response headers
	_, err := clients.Baz.Simple(ctx, clientReqHeaders)
	if err != nil {
		switch err.(type) {
		case *baz.SimpleErr:
			res.SendError(400, errors.Wrap(err, "simple error"))
		default:
			req.Logger.Error("Client request returned error",
				zap.String("error", err.Error()),
			)
			res.SendError(500, errors.Wrap(err, "could not make client request:"))
		}
		return
	}

	// TODO: (lu) convert rpc response to http response status code
	statusCode := 200
	res.WriteJSONBytes(statusCode, nil)
}
