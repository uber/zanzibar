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

package lib

import (
	"sync"
)

// WaitAtLeast signals the wait channel when at least the added count are done.
// This avoids sync.WaitGroup panic when count is negative.
//
// Usage:
//     w := WaitAtLeast{ Wait: make(chan bool) }
//     w.Add(1)
//     <-w.Wait
//     go func() { w.Done() }()
//
type WaitAtLeast struct {
	mutex sync.Mutex
	count int
	Wait  chan bool
}

// Add positive wait count.
func (w *WaitAtLeast) Add(count int) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if count <= 0 {
		panic("non-positive count")
	}
	w.count += count
	if w.count <= 0 {
		w.Wait <- true
	}
}

// Done subtracts one from the wait count.
func (w *WaitAtLeast) Done() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.count--
	if w.count == 0 {
		w.Wait <- true
	}
}
