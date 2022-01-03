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
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoundedRunner(t *testing.T) {
	var tests = []struct {
		name          string
		workItems     []*sampleWork
		parallelCount int
		expected      []string
		isErr         bool
	}{
		{
			name:          "parallelized - bounded",
			workItems:     []*sampleWork{{data: "abc"}, {data: "def"}},
			parallelCount: 2,
			expected:      []string{"abc", "def"},
		},
		{
			name:          "serial - bounded",
			workItems:     []*sampleWork{{data: "abc"}, {data: "def"}},
			parallelCount: 1,
			expected:      []string{"abc", "def"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewBoundedRunner(len(tt.workItems), tt.parallelCount)
			for _, wi := range tt.workItems {
				r.SubmitWork(wi)
			}
			compareResult(r, t, tt.isErr, tt.expected)
		})
	}
}

func TestIOBoundedRunner(t *testing.T) {
	var tests = []struct {
		name      string
		workItems []*sampleWork
		ioBound   bool
		isErr     bool
		expected  []string
	}{
		{
			name:      "parallelized - io",
			workItems: []*sampleWork{{data: "abc"}, {data: "def"}},
			ioBound:   true,
			expected:  []string{"abc", "def"},
		},
		{
			name:      "parallelized - non io work",
			workItems: []*sampleWork{{data: "abc"}, {data: "def"}},
			ioBound:   false,
			expected:  []string{"abc", "def"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewFixedBoundedRunner(len(tt.workItems), tt.ioBound)
			for _, wi := range tt.workItems {
				r.SubmitWork(wi)
			}
			compareResult(r, t, tt.isErr, tt.expected)
		})
	}
}

func TestUnboundedRunner(t *testing.T) {
	var tests = []struct {
		name      string
		workItems []*sampleWork
		ioBound   bool
		isErr     bool
		expected  []string
	}{
		{
			name:      "parallelized - unbounded",
			workItems: []*sampleWork{{data: "abc"}, {data: "def"}},
			ioBound:   true,
			expected:  []string{"abc", "def"},
		},
		{
			name:      "parallelized - unbounded error",
			workItems: []*sampleWork{{data: "abc"}, {data: "def", err: errors.New("error")}},
			ioBound:   false,
			expected:  nil,
			isErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewUnboundedRunner(len(tt.workItems))
			for _, wi := range tt.workItems {
				r.SubmitWork(wi)
			}
			compareResult(r, t, tt.isErr, tt.expected)
		})
	}
}

func TestStatelessWork(t *testing.T) {
	var tests = []struct {
		name      string
		workItems []Work
		ioBound   bool
		isErr     bool
		expected  []string
	}{
		{
			name: "stateless work",
			workItems: []Work{StatelessFunc(func() (interface{}, error) { return "abc", nil }),
				StatelessFunc(func() (interface{}, error) { return "def", nil })},
			ioBound:  true,
			expected: []string{"abc", "def"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewUnboundedRunner(len(tt.workItems))
			for _, wi := range tt.workItems {
				r.SubmitWork(wi)
			}
			compareResult(r, t, tt.isErr, tt.expected)
		})
	}
}

func TestSingleWork(t *testing.T) {
	var tests = []struct {
		name      string
		workItems []*SingleParamWork
		ioBound   bool
		isErr     bool
		expected  []string
	}{
		{
			name: "single param work",
			workItems: []*SingleParamWork{{
				Data: "abc",
				Func: func(data1 interface{}) (interface{}, error) {
					return data1, nil
				}}},
			ioBound:  true,
			expected: []string{"abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewUnboundedRunner(len(tt.workItems))
			for _, wi := range tt.workItems {
				r.SubmitWork(wi)
			}
			compareResult(r, t, tt.isErr, tt.expected)
		})
	}
}

func TestMultiParamWork(t *testing.T) {
	var tests = []struct {
		name      string
		workItems []*MultiParamWork
		ioBound   bool
		isErr     bool
		expected  []string
	}{
		{
			name: "single param work",
			workItems: []*MultiParamWork{{
				Data: []interface{}{"abc", "def"},
				Func: func(data ...interface{}) (interface{}, error) {
					return data[0], nil
				}}},
			ioBound:  true,
			expected: []string{"abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewUnboundedRunner(len(tt.workItems))
			for _, wi := range tt.workItems {
				r.SubmitWork(wi)
			}
			compareResult(r, t, tt.isErr, tt.expected)
		})
	}
}

func compareResult(r *Runner, t *testing.T, isErr bool, expected []string) {
	actual, err := r.GetResult()
	var actualStrResults []string
	for _, a := range actual {
		actualStrResults = append(actualStrResults, a.(string))
	}
	sort.Strings(actualStrResults)
	if !isErr {
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
	}
	assert.Equal(t, expected, actualStrResults)
}

type sampleWork struct {
	data string
	err  error
}

func (sw *sampleWork) Work() (interface{}, error) {
	if sw.err != nil {
		return nil, sw.err
	}
	return sw.data, nil
}
