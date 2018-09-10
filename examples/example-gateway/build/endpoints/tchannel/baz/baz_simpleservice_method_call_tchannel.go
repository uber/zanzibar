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

package baztchannelendpoint

import (
	"context"
	"runtime/debug"
	"strings"
	"time"

	"github.com/pkg/errors"
	tchannel "github.com/uber/tchannel-go"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/zap"

	endpointsTchannelBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/tchannel/baz/baz"
	customBaz "github.com/uber/zanzibar/examples/example-gateway/endpoints/tchannel/baz"

	example_tchannel "github.com/uber/zanzibar/examples/example-gateway/middlewares/example_tchannel"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/tchannel/baz/module"
)

// NewSimpleServiceCallHandler creates a handler to be registered with a thrift server.
func NewSimpleServiceCallHandler(deps *module.Dependencies) *SimpleServiceCallHandler {
	handler := &SimpleServiceCallHandler{
		Deps: deps,
	}
	handler.endpoint = zanzibar.NewTChannelEndpoint(
		deps.Default.Logger, deps.Default.Scope,
		"bazTChannel", "call", "SimpleService::Call",
		zanzibar.NewTchannelStack([]zanzibar.MiddlewareTchannelHandle{
			deps.Middleware.ExampleTchannel.NewMiddlewareHandle(
				example_tchannel.Options{
					Foo: "test",
				},
			),
		}, handler),
	)

	return handler
}

// SimpleServiceCallHandler is the handler for "SimpleService::Call".
type SimpleServiceCallHandler struct {
	Deps     *module.Dependencies
	endpoint *zanzibar.TChannelEndpoint
}

// Register adds the tchannel handler to the gateway's tchannel router
func (h *SimpleServiceCallHandler) Register(g *zanzibar.Gateway) error {
	return g.TChannelRouter.Register(h.endpoint)
}

// Handle handles RPC call of "SimpleService::Call".
func (h *SimpleServiceCallHandler) Handle(
	ctx context.Context,
	reqHeaders map[string]string,
	wireValue *wire.Value,
) (isSuccessful bool, response zanzibar.RWTStruct, headers map[string]string, e error) {
	defer func() {
		if r := recover(); r != nil {
			stacktrace := string(debug.Stack())
			e = errors.Errorf("enpoint panic: %v, stacktrace: %v", r, stacktrace)
			h.Deps.Default.ContextLogger.Error(
				ctx,
				"endpoint panic",
				zap.Error(e),
				zap.String("stacktrace", stacktrace),
				zap.String("endpoint", h.endpoint.EndpointID))

			h.endpoint.Metrics.Panic.Inc(1)
			isSuccessful = false
			response = nil
			headers = nil
		}
	}()

	wfReqHeaders := zanzibar.ServerTChannelHeader(reqHeaders)
	if err := wfReqHeaders.EnsureContext(ctx, []string{"x-uuid", "x-token"}, h.Deps.Default.ContextLogger); err != nil {
		return false, nil, nil, errors.Wrapf(
			err, "%s.%s (%s) missing request headers",
			h.endpoint.EndpointID, h.endpoint.HandlerID, h.endpoint.Method,
		)
	}

	var res endpointsTchannelBazBaz.SimpleService_Call_Result

	var req endpointsTchannelBazBaz.SimpleService_Call_Args
	if err := req.FromWire(*wireValue); err != nil {
		h.Deps.Default.ContextLogger.Warn(ctx, "Error converting request from wire", zap.Error(err))
		return false, nil, nil, errors.Wrapf(
			err, "Error converting %s.%s (%s) request from wire",
			h.endpoint.EndpointID, h.endpoint.HandlerID, h.endpoint.Method,
		)
	}

	if hostPort, ok := reqHeaders["x-deputy-forwarded"]; ok {
		if hostPort != "" {
			return h.redirectToDeputy(ctx, reqHeaders, hostPort, &req, &res)
		}
	}
	workflow := customBaz.NewSimpleServiceCallWorkflow(h.Deps)

	wfResHeaders, err := workflow.Handle(ctx, wfReqHeaders, &req)

	resHeaders := map[string]string{}
	if wfResHeaders != nil {
		for _, key := range wfResHeaders.Keys() {
			resHeaders[key], _ = wfResHeaders.Get(key)
		}
	}

	if err != nil {
		switch v := err.(type) {
		case *endpointsTchannelBazBaz.AuthErr:
			h.Deps.Default.ContextLogger.Warn(
				ctx,
				"Handler returned non-nil error type *endpointsTchannelBazBaz.AuthErr but nil value",
				zap.Error(err),
			)
			if v == nil {
				return false, nil, resHeaders, errors.Errorf(
					"%s.%s (%s) handler returned non-nil error type *endpointsTchannelBazBaz.AuthErr but nil value",
					h.endpoint.EndpointID, h.endpoint.HandlerID, h.endpoint.Method,
				)
			}
			res.AuthErr = v
		default:
			h.Deps.Default.ContextLogger.Warn(ctx, "Handler returned error", zap.Error(err))
			return false, nil, resHeaders, errors.Wrapf(
				err, "%s.%s (%s) handler returned error",
				h.endpoint.EndpointID, h.endpoint.HandlerID, h.endpoint.Method,
			)
		}
	}
	if wfResHeaders == nil {
		return false, nil, nil, errors.Wrapf(
			errors.Errorf(
				"Missing mandatory headers: %s",
				strings.Join([]string{"some-res-header"}, ", "),
			),
			"%s.%s (%s) missing response headers",
			h.endpoint.EndpointID, h.endpoint.HandlerID, h.endpoint.Method,
		)
	}

	if err := wfResHeaders.EnsureContext(ctx, []string{"some-res-header"}, h.Deps.Default.ContextLogger); err != nil {
		return false, nil, nil, errors.Wrapf(
			err, "%s.%s (%s) missing response headers",
			h.endpoint.EndpointID, h.endpoint.HandlerID, h.endpoint.Method,
		)
	}

	return err == nil, &res, resHeaders, nil
}

// redirectToDeputy sends the request to deputy hostPort
func (h *SimpleServiceCallHandler) redirectToDeputy(
	ctx context.Context,
	reqHeaders map[string]string,
	hostPort string,
	req *endpointsTchannelBazBaz.SimpleService_Call_Args,
	res *endpointsTchannelBazBaz.SimpleService_Call_Result,
) (bool, zanzibar.RWTStruct, map[string]string, error) {
	var routingKey string
	if h.Deps.Default.Config.ContainsKey("tchannel.routingKey") {
		routingKey = h.Deps.Default.Config.MustGetString("tchannel.routingKey")
	}

	serviceName := h.Deps.Default.Config.MustGetString("tchannel.serviceName")
	timeout := time.Millisecond * time.Duration(
		h.Deps.Default.Config.MustGetInt("tchannel.deputy.timeout"),
	)

	timeoutPerAttempt := time.Millisecond * time.Duration(
		h.Deps.Default.Config.MustGetInt("tchannel.deputy.timeoutPerAttempt"),
	)

	methodNames := map[string]string{
		"SimpleService::Call": "Call",
	}

	sub := h.Deps.Default.Channel.GetSubChannel(serviceName, tchannel.Isolated)
	sub.Peers().Add(hostPort)
	client := zanzibar.NewTChannelClient(
		h.Deps.Default.Channel,
		h.Deps.Default.Logger,
		h.Deps.Default.Scope,
		&zanzibar.TChannelClientOption{
			ServiceName:       serviceName,
			ClientID:          "",
			MethodNames:       methodNames,
			Timeout:           timeout,
			TimeoutPerAttempt: timeoutPerAttempt,
			RoutingKey:        &routingKey,
		},
	)

	success, respHeaders, err := client.Call(ctx, "SimpleService", "Call", reqHeaders, req, res)
	_ = sub.Peers().Remove(hostPort)
	return success, res, respHeaders, err
}
