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
	"time"

	jsonschema "github.com/mcuadros/go-jsonschema-generator"
)

const (
	middlewareRequestLatencyTag  = "middleware.requests.latency"
	middlewareResponseLatencyTag = "middleware.responses.latency"
	middlewareRequestStatusTag   = "middleware.requests.status"
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
	start := time.Now()

	for i := 0; i < len(m.middlewares); i++ {
		ctx, ok := m.middlewares[i].HandleRequest(ctx, req, res, shared)
		// If a middleware errors and writes to the response header
		// then abort the rest of the stack and evaluate the response
		// handlers for the middlewares seen so far.
		if ok == false {
			//record latency for middlewares requests in unsuccesful case as the middleware requests calls are terminated
			m.recordLatency(middlewareRequestLatencyTag, start, req)

			start = time.Now() // start the timer for middleware responses
			for j := i; j >= 0; j-- {
				m.middlewares[j].HandleResponse(ctx, res, shared)
			}
			//record latency for middlewares responses in unsuccesful case
			m.recordLatency(middlewareResponseLatencyTag, start, req)

			//for error metrics only emit when there is gateway error and not request error
			if res.pendingStatusCode >= 500 {
				m.emitAvailability(middlewareRequestStatusTag, "error", req)
			} else {
				m.emitAvailability(middlewareRequestStatusTag, "success", req)
			}
			return ctx
		}
	}
	// record latency for middlewares requests in successful case
	m.recordLatency(middlewareRequestLatencyTag, start, req)
	//emit success metric for middleware requests
	m.emitAvailability(middlewareRequestStatusTag, "success", req)

	ctx = m.handle(ctx, req, res)

	start = time.Now()
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		m.middlewares[i].HandleResponse(ctx, res, shared)
	}
	// record latency for middlewares responses in successful case
	m.recordLatency(middlewareResponseLatencyTag, start, req)
	return ctx
}

// recordLatency measures the latency as per the tagName and start time given.
func (m *MiddlewareStack) recordLatency(tagName string, start time.Time, req *ServerHTTPRequest) {
	elapsed := time.Now().Sub(start)
	req.scope.Timer(tagName).Record(elapsed)
}

// emitAvailability is used to increment the success/error counter for a particular tagName.
func (m *MiddlewareStack) emitAvailability(tagName string, status string, req *ServerHTTPRequest) {
	tagged := req.scope.Tagged(map[string]string{
		scopeTagStatus: status,
	})
	tagged.Counter(tagName).Inc(1)
}
