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

import "context"

// ExecutionStack is a stack of Adapter and Middleware Handlers that can be invoked as an Handle.
// ExecutionStack adapters and middlewares are evaluated for requests in the order that they are added to the stack
// followed by the underlying HandlerFn. The adapter responses are then executed in reverse.
type ExecutionStack struct {
	adapters    []AdapterHandle
	middlewares []MiddlewareHandle
	handle      HandlerFn
}

// NewExecutionStack returns a new ExecutionStack instance with no adapter or middleware preconfigured.
func NewExecutionStack(
	adapters []AdapterHandle,
	middlewares []MiddlewareHandle,
	handle HandlerFn) *ExecutionStack {
	return &ExecutionStack{
		handle:      handle,
		adapters:    adapters,
		middlewares: middlewares,
	}
}

// Adapters returns a list of all the adapter handlers in the current ExecutionStack.
func (m *ExecutionStack) Adapters() []AdapterHandle {
	return m.adapters
}

// Middlewares returns a list of all the middleware handlers in the current ExecutionStack.
func (m *ExecutionStack) Middlewares() []MiddlewareHandle {
	return m.middlewares
}

// SharedState used to access other adapters and middlewares in the chain.
type SharedState struct {
	// TODO(rnkim): One issue I can see is if a middleware and adapter have the same name
	stateDict map[string]interface{}
}

// NewSharedState constructs a ShardState
func NewSharedState(adapters []AdapterHandle, middlewares []MiddlewareHandle) SharedState {
	sharedState := SharedState{}
	sharedState.stateDict = make(map[string]interface{})

	for i := 0; i < len(adapters); i++ {
		sharedState.stateDict[adapters[i].Name()] = nil
	}
	for i := 0; i < len(middlewares); i++ {
		sharedState.stateDict[middlewares[i].Name()] = nil
	}

	return sharedState
}

// GetState returns the state from a different adapter or middleware
func (s SharedState) GetState(name string) interface{} {
	return s.stateDict[name]
}

// SetAdapterState sets value of an adapter shared state
func (s SharedState) SetAdapterState(m AdapterHandle, state interface{}) {
	s.stateDict[m.Name()] = state
}

// SetMiddlewareState sets value of an middleware shared state
func (s SharedState) SetMiddlewareState(m MiddlewareHandle, state interface{}) {
	s.stateDict[m.Name()] = state
}

// Handle executes the adapters and middlewares in a stack and underlying handler.
func (m *ExecutionStack) Handle(
	ctx context.Context,
	req *ServerHTTPRequest,
	res *ServerHTTPResponse) {

	shared := NewSharedState(m.adapters, m.middlewares)

	for i := 0; i < len(m.adapters); i++ {
		ok := m.adapters[i].HandleRequest(ctx, req, res, shared)
		// If a adapter errors and writes to the response header
		// then abort the rest of the stack and evaluate the response
		// handlers for the adapters seen so far.
		if ok == false {
			for j := i; j >= 0; j-- {
				m.adapters[j].HandleResponse(ctx, res, shared)
			}
			return
		}
	}

	for i := 0; i < len(m.middlewares); i++ {
		ok := m.middlewares[i].HandleRequest(ctx, req, res, shared)
		// If a middleware errors and writes to the response header
		// then abort the rest of the stack and evaluate the response
		// handlers for the adapters and middlewares seen so far.
		if ok == false {
			for j := i; j >= 0; j-- {
				m.middlewares[j].HandleResponse(ctx, res, shared)
			}

			for j := len(m.adapters) - 1; j >= 0; j-- {
				m.adapters[j].HandleResponse(ctx, res, shared)
			}
			return
		}
	}

	m.handle(ctx, req, res)

	for i := len(m.middlewares) - 1; i >= 0; i-- {
		m.middlewares[i].HandleResponse(ctx, res, shared)
	}

	for i := len(m.adapters) - 1; i >= 0; i-- {
		m.adapters[i].HandleResponse(ctx, res, shared)
	}
}
