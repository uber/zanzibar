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
	"net/http"
)

// ServerHeader defines methods on ServerHeaders
type ServerHeader interface {
	Get(key string) (string, bool)
	Add(key string, value string)
	Set(key string, value string)
	Keys() []string
}

// ServerHTTPHeader wrapper to implement ServerHeaderInterface
// on http.Header
type ServerHTTPHeader http.Header

// TODO: Add a TChannel Implementation

// Get retrieves the first string stored on a given header. Bool
// return value is used to distinguish between the presence of a
// header with golang's zerovalue string and the absence of the string.
func (zh ServerHTTPHeader) Get(
	key string,
) (string, bool) {
	// TODO: Canonicalize strings before lookup.
	// Use textproto.CanonicalMIMEHeaderKey
	h := zh[key]
	if len(h) > 0 {
		return h[0], true
	}
	return "", false
}

// Add appends a value for a given header.
func (zh ServerHTTPHeader) Add(
	key string, value string,
) {
	// TODO: Canonicalize strings before inserting
	// Use textproto.CanonicalMIMEHeaderKey
	zh[key] = append(zh[key], value)
}

// Set sets a value for a given header, overwriting all previous values.
func (zh ServerHTTPHeader) Set(
	key string, value string,
) {
	// TODO: Canonicalize strings before inserting
	// Use textproto.CanonicalMIMEHeaderKey
	h := make([]string, 1)
	h[0] = value
	zh[key] = h
}

// Keys returns a slice of header keys.
func (zh ServerHTTPHeader) Keys() []string {
	keys := make([]string, len(zh))
	i := 0
	for k := range zh {
		keys[i] = k
		i++
	}
	return keys
}
