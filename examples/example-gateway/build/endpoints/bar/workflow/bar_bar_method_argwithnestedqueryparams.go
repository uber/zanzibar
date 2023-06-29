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

// BarArgWithNestedQueryParamsWorkflow defines the interface for BarArgWithNestedQueryParams workflow
type BarArgWithNestedQueryParamsWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
		r *endpointsIDlEndpointsBarBar.Bar_ArgWithNestedQueryParams_Args,
	) (context.Context, *endpointsIDlEndpointsBarBar.BarResponse, zanzibar.Header, error)
}

// NewBarArgWithNestedQueryParamsWorkflow creates a workflow
func NewBarArgWithNestedQueryParamsWorkflow(deps *module.Dependencies) BarArgWithNestedQueryParamsWorkflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.bar.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.bar.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &barArgWithNestedQueryParamsWorkflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
		defaultDeps:               deps.Default,
		errorBuilder:              zanzibar.NewErrorBuilder("endpoint", "bar"),
	}
}

// barArgWithNestedQueryParamsWorkflow calls thrift client Bar.ArgWithNestedQueryParams
type barArgWithNestedQueryParamsWorkflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
	defaultDeps               *zanzibar.DefaultDependencies
	errorBuilder              zanzibar.ErrorBuilder
}

// Handle calls thrift client.
func (w barArgWithNestedQueryParamsWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsIDlEndpointsBarBar.Bar_ArgWithNestedQueryParams_Args,
) (context.Context, *endpointsIDlEndpointsBarBar.BarResponse, zanzibar.Header, error) {
	clientRequest := convertToArgWithNestedQueryParamsClientRequest(r)

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
	if w.defaultDeps.Config.ContainsKey("endpoints.bar.argWithNestedQueryParams.timeoutPerAttempt") {
		scaleFactor := w.defaultDeps.Config.MustGetFloat("endpoints.bar.argWithNestedQueryParams.scaleFactor")
		maxRetry := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.argWithNestedQueryParams.retryCount"))

		backOffTimeAcrossRetriesCfg := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.argWithNestedQueryParams.backOffTimeAcrossRetries"))
		timeoutPerAttemptConf := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.argWithNestedQueryParams.timeoutPerAttempt"))

		timeoutAndRetryConfig = zanzibar.BuildTimeoutAndRetryConfig(int(timeoutPerAttemptConf), backOffTimeAcrossRetriesCfg, maxRetry, scaleFactor)
		ctx = zanzibar.WithTimeAndRetryOptions(ctx, timeoutAndRetryConfig)
	}

	ctx, clientRespBody, cliRespHeaders, err := w.Clients.Bar.ArgWithNestedQueryParams(
		ctx, clientHeaders, clientRequest,
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
		if zErr != nil {
			err = w.errorBuilder.Rebuild(zErr, err)
		}
		return ctx, nil, nil, err
	}

	// Filter and map response headers from client to server response.
	resHeaders := zanzibar.ServerHTTPHeader{}

	response := convertBarArgWithNestedQueryParamsClientResponse(clientRespBody)
	if val, ok := cliRespHeaders[zanzibar.ClientResponseDurationKey]; ok {
		resHeaders.Set(zanzibar.ClientResponseDurationKey, val)
	}

	resHeaders.Set(zanzibar.ClientTypeKey, "http")
	return ctx, response, resHeaders, nil
}

func convertToArgWithNestedQueryParamsClientRequest(in *endpointsIDlEndpointsBarBar.Bar_ArgWithNestedQueryParams_Args) *clientsIDlClientsBarBar.Bar_ArgWithNestedQueryParams_Args {
	out := &clientsIDlClientsBarBar.Bar_ArgWithNestedQueryParams_Args{}

	if in.Request != nil {
		out.Request = &clientsIDlClientsBarBar.QueryParamsStruct{}
		out.Request.Name = string(in.Request.Name)
		out.Request.UserUUID = (*string)(in.Request.UserUUID)
		out.Request.AuthUUID = (*string)(in.Request.AuthUUID)
		out.Request.AuthUUID2 = (*string)(in.Request.AuthUUID2)
		out.Request.Foo = make([]string, len(in.Request.Foo))
		for index1, value2 := range in.Request.Foo {
			out.Request.Foo[index1] = string(value2)
		}
	} else {
		out.Request = nil
	}
	if in.Opt != nil {
		out.Opt = &clientsIDlClientsBarBar.QueryParamsOptsStruct{}
		out.Opt.Name = string(in.Opt.Name)
		out.Opt.UserUUID = (*string)(in.Opt.UserUUID)
		out.Opt.AuthUUID = (*string)(in.Opt.AuthUUID)
		out.Opt.AuthUUID2 = (*string)(in.Opt.AuthUUID2)
	} else {
		out.Opt = nil
	}

	return out
}

func convertBarArgWithNestedQueryParamsClientResponse(in *clientsIDlClientsBarBar.BarResponse) *endpointsIDlEndpointsBarBar.BarResponse {
	out := &endpointsIDlEndpointsBarBar.BarResponse{}

	out.StringField = string(in.StringField)
	out.IntWithRange = int32(in.IntWithRange)
	out.IntWithoutRange = int32(in.IntWithoutRange)
	out.MapIntWithRange = make(map[endpointsIDlEndpointsBarBar.UUID]int32, len(in.MapIntWithRange))
	for key1, value2 := range in.MapIntWithRange {
		out.MapIntWithRange[endpointsIDlEndpointsBarBar.UUID(key1)] = int32(value2)
	}
	out.MapIntWithoutRange = make(map[string]int32, len(in.MapIntWithoutRange))
	for key3, value4 := range in.MapIntWithoutRange {
		out.MapIntWithoutRange[key3] = int32(value4)
	}
	out.BinaryField = []byte(in.BinaryField)
	var convertBarResponseHelper5 func(in *clientsIDlClientsBarBar.BarResponse) (out *endpointsIDlEndpointsBarBar.BarResponse)
	convertBarResponseHelper5 = func(in *clientsIDlClientsBarBar.BarResponse) (out *endpointsIDlEndpointsBarBar.BarResponse) {
		if in != nil {
			out = &endpointsIDlEndpointsBarBar.BarResponse{}
			out.StringField = string(in.StringField)
			out.IntWithRange = int32(in.IntWithRange)
			out.IntWithoutRange = int32(in.IntWithoutRange)
			out.MapIntWithRange = make(map[endpointsIDlEndpointsBarBar.UUID]int32, len(in.MapIntWithRange))
			for key6, value7 := range in.MapIntWithRange {
				out.MapIntWithRange[endpointsIDlEndpointsBarBar.UUID(key6)] = int32(value7)
			}
			out.MapIntWithoutRange = make(map[string]int32, len(in.MapIntWithoutRange))
			for key8, value9 := range in.MapIntWithoutRange {
				out.MapIntWithoutRange[key8] = int32(value9)
			}
			out.BinaryField = []byte(in.BinaryField)
			out.NextResponse = convertBarResponseHelper5(in.NextResponse)
		} else {
			out = nil
		}
		return
	}
	out.NextResponse = convertBarResponseHelper5(in.NextResponse)

	return out
}
