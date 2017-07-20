// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package bazHandler

import (
	"context"

	"github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"

	clientBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	endpointBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/tchannel/baz/baz"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/tchannel/baz/module"
)

// CallEndpoint ...
type CallEndpoint struct {
	Clients *module.ClientDependencies
	Logger  *zap.Logger
}

// Handle ...
func (w CallEndpoint) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	req *endpointBaz.SimpleService_Call_Args,
) (zanzibar.Header, error) {
	var clientReqHeaders map[string]string

	clientReq := &clientBaz.SimpleService_Call_Args{
		Arg: (*clientBaz.BazRequest)(req.Arg),
	}

	clientRespHeaders, err := w.Clients.Baz.Call(ctx, clientReqHeaders, clientReq)
	respHeaders := zanzibar.ServerTChannelHeader(clientRespHeaders)
	if err != nil {
		switch v := err.(type) {
		case *clientBaz.AuthErr:
			return respHeaders, (*endpointBaz.AuthErr)(v)
		default:
			w.Logger.Error(
				"baz.Call returned unexpected error",
				zap.String("error", err.Error()),
			)
			return respHeaders, err
		}
	}
	return respHeaders, nil
}
