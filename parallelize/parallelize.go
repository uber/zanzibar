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

package parallelize

import (
	"runtime"
	"sync"
)

// NewUnboundedRunner creates unbounded goroutines
func NewUnboundedRunner(workSize int) *Runner {
	r := getRunner(workSize)
	r.wg.Add(workSize)
	go r.consumeMessagesUnbounded()
	return r
}

func getRunner(workSize int) *Runner {
	r := &Runner{
		resultQueue: make(chan result, workSize),
		wg:          &sync.WaitGroup{},
		workQueue:   make(chan Work, workSize),
	}
	return r
}

func (r *Runner) consumeMessagesUnbounded() {
	func() {
		for {
			wrk, ok := <-r.workQueue
			if !ok {
				return
			}

			go func(wrk Work) {
				defer r.wg.Done()
				res, err := wrk.Work()
				r.resultQueue <- result{
					data: res,
					err:  err,
				}
			}(wrk)
		}
	}()
}

// NewFixedBoundedRunner creates fixed bounded goroutines by factor of available CPU based on default setting.
// Default setting is to have parallelization as factor of 4x if iobound work which is supposed to be short.
// If its a long work use NewBoundedRunner instead with optimal config.
// If cpu bound work parallel go routine will be bound by cpu.
func NewFixedBoundedRunner(workSize int, ioBound bool) *Runner {
	parallelCount := runtime.NumCPU()
	// go routine busy doing io would be swapped out, hence 4x.
	if ioBound {
		parallelCount = parallelCount * 4
	}
	return NewBoundedRunner(workSize, parallelCount)
}

// NewBoundedRunner creates bounded goroutines by factor parallel count
func NewBoundedRunner(workSize, parallelCount int) *Runner {
	r := &Runner{
		resultQueue: make(chan result, workSize),
		wg:          &sync.WaitGroup{},
		workQueue:   make(chan Work, workSize),
	}
	r.wg.Add(parallelCount)
	go r.consumeMessagesBounded(parallelCount)
	return r
}

func (r *Runner) consumeMessagesBounded(parallelCount int) {
	func() {
		for i := 0; i < parallelCount; i++ {
			go func() {
				defer r.wg.Done()
				for {
					wrk, ok := <-r.workQueue
					if !ok {
						return
					}

					res, err := wrk.Work()
					r.resultQueue <- result{
						data: res,
						err:  err,
					}
				}
			}()
		}

	}()
}

type result struct {
	data interface{}
	err  error
}

// Runner holds data for initiating parallel work
type Runner struct {
	resultQueue chan result
	wg          *sync.WaitGroup
	workQueue   chan Work
}

// SubmitWork submits a unit of work to be executed
func (r *Runner) SubmitWork(wrk Work) {
	r.workQueue <- wrk
}

// GetResult returns array of responses from executing Work and returns early on first error it gets.
// Also after calling this no more work can be submitted to the Runner
func (r *Runner) GetResult() ([]interface{}, error) {
	go func() {
		close(r.workQueue)
		r.wg.Wait()
		close(r.resultQueue)
	}()

	var results []interface{}
	for ele := range r.resultQueue {
		results = append(results, ele.data)
		if ele.err != nil {
			return nil, ele.err
		}
	}
	return results, nil
}

// Work is a unit of work set to be executed by this Runner
type Work interface {
	Work() (interface{}, error)
}

// StatelessFunc is defined to simulate anonymous implementation directly lacking in golang.
// It will avoid creating boilerplate implementation of Work interface
type StatelessFunc func() (interface{}, error)

// Work satisfies Work interface. So we can now pass an anonymous function casted to StatelessFunc
func (sf StatelessFunc) Work() (interface{}, error) {
	return sf()
}

// SingleParamWork is a utility for doing a single param work
type SingleParamWork struct {
	Data interface{}
	Func func(data interface{}) (interface{}, error)
}

// Work satisfies Work interface. So we can now pass an anonymous function casted to SingleParamWork
func (spw *SingleParamWork) Work() (interface{}, error) {
	return spw.Func(spw.Data)
}

// TwoParamWork is a utility for doing a two param work
type TwoParamWork struct {
	Data1 interface{}
	Data2 interface{}
	Func  func(data1 interface{}, data2 interface{}) (interface{}, error)
}

// Work satisfies Work interface. So we can now pass an anonymous function casted to TwoParamWork
func (tpw *TwoParamWork) Work() (interface{}, error) {
	return tpw.Func(tpw.Data1, tpw.Data2)
}

// ThreeParamWork is a utility for doing a three param work
type ThreeParamWork struct {
	Data1 interface{}
	Data2 interface{}
	Data3 interface{}
	Func  func(data1 interface{}, data2 interface{}, data3 interface{}) (interface{}, error)
}

// Work satisfies Work interface. So we can now pass an anonymous function casted to ThreeParamWork
func (tpw *ThreeParamWork) Work() (interface{}, error) {
	return tpw.Func(tpw.Data1, tpw.Data2, tpw.Data3)
}

// MultiParamWork is a utility for doing a multi param work
type MultiParamWork struct {
	Data []interface{}
	Func func(...interface{}) (interface{}, error)
}

// Work satisfies Work interface. So we can now pass an anonymous function casted to MultiParamWork
func (mpw *MultiParamWork) Work() (interface{}, error) {
	return mpw.Func(mpw.Data...)
}
