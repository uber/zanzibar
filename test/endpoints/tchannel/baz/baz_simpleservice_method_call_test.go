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

package bazTchannel

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	endpointsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/tchannel/baz/baz"
	"github.com/uber/zanzibar/test/lib/test_gateway"
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

	gateway.TChannelBackends()["baz"].Register(
		"baz", "call", "SimpleService::call",
		bazClient.NewSimpleServiceCallHandler(fakeCall),
	)

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
	assert.Equal(t, 5, len(allLogs))
	assert.Equal(t, 1, len(allLogs["Started ExampleGateway"]))
	assert.Equal(t, 1, len(allLogs["Inbound connection is active."]))
	assert.Equal(t, 1, len(allLogs["Outbound connection is active."]))
	assert.Equal(t, 1, len(allLogs["Finished an outgoing client TChannel request"]))
	assert.Equal(t, 1, len(allLogs["Finished an incoming server TChannel request"]))

	tags := allLogs["Finished an incoming server TChannel request"][0]
	assert.Equal(t, "info", tags["level"])
	assert.Equal(t, "Finished an incoming server TChannel request", tags["msg"])
	assert.Equal(t, "bazTChannel", tags["endpointID"])
	assert.Equal(t, "bazTChannel", tags["handlerID"])
	assert.Equal(t, "SimpleService::Call", tags["method"])
	assert.Equal(t, "token", tags["Request-Header-x-token"])
	assert.Equal(t, "uuid", tags["Request-Header-x-uuid"])
	assert.Equal(t, "something", tags["Response-Header-some-res-header"])
	assert.Equal(t, "test-gateway", tags["calling-service"])
	assert.Contains(t, tags["remoteAddr"], getLocalAddr(t))
	assert.NotNil(t, tags["timestamp-started"])
	assert.NotNil(t, tags["timestamp-finished"])

	tags = allLogs["Finished an outgoing client TChannel request"][0]
	assert.Equal(t, "info", tags["level"])
	assert.Equal(t, "Finished an outgoing client TChannel request", tags["msg"])
	assert.Equal(t, "baz", tags["clientID"])
	assert.Equal(t, "bazService", tags["serviceName"])
	assert.Equal(t, "Call", tags["methodName"])
	assert.Equal(t, "SimpleService::call", tags["serviceMethod"])
	assert.Equal(t, "token", tags["Request-Header-x-token"])
	assert.Equal(t, "uuid", tags["Request-Header-x-uuid"])
	assert.Equal(t, "something", tags["Response-Header-some-res-header"])
	assert.Contains(t, tags["remoteAddr"], "127.0.0.1")
	assert.NotNil(t, tags["timestamp-started"])
	assert.NotNil(t, tags["timestamp-finished"])
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
