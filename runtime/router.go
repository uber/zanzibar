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
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

const (
	notFound         = "NotFound"
	methodNotAllowed = "MethodNotAllowed"
)

// HandlerFn is a func that handles ServerHTTPRequest
type HandlerFn func(
	context.Context,
	*ServerHTTPRequest,
	*ServerHTTPResponse,
)

// RouterEndpoint struct represents an endpoint that can be registered
// into the router itself.
type RouterEndpoint struct {
	EndpointName string
	HandlerName  string
	HandlerFn    HandlerFn

	logger  *zap.Logger
	metrics *InboundHTTPMetrics
	tracer  opentracing.Tracer
}

// NewRouterEndpoint creates an endpoint with all the necessary data
func NewRouterEndpoint(
	logger *zap.Logger,
	scope tally.Scope,
	tracer opentracing.Tracer,
	endpointID string,
	handlerID string,
	handler HandlerFn,
) *RouterEndpoint {
	logger = logger.With(
		zap.String("endpointID", endpointID),
		zap.String("handlerID", handlerID),
	)
	scope = scope.Tagged(map[string]string{
		"endpoint": endpointID,
		"handler":  handlerID,
	})
	return &RouterEndpoint{
		EndpointName: endpointID,
		HandlerName:  handlerID,
		HandlerFn:    handler,
		logger:       logger,
		metrics:      NewInboundHTTPMetrics(scope),
		tracer:       tracer,
	}
}

// HandleRequest is called by the router and starts the request
func (endpoint *RouterEndpoint) HandleRequest(
	w http.ResponseWriter,
	r *http.Request,
	params httprouter.Params,
) {

	ctx := withRequestFields(r.Context())
	reqUUID := ""
	if ctxUUID := GetRequestUUIDFromCtx(ctx); ctxUUID != nil {
		reqUUID = ctxUUID.String()
	}
	req := NewServerHTTPRequest(w, r, params, endpoint)
	req.Logger = req.Logger.With(
		zap.String(string(requestUUIDKey), reqUUID),
	)

	// TODO: (lu) get timeout from endpoint config
	//_, ok := ctx.Deadline()
	//if !ok {
	//	var cancel context.CancelFunc
	//	ctx, cancel = context.WithTimeout(ctx, time.Duration(100)*time.Millisecond)
	//	defer cancel()
	//}

	endpoint.HandlerFn(ctx, req, req.res)
	req.res.flush()
}

// HTTPRouter data structure to handle and register endpoints
type HTTPRouter struct {
	gateway                  *Gateway
	httpRouter               *httprouter.Router
	notFoundEndpoint         *RouterEndpoint
	methodNotAllowedEndpoint *RouterEndpoint
	panicCount               tally.Counter
	routeMap                 map[string]*RouterEndpoint
}

// NewHTTPRouter allocates a HTTP router
func NewHTTPRouter(gateway *Gateway) *HTTPRouter {
	router := &HTTPRouter{
		notFoundEndpoint: NewRouterEndpoint(
			gateway.Logger, gateway.AllHostScope, gateway.Tracer,
			notFound, notFound, nil,
		),
		methodNotAllowedEndpoint: NewRouterEndpoint(
			gateway.Logger, gateway.AllHostScope, gateway.Tracer,
			methodNotAllowed, methodNotAllowed, nil,
		),
		gateway:    gateway,
		panicCount: gateway.PerHostScope.Counter("runtime.router.panic"),
		routeMap:   make(map[string]*RouterEndpoint),
	}
	router.httpRouter = &httprouter.Router{
		// We handle trailing slash in Register() without redirect
		RedirectTrailingSlash:  false,
		RedirectFixedPath:      false,
		HandleMethodNotAllowed: true,
		NotFound:               router.handleNotFound,
		MethodNotAllowed:       router.handleMethodNotAllowed,
		PanicHandler:           router.handlePanic,
	}
	return router
}

func (router *HTTPRouter) handlePanic(
	w http.ResponseWriter, r *http.Request, v interface{},
) {
	err, ok := v.(error)
	if !ok {
		err = errors.Errorf("http router panic: %v", v)
	}
	_, ok = err.(fmt.Formatter)
	if !ok {
		err = errors.Wrap(err, "wrapped")
	}

	router.gateway.Logger.Error(
		"A http request handler paniced",
		zap.Error(err),
		zap.String("pathname", r.URL.RequestURI()),
	)
	router.panicCount.Inc(1)

	http.Error(w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

func (router *HTTPRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.httpRouter.ServeHTTP(w, r)
}

// RegisterRaw register a raw handler function.
// Such a function take raw http req/writer.
// Use this only to integrated third-party, like pprof debug handlers
func (router *HTTPRouter) RegisterRaw(
	method, prefix string,
	handler http.HandlerFunc,
) {
	router.httpRouter.Handler(method, prefix, handler)
}

// Register will register an endpoint with the router.
func (router *HTTPRouter) Register(
	method, urlpattern string,
	endpoint *RouterEndpoint,
) error {
	canonicalPattern := urlpattern
	if canonicalPattern[len(canonicalPattern)-1] == '/' {
		canonicalPattern = canonicalPattern[:len(canonicalPattern)-1]
	}

	key := urlpattern + "|" + method
	if _, ok := router.routeMap[key]; ok {
		return fmt.Errorf("handler for '%s %s' is already registered", method, urlpattern)
	}
	router.routeMap[key] = endpoint

	// Support trailing slash going to the same endpoint.
	router.httpRouter.Handle(
		method, canonicalPattern, endpoint.HandleRequest,
	)
	router.httpRouter.Handle(
		method, canonicalPattern+"/", endpoint.HandleRequest,
	)

	return nil
}

func (router *HTTPRouter) handleNotFound(
	w http.ResponseWriter,
	r *http.Request,
) {
	req := NewServerHTTPRequest(w, r, nil, router.notFoundEndpoint)
	http.NotFound(w, r)
	req.res.StatusCode = http.StatusNotFound
	req.res.finish()
}

func (router *HTTPRouter) handleMethodNotAllowed(
	w http.ResponseWriter,
	r *http.Request,
) {
	req := NewServerHTTPRequest(w, r, nil, router.methodNotAllowedEndpoint)
	http.Error(w,
		http.StatusText(http.StatusMethodNotAllowed),
		http.StatusMethodNotAllowed,
	)
	req.res.StatusCode = http.StatusMethodNotAllowed
	req.res.finish()
}
