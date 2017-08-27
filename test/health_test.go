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

	// Expect three metrics
	numMetrics := 3
	cgateway.MetricsWaitGroup.Add(numMetrics)

	res, err := gateway.MakeRequest("GET", "/health", nil, nil)
	if !assert.NoError(t, err, "got http error") {
		return
	}
	assert.Equal(t, res.Status, "200 OK", "got http 200")

	cgateway.MetricsWaitGroup.Wait()

	metrics := cgateway.M3Service.GetMetrics()
	assert.Equal(t, numMetrics, len(metrics), "expected 3 metrics")
	names := []string{
		"inbound.calls.latency",
		"inbound.calls.recvd",
		"inbound.calls.status.200",
	}
	tags := map[string]string{
		"endpoint": "health",
		"handler":  "health",
		"service":  "test-gateway",
		"env":      "test",
		"host":     "all",
	}
	for _, name := range names {
		key := tally.KeyForPrefixedStringMap(name, tags)
		assert.Contains(t, metrics, key, "expected metric: %s", key)
	}

	latencyMetric := metrics[tally.KeyForPrefixedStringMap("inbound.calls.latency", tags)]
	value := *latencyMetric.MetricValue.Timer.I64Value
	assert.True(t, value > 1000, "expected timer to be >1000 nano seconds")
	assert.True(t, value < 1000*1000*1000, "expected timer to be <1 second")

	recvdMetric := metrics[tally.KeyForPrefixedStringMap("inbound.calls.recvd", tags)]
	value = *recvdMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")

	statusCodeMetric := metrics[tally.KeyForPrefixedStringMap("inbound.calls.status.200", tags)]
	value = *statusCodeMetric.MetricValue.Count.I64Value
	assert.Equal(t, int64(1), value, "expected counter to be 1")
}

func TestRuntimeMetrics(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		CountMetrics:         true,
		EnableRuntimeMetrics: true,
		TestBinary:           util.DefaultMainFile("example-gateway"),
		ConfigFiles:          util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "must be able to create gateway") {
		return
	}
	defer gateway.Close()

	cgateway := gateway.(*testGateway.ChildProcessGateway)

	// Expect 30 runtime metrics
	numMetrics := 30
	cgateway.MetricsWaitGroup.Add(numMetrics)
	cgateway.MetricsWaitGroup.Wait()

	metrics := cgateway.M3Service.GetMetrics()
	assert.Equal(t, numMetrics, len(metrics), "expected 30 metrics")
	names := []string{
		"runtime.cpu.cgoCalls",
		"runtime.cpu.count",
		"runtime.cpu.goMaxProcs",
		"runtime.cpu.goroutines",
		"runtime.mem.alloc",
		"runtime.mem.frees",
		"runtime.mem.gc.count",
		"runtime.mem.gc.cpuFraction",
		"runtime.mem.gc.last",
		"runtime.mem.gc.next",
		"runtime.mem.gc.pause",
		"runtime.mem.gc.pauseTotal",
		"runtime.mem.gc.sys",
		"runtime.mem.heap.alloc",
		"runtime.mem.heap.idle",
		"runtime.mem.heap.inuse",
		"runtime.mem.heap.objects",
		"runtime.mem.heap.released",
		"runtime.mem.heap.sys",
		"runtime.mem.lookups",
		"runtime.mem.malloc",
		"runtime.mem.otherSys",
		"runtime.mem.stack.inuse",
		"runtime.mem.stack.mcacheInuse",
		"runtime.mem.stack.mcacheSys",
		"runtime.mem.stack.mspanInuse",
		"runtime.mem.stack.mspanSys",
		"runtime.mem.stack.sys",
		"runtime.mem.sys",
		"runtime.mem.total",
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
