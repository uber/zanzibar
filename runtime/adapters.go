// Copyright (c) 2019 Uber Technologies, Inc.
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
	"context"
	"github.com/mcuadros/go-jsonschema-generator"
)

// AdapterStack is a stack of Adapter Handlers that can be invoked as an Handle.
// AdapterStack adapters are evaluated for requests in the order that they are added to the stack
// followed by the underlying HandlerFn. The adapter responses are then executed in reverse.
type AdapterStack struct {
	adapters []AdapterHandle
	handle   HandlerFn
}

// NewAdapterStack returns a new AdapterStack instance with no adapter preconfigured.
func NewAdapterStack(adapters []AdapterHandle, handle HandlerFn) *AdapterStack {
	return &AdapterStack{handle: handle, adapters: adapters}
}

// Adapters returns a list of all the handlers in the current AdapterStack.
func (m *AdapterStack) Adapters() []AdapterHandle {
	return m.adapters
}

// AdapterHandle used to define adapter
type AdapterHandle interface {
	// implement HandleRequest for your adapter. Return false
	// if the handler writes to the response body.
	HandleRequest(
		ctx context.Context,
		req *ServerHTTPRequest,
		res *ServerHTTPResponse,
	) bool
	// implement HandleResponse for your adapter. Return false
	// if the handler writes to the response body.
	HandleResponse(
		ctx context.Context,
		res *ServerHTTPResponse,
	)
	JSONSchema() *jsonschema.Document
	Name() string
}

// Handle executes the adapters in a stack and underlying handler.
func (m *AdapterStack) Handle(
	ctx context.Context,
	req *ServerHTTPRequest,
	res *ServerHTTPResponse) {

	for i := 0; i < len(m.adapters); i++ {
		ok := m.adapters[i].HandleRequest(ctx, req, res)
		// If a adapter errors and writes to the response header
		// then abort the rest of the stack and evaluate the response
		// handlers for the adapters seen so far.
		if ok == false {
			for j := i; j >= 0; j-- {
				m.adapters[j].HandleResponse(ctx, res)
			}
			return
		}
	}

	m.handle(ctx, req, res)

	for i := len(m.adapters) - 1; i >= 0; i-- {
		m.adapters[i].HandleResponse(ctx, res)
	}
}
