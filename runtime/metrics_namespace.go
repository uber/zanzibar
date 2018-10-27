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
	"net/http"

	"github.com/uber/tchannel-go"
)

const (
	inboundCallsRecvd        = "inbound.calls.recvd"
	inboundCallsLatency      = "inbound.calls.latency"
	inboundCallsSuccess      = "inbound.calls.success"
	inboundCallsAppErrors    = "inbound.calls.app-errors"
	inboundCallsSystemErrors = "inbound.calls.system-errors"
	inboundCallsErrors       = "inbound.calls.errors"
	inboundCallsStatus       = "inbound.calls.status"

	// InboundCallsPanic is endpoint level panic counter
	InboundCallsPanic = "inbound.calls.panic"

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
