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
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	m3 "github.com/uber-go/tally/m3/thrift"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
	"github.com/uber/zanzibar/test/lib/test_gateway"

	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints"
)

var testBinary = filepath.Join(
	getDirName(),
	"..", "examples", "example-gateway",
	"build", "services", "example-gateway", "main.go",
)

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(file)
}

func TestHealthCall(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		TestBinary: testBinary,
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
		nil,
		nil,
		clients.CreateClients,
		endpoints.Register,
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

type SortMetricByName []*m3.Metric

func (a SortMetricByName) Len() int {
	return len(a)
}
func (a SortMetricByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a SortMetricByName) Less(i, j int) bool {
	return a[i].GetName() < a[j].GetName()
}

func TestHealthMetrics(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		CountMetrics: true,
		TestBinary:   testBinary,
	})
	if !assert.NoError(t, err, "must be able to create gateway") {
		return
	}
	defer gateway.Close()

	cgateway := gateway.(*testGateway.ChildProcessGateway)

	// Expect three metrics
	cgateway.MetricsWaitGroup.Add(3)

	res, err := gateway.MakeRequest("GET", "/health", nil, nil)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, res.Status, "200 OK", "got http 200")

	cgateway.MetricsWaitGroup.Wait()
	metrics := cgateway.M3Service.GetMetrics()
	sort.Sort(SortMetricByName(metrics))

	assert.Equal(t, len(metrics), 3, "expected one metric")

	latencyMetric := metrics[0]

	assert.Equal(t,
		"test-gateway.production.all-workers.inbound.calls.latency",
		latencyMetric.GetName(),
		"expected correct name",
	)

	value := *latencyMetric.MetricValue.Timer.I64Value

	assert.True(t, value > 1000,
		"expected timer to be >1000 nano seconds",
	)
	assert.True(t, value < 1000*1000*1000,
		"expected timer to be <1 second",
	)

	tags := latencyMetric.GetTags()
	assert.Equal(t, 2, len(tags), "expected 2 tags")
	expectedTags := map[string]string{
		"endpoint": "health",
		"handler":  "health",
	}
	for tag := range tags {
		assert.Equal(t,
			expectedTags[tag.GetTagName()],
			tag.GetTagValue(),
			"expected tag value to be correct",
		)
	}

	recvdMetric := metrics[1]

	assert.Equal(t,
		"test-gateway.production.all-workers.inbound.calls.recvd",
		recvdMetric.GetName(),
		"expected correct name",
	)

	value = *recvdMetric.MetricValue.Count.I64Value
	assert.Equal(t,
		int64(1),
		value,
		"expected counter to be 1",
	)

	tags = recvdMetric.GetTags()
	assert.Equal(t, 2, len(tags), "expected 2 tags")
	expectedTags = map[string]string{
		"endpoint": "health",
		"handler":  "health",
	}
	for tag := range tags {
		assert.Equal(t,
			expectedTags[tag.GetTagName()],
			tag.GetTagValue(),
			"expected tag value to be correct",
		)
	}

	statusCodeMetric := metrics[2]

	assert.Equal(t,
		"test-gateway.production.all-workers.inbound.calls.status.200",
		statusCodeMetric.GetName(),
		"expected correct name",
	)

	value = *statusCodeMetric.MetricValue.Count.I64Value
	assert.Equal(t,
		int64(1),
		value,
		"expected counter to be 1",
	)

	tags = statusCodeMetric.GetTags()
	assert.Equal(t, 2, len(tags), "expected 2 tags")
	expectedTags = map[string]string{
		"endpoint": "health",
		"handler":  "health",
	}
	for tag := range tags {
		assert.Equal(t,
			expectedTags[tag.GetTagName()],
			tag.GetTagValue(),
			"expected tag value to be correct",
		)
	}
}
