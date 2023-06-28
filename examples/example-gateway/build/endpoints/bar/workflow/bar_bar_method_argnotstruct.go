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

	clientsIDlClientsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/bar/bar"
	endpointsIDlEndpointsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/bar/bar"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/bar/module"
	"go.uber.org/zap"
)

// BarArgNotStructWorkflow defines the interface for BarArgNotStruct workflow
type BarArgNotStructWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
		r *endpointsIDlEndpointsBarBar.Bar_ArgNotStruct_Args,
	) (context.Context, zanzibar.Header, error)
}

// NewBarArgNotStructWorkflow creates a workflow
func NewBarArgNotStructWorkflow(deps *module.Dependencies) BarArgNotStructWorkflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.bar.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.bar.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &barArgNotStructWorkflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
		defaultDeps:               deps.Default,
		errorBuilder:              zanzibar.NewErrorBuilder("endpoint", "bar"),
	}
}

// barArgNotStructWorkflow calls thrift client Bar.ArgNotStruct
type barArgNotStructWorkflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
	defaultDeps               *zanzibar.DefaultDependencies
	errorBuilder              zanzibar.ErrorBuilder
}

// Handle calls thrift client.
func (w barArgNotStructWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsIDlEndpointsBarBar.Bar_ArgNotStruct_Args,
) (context.Context, zanzibar.Header, error) {
	clientRequest := convertToArgNotStructClientRequest(r)

	clientHeaders := map[string]string{}

	var ok bool
	var h string
	var k string

	k = textproto.CanonicalMIMEHeaderKey("x-uber-foo")
	h, ok = reqHeaders.Get(k)
	if ok {
		clientHeaders[k] = h
	}
	k = textproto.CanonicalMIMEHeaderKey("x-uber-bar")
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

	//when maxRetry is 0, timeout per client level is used & one attempt is made, and timoutPerAttempt is not used
	var timeoutAndRetryConfig = zanzibar.TimeoutAndRetryOptions{}

	//when endpoint level timeout information is available, override it with client level config
	if w.defaultDeps.Config.ContainsKey("endpoints.bar.argNotStruct.timeoutPerAttempt") {
		scaleFactor := w.defaultDeps.Config.MustGetFloat("endpoints.bar.argNotStruct.scaleFactor")
		maxRetry := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.argNotStruct.retryCount"))

		backOffTimeAcrossRetriesCfg := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.argNotStruct.backOffTimeAcrossRetries"))
		timeoutPerAttemptConf := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.argNotStruct.timeoutPerAttempt"))

		timeoutAndRetryConfig = zanzibar.BuildTimeoutAndRetryConfig(int(timeoutPerAttemptConf), backOffTimeAcrossRetriesCfg, maxRetry, scaleFactor)
	}

	ctx, _, err := w.Clients.Bar.ArgNotStruct(
		ctx, clientHeaders, clientRequest, &timeoutAndRetryConfig,
	)

	if err != nil {
		zErr, ok := err.(zanzibar.Error)
		if ok {
			err = zErr.Unwrap()
		}
		switch errValue := err.(type) {

		case *clientsIDlClientsBarBar.BarException:
			err = convertArgNotStructBarException(
				errValue,
			)

		default:
			w.Logger.Warn("Client failure: could not make client request",
				zap.Error(errValue),
				zap.String("client", "Bar"),
			)
		}
		err = w.errorBuilder.Rebuild(zErr, err)
		return ctx, nil, err
	}

	// Filter and map response headers from client to server response.
	resHeaders := zanzibar.ServerHTTPHeader{}

	return ctx, resHeaders, nil
}

func convertToArgNotStructClientRequest(in *endpointsIDlEndpointsBarBar.Bar_ArgNotStruct_Args) *clientsIDlClientsBarBar.Bar_ArgNotStruct_Args {
	out := &clientsIDlClientsBarBar.Bar_ArgNotStruct_Args{}

	out.Request = string(in.Request)

	return out
}

func convertArgNotStructBarException(
	clientError *clientsIDlClientsBarBar.BarException,
) *endpointsIDlEndpointsBarBar.BarException {
	// TODO: Add error fields mapping here.
	serverError := &endpointsIDlEndpointsBarBar.BarException{}
	return serverError
}
