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

	clientsIDlClientsWithexceptionsWithexceptions "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/withexceptions/withexceptions"
	endpointsIDlEndpointsWithexceptionsWithexceptions "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/withexceptions/withexceptions"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/withexceptions/module"
	"go.uber.org/zap"
)

// WithExceptionsFunc1Workflow defines the interface for WithExceptionsFunc1 workflow
type WithExceptionsFunc1Workflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
	) (context.Context, *endpointsIDlEndpointsWithexceptionsWithexceptions.Response, zanzibar.Header, error)
}

// NewWithExceptionsFunc1Workflow creates a workflow
func NewWithExceptionsFunc1Workflow(deps *module.Dependencies) WithExceptionsFunc1Workflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.withexceptions.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.withexceptions.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &withExceptionsFunc1Workflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
		defaultDeps:               deps.Default,
		errorBuilder:              zanzibar.NewErrorBuilder("endpoint", "withexceptions"),
	}
}

// withExceptionsFunc1Workflow calls thrift client Withexceptions.Func1
type withExceptionsFunc1Workflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
	defaultDeps               *zanzibar.DefaultDependencies
	errorBuilder              zanzibar.ErrorBuilder
}

// Handle calls thrift client.
func (w withExceptionsFunc1Workflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
) (context.Context, *endpointsIDlEndpointsWithexceptionsWithexceptions.Response, zanzibar.Header, error) {

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
	var timeoutAndRetryConfig *zanzibar.TimeoutAndRetryOptions

	//when endpoint level timeout information is available, override it with client level config
	if w.defaultDeps.Config.ContainsKey("endpoints.withexceptions.Func1.timeoutPerAttempt") {
		scaleFactor := w.defaultDeps.Config.MustGetFloat("endpoints.withexceptions.Func1.scaleFactor")
		maxRetry := int(w.defaultDeps.Config.MustGetInt("endpoints.withexceptions.Func1.retryCount"))

		backOffTimeAcrossRetriesCfg := int(w.defaultDeps.Config.MustGetInt("endpoints.withexceptions.Func1.backOffTimeAcrossRetries"))
		timeoutPerAttemptConf := int(w.defaultDeps.Config.MustGetInt("endpoints.withexceptions.Func1.timeoutPerAttempt"))

		timeoutAndRetryConfig = zanzibar.BuildTimeoutAndRetryConfig(int(timeoutPerAttemptConf), backOffTimeAcrossRetriesCfg, maxRetry, scaleFactor)
		ctx = zanzibar.WithTimeAndRetryOptions(ctx, timeoutAndRetryConfig)
	}

	ctx, clientRespBody, cliRespHeaders, err := w.Clients.Withexceptions.Func1(
		ctx, clientHeaders,
	)

	if err != nil {
		zErr, ok := err.(zanzibar.Error)
		if ok {
			err = zErr.Unwrap()
		}
		switch errValue := err.(type) {

		case *clientsIDlClientsWithexceptionsWithexceptions.ExceptionType1:
			err = convertFunc1E1(
				errValue,
			)

		case *clientsIDlClientsWithexceptionsWithexceptions.ExceptionType2:
			err = convertFunc1E2(
				errValue,
			)

		default:
			w.Logger.Warn("Client failure: could not make client request",
				zap.Error(errValue),
				zap.String("client", "Withexceptions"),
			)
		}
		err = w.errorBuilder.Rebuild(zErr, err)
		return ctx, nil, nil, err
	}

	// Filter and map response headers from client to server response.
	resHeaders := zanzibar.ServerHTTPHeader{}

	response := convertWithExceptionsFunc1ClientResponse(clientRespBody)
	if val, ok := cliRespHeaders[zanzibar.ClientResponseDurationKey]; ok {
		resHeaders.Set(zanzibar.ClientResponseDurationKey, val)
	}

	resHeaders.Set(zanzibar.ClientTypeKey, "http")
	return ctx, response, resHeaders, nil
}

func convertFunc1E1(
	clientError *clientsIDlClientsWithexceptionsWithexceptions.ExceptionType1,
) *endpointsIDlEndpointsWithexceptionsWithexceptions.EndpointExceptionType1 {
	// TODO: Add error fields mapping here.
	serverError := &endpointsIDlEndpointsWithexceptionsWithexceptions.EndpointExceptionType1{}
	return serverError
}
func convertFunc1E2(
	clientError *clientsIDlClientsWithexceptionsWithexceptions.ExceptionType2,
) *endpointsIDlEndpointsWithexceptionsWithexceptions.EndpointExceptionType2 {
	// TODO: Add error fields mapping here.
	serverError := &endpointsIDlEndpointsWithexceptionsWithexceptions.EndpointExceptionType2{}
	return serverError
}

func convertWithExceptionsFunc1ClientResponse(in *clientsIDlClientsWithexceptionsWithexceptions.Response) *endpointsIDlEndpointsWithexceptionsWithexceptions.Response {
	out := &endpointsIDlEndpointsWithexceptionsWithexceptions.Response{}

	return out
}
