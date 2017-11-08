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

package barEndpoint

import (
	"context"

	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"

	clientsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/bar/bar"
	endpointsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/bar/bar"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/bar/module"
)

// BarNoRequestHandler is the handler for "/bar/no-request-path"
type BarNoRequestHandler struct {
	Clients  *module.ClientDependencies
	endpoint *zanzibar.RouterEndpoint
}

// NewBarNoRequestHandler creates a handler
func NewBarNoRequestHandler(
	g *zanzibar.Gateway,
	deps *module.Dependencies,
) *BarNoRequestHandler {
	handler := &BarNoRequestHandler{
		Clients: deps.Client,
	}
	handler.endpoint = zanzibar.NewRouterEndpoint(
		deps.Default.Logger, deps.Default.Scope,
		"bar", "noRequest",
		handler.HandleRequest,
	)
	return handler
}

// Register adds the http handler to the gateway's http router
func (h *BarNoRequestHandler) Register(g *zanzibar.Gateway) error {
	g.HTTPRouter.Register(
		"GET", "/bar/no-request-path",
		h.endpoint,
	)
	// TODO: register should return errors on route conflicts
	return nil
}

// HandleRequest handles "/bar/no-request-path".
func (h *BarNoRequestHandler) HandleRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
) {

	workflow := NoRequestEndpoint{
		Clients: h.Clients,
		Logger:  req.Logger,
		Request: req,
	}

	response, cliRespHeaders, err := workflow.Handle(ctx, req.Header)
	if err != nil {
		switch errValue := err.(type) {

		case *endpointsBarBar.BarException:
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
	// TODO(jakev): implement writing fields into response headers

	res.WriteJSON(200, cliRespHeaders, response)
}

// NoRequestEndpoint calls thrift client Bar.NoRequest
type NoRequestEndpoint struct {
	Clients *module.ClientDependencies
	Logger  *zap.Logger
	Request *zanzibar.ServerHTTPRequest
}

// Handle calls thrift client.
func (w NoRequestEndpoint) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
) (*endpointsBarBar.BarResponse, zanzibar.Header, error) {

	clientHeaders := map[string]string{}

	clientRespBody, _, err := w.Clients.Bar.NoRequest(
		ctx, clientHeaders,
	)

	if err != nil {
		switch errValue := err.(type) {

		case *clientsBarBar.BarException:
			serverErr := convertNoRequestBarException(
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

	response := convertNoRequestClientResponse(clientRespBody)
	return response, resHeaders, nil
}

func convertNoRequestBarException(
	clientError *clientsBarBar.BarException,
) *endpointsBarBar.BarException {
	// TODO: Add error fields mapping here.
	serverError := &endpointsBarBar.BarException{}
	return serverError
}

func convertNoRequestClientResponse(in *clientsBarBar.BarResponse) *endpointsBarBar.BarResponse {
	out := &endpointsBarBar.BarResponse{}

	out.StringField = string(in.StringField)
	out.IntWithRange = int32(in.IntWithRange)
	out.IntWithoutRange = int32(in.IntWithoutRange)
	out.MapIntWithRange = make(map[endpointsBarBar.UUID]int32, len(in.MapIntWithRange))
	for key, value := range in.MapIntWithRange {
		out.MapIntWithRange[endpointsBarBar.UUID(key)] = int32(value)
	}
	out.MapIntWithoutRange = make(map[string]int32, len(in.MapIntWithoutRange))
	for key, value := range in.MapIntWithoutRange {
		out.MapIntWithoutRange[key] = int32(value)
	}
	out.BinaryField = []byte(in.BinaryField)

	return out
}
