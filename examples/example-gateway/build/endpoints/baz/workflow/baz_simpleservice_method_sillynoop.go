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
	"net/textproto"

	"github.com/uber/zanzibar/config"

	zanzibar "github.com/uber/zanzibar/runtime"

	clientsBazBase "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/base"
	clientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	endpointsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/baz/baz"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/baz/module"
	"go.uber.org/zap"
)

// SimpleServiceSillyNoopWorkflow defines the interface for SimpleServiceSillyNoop workflow
type SimpleServiceSillyNoopWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
	) (zanzibar.Header, error)
}

// NewSimpleServiceSillyNoopWorkflow creates a workflow
func NewSimpleServiceSillyNoopWorkflow(deps *module.Dependencies) SimpleServiceSillyNoopWorkflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.baz.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.baz.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &simpleServiceSillyNoopWorkflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
	}
}

// simpleServiceSillyNoopWorkflow calls thrift client Baz.DeliberateDiffNoop
type simpleServiceSillyNoopWorkflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
}

// Handle calls thrift client.
func (w simpleServiceSillyNoopWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
) (zanzibar.Header, error) {

	clientHeaders := map[string]string{}

	var ok bool
	var h string

	k := textproto.CanonicalMIMEHeaderKey("x-uber-foo")
	h, ok = reqHeaders.Get(k)
	if ok {
		clientHeaders[k] = h
	}
	k := textproto.CanonicalMIMEHeaderKey("x-uber-bar")
	h, ok = reqHeaders.Get(k)
	if ok {
		clientHeaders[k] = h
	}

	h, ok = reqHeaders.Get("X-Deputy-Forwarded")
	if ok {
		clientHeaders["X-Deputy-Forwarded"] = h
	}
	for _, whitelistedHeader := range w.whitelistedDynamicHeaders {
		headerVal, ok := reqHeaders.Get(whitelistedHeader)
		if ok {
			clientHeaders[whitelistedHeader] = headerVal
		}
	}

	_, err := w.Clients.Baz.DeliberateDiffNoop(ctx, clientHeaders)

	if err != nil {
		switch errValue := err.(type) {

		case *clientsBazBaz.AuthErr:
			serverErr := convertSillyNoopAuthErr(
				errValue,
			)

			return nil, serverErr

		case *clientsBazBase.ServerErr:
			serverErr := convertSillyNoopServerErr(
				errValue,
			)

			return nil, serverErr

		default:
			w.Logger.Warn("Client failure: could not make client request",
				zap.Error(errValue),
				zap.String("client", "Baz"),
			)

			return nil, err

		}
	}

	// Filter and map response headers from client to server response.
	resHeaders := zanzibar.ServerHTTPHeader{}

	return resHeaders, nil
}

func convertSillyNoopAuthErr(
	clientError *clientsBazBaz.AuthErr,
) *endpointsBazBaz.AuthErr {
	// TODO: Add error fields mapping here.
	serverError := &endpointsBazBaz.AuthErr{}
	return serverError
}
func convertSillyNoopServerErr(
	clientError *clientsBazBase.ServerErr,
) *endpointsBazBaz.ServerErr {
	// TODO: Add error fields mapping here.
	serverError := &endpointsBazBaz.ServerErr{}
	return serverError
}
