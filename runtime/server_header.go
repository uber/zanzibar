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
	"net/http"
	"net/textproto"
	"strings"

	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Header defines methods on ServerHeaders
type Header interface {
	// Get mirrors the implementation of http.header and returns a single header value.
	// When a key contains multiple values, -only- the first one is returned.
	// Ref: https://golang.org/pkg/net/http/#Header.Get
	Get(key string) (string, bool)
	// Values mirrors the implementation of http.header and returns a slice of header values.
	// When a key contains multiple values, the entire collection is returned.
	// Ref: https://golang.org/pkg/net/http/#Header.Values
	Values(key string) ([]string, bool)
	Add(key string, value string)
	Set(key string, value string)
	// Unset unsets the value for a given header. Can be safely called multiple times
	Unset(key string)
	Keys() []string
	// Deprecated: Use EnsureContext instead
	Ensure(keys []string, logger *zap.Logger) error
	EnsureContext(ctx context.Context, keys []string, logger ContextLogger) error
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
// header with golang's zerovalue string and the absence of the header.
func (zh ServerHTTPHeader) Get(key string) (string, bool) {
	httpKey := textproto.CanonicalMIMEHeaderKey(key)
	h := zh[httpKey]
	if len(h) > 0 {
		return h[0], true
	}
	return "", false
}

// Values retrieves the entire collection of values stored on a given header.
// Bool return value is used to distinguish between the presence of a
// header with golang's zerovalue slice and the absence of the header.
func (zh ServerHTTPHeader) Values(key string) ([]string, bool) {
	httpKey := textproto.CanonicalMIMEHeaderKey(key)
	if _, ok := zh[httpKey]; ok {
		return textproto.MIMEHeader(zh).Values(key), true
	}
	return []string{}, false
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

// Unset unsets the value for a given header. Can be safely called multiple times
func (zh ServerHTTPHeader) Unset(key string) {
	httpKey := textproto.CanonicalMIMEHeaderKey(key)
	delete(zh, httpKey)
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
//
// Deprecated: Use EnsureContext instead
func (zh ServerHTTPHeader) Ensure(keys []string, logger *zap.Logger) error {
	loggerCtx := NewContextLogger(logger)
	ctx := context.Background()
	return zh.EnsureContext(ctx, keys, loggerCtx)
}

// EnsureContext returns error if the headers do not have the given keys
func (zh ServerHTTPHeader) EnsureContext(ctx context.Context, keys []string, logger ContextLogger) error {
	missing := make([]string, 0, len(keys))
	for _, header := range keys {
		h := textproto.CanonicalMIMEHeaderKey(header)
		if _, ok := zh[h]; !ok {
			missing = append(missing, header)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	err := errors.Errorf(
		"Missing mandatory headers: %s",
		strings.Join(missing, ", "),
	)
	logger.WarnZ(ctx, "Missing mandatory headers",
		zap.Error(err),
		zap.Strings("headers", missing),
	)
	return err
}

// ServerTChannelHeader wrapper to implement zanzibar Header interface
// on map[string]string
// Unlike http.Header, tchannel headers are case sensitive and should be
// keyed with lower case. TChannel protocol does not mention header case
// sensitivity, so it is up to implementation.
type ServerTChannelHeader map[string]string

// Get retrieves the string value stored on a given header. Bool
// return value is used to distinguish between the presence of a
// header with golang's zerovalue string and the absence of the string.
func (th ServerTChannelHeader) Get(key string) (string, bool) {
	value, ok := th[key]
	return value, ok
}

// Values retrieves the entire collection of values stored on a given header.
// Bool return value is used to distinguish between the presence of a
// header with golang's zerovalue slice and the absence of the header.
func (th ServerTChannelHeader) Values(key string) ([]string, bool) {
	if value, ok := th.Get(key); ok {
		// In the case of TChannel, ServerTChannelHeader does not support
		//  multiple, disparate values so we defer to Get and package it
		// in a slice to meet the interface's requirement.
		return []string{value}, ok
	}
	return []string{}, false
}

// Add is an alias to Set.
func (th ServerTChannelHeader) Add(key string, value string) {
	th.Set(key, value)
}

// Set sets a value for a given header, overwriting the previous value.
func (th ServerTChannelHeader) Set(key string, value string) {
	th[key] = value
}

// Unset unsets the value for a given header. Can be safely called multiple times
func (th ServerTChannelHeader) Unset(key string) {
	delete(th, key)
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
//
// Deprecated: Use EnsureContext instead
func (th ServerTChannelHeader) Ensure(keys []string, logger *zap.Logger) error {
	loggerCtx := NewContextLogger(logger)
	ctx := context.Background()
	return th.EnsureContext(ctx, keys, loggerCtx)
}

// EnsureContext returns error if the headers do not have the given keys
func (th ServerTChannelHeader) EnsureContext(ctx context.Context, keys []string, logger ContextLogger) error {
	missing := make([]string, 0, len(keys))
	for _, header := range keys {
		if _, ok := th[header]; !ok {
			missing = append(missing, header)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	err := errors.Errorf(
		"Missing mandatory headers: %s",
		strings.Join(missing, ", "),
	)
	logger.WarnZ(ctx, "Missing mandatory headers",
		zap.Error(err),
		zap.Strings("headers", missing),
	)
	return err
}
