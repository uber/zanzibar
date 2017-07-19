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
	"sync"

	"github.com/uber-go/tally"
	jaeger_metrics "github.com/uber/jaeger-lib/metrics"
)

type jaegerGauge struct {
	gauge tally.Gauge
}

func (j *jaegerGauge) Update(value int64) {
	j.gauge.Update(float64(value))
}

type jaegerMetricsFactory struct {
	scope  tally.Scope
	gauges map[string]*jaegerGauge
	gm     sync.RWMutex
}

// NewJaegerMetricsFactory returns a new jaeger_metrics.Factory backed by a tally metrics scope.
func NewJaegerMetricsFactory(scope tally.Scope) jaeger_metrics.Factory {
	return &jaegerMetricsFactory{
		scope:  scope,
		gauges: make(map[string]*jaegerGauge),
	}
}

// Counter returns an jaeger stats counter.
func (f *jaegerMetricsFactory) Counter(name string, tags map[string]string) jaeger_metrics.Counter {
	return f.scope.Tagged(tags).Counter(name)
}

// Gauge returns an jaeger stats gauge.
func (f *jaegerMetricsFactory) Gauge(name string, tags map[string]string) jaeger_metrics.Gauge {
	key := tally.KeyForPrefixedStringMap(name, tags)
	f.gm.RLock()
	g, ok := f.gauges[key]
	f.gm.RUnlock()
	if !ok {
		f.gm.Lock()
		g = &jaegerGauge{f.scope.Tagged(tags).Gauge(name)}
		f.gauges[key] = g
		f.gm.Unlock()
	}
	return g
}

// Timer returns an jaeger stats timer.
func (f *jaegerMetricsFactory) Timer(name string, tags map[string]string) jaeger_metrics.Timer {
	return f.scope.Tagged(tags).Timer(name)
}

// Namespace returns an jaeger_metrics.Factory backed with a tally sub scope.
func (f *jaegerMetricsFactory) Namespace(name string, tags map[string]string) jaeger_metrics.Factory {
	newScope := f.scope
	if name != "" {
		newScope = newScope.SubScope(name)
	}
	if len(tags) > 0 {
		newScope = newScope.Tagged(tags)
	}
	return NewJaegerMetricsFactory(newScope)
}
