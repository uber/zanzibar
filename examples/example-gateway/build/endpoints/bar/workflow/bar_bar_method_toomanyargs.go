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
	clientsIDlClientsFooBaseBase "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/foo/base/base"
	clientsIDlClientsFooFoo "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/foo/foo"
	endpointsIDlEndpointsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/bar/bar"
	endpointsIDlEndpointsFooFoo "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/foo/foo"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/bar/module"
	"go.uber.org/zap"
)

// BarTooManyArgsWorkflow defines the interface for BarTooManyArgs workflow
type BarTooManyArgsWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
		r *endpointsIDlEndpointsBarBar.Bar_TooManyArgs_Args,
	) (context.Context, *endpointsIDlEndpointsBarBar.BarResponse, zanzibar.Header, error)
}

// NewBarTooManyArgsWorkflow creates a workflow
func NewBarTooManyArgsWorkflow(deps *module.Dependencies) BarTooManyArgsWorkflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.bar.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.bar.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &barTooManyArgsWorkflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
		defaultDeps:               deps.Default,
		errorBuilder:              zanzibar.NewErrorBuilder("endpoint", "bar"),
	}
}

// barTooManyArgsWorkflow calls thrift client Bar.TooManyArgs
type barTooManyArgsWorkflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
	defaultDeps               *zanzibar.DefaultDependencies
	errorBuilder              zanzibar.ErrorBuilder
}

// Handle calls thrift client.
func (w barTooManyArgsWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsIDlEndpointsBarBar.Bar_TooManyArgs_Args,
) (context.Context, *endpointsIDlEndpointsBarBar.BarResponse, zanzibar.Header, error) {
	clientRequest := convertToTooManyArgsClientRequest(r)

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
	h, ok = reqHeaders.Get("X-Token")
	if ok {
		clientHeaders["X-Token"] = h
	}
	h, ok = reqHeaders.Get("X-Uuid")
	if ok {
		clientHeaders["X-Uuid"] = h
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
	if w.defaultDeps.Config.ContainsKey("endpoints.bar.tooManyArgs.timeoutPerAttempt") {
		scaleFactor := w.defaultDeps.Config.MustGetFloat("endpoints.bar.tooManyArgs.scaleFactor")
		maxRetry := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.tooManyArgs.retryCount"))

		backOffTimeAcrossRetriesCfg := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.tooManyArgs.backOffTimeAcrossRetries"))
		timeoutPerAttemptConf := int(w.defaultDeps.Config.MustGetInt("endpoints.bar.tooManyArgs.timeoutPerAttempt"))

		timeoutAndRetryConfig = zanzibar.BuildTimeoutAndRetryConfig(int(timeoutPerAttemptConf), backOffTimeAcrossRetriesCfg, maxRetry, scaleFactor)
		ctx = zanzibar.WithTimeAndRetryOptions(ctx, timeoutAndRetryConfig)
	}

	ctx, clientRespBody, cliRespHeaders, err := w.Clients.Bar.TooManyArgs(
		ctx, clientHeaders, clientRequest,
	)

	if err != nil {
		zErr, ok := err.(zanzibar.Error)
		if ok {
			err = zErr.Unwrap()
		}
		switch errValue := err.(type) {

		case *clientsIDlClientsBarBar.BarException:
			err = convertTooManyArgsBarException(
				errValue,
			)

		case *clientsIDlClientsFooFoo.FooException:
			err = convertTooManyArgsFooException(
				errValue,
			)

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
	if cliRespHeaders != nil {
		resHeaders.Set("X-Token", cliRespHeaders["X-Token"])
	}
	if cliRespHeaders != nil {
		resHeaders.Set("X-Uuid", cliRespHeaders["X-Uuid"])
	}

	response := convertBarTooManyArgsClientResponse(clientRespBody)
	if val, ok := cliRespHeaders[zanzibar.ClientResponseDurationKey]; ok {
		resHeaders.Set(zanzibar.ClientResponseDurationKey, val)
	}

	resHeaders.Set(zanzibar.ClientTypeKey, "http")
	return ctx, response, resHeaders, nil
}

func convertToTooManyArgsClientRequest(in *endpointsIDlEndpointsBarBar.Bar_TooManyArgs_Args) *clientsIDlClientsBarBar.Bar_TooManyArgs_Args {
	out := &clientsIDlClientsBarBar.Bar_TooManyArgs_Args{}

	if in.Request != nil {
		out.Request = &clientsIDlClientsBarBar.BarRequest{}
		out.Request.StringField = string(in.Request.StringField)
		out.Request.BoolField = bool(in.Request.BoolField)
		out.Request.BinaryField = []byte(in.Request.BinaryField)
		out.Request.Timestamp = clientsIDlClientsBarBar.Timestamp(in.Request.Timestamp)
		out.Request.EnumField = clientsIDlClientsBarBar.Fruit(in.Request.EnumField)
		out.Request.LongField = clientsIDlClientsBarBar.Long(in.Request.LongField)
	} else {
		out.Request = nil
	}
	if in.Foo != nil {
		out.Foo = &clientsIDlClientsFooFoo.FooStruct{}
		out.Foo.FooString = string(in.Foo.FooString)
		out.Foo.FooI32 = (*int32)(in.Foo.FooI32)
		out.Foo.FooI16 = (*int16)(in.Foo.FooI16)
		out.Foo.FooDouble = (*float64)(in.Foo.FooDouble)
		out.Foo.FooBool = (*bool)(in.Foo.FooBool)
		out.Foo.FooMap = make(map[string]string, len(in.Foo.FooMap))
		for key1, value2 := range in.Foo.FooMap {
			out.Foo.FooMap[key1] = string(value2)
		}
		if in.Foo.Message != nil {
			out.Foo.Message = &clientsIDlClientsFooBaseBase.Message{}
			out.Foo.Message.Body = string(in.Foo.Message.Body)
		} else {
			out.Foo.Message = nil
		}
	} else {
		out.Foo = nil
	}

	return out
}

func convertTooManyArgsBarException(
	clientError *clientsIDlClientsBarBar.BarException,
) *endpointsIDlEndpointsBarBar.BarException {
	// TODO: Add error fields mapping here.
	serverError := &endpointsIDlEndpointsBarBar.BarException{}
	return serverError
}
func convertTooManyArgsFooException(
	clientError *clientsIDlClientsFooFoo.FooException,
) *endpointsIDlEndpointsFooFoo.FooException {
	// TODO: Add error fields mapping here.
	serverError := &endpointsIDlEndpointsFooFoo.FooException{}
	return serverError
}

func convertBarTooManyArgsClientResponse(in *clientsIDlClientsBarBar.BarResponse) *endpointsIDlEndpointsBarBar.BarResponse {
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
