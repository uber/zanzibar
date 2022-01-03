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
	"runtime"
	"sync"
	"time"

	"github.com/uber-go/tally"
)

const _numGCThreshold = uint32(len(runtime.MemStats{}.PauseEnd))

// RuntimeMetricsOptions configuration.
type RuntimeMetricsOptions struct {
	EnableCPUMetrics bool          `json:"enableCPUMetrics"`
	EnableMemMetrics bool          `json:"enableMemMetrics"`
	EnableGCMetrics  bool          `json:"enableGCMetrics"`
	CollectInterval  time.Duration `json:"collectInterval"`
}

// RuntimeMetricsCollector interface.
type RuntimeMetricsCollector interface {
	Start()
	Stop()
	IsRunning() bool
}

type runtimeMetrics struct {
	// maximum number of CPUs which are executing simultaneously
	goMaxProcs tally.Gauge
	// number of logical CPUs usable by the current process
	numCPUs tally.Gauge
	// number of goroutines that currently exist
	numGoRoutines tally.Gauge

	// bytes of allocated heap objects
	heapAlloc tally.Gauge
	// bytes in idle (unused) spans
	heapIdle tally.Gauge
	// bytes in in-use spans
	heapInuse tally.Gauge
	// bytes in stack spans
	stackInuse tally.Gauge

	// number of completed GC cycles
	numGC tally.Counter
	// GC pause time
	gcPauseMs     tally.Timer
	gcPauseMsHist tally.Histogram
}

// runtimeCollector keeps the current state of runtime metrics
type runtimeCollector struct {
	opts         RuntimeMetricsOptions
	scope        tally.Scope
	metrics      runtimeMetrics
	runningMutex sync.RWMutex
	running      bool // protected by runningMutex
	stop         chan struct{}
	lastNumGC    uint32
}

// StartRuntimeMetricsCollector starts collecting runtime metrics periodically.
// Recommended usage:
//     rm := StartRuntimeMetricsCollector(rootScope.Scope("runtime"), opts)
//     ...
//     rm.Stop()
func StartRuntimeMetricsCollector(
	config RuntimeMetricsOptions,
	scope tally.Scope,
) RuntimeMetricsCollector {
	if !config.EnableCPUMetrics && !config.EnableMemMetrics && !config.EnableGCMetrics {
		return nil
	}
	rm := NewRuntimeMetricsCollector(
		config, scope.SubScope("runtime"),
	)
	rm.Start()
	return rm
}

// NewRuntimeMetricsCollector creates a new runtime metrics collector.
func NewRuntimeMetricsCollector(
	opts RuntimeMetricsOptions,
	scope tally.Scope,
) RuntimeMetricsCollector {
	var memstats runtime.MemStats
	runtime.ReadMemStats(&memstats)

	return &runtimeCollector{
		opts:  opts,
		scope: scope,
		metrics: runtimeMetrics{
			// CPU
			goMaxProcs:    scope.Gauge("gomaxprocs"),
			numCPUs:       scope.Gauge("num-cpu"),
			numGoRoutines: scope.Gauge("num-goroutines"),

			// Memory
			heapAlloc:  scope.Gauge("memory.heap"),
			heapIdle:   scope.Gauge("memory.heapidle"),
			heapInuse:  scope.Gauge("memory.heapinuse"),
			stackInuse: scope.Gauge("memory.stack"),

			// GC
			numGC:         scope.Counter("memory.num-gc"),
			gcPauseMs:     scope.Timer("memory.gc-pause-ms"),
			gcPauseMsHist: scope.Histogram("memory.gc-pause-ms-hist", tally.DefaultBuckets),
		},
		running:   false,
		stop:      make(chan struct{}),
		lastNumGC: memstats.NumGC,
	}
}

// Start collecting runtime metrics periodically.
func (r *runtimeCollector) Start() {
	r.runningMutex.RLock()
	if r.running {
		r.runningMutex.RUnlock()
		return
	}
	r.runningMutex.RUnlock()
	if r.opts.EnableCPUMetrics || r.opts.EnableMemMetrics || r.opts.EnableGCMetrics {
		go func() {
			ticker := time.NewTicker(r.opts.CollectInterval)
			for {
				select {
				case <-ticker.C:
					r.collect()
				case <-r.stop:
					ticker.Stop()
					return
				}
			}
		}()
		r.runningMutex.Lock()
		r.running = true
		r.runningMutex.Unlock()
	}
}

// Stop collecting runtime metrics. It cannot be restarted once stopped.
func (r *runtimeCollector) Stop() {
	r.runningMutex.Lock()
	defer r.runningMutex.Unlock()
	close(r.stop)
	r.running = false
}

// IsRunning returns true if the runtime metrics collector was running; otherwise false.
func (r *runtimeCollector) IsRunning() bool {
	r.runningMutex.RLock()
	defer r.runningMutex.RUnlock()
	return r.running
}

func (r *runtimeCollector) collect() {
	var memStats runtime.MemStats
	if r.opts.EnableMemMetrics || r.opts.EnableGCMetrics {
		runtime.ReadMemStats(&memStats)
	}

	if r.opts.EnableCPUMetrics {
		r.collectCPUMetrics()
	}
	if r.opts.EnableMemMetrics {
		r.collectMemMetrics(&memStats)
	}
	if r.opts.EnableGCMetrics {
		r.collectGCMetrics(&memStats)
	}
}

func (r *runtimeCollector) collectCPUMetrics() {
	r.metrics.goMaxProcs.Update(float64(runtime.GOMAXPROCS(0)))
	r.metrics.numCPUs.Update(float64(runtime.NumCPU()))
	r.metrics.numGoRoutines.Update(float64(runtime.NumGoroutine()))
}

func (r *runtimeCollector) collectMemMetrics(memStats *runtime.MemStats) {
	r.metrics.heapAlloc.Update(float64(memStats.HeapAlloc))
	r.metrics.heapIdle.Update(float64(memStats.HeapIdle))
	r.metrics.heapInuse.Update(float64(memStats.HeapInuse))
	r.metrics.stackInuse.Update(float64(memStats.StackInuse))
}

func (r *runtimeCollector) collectGCMetrics(memStats *runtime.MemStats) {
	num := memStats.NumGC
	lastNum := r.lastNumGC
	r.lastNumGC = num

	if delta := num - lastNum; delta > 0 {
		r.metrics.numGC.Inc(int64(delta))
		if delta >= _numGCThreshold {
			/* coverage ignore next line */
			lastNum = num - _numGCThreshold
		}

		for i := lastNum; i != num; i++ {
			pause := memStats.PauseNs[i%uint32(len(memStats.PauseNs))]
			r.metrics.gcPauseMs.Record(time.Duration(pause))
			r.metrics.gcPauseMsHist.RecordDuration(time.Duration(pause))
		}
	}

}
