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
	"net/http"
	"time"
)

const (
	endpointRequest             = "endpoint.request"
	endpointSuccess             = "endpoint.success"
	endpointStatus              = "endpoint.status"
	endpointSystemErrors        = "endpoint.system-errors"
	endpointLatency             = "endpoint.latency"
	endpointLatencyHist         = "endpoint.latency-hist"
	endpointOverheadLatency     = "endpoint.overhead.latency"
	endpointOverheadLatencyHist = "endpoint.overhead.latency-hist"
	endpointOverheadRatio       = "endpoint.overhead.latency.ratio"

	// MetricEndpointPanics is endpoint level panic counter
	MetricEndpointPanics = "endpoint.panic"

	// endpointAppErrors is the metric name for endpoint level application error for HTTP
	endpointAppErrors = "endpoint.app-errors"
	// MetricEndpointAppErrors is the metric name for endpoint level application error for TChannel
	MetricEndpointAppErrors = "endpoint.app-errors"

	clientRequest      = "client.request"
	clientSuccess      = "client.success"
	clientStatus       = "client.status"
	clientErrors       = "client.errors"
	clientAppErrors    = "client.app-errors"
	clientSystemErrors = "client.system-errors"
	clientLatency      = "client.latency"
	clientLatencyHist  = "client.latency-hist"

	// clientHTTPUnmarshalError is the metric for tracking errors due to unmarshalling json responses
	clientHTTPUnmarshalError = "client.http-unmarshal-error"
	// clientTchannelReadError is the metric for tracking errors in reading tchannel response
	clientTchannelUnmarshalError = "client.tchannel-unmarshal-error"

	// shadow headers and environment
	shadowEnvironment = "shadow"
	environmentKey    = "env"
	apienvironmentKey = "apienvironment"

	// TraceIDKey is the log field key containing the associated trace id
	TraceIDKey = "trace.traceId"
	// TraceSpanKey is the log field key containing the associated span id
	TraceSpanKey = "trace.span"
	// TraceSampledKey is the log field key for whether a trace was sampled or not
	TraceSampledKey = "trace.sampled"

	// ClientResponseDurationKey is the key denoting a downstream response duration
	ClientResponseDurationKey = "client.response.duration"
	// ClientTypeKey denotes the type of the client, usually http / tchannel / client-less / custom
	ClientTypeKey = "client.type"
)

var knownMetrics = []string{
	"inbound.calls.recvd",
	"inbound.calls.latency",
	"inbound.calls.success",
	"inbound.calls.app-errors",
	"inbound.calls.system-errors",
	"inbound.calls.errors",
	"inbound.calls.status",

	// TChannel docs say it emits 'outbound.calls.sent':
	// http://tchannel.readthedocs.io/en/latest/metrics/#call-metrics
	// but uber/tchannel-go emits 'outbound.calls.send':
	// https://github.com/uber/tchannel-go/blob/3abb4c025c1663b383452339a22d918cf9d0be0b/outbound.go#L196
	"outbound.calls.send",
	"outbound.calls.sent",
	"outbound.calls.latency",
	"outbound.calls.success",
	"outbound.calls.app-errors",
	"outbound.calls.system-errors",
	"outbound.calls.errors",
	"outbound.calls.status",
}

var knownStatusCodes = map[int]bool{
	http.StatusContinue:                      true, // 100
	http.StatusSwitchingProtocols:            true, // 101
	http.StatusProcessing:                    true, // 102
	http.StatusOK:                            true, // 200
	http.StatusCreated:                       true, // 201
	http.StatusAccepted:                      true, // 202
	http.StatusNonAuthoritativeInfo:          true, // 203
	http.StatusNoContent:                     true, // 204
	http.StatusResetContent:                  true, // 205
	http.StatusPartialContent:                true, // 206
	http.StatusMultiStatus:                   true, // 207
	http.StatusAlreadyReported:               true, // 208
	http.StatusIMUsed:                        true, // 226
	http.StatusMultipleChoices:               true, // 300
	http.StatusMovedPermanently:              true, // 301
	http.StatusFound:                         true, // 302
	http.StatusSeeOther:                      true, // 303
	http.StatusNotModified:                   true, // 304
	http.StatusUseProxy:                      true, // 305
	http.StatusTemporaryRedirect:             true, // 307
	http.StatusPermanentRedirect:             true, // 308
	http.StatusBadRequest:                    true, // 400
	http.StatusUnauthorized:                  true, // 401
	http.StatusPaymentRequired:               true, // 402
	http.StatusForbidden:                     true, // 403
	http.StatusNotFound:                      true, // 404
	http.StatusMethodNotAllowed:              true, // 405
	http.StatusNotAcceptable:                 true, // 406
	http.StatusProxyAuthRequired:             true, // 407
	http.StatusRequestTimeout:                true, // 408
	http.StatusConflict:                      true, // 409
	http.StatusGone:                          true, // 410
	http.StatusLengthRequired:                true, // 411
	http.StatusPreconditionFailed:            true, // 412
	http.StatusRequestEntityTooLarge:         true, // 413
	http.StatusRequestURITooLong:             true, // 414
	http.StatusUnsupportedMediaType:          true, // 415
	http.StatusRequestedRangeNotSatisfiable:  true, // 416
	http.StatusExpectationFailed:             true, // 417
	http.StatusTeapot:                        true, // 418
	http.StatusUnprocessableEntity:           true, // 422
	http.StatusLocked:                        true, // 423
	http.StatusFailedDependency:              true, // 424
	http.StatusUpgradeRequired:               true, // 426
	http.StatusPreconditionRequired:          true, // 428
	http.StatusTooManyRequests:               true, // 429
	http.StatusRequestHeaderFieldsTooLarge:   true, // 431
	http.StatusUnavailableForLegalReasons:    true, // 451
	http.StatusInternalServerError:           true, // 500
	http.StatusNotImplemented:                true, // 501
	http.StatusBadGateway:                    true, // 502
	http.StatusServiceUnavailable:            true, // 503
	http.StatusGatewayTimeout:                true, // 504
	http.StatusHTTPVersionNotSupported:       true, // 505
	http.StatusVariantAlsoNegotiates:         true, // 506
	http.StatusInsufficientStorage:           true, // 507
	http.StatusLoopDetected:                  true, // 508
	http.StatusNotExtended:                   true, // 510
	http.StatusNetworkAuthenticationRequired: true, // 511
}

var noContentStatusCodes = map[int]bool{
	http.StatusNoContent:   true, // 204
	http.StatusNotModified: true, // 304
}

//DefaultBackOffTimeAcrossRetriesConf is the time to wait before attempting new attempt
var DefaultBackOffTimeAcrossRetriesConf = 10

//DefaultBackOffTimeAcrossRetries is the time in MS to wait before attempting new attempt
var DefaultBackOffTimeAcrossRetries = time.Duration(DefaultBackOffTimeAcrossRetriesConf) * time.Millisecond

//DefaultScaleFactor is multiplied with timeoutPerAttempt
var DefaultScaleFactor = 1.1
