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
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/julienschmidt/httprouter"
	"github.com/uber-go/tally"
	"go.uber.org/zap/zapcore"
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

const statusCodeZapName = "statusCode"

// HandlerFn is a func that handles ServerHTTPRequest
type HandlerFn func(
	context.Context,
	*ServerHTTPRequest,
	*ServerHTTPResponse,
)

// EndpointMetrics contains pre allocated metrics structures
// These are pre-allocated to cache tags maps and for performance
type EndpointMetrics struct {
	requestRecvd   tally.Counter
	requestLatency tally.Timer
	statusCodes    map[int]tally.Counter
}

// RouterEndpoint struct represents an endpoint that can be registered
// into the router itself.
type RouterEndpoint struct {
	EndpointName string
	HandlerName  string
	HandlerFn    HandlerFn

	metrics EndpointMetrics
	gateway *Gateway
}

// NewRouterEndpoint creates an endpoint with all the necessary data
func NewRouterEndpoint(
	gateway *Gateway,
	endpointName string,
	handlerName string,
	handler HandlerFn,
) *RouterEndpoint {
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

	return &RouterEndpoint{
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
func (endpoint *RouterEndpoint) HandleRequest(
	w http.ResponseWriter, r *http.Request, params httprouter.Params,
) {
	reqFields := logRequestFields(r)
	var resFields []zapcore.Field

	defer func() {
		writeLogs(endpoint.gateway.Logger, reqFields, resFields)
	}()

	req := NewServerHTTPRequest(w, r, params, endpoint)

	fn := endpoint.HandlerFn

	ctx := r.Context()

	// TODO: (lu) get timeout from endpoint config
	//_, ok := ctx.Deadline()
	//if !ok {
	//	var cancel context.CancelFunc
	//	ctx, cancel = context.WithTimeout(ctx, time.Duration(100)*time.Millisecond)
	//	defer cancel()
	//}

	fn(ctx, req, req.res)

	req.res.flush()
	resFields = logResponseFields(req.res)
}

// HTTPRouter data structure to handle and register endpoints
type HTTPRouter struct {
	httpRouter *httprouter.Router
	gateway    *Gateway
}

// NewHTTPRouter allocates a HTTP router
func NewHTTPRouter(gateway *Gateway) *HTTPRouter {
	router := &HTTPRouter{
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

func (router *HTTPRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.httpRouter.ServeHTTP(w, r)
}

// RegisterRaw register a raw handler function.
// Such a function take raw http req/writer.
// Use this only to integrated third-party, like pprof debug handlers
func (router *HTTPRouter) RegisterRaw(
	method string, prefix string, handler http.HandlerFunc,
) {
	router.httpRouter.Handler(method, prefix, handler)
}

// Register will register an endpoint with the router.
func (router *HTTPRouter) Register(
	method string, urlpattern string, endpoint *RouterEndpoint,
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

func (router *HTTPRouter) handleNotFound(w http.ResponseWriter, r *http.Request) {
	resFields := []zapcore.Field{
		zap.Int(statusCodeZapName, 404),
	}
	writeLogs(router.gateway.Logger, logRequestFields(r), resFields)
	// TODO custom NotFound
	// A NotFound request is not started...
	// TODO: inc.finish()
	http.NotFound(w, r)
}

func (router *HTTPRouter) handleMethodNotAllowed(
	w http.ResponseWriter, r *http.Request,
) {
	resFields := []zapcore.Field{
		zap.Int(statusCodeZapName, 405),
	}
	writeLogs(router.gateway.Logger, logRequestFields(r), resFields)
	// TODO: Remove coverage ignore when body unmarshaling supported.
	// TODO custom MethodNotAllowed
	http.Error(w,
		http.StatusText(http.StatusMethodNotAllowed),
		http.StatusMethodNotAllowed,
	)
}

func logRequestFields(r *http.Request) []zapcore.Field {
	// TODO: Allocating a fixed size array causes the zap logger to fail
	// with ``unknown field type: { 0 0  <nil>}'' errors. Investigate this
	// further to see if we can avoid reallocating underlying arrays for slices.
	var fields []zapcore.Field
	for k, v := range r.Header {
		if len(v) > 0 {
			fields = append(fields, zap.String("Request-Header-"+k, v[0]))
		}
	}

	fields = append(fields, zap.String("method", r.Method))
	fields = append(fields, zap.String("remoteAddr", r.RemoteAddr))
	fields = append(fields, zap.String("pathname", r.URL.RequestURI()))
	fields = append(fields, zap.String("host", r.Host))
	fields = append(fields, zap.Time("timestamp", time.Now().UTC()))
	// TODO add endpoint.id and endpoint.handlerId
	// TODO log jaeger trace span

	// TODO: Do not log body by default because PII and bandwidth.
	// Temporarily log during the developement cycle
	// TODO: Add a gateway level configurable body unmarshaller
	// to extract only non-PII info.

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		fields = append(fields, zap.String("Request-Body", string(body)))
	}
	return fields
}

func logResponseFields(res *ServerHTTPResponse) []zapcore.Field {
	var fields []zapcore.Field

	fields = append(fields, zap.Int(statusCodeZapName, res.pendingStatusCode))
	fields = append(fields, zap.Time("timestamp-finished", res.finishTime))
	return fields
}

func writeLogs(l *zap.Logger, reqFlds []zapcore.Field, resFlds []zapcore.Field) {
	fields := reqFlds
	if resFlds != nil {
		fields = append(reqFlds, resFlds...)
	}
	l.Info(
		"Finished an incoming server HTTP request",
		fields...,
	)
}
