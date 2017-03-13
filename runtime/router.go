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
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/uber-go/tally"
)

var knownStatusCodes = []int{
	http.StatusContinue,           // 100
	http.StatusSwitchingProtocols, // 101
	http.StatusProcessing,         // 102

	http.StatusOK,                   // 200
	http.StatusCreated,              // 201
	http.StatusAccepted,             // 202
	http.StatusNonAuthoritativeInfo, // 203
	http.StatusNoContent,            // 204
	http.StatusResetContent,         // 205
	http.StatusPartialContent,       // 206
	http.StatusMultiStatus,          // 207
	http.StatusAlreadyReported,      // 208
	http.StatusIMUsed,               // 226

	http.StatusMultipleChoices,   // 300
	http.StatusMovedPermanently,  // 301
	http.StatusFound,             // 302
	http.StatusSeeOther,          // 303
	http.StatusNotModified,       // 304
	http.StatusUseProxy,          // 305
	http.StatusTemporaryRedirect, // 307
	http.StatusPermanentRedirect, // 308

	http.StatusBadRequest,                   // 400
	http.StatusUnauthorized,                 // 401
	http.StatusPaymentRequired,              // 402
	http.StatusForbidden,                    // 403
	http.StatusNotFound,                     // 404
	http.StatusMethodNotAllowed,             // 405
	http.StatusNotAcceptable,                // 406
	http.StatusProxyAuthRequired,            // 407
	http.StatusRequestTimeout,               // 408
	http.StatusConflict,                     // 409
	http.StatusGone,                         // 410
	http.StatusLengthRequired,               // 411
	http.StatusPreconditionFailed,           // 412
	http.StatusRequestEntityTooLarge,        // 413
	http.StatusRequestURITooLong,            // 414
	http.StatusUnsupportedMediaType,         // 415
	http.StatusRequestedRangeNotSatisfiable, // 416
	http.StatusExpectationFailed,            // 417
	http.StatusTeapot,                       // 418
	http.StatusUnprocessableEntity,          // 422
	http.StatusLocked,                       // 423
	http.StatusFailedDependency,             // 424
	http.StatusUpgradeRequired,              // 426
	http.StatusPreconditionRequired,         // 428
	http.StatusTooManyRequests,              // 429
	http.StatusRequestHeaderFieldsTooLarge,  // 431
	http.StatusUnavailableForLegalReasons,   // 451

	http.StatusInternalServerError,           // 500
	http.StatusNotImplemented,                // 501
	http.StatusBadGateway,                    // 502
	http.StatusServiceUnavailable,            // 503
	http.StatusGatewayTimeout,                // 504
	http.StatusHTTPVersionNotSupported,       // 505
	http.StatusVariantAlsoNegotiates,         // 506
	http.StatusInsufficientStorage,           // 507
	http.StatusLoopDetected,                  // 508
	http.StatusNotExtended,                   // 510
	http.StatusNetworkAuthenticationRequired, // 511
}

// HandlerFn is a func that handles IncomingHTTPRequest
type HandlerFn func(context.Context, *IncomingHTTPRequest, *Gateway)

// EndpointMetrics contains pre allocated metrics structures
// These are pre-allocated to cache tags maps and for performance
type EndpointMetrics struct {
	requestRecvd   tally.Counter
	requestLatency tally.Timer
	statusCodes    map[int]tally.Counter
}

// Endpoint struct represents an endpoint that can be registered
// into the router itself.
type Endpoint struct {
	EndpointName string
	HandlerName  string
	HandlerFn    HandlerFn

	metrics EndpointMetrics
	gateway *Gateway
}

// NewEndpoint creates an endpoint with all the necessary data
func NewEndpoint(
	gateway *Gateway,
	endpointName string,
	handlerName string,
	handler HandlerFn,
) *Endpoint {
	endpointTags := map[string]string{
		"endpoint": endpointName,
		"handler":  handlerName,
	}
	endpointScope := gateway.MetricScope.Tagged(endpointTags)
	requestRecvd := endpointScope.Counter("inbound.calls.recvd")
	requestLatency := endpointScope.Timer("inbound.calls.latency")
	statusCodes := make(map[int]tally.Counter, len(knownStatusCodes))

	for _, statusCode := range knownStatusCodes {
		metricName := "inbound.calls.status." + strconv.Itoa(statusCode)
		statusCodes[statusCode] = endpointScope.Counter(metricName)
	}

	return &Endpoint{
		EndpointName: endpointName,
		HandlerName:  handlerName,
		HandlerFn:    handler,
		gateway:      gateway,

		metrics: EndpointMetrics{
			requestRecvd:   requestRecvd,
			statusCodes:    statusCodes,
			requestLatency: requestLatency,
		},
	}
}

// HandleRequest is called by the router and starts the request
func (endpoint *Endpoint) HandleRequest(
	w http.ResponseWriter, r *http.Request, params httprouter.Params,
) {
	inc := NewIncomingHTTPRequest(w, r, params, endpoint)

	fn := endpoint.HandlerFn
	fn(r.Context(), inc, endpoint.gateway)
}

// Router data structure to handle and register endpoints
type Router struct {
	httpRouter *httprouter.Router
	gateway    *Gateway
}

// NewRouter allocates a router
func NewRouter(gateway *Gateway) *Router {
	router := &Router{
		gateway: gateway,
	}

	router.httpRouter = &httprouter.Router{
		// We handle trailing slash in Register() without redirect
		RedirectTrailingSlash:  false,
		RedirectFixedPath:      false,
		HandleMethodNotAllowed: true,
		NotFound:               router.handleNotFound,
		MethodNotAllowed:       router.handleMethodNotAllowed,
		// TODO add panic handler
		// PanicHandler:           router.handlePanic,
	}

	return router
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.httpRouter.ServeHTTP(w, r)
}

// RegisterRaw register a raw handler function.
// Such a function take raw http req/writer.
// Use this only to integrated third-party, like pprof debug handlers
func (router *Router) RegisterRaw(
	method string, prefix string, handler http.HandlerFunc,
) {
	router.httpRouter.Handler(method, prefix, handler)
}

// Register will register an endpoint with the router.
func (router *Router) Register(
	method string, urlpattern string, endpoint *Endpoint,
) {
	canonicalPattern := urlpattern
	if canonicalPattern[len(canonicalPattern)-1] == '/' {
		canonicalPattern = canonicalPattern[:len(canonicalPattern)-1]
	}

	// Support trailing slash going to the same endpoint.
	router.httpRouter.Handle(
		method, canonicalPattern, endpoint.HandleRequest,
	)
	router.httpRouter.Handle(
		method, canonicalPattern+"/", endpoint.HandleRequest,
	)
}

func (router *Router) handleNotFound(w http.ResponseWriter, r *http.Request) {
	// TODO custom NotFound
	http.NotFound(w, r)
}

func (router *Router) handleMethodNotAllowed(
	w http.ResponseWriter, r *http.Request,
) {
	// TODO custom MethodNotAllowed
	http.Error(w,
		http.StatusText(http.StatusMethodNotAllowed),
		http.StatusMethodNotAllowed,
	)
}
