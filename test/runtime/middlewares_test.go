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

package runtime_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	// TODO(sindelar): Refactor into a unit test and remove the
	// example middleware (which creates a cyclic dependency)
	"github.com/uber/zanzibar/examples/example-gateway/middlewares/example"

	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
)

// helper setup functions
func setUpRequest() (context.Context, *zanzibar.ServerHTTPRequest, *zanzibar.ServerHTTPResponse) {
	responseWriter := httptest.NewRecorder()

	httpReq, _ := http.NewRequest(
		"POST",
		"/googlenow/add-credentials",
		bytes.NewReader([]byte("{\"foo\":\"bar\"}")))

	request := zanzibar.NewServerHTTPRequest(
		responseWriter,
		httpReq,
		nil, // Params
		nil, // Endpoint
	)

	response := zanzibar.NewServerHTTPResponse(responseWriter,
		request)
	ctx := context.Background()
	return ctx, request, response
}

// Ensures that a middleware stack can correctly return all of its handlers.
func TestHandlers(t *testing.T) {

	middlewareStack := zanzibar.NewStack()
	handlers := middlewareStack.Handlers()
	assert.Equal(t, 0, len(handlers))

	// Test unnammed anonymous middleware
	middlewareStack.UseFunc(func(
		ctx context.Context,
		req *zanzibar.ServerHTTPRequest,
		res *zanzibar.ServerHTTPResponse,
		next zanzibar.HandlerFn) {
		next(ctx, req, res)
		res.StatusCode = http.StatusOK
	})

	ex := example.NewMiddleWare(
		nil, // *zanzibar.Gateway
		example.Options{
			Foo: "foo",
			Bar: 2,
		},
		noop,
	)
	middlewareStack.UseHandlerFn(ex)

	// Verify the custom middleware has been added.
	handlers = middlewareStack.Handlers()
	assert.Equal(t, 2, len(handlers))

	// Run the zanzibar.HandleFn of composed middlewares.
	// TODO(sindelar): Refactor. We some helpers to build zanzibar
	// request/responses without setting up a backend and register.
	// Currently they require endpoints to instantiate.
	gateway, err := benchGateway.CreateGateway(nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.ActualGateway.Router.Register(
		"GET", "/foo",
		zanzibar.NewEndpoint(
			bgateway.ActualGateway,
			"foo",
			"foo",
			middlewareStack.Handle,
		),
	)
	resp, err := gateway.MakeRequest("GET", "/foo", nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

// Ensures that a middleware can read state from a middeware earlier in the stack.
func TestMiddlewareSharedStates(t *testing.T) {
	//contxt, request, response := setUpRequest()

	middlewareStack := zanzibar.NewStack()
	handlers := middlewareStack.Handlers()
	assert.Equal(t, 0, len(handlers))

	// Read from a shared state
	readFn := func(
		ctx context.Context,
		req *zanzibar.ServerHTTPRequest,
		res *zanzibar.ServerHTTPResponse) {
		sharedState := ctx.Value(example.MiddlewareStateName).(example.MiddlewareState)
		if sharedState.Baz == "test_state" {
			res.StatusCode = http.StatusOK
		}
	}

	// Write to a shared state
	ex := example.NewMiddleWare(
		nil, // nil Gateway
		example.Options{
			Foo: "test_state",
			Bar: 2,
		},
		readFn,
	)
	middlewareStack.UseHandlerFn(ex)

	// Run the zanzibar.HandleFn of composed middlewares.
	// TODO(sindelar): Refactor. We some helpers to build zanzibar
	// request/responses without setting up a backend and register.
	// Currently they require endpoints to instantiate.
	gateway, err := benchGateway.CreateGateway(nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.ActualGateway.Router.Register(
		"GET", "/foo",
		zanzibar.NewEndpoint(
			bgateway.ActualGateway,
			"foo",
			"foo",
			middlewareStack.Handle,
		),
	)
	resp, err := gateway.MakeRequest("GET", "/foo", nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

// Ensures that a middleware stack can accept Http middlewares.
func TestAdaptedHttpHandlers(t *testing.T) {
	// TODO(sindelar)
}

func noop(ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse) {
	return
}
