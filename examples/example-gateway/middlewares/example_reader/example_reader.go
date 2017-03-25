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

package exampleReader

import (
	"net/http"

	"github.com/mcuadros/go-jsonschema-generator"
	"github.com/uber/zanzibar/examples/example-gateway/middlewares/example"
	zanzibar "github.com/uber/zanzibar/runtime"
)

type exampleReaderMiddleware struct {
	middlewareState MiddlewareState
	options         Options
}

// Options for middleware configuration
type Options struct {
	Foo string `json:"Foo"` // string to verify in shared
}

// MiddlewareState accessible by other middlewares and endpoint handler
// though the context object.
type MiddlewareState struct{}

// NewMiddleWare creates a new middleware that executes the next middleware
// after performing it's operations.
func NewMiddleWare(
	gateway *zanzibar.Gateway,
	options Options) zanzibar.MiddlewareHandle {
	return &exampleReaderMiddleware{
		options: options,
	}
}

func (m exampleReaderMiddleware) OwnState() interface{} {
	return m.middlewareState
}

// HandleRequest handles the requests before calling lower level middlewares.
func (m exampleReaderMiddleware) HandleRequest(
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
	shared zanzibar.SharedState) error {
	return nil
}

func (m exampleReaderMiddleware) HandleResponse(
	res *zanzibar.ServerHTTPResponse,
	shared zanzibar.SharedState) error {
	ss := shared.GetState("example").(example.MiddlewareState)
	if ss.Baz == m.options.Foo {
		res.StatusCode = http.StatusOK
		return nil
	}
	res.StatusCode = http.StatusNotFound
	return nil
}

// JSONSchema returns a schema definition of the configuration options for a middlware
func (m exampleReaderMiddleware) JSONSchema() *jsonschema.Document {
	s := &jsonschema.Document{}
	s.Read(&Options{})
	return s
}

func (m exampleReaderMiddleware) Name() string {
	return "example_reader"
}
