// Copyright (c) 2020 Uber Technologies, Inc.
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

package jsonwrapper

import (
	"encoding/json"
)

// DefaultJSONWrapper implements the JSONWrapper interface with the golang encoding/json lib
type DefaultJSONWrapper struct {
}

// NewDefaultJSONWrapper returns an instance of DefaultJSONWrapper
func NewDefaultJSONWrapper() JSONWrapper {
	return &DefaultJSONWrapper{}
}

// Unmarshal converts a byte array into its Go representation
func (j *DefaultJSONWrapper) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Marshal serializes to a byte array
func (j *DefaultJSONWrapper) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
