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
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	zanzibar "github.com/uber/zanzibar/runtime"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"

	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
)

func TestInvalidStatusCode(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpointContext(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				res.WriteJSONBytes(999, nil, []byte("true"))
			},
		),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.Status, "999 status code 999")
	assert.Equal(t, resp.StatusCode, 999)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "true", string(bytes))

	logLines := bgateway.Logs("error", "Could not emit statusCode metric")

	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))

	lineStruct := logLines[0]
	code := lineStruct["UnknownStatusCode"].(float64)
	assert.Equal(t, 999.0, code)
}

func TestCallingWriteJSONWithNil(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpointContext(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				res.WriteJSON(200, nil, nil)
			},
		),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.Status, "500 Internal Server Error")
	assert.Equal(t, resp.StatusCode, 500)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not serialize json response"}`,
		string(bytes),
	)

	logLines := bgateway.Logs("error", "Could not serialize nil pointer body")

	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))
}

type failingJsonObj struct {
}

func (f failingJsonObj) MarshalJSON() ([]byte, error) {
	return nil, errors.New("cannot serialize")
}

func TestCallWriteJSONWithBadJSON(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpointContext(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				res.WriteJSON(200, nil, failingJsonObj{})
			},
		),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.Status, "500 Internal Server Error")
	assert.Equal(t, resp.StatusCode, 500)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not serialize json response"}`,
		string(bytes),
	)

	logLines := bgateway.Logs("error", "Could not serialize json response")

	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))

	lineStruct := logLines[0]
	errorText := lineStruct["error"].(string)
	assert.Equal(t, "cannot serialize", errorText)
}

//easyjson:json
type MyBodyClient struct {
	Token string
}

//easyjson:json
type MyBody struct {
	Client MyBodyClient
	Token  string
}

func TestResponsePeekBody(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpointContext(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				res.WriteJSON(200, nil, &MyBody{
					Token: "myToken",
					Client: MyBodyClient{
						Token: "myClientToken",
					},
				})

				value, vType, err := res.PeekBody("Token")
				assert.NoError(t, err, "do not expect error")
				assert.Equal(t, []byte(`myToken`), value)
				assert.Equal(t, vType, jsonparser.String)

				value, vType, err = res.PeekBody("Client", "Token")
				assert.NoError(t, err, "do not expect error")
				assert.Equal(t, []byte(`myClientToken`), value)
				assert.Equal(t, vType, jsonparser.String)
			},
		),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.StatusCode, 200)
	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(
		t,
		`{"Client":{"Token":"myClientToken"},"Token":"myToken"}`,
		string(bytes),
	)
}

func TestResponseSetHeaders(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)

	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	headers := zanzibar.ServerHTTPHeader{}
	headers.Set("foo", "bar")

	bgateway := gateway.(*benchGateway.BenchGateway)
	err = bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/foo", zanzibar.NewRouterEndpointContext(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				res.WriteJSON(200, headers, &MyBody{
					Token: "myToken",
					Client: MyBodyClient{
						Token: "myClientToken",
					},
				})
			},
		),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.StatusCode, 200)
	assert.Equal(
		t,
		resp.Header.Get("foo"),
		"bar",
	)
}

func TestResponsePeekBodyError(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpointContext(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				res.WriteJSON(200, nil, &MyBody{
					Token: "myToken",
					Client: MyBodyClient{
						Token: "myClientToken",
					},
				})

				_, _, err := res.PeekBody("Token2")
				assert.Error(t, err)
				assert.Equal(t, "Key path not found", err.Error())
			},
		),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.StatusCode, 200)
	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(
		t,
		`{"Client":{"Token":"myClientToken"},"Token":"myToken"}`,
		string(bytes),
	)
}

func TestPendingResponseBody(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpointContext(
			bgateway.ActualGateway.ContextExtractor,
			bgateway.ActualGateway.ContextMetrics,
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.Tracer,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				obj := &MyBody{
					Token: "myToken",
					Client: MyBodyClient{
						Token: "myClientToken",
					},
				}
				bytes, err := json.Marshal(obj)
				statusCode := 200
				assert.NoError(t, err)
				res.WriteJSON(statusCode, nil, obj)

				pendingBytes, pendingStatusCode := res.GetPendingResponse()
				assert.Equal(t, bytes, pendingBytes)
				assert.Equal(t, statusCode, pendingStatusCode)

			},
		),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.StatusCode, 200)
	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(
		t,
		`{"Client":{"Token":"myClientToken"},"Token":"myToken"}`,
		string(bytes),
	)
}
