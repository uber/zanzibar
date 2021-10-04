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

	"github.com/uber/zanzibar/v1/config"

	zanzibar "github.com/uber/zanzibar/v1/runtime"

	clientsIDlClientsBarBar "github.com/uber/zanzibar/v1/examples/example-gateway/build/gen-code/clients-idl/clients/bar/bar"
	endpointsIDlEndpointsBarBar "github.com/uber/zanzibar/v1/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/bar/bar"

	module "github.com/uber/zanzibar/v1/examples/example-gateway/build/endpoints/bar/module"
	"go.uber.org/zap"
)

// BarNormalWorkflow defines the interface for BarNormal workflow
type BarNormalWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
		r *endpointsIDlEndpointsBarBar.Bar_Normal_Args,
	) (context.Context, *endpointsIDlEndpointsBarBar.BarResponse, zanzibar.Header, error)
}

// NewBarNormalWorkflow creates a workflow
func NewBarNormalWorkflow(deps *module.Dependencies) BarNormalWorkflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.bar.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.bar.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &barNormalWorkflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
	}
}

// barNormalWorkflow calls thrift client Bar.Normal
type barNormalWorkflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
}

// Handle calls thrift client.
func (w barNormalWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsIDlEndpointsBarBar.Bar_Normal_Args,
) (context.Context, *endpointsIDlEndpointsBarBar.BarResponse, zanzibar.Header, error) {
	clientRequest := convertToNormalClientRequest(r)

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

	ctx, clientRespBody, _, err := w.Clients.Bar.Normal(
		ctx, clientHeaders, clientRequest,
	)

	if err != nil {
		switch errValue := err.(type) {

		case *clientsIDlClientsBarBar.BarException:
			serverErr := convertNormalBarException(
				errValue,
			)

			return ctx, nil, nil, serverErr

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

	response := convertBarNormalClientResponse(clientRespBody)
	return ctx, response, resHeaders, nil
}

func convertToNormalClientRequest(in *endpointsIDlEndpointsBarBar.Bar_Normal_Args) *clientsIDlClientsBarBar.Bar_Normal_Args {
	out := &clientsIDlClientsBarBar.Bar_Normal_Args{}

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

	return out
}

func convertNormalBarException(
	clientError *clientsIDlClientsBarBar.BarException,
) *endpointsIDlEndpointsBarBar.BarException {
	// TODO: Add error fields mapping here.
	serverError := &endpointsIDlEndpointsBarBar.BarException{}
	return serverError
}

func convertBarNormalClientResponse(in *clientsIDlClientsBarBar.BarResponse) *endpointsIDlEndpointsBarBar.BarResponse {
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
