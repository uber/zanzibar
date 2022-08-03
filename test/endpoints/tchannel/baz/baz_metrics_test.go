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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	bazClient "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/baz/baz"
	endpointsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/tchannel/baz/baz"
	zanzibar "github.com/uber/zanzibar/runtime"
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

	numMetrics := 11
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

	cg.MetricsWaitGroup.Wait()
	metrics := cg.M3Service.GetMetrics()
	// we don't care about jaeger emitted metrics
	for key := range metrics {
		if strings.HasPrefix(key, "jaeger") || strings.Contains(key, "circuitbreaker") {
			delete(metrics, key)
		}
	}
	assert.Equal(t, numMetrics, len(metrics)-2) // number to go away after we remove the gratuitous histograms for timer metrics

	endpointNames := []string{
		"endpoint.request",
		"endpoint.success",
	}
	endpointTags := map[string]string{
		"env":            "test",
		"service":        "test-gateway",
		"endpointmethod": "SimpleService__Call",
		"dc":             "unknown",
		"endpointid":     "bazTChannel",
		"handlerid":      "call",
		"device":         "ios",
		"deviceversion":  "carbon",
		"regionname":     "san_francisco",
		"protocol":       "TChannel",
	}

	for _, name := range endpointNames {
		key := tally.KeyForPrefixedStringMap(name, endpointTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
		assert.Equal(t, int64(1), metrics[key].Value.Count)
	}

	key := tally.KeyForPrefixedStringMap("endpoint.latency", endpointTags)
	assert.Contains(t, metrics, key, "expected metric: %s", key)
	value := metrics[key].Value.Timer
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 50*1000*1000, "expected timer to be < 10 milli seconds")

	key = "endpoint.latency-hist"
	keyFound := false
	for metricKeyName := range metrics {
		if strings.Contains(metricKeyName, key) {
			if mapValue, ok := metrics[metricKeyName]; ok {
				assert.Equal(t, int64(1), mapValue.Value.Count, fmt.Sprintf("key: %s, metric: %v\n", key, metrics[key]))
				keyFound = true
			}
		}
	}
	assert.True(t, keyFound, fmt.Sprintf("expected the key: %s to be in metrics", key))

	tchannelOutboundNames := []string{
		"outbound.calls.per-attempt.latency",
	}
	tchannelOutboundTags := map[string]string{
		"app": "test-gateway",
		// this host tag is added by tchannel library, which we don't have control with
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
	value = outboundLatency.Value.Timer
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 10*1000*1000, "expected timer to be 10 milli second")

	clientNames := []string{
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
		"device":         "ios",
		"deviceversion":  "carbon",
		"regionname":     "san_francisco",
		"endpointid":     "bazTChannel",
		"handlerid":      "call",
		"protocol":       "TChannel",
	}

	for _, name := range clientNames {
		key := tally.KeyForPrefixedStringMap(name, clientTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
		assert.Equal(t, int64(1), metrics[key].Value.Count, "expected counter to be 1")
	}

	key = tally.KeyForPrefixedStringMap("client.latency", clientTags)
	assert.Contains(t, metrics, key, "expected metric: %s", key)
	value = metrics[key].Value.Timer
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 50*1000*1000, "expected timer to be < 10 milli seconds")

	key = "client.latency-hist"
	keyFound = false
	for metricKeyName := range metrics {
		if strings.Contains(metricKeyName, key) {
			if mapValue, ok := metrics[metricKeyName]; ok {
				assert.Equal(t, int64(1), mapValue.Value.Count, fmt.Sprintf("key: %s, metric: %v\n", key, metrics[key]))
				keyFound = true
			}
		}
	}
	assert.True(t, keyFound, fmt.Sprintf("expected the key: %s to be in metrics", key))
}
