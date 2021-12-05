// Copyright (c) 2018 Uber Technologies, Inc.
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

package examplecustom

import (
	"context"

	"github.com/mcuadros/go-jsonschema-generator"
	zanzibar "github.com/uber/zanzibar/runtime"
)

// Dependencies ...
type Dependencies struct {
	Default *zanzibar.DefaultDependencies
}

// Middleware ...
type Middleware struct {
	Dependencies *Dependencies
}

// MiddlewareState accessible by other middlewares and endpoint handler
// though the context object.
type MiddlewareState struct {
	Baz string
}

// NewMiddleware creates a new middleware that executes the next middleware
// after performing it's operations.
func NewMiddleware(dependencies *Dependencies) Middleware {
	return Middleware{
		Dependencies: dependencies,
	}
}

// HandleRequest handles the requests before calling lower level middlewares.
func (m *Middleware) HandleRequest(
	ctx context.Context,
	_ *zanzibar.ServerHTTPRequest,
	_ *zanzibar.ServerHTTPResponse,
	_ zanzibar.SharedState,
) (context.Context, bool) {
	return ctx, true
}

// HandleResponse ...
func (m *Middleware) HandleResponse(
	ctx context.Context,
	_ *zanzibar.ServerHTTPResponse,
	_ zanzibar.SharedState,
) context.Context {
	return ctx
}

// JSONSchema returns a schema definition of the configuration options for a middlware
func (m *Middleware) JSONSchema() *jsonschema.Document {
	return &jsonschema.Document{}
}

// Name ...
func (m *Middleware) Name() string {
	return "example_custom"
}
