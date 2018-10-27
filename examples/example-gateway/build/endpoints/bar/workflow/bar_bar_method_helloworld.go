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

// BarHelloWorldWorkflow defines the interface for BarHelloWorld workflow
type BarHelloWorldWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
	) (string, zanzibar.Header, error)
}

// NewBarHelloWorldWorkflow creates a workflow
func NewBarHelloWorldWorkflow(deps *module.Dependencies) BarHelloWorldWorkflow {
	return &barHelloWorldWorkflow{
		Clients: deps.Client,
		Logger:  deps.Default.Logger,
	}
}

// barHelloWorldWorkflow calls thrift client Bar.Hello
type barHelloWorldWorkflow struct {
	Clients *module.ClientDependencies
	Logger  *zap.Logger
}

// Handle calls thrift client.
func (w barHelloWorldWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
) (string, zanzibar.Header, error) {

	clientHeaders := map[string]string{}

	var ok bool
	var h string
	h, ok = reqHeaders.Get("X-Deputy-Forwarded")
	if ok {
		clientHeaders["X-Deputy-Forwarded"] = h
	}
	h, ok = reqHeaders.Get("X-Zanzibar-Use-Staging")
	if ok {
		clientHeaders["X-Zanzibar-Use-Staging"] = h
	}

	clientRespBody, _, err := w.Clients.Bar.Hello(
		ctx, clientHeaders,
	)

	if err != nil {
		switch errValue := err.(type) {

		case *clientsBarBar.BarException:
			serverErr := convertHelloWorldBarException(
				errValue,
			)

			return "", nil, serverErr

		default:
			w.Logger.Warn("Could not make client request",
				zap.Error(errValue),
				zap.String("client", "Bar"),
			)

			return "", nil, err

		}
	}

	// Filter and map response headers from client to server response.
	resHeaders := zanzibar.ServerHTTPHeader{}

	response := convertBarHelloWorldClientResponse(clientRespBody)
	return response, resHeaders, nil
}

func convertHelloWorldBarException(
	clientError *clientsBarBar.BarException,
) *endpointsBarBar.BarException {
	// TODO: Add error fields mapping here.
	serverError := &endpointsBarBar.BarException{}
	return serverError
}

func convertBarHelloWorldClientResponse(in string) string {
	out := in

	return out
}
