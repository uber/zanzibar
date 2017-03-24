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

package zanzibar

import (
	"context"
	"net/http"
)

// Handler is an interface that compatable middlewares must implement to be used
// in the middleware stack. Handle should yield to the next middleware in the stack
// by invoking the next HandlerFn passed in.
type Handler interface {
	Handle(
		ctx context.Context,
		req *ServerHTTPRequest,
		res *ServerHTTPResponse,
		next HandlerFn)
}

// HandlerFunc is an adapter to allow functions to be directly added as MiddlewareStack handlers.
type HandlerFunc func(
	ctx context.Context,
	req *ServerHTTPRequest,
	res *ServerHTTPResponse,
	next HandlerFn)

// Handle executes the HandlerFunc.
func (h HandlerFunc) Handle(
	ctx context.Context,
	req *ServerHTTPRequest,
	res *ServerHTTPResponse,
	next HandlerFn) {
	h(ctx, req, res, next)
}

type middleware struct {
	handler Handler
	next    *middleware
}

func (m middleware) Handle(
	ctx context.Context,
	req *ServerHTTPRequest,
	res *ServerHTTPResponse) {
	m.handler.Handle(ctx, req, res, m.next.Handle)
}

// MiddlewareStack is a stack of Middleware Handlers that can be invoked as an Handler.
// MiddlewareStack middleware is evaluated in the order that they are added to the stack using
// the Use and UseHandler methods.
type MiddlewareStack struct {
	middleware middleware
	handlers   []Handler
}

// NewStack returns a new MiddlewareStack instance with no middleware preconfigured.
func NewStack(handlers ...Handler) *MiddlewareStack {
	return &MiddlewareStack{
		handlers:   handlers,
		middleware: build(handlers),
	}

}

// With returns a new MiddlewareStack instance that is a combination of the MiddlewareStack
// receiver's handlers and the provided handlers.
func (m *MiddlewareStack) With(handlers ...Handler) *MiddlewareStack {
	return NewStack(
		append(m.handlers, handlers...)...,
	)
}

// Handle calls Handle on the the handlers in the stack.
func (m *MiddlewareStack) Handle(
	ctx context.Context,
	req *ServerHTTPRequest,
	res *ServerHTTPResponse) {
	m.middleware.Handle(ctx, req, res)
}

// Use adds a Handler onto the middleware stack. Handlers are invoked in the order they are added to a MiddlewareStack.
func (m *MiddlewareStack) Use(handler Handler) {
	if handler == nil {
		panic("handler cannot be nil")
	}

	m.handlers = append(m.handlers, handler)
	m.middleware = build(m.handlers)
}

// UseFunc adds a MiddlewareStack-style handler function onto the middleware stack.
func (m *MiddlewareStack) UseFunc(
	handlerFunc func(
		ctx context.Context,
		req *ServerHTTPRequest,
		res *ServerHTTPResponse,
		next HandlerFn)) {
	m.Use(HandlerFunc(handlerFunc))
}

// Handlers returns a list of all the handlers in the current MiddlewareStack.
func (m *MiddlewareStack) Handlers() []Handler {
	return m.handlers
}

func build(handlers []Handler) middleware {
	if len(handlers) == 0 {
		return voidMiddleware()
	}

	var next middleware
	if len(handlers) > 1 {
		next = build(handlers[1:])
	} else {
		next = voidMiddleware()
	}

	return middleware{handlers[0], &next}
}

func voidMiddleware() middleware {
	return middleware{
		HandlerFunc(
			func(
				ctx context.Context,
				req *ServerHTTPRequest,
				res *ServerHTTPResponse,
				next HandlerFn) {
			}),
		&middleware{},
	}
}

// Wrap converts a zanzibar.HandlerFn into a middleware.Handler so it can be used
// in the middleware stack. The following zanzibar.HandlerFn is called after execution.
func Wrap(handlerFn HandlerFn) Handler {
	return HandlerFunc(
		func(
			ctx context.Context,
			req *ServerHTTPRequest,
			res *ServerHTTPResponse,
			next HandlerFn) {
			handlerFn(ctx, req, res)
			next(ctx, req, res)
		})
}

// UseHandlerFn adds a zanzibar.HandlerFn handler function onto the middleware stack.
func (m *MiddlewareStack) UseHandlerFn(
	handlerFunc func(
		ctx context.Context,
		req *ServerHTTPRequest,
		res *ServerHTTPResponse)) {
	m.Use(Wrap(handlerFunc))
}

// AdaptHTTPHandler wraps a http.Handler to support standard non-zanzibar middlewares.
func AdaptHTTPHandler(handler http.Handler) HandlerFunc {
	return func(
		ctx context.Context,
		req *ServerHTTPRequest,
		res *ServerHTTPResponse,
		next HandlerFn) {
		// TODO(sindelar): Implement this with a recorder and context propogation.
		handler.ServeHTTP(res.responseWriter, req.httpRequest)
		next(ctx, req, res)
	}
}

// DefaultStack returns a new middleware stack instance with the default middlewares.
//
// Logger - Request/Response Logging
// Tracer - Cross service request tracing
func DefaultStack() *MiddlewareStack {
	//TODO(sindelar): implement this.
	return nil
}
