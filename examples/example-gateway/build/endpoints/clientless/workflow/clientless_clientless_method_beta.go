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

package clientlessworkflow

import (
	"context"

	zanzibar "github.com/uber/zanzibar/v2/runtime"

	endpointsIDlEndpointsClientlessClientless "github.com/uber/zanzibar/v2/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/clientless/clientless"

	module "github.com/uber/zanzibar/v2/examples/example-gateway/build/endpoints/clientless/module"
	"go.uber.org/zap"
)

// ClientlessBetaWorkflow defines the interface for ClientlessBeta workflow
type ClientlessBetaWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
		r *endpointsIDlEndpointsClientlessClientless.Clientless_Beta_Args,
	) (context.Context, *endpointsIDlEndpointsClientlessClientless.Response, zanzibar.Header, error)
}

// NewClientlessBetaWorkflow creates a workflow
func NewClientlessBetaWorkflow(deps *module.Dependencies) ClientlessBetaWorkflow {

	return &clientlessBetaWorkflow{
		Logger: deps.Default.Logger,
	}
}

// clientlessBetaWorkflow calls thrift client .
type clientlessBetaWorkflow struct {
	Logger *zap.Logger
}

// Handle processes the request without a downstream
func (w clientlessBetaWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsIDlEndpointsClientlessClientless.Clientless_Beta_Args,
) (context.Context, *endpointsIDlEndpointsClientlessClientless.Response, zanzibar.Header, error) {
	response := convertBetaDummyResponse(r)

	clientlessHeaders := map[string]string{}

	var ok bool
	var h string
	h, ok = reqHeaders.Get("X-Deputy-Forwarded")
	if ok {
		clientlessHeaders["X-Deputy-Forwarded"] = h
	}

	// Filter and map response headers from client to server response.
	resHeaders := zanzibar.ServerHTTPHeader{}

	resHeaders.Set(zanzibar.ClientTypeKey, "clientless")
	return ctx, response, resHeaders, nil
}

func convertBetaDummyResponse(in *endpointsIDlEndpointsClientlessClientless.Clientless_Beta_Args) *endpointsIDlEndpointsClientlessClientless.Response {
	out := &endpointsIDlEndpointsClientlessClientless.Response{}

	if in.Request != nil {
		out.FirstName = (*string)(in.Request.FirstName)
	}
	if in.Request != nil {
		out.LastName1 = (*string)(in.Request.LastName)
	}

	return out
}
