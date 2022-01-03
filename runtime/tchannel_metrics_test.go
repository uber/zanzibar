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

package zanzibar

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
)

var tags = map[string]string{
	"app":            "app",
	"host":           "hostname",
	"service":        "my-service",
	"targetservice":  "other-service",
	"targetendpoint": "operation",
}

var (
	prefix            = "prefix"
	counterName       = "outbound.counter"
	gaugeName         = "outbound.gauge"
	timerName         = "outbound.timer"
	prefixCounterName = prefix + "." + counterName
	prefixGaugeName   = prefix + "." + gaugeName
	prefixTimerName   = prefix + "." + timerName
)

func TestNewTChannelStatsReporter(t *testing.T) {
	testScope := tally.NewTestScope(prefix, nil)
	reporter := NewTChannelStatsReporter(testScope)

	reporter.IncCounter(counterName, tags, 1)
	reporter.IncCounter(counterName, tags, 41)
	for _, m := range knownMetrics {
		reporter.IncCounter(m, tags, 5)
	}

	reporter.UpdateGauge(gaugeName, tags, 1)
	reporter.UpdateGauge(gaugeName, tags, 42)
	reporter.UpdateGauge(gaugeName, tags, 13)
	for _, m := range knownMetrics {
		reporter.UpdateGauge(m, tags, 17)
	}

	reporter.RecordTimer(timerName, tags, 100)
	reporter.RecordTimer(timerName, tags, 200)
	reporter.RecordTimer(timerName, tags, 400)
	for _, m := range knownMetrics {
		reporter.RecordTimer(m, tags, 1000)
	}

	snapshot := testScope.Snapshot()

	snapshotCounterName := tally.KeyForPrefixedStringMap(prefixCounterName, tags)
	counterSnapshot, ok := snapshot.Counters()[snapshotCounterName]
	assert.True(t, ok)
	assert.Equal(t, prefixCounterName, counterSnapshot.Name())
	assert.Equal(t, int64(42), counterSnapshot.Value())

	snapshotGaugeName := tally.KeyForPrefixedStringMap(prefixGaugeName, tags)
	gaugeSnapshot, ok := snapshot.Gauges()[snapshotGaugeName]
	assert.True(t, ok)
	assert.Equal(t, prefixGaugeName, gaugeSnapshot.Name())
	assert.Equal(t, float64(13), gaugeSnapshot.Value())

	snapshotTimerName := tally.KeyForPrefixedStringMap(prefixTimerName, tags)
	timerSnapshot, ok := snapshot.Timers()[snapshotTimerName]
	assert.True(t, ok)
	assert.Equal(t, prefixTimerName, timerSnapshot.Name())
	assert.Equal(t, []time.Duration{100, 200, 400}, timerSnapshot.Values())

	for _, m := range knownMetrics {
		name := tally.KeyForPrefixedStringMap(prefix+"."+m, tags)
		assert.Nil(t, snapshot.Counters()[name])
		assert.Nil(t, snapshot.Gauges()[name])
		assert.Nil(t, snapshot.Timers()[name])
	}
}
