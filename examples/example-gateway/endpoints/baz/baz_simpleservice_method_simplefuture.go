// TODO: (lu) to be generated

package baz

import (
	"context"
	"net/http"

	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"

	bazClientStructs "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
)

// SimpleFutureEndpoint is the simple future handler struct
type SimpleFutureEndpoint struct {
	Clients *clients.Clients
	Logger  *zap.Logger
	Request *zanzibar.ServerHTTPRequest
}

// Handle "/baz/simple-future-path".
func (e SimpleFutureEndpoint) Handle(
	ctx context.Context,
	headers http.Header,
) (map[string]string, error) {

	clientHeaders := map[string]string{}
	for k := range headers {
		clientHeaders[k] = headers.Get(k)
	}

	cRespHeaders, err := e.Clients.Baz.SimpleFuture(ctx, clientHeaders)
	if err != nil {
		// TODO: (lu) error type handling
		switch err.(type) {
		case *bazClientStructs.SimpleErr:
			return nil, err
		case *bazClientStructs.NewErr:
			return nil, err
		default:
			return nil, err
		}
	}

	return cRespHeaders, nil
}
