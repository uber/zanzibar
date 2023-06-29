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
	"strconv"

	"github.com/uber/zanzibar/config"

	zanzibar "github.com/uber/zanzibar/runtime"

	clientsIDlClientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/baz/baz"
	endpointsIDlEndpointsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/baz/baz"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/baz/module"
	"go.uber.org/zap"
)

// SimpleServiceTransHeadersTypeWorkflow defines the interface for SimpleServiceTransHeadersType workflow
type SimpleServiceTransHeadersTypeWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
		r *endpointsIDlEndpointsBazBaz.SimpleService_TransHeadersType_Args,
	) (context.Context, *endpointsIDlEndpointsBazBaz.TransHeader, zanzibar.Header, error)
}

// NewSimpleServiceTransHeadersTypeWorkflow creates a workflow
func NewSimpleServiceTransHeadersTypeWorkflow(deps *module.Dependencies) SimpleServiceTransHeadersTypeWorkflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.baz.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.baz.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &simpleServiceTransHeadersTypeWorkflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
		defaultDeps:               deps.Default,
		errorBuilder:              zanzibar.NewErrorBuilder("endpoint", "baz"),
	}
}

// simpleServiceTransHeadersTypeWorkflow calls thrift client Baz.TransHeadersType
type simpleServiceTransHeadersTypeWorkflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
	defaultDeps               *zanzibar.DefaultDependencies
	errorBuilder              zanzibar.ErrorBuilder
}

// Handle calls thrift client.
func (w simpleServiceTransHeadersTypeWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsIDlEndpointsBazBaz.SimpleService_TransHeadersType_Args,
) (context.Context, *endpointsIDlEndpointsBazBaz.TransHeader, zanzibar.Header, error) {
	clientRequest := convertToTransHeadersTypeClientRequest(r)
	clientRequest = propagateHeadersTransHeadersTypeClientRequests(clientRequest, reqHeaders)

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
	if w.defaultDeps.Config.ContainsKey("endpoints.baz.transHeadersType.timeoutPerAttempt") {
		scaleFactor := w.defaultDeps.Config.MustGetFloat("endpoints.baz.transHeadersType.scaleFactor")
		maxRetry := int(w.defaultDeps.Config.MustGetInt("endpoints.baz.transHeadersType.retryCount"))

		backOffTimeAcrossRetriesCfg := int(w.defaultDeps.Config.MustGetInt("endpoints.baz.transHeadersType.backOffTimeAcrossRetries"))
		timeoutPerAttemptConf := int(w.defaultDeps.Config.MustGetInt("endpoints.baz.transHeadersType.timeoutPerAttempt"))

		timeoutAndRetryConfig = zanzibar.BuildTimeoutAndRetryConfig(int(timeoutPerAttemptConf), backOffTimeAcrossRetriesCfg, maxRetry, scaleFactor)
	}

	ctx, clientRespBody, cliRespHeaders, err := w.Clients.Baz.TransHeadersType(
		ctx, clientHeaders, clientRequest, &timeoutAndRetryConfig,
	)

	if err != nil {
		zErr, ok := err.(zanzibar.Error)
		if ok {
			err = zErr.Unwrap()
		}
		switch errValue := err.(type) {

		case *clientsIDlClientsBazBaz.AuthErr:
			err = convertTransHeadersTypeAuthErr(
				errValue,
			)

		case *clientsIDlClientsBazBaz.OtherAuthErr:
			err = convertTransHeadersTypeOtherAuthErr(
				errValue,
			)

		default:
			w.Logger.Warn("Client failure: could not make client request",
				zap.Error(errValue),
				zap.String("client", "Baz"),
			)
		}
		if zErr != nil {
			err = w.errorBuilder.Rebuild(zErr, err)
		}
		return ctx, nil, nil, err
	}

	// Filter and map response headers from client to server response.
	resHeaders := zanzibar.ServerHTTPHeader{}

	response := convertSimpleServiceTransHeadersTypeClientResponse(clientRespBody)
	if val, ok := cliRespHeaders[zanzibar.ClientResponseDurationKey]; ok {
		resHeaders.Set(zanzibar.ClientResponseDurationKey, val)
	}

	resHeaders.Set(zanzibar.ClientTypeKey, "tchannel")
	return ctx, response, resHeaders, nil
}

func convertToTransHeadersTypeClientRequest(in *endpointsIDlEndpointsBazBaz.SimpleService_TransHeadersType_Args) *clientsIDlClientsBazBaz.SimpleService_TransHeadersType_Args {
	out := &clientsIDlClientsBazBaz.SimpleService_TransHeadersType_Args{}

	if in.Req != nil {
		out.Req = &clientsIDlClientsBazBaz.TransHeaderType{}
	} else {
		out.Req = nil
	}

	return out
}

func convertTransHeadersTypeAuthErr(
	clientError *clientsIDlClientsBazBaz.AuthErr,
) *endpointsIDlEndpointsBazBaz.AuthErr {
	// TODO: Add error fields mapping here.
	serverError := &endpointsIDlEndpointsBazBaz.AuthErr{}
	return serverError
}
func convertTransHeadersTypeOtherAuthErr(
	clientError *clientsIDlClientsBazBaz.OtherAuthErr,
) *endpointsIDlEndpointsBazBaz.OtherAuthErr {
	// TODO: Add error fields mapping here.
	serverError := &endpointsIDlEndpointsBazBaz.OtherAuthErr{}
	return serverError
}

func convertSimpleServiceTransHeadersTypeClientResponse(in *clientsIDlClientsBazBaz.TransHeaderType) *endpointsIDlEndpointsBazBaz.TransHeader {
	out := &endpointsIDlEndpointsBazBaz.TransHeader{}

	return out
}

func propagateHeadersTransHeadersTypeClientRequests(in *clientsIDlClientsBazBaz.SimpleService_TransHeadersType_Args, headers zanzibar.Header) *clientsIDlClientsBazBaz.SimpleService_TransHeadersType_Args {
	if in == nil {
		in = &clientsIDlClientsBazBaz.SimpleService_TransHeadersType_Args{}
	}
	if key, ok := headers.Get("x-boolean"); ok {
		if in.Req == nil {
			in.Req = &clientsIDlClientsBazBaz.TransHeaderType{}
		}
		if v, err := strconv.ParseBool(key); err == nil {
			in.Req.B1 = v
		}

	}
	if key, ok := headers.Get("x-float"); ok {
		if in.Req == nil {
			in.Req = &clientsIDlClientsBazBaz.TransHeaderType{}
		}
		if v, err := strconv.ParseFloat(key, 64); err == nil {
			in.Req.F3 = &v
		}

	}
	if key, ok := headers.Get("x-int"); ok {
		if in.Req == nil {
			in.Req = &clientsIDlClientsBazBaz.TransHeaderType{}
		}
		if v, err := strconv.ParseInt(key, 10, 32); err == nil {
			val := int32(v)
			in.Req.I1 = &val
		}

	}
	if key, ok := headers.Get("x-int"); ok {
		if in.Req == nil {
			in.Req = &clientsIDlClientsBazBaz.TransHeaderType{}
		}
		if v, err := strconv.ParseInt(key, 10, 64); err == nil {
			in.Req.I2 = v
		}

	}
	if key, ok := headers.Get("x-string"); ok {
		if in.Req == nil {
			in.Req = &clientsIDlClientsBazBaz.TransHeaderType{}
		}
		in.Req.S6 = key

	}
	if key, ok := headers.Get("x-string"); ok {
		if in.Req == nil {
			in.Req = &clientsIDlClientsBazBaz.TransHeaderType{}
		}
		val := clientsIDlClientsBazBaz.UUID(key)
		in.Req.U4 = val

	}
	if key, ok := headers.Get("x-string"); ok {
		if in.Req == nil {
			in.Req = &clientsIDlClientsBazBaz.TransHeaderType{}
		}
		val := clientsIDlClientsBazBaz.UUID(key)
		in.Req.U5 = &val

	}
	return in
}
