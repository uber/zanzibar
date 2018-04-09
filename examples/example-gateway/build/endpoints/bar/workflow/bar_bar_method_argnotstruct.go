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

package workflow

import (
	"context"

	zanzibar "github.com/uber/zanzibar/runtime"

	clientsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/bar/bar"
	endpointsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/bar/bar"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/bar/module"
	"go.uber.org/zap"
)

// BarArgNotStructWorkflow defines the interface for BarArgNotStruct workflow
type BarArgNotStructWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
		r *endpointsBarBar.Bar_ArgNotStruct_Args,
	) (zanzibar.Header, error)
}

// NewBarArgNotStructWorkflow creates a workflow
func NewBarArgNotStructWorkflow(clients *module.ClientDependencies, logger *zap.Logger) BarArgNotStructWorkflow {
	return &barArgNotStructWorkflow{
		Clients: clients,
		Logger:  logger,
	}
}

// barArgNotStructWorkflow calls thrift client Bar.ArgNotStruct
type barArgNotStructWorkflow struct {
	Clients *module.ClientDependencies
	Logger  *zap.Logger
}

// Handle calls thrift client.
func (w barArgNotStructWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsBarBar.Bar_ArgNotStruct_Args,
) (zanzibar.Header, error) {
	clientRequest := convertToArgNotStructClientRequest(r)

	clientHeaders := map[string]string{}

	var ok bool
	var h string
	h, ok = reqHeaders.Get("X-Zanzibar-Use-Staging")
	if ok {
		clientHeaders["X-Zanzibar-Use-Staging"] = h
	}

	_, err := w.Clients.Bar.ArgNotStruct(
		ctx, clientHeaders, clientRequest,
	)

	if err != nil {
		switch errValue := err.(type) {

		case *clientsBarBar.BarException:
			serverErr := convertArgNotStructBarException(
				errValue,
			)
			// TODO(sindelar): Consider returning partial headers

			return nil, serverErr

		default:
			w.Logger.Warn("Could not make client request",
				zap.Error(errValue),
				zap.String("client", "Bar"),
			)

			// TODO(sindelar): Consider returning partial headers

			return nil, err

		}
	}

	// Filter and map response headers from client to server response.

	// TODO: Add support for TChannel Headers with a switch here
	resHeaders := zanzibar.ServerHTTPHeader{}

	return resHeaders, nil
}

func convertToArgNotStructClientRequest(in *endpointsBarBar.Bar_ArgNotStruct_Args) *clientsBarBar.Bar_ArgNotStruct_Args {
	out := &clientsBarBar.Bar_ArgNotStruct_Args{}

	out.Request = string(in.Request)

	return out
}

func convertArgNotStructBarException(
	clientError *clientsBarBar.BarException,
) *endpointsBarBar.BarException {
	// TODO: Add error fields mapping here.
	serverError := &endpointsBarBar.BarException{}
	return serverError
}
