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

package bar_test

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

func TestCallMetrics(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		CountMetrics:      true,
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte(`{
				"stringField": "stringValue",
				"intWithRange": 0,
				"intWithoutRange": 0,
				"mapIntWithRange": {},
				"mapIntWithoutRange": {}
			}`)); err != nil {
				t.Fatal("can't write fake response")
			}
		},
	)

	numMetrics := 9
	cg := gateway.(*testGateway.ChildProcessGateway)
	cg.MetricsWaitGroup.Add(numMetrics)

	_, err = gateway.MakeRequest(
		"POST", "/bar/bar-path", nil,
		bytes.NewReader([]byte(`{
			"request":{"stringField":"foo","boolField":true}
		}`)),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	cg.MetricsWaitGroup.Wait()
	time.Sleep(100 * time.Millisecond)
	metrics := cg.M3Service.GetMetrics()
	assert.Equal(t, numMetrics, len(metrics))

	endpointNames := []string{
		"inbound.calls.latency",
		"inbound.calls.recvd",
		"inbound.calls.success",
		"inbound.calls.status.200",
	}
	endpointTags := map[string]string{
		"env":      "test",
		"service":  "test-gateway",
		"endpoint": "bar",
		"handler":  "normal",
	}
	for _, name := range endpointNames {
		key := tally.KeyForPrefixedStringMap(name, endpointTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	inboundLatency := metrics[tally.KeyForPrefixedStringMap("inbound.calls.latency", endpointTags)]
	value := *inboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	inboundRecvd := metrics[tally.KeyForPrefixedStringMap("inbound.calls.recvd", endpointTags)]
	value = *inboundRecvd.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value)

	inboundStatus := metrics[tally.KeyForPrefixedStringMap("inbound.calls.status.200", endpointTags)]
	value = *inboundStatus.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	httpClientNames := []string{
		"outbound.calls.latency",
		"outbound.calls.sent",
		"outbound.calls.success",
		"outbound.calls.status.200",
	}
	httpClientTags := map[string]string{
		"env":     "test",
		"service": "test-gateway",
		"client":  "bar",
		"method":  "Normal",
	}
	for _, name := range httpClientNames {
		key := tally.KeyForPrefixedStringMap(name, httpClientTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	outboundLatency := metrics[tally.KeyForPrefixedStringMap("outbound.calls.latency", httpClientTags)]
	value = *outboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	outboundSent := metrics[tally.KeyForPrefixedStringMap("outbound.calls.sent", httpClientTags)]
	value = *outboundSent.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	outboundSuccess := metrics[tally.KeyForPrefixedStringMap("outbound.calls.success", httpClientTags)]
	value = *outboundSuccess.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	statusSuccess := metrics[tally.KeyForPrefixedStringMap("outbound.calls.status.200", httpClientTags)]
	value = *statusSuccess.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	defaultTags := map[string]string{
		"env":     "test",
		"service": "test-gateway",
	}

	loggedMetrics := metrics[tally.KeyForPrefixedStringMap("zap.logged.info", defaultTags)]
	value = *loggedMetrics.MetricValue.Count.I64Value
	assert.Equal(t, int64(5), value, "expected counter to be 5")
}
