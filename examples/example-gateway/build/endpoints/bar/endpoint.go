// Code generated by zanzibar
// @generated

// Copyright (c) 2017 Uber Technologies, Inc.
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

package barEndpoint

import (
	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/bar/module"
	zanzibar "github.com/uber/zanzibar/runtime"
)

// Endpoint registers a request handler on a gateway
type Endpoint interface {
	Register(*zanzibar.Gateway) error
}

// NewEndpoint returns a collection of endpoints that can be registered on
// a gateway
func NewEndpoint(g *zanzibar.Gateway, deps *module.Dependencies) Endpoint {
	return &EndpointHandlers{
		BarArgNotStructHandler:             NewBarArgNotStructHandler(g, deps),
		BarArgWithHeadersHandler:           NewBarArgWithHeadersHandler(g, deps),
		BarArgWithQueryParamsHandler:       NewBarArgWithQueryParamsHandler(g, deps),
		BarArgWithNestedQueryParamsHandler: NewBarArgWithNestedQueryParamsHandler(g, deps),
		BarArgWithQueryHeaderHandler:       NewBarArgWithQueryHeaderHandler(g, deps),
		BarArgWithParamsHandler:            NewBarArgWithParamsHandler(g, deps),
		BarArgWithManyQueryParamsHandler:   NewBarArgWithManyQueryParamsHandler(g, deps),
		BarMissingArgHandler:               NewBarMissingArgHandler(g, deps),
		BarNoRequestHandler:                NewBarNoRequestHandler(g, deps),
		BarNormalHandler:                   NewBarNormalHandler(g, deps),
		BarTooManyArgsHandler:              NewBarTooManyArgsHandler(g, deps),
	}
}

// EndpointHandlers is a collection of individual endpoint handlers
type EndpointHandlers struct {
	BarArgNotStructHandler             *BarArgNotStructHandler
	BarArgWithHeadersHandler           *BarArgWithHeadersHandler
	BarArgWithQueryParamsHandler       *BarArgWithQueryParamsHandler
	BarArgWithNestedQueryParamsHandler *BarArgWithNestedQueryParamsHandler
	BarArgWithQueryHeaderHandler       *BarArgWithQueryHeaderHandler
	BarArgWithParamsHandler            *BarArgWithParamsHandler
	BarArgWithManyQueryParamsHandler   *BarArgWithManyQueryParamsHandler
	BarMissingArgHandler               *BarMissingArgHandler
	BarNoRequestHandler                *BarNoRequestHandler
	BarNormalHandler                   *BarNormalHandler
	BarTooManyArgsHandler              *BarTooManyArgsHandler
}

// Register registers the endpoint handlers with the gateway
func (handlers *EndpointHandlers) Register(gateway *zanzibar.Gateway) error {
	err0 := handlers.BarArgNotStructHandler.Register(gateway)
	if err0 != nil {
		return err0
	}
	err1 := handlers.BarArgWithHeadersHandler.Register(gateway)
	if err1 != nil {
		return err1
	}
	err2 := handlers.BarArgWithQueryParamsHandler.Register(gateway)
	if err2 != nil {
		return err2
	}
	err3 := handlers.BarArgWithNestedQueryParamsHandler.Register(gateway)
	if err3 != nil {
		return err3
	}
	err4 := handlers.BarArgWithQueryHeaderHandler.Register(gateway)
	if err4 != nil {
		return err4
	}
	err5 := handlers.BarArgWithParamsHandler.Register(gateway)
	if err5 != nil {
		return err5
	}
	err6 := handlers.BarArgWithManyQueryParamsHandler.Register(gateway)
	if err6 != nil {
		return err6
	}
	err7 := handlers.BarMissingArgHandler.Register(gateway)
	if err7 != nil {
		return err7
	}
	err8 := handlers.BarNoRequestHandler.Register(gateway)
	if err8 != nil {
		return err8
	}
	err9 := handlers.BarNormalHandler.Register(gateway)
	if err9 != nil {
		return err9
	}
	err10 := handlers.BarTooManyArgsHandler.Register(gateway)
	if err10 != nil {
		return err10
	}
	return nil
}
