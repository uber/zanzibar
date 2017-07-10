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
	"time"

	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/stats"
)

type statsReporter struct {
	scope     tally.Scope
	metricKey MetricKeyFunc
}

// MetricKeyFunc converts a tchannel metric name and a list of tags to a
// key that is used when reporting to M3.
type MetricKeyFunc func(name string, tags map[string]string) string

// NewDefaultTChannelStatsReporter returns a StatsReporter that reports stats
// using the given tally.Scope. It only reports a metric key that is generated
// from the given list of tags.
func NewDefaultTChannelStatsReporter(s tally.Scope) tchannel.StatsReporter {
	return &statsReporter{
		scope: s,
		metricKey: func(name string, tags map[string]string) string {
			return stats.MetricWithPrefix("", name, tags)
		},
	}
}

// NewTChannelStatsReporter returns a StatsReporter that reports stats
// using the given tally.Scope, using the given metric key function.
func NewTChannelStatsReporter(s tally.Scope, m MetricKeyFunc) tchannel.StatsReporter {
	return &statsReporter{s, m}
}

// IncCounter ...
func (s statsReporter) IncCounter(name string, tags map[string]string, value int64) {
	s.scope.Counter(s.metricKey(name, tags)).Inc(value)
}

// UpdateGauge ...
func (s statsReporter) UpdateGauge(name string, tags map[string]string, value int64) {
	s.scope.Gauge(s.metricKey(name, tags)).Update(float64(value))
}

// RecordTimer ...
func (s statsReporter) RecordTimer(name string, tags map[string]string, d time.Duration) {
	s.scope.Timer(s.metricKey(name, tags)).Record(d)
}
