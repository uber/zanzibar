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

package gateway

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	jaegerGen "github.com/uber/jaeger-client-go/thrift-gen/jaeger"
	"github.com/uber/zanzibar/test/lib/util"

	bazClient "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/baz/baz"
	clientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/baz/baz"
	endpointsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/tchannel/baz/baz"
	zanzibar "github.com/uber/zanzibar/runtime"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
)

func TestHTTPEndpointToHTTPClient(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"clients.bar.serviceName": "barService",
	}, &testGateway.Options{
		CountMetrics:      true,
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
		JaegerFlushMillis: 1,
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	cg := gateway.(*testGateway.ChildProcessGateway)
	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/hello",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte(`"hello"`)); err != nil {
				t.Fatal("can't write fake response")
			}
		},
	)

	_, err = gateway.MakeRequest(
		"GET",
		"/bar/hello",
		nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	// Wait until all spans are flushed
	time.Sleep(time.Millisecond * 100)

	batches := cg.JaegerAgent.GetJaegerBatches()

	var spans []*jaegerGen.Span
	for _, batch := range batches {
		for _, span := range batch.Spans {
			spans = append(spans, span)
		}
	}

	assert.Equal(t, 2, len(spans))
	endpointSpan, clientSpan := spans[1], spans[0]
	assert.Equal(t, "bar.Hello(Bar::helloWorld)", clientSpan.GetOperationName())
	assert.Equal(t, "bar.helloWorld", endpointSpan.GetOperationName())
	assert.Equal(t, endpointSpan.GetSpanId(), clientSpan.GetParentSpanId())
}

func TestHTTPEndpointToHTTPClientWithUpstreamSpan(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"clients.bar.serviceName": "barService",
	}, &testGateway.Options{
		CountMetrics:      true,
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
		JaegerFlushMillis: 1,
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	cg := gateway.(*testGateway.ChildProcessGateway)
	gateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/hello",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte(`"hello"`)); err != nil {
				t.Fatal("can't write fake response")
			}
		},
	)

	tracer, closer, err := config.Configuration{
		ServiceName: "upstream",
	}.NewTracer()
	if !assert.NoError(t, err, "error creating upstream tracer") {
		return
	}

	upstream := cg.HTTPClient

	url := "http://" + cg.RealHTTPAddr + "/bar/hello"
	req, err := http.NewRequest("GET", url, nil)
	if !assert.NoError(t, err, "error creating http request") {
		return
	}

	upstreamSpanID := jaeger.SpanID(42)
	spanContext := jaeger.NewSpanContext(
		jaeger.TraceID{High: 255, Low: 255},
		upstreamSpanID,
		jaeger.SpanID(0),
		true,
		nil,
	)
	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	sp := tracer.StartSpan("upstream")
	err = tracer.Inject(spanContext, opentracing.HTTPHeaders, carrier)
	if !assert.NoError(t, err, "failed to inject span context") {
		return
	}

	_, err = upstream.Do(req)
	if !assert.NoError(t, err, "got http error") {
		return
	}
	sp.Finish()
	_ = closer.Close()

	// Wait until all spans are flushed
	time.Sleep(time.Millisecond * 100)

	batches := cg.JaegerAgent.GetJaegerBatches()

	var spans []*jaegerGen.Span
	for _, batch := range batches {
		for _, span := range batch.Spans {
			spans = append(spans, span)
		}
	}

	assert.Equal(t, 2, len(spans))
	endpointSpan, clientSpan := spans[1], spans[0]
	assert.Equal(t, "bar.Hello(Bar::helloWorld)", clientSpan.GetOperationName())
	assert.Equal(t, "bar.helloWorld", endpointSpan.GetOperationName())
	assert.Equal(t, int64(upstreamSpanID), endpointSpan.GetParentSpanId())
	assert.Equal(t, endpointSpan.GetSpanId(), clientSpan.GetParentSpanId())
}

func TestHTTPEndpointToTChannelClient(t *testing.T) {
	t.Skip("Flaky test")
	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"clients.baz.serviceName": "bazService",
	}, &testGateway.Options{
		CountMetrics:          true,
		KnownTChannelBackends: []string{"baz"},
		TestBinary:            util.DefaultMainFile("example-gateway"),
		ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		JaegerFlushMillis:     1,
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	cg := gateway.(*testGateway.ChildProcessGateway)

	fakeCall := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SimpleService_Call_Args,
	) (map[string]string, error) {

		var resHeaders map[string]string

		return resHeaders, nil
	}

	err = gateway.TChannelBackends()["baz"].Register(
		"baz", "call", "SimpleService::call",
		bazClient.NewSimpleServiceCallHandler(fakeCall),
	)
	assert.NoError(t, err)

	headers := map[string]string{}
	headers["x-token"] = "token"
	headers["x-uuid"] = "uuid"

	_, err = gateway.MakeRequest(
		"POST",
		"/baz/call",
		headers,
		bytes.NewReader([]byte(`{"arg":{"b1":true,"i3":42,"s2":"hello"}}`)),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	// Wait until all spans are flushed
	time.Sleep(time.Millisecond * 100)

	batches := cg.JaegerAgent.GetJaegerBatches()

	var spans []*jaegerGen.Span
	for _, batch := range batches {
		for _, span := range batch.Spans {
			spans = append(spans, span)
		}
	}

	assert.Equal(t, 2, len(spans))
	endpointSpan, clientSpan := spans[1], spans[0]
	assert.Equal(t, "SimpleService::call", clientSpan.GetOperationName())
	assert.Equal(t, "baz.call", endpointSpan.GetOperationName())
	assert.Equal(t, endpointSpan.GetSpanId(), clientSpan.GetParentSpanId())
}

func TestHTTPEndpointToTChannelClientWithUpstreamSpan(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"clients.baz.serviceName": "bazService",
	}, &testGateway.Options{
		CountMetrics:          true,
		KnownTChannelBackends: []string{"baz"},
		TestBinary:            util.DefaultMainFile("example-gateway"),
		ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		JaegerFlushMillis:     1,
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	cg := gateway.(*testGateway.ChildProcessGateway)

	fakeCall := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SimpleService_Call_Args,
	) (map[string]string, error) {

		var resHeaders map[string]string

		return resHeaders, nil
	}

	err = gateway.TChannelBackends()["baz"].Register(
		"baz", "call", "SimpleService::call",
		bazClient.NewSimpleServiceCallHandler(fakeCall),
	)
	assert.NoError(t, err)

	headers := map[string]string{}
	headers["x-token"] = "token"
	headers["x-uuid"] = "uuid"

	tracer, closer, err := config.Configuration{
		ServiceName: "upstream",
	}.NewTracer()
	if !assert.NoError(t, err, "error creating upstream tracer") {
		return
	}

	upstream := cg.HTTPClient

	url := "http://" + cg.RealHTTPAddr + "/baz/call"
	body := bytes.NewReader([]byte(`{"arg":{"b1":true,"i3":42,"s2":"hello"}}`))
	req, err := http.NewRequest("POST", url, body)
	if !assert.NoError(t, err, "error creating http request") {
		return
	}
	for headerName, headerValue := range headers {
		req.Header.Set(headerName, headerValue)
	}

	upstreamSpanID := jaeger.SpanID(42)
	spanContext := jaeger.NewSpanContext(
		jaeger.TraceID{High: 255, Low: 255},
		upstreamSpanID,
		jaeger.SpanID(0),
		true,
		nil,
	)
	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	sp := tracer.StartSpan("upstream")
	err = tracer.Inject(spanContext, opentracing.HTTPHeaders, carrier)
	if !assert.NoError(t, err, "failed to inject span context") {
		return
	}

	_, err = upstream.Do(req)
	if !assert.NoError(t, err, "got http error") {
		return
	}
	sp.Finish()
	_ = closer.Close()

	// Wait until all spans are flushed
	time.Sleep(time.Millisecond * 100)

	batches := cg.JaegerAgent.GetJaegerBatches()

	var spans []*jaegerGen.Span
	for _, batch := range batches {
		for _, span := range batch.Spans {
			spans = append(spans, span)
		}
	}

	assert.Equal(t, 2, len(spans))
	endpointSpan, clientSpan := spans[1], spans[0]
	assert.Equal(t, "SimpleService::call", clientSpan.GetOperationName())
	assert.Equal(t, "baz.call", endpointSpan.GetOperationName())
	assert.Equal(t, int64(upstreamSpanID), endpointSpan.GetParentSpanId())
	assert.Equal(t, endpointSpan.GetSpanId(), clientSpan.GetParentSpanId())
}

func TestTChannelEndpoint(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"clients.baz.serviceName": "bazService",
	}, &testGateway.Options{
		CountMetrics:          true,
		KnownTChannelBackends: []string{"baz"},
		TestBinary:            util.DefaultMainFile("example-gateway"),
		ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		JaegerFlushMillis:     1,
		TChannelClientMethods: map[string]string{
			"SimpleService::Call": "Call",
		},
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	cg := gateway.(*testGateway.ChildProcessGateway)

	fakeCall := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBaz.SimpleService_Call_Args,
	) (map[string]string, error) {

		return map[string]string{
			"some-res-header": "something",
		}, nil
	}

	err = gateway.TChannelBackends()["baz"].Register(
		"baz", "call", "SimpleService::call",
		bazClient.NewSimpleServiceCallHandler(fakeCall),
	)
	assert.NoError(t, err)

	ctx := context.Background()
	reqHeaders := map[string]string{
		"x-token": "token",
		"x-uuid":  "uuid",
	}
	args := &endpointsBaz.SimpleService_Call_Args{
		Arg: &endpointsBaz.BazRequest{
			B1: true,
			S2: "hello",
			I3: 42,
		},
	}
	var result endpointsBaz.SimpleService_Call_Result

	_, _, err = gateway.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	if !assert.NoError(t, err, "got tchannel error") {
		return
	}

	// Wait until all spans are flushed
	time.Sleep(time.Millisecond * 100)

	batches := cg.JaegerAgent.GetJaegerBatches()

	var spans []*jaegerGen.Span
	for _, batch := range batches {
		for _, span := range batch.Spans {
			spans = append(spans, span)
		}
	}

	assert.Equal(t, 2, len(spans))
	endpointSpan, clientSpan := spans[1], spans[0]
	assert.Equal(t, "SimpleService::call", clientSpan.GetOperationName())
	assert.Equal(t, "SimpleService::Call", endpointSpan.GetOperationName())
	assert.Equal(t, endpointSpan.GetSpanId(), clientSpan.GetParentSpanId())
}
