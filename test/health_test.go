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

package gateway_test

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	"github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
	"github.com/uber/zanzibar/test/lib/test_gateway"
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
	numMetrics := 5
	cgateway.MetricsWaitGroup.Add(numMetrics)

	res, err := gateway.MakeRequest("GET", "/health", nil, nil)
	if !assert.NoError(t, err, "got http error") {
		return
	}
	assert.Equal(t, res.Status, "200 OK", "got http 200")

	cgateway.MetricsWaitGroup.Wait()

	// sleep to avoid race conditions with m3.Client
	time.Sleep(100 * time.Millisecond)
	metrics := cgateway.M3Service.GetMetrics()
	assert.Equal(t, numMetrics, len(metrics), "expected 5 metrics")
	names := []string{
		"inbound.calls.latency",
		"inbound.calls.recvd",
		"inbound.calls.success",
		"inbound.calls.status.200",
	}
	tags := map[string]string{
		"env":      "test",
		"service":  "test-gateway",
		"endpoint": "health",
		"handler":  "health",
	}
	defaultTags := map[string]string{
		"env":     "test",
		"service": "test-gateway",
	}

	for _, name := range names {
		key := tally.KeyForPrefixedStringMap(name, tags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	loggedKey := tally.KeyForPrefixedStringMap("zap.logged.info", defaultTags)
	assert.Contains(t, metrics, loggedKey, "expected metrics: %s", loggedKey)

	latencyMetric := metrics[tally.KeyForPrefixedStringMap("inbound.calls.latency", tags)]
	value := *latencyMetric.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	recvdMetric := metrics[tally.KeyForPrefixedStringMap("inbound.calls.recvd", tags)]
	value = *recvdMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	successMetric := metrics[tally.KeyForPrefixedStringMap("inbound.calls.success", tags)]
	value = *successMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	statusMetric := metrics[tally.KeyForPrefixedStringMap("inbound.calls.status.200", tags)]
	value = *statusMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	loggedMetrics := metrics[tally.KeyForPrefixedStringMap("zap.logged.info", defaultTags)]
	value = *loggedMetrics.MetricValue.Count.I64Value
	assert.Equal(t, int64(2), value, "expected counter to be 2")
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

	// Expect 30 runtime metrics + 1 logged metric
	numMetrics := 31
	cgateway.MetricsWaitGroup.Add(numMetrics)
	cgateway.MetricsWaitGroup.Wait()

	metrics := cgateway.M3Service.GetMetrics()
	assert.Equal(t, numMetrics, len(metrics), "expected 31 metrics")
	names := []string{
		"per-worker.runtime.cpu.cgoCalls",
		"per-worker.runtime.cpu.count",
		"per-worker.runtime.cpu.goMaxProcs",
		"per-worker.runtime.cpu.goroutines",
		"per-worker.runtime.mem.alloc",
		"per-worker.runtime.mem.frees",
		"per-worker.runtime.mem.gc.count",
		"per-worker.runtime.mem.gc.cpuFraction",
		"per-worker.runtime.mem.gc.last",
		"per-worker.runtime.mem.gc.next",
		"per-worker.runtime.mem.gc.pause",
		"per-worker.runtime.mem.gc.pauseTotal",
		"per-worker.runtime.mem.gc.sys",
		"per-worker.runtime.mem.heap.alloc",
		"per-worker.runtime.mem.heap.idle",
		"per-worker.runtime.mem.heap.inuse",
		"per-worker.runtime.mem.heap.objects",
		"per-worker.runtime.mem.heap.released",
		"per-worker.runtime.mem.heap.sys",
		"per-worker.runtime.mem.lookups",
		"per-worker.runtime.mem.malloc",
		"per-worker.runtime.mem.otherSys",
		"per-worker.runtime.mem.stack.inuse",
		"per-worker.runtime.mem.stack.mcacheInuse",
		"per-worker.runtime.mem.stack.mcacheSys",
		"per-worker.runtime.mem.stack.mspanInuse",
		"per-worker.runtime.mem.stack.mspanSys",
		"per-worker.runtime.mem.stack.sys",
		"per-worker.runtime.mem.sys",
		"per-worker.runtime.mem.total",
	}
	tags := map[string]string{
		"env":     "test",
		"service": "test-gateway",
		"host":    zanzibar.GetHostname(),
	}
	for _, name := range names {
		key := tally.KeyForPrefixedStringMap(name, tags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}
}
