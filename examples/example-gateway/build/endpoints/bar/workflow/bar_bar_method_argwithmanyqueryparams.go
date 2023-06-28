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

// BarArgWithManyQueryParamsWorkflow defines the interface for BarArgWithManyQueryParams workflow
type BarArgWithManyQueryParamsWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
		r *endpointsIDlEndpointsBarBar.Bar_ArgWithManyQueryParams_Args,
	) (context.Context, *endpointsIDlEndpointsBarBar.BarResponse, zanzibar.Header, error)
}

// NewBarArgWithManyQueryParamsWorkflow creates a workflow
func NewBarArgWithManyQueryParamsWorkflow(deps *module.Dependencies) BarArgWithManyQueryParamsWorkflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.bar.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.bar.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &barArgWithManyQueryParamsWorkflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
		defaultDeps:               deps.Default,
		errorBuilder:              zanzibar.NewErrorBuilder("endpoint", "bar"),
	}
}

// barArgWithManyQueryParamsWorkflow calls thrift client Bar.ArgWithManyQueryParams
type barArgWithManyQueryParamsWorkflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
	defaultDeps               *zanzibar.DefaultDependencies
	errorBuilder              zanzibar.ErrorBuilder
}

// Handle calls thrift client.
func (w barArgWithManyQueryParamsWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsIDlEndpointsBarBar.Bar_ArgWithManyQueryParams_Args,
) (context.Context, *endpointsIDlEndpointsBarBar.BarResponse, zanzibar.Header, error) {
	clientRequest := convertToArgWithManyQueryParamsClientRequest(r)

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
	if w.defaultDeps.Config.ContainsKey("endpoints.bar.argWithManyQueryParams.timeoutPerAttempt") {
		scaleFactor := w.defaultDeps.Config.MustGetFloat("endpoints.bar.argWithManyQueryParams.scaleFactor")
		maxRetry := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.argWithManyQueryParams.retryCount"))

		backOffTimeAcrossRetriesCfg := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.argWithManyQueryParams.backOffTimeAcrossRetries"))
		timeoutPerAttemptConf := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.argWithManyQueryParams.timeoutPerAttempt"))

		timeoutAndRetryConfig = zanzibar.BuildTimeoutAndRetryConfig(int(timeoutPerAttemptConf), backOffTimeAcrossRetriesCfg, maxRetry, scaleFactor)
	}

	ctx, clientRespBody, cliRespHeaders, err := w.Clients.Bar.ArgWithManyQueryParams(
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
		return ctx, nil, nil, err
	}

	// Filter and map response headers from client to server response.
	resHeaders := zanzibar.ServerHTTPHeader{}

	response := convertBarArgWithManyQueryParamsClientResponse(clientRespBody)
	if val, ok := cliRespHeaders[zanzibar.ClientResponseDurationKey]; ok {
		resHeaders.Set(zanzibar.ClientResponseDurationKey, val)
	}

	resHeaders.Set(zanzibar.ClientTypeKey, "http")
	return ctx, response, resHeaders, nil
}

func convertToArgWithManyQueryParamsClientRequest(in *endpointsIDlEndpointsBarBar.Bar_ArgWithManyQueryParams_Args) *clientsIDlClientsBarBar.Bar_ArgWithManyQueryParams_Args {
	out := &clientsIDlClientsBarBar.Bar_ArgWithManyQueryParams_Args{}

	out.AStr = string(in.AStr)
	out.AnOptStr = (*string)(in.AnOptStr)
	out.ABool = bool(in.ABool)
	out.AnOptBool = (*bool)(in.AnOptBool)
	out.AInt8 = int8(in.AInt8)
	out.AnOptInt8 = (*int8)(in.AnOptInt8)
	out.AInt16 = int16(in.AInt16)
	out.AnOptInt16 = (*int16)(in.AnOptInt16)
	out.AInt32 = int32(in.AInt32)
	out.AnOptInt32 = (*int32)(in.AnOptInt32)
	out.AInt64 = int64(in.AInt64)
	out.AnOptInt64 = (*int64)(in.AnOptInt64)
	out.AFloat64 = float64(in.AFloat64)
	out.AnOptFloat64 = (*float64)(in.AnOptFloat64)
	out.AUUID = clientsIDlClientsBarBar.UUID(in.AUUID)
	out.AnOptUUID = (*clientsIDlClientsBarBar.UUID)(in.AnOptUUID)
	out.AListUUID = make([]clientsIDlClientsBarBar.UUID, len(in.AListUUID))
	for index1, value2 := range in.AListUUID {
		out.AListUUID[index1] = clientsIDlClientsBarBar.UUID(value2)
	}
	out.AnOptListUUID = make([]clientsIDlClientsBarBar.UUID, len(in.AnOptListUUID))
	for index3, value4 := range in.AnOptListUUID {
		out.AnOptListUUID[index3] = clientsIDlClientsBarBar.UUID(value4)
	}
	out.AStringList = make([]string, len(in.AStringList))
	for index5, value6 := range in.AStringList {
		out.AStringList[index5] = string(value6)
	}
	out.AnOptStringList = make([]string, len(in.AnOptStringList))
	for index7, value8 := range in.AnOptStringList {
		out.AnOptStringList[index7] = string(value8)
	}
	out.AUUIDList = make([]clientsIDlClientsBarBar.UUID, len(in.AUUIDList))
	for index9, value10 := range in.AUUIDList {
		out.AUUIDList[index9] = clientsIDlClientsBarBar.UUID(value10)
	}
	out.AnOptUUIDList = make([]clientsIDlClientsBarBar.UUID, len(in.AnOptUUIDList))
	for index11, value12 := range in.AnOptUUIDList {
		out.AnOptUUIDList[index11] = clientsIDlClientsBarBar.UUID(value12)
	}
	out.ATs = clientsIDlClientsBarBar.Timestamp(in.ATs)
	out.AnOptTs = (*clientsIDlClientsBarBar.Timestamp)(in.AnOptTs)
	out.AReqDemo = clientsIDlClientsBarBar.DemoType(in.AReqDemo)
	out.AnOptFruit = (*clientsIDlClientsBarBar.Fruit)(in.AnOptFruit)
	out.AReqFruits = make([]clientsIDlClientsBarBar.Fruit, len(in.AReqFruits))
	for index13, value14 := range in.AReqFruits {
		out.AReqFruits[index13] = clientsIDlClientsBarBar.Fruit(value14)
	}
	out.AnOptDemos = make([]clientsIDlClientsBarBar.DemoType, len(in.AnOptDemos))
	for index15, value16 := range in.AnOptDemos {
		out.AnOptDemos[index15] = clientsIDlClientsBarBar.DemoType(value16)
	}

	return out
}

func convertBarArgWithManyQueryParamsClientResponse(in *clientsIDlClientsBarBar.BarResponse) *endpointsIDlEndpointsBarBar.BarResponse {
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
