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

package baz

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	"github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/test_gateway"
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

	gateway.TChannelBackends()["baz"].Register(
		"baz", "call", "SimpleService::call",
		bazClient.NewSimpleServiceCallHandler(fakeCall),
	)

	headers := map[string]string{}
	headers["x-token"] = "token"
	headers["x-uuid"] = "uuid"

	numMetrics := 15
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
		"inbound.calls.latency",
		"inbound.calls.recvd",
		"inbound.calls.success",
		"inbound.calls.status.204",
	}
	endpointTags := map[string]string{
		"env":      "test",
		"service":  "test-gateway",
		"endpoint": "baz",
		"handler":  "call",
	}
	for _, name := range endpointNames {
		key := tally.KeyForPrefixedStringMap(name, endpointTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	latencyMetric := metrics[tally.KeyForPrefixedStringMap("inbound.calls.latency", endpointTags)]
	value := *latencyMetric.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	recvdMetric := metrics[tally.KeyForPrefixedStringMap("inbound.calls.recvd", endpointTags)]
	value = *recvdMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	successMetric := metrics[tally.KeyForPrefixedStringMap("inbound.calls.success", endpointTags)]
	value = *successMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	statusMetric := metrics[tally.KeyForPrefixedStringMap("inbound.calls.status.204", endpointTags)]
	value = *statusMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	tchannelNames := []string{
		"tchannel.outbound.calls.latency",
		"tchannel.outbound.calls.per-attempt.latency",
		"tchannel.outbound.calls.send",
		"tchannel.outbound.calls.success",
	}
	tchannelTags := map[string]string{
		"env":             "test",
		"app":             "test-gateway",
		"service":         "test-gateway",
		"target-service":  "bazService",
		"target-endpoint": "SimpleService::call",
		"host":            zanzibar.GetHostname(),
	}
	for _, name := range tchannelNames {
		key := tally.KeyForPrefixedStringMap(name, tchannelTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	outboundLatency := metrics[tally.KeyForPrefixedStringMap("tchannel.outbound.calls.latency", tchannelTags)]
	value = *outboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	perAttemptOutboundLatency := metrics[tally.KeyForPrefixedStringMap("tchannel.outbound.calls.per-attempt.latency", tchannelTags)]
	value = *perAttemptOutboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	outboundSend := metrics[tally.KeyForPrefixedStringMap("tchannel.outbound.calls.send", tchannelTags)]
	value = *outboundSend.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	outboundSuccess := metrics[tally.KeyForPrefixedStringMap("tchannel.outbound.calls.success", tchannelTags)]
	value = *outboundSuccess.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	clientNames := []string{
		"outbound.calls.latency",
		"outbound.calls.sent",
		"outbound.calls.success",
	}
	clientTags := map[string]string{
		"env":             "test",
		"service":         "test-gateway",
		"client":          "baz",
		"method":          "Call",
		"target-service":  "bazService",
		"target-endpoint": "SimpleService::call",
	}
	for _, name := range clientNames {
		key := tally.KeyForPrefixedStringMap(name, clientTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	outboundLatency = metrics[tally.KeyForPrefixedStringMap("outbound.calls.latency", clientTags)]
	value = *outboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	outboundSend = metrics[tally.KeyForPrefixedStringMap("outbound.calls.sent", clientTags)]
	value = *outboundSend.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	outboundSuccess = metrics[tally.KeyForPrefixedStringMap("outbound.calls.success", clientTags)]
	value = *outboundSuccess.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")
}
