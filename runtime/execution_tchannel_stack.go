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

	"go.uber.org/thriftrw/wire"
)

// ExecutionTchannelStack is a stack of Adapter and Middleware Handlers that can be invoked as an Handle.
// ExecutionTchannelStack adapters and middlewares are evaluated for requests in the order that they are added to the stack
// followed by the underlying HandlerFn. The adapter and middleware responses are then executed in reverse.
type ExecutionTchannelStack struct {
	adapters        []AdapterTchannelHandle
	middlewares     []MiddlewareTchannelHandle
	tchannelHandler TChannelHandler
}

// NewExecutionTchannelStack returns a new ExecutionTchannelStack instance with no adapter or middleware preconfigured.
func NewExecutionTchannelStack(
	adapters []AdapterTchannelHandle,
	middlewares []MiddlewareTchannelHandle,
	handler TChannelHandler) *ExecutionTchannelStack {
	return &ExecutionTchannelStack{
		tchannelHandler: handler,
		adapters:        adapters,
		middlewares:     middlewares,
	}
}

// TchannelAdapters returns a list of all the adapter handlers in the current ExecutionTchannelStack.
func (m *ExecutionTchannelStack) TchannelAdapters() []AdapterTchannelHandle {
	return m.adapters
}

// TchannelMiddlewares returns a list of all the middleware handlers in the current ExecutionTchannelStack.
func (m *ExecutionTchannelStack) TchannelMiddlewares() []MiddlewareTchannelHandle {
	return m.middlewares
}

// TchannelSharedState used to access other adapters or middlewares in the chain.
type TchannelSharedState struct {
	stateDict map[string]interface{}
}

// NewTchannelSharedState constructs a ShardState
func NewTchannelSharedState(adapters []AdapterTchannelHandle,
	middlewares []MiddlewareTchannelHandle) TchannelSharedState {
	sharedState := TchannelSharedState{}
	sharedState.stateDict = make(map[string]interface{})

	for i := 0; i < len(adapters); i++ {
		sharedState.stateDict[adapters[i].Name()] = nil
	}
	for i := 0; i < len(adapters); i++ {
		sharedState.stateDict[middlewares[i].Name()] = nil
	}

	return sharedState
}

// GetTchannelState returns the state from a different adapter or middleware
func (s TchannelSharedState) GetTchannelState(name string) interface{} {
	return s.stateDict[name]
}

// SetTchannelAdapterState sets value of an adapter shared state
func (s TchannelSharedState) SetTchannelAdapterState(m AdapterTchannelHandle, state interface{}) {
	s.stateDict[m.Name()] = state
}

// SetTchannelMiddlewareState sets value of a middleware shared state
func (s TchannelSharedState) SetTchannelMiddlewareState(m MiddlewareTchannelHandle, state interface{}) {
	s.stateDict[m.Name()] = state
}

// Handle executes the adapters and middlewares in a stack and underlying handler.
func (m *ExecutionTchannelStack) Handle(
	ctx context.Context,
	reqHeaders map[string]string,
	wireValue *wire.Value) (bool, RWTStruct, map[string]string, error) {
	var res RWTStruct
	var ok bool

	shared := NewTchannelSharedState(m.adapters, m.middlewares)

	for i := 0; i < len(m.adapters); i++ {
		ok, err := m.adapters[i].HandleRequest(ctx, reqHeaders, wireValue, shared)
		if ok == false {
			for j := i; j >= 0; j-- {
				m.adapters[j].HandleResponse(ctx, res, shared)
			}
			return ok, nil, map[string]string{}, err
		}
	}

	for i := 0; i < len(m.middlewares); i++ {
		ok, err := m.middlewares[i].HandleRequest(ctx, reqHeaders, wireValue, shared)
		if ok == false {
			for j := i; j >= 0; j-- {
				m.middlewares[j].HandleResponse(ctx, res, shared)
			}

			for j := len(m.adapters) - 1; j >= 0; j-- {
				m.adapters[j].HandleResponse(ctx, res, shared)
			}
			return ok, nil, map[string]string{}, err
		}
	}

	ok, res, resHeaders, err := m.tchannelHandler.Handle(ctx, reqHeaders, wireValue)

	for i := len(m.middlewares) - 1; i >= 0; i-- {
		res = m.middlewares[i].HandleResponse(ctx, res, shared)
	}

	for i := len(m.adapters) - 1; i >= 0; i-- {
		res = m.adapters[i].HandleResponse(ctx, res, shared)
	}

	return ok, res, resHeaders, err
}
