package zanzibar

import (
	"github.com/uber-go/tally"
	"time"
	"sync"
)

type Metrics interface {
	IncrementCounter()
	RecordLatency(time.Duration)
	IncrementMetric(string)
}

// Metrics contains pre allocated metrics structures
// These are pre-allocated to cache tags maps and for performance.
type metrics struct {
	root tally.Scope

	requestCounter   tally.Counter
	requestLatency tally.Timer
	requestMetrics    map[string]tally.Counter

	sync.RWMutex
}

func NewMetrics(
	root tally.Scope,
	counter, latency string,
) Metrics {
	return &metrics{
		root: root,
		requestCounter: root.Counter(counter),
		requestLatency: root.Timer(latency),
		requestMetrics: make(map[string]tally.Counter),
	}
}

func (m *metrics) IncrementCounter() {
	m.requestCounter.Inc(1)
}

func (m *metrics) RecordLatency(t time.Duration) {
	m.requestLatency.Record(t)
}

func (m *metrics) IncrementMetric(name string) {
	m.RLock()
	ctr, ok := m.requestMetrics[name]
	m.RUnlock()
	if !ok {
		m.Lock()
		ctr = m.root.Counter(name)
		m.requestMetrics[name] = ctr
		m.Unlock()
	}
	ctr.Inc(1)
}
