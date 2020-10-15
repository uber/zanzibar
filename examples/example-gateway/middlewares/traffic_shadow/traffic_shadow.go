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

package examplereader

import (
	"context"
	"fmt"
	"github.com/mcuadros/go-jsonschema-generator"
	"github.com/uber/zanzibar/examples/example-gateway/build/middlewares/traffic_shadow/module"
	zanzibar "github.com/uber/zanzibar/runtime"
)

type trafficShadowMiddleware struct {
	deps    *module.Dependencies
	options Options
}

// Options for middleware configuration
type Options struct {
	Foo string `json:"Foo"` // string to verify in shared
}

// MiddlewareState accessible by other middlewares and endpoint handler
// though the context object.
type MiddlewareState struct{}

// NewMiddleware creates a new middleware that executes the next middleware
// after performing it's operations.
func NewMiddleware(
	deps *module.Dependencies,
	options Options,
) zanzibar.MiddlewareHandle {
	return &trafficShadowMiddleware{
		deps:    deps,
		options: options,
	}
}

// HandleRequest handles the requests before calling lower level middlewares.
func (m *trafficShadowMiddleware) HandleRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
	shared zanzibar.SharedState,
) bool {
	return true
}

func (m *trafficShadowMiddleware) HandleResponse(
	ctx context.Context,
	res *zanzibar.ServerHTTPResponse,
	shared zanzibar.SharedState,
) {

	fmt.Println("Hello m in traffic_shadow middleware")
	for i, num := range res.Request.Header.Keys() {
		fmt.Println(i , " : " , num)
	}
	//ss := shared.GetState("example").(example.MiddlewareState)
	//if ss.Baz == m.options.Foo {
	//	res.StatusCode = http.StatusOK
	//}
	//res.StatusCode = http.StatusNotFound
}

// JSONSchema returns a schema definition of the configuration options for a middlware
func (m *trafficShadowMiddleware) JSONSchema() *jsonschema.Document {
	s := &jsonschema.Document{}
	s.Read(&Options{})
	return s
}

func (m *trafficShadowMiddleware) Name() string {
	return "traffic_shadow"
}
