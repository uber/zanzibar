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

package lib

import (
	"sync"
)

// WaitAtLeast wait group signals the wait channel when at least the added
// count are done. This avoids sync.WaitGroup panic when count is negative.
//
// Usage:
//     wg := WaitAtLeast{}
//     wg.Add(1)
//     wg.Wait()
//     go func() {
//         wg.Done()
//     }()
//
type WaitAtLeast struct {
	mutex sync.Mutex
	wg    sync.WaitGroup
	add   int
	done  int
}

// Add positive wait count.
func (w *WaitAtLeast) Add(count int) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if count <= 0 {
		panic("non-positive count")
	}

	w.add += count
	w.wg.Add(count)

	// complete matching add/done
	done := w.done
	if w.add < w.done {
		done = w.add
	}
	for i := 0; i < done; i++ {
		w.done--
		w.add--
		w.wg.Done()
	}
}

// Done subtracts one from the wait count.
func (w *WaitAtLeast) Done() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// complete matching add/done
	if w.add > 0 {
		w.add--
		w.wg.Done()
		return
	}

	w.done++
}

// Wait for matching add/done.
func (w *WaitAtLeast) Wait() {
	w.wg.Wait()
}
