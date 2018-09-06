// Copyright (c) 2018 Uber Technologies, Inc.
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

package panichandler

import (
	"context"

	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/tchannel/baz/module"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/tchannel/baz/workflow"
	clientBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	endpointBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/tchannel/baz/baz"
	"github.com/uber/zanzibar/runtime"

	"go.uber.org/zap"
)

// NewSimpleServiceCallWorkflow ...
func NewSimpleServiceCallWorkflow(
	deps *module.Dependencies,
) workflow.SimpleServiceCallWorkflow {
	return &Workflow{
		Clients: deps.Client,
		Logger:  deps.Default.Logger,
	}
}

// Workflow ...
type Workflow struct {
	Clients *module.ClientDependencies
	Logger  *zap.Logger
}

// Handle ...
func (w Workflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	req *endpointBaz.SimpleService_Call_Args,
) (zanzibar.Header, error) {
	clientReqHeaders := make(map[string]string)
	// passing endpoint reqHeaders to downstream client
	for _, k := range reqHeaders.Keys() {
		if v, ok := reqHeaders.Get(k); ok {
			clientReqHeaders[k] = v
		}
	}

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
			zanzibar.LogErrorWarnTimeout(
				w.Logger, err,
				"baz.Call returned error",
			)
			return respHeaders, err
		}
	}

	if val, ok := clientReqHeaders["x-nil-response-header"]; ok {
		if "true" == val {
			return nil, nil
		}
	}

	return respHeaders, nil
}
