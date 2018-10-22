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

package baz

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	bazClient "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	"github.com/uber/zanzibar/runtime"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

func TestCallMetrics(t *testing.T) {
	testcallCounter := 0

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
		args *clientsBazBaz.SimpleService_Call_Args,
	) (map[string]string, error) {
		testcallCounter++

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
	headers["regionname"] = "san_francisco"
	headers["device"] = "ios"
	headers["deviceversion"] = "carbon"

	numMetrics := 14
	cg.MetricsWaitGroup.Add(numMetrics)

	_, err = gateway.MakeRequest(
		"POST",
		"/baz/call",
		headers,
		bytes.NewReader([]byte(`{"arg":{"b1":true,"i3":42,"s2":"hello"}}`)),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	cg.MetricsWaitGroup.Wait()

	metrics := cg.M3Service.GetMetrics()
	assert.Equal(t, numMetrics, len(metrics))

	endpointNames := []string{
		"endpoint.latency",
		"endpoint.request",
	}
	endpointTags := map[string]string{
		"env":           "test",
		"service":       "test-gateway",
		"endpointid":    "baz",
		"handlerid":     "call",
		"regionname":    "san_francisco",
		"device":        "ios",
		"deviceversion": "carbon",
		"dc":            "unknown",
		"host":          zanzibar.GetHostname(),
		"protocal":      "HTTP",
	}
	statusTags := map[string]string{
		"env":           "test",
		"service":       "test-gateway",
		"status":        "204",
		"endpointid":    "baz",
		"handlerid":     "call",
		"regionname":    "san_francisco",
		"device":        "ios",
		"deviceversion": "carbon",
		"dc":            "unknown",
		"host":          zanzibar.GetHostname(),
		"protocal":      "HTTP",
	}
	for _, name := range endpointNames {
		key := tally.KeyForPrefixedStringMap(name, endpointTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	latencyMetric := metrics[tally.KeyForPrefixedStringMap(
		"endpoint.latency", endpointTags,
	)]
	value := *latencyMetric.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	recvdMetric := metrics[tally.KeyForPrefixedStringMap(
		"endpoint.request", endpointTags,
	)]
	value = *recvdMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	statusMetric := metrics[tally.KeyForPrefixedStringMap(
		"endpoint.status", statusTags,
	)]
	value = *statusMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	tchannelNames := []string{
		"outbound.calls.per-attempt.latency",
	}
	tchannelTags := map[string]string{
		"env":             "test",
		"app":             "test-gateway",
		"service":         "test-gateway",
		"target-service":  "bazService",
		"target-endpoint": "SimpleService__call",
		"host":            zanzibar.GetHostname(),
		"dc":              "unknown",
	}
	for _, name := range tchannelNames {
		key := tally.KeyForPrefixedStringMap(name, tchannelTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	perAttemptOutboundLatency := metrics[tally.KeyForPrefixedStringMap(
		"outbound.calls.per-attempt.latency",
		tchannelTags,
	)]
	value = *perAttemptOutboundLatency.MetricValue.Timer.I64Value
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
		"targetservice":  "bazService",
		"targetendpoint": "SimpleService__call",
		"dc":             "unknown",
		"host":           zanzibar.GetHostname(),
		"endpointid":     "baz",
		"handlerid":      "call",
		"regionname":     "san_francisco",
		"device":         "ios",
		"deviceversion":  "carbon",
		"protocal":       "HTTP",
	}
	for _, name := range clientNames {
		key := tally.KeyForPrefixedStringMap(name, clientTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	outboundLatency := metrics[tally.KeyForPrefixedStringMap(
		"client.latency", clientTags,
	)]
	value = *outboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	outboundSent := metrics[tally.KeyForPrefixedStringMap(
		"client.request", clientTags,
	)]
	value = *outboundSent.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	outboundSuccess := metrics[tally.KeyForPrefixedStringMap(
		"client.success", clientTags,
	)]
	value = *outboundSuccess.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	defaultTags := map[string]string{
		"env":     "test",
		"service": "test-gateway",
		"dc":      "unknown",
		"host":    zanzibar.GetHostname(),
	}

	loggedMetrics := metrics[tally.KeyForPrefixedStringMap(
		"zap.logged.info", defaultTags,
	)]
	value = *loggedMetrics.MetricValue.Count.I64Value
	assert.Equal(t, int64(4), value, "expected counter to be 4")
}
