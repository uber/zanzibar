// Copyright (c) 2019 Uber Technologies, Inc.
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

// Package plugins allows users to operate on statistics recorded for each circuit operation.
// Plugins should be careful to be lightweight as they will be called frequently.
package plugins

import (
	"github.com/uber-go/tally"
	"strings"
	"time"

	"github.com/afex/hystrix-go/hystrix/metric_collector"
)

// M3Collector fulfills the metricCollector interface allowing users to ship circuit
// stats to a m3 backend. To use users must call InitializeM3Collector before
// circuits are started. Then register NewM3Collector with metricCollector.Registry.Register(NewM3Collector).
type M3Collector struct {
	scope                   tally.Scope
	attemptsPrefix          string
	errorsPrefix            string
	successesPrefix         string
	failuresPrefix          string
	rejectsPrefix           string
	shortCircuitsPrefix     string
	timeoutsPrefix          string
	fallbackSuccessesPrefix string
	fallbackFailuresPrefix  string
	totalDurationPrefix     string
	runDurationPrefix       string
}

// M3CollectorClient provides configuration that the m3 client will need.
type M3CollectorClient struct {
	scope tally.Scope
}

// InitializeM3Collector creates the connection to the m3
func InitializeM3Collector(scope tally.Scope) *M3CollectorClient {
	return &M3CollectorClient{
		scope: scope,
	}

}

// NewM3Collector creates a collector for a specific circuit. The
// prefix given to this circuit will be {config.Prefix}.{circuit_name}.{metric}.
// Circuits with "/" in their names will have them replaced with ".".
func (m *M3CollectorClient) NewM3Collector(name string) metricCollector.MetricCollector {
	name = strings.Replace(name, "/", "-", -1)
	name = strings.Replace(name, ":", "-", -1)
	name = strings.Replace(name, ".", "-", -1)
	return &M3Collector{
		scope:                   m.scope,
		attemptsPrefix:          name + ".attempts",
		errorsPrefix:            name + ".errors",
		successesPrefix:         name + ".successes",
		failuresPrefix:          name + ".failures",
		rejectsPrefix:           name + ".rejects",
		shortCircuitsPrefix:     name + ".shortCircuits",
		timeoutsPrefix:          name + ".timeouts",
		fallbackSuccessesPrefix: name + ".fallbackSuccesses",
		fallbackFailuresPrefix:  name + ".fallbackFailures",
		totalDurationPrefix:     name + ".totalDuration",
		runDurationPrefix:       name + ".runDuration",
	}
}

func (g *M3Collector) incrementCounterMetric(prefix string, i float64) {
	if i == 0 {
		return
	}
	c := g.scope.Counter(prefix)
	c.Inc(int64(i))
}

func (g *M3Collector) updateTimerMetric(prefix string, dur time.Duration) {
	c := g.scope.Timer(prefix)

	c.Record(dur)
}

// Update is a callback by hystrix lib to relay the metrics to m3
func (g *M3Collector) Update(r metricCollector.MetricResult) {
	g.incrementCounterMetric(g.attemptsPrefix, r.Attempts)
	g.incrementCounterMetric(g.errorsPrefix, r.Errors)
	g.incrementCounterMetric(g.successesPrefix, r.Successes)
	g.incrementCounterMetric(g.failuresPrefix, r.Failures)
	g.incrementCounterMetric(g.rejectsPrefix, r.Rejects)
	g.incrementCounterMetric(g.shortCircuitsPrefix, r.ShortCircuits)
	g.incrementCounterMetric(g.timeoutsPrefix, r.Timeouts)
	g.incrementCounterMetric(g.fallbackSuccessesPrefix, r.FallbackSuccesses)
	g.incrementCounterMetric(g.fallbackFailuresPrefix, r.FallbackFailures)
	g.updateTimerMetric(g.totalDurationPrefix, r.TotalDuration)
	g.updateTimerMetric(g.runDurationPrefix, r.RunDuration)
}

// Reset is a noop operation in this collector.
func (g *M3Collector) Reset() {}
