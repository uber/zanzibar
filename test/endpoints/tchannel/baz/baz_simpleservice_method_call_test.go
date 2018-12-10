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

package baztchannel

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	bazClient "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	endpointsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/tchannel/baz/baz"
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

	success, resHeaders, err := gateway.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result,
	)

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
	assert.Equal(t, 1, len(allLogs["Started ExampleGateway"]))
	assert.Equal(t, 2, len(allLogs["Created new active connection."]))
	assert.Equal(t, 1, len(allLogs["Finished an outgoing client TChannel request"]))
	assert.Equal(t, 1, len(allLogs["Finished an incoming server TChannel request"]))

	tags := allLogs["Finished an incoming server TChannel request"][0]
	dynamicHeaders := []string{
		"requestUUID",
		"timestamp-started",
		"timestamp-finished",
		"remoteAddr",
		"ts",
		"hostname",
		"pid",
	}
	for _, dynamicValue := range dynamicHeaders {
		assert.Contains(t, tags, dynamicValue)
		delete(tags, dynamicValue)
	}

	expectedValues := map[string]interface{}{
		"level":                           "info",
		"env":                             "test",
		"service":                         "example-gateway",
		"msg":                             "Finished an incoming server TChannel request",
		"endpointID":                      "bazTChannel",
		"handlerID":                       "call",
		"method":                          "SimpleService::Call",
		"Request-Header-x-token":          "token",
		"Request-Header-x-uuid":           "uuid",
		"Response-Header-some-res-header": "something",
		"calling-service":                 "test-gateway",
		"zone":                            "unknown",
	}
	for actualKey, actualValue := range tags {
		assert.Equal(t, expectedValues[actualKey], actualValue, "unexpected header %q", actualKey)
	}
	for expectedKey, expectedValue := range expectedValues {
		assert.Equal(t, expectedValue, tags[expectedKey], "unexpected header %q", expectedKey)
	}

	tags = allLogs["Finished an outgoing client TChannel request"][0]
	dynamicHeaders = []string{
		"remoteAddr",
		"timestamp-started",
		"ts",
		"hostname",
		"pid",
		"timestamp-finished",
		"Request-Header-$tracing$uber-trace-id",
	}
	for _, dynamicValue := range dynamicHeaders {
		assert.Contains(t, tags, dynamicValue)
		delete(tags, dynamicValue)
	}

	expectedValues = map[string]interface{}{
		"msg":                             "Finished an outgoing client TChannel request",
		"env":                             "test",
		"clientID":                        "baz",
		"level":                           "info",
		"serviceName":                     "bazService",
		"serviceMethod":                   "SimpleService::call",
		"methodName":                      "Call",
		"zone":                            "unknown",
		"service":                         "example-gateway",
		"Request-Header-x-uuid":           "uuid",
		"Request-Header-x-token":          "token",
		"Response-Header-some-res-header": "something",
	}
	for actualKey, actualValue := range tags {
		assert.Equal(t, expectedValues[actualKey], actualValue, "unexpected header %q", actualKey)
	}
	for expectedKey, expectedValue := range expectedValues {
		assert.Equal(t, expectedValue, tags[expectedKey], "unexpected header %q", expectedKey)
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

	success, resHeaders, err := gateway.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result,
	)
	assert.Error(t, err, "excepting tchannel error")
	assert.Nil(t, resHeaders)
	assert.False(t, success)

	allLogs := gateway.AllLogs()
	assert.Len(t, allLogs, 8)
	assert.Len(t, gateway.Logs("info", "Started ExampleGateway"), 1)
	assert.Len(t, gateway.Logs("info", "Created new active connection."), 2)

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
