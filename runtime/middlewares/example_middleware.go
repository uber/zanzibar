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

package exampleMiddleware

import (
	"net/http"

	zanzibar "github.com/uber/zanzibar/runtime"
)

type Options struct {
	foo string
	bar int
}

type MiddlewareState struct {
	baz string
}

//func middlewareFoo(next zanzibar.HandlerFn) zanzibar.HandlerFn {
//	ctx.Put("token", "c9e452805dee5044ba520198628abcaa")
//	next.ServeHTTP(w, r)
//}

func WithHeader(key, value string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header.Add(key, value)
			h.ServeHTTP(w, r)
		})
	}
}

func NewMiddleWare(gateway zanzibar.Gateway, options Options, next zanzibar.HandlerFn) zanzibar.HandlerFn {
	return func(h zanzibar.HandlerFn) zanzibar.HandlerFn {
		h.ctx.Put("token", MiddlewareState{baz: "c9e452805dee5044ba520198628abcaa"})
		next.ServeHTTP(w, r)
	}
}
