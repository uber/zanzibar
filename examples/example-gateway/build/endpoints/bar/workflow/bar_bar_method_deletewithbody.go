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

// BarDeleteWithBodyWorkflow defines the interface for BarDeleteWithBody workflow
type BarDeleteWithBodyWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
		r *endpointsIDlEndpointsBarBar.Bar_DeleteWithBody_Args,
	) (context.Context, zanzibar.Header, error)
}

// NewBarDeleteWithBodyWorkflow creates a workflow
func NewBarDeleteWithBodyWorkflow(deps *module.Dependencies) BarDeleteWithBodyWorkflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.bar.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.bar.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &barDeleteWithBodyWorkflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
		defaultDeps:               deps.Default,
		errorBuilder:              zanzibar.NewErrorBuilder("endpoint", "bar"),
	}
}

// barDeleteWithBodyWorkflow calls thrift client Bar.DeleteWithBody
type barDeleteWithBodyWorkflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
	defaultDeps               *zanzibar.DefaultDependencies
	errorBuilder              zanzibar.ErrorBuilder
}

// Handle calls thrift client.
func (w barDeleteWithBodyWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsIDlEndpointsBarBar.Bar_DeleteWithBody_Args,
) (context.Context, zanzibar.Header, error) {
	clientRequest := convertToDeleteWithBodyClientRequest(r)

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
	if w.defaultDeps.Config.ContainsKey("endpoints.bar.deleteWithBody.timeoutPerAttempt") {
		scaleFactor := w.defaultDeps.Config.MustGetFloat("endpoints.bar.deleteWithBody.scaleFactor")
		maxRetry := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.deleteWithBody.retryCount"))

		backOffTimeAcrossRetriesCfg := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.deleteWithBody.backOffTimeAcrossRetries"))
		timeoutPerAttemptConf := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.deleteWithBody.timeoutPerAttempt"))

		timeoutAndRetryConfig = zanzibar.BuildTimeoutAndRetryConfig(int(timeoutPerAttemptConf), backOffTimeAcrossRetriesCfg, maxRetry, scaleFactor)
	}

	ctx, _, err := w.Clients.Bar.DeleteWithBody(
		ctx, clientHeaders, clientRequest, &timeoutAndRetryConfig,
	)

	if err != nil {
		zErr, ok := err.(zanzibar.Error)
		if ok {
			err = zErr.Unwrap()
		}
		switch errValue := err.(type) {

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

func convertToDeleteWithBodyClientRequest(in *endpointsIDlEndpointsBarBar.Bar_DeleteWithBody_Args) *clientsIDlClientsBarBar.Bar_DeleteWithBody_Args {
	out := &clientsIDlClientsBarBar.Bar_DeleteWithBody_Args{}

	out.Filter = string(in.Filter)
	out.Count = (*int32)(in.Count)

	return out
}
