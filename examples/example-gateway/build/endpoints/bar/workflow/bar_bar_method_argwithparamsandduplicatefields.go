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

// BarArgWithParamsAndDuplicateFieldsWorkflow defines the interface for BarArgWithParamsAndDuplicateFields workflow
type BarArgWithParamsAndDuplicateFieldsWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
		r *endpointsIDlEndpointsBarBar.Bar_ArgWithParamsAndDuplicateFields_Args,
	) (context.Context, *endpointsIDlEndpointsBarBar.BarResponse, zanzibar.Header, error)
}

// NewBarArgWithParamsAndDuplicateFieldsWorkflow creates a workflow
func NewBarArgWithParamsAndDuplicateFieldsWorkflow(deps *module.Dependencies) BarArgWithParamsAndDuplicateFieldsWorkflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.bar.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.bar.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &barArgWithParamsAndDuplicateFieldsWorkflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
		defaultDeps:               deps.Default,
	}
}

// barArgWithParamsAndDuplicateFieldsWorkflow calls thrift client Bar.ArgWithParamsAndDuplicateFields
type barArgWithParamsAndDuplicateFieldsWorkflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
	defaultDeps               *zanzibar.DefaultDependencies
}

// Handle calls thrift client.
func (w barArgWithParamsAndDuplicateFieldsWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsIDlEndpointsBarBar.Bar_ArgWithParamsAndDuplicateFields_Args,
) (context.Context, *endpointsIDlEndpointsBarBar.BarResponse, zanzibar.Header, error) {
	clientRequest := convertToArgWithParamsAndDuplicateFieldsClientRequest(r)

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
	if w.defaultDeps.Config.ContainsKey("endpoints.bar.argWithParamsAndDuplicateFields.timeoutPerAttempt") {
		scaleFactor := w.defaultDeps.Config.MustGetFloat("endpoints.bar.argWithParamsAndDuplicateFields.scaleFactor")
		maxRetry := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.argWithParamsAndDuplicateFields.retryCount"))

		backOffTimeAcrossRetriesCfg := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.argWithParamsAndDuplicateFields.backOffTimeAcrossRetries"))
		timeoutPerAttemptConf := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.argWithParamsAndDuplicateFields.timeoutPerAttempt"))

		timeoutAndRetryConfig = zanzibar.BuildTimeoutAndRetryConfig(int(timeoutPerAttemptConf), backOffTimeAcrossRetriesCfg, maxRetry, scaleFactor)
		ctx = zanzibar.WithTimeAndRetryOptions(ctx, timeoutAndRetryConfig)
	}

	ctx, clientRespBody, cliRespHeaders, err := w.Clients.Bar.ArgWithParamsAndDuplicateFields(
		ctx, clientHeaders, clientRequest,
	)

	if err != nil {
		switch errValue := err.(type) {

		default:
			w.Logger.Warn("Client failure: could not make client request",
				zap.Error(errValue),
				zap.String("client", "Bar"),
			)

			return ctx, nil, nil, err

		}
	}

	// Filter and map response headers from client to server response.
	resHeaders := zanzibar.ServerHTTPHeader{}

	response := convertBarArgWithParamsAndDuplicateFieldsClientResponse(clientRespBody)
	if val, ok := cliRespHeaders[zanzibar.ClientResponseDurationKey]; ok {
		resHeaders.Set(zanzibar.ClientResponseDurationKey, val)
	}

	resHeaders.Set(zanzibar.ClientTypeKey, "http")
	return ctx, response, resHeaders, nil
}

func convertToArgWithParamsAndDuplicateFieldsClientRequest(in *endpointsIDlEndpointsBarBar.Bar_ArgWithParamsAndDuplicateFields_Args) *clientsIDlClientsBarBar.Bar_ArgWithParamsAndDuplicateFields_Args {
	out := &clientsIDlClientsBarBar.Bar_ArgWithParamsAndDuplicateFields_Args{}

	if in.Request != nil {
		out.Request = &clientsIDlClientsBarBar.RequestWithDuplicateType{}
		if in.Request.Request1 != nil {
			out.Request.Request1 = &clientsIDlClientsBarBar.BarRequest{}
			out.Request.Request1.StringField = string(in.Request.Request1.StringField)
			out.Request.Request1.BoolField = bool(in.Request.Request1.BoolField)
			out.Request.Request1.BinaryField = []byte(in.Request.Request1.BinaryField)
			out.Request.Request1.Timestamp = clientsIDlClientsBarBar.Timestamp(in.Request.Request1.Timestamp)
			out.Request.Request1.EnumField = clientsIDlClientsBarBar.Fruit(in.Request.Request1.EnumField)
			out.Request.Request1.LongField = clientsIDlClientsBarBar.Long(in.Request.Request1.LongField)
		} else {
			out.Request.Request1 = nil
		}
		if in.Request.Request2 != nil {
			out.Request.Request2 = &clientsIDlClientsBarBar.BarRequest{}
			out.Request.Request2.StringField = string(in.Request.Request2.StringField)
			out.Request.Request2.BoolField = bool(in.Request.Request2.BoolField)
			out.Request.Request2.BinaryField = []byte(in.Request.Request2.BinaryField)
			out.Request.Request2.Timestamp = clientsIDlClientsBarBar.Timestamp(in.Request.Request2.Timestamp)
			out.Request.Request2.EnumField = clientsIDlClientsBarBar.Fruit(in.Request.Request2.EnumField)
			out.Request.Request2.LongField = clientsIDlClientsBarBar.Long(in.Request.Request2.LongField)
		} else {
			out.Request.Request2 = nil
		}
	} else {
		out.Request = nil
	}
	out.EntityUUID = string(in.EntityUUID)

	return out
}

func convertBarArgWithParamsAndDuplicateFieldsClientResponse(in *clientsIDlClientsBarBar.BarResponse) *endpointsIDlEndpointsBarBar.BarResponse {
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
