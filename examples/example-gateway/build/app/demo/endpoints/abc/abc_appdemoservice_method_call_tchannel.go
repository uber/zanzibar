// Code generated by zanzibar
// @generated

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

package appdemoabcendpoint

import (
	"context"
	"runtime/debug"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/zap"

	zanzibar "github.com/uber/zanzibar/v2/runtime"

	customAbc "github.com/uber/zanzibar/v2/examples/example-gateway/app/demo/endpoints/abc"
	endpointsIDlEndpointsAppDemoEndpointsAbc "github.com/uber/zanzibar/v2/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/app/demo/endpoints/abc"

	defaultExampleTchannel "github.com/uber/zanzibar/v2/examples/example-gateway/middlewares/default/default_example_tchannel"

	module "github.com/uber/zanzibar/v2/examples/example-gateway/build/app/demo/endpoints/abc/module"
)

// NewAppDemoServiceCallHandler creates a handler to be registered with a thrift server.
func NewAppDemoServiceCallHandler(deps *module.Dependencies) *AppDemoServiceCallHandler {
	handler := &AppDemoServiceCallHandler{
		Deps: deps,
	}
	handler.endpoint = zanzibar.NewTChannelEndpoint(
		"appDemoAbc", "call", "AppDemoService::Call",
		zanzibar.NewTchannelStack([]zanzibar.MiddlewareTchannelHandle{
			deps.Middleware.DefaultExampleTchannel.NewMiddlewareHandle(
				defaultExampleTchannel.Options{},
			),
		}, handler),
	)

	return handler
}

// AppDemoServiceCallHandler is the handler for "AppDemoService::Call".
type AppDemoServiceCallHandler struct {
	Deps     *module.Dependencies
	endpoint *zanzibar.TChannelEndpoint
}

// Register adds the tchannel handler to the gateway's tchannel router
func (h *AppDemoServiceCallHandler) Register(g *zanzibar.Gateway) error {
	return g.ServerTChannelRouter.Register(h.endpoint)
}

// Handle handles RPC call of "AppDemoService::Call".
func (h *AppDemoServiceCallHandler) Handle(
	ctx context.Context,
	reqHeaders map[string]string,
	wireValue *wire.Value,
) (ctxRes context.Context, isSuccessful bool, response zanzibar.RWTStruct, headers map[string]string, e error) {
	defer func() {
		if r := recover(); r != nil {
			stacktrace := string(debug.Stack())
			e = errors.Errorf("enpoint panic: %v, stacktrace: %v", r, stacktrace)
			ctx = h.Deps.Default.ContextLogger.ErrorZ(
				ctx,
				"Endpoint failure: endpoint panic",
				zap.Error(e),
				zap.String("stacktrace", stacktrace),
				zap.String("endpoint", h.endpoint.EndpointID))

			h.Deps.Default.ContextMetrics.IncCounter(ctx, zanzibar.MetricEndpointPanics, 1)
			isSuccessful = false
			response = nil
			headers = nil
		}
	}()

	wfReqHeaders := zanzibar.ServerTChannelHeader(reqHeaders)

	var res endpointsIDlEndpointsAppDemoEndpointsAbc.AppDemoService_Call_Result

	workflow := customAbc.NewAppDemoServiceCallWorkflow(h.Deps)

	ctx, r, wfResHeaders, err := workflow.Handle(ctx, wfReqHeaders)

	resHeaders := map[string]string{}
	if wfResHeaders != nil {
		for _, key := range wfResHeaders.Keys() {
			resHeaders[key], _ = wfResHeaders.Get(key)
		}
	}

	if err != nil {
		ctx = h.Deps.Default.ContextLogger.ErrorZ(ctx, "Endpoint failure: handler returned error", zap.Error(err))
		return ctx, false, nil, resHeaders, err
	}
	res.Success = &r

	return ctx, err == nil, &res, resHeaders, nil
}
