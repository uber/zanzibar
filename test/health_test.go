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

package gateway_test

import (
	"io/ioutil"
	"testing"

	"github.com/uber-go/tally/m3"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	zanzibar "github.com/uber/zanzibar/runtime"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

func TestHealthCall(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		TestBinary:  util.DefaultMainFile("example-gateway"),
		ConfigFiles: util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "must be able to create gateway") {
		return
	}
	defer gateway.Close()

	assert.NotNil(t, gateway, "gateway exists")

	res, err := gateway.MakeRequest("GET", "/health", nil, nil)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, res.Status, "200 OK", "got http 200")
}

func BenchmarkHealthCall(b *testing.B) {
	gateway, err := benchGateway.CreateGateway(
		map[string]interface{}{
			"clients.baz.serviceName": "baz",
		},
		&testGateway.Options{
			TestBinary:            util.DefaultMainFile("example-gateway"),
			ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
			KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
			KnownTChannelBackends: []string{"baz"},
		},
		exampleGateway.CreateGateway,
	)
	if err != nil {
		b.Error("got bootstrap err: " + err.Error())
		return
	}

	b.ResetTimer()

	// b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := gateway.MakeRequest("GET", "/health", nil, nil)
			if err != nil {
				b.Error("got http error: " + err.Error())
				break
			}
			if res.Status != "200 OK" {
				b.Error("got bad status error: " + res.Status)
				break
			}

			_, err = ioutil.ReadAll(res.Body)
			if err != nil {
				b.Error("could not write response: " + res.Status)
				break
			}
		}
	})

	b.StopTimer()
	gateway.Close()
	b.StartTimer()
}

func TestHealthMetrics(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		CountMetrics: true,
		TestBinary:   util.DefaultMainFile("example-gateway"),
		ConfigFiles:  util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "must be able to create gateway") {
		return
	}
	defer gateway.Close()

	cgateway := gateway.(*testGateway.ChildProcessGateway)
	numMetrics := 11
	cgateway.MetricsWaitGroup.Add(numMetrics)

	headers := make(map[string]string)
	headers["regionname"] = "san_francisco"
	headers["device"] = "ios"
	headers["deviceversion"] = "carbon"

	res, err := gateway.MakeRequest("GET", "/health", headers, nil)
	if !assert.NoError(t, err, "got http error") {
		return
	}
	assert.Equal(t, res.Status, "200 OK", "got http 200")

	cgateway.MetricsWaitGroup.Wait()

	metrics := cgateway.M3Service.GetMetrics()
	assert.Equal(t, numMetrics, len(metrics))
	tags := map[string]string{
		"env":            "test",
		"service":        "test-gateway",
		"endpointid":     "health",
		"handlerid":      "health",
		"regionname":     "san_francisco",
		"device":         "ios",
		"deviceversion":  "carbon",
		"dc":             "unknown",
		"protocol":       "HTTP",
		"apienvironment": "production",
	}
	statusTags := map[string]string{
		"status":     "200",
		"clienttype": "",
	}
	for k, v := range tags {
		statusTags[k] = v
	}
	histogramTags := map[string]string{
		m3.DefaultHistogramBucketName:   "0-10ms", // TODO(argouber): There must be a better way than this hard-coding
		m3.DefaultHistogramBucketIDName: "0001",
	}
	for k, v := range statusTags {
		histogramTags[k] = v
	}

	key := tally.KeyForPrefixedStringMap("endpoint.request", tags)
	assert.Contains(t, metrics, key, "expected metric: %s", key)

	key = tally.KeyForPrefixedStringMap("endpoint.latency", statusTags)
	assert.Contains(t, metrics, key, "expected metric: %s", key)
	key = tally.KeyForPrefixedStringMap("endpoint.latency-hist", histogramTags)
	assert.Contains(t, metrics, key, "expected metric: %s", key)

	statusKey := tally.KeyForPrefixedStringMap(
		"endpoint.status", statusTags,
	)
	assert.Contains(t, metrics, statusKey, "expected metrics: %s", statusKey)

	latencyMetric := metrics[tally.KeyForPrefixedStringMap("endpoint.latency", statusTags)]
	value := latencyMetric.Value.Timer
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 10*1000*1000, "expected timer to be <10 milli seconds")

	latencyHistMetric := metrics[tally.KeyForPrefixedStringMap("endpoint.latency-hist", histogramTags)]
	value = latencyHistMetric.Value.Count
	assert.Equal(t, int64(1), value)

	recvdMetric := metrics[tally.KeyForPrefixedStringMap(
		"endpoint.request", tags,
	)]
	value = recvdMetric.Value.Count
	assert.Equal(t, int64(1), value)

	statusMetric := metrics[tally.KeyForPrefixedStringMap(
		"endpoint.status", statusTags,
	)]
	value = statusMetric.Value.Count
	assert.Equal(t, int64(1), value)
}

func TestRuntimeMetrics(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		CountMetrics:         true,
		EnableRuntimeMetrics: true,
		MaxMetrics:           31,
		TestBinary:           util.DefaultMainFile("example-gateway"),
		ConfigFiles:          util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "must be able to create gateway") {
		return
	}
	defer gateway.Close()

	cgateway := gateway.(*testGateway.ChildProcessGateway)

	names := []string{
		"runtime.num-cpu",
		"runtime.gomaxprocs",
		"runtime.num-goroutines",

		"runtime.memory.heap",
		"runtime.memory.heapidle",
		"runtime.memory.heapinuse",
		"runtime.memory.stack",

		"runtime.memory.num-gc",
		"runtime.memory.gc-pause-ms",
	}
	histogramName := "runtime.memory.gc-pause-ms-hist"

	// this is a shame because first GC takes 20s to kick in
	// only then gc stats can be collected
	// oh and the magic number 2 are 2 other stats produced
	cgateway.MetricsWaitGroup.Add(len(names) + 2)
	cgateway.MetricsWaitGroup.Wait()

	metrics := cgateway.M3Service.GetMetrics()

	tags := map[string]string{
		"env":     "test",
		"service": "test-gateway",
		"host":    zanzibar.GetHostname(),
		"dc":      "unknown",
	}
	for _, name := range names {
		key := tally.KeyForPrefixedStringMap(name, tags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}
	histogramTags := map[string]string{
		m3.DefaultHistogramBucketName:   "0-10ms",
		m3.DefaultHistogramBucketIDName: "0001",
	}
	for k, v := range tags {
		histogramTags[k] = v
	}
	assert.Contains(t, metrics, tally.KeyForPrefixedStringMap(histogramName, histogramTags))
}
