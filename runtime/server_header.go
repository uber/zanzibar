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
	"net/textproto"

	"github.com/pkg/errors"
)

// Header defines methods on ServerHeaders
type Header interface {
	Get(key string) (string, bool)
	Add(key string, value string)
	Set(key string, value string)
	Keys() []string
	Ensure(keys []string) error
}

// ServerHTTPHeader wrapper to implement zanzibar Header interface
// on http.Header
type ServerHTTPHeader http.Header

// NewServerHTTPHeader creates a server http header
func NewServerHTTPHeader(h http.Header) ServerHTTPHeader {
	return ServerHTTPHeader(h)
}

// Get retrieves the first string stored on a given header. Bool
// return value is used to distinguish between the presence of a
// header with golang's zerovalue string and the absence of the string.
func (zh ServerHTTPHeader) Get(key string) (string, bool) {
	httpKey := textproto.CanonicalMIMEHeaderKey(key)
	h := zh[httpKey]
	if len(h) > 0 {
		return h[0], true
	}
	return "", false
}

// GetOrEmptyStr retrieves the first string stored on a given header or
// the empty string (golang's zero vlaue for string types)
func (zh ServerHTTPHeader) GetOrEmptyStr(key string) string {
	value, ok := zh.Get(key)
	if ok {
		return value
	}
	return ""
}

// GetAll retries all strings stored for this header.
func (zh ServerHTTPHeader) GetAll(key string) []string {
	httpKey := textproto.CanonicalMIMEHeaderKey(key)
	return zh[httpKey]
}

// Add appends a value for a given header.
func (zh ServerHTTPHeader) Add(key string, value string) {
	httpKey := textproto.CanonicalMIMEHeaderKey(key)
	zh[httpKey] = append(zh[httpKey], value)
}

// Set sets a value for a given header, overwriting all previous values.
func (zh ServerHTTPHeader) Set(key string, value string) {
	httpKey := textproto.CanonicalMIMEHeaderKey(key)
	zh[httpKey] = []string{value}
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

// Ensure returns error if the headers do not have the given keys
func (zh ServerHTTPHeader) Ensure(keys []string) error {
	for _, headerName := range keys {
		httpHeaderName := textproto.CanonicalMIMEHeaderKey(headerName)
		_, ok := zh[httpHeaderName]
		if !ok {
			return errors.New("Missing manditory header: " + headerName)
		}
	}
	return nil
}

// ServerTChannelHeader wrapper to implement zanzibar Header interface
// on map[string]string
type ServerTChannelHeader map[string]string

// Get retrieves the string value stored on a given header. Bool
// return value is used to distinguish between the presence of a
// header with golang's zerovalue string and the absence of the string.
func (th ServerTChannelHeader) Get(key string) (string, bool) {
	value, ok := th[key]
	return value, ok
}

// Add is an alias to Set.
func (th ServerTChannelHeader) Add(key string, value string) {
	th.Set(key, value)
}

// Set sets a value for a given header, overwriting the previous value.
func (th ServerTChannelHeader) Set(key string, value string) {
	// TODO: Canonicalize strings before inserting
	// Use textproto.CanonicalMIMEHeaderKey
	th[key] = value
}

// Keys returns a slice of header keys.
func (th ServerTChannelHeader) Keys() []string {
	keys := make([]string, len(th))
	i := 0
	for k := range th {
		keys[i] = k
		i++
	}
	return keys
}

// Ensure returns error if the headers do not have the given keys
func (th ServerTChannelHeader) Ensure(keys []string) error {
	for _, headerName := range keys {
		_, ok := th[headerName]
		if !ok {
			return errors.New("Missing manditory header: " + headerName)
		}
	}
	return nil
}
