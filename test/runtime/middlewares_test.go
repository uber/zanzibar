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
	"github.com/uber/zanzibar/examples/example-gateway/middlewares/example"
	zanzibar "github.com/uber/zanzibar/runtime"
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
	contxt, request, response := setUpRequest()

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
		nil, // nil Gateway
		example.Options{
			Foo: "foo",
			Bar: 2,
		},
		nil,
	)
	middlewareStack.UseHandlerFn(ex)

	// Verify the custom middleware has been added.
	handlers = middlewareStack.Handlers()
	assert.Equal(t, 2, len(handlers))

	// Run the zanzibar.HandleFn of composed middlewares.
	middlewareStack.Handle(contxt, request, response)
	assert.Equal(t, response.StatusCode, http.StatusOK)
}

// Ensures that a middleware can read state from a middeware earlier in the stack.
func TestMiddlewareSharedStates(t *testing.T) {
	contxt, request, response := setUpRequest()

	middlewareStack := zanzibar.NewStack()
	handlers := middlewareStack.Handlers()
	assert.Equal(t, 0, len(handlers))

	// Write to a shared state
	ex := example.NewMiddleWare(
		nil, // nil Gateway
		example.Options{
			Foo: "test_state",
			Bar: 2,
		},
		nil,
	)
	middlewareStack.UseHandlerFn(ex)

	// Read from a shared state
	middlewareStack.UseFunc(func(
		ctx context.Context,
		req *zanzibar.ServerHTTPRequest,
		res *zanzibar.ServerHTTPResponse,
		next zanzibar.HandlerFn) {
		next(ctx, req, res)
		sharedState := ctx.Value(example.MiddlewareStateName).(example.MiddlewareState)
		if sharedState.Baz == "test_state" {
			res.StatusCode = http.StatusOK
		}
	})

	// Run the zanzibar.HandleFn of composed middlewares.
	middlewareStack.Handle(contxt, request, response)
	assert.Equal(t, response.StatusCode, http.StatusOK)

}

// Ensures that a middleware stack can accept Http middlewares.
func TestAdaptedHttpHandlers(t *testing.T) {
	// TODO(sindelar)
}
