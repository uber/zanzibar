// Plugins allows users to operate on statistics recorded for each circuit operation.
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
