// Copyright (c) 2022 Uber Technologies, Inc.
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
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/buger/jsonparser"
	"github.com/stretchr/testify/assert"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
	zanzibar "github.com/uber/zanzibar/runtime"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

func TestInvalidReadAndUnmarshalBody(t *testing.T) {
	gateway, err := createTestGateway()
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()
	bgateway := gateway.(*benchGateway.BenchGateway)
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	endpoint := zanzibar.NewRouterEndpoint(
		bgateway.ActualGateway.ContextExtractor,
		deps,
		"foo", "foo",
		func(
			ctx context.Context,
			req *zanzibar.ServerHTTPRequest,
			res *zanzibar.ServerHTTPResponse,
		) context.Context {
			res.WriteJSON(200, nil, nil)
			return ctx
		},
	)

	httpReq, _ := http.NewRequest("GET", "/health-check", &corruptReader{})
	req := zanzibar.NewServerHTTPRequest(httptest.NewRecorder(), httpReq, nil, endpoint)
	dJ := &dummyJson{}
	assert.False(t, req.ReadAndUnmarshalBody(dJ))

	logLines := gateway.Logs("error", "Could not read request body")
	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))
}

func createTestGateway() (testGateway.TestGateway, error) {
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
	return gateway, err
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				foo1, _ := req.GetQueryValue("foo")
				assert.Equal(t, "", foo1)

				foo2, _ := req.GetQueryValue("foo")
				assert.Equal(t, "", foo2)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	logs := bgateway.AllLogs()

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request with 400 status code"]))
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				foo1, _ := req.GetQueryBool("foo")
				assert.Equal(t, false, foo1)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	logs := bgateway.AllLogs()

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request with 400 status code"]))
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				foo1, _ := req.GetQueryInt8("foo")
				assert.Equal(t, int8(0), foo1)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	logs := bgateway.AllLogs()

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request with 400 status code"]))
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				ok := req.HasQueryValue("foo")
				assert.Equal(t, false, ok)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	logs := bgateway.AllLogs()

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request with 400 status code"]))
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				foo1, _ := req.GetQueryInt16("foo")
				assert.Equal(t, int16(0), foo1)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	logs := bgateway.AllLogs()

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request with 400 status code"]))
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				foo1, _ := req.GetQueryInt32("foo")
				assert.Equal(t, int32(0), foo1)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	logs := bgateway.AllLogs()

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request with 400 status code"]))
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				foo1, _ := req.GetQueryInt64("foo")
				assert.Equal(t, int64(0), foo1)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	logs := bgateway.AllLogs()

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request with 400 status code"]))
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				foo1, _ := req.GetQueryFloat64("foo")
				assert.Equal(t, float64(0), foo1)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	logs := bgateway.AllLogs()

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request with 400 status code"]))
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				ok := req.HasQueryPrefix("foo")
				assert.Equal(t, false, ok)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	logs := bgateway.AllLogs()

	// Assert that there is only one log even though
	// we double call GetQueryValue
	assert.Equal(t, 1, len(logs["Got request with invalid query string"]))
	assert.Equal(t, 1, len(logs["Finished an incoming server HTTP request with 400 status code"]))
}

func TestGetQueryBoolList(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryBoolList("a")
				assert.True(t, ok)
				assert.Equal(t, []bool{true, true, false}, l)
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=true&a=true&a=false", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestFailingGetQueryBoolList(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryBoolList("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=truer", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err = ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestGetQueryInt8List(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt8List("a")
				assert.True(t, ok)
				sort.Slice(l, func(i, j int) bool { return l[i] < l[j] })
				assert.Equal(t, []int8{42, 49}, l)
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42&a=49", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestFailingGetQueryInt8List(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt8List("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42&a=49er", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err = ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestGetQueryInt8Set(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt8Set("a")
				assert.True(t, ok)
				expected := []int8{42, 49}
				sort.Slice(l, func(i, j int) bool { return l[i] < l[j] })
				assert.Equal(t, expected, l)
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	queries := []string{"a=42&a=49", "a=42&a=42&a=49"}
	for _, query := range queries {
		resp, err := mockService.MakeHTTPRequest("GET", "/foo?"+query, nil, nil)
		assert.NoError(t, err)

		assert.Equal(t, "200 OK", resp.Status)
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestFailingGetQueryInt8Set(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt8Set("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42&a=49er", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err = ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestGetQueryInt16List(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt16List("a")
				assert.True(t, ok)
				assert.Equal(t, []int16{42, 49}, l)
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42&a=49", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestFailingGetQueryInt16List(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt16List("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42&a=49er", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err = ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestGetQueryInt16Set(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt16Set("a")
				assert.True(t, ok)
				expected := []int16{42, 49}
				sort.Slice(l, func(i, j int) bool { return l[i] < l[j] })
				assert.Equal(t, expected, l)
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	queries := []string{"a=42&a=49", "a=42&a=42&a=49"}
	for _, query := range queries {
		resp, err := mockService.MakeHTTPRequest("GET", "/foo?"+query, nil, nil)
		assert.NoError(t, err)

		assert.Equal(t, "200 OK", resp.Status)
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestFailingGetQueryInt16Set(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt16Set("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42&a=49er", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err = ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestGetQueryInt32List(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt32List("a")
				assert.True(t, ok)
				assert.Equal(t, []int32{42, 49}, l)
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42&a=49", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestFailingGetQueryInt32List(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt32List("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42&a=49er", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err = ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestGetQueryInt32Set(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt32Set("a")
				assert.True(t, ok)
				expected := []int32{42, 49}
				sort.Slice(l, func(i, j int) bool { return l[i] < l[j] })
				assert.Equal(t, expected, l)
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	queries := []string{"a=42&a=49", "a=42&a=42&a=49"}
	for _, query := range queries {
		resp, err := mockService.MakeHTTPRequest("GET", "/foo?"+query, nil, nil)
		assert.NoError(t, err)

		assert.Equal(t, "200 OK", resp.Status)
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestFailingGetQueryInt32Set(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt32Set("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42&a=49er", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err = ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestGetQueryInt64List(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt64List("a")
				assert.True(t, ok)
				assert.Equal(t, []int64{42, 49}, l)
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42&a=49", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestFailingGetQueryInt64List(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt64List("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42&a=49er", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err = ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestGetQueryInt64Set(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt64Set("a")
				assert.True(t, ok)
				expected := []int64{42, 49}
				sort.Slice(l, func(i, j int) bool { return l[i] < l[j] })
				assert.Equal(t, expected, l)
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	queries := []string{"a=42&a=49", "a=42&a=42&a=49"}
	for _, query := range queries {
		resp, err := mockService.MakeHTTPRequest("GET", "/foo?"+query, nil, nil)
		assert.NoError(t, err)

		assert.Equal(t, "200 OK", resp.Status)
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestFailingGetQueryInt64Set(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryInt64Set("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42&a=49er", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err = ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestGetQueryFloat64List(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryFloat64List("a")
				assert.True(t, ok)
				assert.InEpsilonSlice(t, []float64{42.24, 49.94}, l, float64(0.005))
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42.42&a=49.94", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestFailingGetQueryFloat64List(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryFloat64List("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42.24&a=49.94er", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err = ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestGetQueryFloat64Set(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryFloat64Set("a")
				assert.True(t, ok)
				sort.Slice(l, func(i, j int) bool { return l[i] < l[j] })
				assert.InEpsilonSlice(t, []float64{42.24, 49.94}, l, float64(0.005))
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	queries := []string{"a=42.24&a=49.94", "a=42.24&a=42.24&a=49.94"}
	for _, query := range queries {
		resp, err := mockService.MakeHTTPRequest("GET", "/foo?"+query, nil, nil)
		assert.NoError(t, err)

		assert.Equal(t, "200 OK", resp.Status)
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestFailingGetQueryFloat64Set(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryFloat64Set("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?a=42.24&a=49.94er", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err = ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestFailingGetQueryValueList(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryValueList("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestGetQueryValueList(t *testing.T) {
	lastQueryParam := []string{}

	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				params, ok := req.GetQueryValueList("foo")
				if !assert.Equal(t, true, ok) {
					return ctx
				}

				lastQueryParam = params
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?foo=bar", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, []string{"bar"}, lastQueryParam)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?foo=baz&foo=baz2", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, []string{"baz", "baz2"}, lastQueryParam)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?bar=bar", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, []string(nil), lastQueryParam)
	assert.Equal(t, 0, len(lastQueryParam))
}

func TestGetQueryValueSet(t *testing.T) {
	lastQueryParam := []string{}

	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				params, ok := req.GetQueryValueSet("foo")
				if !assert.True(t, ok) {
					return ctx
				}
				lastQueryParam = params
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?foo=bar", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, []string{"bar"}, lastQueryParam)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?foo=baz&foo=baz2&foo=baz", nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "200 OK", resp.Status)
	sort.Strings(lastQueryParam)
	assert.Equal(t, []string{"baz", "baz2"}, lastQueryParam)

	resp, err = mockService.MakeHTTPRequest("GET", "/foo?bar=bar", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, []string{}, lastQueryParam)
	assert.Equal(t, 0, len(lastQueryParam))
}

func TestFailingGetQueryValueSet(t *testing.T) {
	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				l, ok := req.GetQueryValueSet("a")
				assert.False(t, ok)
				assert.Nil(t, l)
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := mockService.MakeHTTPRequest("GET", "/foo?%gh&%ij", nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "400 Bad Request", resp.Status)
	assert.Equal(t, 400, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t,
		`{"error":"Could not parse query string"}`,
		string(respBytes),
	)
}

func TestSetQueryValue(t *testing.T) {
	lastQueryParam := []string{}

	mockService := ms.MustCreateTestService(t)
	mockService.Start()
	defer mockService.Stop()

	g := mockService.Server()
	deps := &zanzibar.DefaultDependencies{
		Scope:         g.RootScope,
		Logger:        g.Logger,
		ContextLogger: g.ContextLogger,
		Tracer:        g.Tracer,
	}
	err := g.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			g.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				req.SetQueryValue("foo", "changed")
				params, ok := req.GetQueryValues("foo")
				if !assert.Equal(t, true, ok) {
					return ctx
				}
				lastQueryParam = params
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	for _, query := range []string{"", "?foo=bar", "?foo=baz&foo=baz2", "?bar=bar"} {
		resp, err := mockService.MakeHTTPRequest("GET", "/foo"+query, nil, nil)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, "200 OK", resp.Status)
		assert.Equal(t, []string{"changed"}, lastQueryParam)
	}
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"POST", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
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
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("POST", "/foo?foo=bar", nil, bytes.NewReader([]byte(`{"arg1":{"b1":{"c1":"result"}}}`)))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "200 OK", resp.Status)
}

func TestReplaceBody(t *testing.T) {
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"POST", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				_, success := req.ReadAll()
				assert.True(t, success)
				assert.NotEqual(t, 0, len(req.GetRawBody()))

				value, vType, err := req.PeekBody("arg1", "b1", "c1")
				assert.NoError(t, err, "do not expect error")
				assert.Equal(t, []byte("result"), value)
				assert.Equal(t, jsonparser.String, vType)

				req.ReplaceBody([]byte(`{"arg1":{"b2":"foo"}}`))
				_, _, err = req.PeekBody("arg1", "b1")
				assert.Error(t, err, "expected this to have been replaced already")
				value, vType, err = req.PeekBody("arg1", "b2")
				assert.NoError(t, err, "do not expect error")
				assert.Equal(t, []byte("foo"), value)
				assert.Equal(t, jsonparser.String, vType)
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("POST", "/foo?foo=bar", nil, bytes.NewReader([]byte(`{"arg1":{"b1":{"c1":"result","d1":"efg"}}}`)))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "200 OK", resp.Status)
}

func TestSpanCreated(t *testing.T) {
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"POST", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				assert.NotEmpty(t, req.StartTime(), "startTime is not initialized")
				span := req.GetSpan()
				assert.NotNil(t, span)
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	resp, err := gateway.MakeRequest("POST", "/foo?foo=bar", nil, bytes.NewReader([]byte(`{"arg1":{"b1":{"c1":"result"}}}`)))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "200 OK", resp.Status)
}

func TestGetAPIEnvironment(t *testing.T) {
	cases := []struct {
		name                   string
		config                 *zanzibar.StaticConfig
		requestAPIEnvironment  string
		expectedApiEnvironment string
	}{
		{
			name:                   "Endpoint Request with no config and no api environment header",
			expectedApiEnvironment: "production",
		},
		{
			name:                   "Endpoint Request with no config and api environment header :: is ignored",
			requestAPIEnvironment:  "sandbox",
			expectedApiEnvironment: "production",
		},
		{
			name:                   "Endpoint Request with config and matching api environment header :: is used",
			requestAPIEnvironment:  "sandbox",
			expectedApiEnvironment: "sandbox",
			config: zanzibar.NewStaticConfigOrDie(nil, map[string]interface{}{
				"apiEnvironmentHeader": "x-api-environment",
			}),
		},
		{
			name:                   "Endpoint Request with config and non matching api environment header :: is ignored",
			requestAPIEnvironment:  "sandbox",
			expectedApiEnvironment: "production",
			config: zanzibar.NewStaticConfigOrDie(nil, map[string]interface{}{
				"apiEnvironmentHeader": "x-api-environment-mod",
			}),
		},
	}

	for _, tc := range cases {
		deps := &zanzibar.DefaultDependencies{Config: tc.config}
		endpoint := zanzibar.NewRouterEndpoint(nil, deps, "", "", nil)
		request := httptest.NewRequest(http.MethodGet, "/health", nil)
		request.Header.Add("x-api-environment", tc.requestAPIEnvironment)
		environment := zanzibar.GetAPIEnvironment(endpoint, request)
		assert.Equal(t, tc.expectedApiEnvironment, environment)
	}
}

func testIncomingHTTPRequestServerLog(t *testing.T, isShadowRequest bool, environment string) {
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
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
		Config:        bgateway.ActualGateway.Config,
	}
	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo", http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) context.Context {
				res.WriteJSONBytes(200, nil, []byte(`{"ok":true}`))
				return ctx
			},
		).HandleRequest),
	)
	assert.NoError(t, err)

	var headers map[string]string
	if isShadowRequest {
		headers = map[string]string{
			"X-Shadow-Request": "true",
		}
	}
	_, err = gateway.MakeRequest("GET", "/foo?bar=bar", headers, nil)
	assert.NoError(t, err)

	allLogs := bgateway.AllLogs()
	assert.Equal(t, 1, len(allLogs["Finished an incoming server HTTP request with 200 status code"]))

	tags := allLogs["Finished an incoming server HTTP request with 200 status code"][0]

	dynamicHeaders := []string{
		"requestUUID",
		"remoteAddr",
		"timestamp-started",
		"ts",
		"hostname",
		"host",
		"pid",
		"timestamp-finished",
		zanzibar.TraceIDKey,
		zanzibar.TraceSpanKey,
		zanzibar.TraceSampledKey,
	}

	if isShadowRequest {
		dynamicHeaders = append(dynamicHeaders, "X-Shadow-Request")
	}

	for _, dynamicValue := range dynamicHeaders {
		assert.Contains(t, tags, dynamicValue)
		delete(tags, dynamicValue)
	}

	expectedValues := map[string]interface{}{
		"msg":             "Finished an incoming server HTTP request with 200 status code",
		"env":             environment,
		"level":           "debug",
		"zone":            "unknown",
		"service":         "example-gateway",
		"method":          "GET",
		"pathname":        "/foo?bar=bar",
		"statusCode":      float64(200),
		"endpointHandler": "foo",
		"endpointID":      "foo",
		"url":             "/foo",

		"Accept-Encoding":         "gzip",
		"User-Agent":              "Go-http-client/1.1",
		"Res-Header-Content-Type": "application/json",
		"apienvironment":          "production",
	}
	for actualKey, actualValue := range tags {
		assert.Equal(t, expectedValues[actualKey], actualValue, "unexpected header %q", actualKey)
	}
	for expectedKey, expectedValue := range expectedValues {
		assert.Equal(t, expectedValue, tags[expectedKey], "unexpected header %q", expectedKey)
	}
}

func TestIncomingHTTPRequestServerLogForDiffRequestTypes(t *testing.T) {
	tests := []struct {
		name                string
		isShadowRequest     bool
		expectedEnvironment string
	}{
		{
			"Test incoming http request server log for normal request",
			false,
			"test",
		},
		{
			"Test incoming http request server log for shadow request",
			true,
			"shadow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testIncomingHTTPRequestServerLog(t, tt.isShadowRequest, tt.expectedEnvironment)
		})
	}
}
