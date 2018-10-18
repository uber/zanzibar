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
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go"
)

const (
	inboundCallsRecvd        = "inbound.calls.recvd"
	inboundCallsLatency      = "inbound.calls.latency"
	inboundCallsSuccess      = "inbound.calls.success"
	inboundCallsAppErrors    = "inbound.calls.app-errors"
	inboundCallsSystemErrors = "inbound.calls.system-errors"
	inboundCallsErrors       = "inbound.calls.errors"
	inboundCallsPanic        = "endpoints.panic"
	inboundCallsStatus       = "inbound.calls.status"

	// TChannel docs say it emits 'outbound.calls.sent':
	// http://tchannel.readthedocs.io/en/latest/metrics/#call-metrics
	// but uber/tchannel-go emits 'outbound.calls.send':
	// https://github.com/uber/tchannel-go/blob/3abb4c025c1663b383452339a22d918cf9d0be0b/outbound.go#L196
	outboundCallsSend = "outbound.calls.send"

	outboundCallsSent         = "outbound.calls.sent"
	outboundCallsLatency      = "outbound.calls.latency"
	outboundCallsSuccess      = "outbound.calls.success"
	outboundCallsAppErrors    = "outbound.calls.app-errors"
	outboundCallsSystemErrors = "outbound.calls.system-errors"
	outboundCallsErrors       = "outbound.calls.errors"
	outboundCallsStatus       = "outbound.calls.status"
)

var knownMetrics = []string{
	inboundCallsRecvd,
	inboundCallsLatency,
	inboundCallsSuccess,
	inboundCallsAppErrors,
	inboundCallsSystemErrors,
	inboundCallsErrors,
	inboundCallsStatus,

	outboundCallsSend,
	outboundCallsSent,
	outboundCallsLatency,
	outboundCallsSuccess,
	outboundCallsAppErrors,
	outboundCallsSystemErrors,
	outboundCallsErrors,
	outboundCallsStatus,
}

var knownTchannelErrors = map[tchannel.SystemErrCode]bool{
	tchannel.ErrCodeTimeout:    true,
	tchannel.ErrCodeBadRequest: true,
	tchannel.ErrCodeProtocol:   true,
	tchannel.ErrCodeCancelled:  true,
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

type systemErrorMap map[tchannel.SystemErrCode]tally.Counter

func newsystemErrorMap(scope tally.Scope, key string) systemErrorMap {
	ret := make(systemErrorMap)
	ret[tchannel.ErrCodeInvalid] = scope.Counter(key)
	for errorCode := range knownTchannelErrors {
		ret[errorCode] = scope.Tagged(map[string]string{
			"error": errorCode.MetricsKey(),
		}).Counter(key)
	}
	return ret
}

// IncrErr will increase according error counter
func (sem systemErrorMap) IncrErr(err error, count int64) {
	errCause := tchannel.GetSystemErrorCode(errors.Cause(err))
	counter, ok := sem[errCause]
	if !ok {
		counter = sem[0x00]
	}
	counter.Inc(count)
}

// HTTPStatusMap is statusCode -> according counter map
type HTTPStatusMap map[int]tally.Counter

func newHTTPStatusMap(scope tally.Scope, key string) HTTPStatusMap {
	ret := make(HTTPStatusMap)
	for statusCode := range knownStatusCodes {
		ret[statusCode] = scope.Tagged(map[string]string{
			"status": fmt.Sprintf("%d", statusCode),
		}).Counter(fmt.Sprintf("%s.%d", key, statusCode))
	}
	return ret
}

// IncrStatus will increase according status code counter
func (hsm HTTPStatusMap) IncrStatus(statusCode int, count int64) {
	if counter, ok := hsm[statusCode]; ok {
		counter.Inc(count)
	}
}

type inboundMetrics struct {
	Recvd tally.Counter // inbound.calls.recvd
}

type outboundMetrics struct {
	Sent tally.Counter // outbound.calls.sent
}

type commonMetrics struct {
	Latency tally.Timer   // [inbound|outbound].calls.latency
	Success tally.Counter // [inbound|outbound].calls.success
}

// EndpointMetrics ...
type EndpointMetrics struct {
	Panic tally.Counter // [inbound|outbound].calls.panics
	Recvd tally.Counter // [inbound|outbound].calls.panics
}

type tchannelMetrics struct {
	commonMetrics
	AppErrors    tally.Counter  // [inbound|outbound].calls.app-errors
	SystemErrors systemErrorMap // [inbound|outbound].calls.system-errors*
}

type httpMetrics struct {
	commonMetrics
	Errors tally.Counter // [inbound|outbound].calls.errors
	Status HTTPStatusMap // [inbound|outbound].calls.status.XXX
}

// InboundHTTPMetrics ...
type InboundHTTPMetrics struct {
	inboundMetrics
	httpMetrics
}

// InboundTChannelMetrics ...
type InboundTChannelMetrics struct {
	inboundMetrics
	tchannelMetrics
}

// OutboundHTTPMetrics ...
type OutboundHTTPMetrics struct {
	outboundMetrics
	httpMetrics
}

// OutboundTChannelMetrics ...
type OutboundTChannelMetrics struct {
	outboundMetrics
	tchannelMetrics
}

// NewEndpointMetrics returns endpoint panic metrics
func NewEndpointMetrics(scope tally.Scope) *EndpointMetrics {
	metrics := EndpointMetrics{}
	metrics.Panic = scope.Counter(inboundCallsPanic)
	metrics.Recvd = scope.Counter(inboundCallsRecvd)
	return &metrics
}

// NewInboundHTTPMetrics returns inbound HTTP metrics
func NewInboundHTTPMetrics(scope tally.Scope) *InboundHTTPMetrics {
	metrics := InboundHTTPMetrics{}
	metrics.Recvd = scope.Counter(inboundCallsRecvd)
	metrics.Latency = scope.Timer(inboundCallsLatency)
	metrics.Success = scope.Counter(inboundCallsSuccess)
	metrics.Errors = scope.Counter(inboundCallsErrors)
	metrics.Status = newHTTPStatusMap(scope, inboundCallsStatus)
	return &metrics
}

// NewInboundTChannelMetrics returns inbound TChannel metrics
func NewInboundTChannelMetrics(scope tally.Scope) *InboundTChannelMetrics {
	metrics := InboundTChannelMetrics{}
	metrics.Recvd = scope.Counter(inboundCallsRecvd)
	metrics.Latency = scope.Timer(inboundCallsLatency)
	metrics.Success = scope.Counter(inboundCallsSuccess)
	metrics.AppErrors = scope.Counter(inboundCallsAppErrors)
	metrics.SystemErrors = newsystemErrorMap(scope, inboundCallsSystemErrors)
	return &metrics
}

// NewOutboundHTTPMetrics returns outbound HTTP metrics
func NewOutboundHTTPMetrics(scope tally.Scope) *OutboundHTTPMetrics {
	metrics := OutboundHTTPMetrics{}
	metrics.Sent = scope.Counter(outboundCallsSent)
	metrics.Latency = scope.Timer(outboundCallsLatency)
	metrics.Success = scope.Counter(outboundCallsSuccess)
	metrics.Errors = scope.Counter(outboundCallsErrors)
	metrics.Status = newHTTPStatusMap(scope, outboundCallsStatus)
	return &metrics
}

// NewOutboundTChannelMetrics returns outbound TChannel metrics
func NewOutboundTChannelMetrics(scope tally.Scope) *OutboundTChannelMetrics {
	metrics := OutboundTChannelMetrics{}
	metrics.Sent = scope.Counter(outboundCallsSent)
	metrics.Latency = scope.Timer(outboundCallsLatency)
	metrics.Success = scope.Counter(outboundCallsSuccess)
	metrics.AppErrors = scope.Counter(outboundCallsAppErrors)
	metrics.SystemErrors = newsystemErrorMap(scope, outboundCallsSystemErrors)
	return &metrics
}
