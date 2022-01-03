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
	"time"

	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go"
)

type statsReporter struct {
	scope        tally.Scope
	knownMetrics map[string]bool
}

// NewTChannelStatsReporter returns a StatsReporter using the given tally.Scope.
func NewTChannelStatsReporter(scope tally.Scope) tchannel.StatsReporter {
	s := &statsReporter{
		scope:        scope,
		knownMetrics: make(map[string]bool, len(knownMetrics)),
	}
	for _, m := range knownMetrics {
		s.knownMetrics[m] = true
	}
	return s
}

// IncCounter ...
func (s statsReporter) IncCounter(name string, tags map[string]string, value int64) {
	if !s.knownMetrics[name] {
		s.scope.Tagged(tags).Counter(name).Inc(value)
	}
}

// UpdateGauge ...
func (s statsReporter) UpdateGauge(name string, tags map[string]string, value int64) {
	if !s.knownMetrics[name] {
		s.scope.Tagged(tags).Gauge(name).Update(float64(value))
	}
}

// RecordTimer ...
func (s statsReporter) RecordTimer(name string, tags map[string]string, d time.Duration) {
	if !s.knownMetrics[name] {
		s.scope.Tagged(tags).Timer(name).Record(d)
	}
}
