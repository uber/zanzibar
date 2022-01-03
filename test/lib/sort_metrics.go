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
	"github.com/uber-go/tally"
	m3 "github.com/uber-go/tally/m3/thrift/v2"
)

// SortMetricsByNameAndTags ...
type SortMetricsByNameAndTags []m3.Metric

// Len ...
func (a SortMetricsByNameAndTags) Len() int {
	return len(a)
}

// Swap ...
func (a SortMetricsByNameAndTags) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less ...
func (a SortMetricsByNameAndTags) Less(i, j int) bool {
	if a[i].Name == a[j].Name {
		ti := tally.KeyForStringMap(tagsStringMap(a[i]))
		tj := tally.KeyForStringMap(tagsStringMap(a[j]))
		return ti < tj

	}
	return a[i].Name < a[j].Name
}

func tagsStringMap(m m3.Metric) map[string]string {
	out := make(map[string]string, len(m.Tags))
	for _, tag := range m.Tags {
		out[tag.Name+":"+tag.Value] = ""
	}
	return out
}
