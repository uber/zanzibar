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

package zanzibar_test

import (
	"bytes"
	"context"
	"testing"

	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	zanzibar "github.com/uber/zanzibar/runtime"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"

	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
)

func TestTrailingSlashRoutes(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)
	routerEndpoint := zanzibar.NewRouterEndpoint(
		bgateway.ActualGateway.ContextExtractor,
		bgateway.ActualGateway.ContextMetrics,
		bgateway.ActualGateway.Logger,
		bgateway.ActualGateway.Tracer,
		"foo", "foo",
		func(
			ctx context.Context,
			req *zanzibar.ServerHTTPRequest,
			resp *zanzibar.ServerHTTPResponse,
		) {
			resp.WriteJSONBytes(200, nil, []byte("foo\n"))
		},
	)

	err = bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/foo",
		routerEndpoint,
	)
	assert.NoError(t, err)
	err = bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/bar/",
		zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"bar", "bar",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				resp *zanzibar.ServerHTTPResponse,
			) {
				resp.WriteJSONBytes(200, nil, []byte("bar\n"))
			},
		),
	)
	assert.NoError(t, err)

	testRequests := []struct {
		url      string
		expected string
	}{
		{"/foo", "foo\n"},
		{"/foo/", "foo\n"},
		{"/bar", "bar\n"},
		{"/bar/", "bar\n"},
	}

	for _, testReq := range testRequests {
		resp, err := gateway.MakeRequest(
			"GET",
			testReq.url,
			nil,
			bytes.NewReader([]byte("{\"baz\":\"bat\"}")),
		)
		if !assert.NoError(t, err) {
			return
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, []byte(testReq.expected), bytes)
		assert.Equal(t, 1, len(
			bgateway.Logs("info", "Finished an incoming server HTTP request"),
		))
	}
}

func TestRouterNotFound(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.Status, "404 Not Found")
	assert.Equal(t, resp.StatusCode, 404)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, bytes, []byte("404 page not found\n"))
	assert.Equal(t, 1, len(
		gateway.Logs("info", "Finished an incoming server HTTP request"),
	))
}

func TestRouterInvalidMethod(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	resp, err := gateway.MakeRequest("POST", "/health", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.Status, "405 Method Not Allowed")
	assert.Equal(t, resp.StatusCode, 405)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, bytes, []byte("Method Not Allowed\n"))
	assert.Equal(t, 1, len(
		gateway.Logs("info", "Finished an incoming server HTTP request"),
	))
}

func TestRouterPanic(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	err = bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/panic",
		zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"panic", "panic",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				resp *zanzibar.ServerHTTPResponse,
			) {
				panic("a string")
			},
		),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/panic", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.Status, "500 Internal Server Error")
	assert.Equal(t, resp.StatusCode, 500)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, bytes, []byte("Internal Server Error\n"))

	allLogs := gateway.AllLogs()
	assert.Equal(t, 1, len(allLogs))

	logLines := allLogs["A http request handler paniced"]
	assert.Equal(t, 1, len(logLines))

	line := logLines[0]
	assert.Equal(t, "http router panic: a string", line["error"])
	assert.Equal(t, "/panic", line["pathname"])
	assert.Contains(
		t,
		line["errorVerbose"],
		"runtime_test.TestRouterPanic.func1",
	)
}

func TestRouterPanicObject(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	err = bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/panic",
		zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"panic", "panic",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				resp *zanzibar.ServerHTTPResponse,
			) {
				panic(errors.New("an error"))
			},
		),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/panic", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.Status, "500 Internal Server Error")
	assert.Equal(t, resp.StatusCode, 500)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, bytes, []byte("Internal Server Error\n"))

	allLogs := gateway.AllLogs()
	assert.Equal(t, 1, len(allLogs))

	logLines := allLogs["A http request handler paniced"]
	assert.Equal(t, 1, len(logLines))

	line := logLines[0]
	assert.Equal(t, "an error", line["error"])
	assert.Equal(t, "/panic", line["pathname"])
	assert.Contains(
		t,
		line["errorVerbose"],
		"runtime_test.TestRouterPanicObject.func1",
	)
}

func TestRouterPanicNilPointer(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	err = bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/panic",
		zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"panic", "panic",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				resp *zanzibar.ServerHTTPResponse,
			) {
				var foo *string = nil
				_ = *foo
			},
		),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/panic", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.Status, "500 Internal Server Error")
	assert.Equal(t, resp.StatusCode, 500)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, bytes, []byte("Internal Server Error\n"))

	allLogs := gateway.AllLogs()
	assert.Equal(t, 1, len(allLogs))

	logLines := allLogs["A http request handler paniced"]
	assert.Equal(t, 1, len(logLines))

	line := logLines[0]
	assert.Equal(t,
		"wrapped: runtime error: "+
			"invalid memory address or nil pointer dereference",
		line["error"],
	)
	assert.Equal(t, "/panic", line["pathname"])
	assert.Contains(
		t,
		line["errorVerbose"],
		"runtime_test.TestRouterPanicNilPointer.func1",
	)
}
func TestConflictingRoutes(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	err = bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/foo",
		zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				resp *zanzibar.ServerHTTPResponse,
			) {
				resp.WriteJSONBytes(200, nil, []byte("foo\n"))
			},
		),
	)
	assert.Nil(t, err)
	err = bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/foo",
		zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"bar", "bar",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				resp *zanzibar.ServerHTTPResponse,
			) {
				resp.WriteJSONBytes(200, nil, []byte("bar\n"))
			},
		),
	)
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), "handler for 'GET /foo' is already registered")
	}
}
