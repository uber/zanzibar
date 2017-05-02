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

package zanzibar_test

import (
	"context"
	"net/http"
	"testing"

	jsonschema "github.com/mcuadros/go-jsonschema-generator"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/examples/example-gateway/middlewares/example"
	"github.com/uber/zanzibar/examples/example-gateway/middlewares/example_reader"

	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints"
	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
)

// Ensures that a middleware stack can correctly return all of its handlers.
func TestHandlers(t *testing.T) {
	ex := example.NewMiddleWare(
		nil, // *zanzibar.Gateway
		example.Options{
			Foo: "foo",
			Bar: 2,
		},
	)

	middles := []zanzibar.MiddlewareHandle{ex}
	middlewareStack := zanzibar.NewStack(middles, noopHandlerFn)

	// Verify the custom middleware has been added.
	middlewares := middlewareStack.Middlewares()
	assert.Equal(t, 1, len(middlewares))

	// Run the zanzibar.HandleFn of composed middlewares.
	// TODO(sindelar): Refactor. We some helpers to build zanzibar
	// request/responses without setting up a backend and register.
	// Currently they require endpoints to instantiate.
	gateway, err := benchGateway.CreateGateway(
		nil,
		nil,
		clients.CreateClients,
		endpoints.Register,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/foo",
		zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway,
			"foo",
			"foo",
			middlewareStack.Handle,
		),
	)
	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

type countMiddleware struct {
	name       string
	reqCounter int
	resCounter int
	reqBail    bool
	resBail    bool
}

func (c *countMiddleware) HandleRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
	shared zanzibar.SharedState,
) bool {
	c.reqCounter++
	if !c.reqBail {
		res.WriteJSONBytes(200, nil, []byte(""))
	}

	return !c.reqBail
}

func (c *countMiddleware) HandleResponse(
	ctx context.Context,
	res *zanzibar.ServerHTTPResponse,
	shared zanzibar.SharedState,
) {
	c.resCounter++
}

func (c *countMiddleware) JSONSchema() *jsonschema.Document {
	return nil
}

func (c *countMiddleware) Name() string {
	return c.name
}

// Ensures that a middleware stack can correctly return all of its handlers.
func TestMiddlewareRequestAbort(t *testing.T) {
	mid1 := &countMiddleware{
		name: "mid1",
	}
	mid2 := &countMiddleware{
		name:    "mid2",
		reqBail: true,
	}
	mid3 := &countMiddleware{
		name: "mid3",
	}

	middles := []zanzibar.MiddlewareHandle{mid1, mid2, mid3}
	middlewareStack := zanzibar.NewStack(middles, noopHandlerFn)

	// Verify the custom middleware has been added.
	middlewares := middlewareStack.Middlewares()
	assert.Equal(t, 3, len(middlewares))

	gateway, err := benchGateway.CreateGateway(
		nil,
		nil,
		clients.CreateClients,
		endpoints.Register,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/foo",
		zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway,
			"foo",
			"foo",
			middlewareStack.Handle,
		),
	)
	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	assert.Equal(t, mid1.reqCounter, 1)
	assert.Equal(t, mid1.resCounter, 1)
	assert.Equal(t, mid2.reqCounter, 1)
	assert.Equal(t, mid2.resCounter, 1)
	assert.Equal(t, mid3.reqCounter, 0)
	assert.Equal(t, mid3.resCounter, 0)
}

// Ensures that a middleware stack can correctly return all of its handlers.
func TestMiddlewareResponseAbort(t *testing.T) {
	mid1 := &countMiddleware{
		name: "mid1",
	}
	mid2 := &countMiddleware{
		name: "mid2",
	}
	mid3 := &countMiddleware{
		name: "mid3",
	}

	middles := []zanzibar.MiddlewareHandle{mid1, mid2, mid3}
	middlewareStack := zanzibar.NewStack(middles, noopHandlerFn)

	// Verify the custom middleware has been added.
	middlewares := middlewareStack.Middlewares()
	assert.Equal(t, 3, len(middlewares))

	gateway, err := benchGateway.CreateGateway(
		nil,
		nil,
		clients.CreateClients,
		endpoints.Register,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/foo",
		zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway,
			"foo",
			"foo",
			middlewareStack.Handle,
		),
	)
	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	assert.Equal(t, mid1.reqCounter, 1)
	assert.Equal(t, mid1.resCounter, 1)
	assert.Equal(t, mid2.reqCounter, 1)
	assert.Equal(t, mid2.resCounter, 1)
	assert.Equal(t, mid3.reqCounter, 1)
	assert.Equal(t, mid3.resCounter, 1)
}

// Ensures that a middleware can read state from a middeware earlier in the stack.
func TestMiddlewareSharedStates(t *testing.T) {
	ex := example.NewMiddleWare(
		nil, // nil Gateway
		example.Options{
			Foo: "test_state",
			Bar: 2,
		},
	)
	exReader := exampleReader.NewMiddleWare(
		nil, // *zanzibar.Gateway
		exampleReader.Options{
			Foo: "foo",
		},
	)

	middles := []zanzibar.MiddlewareHandle{ex, exReader}
	middlewareStack := zanzibar.NewStack(middles, noopHandlerFn)

	// Verify the custom middleware has been added.
	middlewares := middlewareStack.Middlewares()
	assert.Equal(t, 2, len(middlewares))

	// Run the zanzibar.HandleFn of composed middlewares.
	// TODO(sindelar): Refactor. We some helpers to build zanzibar
	// request/responses without setting up a backend and register.
	// Currently they require endpoints to instantiate.
	gateway, err := benchGateway.CreateGateway(
		nil,
		nil,
		clients.CreateClients,
		endpoints.Register,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/foo",
		zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway,
			"foo",
			"foo",
			middlewareStack.Handle,
		),
	)
	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func noopHandlerFn(ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
) {
	res.WriteJSONBytes(200, nil, []byte(""))
	return
}
