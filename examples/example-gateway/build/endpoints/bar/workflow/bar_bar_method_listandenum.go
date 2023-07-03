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

// BarListAndEnumWorkflow defines the interface for BarListAndEnum workflow
type BarListAndEnumWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
		r *endpointsIDlEndpointsBarBar.Bar_ListAndEnum_Args,
	) (context.Context, string, zanzibar.Header, error)
}

// NewBarListAndEnumWorkflow creates a workflow
func NewBarListAndEnumWorkflow(deps *module.Dependencies) BarListAndEnumWorkflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.bar.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.bar.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &barListAndEnumWorkflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
		defaultDeps:               deps.Default,
		errorBuilder:              zanzibar.NewErrorBuilder("endpoint", "bar"),
	}
}

// barListAndEnumWorkflow calls thrift client Bar.ListAndEnum
type barListAndEnumWorkflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
	defaultDeps               *zanzibar.DefaultDependencies
	errorBuilder              zanzibar.ErrorBuilder
}

// Handle calls thrift client.
func (w barListAndEnumWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsIDlEndpointsBarBar.Bar_ListAndEnum_Args,
) (context.Context, string, zanzibar.Header, error) {
	clientRequest := convertToListAndEnumClientRequest(r)

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
	if w.defaultDeps.Config.ContainsKey("endpoints.bar.listAndEnum.timeoutPerAttempt") {
		scaleFactor := w.defaultDeps.Config.MustGetFloat("endpoints.bar.listAndEnum.scaleFactor")
		maxRetry := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.listAndEnum.retryCount"))

		backOffTimeAcrossRetriesCfg := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.listAndEnum.backOffTimeAcrossRetries"))
		timeoutPerAttemptConf := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.listAndEnum.timeoutPerAttempt"))

		timeoutAndRetryConfig = zanzibar.BuildTimeoutAndRetryConfig(int(timeoutPerAttemptConf), backOffTimeAcrossRetriesCfg, maxRetry, scaleFactor)
		ctx = zanzibar.WithTimeAndRetryOptions(ctx, timeoutAndRetryConfig)
	}

	ctx, clientRespBody, cliRespHeaders, err := w.Clients.Bar.ListAndEnum(
		ctx, clientHeaders, clientRequest,
	)

	if err != nil {
		zErr, ok := err.(zanzibar.Error)
		if ok {
			err = zErr.Unwrap()
		}
		switch errValue := err.(type) {

		case *clientsIDlClientsBarBar.BarException:
			err = convertListAndEnumBarException(
				errValue,
			)

		default:
			w.Logger.Warn("Client failure: could not make client request",
				zap.Error(errValue),
				zap.String("client", "Bar"),
			)
		}
		if zErr != nil {
			err = w.errorBuilder.Rebuild(zErr, err)
		}
		return ctx, "", nil, err
	}

	// Filter and map response headers from client to server response.
	resHeaders := zanzibar.ServerHTTPHeader{}

	response := convertBarListAndEnumClientResponse(clientRespBody)
	if val, ok := cliRespHeaders[zanzibar.ClientResponseDurationKey]; ok {
		resHeaders.Set(zanzibar.ClientResponseDurationKey, val)
	}

	resHeaders.Set(zanzibar.ClientTypeKey, "http")
	return ctx, response, resHeaders, nil
}

func convertToListAndEnumClientRequest(in *endpointsIDlEndpointsBarBar.Bar_ListAndEnum_Args) *clientsIDlClientsBarBar.Bar_ListAndEnum_Args {
	out := &clientsIDlClientsBarBar.Bar_ListAndEnum_Args{}

	out.DemoIds = make([]string, len(in.DemoIds))
	for index1, value2 := range in.DemoIds {
		out.DemoIds[index1] = string(value2)
	}
	out.DemoType = (*clientsIDlClientsBarBar.DemoType)(in.DemoType)
	out.Demos = make([]clientsIDlClientsBarBar.DemoType, len(in.Demos))
	for index3, value4 := range in.Demos {
		out.Demos[index3] = clientsIDlClientsBarBar.DemoType(value4)
	}

	return out
}

func convertListAndEnumBarException(
	clientError *clientsIDlClientsBarBar.BarException,
) *endpointsIDlEndpointsBarBar.BarException {
	// TODO: Add error fields mapping here.
	serverError := &endpointsIDlEndpointsBarBar.BarException{}
	return serverError
}

func convertBarListAndEnumClientResponse(in string) string {
	out := in

	return out
}
