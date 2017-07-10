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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
)

func TestNewJaegerMetricsFactory(t *testing.T) {
	testScope := tally.NewTestScope("", nil)
	f := NewJaegerMetricsFactory(testScope)

	f.Counter("counter", nil).Inc(1)
	f.Counter("counter", nil).Inc(2)
	f.Gauge("gauge", nil).Update(13)
	f.Gauge("gauge", nil).Update(42)
	f.Timer("timer", nil).Record(100)
	f.Timer("timer", nil).Record(200)

	f2 := f.Namespace("foo", map[string]string{
		"hello": "hola",
	})
	f2.Counter("counter", nil).Inc(3)
	f2.Counter("counter", nil).Inc(6)
	f2.Gauge("gauge", nil).Update(42)
	f2.Gauge("gauge", nil).Update(13)
	f2.Timer("timer", nil).Record(300)
	f2.Timer("timer", nil).Record(400)

	snapshot := testScope.Snapshot()

	key := "counter"
	counterSnapshot, ok := snapshot.Counters()[key]
	assert.True(t, ok)
	assert.Equal(t, key, counterSnapshot.Name())
	assert.Equal(t, int64(3), counterSnapshot.Value())

	key = "gauge"
	gaugeSnapshot, ok := snapshot.Gauges()[key]
	assert.True(t, ok)
	assert.Equal(t, key, gaugeSnapshot.Name())
	assert.Equal(t, float64(42), gaugeSnapshot.Value())

	key = "timer"
	timerSnapshot, ok := snapshot.Timers()[key]
	assert.True(t, ok)
	assert.Equal(t, key, timerSnapshot.Name())
	assert.Equal(t, []time.Duration{100, 200}, timerSnapshot.Values())

	key = "foo.counter"
	f2CounterSnapshot, ok := snapshot.Counters()[key]
	assert.True(t, ok)
	assert.Equal(t, key, f2CounterSnapshot.Name())
	assert.Equal(t, int64(9), f2CounterSnapshot.Value())

	key = "foo.gauge"
	f2GaugeSnapshot, ok := snapshot.Gauges()[key]
	assert.True(t, ok)
	assert.Equal(t, key, f2GaugeSnapshot.Name())
	assert.Equal(t, float64(13), f2GaugeSnapshot.Value())

	key = "foo.timer"
	f2TimerSnapshot, ok := snapshot.Timers()[key]
	assert.True(t, ok)
	assert.Equal(t, key, f2TimerSnapshot.Name())
	assert.Equal(t, []time.Duration{300, 400}, f2TimerSnapshot.Values())
}
