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
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"

	"github.com/buger/jsonparser"
	"github.com/stretchr/testify/assert"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
	"github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

func TestInvalidReadAndUnmarshalBody(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		&testGateway.Options{
			LogWhitelist: map[string]bool{
				"Could not read request body": true,
			},
			KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
			KnownTChannelBackends: []string{"baz"},
			ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		},
		exampleGateway.CreateGateway,
	)

	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)
	endpoint := zanzibar.NewRouterEndpoint(
		bgateway.ActualGateway.Logger,
		bgateway.ActualGateway.AllHostScope,
		"foo", "foo",
		func(
			ctx context.Context,
			req *zanzibar.ServerHTTPRequest,
			res *zanzibar.ServerHTTPResponse,
		) {
			res.WriteJSON(200, nil, nil)
		},
	)

	httpReq, _ := http.NewRequest(
		"GET",
		"/health-check",
		&corruptReader{},
	)

	req := zanzibar.NewServerHTTPRequest(
		httptest.NewRecorder(),
		httpReq,
		nil,
		endpoint,
	)
	dJ := &dummyJson{}
	assert.False(t, req.ReadAndUnmarshalBody(dJ))

	logLines := gateway.Logs("error", "Could not read request body")
	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))
}

type dummyJson struct {
	foo string
}

func (d *dummyJson) UnmarshalJSON(b []byte) (err error) {
	return errors.New("Failed to unmarshal")
}

type corruptReader struct{}

func (c *corruptReader) Read(b []byte) (n int, err error) {
	return 0, errors.New("Failed to read body")
}

func TestDoubleParseQueryValues(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				foo1, _ := req.GetQueryValue("foo")
				assert.Equal(t, "", foo1)

				foo2, _ := req.GetQueryValue("foo")
				assert.Equal(t, "", foo2)
			},
		),
	)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(bytes),
	)

	logs := bgateway.AllLogs()
	assert.Equal(t, 2, len(logs))

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request"]))
}

func TestFailingGetQueryBool(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				foo1, _ := req.GetQueryBool("foo")
				assert.Equal(t, false, foo1)
			},
		),
	)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(bytes),
	)

	logs := bgateway.AllLogs()
	assert.Equal(t, 2, len(logs))

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request"]))
}

func TestFailingGetQueryInt8(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				foo1, _ := req.GetQueryInt8("foo")
				assert.Equal(t, int8(0), foo1)
			},
		),
	)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(bytes),
	)

	logs := bgateway.AllLogs()
	assert.Equal(t, 2, len(logs))

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request"]))
}

func TestFailingHasQueryValue(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				ok := req.HasQueryValue("foo")
				assert.Equal(t, false, ok)
			},
		),
	)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(bytes),
	)

	logs := bgateway.AllLogs()
	assert.Equal(t, 2, len(logs))

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request"]))
}

func TestFailingGetQueryInt16(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				foo1, _ := req.GetQueryInt16("foo")
				assert.Equal(t, int16(0), foo1)
			},
		),
	)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(bytes),
	)

	logs := bgateway.AllLogs()
	assert.Equal(t, 2, len(logs))

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request"]))
}

func TestFailingGetQueryInt32(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				foo1, _ := req.GetQueryInt32("foo")
				assert.Equal(t, int32(0), foo1)
			},
		),
	)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(bytes),
	)

	logs := bgateway.AllLogs()
	assert.Equal(t, 2, len(logs))

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request"]))
}

func TestFailingGetQueryInt64(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				foo1, _ := req.GetQueryInt64("foo")
				assert.Equal(t, int64(0), foo1)
			},
		),
	)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(bytes),
	)

	logs := bgateway.AllLogs()
	assert.Equal(t, 2, len(logs))

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request"]))
}

func TestFailingGetQueryFloat64(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				foo1, _ := req.GetQueryFloat64("foo")
				assert.Equal(t, float64(0), foo1)
			},
		),
	)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(bytes),
	)

	logs := bgateway.AllLogs()
	assert.Equal(t, 2, len(logs))

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request"]))
}

func TestFailingHasQueryPrefix(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				ok := req.HasQueryPrefix("foo")
				assert.Equal(t, false, ok)
			},
		),
	)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(bytes),
	)

	logs := bgateway.AllLogs()
	assert.Equal(t, 2, len(logs))

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request"]))
}

func TestFailingGetQueryValues(t *testing.T) {
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
		"GET", "/foo", zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				foo1, _ := req.GetQueryValues("foo")
				assert.Equal(t, 0, len(foo1))
			},
		),
	)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(bytes),
	)

	logs := bgateway.AllLogs()
	assert.Equal(t, 2, len(logs))

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request"]))
}

func TestGetQueryValues(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)

	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	lastQueryParam := []string{}

	bgateway := gateway.(*benchGateway.BenchGateway)
	bgateway.ActualGateway.HTTPRouter.Register(
		"GET", "/foo", zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				params, ok := req.GetQueryValues("foo")
				if !assert.Equal(t, true, ok) {
					return
				}

				lastQueryParam = params
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
			},
		),
	)

	resp, err := gateway.MakeRequest("GET", "/foo?foo=bar", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, []string{"bar"}, lastQueryParam)

	resp, err = gateway.MakeRequest("GET", "/foo?foo=baz&foo=baz2", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, []string{"baz", "baz2"}, lastQueryParam)

	resp, err = gateway.MakeRequest("GET", "/foo?bar=bar", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, []string(nil), lastQueryParam)
	assert.Equal(t, 0, len(lastQueryParam))
}

func TestPeekBody(t *testing.T) {
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
		"POST", "/foo", zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.Logger,
			bgateway.ActualGateway.AllHostScope,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				value, _, err := req.PeekBody("arg1", "b1", "c1")
				assert.Error(t, err, "do not expect error")
				assert.Nil(t, value)

				assert.Equal(t, len(req.GetRawBody()), 0)
				_, success := req.ReadAll()
				assert.NotEqual(t, len(req.GetRawBody()), 0)
				assert.True(t, success)

				_, success = req.ReadAll()
				assert.True(t, success)
				value, vType, err := req.PeekBody("arg1", "b1", "c1")
				assert.NoError(t, err, "do not expect error")
				assert.Equal(t, []byte(`result`), value)
				assert.Equal(t, vType, jsonparser.String)
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
			},
		),
	)

	resp, err := gateway.MakeRequest("POST", "/foo?foo=bar", nil, bytes.NewReader([]byte(`{"arg1":{"b1":{"c1":"result"}}}`)))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "200 OK", resp.Status)
}
