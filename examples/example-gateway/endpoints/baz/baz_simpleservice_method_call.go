// TODO: (lu) to be generated

package baz

import (
	"context"

	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"

	bazClientStructs "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	endpointBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/baz/baz"
)

// CallEndpoint is the call handler struct
type CallEndpoint struct {
	Clients *clients.Clients
	Logger  *zap.Logger
	Request *zanzibar.ServerHTTPRequest
}

// Handle "/baz/call-path".
func (e CallEndpoint) Handle(
	ctx context.Context,
	headers map[string]string,
	r *endpointBaz.BazRequest,
) (*endpointBaz.BazResponse, map[string]string, error) {
	req := (*bazClientStructs.BazRequest)(r)

	cResp, cRespHeaders, err := e.Clients.Baz.Call(ctx, headers, req)
	if err != nil {
		switch err.(type) {
		default:
			e.Logger.Error("Client request returned unexpected error",
				zap.String("error", err.Error()),
			)
			return nil, nil, err
		}
	}

	resp := (*endpointBaz.BazResponse)(cResp)
	return resp, cRespHeaders, nil
}
