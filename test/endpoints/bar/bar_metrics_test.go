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

package bar_test

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
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

	numMetrics := 16
	cg := gateway.(*testGateway.ChildProcessGateway)
	cg.MetricsWaitGroup.Add(numMetrics)

	headers := make(map[string]string)
	headers["x-uuid"] = "uuid"
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
	cbKeys := make([]string, 0)
	for key := range metrics {
		if strings.Contains(key, "circuitbreaker") {
			cbKeys = append(cbKeys, key)
		}
	}
	assert.Equal(t, 6, len(cbKeys)) // number off because of gratuitous histogram metric
	for key := range metrics {
		if strings.HasPrefix(key, "jaeger") {
			delete(metrics, key)
		}
	}
	assert.Equal(t, numMetrics+4, len(metrics))

	endpointTags := map[string]string{
		"env":            "test",
		"service":        "test-gateway",
		"endpointid":     "bar",
		"handlerid":      "normal",
		"regionname":     "san_francisco",
		"device":         "ios",
		"deviceversion":  "carbon",
		"dc":             "unknown",
		"protocol":       "HTTP",
		"apienvironment": "production",
	}
	statusTags := map[string]string{
		"status":     "200",
		"clienttype": "http",
	}
	for k, v := range endpointTags {
		statusTags[k] = v
	}

	key := tally.KeyForPrefixedStringMap("endpoint.request", endpointTags)
	assert.Contains(t, metrics, key, "expected metric: %s", key)
	assert.Equal(t, int64(1), metrics[key].Value.Count)

	key = tally.KeyForPrefixedStringMap("endpoint.latency", statusTags)
	assert.Contains(t, metrics, key, "expected metric: %s", key)
	value := metrics[key].Value.Timer

	assert.True(t, value > 1000, fmt.Sprintf("expected latency > 1000 nano seconds but found to be: %d", value))
	assert.True(t, value < 50*1000*1000, fmt.Sprintf("expected latency < 50 milli seconds but found to be: %d", value))

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

	inboundStatus := metrics[tally.KeyForPrefixedStringMap(
		"endpoint.status", statusTags,
	)]
	value = inboundStatus.Value.Count
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	httpClientTags := map[string]string{
		"env":            "test",
		"service":        "test-gateway",
		"clientid":       "bar",
		"clientmethod":   "Normal",
		"targetendpoint": "Bar--normal",
		"dc":             "unknown",
		"endpointid":     "bar",
		"handlerid":      "normal",
		"regionname":     "san_francisco",
		"device":         "ios",
		"deviceversion":  "carbon",
		"protocol":       "HTTP",
		"apienvironment": "production",
	}
	cStatusTags := map[string]string{
		"status": "200",
	}
	for k, v := range httpClientTags {
		cStatusTags[k] = v
	}

	key = tally.KeyForPrefixedStringMap("client.request", httpClientTags)
	assert.Contains(t, metrics, key, "expected metric: %s", key)
	assert.Equal(t, int64(1), metrics[key].Value.Count, "expected counter to be 1")

	key = tally.KeyForPrefixedStringMap("client.status", cStatusTags)
	assert.Contains(t, metrics, key, "expected metric: %s", key)
	assert.Equal(t, int64(1), metrics[key].Value.Count, "expected counter to be 1")

	key = tally.KeyForPrefixedStringMap("client.latency", httpClientTags)
	assert.Contains(t, metrics, key, "expected metric: %s", key)
	value = metrics[key].Value.Timer
	assert.True(t, value > 1000, "expected latency > 1000 nano second")
	assert.True(t, value < 50*1000*1000, "expected latency < 50 milli second")

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

	allLogs := gateway.AllLogs()

	logMsgs := allLogs["Finished an outgoing client HTTP request"]
	assert.Len(t, logMsgs, 1)
	logMsg := logMsgs[0]
	dynamicHeaders := []string{
		"url",
		"timestamp-finished",
		"Client-Req-Header-Uber-Trace-Id",
		"Client-Req-Header-X-Request-Uuid",
		"Client-Res-Header-Content-Length",
		"Client-Res-Header-Date",
		"timestamp-started",
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
		"env":        "test",
		"level":      "debug",
		"msg":        "Finished an outgoing client HTTP request",
		"statusCode": float64(200),
		//"clientID":                       "bar",
		//"clientMethod":                   "Normal",
		//"clientThriftMethod":             "Bar::normal",
		"clientHTTPMethod":               "POST",
		"Client-Req-Header-X-Client-Id":  "bar",
		"Client-Req-Header-Content-Type": "application/json",
		"Client-Req-Header-Accept":       "application/json",
		"Client-Res-Header-Content-Type": "text/plain; charset=utf-8",

		"zone":            "unknown",
		"service":         "example-gateway",
		"endpointID":      "bar",
		"endpointHandler": "normal",
		"X-Uuid":          "uuid",
		"Regionname":      "san_francisco",
		"Device":          "ios",
		"Deviceversion":   "carbon",
		"Content-Length":  "128",
		"User-Agent":      "Go-http-client/1.1",
		"Accept-Encoding": "gzip",
		"apienvironment":  "production",
	}
	for actualKey, actualValue := range logMsg {
		assert.Equal(t, expectedValues[actualKey], actualValue, "unexpected field %q", actualKey)
	}
	for expectedKey, expectedValue := range expectedValues {
		assert.Equal(t, expectedValue, logMsg[expectedKey], "unexpected field %q", expectedKey)
	}
}
