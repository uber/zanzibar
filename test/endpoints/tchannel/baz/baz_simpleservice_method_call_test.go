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

package baztchannel

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	bazClient "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/baz/baz"
	endpointsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/tchannel/baz/baz"
	zanzibar "github.com/uber/zanzibar/runtime"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

func TestCallTChannelSuccessfulRequestOKResponse(t *testing.T) {
	testCallCounter := 0

	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"clients.baz.serviceName": "bazService",
	}, &testGateway.Options{
		KnownTChannelBackends: []string{"baz"},
		TestBinary:            util.DefaultMainFile("example-gateway"),
		ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		TChannelClientMethods: map[string]string{
			"SimpleService::Call": "Call",
		},
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	fakeCall := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBaz.SimpleService_Call_Args,
	) (map[string]string, error) {
		testCallCounter++

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
		"x-token":       "token",
		"x-uuid":        "uuid",
		"Regionname":    "sf",
		"Device":        "ios",
		"Deviceversion": "1.0",
	}
	args := &endpointsBaz.SimpleService_Call_Args{
		Arg: &endpointsBaz.BazRequest{
			B1: true,
			S2: "hello",
			I3: 42,
		},
	}
	var result endpointsBaz.SimpleService_Call_Result

	success, resHeaders, err := gateway.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  0,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	dynamicRespHeaders := []string{
		"client.response.duration",
	}
	for _, dynamicValue := range dynamicRespHeaders {
		assert.Contains(t, resHeaders, dynamicValue)
		delete(resHeaders, dynamicValue)
	}

	if !assert.NoError(t, err, "got tchannel error") {
		return
	}
	expectedHeaders := map[string]string{
		"some-res-header": "something",
	}
	assert.Equal(t, 1, testCallCounter)
	assert.Equal(t, expectedHeaders, resHeaders)
	assert.True(t, success)

	allLogs := gateway.AllLogs()
	assert.Equal(t, 1, len(allLogs["Started Example-gateway"]))
	assert.Equal(t, 1, len(allLogs["Finished an outgoing client TChannel request"]))
	assert.Equal(t, 1, len(allLogs["Finished an incoming server TChannel request"]))

	logs := allLogs["Finished an incoming server TChannel request"][0]
	dynamicHeaders := []string{
		"requestUUID",
		"timestamp-started",
		"timestamp-finished",
		"remoteAddr",
		"ts",
		"hostname",
		"pid",
		"Res-Header-client.response.duration",
	}
	for _, dynamicValue := range dynamicHeaders {
		assert.Contains(t, logs, dynamicValue)
		delete(logs, dynamicValue)
	}

	expectedValues := map[string]string{
		"level":                      "debug",
		"msg":                        "Finished an incoming server TChannel request",
		"env":                        "test",
		"service":                    "example-gateway",
		"endpointID":                 "bazTChannel",
		"endpointHandler":            "call",
		"endpointThriftMethod":       "SimpleService::Call",
		"x-uuid":                     "uuid",
		"calling-service":            "test-gateway",
		"zone":                       "unknown",
		"Device":                     "ios",
		"Regionname":                 "sf",
		"Deviceversion":              "1.0",
		"Res-Header-some-res-header": "something",
	}
	for actualKey, actualValue := range logs {
		assert.Equal(t, expectedValues[actualKey], actualValue, "unexpected field %q", actualKey)
	}
	for expectedKey, expectedValue := range expectedValues {
		assert.Equal(t, expectedValue, logs[expectedKey], "unexpected field %q", expectedKey)
	}

	logs = allLogs["Finished an outgoing client TChannel request"][0]
	dynamicHeaders = []string{
		"requestUUID",
		"remoteAddr",
		"timestamp-started",
		"ts",
		"hostname",
		"pid",
		"timestamp-finished",
		"Client-Req-Header-x-request-uuid",
		"Client-Req-Header-$tracing$uber-trace-id",
		zanzibar.TraceIDKey,
		zanzibar.TraceSpanKey,
		zanzibar.TraceSampledKey,
	}
	for _, dynamicValue := range dynamicHeaders {
		assert.Contains(t, logs, dynamicValue)
		delete(logs, dynamicValue)
	}

	expectedValues = map[string]string{
		"msg":     "Finished an outgoing client TChannel request",
		"env":     "test",
		"level":   "debug",
		"zone":    "unknown",
		"service": "example-gateway",

		// contextual logs
		"endpointID":           "bazTChannel",
		"endpointThriftMethod": "SimpleService::Call",
		"endpointHandler":      "call",
		"x-uuid":               "uuid",
		"Deviceversion":        "1.0",
		"Device":               "ios",
		"Regionname":           "sf",

		// client specific logs
		//"clientID":                          "baz",
		//"clientService":                     "bazService",
		//"clientThriftMethod":                "SimpleService::call",
		//"clientMethod":                      "Call",
		"Client-Req-Header-Device":          "ios",
		"Client-Req-Header-x-uuid":          "uuid",
		"Client-Req-Header-Regionname":      "sf",
		"Client-Req-Header-Deviceversion":   "1.0",
		"Client-Res-Header-some-res-header": "something",
	}
	for actualKey, actualValue := range logs {
		assert.Equal(t, expectedValues[actualKey], actualValue, "unexpected field %q", actualKey)
	}
	for expectedKey, expectedValue := range expectedValues {
		assert.Equal(t, expectedValue, logs[expectedKey], "unexpected field %q", expectedKey)
	}
}

func getLocalAddr(t *testing.T) string {
	addrs, err := net.InterfaceAddrs()
	assert.NoError(t, err)
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "unknown"
}

func TestCallTChannelTimeout(t *testing.T) {
	testCallCounter := 0

	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"clients.baz.serviceName": "bazService",
		"tchannel.client.timeout": 200,
	}, &testGateway.Options{
		KnownTChannelBackends: []string{"baz"},
		TestBinary:            util.DefaultMainFile("example-gateway"),
		ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		TChannelClientMethods: map[string]string{
			"SimpleService::Call": "Call",
		},
		LogWhitelist: map[string]bool{
			"Endpoint failure: handler returned error": true,
		},
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	fakeCall := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBaz.SimpleService_Call_Args,
	) (map[string]string, error) {
		testCallCounter++
		time.Sleep(400 * time.Millisecond)

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

	success, _, err := gateway.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  0,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.Error(t, err, "excepting tchannel error")
	assert.False(t, success)

	assert.Len(t, gateway.Logs("info", "Started Example-gateway"), 1)

	// logged from tchannel client runtime
	assert.Len(t, gateway.Logs("info", "Failed after non-retriable error."), 1)
	assert.Len(t, gateway.Logs("warn", "Failed to send outgoing client TChannel request"), 1)

	// logged from generated client
	assert.Len(t, gateway.Logs("warn", "Client failure: TChannel client call returned error"), 1)

	// logged from generated endpoint
	assert.Len(t, gateway.Logs("error", "Endpoint failure: handler returned error"), 1)

	// logged from tchannel server runtime
	assert.Len(t, gateway.Logs("warn", "Failed to serve incoming TChannel request"), 1)
	assert.Len(t, gateway.Logs("warn", "Unexpected tchannel system error"), 1)
}

func TestMetricsAppError(t *testing.T) {
	// https://github.com/uber/zanzibar/issues/547
	t.Skip()

	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"clients.baz.serviceName": "bazService",
	}, &testGateway.Options{
		KnownTChannelBackends: []string{"baz"},
		TestBinary:            util.DefaultMainFile("example-gateway"),
		ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		TChannelClientMethods: map[string]string{
			"SimpleService::Call": "Call",
		},
	})
	require.NoError(t, err, "got bootstrap err")
	defer gateway.Close()

	fakeCall := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBaz.SimpleService_Call_Args,
	) (map[string]string, error) {
		return map[string]string{"some-res-header": "true"}, &clientsBaz.AuthErr{
			Message: "authentication is required",
		}
	}

	err = gateway.TChannelBackends()["baz"].Register(
		"baz", "call", "SimpleService::call",
		bazClient.NewSimpleServiceCallHandler(fakeCall),
	)
	assert.NoError(t, err)

	cg := gateway.(*testGateway.ChildProcessGateway)
	ctx := context.Background()
	args := &endpointsBaz.SimpleService_Call_Args{
		Arg: &endpointsBaz.BazRequest{},
	}
	var result endpointsBaz.SimpleService_Call_Result

	reqHeaders := map[string]string{
		"x-uuid":  "test",
		"x-token": "token",
	}
	_, _, err = gateway.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	require.NoError(t, err, "got unexpected tchannel error")

	metrics := cg.M3Service.GetMetrics()
	// we don't care about jaeger emitted metrics
	for key := range metrics {
		if strings.HasPrefix(key, "jaeger") {
			delete(metrics, key)
		}
	}

	endpointTags := map[string]string{
		"app-error":      "endpointsTchannelBazBaz.AuthErr",
		"dc":             "unknown",
		"device":         "",
		"deviceversion":  "",
		"endpointid":     "bazTChannel",
		"endpointmethod": "SimpleService__Call",
		"env":            "test",
		"handlerid":      "call",
		"host":           "jacobg-C02WC093HTDG",
		"protocol":       "TChannel",
		"regionname":     "",
		"service":        "test-gateway",
	}

	// t.Logf("metrics received: %+v", metrics)
	metricAppError := "endpoint.app-errors"
	key := tally.KeyForPrefixedStringMap(metricAppError, endpointTags)
	assert.Contains(t, metrics, key, "expected metric: %s", key)
}
