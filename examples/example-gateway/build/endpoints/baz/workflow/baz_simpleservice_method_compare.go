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

	clientsIDlClientsBazBase "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/baz/base"
	clientsIDlClientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/baz/baz"
	endpointsIDlEndpointsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/baz/baz"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/baz/module"
	"go.uber.org/zap"
)

// SimpleServiceCompareWorkflow defines the interface for SimpleServiceCompare workflow
type SimpleServiceCompareWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
		r *endpointsIDlEndpointsBazBaz.SimpleService_Compare_Args,
	) (*endpointsIDlEndpointsBazBaz.BazResponse, zanzibar.Header, error)
}

// NewSimpleServiceCompareWorkflow creates a workflow
func NewSimpleServiceCompareWorkflow(deps *module.Dependencies) SimpleServiceCompareWorkflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.baz.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.baz.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &simpleServiceCompareWorkflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
	}
}

// simpleServiceCompareWorkflow calls thrift client Baz.Compare
type simpleServiceCompareWorkflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
}

// Handle calls thrift client.
func (w simpleServiceCompareWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	r *endpointsIDlEndpointsBazBaz.SimpleService_Compare_Args,
) (*endpointsIDlEndpointsBazBaz.BazResponse, zanzibar.Header, error) {
	clientRequest := convertToCompareClientRequest(r)

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

	clientRespBody, _, err := w.Clients.Baz.Compare(
		ctx, clientHeaders, clientRequest,
	)

	if err != nil {
		switch errValue := err.(type) {

		case *clientsIDlClientsBazBaz.AuthErr:
			serverErr := convertCompareAuthErr(
				errValue,
			)

			return nil, nil, serverErr

		case *clientsIDlClientsBazBaz.OtherAuthErr:
			serverErr := convertCompareOtherAuthErr(
				errValue,
			)

			return nil, nil, serverErr

		default:
			w.Logger.Warn("Client failure: could not make client request",
				zap.Error(errValue),
				zap.String("client", "Baz"),
			)

			return nil, nil, err

		}
	}

	// Filter and map response headers from client to server response.
	resHeaders := zanzibar.ServerHTTPHeader{}

	response := convertSimpleServiceCompareClientResponse(clientRespBody)
	return response, resHeaders, nil
}

func convertToCompareClientRequest(in *endpointsIDlEndpointsBazBaz.SimpleService_Compare_Args) *clientsIDlClientsBazBaz.SimpleService_Compare_Args {
	out := &clientsIDlClientsBazBaz.SimpleService_Compare_Args{}

	if in.Arg1 != nil {
		out.Arg1 = &clientsIDlClientsBazBaz.BazRequest{}
		out.Arg1.B1 = bool(in.Arg1.B1)
		out.Arg1.S2 = string(in.Arg1.S2)
		out.Arg1.I3 = int32(in.Arg1.I3)
	} else {
		out.Arg1 = nil
	}
	if in.Arg2 != nil {
		out.Arg2 = &clientsIDlClientsBazBaz.BazRequest{}
		out.Arg2.B1 = bool(in.Arg2.B1)
		out.Arg2.S2 = string(in.Arg2.S2)
		out.Arg2.I3 = int32(in.Arg2.I3)
	} else {
		out.Arg2 = nil
	}

	return out
}

func convertCompareAuthErr(
	clientError *clientsIDlClientsBazBaz.AuthErr,
) *endpointsIDlEndpointsBazBaz.AuthErr {
	// TODO: Add error fields mapping here.
	serverError := &endpointsIDlEndpointsBazBaz.AuthErr{}
	return serverError
}
func convertCompareOtherAuthErr(
	clientError *clientsIDlClientsBazBaz.OtherAuthErr,
) *endpointsIDlEndpointsBazBaz.OtherAuthErr {
	// TODO: Add error fields mapping here.
	serverError := &endpointsIDlEndpointsBazBaz.OtherAuthErr{}
	return serverError
}

func convertSimpleServiceCompareClientResponse(in *clientsIDlClientsBazBase.BazResponse) *endpointsIDlEndpointsBazBaz.BazResponse {
	out := &endpointsIDlEndpointsBazBaz.BazResponse{}

	out.Message = string(in.Message)

	return out
}
