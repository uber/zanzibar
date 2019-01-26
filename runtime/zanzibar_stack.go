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

type ZanzibarStack struct {
	adapters    []AdapterHandle
	middlewares []MiddlewareHandle
	handle      HandlerFn
}

func NewZanzibarStack(
	adapters 	[]AdapterHandle,
	middlewares []MiddlewareHandle,
	handle      HandlerFn) *ZanzibarStack {
	return &ZanzibarStack{
		handle:      handle,
		adapters:    adapters,
		middlewares: middlewares,
	}
}

func (m *ZanzibarStack) Adapters() []AdapterHandle {
	return m.adapters
}

func (m *ZanzibarStack) Middlewares() []MiddlewareHandle {
	return m.middlewares
}

// SharedState used to access other middlewares and adapters in the chain.
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

// SetState sets value of a middleware shared state
func (s SharedState) SetState(m MiddlewareHandle, state interface{}) {
	s.stateDict[m.Name()] = state
}

// Handle executes the adapters and middlewares in a stack and underlying handler.
func (m *ZanzibarStack) Handle(
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
		// handlers for the adapters seen so far.
		if ok == false {
			for j := i; j >= 0; j-- {
				m.middlewares[j].HandleResponse(ctx, res, shared)
			}

			for j := len(m.adapters)-1; j >= 0; j-- {
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