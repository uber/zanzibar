// TODO: (lu) to be generated

package baz

import (
	"context"

	"github.com/uber-go/zap"
	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	zanzibar "github.com/uber/zanzibar/runtime"

	bazClientStructs "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/clients/baz/baz"
)

// SimpleEndpoint is the simple handler struct
type SimpleEndpoint struct {
	Clients *clients.Clients
	Logger  zap.Logger
	Request *zanzibar.ServerHTTPRequest
}

// Handle "/baz/simple-path".
func (e SimpleEndpoint) Handle(
	ctx context.Context,
	headers map[string]string,
) (map[string]string, error) {
	cRespHeaders, err := e.Clients.Baz.Simple(ctx, headers)
	if err != nil {
		// TODO: (lu) error type handling
		switch err.(type) {
		case *bazClientStructs.SimpleErr:
			return nil, err
		default:
			return nil, err
		}
	}

	return cRespHeaders, nil
}

