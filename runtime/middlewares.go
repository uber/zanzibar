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
	"context"

	jsonschema "github.com/mcuadros/go-jsonschema-generator"
	"github.com/uber-go/tally"
)

const (
	middlewareRequestStatusTag = "middleware.request.status"
)

// MiddlewareStack is a stack of Middleware Handlers that can be invoked as an Handle.
// MiddlewareStack middlewares are evaluated for requests in the order that they are added to the stack
// followed by the underlying HandlerFn. The middleware responses are then executed in reverse.
type MiddlewareStack struct {
	middlewares []MiddlewareHandle
	handle      HandlerFn
}

// NewStack returns a new MiddlewareStack instance with no middleware preconfigured.
func NewStack(middlewares []MiddlewareHandle,
	handle HandlerFn) *MiddlewareStack {
	return &MiddlewareStack{
		handle:      handle,
		middlewares: middlewares,
	}
}

// Middlewares returns a list of all the handlers in the current MiddlewareStack.
func (m *MiddlewareStack) Middlewares() []MiddlewareHandle {
	return m.middlewares
}

// MiddlewareHandle used to define middleware
type MiddlewareHandle interface {
	// implement HandleRequest for your middleware. Return false
	// if the handler writes to the response body.
	HandleRequest(
		ctx context.Context,
		req *ServerHTTPRequest,
		res *ServerHTTPResponse,
		shared SharedState,
	) (context.Context, bool)
	// implement HandleResponse for your middleware. Return false
	// if the handler writes to the response body.
	HandleResponse(
		ctx context.Context,
		res *ServerHTTPResponse,
		shared SharedState,
	) context.Context
	// return any shared state for this middleware.
	JSONSchema() *jsonschema.Document
	Name() string
}

// SharedState used to access other middlewares in the chain.
type SharedState struct {
	middlewareDict map[string]interface{}
}

// NewSharedState constructs a ShardState
func NewSharedState(middlewares []MiddlewareHandle) SharedState {
	sharedState := SharedState{}
	sharedState.middlewareDict = make(map[string]interface{})

	for i := 0; i < len(middlewares); i++ {
		sharedState.middlewareDict[middlewares[i].Name()] = nil
	}
	return sharedState
}

// GetState returns the state from a different middleware
func (s SharedState) GetState(name string) interface{} {
	return s.middlewareDict[name]
}

// SetState sets value of a middleware shared state
func (s SharedState) SetState(m MiddlewareHandle, state interface{}) {
	s.middlewareDict[m.Name()] = state
}

// Handle executes the middlewares in a stack and underlying handler.
func (m *MiddlewareStack) Handle(
	ctx context.Context,
	req *ServerHTTPRequest,
	res *ServerHTTPResponse) context.Context {

	shared := NewSharedState(m.middlewares)

	for i := 0; i < len(m.middlewares); i++ {
		ctx, ok := m.middlewares[i].HandleRequest(ctx, req, res, shared)
		// If a middleware errors and writes to the response header
		// then abort the rest of the stack and evaluate the response
		// handlers for the middlewares seen so far.
		if ok == false {
			for j := i; j >= 0; j-- {
				m.middlewares[j].HandleResponse(ctx, res, shared)
			}

			//for error metrics only emit when there is gateway error and not request error
			// the percentage can be calculated via error_count/total_request
			if res.pendingStatusCode >= 500 {
				m.emitAvailabilityError(middlewareRequestStatusTag, m.middlewares[i].Name(), req.scope)
			}
			return ctx
		}
	}

	ctx = m.handle(ctx, req, res)

	for i := len(m.middlewares) - 1; i >= 0; i-- {
		m.middlewares[i].HandleResponse(ctx, res, shared)
	}
	return ctx
}

// emitAvailability is used to increment the error counter for a particular tagName.
func (m *MiddlewareStack) emitAvailabilityError(tagName string, middlewareName string, scope tally.Scope) {
	tagged := scope.Tagged(map[string]string{
		scopeTagStatus:     "error",
		scopeTagMiddleWare: middlewareName,
	})
	tagged.Counter(tagName).Inc(1)
}
