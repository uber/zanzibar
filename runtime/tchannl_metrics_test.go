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

package zanzibar

import (
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"time"
)

var tags = map[string]string{
	"app":             "app",
	"host":            "hostname",
	"service":         "my-service",
	"target-service":  "other-service",
	"target-endpoint": "operation",
}

func TestNewDefaultTChannelStatsReporterIncCounter(t *testing.T) {
	testScope := tally.NewTestScope("", nil)
	reporter := NewDefaultTChannelStatsReporter(testScope)

	counterName := "outbound.success"
	reporter.IncCounter(counterName, tags, 1)
	reporter.IncCounter(counterName, tags, 41)

	expectedKey := fmt.Sprintf(
		"%s.%s.%s.%s",
		counterName,
		tags["service"],
		tags["target-service"],
		tags["target-endpoint"],
	)

	snapshot := testScope.Snapshot()
	counterSnapshot, ok := snapshot.Counters()[expectedKey]
	assert.True(t, ok)
	assert.Equal(t, expectedKey, counterSnapshot.Name())
	assert.Equal(t, int64(42), counterSnapshot.Value())
}

func TestNewDefaultTChannelStatsReporterUpdateGauge(t *testing.T) {
	testScope := tally.NewTestScope("", nil)
	reporter := NewDefaultTChannelStatsReporter(testScope)

	gaugeName := "outbound.pending"
	reporter.UpdateGauge(gaugeName, tags, 1)
	reporter.UpdateGauge(gaugeName, tags, 42)
	reporter.UpdateGauge(gaugeName, tags, 13)

	expectedKey := fmt.Sprintf(
		"%s.%s.%s.%s",
		gaugeName,
		tags["service"],
		tags["target-service"],
		tags["target-endpoint"],
	)

	snapshot := testScope.Snapshot()
	gaugeSnapshot, ok := snapshot.Gauges()[expectedKey]
	assert.True(t, ok)
	assert.Equal(t, expectedKey, gaugeSnapshot.Name())
	assert.Equal(t, float64(13), gaugeSnapshot.Value())
}

func TestNewDefaultTChannelStatsReporterRecordTimer(t *testing.T) {
	testScope := tally.NewTestScope("", nil)
	reporter := NewDefaultTChannelStatsReporter(testScope)

	timerName := "outbound.pending"
	reporter.RecordTimer(timerName, tags, 100)
	reporter.RecordTimer(timerName, tags, 200)
	reporter.RecordTimer(timerName, tags, 400)

	expectedKey := fmt.Sprintf(
		"%s.%s.%s.%s",
		timerName,
		tags["service"],
		tags["target-service"],
		tags["target-endpoint"],
	)

	snapshot := testScope.Snapshot()
	timerSnapshot, ok := snapshot.Timers()[expectedKey]
	assert.True(t, ok)
	assert.Equal(t, expectedKey, timerSnapshot.Name())
	assert.Equal(t, []time.Duration{100, 200, 400}, timerSnapshot.Values())
}

func TestNewTChannelStatsReporter(t *testing.T) {
	testScope := tally.NewTestScope("", nil)
	reporter := NewTChannelStatsReporter(
		testScope,
		func(name string, _ map[string]string) string {
			return "custom." + name
		},
	)

	counterName := "outbound.success"
	reporter.IncCounter(counterName, tags, 1)
	reporter.IncCounter(counterName, tags, 41)

	expectedKey := fmt.Sprintf("custom.%s", counterName)

	snapshot := testScope.Snapshot()
	fmt.Println(snapshot.Counters())
	counterSnapshot, ok := snapshot.Counters()[expectedKey]
	assert.True(t, ok)
	assert.Equal(t, expectedKey, counterSnapshot.Name())
	assert.Equal(t, int64(42), counterSnapshot.Value())
}
