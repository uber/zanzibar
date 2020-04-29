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

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/googlenow/module"
	"go.uber.org/zap"
)

// GoogleNowCheckCredentialsWorkflow defines the interface for GoogleNowCheckCredentials workflow
type GoogleNowCheckCredentialsWorkflow interface {
	Handle(
		ctx context.Context,
		reqHeaders zanzibar.Header,
	) (zanzibar.Header, error)
}

// NewGoogleNowCheckCredentialsWorkflow creates a workflow
func NewGoogleNowCheckCredentialsWorkflow(deps *module.Dependencies) GoogleNowCheckCredentialsWorkflow {
	var whitelistedDynamicHeaders []string
	if deps.Default.Config.ContainsKey("clients.google-now.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.google-now.alternates", &alternateServiceDetail)
		for _, routingConfig := range alternateServiceDetail.RoutingConfigs {
			whitelistedDynamicHeaders = append(whitelistedDynamicHeaders, textproto.CanonicalMIMEHeaderKey(routingConfig.HeaderName))
		}
	}

	return &googleNowCheckCredentialsWorkflow{
		Clients:                   deps.Client,
		Logger:                    deps.Default.Logger,
		whitelistedDynamicHeaders: whitelistedDynamicHeaders,
	}
}

// googleNowCheckCredentialsWorkflow calls thrift client GoogleNow.CheckCredentials
type googleNowCheckCredentialsWorkflow struct {
	Clients                   *module.ClientDependencies
	Logger                    *zap.Logger
	whitelistedDynamicHeaders []string
}

// Handle calls thrift client.
func (w googleNowCheckCredentialsWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
) (zanzibar.Header, error) {

	clientHeaders := map[string]string{}

	var ok bool
	var h string

	h, ok = reqHeaders.Get("x-uber-foo")
	if ok {
		clientHeaders["x-uber-foo"] = h
	}
	h, ok = reqHeaders.Get("x-uber-bar")
	if ok {
		clientHeaders["x-uber-bar"] = h
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

	cliRespHeaders, err := w.Clients.GoogleNow.CheckCredentials(ctx, clientHeaders)

	if err != nil {
		switch errValue := err.(type) {

		default:
			w.Logger.Warn("Client failure: could not make client request",
				zap.Error(errValue),
				zap.String("client", "GoogleNow"),
			)

			return nil, err

		}
	}

	// Filter and map response headers from client to server response.
	resHeaders := zanzibar.ServerHTTPHeader{}
	if cliRespHeaders != nil {
		resHeaders.Set("X-Uuid", cliRespHeaders["X-Uuid"])
	}

	return resHeaders, nil
}
