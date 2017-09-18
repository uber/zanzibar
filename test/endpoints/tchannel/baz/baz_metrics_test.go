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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"

	"github.com/uber-go/tally"
	"github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	endpointsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/tchannel/baz/baz"
	"github.com/uber/zanzibar/runtime"
)

func TestCallMetrics(t *testing.T) {
	testCallCounter := 0

	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"clients.baz.serviceName": "bazService",
	}, &testGateway.Options{
		CountMetrics:          true,
		KnownTChannelBackends: []string{"baz"},
		TestBinary:            util.DefaultMainFile("example-gateway"),
		ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
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
		testCallCounter++

		return map[string]string{
			"some-res-header": "something",
		}, nil
	}

	gateway.TChannelBackends()["baz"].Register(
		"baz", "call", "SimpleService::call",
		bazClient.NewSimpleServiceCallHandler(fakeCall),
	)

	numMetrics := 14
	cg.MetricsWaitGroup.Add(numMetrics)

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

	cg.MetricsWaitGroup.Wait()
	metrics := cg.M3Service.GetMetrics()
	assert.Equal(t, numMetrics, len(metrics))

	tchannelInboundNames := []string{
		"tchannel.inbound.calls.latency",
		"tchannel.inbound.calls.recvd",
		"tchannel.inbound.calls.success",
	}
	tchannelInboundTags := map[string]string{
		"app":             "test-gateway",
		"host":            zanzibar.GetHostname(),
		"env":             "test",
		"service":         "test-gateway",
		"endpoint":        "SimpleService::Call",
		"calling-service": "test-gateway",
	}
	for _, name := range tchannelInboundNames {
		key := tally.KeyForPrefixedStringMap(name, tchannelInboundTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	inboundLatency := metrics[tally.KeyForPrefixedStringMap("tchannel.inbound.calls.latency", tchannelInboundTags)]
	value := *inboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	inboundRecvd := metrics[tally.KeyForPrefixedStringMap("tchannel.inbound.calls.recvd", tchannelInboundTags)]
	value = *inboundRecvd.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value)

	inboundSuccess := metrics[tally.KeyForPrefixedStringMap("tchannel.inbound.calls.success", tchannelInboundTags)]
	value = *inboundSuccess.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	endpointNames := []string{
		"inbound.calls.latency",
		"inbound.calls.recvd",
		"inbound.calls.success",
	}
	endpointTags := map[string]string{
		"env":      "test",
		"service":  "test-gateway",
		"endpoint": "bazTChannel",
		"handler":  "call",
		"method":   "SimpleService::Call",
	}

	for _, name := range endpointNames {
		key := tally.KeyForPrefixedStringMap(name, endpointTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	inboundLatency = metrics[tally.KeyForPrefixedStringMap("inbound.calls.latency", endpointTags)]
	value = *inboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	inboundRecvd = metrics[tally.KeyForPrefixedStringMap("inbound.calls.recvd", endpointTags)]
	value = *inboundRecvd.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value)

	inboundSuccess = metrics[tally.KeyForPrefixedStringMap("inbound.calls.success", endpointTags)]
	value = *inboundSuccess.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	tchannelOutboundNames := []string{
		"tchannel.outbound.calls.latency",
		"tchannel.outbound.calls.per-attempt.latency",
		"tchannel.outbound.calls.send",
		"tchannel.outbound.calls.success",
	}
	tchannelOutboundTags := map[string]string{
		"app":             "test-gateway",
		"host":            zanzibar.GetHostname(),
		"env":             "test",
		"service":         "test-gateway",
		"target-endpoint": "SimpleService::call",
		"target-service":  "bazService",
	}
	for _, name := range tchannelOutboundNames {
		key := tally.KeyForPrefixedStringMap(name, tchannelOutboundTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	outboundLatency := metrics[tally.KeyForPrefixedStringMap("tchannel.outbound.calls.latency", tchannelOutboundTags)]
	value = *outboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	outboundLatency = metrics[tally.KeyForPrefixedStringMap("tchannel.outbound.calls.per-attempt.latency", tchannelOutboundTags)]
	value = *outboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	outboundSend := metrics[tally.KeyForPrefixedStringMap("tchannel.outbound.calls.send", tchannelOutboundTags)]
	value = *outboundSend.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value)

	outboundSuccess := metrics[tally.KeyForPrefixedStringMap("tchannel.outbound.calls.success", tchannelOutboundTags)]
	value = *outboundSuccess.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")
}
