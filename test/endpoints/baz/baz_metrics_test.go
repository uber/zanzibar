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
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/test/lib"
	"github.com/uber/zanzibar/test/lib/test_gateway"

	bazClient "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
)

func TestCallMetrics(t *testing.T) {
	testcallCounter := 0

	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"clients.baz.serviceName": "bazService",
	}, &testGateway.Options{
		CountMetrics:          true,
		KnownTChannelBackends: []string{"baz"},
		TestBinary: filepath.Join(
			getDirName(), "..", "..", "..", "examples", "example-gateway",
			"build", "services", "example-gateway", "main.go",
		),
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
		"SimpleService",
		"call",
		bazClient.NewSimpleServiceCallHandler(fakeCall),
	)

	headers := map[string]string{}
	headers["x-token"] = "token"
	headers["x-uuid"] = "uuid"

	numMetrics := 11
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
	sort.Sort(lib.SortMetricsByNameAndTags(metrics))
	assert.Equal(t, numMetrics, len(metrics))

	expectedEndpoitTags := map[string]string{
		"endpoint": "baz",
		"handler":  "call",
	}

	expectedTchannelTags := map[string]string{
		"app":             "test-gateway",
		"service":         "test-gateway",
		"target-service":  "bazService",
		"target-endpoint": "SimpleService::call",
	}

	latencyMetric := metrics[0]
	assert.Equal(t, "test-gateway.production.all-workers.inbound.calls.latency", latencyMetric.GetName())

	value := *latencyMetric.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	tags := latencyMetric.GetTags()
	assert.Equal(t, 2, len(tags), "expected 2 tags")

	for tag := range tags {
		assert.Equal(t, expectedEndpoitTags[tag.GetTagName()], tag.GetTagValue())
	}

	recvdMetric := metrics[1]
	assert.Equal(t, "test-gateway.production.all-workers.inbound.calls.recvd", recvdMetric.GetName())

	value = *recvdMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value)

	tags = recvdMetric.GetTags()
	assert.Equal(t, 2, len(tags), "expected 2 tags")

	for tag := range tags {
		assert.Equal(t, expectedEndpoitTags[tag.GetTagName()], tag.GetTagValue())
	}

	statusCodeMetric := metrics[2]
	assert.Equal(t,
		"test-gateway.production.all-workers.inbound.calls.status.204", statusCodeMetric.GetName(),
	)

	value = *statusCodeMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	tags = statusCodeMetric.GetTags()
	assert.Equal(t, 2, len(tags), "expected 2 tags")
	for tag := range tags {
		assert.Equal(t, expectedEndpoitTags[tag.GetTagName()], tag.GetTagValue())
	}

	expectedJaegerSpanMetricName := "test-gateway.production.all-workers.jaeger.spans"

	jaegerSpanFinishedMetric := metrics[3]
	assert.Equal(t, expectedJaegerSpanMetricName, jaegerSpanFinishedMetric.GetName())

	value = *jaegerSpanFinishedMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value)

	tags = jaegerSpanFinishedMetric.GetTags()
	assert.Equal(t, 2, len(tags), "expected 2 tags")

	expectedSpanFinishedTags := map[string]string{
		"group": "lifecycle",
		"state": "finished",
	}
	for tag := range tags {
		assert.Equal(t, expectedSpanFinishedTags[tag.GetTagName()], tag.GetTagValue())
	}

	jaegerSpanStartedMetric := metrics[4]
	assert.Equal(t, expectedJaegerSpanMetricName, jaegerSpanStartedMetric.GetName())

	value = *jaegerSpanStartedMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value)

	tags = jaegerSpanStartedMetric.GetTags()
	assert.Equal(t, 2, len(tags), "expected 2 tags")

	expectedSpanStartedTags := map[string]string{
		"group": "lifecycle",
		"state": "started",
	}
	for tag := range tags {
		assert.Equal(t, expectedSpanStartedTags[tag.GetTagName()], tag.GetTagValue())
	}

	jaegerSpanSamplingMetric := metrics[5]
	assert.Equal(t, expectedJaegerSpanMetricName, jaegerSpanSamplingMetric.GetName())

	value = *jaegerSpanSamplingMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value)

	tags = jaegerSpanSamplingMetric.GetTags()
	assert.Equal(t, 2, len(tags), "expected 2 tags")

	expectedSpanSamplingTags := map[string]string{
		"group":   "sampling",
		"sampled": "n",
	}
	for tag := range tags {
		assert.Equal(t, expectedSpanSamplingTags[tag.GetTagName()], tag.GetTagValue())
	}

	jaegerTraceMetric := metrics[6]
	assert.Equal(t, "test-gateway.production.all-workers.jaeger.traces", jaegerTraceMetric.GetName())

	value = *jaegerTraceMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value)

	tags = jaegerTraceMetric.GetTags()
	assert.Equal(t, 2, len(tags), "expected 2 tags")

	expectedTraceTags := map[string]string{
		"state":   "started",
		"sampled": "n",
	}
	for tag := range tags {
		assert.Equal(t, expectedTraceTags[tag.GetTagName()], tag.GetTagValue())
	}

	tchannelOutboundSuccessMetric := metrics[7]
	assert.Equal(t,
		"test-gateway.production.all-workers.tchannel.outbound.calls.latency",
		tchannelOutboundSuccessMetric.GetName(),
	)

	value = *latencyMetric.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	tags = tchannelOutboundSuccessMetric.GetTags()
	assert.Equal(t, 5, len(tags), "expected 2 tags")

	for tag := range tags {
		if tag.GetTagName() == "host" {
			continue
		}
		assert.Equal(t, expectedTchannelTags[tag.GetTagName()], tag.GetTagValue())
	}

	tchannelOutboundSuccessMetric = metrics[8]
	assert.Equal(t,
		"test-gateway.production.all-workers.tchannel.outbound.calls.per-attempt.latency",
		tchannelOutboundSuccessMetric.GetName(),
	)

	value = *latencyMetric.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	tags = tchannelOutboundSuccessMetric.GetTags()
	assert.Equal(t, 5, len(tags), "expected 2 tags")
	for tag := range tags {
		if tag.GetTagName() == "host" {
			continue
		}
		assert.Equal(t, expectedTchannelTags[tag.GetTagName()], tag.GetTagValue())
	}

	tchannelOutboundSuccessMetric = metrics[9]
	assert.Equal(t,
		"test-gateway.production.all-workers.tchannel.outbound.calls.send",
		tchannelOutboundSuccessMetric.GetName(),
	)

	value = *tchannelOutboundSuccessMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	tags = tchannelOutboundSuccessMetric.GetTags()
	assert.Equal(t, 5, len(tags), "expected 2 tags")
	for tag := range tags {
		if tag.GetTagName() == "host" {
			continue
		}
		assert.Equal(t, expectedTchannelTags[tag.GetTagName()], tag.GetTagValue())
	}

	tchannelOutboundSuccessMetric = metrics[10]
	assert.Equal(t,
		"test-gateway.production.all-workers.tchannel.outbound.calls.success",
		tchannelOutboundSuccessMetric.GetName(),
	)

	value = *tchannelOutboundSuccessMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	tags = tchannelOutboundSuccessMetric.GetTags()
	assert.Equal(t, 5, len(tags), "expected 2 tags")
	for tag := range tags {
		if tag.GetTagName() == "host" {
			continue
		}
		assert.Equal(t, expectedTchannelTags[tag.GetTagName()], tag.GetTagValue())
	}
}
