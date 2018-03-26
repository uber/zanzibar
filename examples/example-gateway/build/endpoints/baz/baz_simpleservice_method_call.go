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

package bazendpoint

import (
	"context"
	"fmt"
	"strconv"

	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	clientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	endpointsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/baz/baz"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/baz/module"
)

// SimpleServiceCallHandler is the handler for "/baz/call"
type SimpleServiceCallHandler struct {
	Clients  *module.ClientDependencies
	endpoint *zanzibar.RouterEndpoint
}

// NewSimpleServiceCallHandler creates a handler
func NewSimpleServiceCallHandler(deps *module.Dependencies) *SimpleServiceCallHandler {
	handler := &SimpleServiceCallHandler{
		Clients: deps.Client,
	}
	handler.endpoint = zanzibar.NewRouterEndpoint(
		deps.Default.Logger, deps.Default.Scope,
		"baz", "call",
		handler.HandleRequest,
	)
	return handler
}

// Register adds the http handler to the gateway's http router
func (h *SimpleServiceCallHandler) Register(g *zanzibar.Gateway) error {
	g.HTTPRouter.Register(
		"POST", "/baz/call",
		h.endpoint,
	)
	// TODO: register should return errors on route conflicts
	return nil
}

// HandleRequest handles "/baz/call".
func (h *SimpleServiceCallHandler) HandleRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
) {
	var requestBody endpointsBazBaz.SimpleService_Call_Args
	if ok := req.ReadAndUnmarshalBody(&requestBody); !ok {
		return
	}

	if requestBody.Arg == nil {
		requestBody.Arg = &endpointsBazBaz.BazRequest{}
	}
	xTokenValue, xTokenValueExists := req.Header.Get("x-token")
	if xTokenValueExists {
		body, _ := strconv.ParseInt(xTokenValue, 10, 64)
		requestBody.I64Optional = &body
	}
	xUUIDValue, xUUIDValueExists := req.Header.Get("x-uuid")
	if xUUIDValueExists {
		body := endpointsBazBaz.UUID(xUUIDValue)
		requestBody.TestUUID = &body
	}

	// log endpoint request to downstream services
	zfields := []zapcore.Field{
		zap.String("endpoint", h.endpoint.EndpointName),
	}

	// TODO: potential perf issue, use zap.Object lazy serialization
	zfields = append(zfields, zap.String("body", fmt.Sprintf("%#v", requestBody)))
	var headerOk bool
	var headerValue string
	headerValue, headerOk = req.Header.Get("X-Token")
	if headerOk {
		zfields = append(zfields, zap.String("X-Token", headerValue))
	}
	headerValue, headerOk = req.Header.Get("X-Uuid")
	if headerOk {
		zfields = append(zfields, zap.String("X-Uuid", headerValue))
	}
	headerValue, headerOk = req.Header.Get("X-Zanzibar-Use-Staging")
	if headerOk {
		zfields = append(zfields, zap.String("X-Zanzibar-Use-Staging", headerValue))
	}
	req.Logger.Debug("Endpoint request to downstream", zfields...)

	workflow := SimpleServiceCallEndpoint{
		Clients: h.Clients,
		Logger:  req.Logger,
		Request: req,
	}

	cliRespHeaders, err := workflow.Handle(ctx, req.Header, &requestBody)
	if err != nil {
		switch errValue := err.(type) {

		case *endpointsBazBaz.AuthErr:
			res.WriteJSON(
				403, cliRespHeaders, errValue,
			)
			return

		default:
			res.SendError(500, "Unexpected server error", err)
			return
		}

	}

	res.WriteJSONBytes(204, cliRespHeaders, nil)
}

// SimpleServiceCallEndpoint calls thrift client Baz.Call
type SimpleServiceCallEndpoint struct {
	Clients *module.ClientDependencies
	Logger  *zap.Logger
	Request *zanzibar.ServerHTTPRequest
}

// Handle calls thrift client.
func (w SimpleServiceCallEndpoint) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsBazBaz.SimpleService_Call_Args,
) (zanzibar.Header, error) {
	clientRequest := convertToCallClientRequest(r)

	clientHeaders := map[string]string{}

	var ok bool
	var h string
	h, ok = reqHeaders.Get("X-Token")
	if ok {
		clientHeaders["X-Token"] = h
	}
	h, ok = reqHeaders.Get("X-Uuid")
	if ok {
		clientHeaders["X-Uuid"] = h
	}
	h, ok = reqHeaders.Get("X-Zanzibar-Use-Staging")
	if ok {
		clientHeaders["X-Zanzibar-Use-Staging"] = h
	}

	cliRespHeaders, err := w.Clients.Baz.Call(
		ctx, clientHeaders, clientRequest,
	)

	if err != nil {
		switch errValue := err.(type) {

		case *clientsBazBaz.AuthErr:
			serverErr := convertCallAuthErr(
				errValue,
			)
			// TODO(sindelar): Consider returning partial headers

			return nil, serverErr

		default:
			w.Logger.Warn("Could not make client request",
				zap.Error(errValue),
				zap.String("client", "Baz"),
			)

			// TODO(sindelar): Consider returning partial headers

			return nil, err

		}
	}

	// Filter and map response headers from client to server response.

	// TODO: Add support for TChannel Headers with a switch here
	resHeaders := zanzibar.ServerHTTPHeader{}

	resHeaders.Set("Some-Res-Header", cliRespHeaders["Some-Res-Header"])

	return resHeaders, nil
}

func convertToCallClientRequest(in *endpointsBazBaz.SimpleService_Call_Args) *clientsBazBaz.SimpleService_Call_Args {
	out := &clientsBazBaz.SimpleService_Call_Args{}

	if in.Arg != nil {
		out.Arg = &clientsBazBaz.BazRequest{}
		out.Arg.B1 = bool(in.Arg.B1)
		out.Arg.S2 = string(in.Arg.S2)
		out.Arg.I3 = int32(in.Arg.I3)
	} else {
		out.Arg = nil
	}
	out.I64Optional = (*int64)(in.I64Optional)
	out.TestUUID = (*clientsBazBaz.UUID)(in.TestUUID)

	return out
}

func convertCallAuthErr(
	clientError *clientsBazBaz.AuthErr,
) *endpointsBazBaz.AuthErr {
	// TODO: Add error fields mapping here.
	serverError := &endpointsBazBaz.AuthErr{}
	return serverError
}
