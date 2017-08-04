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
	"runtime"
	"sync"
	"time"

	"github.com/uber-go/tally"
)

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
	// CPU
	goMaxProcs    tally.Gauge // maximum number of CPUs which are executing simultaneously
	numCPUs       tally.Gauge // number of logical CPUs usable by the current process
	numCgoCalls   tally.Gauge // number of cgo calls made by the current process
	numGoRoutines tally.Gauge // number of goroutines that currently exist

	// General
	alloc      tally.Gauge // bytes of allocated heap objects
	totalAlloc tally.Gauge // cumulative bytes allocated for heap objects
	sys        tally.Gauge // total bytes of memory obtained from the OS
	lookups    tally.Gauge // number of pointer lookups performed
	mallocs    tally.Gauge // cumulative count of heap objects allocated
	frees      tally.Gauge // cumulative count of heap objects freed

	// Heap
	heapAlloc    tally.Gauge // bytes of allocated heap objects
	heapSys      tally.Gauge // bytes of heap memory obtained from the OS
	heapIdle     tally.Gauge // bytes in idle (unused) spans
	heapInuse    tally.Gauge // bytes in in-use spans
	heapReleased tally.Gauge // bytes of physical memory returned to the OS
	heapObjects  tally.Gauge // number of allocated heap objects

	// Stack
	stackInuse  tally.Gauge // bytes in stack spans
	stackSys    tally.Gauge // bytes of stack memory obtained from the OS
	mspanInuse  tally.Gauge // bytes of allocated mspan structures
	mspanSys    tally.Gauge // bytes of memory obtained from the OS for mspan
	mcacheInuse tally.Gauge // bytes of allocated mcache structures
	mcacheSys   tally.Gauge // bytes of memory obtained from the OS for mcache structures

	otherSys tally.Gauge // bytes of memory in miscellaneous off-heap runtime allocations

	// GC
	gcSys         tally.Gauge // bytes of memory in garbage collection metadata
	nextGC        tally.Gauge // target heap size of the next GC cycle
	lastGC        tally.Gauge // time the last garbage collection finished, as nanoseconds since epoch
	pauseTotalMs  tally.Gauge // cumulative nanoseconds in GC stop-the-world pauses since the program running
	pauseMs       tally.Gauge // most recent pause timing
	numGC         tally.Gauge // number of completed GC cycles
	gcCPUFraction tally.Gauge // fraction of cpu consumed by GC
}

// runtimeCollector keeps the current state of runtime metrics
type runtimeCollector struct {
	opts         RuntimeMetricsOptions
	scope        tally.Scope
	metrics      runtimeMetrics
	runningMutex sync.RWMutex
	running      bool // protected by runningMutex
	stop         chan struct{}
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
	rm := NewRuntimeMetricsCollector(config, scope.SubScope("runtime"))
	rm.Start()
	return rm
}

// NewRuntimeMetricsCollector creates a new runtime metrics collector.
func NewRuntimeMetricsCollector(
	opts RuntimeMetricsOptions,
	scope tally.Scope,
) RuntimeMetricsCollector {
	return &runtimeCollector{
		opts:  opts,
		scope: scope,
		metrics: runtimeMetrics{
			// CPU
			goMaxProcs:    scope.Gauge("cpu.goMaxProcs"),
			numCPUs:       scope.Gauge("cpu.count"),
			numCgoCalls:   scope.Gauge("cpu.cgoCalls"),
			numGoRoutines: scope.Gauge("cpu.goroutines"),

			// General memory
			alloc:      scope.Gauge("mem.alloc"),
			totalAlloc: scope.Gauge("mem.total"),
			sys:        scope.Gauge("mem.sys"),
			lookups:    scope.Gauge("mem.lookups"),
			mallocs:    scope.Gauge("mem.malloc"),
			frees:      scope.Gauge("mem.frees"),

			// Heap memory
			heapAlloc:    scope.Gauge("mem.heap.alloc"),
			heapSys:      scope.Gauge("mem.heap.sys"),
			heapIdle:     scope.Gauge("mem.heap.idle"),
			heapInuse:    scope.Gauge("mem.heap.inuse"),
			heapReleased: scope.Gauge("mem.heap.released"),
			heapObjects:  scope.Gauge("mem.heap.objects"),

			// Stack memory
			stackInuse:  scope.Gauge("mem.stack.inuse"),
			stackSys:    scope.Gauge("mem.stack.sys"),
			mspanInuse:  scope.Gauge("mem.stack.mspanInuse"),
			mspanSys:    scope.Gauge("mem.stack.mspanSys"),
			mcacheInuse: scope.Gauge("mem.stack.mcacheInuse"),
			mcacheSys:   scope.Gauge("mem.stack.mcacheSys"),

			// Other memory
			otherSys: scope.Gauge("mem.otherSys"),

			// GC
			gcSys:         scope.Gauge("mem.gc.sys"),
			nextGC:        scope.Gauge("mem.gc.next"),
			lastGC:        scope.Gauge("mem.gc.last"),
			pauseTotalMs:  scope.Gauge("mem.gc.pauseTotal"),
			pauseMs:       scope.Gauge("mem.gc.pause"),
			numGC:         scope.Gauge("mem.gc.count"),
			gcCPUFraction: scope.Gauge("mem.gc.cpuFraction"),
		},
		running: false,
		stop:    make(chan struct{}),
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
	r.metrics.numCgoCalls.Update(float64(runtime.NumCgoCall()))
	r.metrics.numGoRoutines.Update(float64(runtime.NumGoroutine()))
}

func (r *runtimeCollector) collectMemMetrics(memStats *runtime.MemStats) {
	r.metrics.alloc.Update(float64(memStats.Alloc))
	r.metrics.totalAlloc.Update(float64(memStats.TotalAlloc))
	r.metrics.sys.Update(float64(memStats.Sys))
	r.metrics.lookups.Update(float64(memStats.Lookups))
	r.metrics.mallocs.Update(float64(memStats.Mallocs))
	r.metrics.frees.Update(float64(memStats.Frees))

	r.metrics.heapAlloc.Update(float64(memStats.HeapAlloc))
	r.metrics.heapSys.Update(float64(memStats.HeapSys))
	r.metrics.heapIdle.Update(float64(memStats.HeapIdle))
	r.metrics.heapInuse.Update(float64(memStats.HeapInuse))
	r.metrics.heapReleased.Update(float64(memStats.HeapReleased))
	r.metrics.heapObjects.Update(float64(memStats.HeapObjects))

	r.metrics.stackInuse.Update(float64(memStats.StackInuse))
	r.metrics.stackSys.Update(float64(memStats.StackSys))
	r.metrics.mspanInuse.Update(float64(memStats.MSpanInuse))
	r.metrics.mspanSys.Update(float64(memStats.MSpanSys))
	r.metrics.mcacheInuse.Update(float64(memStats.MCacheInuse))
	r.metrics.mcacheSys.Update(float64(memStats.MCacheSys))

	r.metrics.otherSys.Update(float64(memStats.OtherSys))
}

func (r *runtimeCollector) collectGCMetrics(memStats *runtime.MemStats) {
	r.metrics.gcSys.Update(float64(memStats.GCSys))
	r.metrics.nextGC.Update(float64(memStats.NextGC))
	r.metrics.lastGC.Update(float64(memStats.LastGC))
	r.metrics.pauseTotalMs.Update(float64(memStats.PauseTotalNs / 1000))
	r.metrics.pauseMs.Update(float64((memStats.PauseNs[(memStats.NumGC+255)%256]) / 1000))
	r.metrics.numGC.Update(float64(memStats.NumGC))
	r.metrics.gcCPUFraction.Update(memStats.GCCPUFraction)
}
