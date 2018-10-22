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

package bar_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	zanzibar "github.com/uber/zanzibar/runtime"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
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
				"mapIntWithoutRange": {},
				"binaryField": "d29ybGQ="
			}`)); err != nil {
				t.Fatal("can't write fake response")
			}
		},
	)

	numMetrics := 13
	cg := gateway.(*testGateway.ChildProcessGateway)
	cg.MetricsWaitGroup.Add(numMetrics)

	headers := make(map[string]string)
	headers["regionname"] = "san_francisco"
	headers["device"] = "ios"
	headers["deviceversion"] = "carbon"
	_, err = gateway.MakeRequest(
		"POST", "/bar/bar-path", headers,
		bytes.NewReader([]byte(`{
			"request":{"stringField":"foo","boolField":true,"binaryField":"aGVsbG8=","timestamp":123,"enumField":0,"longField":123}
		}`)),
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
		"endpointid":    "bar",
		"handlerid":     "normal",
		"regionname":    "san_francisco",
		"device":        "ios",
		"deviceversion": "carbon",
		"dc":            "unknown",
		"host":          zanzibar.GetHostname(),
		"protocal":      "HTTP",
	}
	eStatusTags := map[string]string{
		"env":           "test",
		"service":       "test-gateway",
		"status":        "200",
		"endpointid":    "bar",
		"handlerid":     "normal",
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

	inboundStatus := metrics[tally.KeyForPrefixedStringMap(
		"endpoint.status", eStatusTags,
	)]
	value = *inboundStatus.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	httpClientNames := []string{
		"client.latency",
		"client.request",
	}
	httpClientTags := map[string]string{
		"env":           "test",
		"service":       "test-gateway",
		"clientid":      "bar",
		"clientmethod":  "Normal",
		"dc":            "unknown",
		"host":          zanzibar.GetHostname(),
		"endpointid":    "bar",
		"handlerid":     "normal",
		"regionname":    "san_francisco",
		"device":        "ios",
		"deviceversion": "carbon",
		"protocal":      "HTTP",
	}
	cStatusTags := map[string]string{
		"env":           "test",
		"service":       "test-gateway",
		"clientid":      "bar",
		"clientmethod":  "Normal",
		"status":        "200",
		"dc":            "unknown",
		"host":          zanzibar.GetHostname(),
		"endpointid":    "bar",
		"handlerid":     "normal",
		"regionname":    "san_francisco",
		"device":        "ios",
		"deviceversion": "carbon",
		"protocal":      "HTTP",
	}

	for _, name := range httpClientNames {
		key := tally.KeyForPrefixedStringMap(name, httpClientTags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	outboundLatency := metrics[tally.KeyForPrefixedStringMap(
		"client.latency", httpClientTags,
	)]
	value = *outboundLatency.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	outboundSent := metrics[tally.KeyForPrefixedStringMap(
		"client.request", httpClientTags,
	)]
	value = *outboundSent.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	statusSuccess := metrics[tally.KeyForPrefixedStringMap(
		"client.status",
		cStatusTags,
	)]
	value = *statusSuccess.MetricValue.Count.I64Value
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
	assert.Equal(t, int64(3), value, "expected counter to be 3")

	allLogs := gateway.AllLogs()

	logMsgs := allLogs["Finished an outgoing client HTTP request"]
	assert.Len(t, logMsgs, 1)
	logMsg := logMsgs[0]
	dynamicHeaders := []string{
		"url",
		"timestamp-finished",
		"Request-Header-Uber-Trace-Id",
		"Response-Header-Content-Length",
		"timestamp-started",
		"Response-Header-Date",
		"ts",
		"hostname",
		"pid",
		"requestUUID",
	}
	for _, dynamicValue := range dynamicHeaders {
		assert.Contains(t, logMsg, dynamicValue)
		delete(logMsg, dynamicValue)
	}
	expectedValues := map[string]interface{}{
		"msg":                          "Finished an outgoing client HTTP request",
		"env":                          "test",
		"clientID":                     "bar",
		"statusCode":                   float64(200),
		"Request-Header-Content-Type":  "application/json",
		"Response-Header-Content-Type": "text/plain; charset=utf-8",

		"level":                      "info",
		"methodName":                 "Normal",
		"method":                     "POST",
		"Request-Header-X-Client-Id": "bar",

		"zone":       "unknown",
		"service":    "example-gateway",
		"endpointID": "bar",
		"handlerID":  "normal",
	}
	for actualKey, actualValue := range logMsg {
		assert.Equal(t, expectedValues[actualKey], actualValue, "unexpected header %q", actualKey)
	}
	for expectedKey, expectedValue := range expectedValues {
		assert.Equal(t, expectedValue, logMsg[expectedKey], "unexpected header %q", expectedKey)
	}
}
