// Code generated by zanzibar
// @generated

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

package bazEndpoint

import (
	"context"

	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"

	clientsBazBase "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/base"
	clientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	endpointsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/baz/baz"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/baz/module"
)

// SimpleServiceTransHandler is the handler for "/baz/trans"
type SimpleServiceTransHandler struct {
	Clients  *module.ClientDependencies
	endpoint *zanzibar.RouterEndpoint
}

// NewSimpleServiceTransHandler creates a handler
func NewSimpleServiceTransHandler(
	g *zanzibar.Gateway,
	deps *module.Dependencies,
) *SimpleServiceTransHandler {
	handler := &SimpleServiceTransHandler{
		Clients: deps.Client,
	}
	handler.endpoint = zanzibar.NewRouterEndpoint(
		deps.Default.Logger, deps.Default.Scope,
		"baz", "trans",
		handler.HandleRequest,
	)
	return handler
}

// Register adds the http handler to the gateway's http router
func (h *SimpleServiceTransHandler) Register(g *zanzibar.Gateway) error {
	g.HTTPRouter.Register(
		"POST", "/baz/trans",
		h.endpoint,
	)
	// TODO: register should return errors on route conflicts
	return nil
}

// HandleRequest handles "/baz/trans".
func (h *SimpleServiceTransHandler) HandleRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
) {
	var requestBody endpointsBazBaz.SimpleService_Trans_Args
	if ok := req.ReadAndUnmarshalBody(&requestBody); !ok {
		return
	}

	workflow := TransEndpoint{
		Clients: h.Clients,
		Logger:  req.Logger,
		Request: req,
	}

	response, cliRespHeaders, err := workflow.Handle(ctx, req.Header, &requestBody)
	if err != nil {
		switch errValue := err.(type) {

		case *endpointsBazBaz.AuthErr:
			res.WriteJSON(
				403, cliRespHeaders, errValue,
			)
			return

		case *endpointsBazBaz.OtherAuthErr:
			res.WriteJSON(
				403, cliRespHeaders, errValue,
			)
			return

		default:
			req.Logger.Warn("Workflow for endpoint returned error", zap.Error(errValue))
			res.SendErrorString(500, "Unexpected server error")
			return
		}
	}

	res.WriteJSON(200, cliRespHeaders, response)
}

// TransEndpoint calls thrift client Baz.Trans
type TransEndpoint struct {
	Clients *module.ClientDependencies
	Logger  *zap.Logger
	Request *zanzibar.ServerHTTPRequest
}

// Handle calls thrift client.
func (w TransEndpoint) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsBazBaz.SimpleService_Trans_Args,
) (*endpointsBazBaz.TransStruct, zanzibar.Header, error) {
	clientRequest := convertToTransClientRequest(r)

	clientHeaders := map[string]string{}

	clientRespBody, _, err := w.Clients.Baz.Trans(
		ctx, clientHeaders, clientRequest,
	)

	if err != nil {
		switch errValue := err.(type) {

		case *clientsBazBaz.AuthErr:
			serverErr := convertTransAuthErr(
				errValue,
			)
			// TODO(sindelar): Consider returning partial headers

			return nil, nil, serverErr

		case *clientsBazBaz.OtherAuthErr:
			serverErr := convertTransOtherAuthErr(
				errValue,
			)
			// TODO(sindelar): Consider returning partial headers

			return nil, nil, serverErr

		default:
			w.Logger.Warn("Could not make client request", zap.Error(errValue))
			// TODO(sindelar): Consider returning partial headers

			return nil, nil, err

		}
	}

	// Filter and map response headers from client to server response.

	// TODO: Add support for TChannel Headers with a switch here
	resHeaders := zanzibar.ServerHTTPHeader{}

	response := convertTransClientResponse(clientRespBody)
	return response, resHeaders, nil
}

func convertToTransClientRequest(in *endpointsBazBaz.SimpleService_Trans_Args) *clientsBazBaz.SimpleService_Trans_Args {
	out := &clientsBazBaz.SimpleService_Trans_Args{}

	convertToTransArg1ClientRequest(in.Arg1, out.Arg1)
	convertToTransArg2ClientRequest(in.Arg2, out.Arg2)

	return out
}

func convertToTransArg1ClientRequest(in *endpointsBazBaz.TransStruct, out *clientsBazBase.TransStruct) {
	if in != nil {
		out = &clientsBazBase.TransStruct{}
		out.Message = string(in.Message)
		convertToTransDriverClientRequest(in.Driver, out.Driver)
		convertToTransRiderClientRequest(in.Rider, out.Rider)
	} else {
		out = nil
	}
}

func convertToTransDriverClientRequest(in *endpointsBazBaz.NestedStruct, out *clientsBazBase.NestedStruct) {
	if in != nil {
		out = &clientsBazBase.NestedStruct{}
		out.Msg = string(in.Msg)
		out.Check = (*int32)(in.Check)
	} else {
		out = nil
	}
}

func convertToTransRiderClientRequest(in *endpointsBazBaz.NestedStruct, out *clientsBazBase.NestedStruct) {
	if in != nil {
		out = &clientsBazBase.NestedStruct{}
		out.Msg = string(in.Msg)
		out.Check = (*int32)(in.Check)
	} else {
		out = nil
	}
}

func convertToTransArg2ClientRequest(in *endpointsBazBaz.TransStruct, out *clientsBazBase.TransStruct) {
	if in != nil {
		out = &clientsBazBase.TransStruct{}
		out.Message = string(in.Message)
		convertToTransDriverClientRequest(in.Driver, out.Driver)
		convertToTransRiderClientRequest(in.Rider, out.Rider)
	} else {
		out = nil
	}
}

func convertTransAuthErr(
	clientError *clientsBazBaz.AuthErr,
) *endpointsBazBaz.AuthErr {
	// TODO: Add error fields mapping here.
	serverError := &endpointsBazBaz.AuthErr{}
	return serverError
}
func convertTransOtherAuthErr(
	clientError *clientsBazBaz.OtherAuthErr,
) *endpointsBazBaz.OtherAuthErr {
	// TODO: Add error fields mapping here.
	serverError := &endpointsBazBaz.OtherAuthErr{}
	return serverError
}

func convertTransClientResponse(in *clientsBazBase.TransStruct) *endpointsBazBaz.TransStruct {
	out := &endpointsBazBaz.TransStruct{}

	out.Message = string(in.Message)
	convertToTransDriverClientResponse(in.Driver, out.Driver)
	convertToTransRiderClientResponse(in.Rider, out.Rider)

	return out
}

func convertToTransDriverClientResponse(in *clientsBazBase.NestedStruct, out *endpointsBazBaz.NestedStruct) {
	if in != nil {
		out = &endpointsBazBaz.NestedStruct{}
		out.Msg = string(in.Msg)
		out.Check = (*int32)(in.Check)
	} else {
		out = nil
	}
}

func convertToTransRiderClientResponse(in *clientsBazBase.NestedStruct, out *endpointsBazBaz.NestedStruct) {
	if in != nil {
		out = &endpointsBazBaz.NestedStruct{}
		out.Msg = string(in.Message)
		out.Check = (*int32)(in.Check)
	} else {
		out = nil
	}
}
