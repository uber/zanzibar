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
	"bytes"
	"context"
	"testing"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/bench_gateway"

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

	bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/foo",
		zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
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
	bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/bar/",
		zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
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
