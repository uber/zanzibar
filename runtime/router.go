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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"github.com/uber/zanzibar/runtime/jsonwrapper"
	zrouter "github.com/uber/zanzibar/runtime/router"
	"go.uber.org/zap"
)

const (
	notFound         = "NotFound"
	methodNotAllowed = "MethodNotAllowed"
)

// HTTPRouter provides a HTTP router. It will match patterns in URLs and route them to provided HTTP handlers.
//
// This router has support for decoding path "parameters" in the URL into named values. An example:
//
//	var r zanzibar.HTTPRouter
//
//	r.Handle("GET", "/foo/:bar", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    params := zanzibar.ParamsFromContext(r.Context())
//	    w.Write("%s", params.Get("bar"))
//	}))
type HTTPRouter interface {
	// HTTPRouter implements a http.Handle as a convenience to allow HTTPRouter to be invoked by the standard library HTTP server.
	http.Handler

	// Handle associates a HTTP method and a pattern string to a HTTP handler function. If the method and pattern string
	// already exists, an error is returned.
	Handle(method, pattern string, handler http.Handler) error
}

// ParamsFromContext extracts the URL parameters that are embedded in the context by the Zanzibar HTTP router implementation.
func ParamsFromContext(ctx context.Context) url.Values {
	params := zrouter.ParamsFromContext(ctx)
	urlValues := make(url.Values)
	for _, paramValue := range params {
		urlValues.Add(paramValue.Key, paramValue.Value)
	}
	return urlValues
}

// HandlerFn is a func that handles ServerHTTPRequest
type HandlerFn func(
	context.Context,
	*ServerHTTPRequest,
	*ServerHTTPResponse,
) context.Context

// RouterEndpoint struct represents an endpoint that can be registered
// into the router itself.
type RouterEndpoint struct {
	EndpointName string
	HandlerName  string
	HandlerFn    HandlerFn
	JSONWrapper  jsonwrapper.JSONWrapper

	contextExtractor ContextExtractor
	contextLogger    ContextLogger
	scope            tally.Scope
	tracer           opentracing.Tracer
	config           *StaticConfig
}

// NewRouterEndpoint creates an endpoint that can be registered to HTTPRouter
func NewRouterEndpoint(
	extractor ContextExtractor,
	deps *DefaultDependencies,
	endpointID string,
	handlerID string,
	handler HandlerFn,
) *RouterEndpoint {
	return &RouterEndpoint{
		EndpointName:     endpointID,
		HandlerName:      handlerID,
		HandlerFn:        handler,
		contextExtractor: extractor,
		contextLogger:    deps.ContextLogger,
		scope:            deps.Scope,
		tracer:           deps.Tracer,
		JSONWrapper:      deps.JSONWrapper,
		config:           deps.Config,
	}
}

// HandleRequest is called by the router and starts the request
func (endpoint *RouterEndpoint) HandleRequest(
	w http.ResponseWriter,
	r *http.Request,
) {
	// TODO: (lu) get timeout from endpoint config
	//_, ok := ctx.Deadline()
	//if !ok {
	//	var cancel context.CancelFunc
	//	ctx, cancel = context.WithTimeout(ctx, time.Duration(100)*time.Millisecond)
	//	defer cancel()
	//}

	urlValues := ParamsFromContext(r.Context())
	req := NewServerHTTPRequest(w, r, urlValues, endpoint)
	ctx := req.Context()
	endpoint.HandlerFn(ctx, req, req.res)
	req.res.flush(ctx)
}

// httpRouter data structure to handle and register endpoints
type httpRouter struct {
	gateway                  *Gateway
	httpRouter               *zrouter.Router
	notFoundEndpoint         *RouterEndpoint
	methodNotAllowedEndpoint *RouterEndpoint
	panicCount               tally.Counter
	routeMap                 map[string]*RouterEndpoint

	requestUUIDHeaderKey string
}

var _ HTTPRouter = (*httpRouter)(nil)

// NewHTTPRouter allocates a HTTP router
func NewHTTPRouter(gateway *Gateway) HTTPRouter {
	deps := &DefaultDependencies{
		Logger:        gateway.Logger,
		ContextLogger: gateway.ContextLogger,
		Scope:         gateway.RootScope,
		Tracer:        gateway.Tracer,
		Config:        gateway.Config,
	}

	router := &httpRouter{
		notFoundEndpoint: NewRouterEndpoint(
			gateway.ContextExtractor, deps,
			notFound, notFound, nil,
		),
		methodNotAllowedEndpoint: NewRouterEndpoint(
			gateway.ContextExtractor, deps,
			methodNotAllowed, methodNotAllowed, nil,
		),
		gateway:    gateway,
		panicCount: gateway.RootScope.Counter("runtime.router.panic"),
		routeMap:   make(map[string]*RouterEndpoint),

		requestUUIDHeaderKey: gateway.requestUUIDHeaderKey,
	}

	router.httpRouter = &zrouter.Router{
		HandleMethodNotAllowed: true,
		NotFound:               http.HandlerFunc(router.handleNotFound),
		MethodNotAllowed:       http.HandlerFunc(router.handleMethodNotAllowed),
		PanicHandler:           router.handlePanic,
		WhitelistedPaths:       router.getWhitelistedPaths(),
	}
	return router
}

// Register register a handler function.
func (router *httpRouter) Handle(method, prefix string, handler http.Handler) (err error) {
	h := func(w http.ResponseWriter, r *http.Request) {
		reqUUID := r.Header.Get(router.requestUUIDHeaderKey)
		if reqUUID == "" {
			reqUUID = uuid.New()
		}
		ctx := withRequestUUID(r.Context(), reqUUID)
		ctx = WithLogFields(ctx, zap.String(logFieldRequestUUID, reqUUID))
		r = r.WithContext(ctx)
		handler.ServeHTTP(w, r)
	}

	return router.httpRouter.Handle(method, prefix, http.HandlerFunc(h))
}

// ServeHTTP implements the http.Handle as a convenience to allow HTTPRouter to be invoked by the standard library HTTP server.
func (router *httpRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.httpRouter.ServeHTTP(w, r)
}

func (router *httpRouter) handlePanic(
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
	logger := router.gateway.Logger
	for k, v := range r.Header {
		val, _ := json.Marshal(v)
		logger = logger.With(zap.String(k, string(val)))
	}
	logger.Error(
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

func (router *httpRouter) handleNotFound(
	w http.ResponseWriter,
	r *http.Request,
) {
	scopeTags := map[string]string{
		scopeTagEndpoint: router.notFoundEndpoint.EndpointName,
		scopeTagHandler:  router.notFoundEndpoint.HandlerName,
		scopeTagProtocol: scopeTagHTTP,
	}

	ctx := r.Context()
	ctx = WithScopeTags(ctx, scopeTags)
	r = r.WithContext(ctx)
	req := NewServerHTTPRequest(w, r, nil, router.notFoundEndpoint)
	http.NotFound(w, r)
	req.res.StatusCode = http.StatusNotFound
	req.res.finish(ctx)
}

func (router *httpRouter) handleMethodNotAllowed(
	w http.ResponseWriter,
	r *http.Request,
) {
	scopeTags := map[string]string{
		scopeTagEndpoint: router.methodNotAllowedEndpoint.EndpointName,
		scopeTagHandler:  router.methodNotAllowedEndpoint.HandlerName,
		scopeTagProtocol: scopeTagHTTP,
	}

	ctx := r.Context()
	ctx = WithScopeTags(ctx, scopeTags)
	r = r.WithContext(ctx)
	req := NewServerHTTPRequest(w, r, nil, router.methodNotAllowedEndpoint)
	http.Error(w,
		http.StatusText(http.StatusMethodNotAllowed),
		http.StatusMethodNotAllowed,
	)
	req.res.StatusCode = http.StatusMethodNotAllowed
	req.res.finish(ctx)
}

func (router *httpRouter) getWhitelistedPaths() []string {
	var whitelistedPaths []string
	if router.gateway.Config != nil &&
		router.gateway.Config.ContainsKey("router.whitelistedPaths") {
		router.gateway.Config.MustGetStruct("router.whitelistedPaths", &whitelistedPaths)
	}
	return whitelistedPaths
}
