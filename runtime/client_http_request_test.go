// Copyright (c) 2023 Uber Technologies, Inc.
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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	clientsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/bar/bar"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/runtime/jsonwrapper"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
	"go.uber.org/zap"
)

var defaultTestOptions = &testGateway.Options{
	KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
	KnownTChannelBackends: []string{"baz"},
	ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
}
var defaultTestConfig = map[string]interface{}{
	"clients.baz.serviceName": "baz",

	// disable circuit breaker to avoid race condition when running tests
	// the circuit breaker lib emits metrics in a free goroutine, when the
	// server closes, it attempts to close the channel for emitting metrics
	// but the circuit breaker stats report goroutine could still be running
	"clients.bar.circuitBreakerDisabled": true,
	"apiEnvironmentHeader":               "x-api-environment",
}
var retryOptions = zanzibar.TimeoutAndRetryOptions{
	OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
	RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
	MaxAttempts:                  0,
	BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
}

func TestMakingClientWriteJSONWithBadJSON(t *testing.T) {
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
	client := zanzibar.NewHTTPClientContext(
		bgateway.ActualGateway.ContextLogger,
		bgateway.ActualGateway.ContextMetrics,
		jsonwrapper.NewDefaultJSONWrapper(),
		"clientID",
		map[string]string{
			"DoStuff": "clientID::DoStuff",
		},
		"/",
		map[string]string{},
		time.Second,
		true,
	)
	ctx := zanzibar.WithTimeAndRetryOptions(context.Background(), &retryOptions)
	req := zanzibar.NewClientHTTPRequest(ctx, "clientID", "DoStuff",
		"clientID::DoStuff", client)

	err = req.WriteJSON("GET", "/foo", nil, &failingJsonObj{})
	assert.NotNil(t, err)
	assert.Equal(t,
		"Could not serialize clientID.DoStuff request object: json: error calling MarshalJSON for type *zanzibar_test.failingJsonObj: cannot serialize",
		err.Error(),
	)
}

func TestMakingClientWriteJSONWithBadHTTPMethod(t *testing.T) {
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
	client := zanzibar.NewHTTPClient(
		bgateway.ActualGateway.ContextLogger,
		bgateway.ActualGateway.RootScope,
		jsonwrapper.NewDefaultJSONWrapper(),
		"clientID",
		map[string]string{
			"DoStuff": "clientID::DoStuff",
		},
		"/",
		map[string]string{},
		time.Second,
	)
	ctx := zanzibar.WithTimeAndRetryOptions(context.Background(), &retryOptions)
	req := zanzibar.NewClientHTTPRequest(ctx, "clientID", "DoStuff",
		"clientID::DoStuff", client)

	err = req.WriteJSON("@INVALIDMETHOD", "/foo", nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t,
		"Could not create outbound clientID.DoStuff request: net/http: invalid method \"@INVALIDMETHOD\"",
		err.Error(),
	)
}

func TestMakingClientCallWithHeaders(t *testing.T) {
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

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(r.Header.Get("Example-Header")))
			// Check that the default header got set and actually sent to the server.
			assert.Equal(t, r.Header.Get("X-Client-ID"), "bar")
			assert.Equal(t, r.Header.Get("Accept"), "application/test+json")
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	barClient := deps.Client.Bar
	client := barClient.HTTPClient()

	ctx := zanzibar.WithTimeAndRetryOptions(context.Background(), &retryOptions)
	req := zanzibar.NewClientHTTPRequest(ctx, "bar", "Normal", "bar::Normal", client)

	err = req.WriteJSON(
		"POST",
		client.BaseURL+"/bar-path",
		map[string]string{
			"Example-Header": "Example-Value",
			"Accept":         "application/test+json",
		},
		nil,
	)
	assert.NoError(t, err)

	res, err := req.Do()
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	bytes, err := res.ReadAll()
	assert.NoError(t, err)
	assert.Equal(t, []byte("Example-Value"), bytes)
}

func TestMakingClientCallWithHeadersWithRequestLevelTimeoutAndRetries(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	serverProcessingTime := 700 * time.Millisecond //to mimic the server processing time, always > reqTimeout

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, _ := io.ReadAll(r.Body)
			time.Sleep(serverProcessingTime) //mimic processing time
			w.WriteHeader(200)
			response := map[string]string{"Example-Header": r.Header.Get("Example-Header"), "body": string(bodyBytes)}
			responseBytes, _ := json.Marshal(response)
			_, _ = w.Write(responseBytes)
			// Check that the default header got set and actually sent to the server.
			assert.Equal(t, r.Header.Get("X-Client-ID"), "bar")
			assert.Equal(t, r.Header.Get("Accept"), "application/test+json")
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	barClient := deps.Client.Bar
	client := barClient.HTTPClient()

	//keep reqTimeout < serverProcessingTIme
	retryOptionsCopy := retryOptions
	retryOptionsCopy.RequestTimeoutPerAttemptInMs = 500 * time.Millisecond
	retryOptionsCopy.MaxAttempts = 2

	ctx := zanzibar.WithTimeAndRetryOptions(context.Background(), &retryOptionsCopy)
	req := zanzibar.NewClientHTTPRequest(ctx, "bar", "Normal", "bar::Normal", client)

	err = req.WriteJSON(
		"POST",
		client.BaseURL+"/bar-path",
		map[string]string{
			"Example-Header": "Example-Value",
			"Accept":         "application/test+json",
		},
		"dummy body",
	)
	assert.NoError(t, err)

	startTime := time.Now()

	_, err = req.Do()

	executionTime := time.Now().Sub(startTime).Milliseconds()

	//1 is subtracted because, after pre-last attempt it doesn't wait for back off time
	expectedExecTime := ((retryOptionsCopy.RequestTimeoutPerAttemptInMs+retryOptionsCopy.BackOffTimeAcrossRetriesInMs)*
		time.Duration(retryOptionsCopy.MaxAttempts-1) + retryOptionsCopy.RequestTimeoutPerAttemptInMs).Milliseconds()

	assert.Error(t, err)

	assert.True(t, executionTime >= expectedExecTime, fmt.Sprintf("total execution time must be greater than %d", expectedExecTime))

	assert.True(t, strings.Contains(err.Error(), "errors while making outbound bar.Normal request"),
		"error message not matching")
	assert.True(t, strings.Contains(err.Error(), "context deadline exceeded"),
		"the request should have failed due to context deadline (timeout)")
}

func TestMakingClientCallWithHeadersWithRequestLevelTimeoutAndRetriesSuccess(t *testing.T) {
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

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, _ := io.ReadAll(r.Body)
			w.WriteHeader(200)
			response := map[string]string{"Example-Header": r.Header.Get("Example-Header"), "body": string(bodyBytes)}
			responseBytes, _ := json.Marshal(response)
			_, _ = w.Write(responseBytes)
			// Check that the default header got set and actually sent to the server.
			assert.Equal(t, r.Header.Get("X-Client-ID"), "bar")
			assert.Equal(t, r.Header.Get("Accept"), "application/test+json")
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	barClient := deps.Client.Bar
	client := barClient.HTTPClient()

	retryOptionsCopy := retryOptions
	retryOptionsCopy.RequestTimeoutPerAttemptInMs = 500 * time.Millisecond
	retryOptionsCopy.MaxAttempts = 2

	ctx := zanzibar.WithTimeAndRetryOptions(context.Background(), &retryOptions)
	req := zanzibar.NewClientHTTPRequest(ctx, "bar", "Normal", "bar::Normal", client)

	err = req.WriteJSON(
		"POST",
		client.BaseURL+"/bar-path",
		map[string]string{
			"Example-Header": "Example-Value",
			"Accept":         "application/test+json",
		},
		"dummy body",
	)
	assert.NoError(t, err)

	_, err = req.Do()

	assert.NoError(t, err)
}

func TestBarClientWithoutHeaders(t *testing.T) {
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

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	//override MaxAttempts
	retryOptionsCopy := retryOptions
	retryOptionsCopy.MaxAttempts = 1

	ctx := zanzibar.WithSafeLogFields(context.Background())
	_, _, _, err = bar.EchoI8(
		ctx, nil, &clientsBarBar.Echo_EchoI8_Args{Arg: 42},
	)

	assert.NotNil(t, err)
	assert.Equal(t, "missing mandatory headers: x-uuid", err.Error())

	logFieldsMap := getLogFieldsMapFromContext(ctx)
	expectedFields := []zap.Field{
		zap.Error(err),
		zap.String("error_location", "client::bar"),
	}
	for _, field := range expectedFields {
		_, ok := logFieldsMap[field.Key]
		assert.True(t, ok, "expected field missing: %s", field.Key)
		if ok {
			assert.Equal(t, field, logFieldsMap[field.Key])
		}
	}
}

func TestMakingClientCallWithRespHeaders(t *testing.T) {
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

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Example-Header", "Example-Value")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{
				"stringField":"foo",
				"intWithRange": 0,
				"intWithoutRange": 1,
				"mapIntWithRange": {},
				"mapIntWithoutRange": {},
				"binaryField": "d29ybGQ="
			}`))
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bClient := deps.Client.Bar

	//override MaxAttempts
	retryOptionsCopy := retryOptions
	retryOptionsCopy.MaxAttempts = 1

	ctx := zanzibar.WithSafeLogFields(context.Background())
	_, body, headers, err := bClient.Normal(
		ctx, nil, &clientsBarBar.Bar_Normal_Args{},
	)
	assert.NoError(t, err)
	assert.NotNil(t, body)
	assert.Equal(t, "Example-Value", headers["Example-Header"])

	dynamicFields := []string{
		"Client-Req-Header-Uber-Trace-Id",
		"Client-Res-Header-Content-Length",
		"Client-Res-Header-Date",
	}
	expectedFields := []zap.Field{
		zap.String("Client-Req-Header-X-Client-Id", "bar"),
		zap.String("Client-Req-Header-Content-Type", "application/json"),
		zap.String("Client-Req-Header-Accept", "application/json"),
		zap.String("Client-Res-Header-Example-Header", "Example-Value"),
		zap.String("Client-Res-Header-Content-Type", "text/plain; charset=utf-8"),
		zap.Int64("client_status_code", 200),
	}
	testLogFields(t, ctx, dynamicFields, expectedFields)
}

func testLogFields(t *testing.T, ctx context.Context, dynamicFields []string, expectedFields []zap.Field) {
	expectedFieldsMap := make(map[string]zap.Field)
	for _, field := range expectedFields {
		expectedFieldsMap[field.Key] = field
	}
	logFields := zanzibar.GetLogFieldsFromCtx(ctx)
	logFieldsMap := make(map[string]zap.Field)
	for _, v := range logFields {
		logFieldsMap[v.Key] = v
	}
	for _, k := range dynamicFields {
		_, ok := logFieldsMap[k]
		assert.True(t, ok, "expected field missing %s", k)
		delete(logFieldsMap, k)
	}
	for k, field := range logFieldsMap {
		_, ok := expectedFieldsMap[k]
		assert.True(t, ok, "unexpected log field %s", k)
		if ok {
			assert.Equal(t, expectedFieldsMap[k], field)
		}
	}
	for k, field := range expectedFieldsMap {
		_, ok := logFieldsMap[k]
		assert.True(t, ok, "expected fields missing %s", k)
		if ok {
			assert.Equal(t, field, logFieldsMap[k])
		}
	}
}

func TestMakingClientCallWithThriftException(t *testing.T) {
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

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
			_, _ = w.Write([]byte(`{"stringField":"test"}`))
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bClient := deps.Client.Bar

	//override MaxAttempts
	retryOptionsCopy := retryOptions
	retryOptionsCopy.MaxAttempts = 1

	ctx := zanzibar.WithSafeLogFields(context.Background())
	_, body, _, err := bClient.Normal(
		ctx, nil, &clientsBarBar.Bar_Normal_Args{},
	)
	assert.Error(t, err)
	assert.Nil(t, body)

	realError := err.(*clientsBarBar.BarException)
	assert.Equal(t, realError.StringField, "test")

	logFieldsMap := getLogFieldsMapFromContext(ctx)
	expectedFields := []zap.Field{
		zap.Error(&clientsBarBar.BarException{StringField: "test"}),
		zap.String("error_type", "client_exception"),
		zap.String("error_location", "client::bar"),
	}
	for _, field := range expectedFields {
		_, ok := logFieldsMap[field.Key]
		assert.True(t, ok, "missing field %s", field.Key)
		if ok {
			assert.Equal(t, field, logFieldsMap[field.Key])
		}
	}
}

func getLogFieldsMapFromContext(ctx context.Context) map[string]zap.Field {
	logFieldsMap := make(map[string]zap.Field)
	logFields := zanzibar.GetLogFieldsFromCtx(ctx)
	for _, field := range logFields {
		logFieldsMap[field.Key] = field
	}
	return logFieldsMap
}

func TestMakingClientCallWithBadStatusCode(t *testing.T) {
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

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(402)
			_, _ = w.Write([]byte(`{"stringField":"test"}`))
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bClient := deps.Client.Bar

	//override MaxAttempts
	retryOptionsCopy := retryOptions
	retryOptionsCopy.MaxAttempts = 1

	ctx := zanzibar.WithSafeLogFields(context.Background())
	_, body, _, err := bClient.Normal(
		ctx, nil, &clientsBarBar.Bar_Normal_Args{},
	)
	assert.Error(t, err)
	assert.Nil(t, body)
	assert.Equal(t, "Unexpected http client response (402)", err.Error())

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Received unexpected client response status code"], 1)

	logFieldsMap := getLogFieldsMapFromContext(ctx)
	expectedFields := []zap.Field{
		zap.String("error", "unexpected http response status code: 402"),
		zap.String("error_location", "client::bar"),
		zap.Int("client_status_code", 402),
	}
	for _, field := range expectedFields {
		_, ok := logFieldsMap[field.Key]
		assert.True(t, ok, "missing field %s", field.Key)
		if ok {
			assert.Equal(t, field, logFieldsMap[field.Key])
		}
	}
}

func TestMakingCallWithThriftException(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/arg-not-struct-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
			_, _ = w.Write([]byte(`{"stringField":"test"}`))
		},
	)

	bgateway := gateway.(*benchGateway.BenchGateway)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bClient := deps.Client.Bar

	//override MaxAttempts
	retryOptionsCopy := retryOptions
	retryOptionsCopy.MaxAttempts = 1

	ctx := zanzibar.WithSafeLogFields(context.Background())
	_, _, err = bClient.ArgNotStruct(
		ctx, nil,
		&clientsBarBar.Bar_ArgNotStruct_Args{
			Request: "request",
		},
	)
	assert.Error(t, err)

	realError := err.(*clientsBarBar.BarException)
	assert.Equal(t, realError.StringField, "test")

	logFieldsMap := getLogFieldsMapFromContext(ctx)
	expectedFields := []zap.Field{
		zap.Error(&clientsBarBar.BarException{StringField: "test"}),
		zap.String("error_type", "client_exception"),
		zap.String("error_location", "client::bar"),
	}
	for _, field := range expectedFields {
		_, ok := logFieldsMap[field.Key]
		assert.True(t, ok, "missing field %s", field.Key)
		if ok {
			assert.Equal(t, field, logFieldsMap[field.Key])
		}
	}
}

func TestMakingClientCallWithServerError(t *testing.T) {
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

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(`{}`))
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bClient := deps.Client.Bar

	//override MaxAttempts
	retryOptionsCopy := retryOptions
	retryOptionsCopy.MaxAttempts = 1

	ctx := zanzibar.WithSafeLogFields(context.Background())
	_, body, _, err := bClient.Normal(
		ctx, nil, &clientsBarBar.Bar_Normal_Args{},
	)
	assert.Error(t, err)
	assert.Nil(t, body)
	assert.Equal(t, "Unexpected http client response (500)", err.Error())

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Received unexpected client response status code"], 1)

	logFieldsMap := getLogFieldsMapFromContext(ctx)
	expectedFields := []zap.Field{
		zap.String("error", "unexpected http response status code: 500"),
		zap.String("error_location", "client::bar"),
		zap.Int("client_status_code", 500),
	}
	for _, field := range expectedFields {
		_, ok := logFieldsMap[field.Key]
		assert.True(t, ok, "missing field %s", field.Key)
		if ok {
			assert.Equal(t, field, logFieldsMap[field.Key])
		}
	}
}

func TestInjectSpan(t *testing.T) {
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

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(r.Header.Get("Example-Header")))
			assert.Equal(t, r.Header.Get("X-Client-ID"), "bar")
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	barClient := deps.Client.Bar
	client := barClient.HTTPClient()
	ctx := context.Background()
	req := zanzibar.NewClientHTTPRequest(ctx, "bar", "Normal", "bar::Normal", client)
	err = req.WriteJSON(
		"POST",
		client.BaseURL+"/bar-path",
		map[string]string{
			"Example-Header": "Example-Value",
		},
		nil,
	)
	assert.NoError(t, err)

	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("someSpan")
	err = req.InjectSpanToHeader(span, opentracing.HTTPHeaders)
	assert.NoError(t, err, "failed to inject span context")
	err = req.InjectSpanToHeader(span, "invalid format")
	assert.Error(t, err, "should return error")
}

func TestDefaultRetryPolicy(t *testing.T) {
	backOffTimeAcrossRetriesInMs := time.Duration(15) * time.Millisecond
	options := zanzibar.TimeoutAndRetryOptions{
		OverallTimeoutInMs:           time.Duration(1000) * time.Millisecond,
		RequestTimeoutPerAttemptInMs: time.Duration(1000) * time.Millisecond,
		BackOffTimeAcrossRetriesInMs: backOffTimeAcrossRetriesInMs,
		MaxAttempts:                  1,
	}
	response := http.Response{}
	error := errors.New("simple new error")

	startTime := time.Now()
	shouldRetry := zanzibar.DefaultRetryPolicy(context.Background(), &options, &response, error)

	assert.True(t, shouldRetry, "Expected shouldRetry to be true")
	assert.True(t, time.Now().Sub(startTime).Milliseconds() >= backOffTimeAcrossRetriesInMs.Milliseconds(),
		fmt.Sprintf("expected runtime duration for DefaultRetryPolicy method is >= %d MS", backOffTimeAcrossRetriesInMs.Milliseconds()))
}
