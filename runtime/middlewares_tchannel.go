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

package zanzibar

import (
	"context"

	"github.com/mcuadros/go-jsonschema-generator"
	"go.uber.org/thriftrw/wire"
)

// MiddlewareTchannelStack is a stack of Middleware Handlers that can be invoked as an Handler.
// MiddlewareTchannelStack middlewares are evaluated for requests in the order that they are added to the stack
// followed by the underlying HandlerFn. The middleware responses are then executed in reverse.
type MiddlewareTchannelStack struct {
	middlewares     []MiddlewareTchannelHandle
	tchannelHandler TChannelHandler
}

// NewTchannelStack returns a new MiddlewareStack instance with no middleware preconfigured.
func NewTchannelStack(middlewares []MiddlewareTchannelHandle,
	handler TChannelHandler) *MiddlewareTchannelStack {
	return &MiddlewareTchannelStack{
		tchannelHandler: handler,
		middlewares:     middlewares,
	}
}

// TchannelMiddlewares returns a list of all the handlers in the current MiddlewareStack.
func (m *MiddlewareTchannelStack) TchannelMiddlewares() []MiddlewareTchannelHandle {
	return m.middlewares
}

// MiddlewareTchannelHandle used to define middleware
type MiddlewareTchannelHandle interface {
	// implement HandleRequest for your middleware. Return false
	// if the handler writes to the response body.
	HandleRequest(
		ctx context.Context,
		reqHeaders map[string]string,
		wireValue *wire.Value,
		shared TchannelSharedState) bool

	// implement HandleResponse for your middleware. Return false
	// if the handler writes to the response body.
	HandleResponse(
		ctx context.Context,
		wireValue *wire.Value,
		shared TchannelSharedState) RWTStruct

	// return any shared state for this middleware.
	JSONSchema() *jsonschema.Document
	Name() string
}

// TchannelSharedState used to access other middlewares in the chain.
type TchannelSharedState struct {
	middlewareDict map[string]interface{}
}

// NewTchannelSharedState constructs a ShardState
func NewTchannelSharedState(middlewares []MiddlewareTchannelHandle) TchannelSharedState {
	sharedState := TchannelSharedState{}
	sharedState.middlewareDict = make(map[string]interface{})
	for i := 0; i < len(middlewares); i++ {
		sharedState.middlewareDict[middlewares[i].Name()] = nil
	}

	return sharedState
}

// GetTchannelState returns the state from a different middleware
func (s TchannelSharedState) GetTchannelState(name string) interface{} {
	return s.middlewareDict[name]
}

// SetTchannelState sets value of a middleware shared state
func (s TchannelSharedState) SetTchannelState(m MiddlewareTchannelHandle, state interface{}) {
	s.middlewareDict[m.Name()] = state
}

// Handle executes the middlewares in a stack and underlying handler.
func (m *MiddlewareTchannelStack) Handle(
	ctx context.Context,
	reqHeaders map[string]string,
	wireValue *wire.Value) (bool, RWTStruct, map[string]string, error) {
	var res RWTStruct
	var ok bool

	shared := NewTchannelSharedState(m.middlewares)
	for i := 0; i < len(m.middlewares); i++ {
		ok := m.middlewares[i].HandleRequest(ctx, reqHeaders, wireValue, shared)

		// If a middleware errors and writes to the response header
		// then abort the rest of the stack and evaluate the response
		// handlers for the middlewares seen so far.
		if ok == false {
			for j := i; j >= 0; j-- {
				res = m.middlewares[j].HandleResponse(ctx, wireValue, shared)
			}

			return ok, res, nil, nil
		}
	}

	ok, res, resHeaders, err := m.tchannelHandler.Handle(ctx, reqHeaders, wireValue)
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		val, error := res.ToWire()
		if error != nil {
			res = m.middlewares[i].HandleResponse(ctx, &val, shared)
		}
	}

	return ok, res, resHeaders, err
}
