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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	bazClient "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	endpointsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/tchannel/baz/baz"
	"github.com/uber/zanzibar/runtime"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
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

	numMetrics := 15
	cg.MetricsWaitGroup.Add(numMetrics)

	ctx := context.Background()
	reqHeaders := map[string]string{
		"x-token":       "token",
		"x-uuid":        "uuid",
		"Regionname":    "san_francisco",
		"Device":        "ios",
		"Deviceversion": "carbon",
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

	endpointNames := []string{
		"endpoint.latency",
		"endpoint.request",
		"endpoint.success",
	}
	endpointTags := map[string]string{
		"env":            "test",
		"service":        "test-gateway",
		"endpointmethod": "SimpleService__Call",
		"dc":             "unknown",
		"host":           zanzibar.GetHostname(),
		"endpointid":     "bazTChannel",
		"handlerid":      "call",
		"device":         "ios",
		"deviceversion":  "carbon",
		"regionname":     "san_francisco",
		"protocal":       "TChannel",
	}

	for _, name := range endpointNames {
		key := tally.KeyForPrefixedStringMap(name, endpointTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	inboundLatency := metrics[tally.KeyForPrefixedStringMap(
		"endpoint.latency", endpointTags,
	)]
	value := *inboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	inboundRecvd := metrics[tally.KeyForPrefixedStringMap(
		"endpoint.request", endpointTags,
	)]
	value = *inboundRecvd.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value)

	inboundSuccess := metrics[tally.KeyForPrefixedStringMap(
		"endpoint.success", endpointTags,
	)]
	value = *inboundSuccess.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	tchannelOutboundNames := []string{
		"outbound.calls.per-attempt.latency",
	}
	tchannelOutboundTags := map[string]string{
		"app":             "test-gateway",
		"host":            zanzibar.GetHostname(),
		"env":             "test",
		"service":         "test-gateway",
		"target-endpoint": "SimpleService__call",
		"target-service":  "bazService",
		"dc":              "unknown",
	}

	for _, name := range tchannelOutboundNames {
		key := tally.KeyForPrefixedStringMap(name, tchannelOutboundTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	outboundLatency := metrics[tally.KeyForPrefixedStringMap(
		"outbound.calls.per-attempt.latency",
		tchannelOutboundTags,
	)]
	value = *outboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	clientNames := []string{
		"client.latency",
		"client.request",
		"client.success",
	}
	clientTags := map[string]string{
		"env":            "test",
		"service":        "test-gateway",
		"clientid":       "baz",
		"clientmethod":   "call",
		"endpointmethod": "SimpleService__Call",
		"targetservice":  "bazService",
		"targetendpoint": "SimpleService__call",
		"dc":             "unknown",
		"host":           zanzibar.GetHostname(),
		"endpointid":     "bazTChannel",
		"handlerid":      "call",
		"protocal":       "TChannel",
	}

	for _, name := range clientNames {
		key := tally.KeyForPrefixedStringMap(name, clientTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	outboundLatency = metrics[tally.KeyForPrefixedStringMap(
		"client.latency", clientTags,
	)]
	value = *outboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	outboundSent := metrics[tally.KeyForPrefixedStringMap(
		"client.request", clientTags,
	)]
	value = *outboundSent.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value)

	outboundSuccess := metrics[tally.KeyForPrefixedStringMap(
		"client.success", clientTags,
	)]
	value = *outboundSuccess.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")
}
