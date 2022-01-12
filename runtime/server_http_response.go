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
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/uber/jaeger-client-go"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"github.com/uber/zanzibar/runtime/jsonwrapper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ServerHTTPResponse struct manages server http response
type ServerHTTPResponse struct {
	Request              *ServerHTTPRequest
	StatusCode           int
	responseWriter       http.ResponseWriter
	flushed              bool
	finished             bool
	finishTime           time.Time
	DownstreamFinishTime time.Duration
	ClientType           string
	pendingBodyBytes     []byte
	pendingBodyObj       interface{}
	pendingStatusCode    int
	contextLogger        ContextLogger
	scope                tally.Scope
	jsonWrapper          jsonwrapper.JSONWrapper
	Err                  error
}

// NewServerHTTPResponse is helper function to alloc ServerHTTPResponse
func NewServerHTTPResponse(
	w http.ResponseWriter,
	req *ServerHTTPRequest,
) *ServerHTTPResponse {
	return &ServerHTTPResponse{
		Request:        req,
		StatusCode:     200,
		responseWriter: w,
		contextLogger:  req.contextLogger,
		scope:          req.scope,
		jsonWrapper:    req.jsonWrapper,
	}
}

// finish will handle final logic, like metrics
func (res *ServerHTTPResponse) finish(ctx context.Context) {
	logFields := GetLogFieldsFromCtx(ctx)
	if !res.Request.started {
		/* coverage ignore next line */
		res.contextLogger.Error(ctx,
			"Forgot to start server response",
			append(logFields, zap.String("path", res.Request.URL.Path))...,
		)
		/* coverage ignore next line */
		return
	}
	if res.finished {
		/* coverage ignore next line */
		res.contextLogger.Error(ctx,
			"Finished a server response multiple times",
			append(logFields, zap.String("path", res.Request.URL.Path))...,
		)
		/* coverage ignore next line */
		return
	}
	res.finished = true
	res.finishTime = time.Now()

	_, known := knownStatusCodes[res.StatusCode]

	tagged := res.scope.Tagged(map[string]string{
		scopeTagStatus:     fmt.Sprintf("%d", res.StatusCode), // no need to put this tag on the context because this is the end of response life cycle
		scopeTagClientType: res.ClientType,
	})
	delta := res.finishTime.Sub(res.Request.startTime)
	tagged.Timer(endpointLatency).Record(delta)
	tagged.Histogram(endpointLatencyHist, tally.DefaultBuckets).RecordDuration(delta)

	if res.DownstreamFinishTime != 0 {
		overhead := delta - res.DownstreamFinishTime
		overheadRatio := overhead.Seconds() / delta.Seconds()
		tagged.Timer(endpointOverheadLatency).Record(overhead)
		tagged.Histogram(endpointOverheadLatencyHist, tally.DefaultBuckets).RecordDuration(overhead)
		tagged.Gauge(endpointOverheadRatio).Update(overheadRatio)
	}

	if !known {
		res.contextLogger.Error(ctx,
			"Unknown status code",
			append(logFields, zap.Int("UnknownStatusCode", res.StatusCode))...,
		)
	} else {
		tagged.Counter(endpointStatus).Inc(1)
	}

	logFn := res.contextLogger.Debug
	if !known || res.StatusCode >= 400 && res.StatusCode < 600 {
		tagged.Counter(endpointAppErrors).Inc(1)
		logFn = res.contextLogger.WarnZ
	}

	span := res.Request.GetSpan()
	if span != nil {
		span.Finish()
	}

	logFn(ctx,
		fmt.Sprintf("Finished an incoming server HTTP request with %d status code", res.StatusCode),
		append(logFields, serverHTTPLogFields(res.Request, res)...)...,
	)
}

func serverHTTPLogFields(req *ServerHTTPRequest, res *ServerHTTPResponse) []zapcore.Field {
	fields := []zapcore.Field{
		zap.String("method", req.httpRequest.Method),
		zap.String("remoteAddr", req.httpRequest.RemoteAddr),
		zap.String("pathname", req.httpRequest.URL.RequestURI()),
		zap.String("host", req.httpRequest.Host),
		zap.Time("timestamp-started", req.startTime),
		zap.Time("timestamp-finished", res.finishTime),
		zap.Int("statusCode", res.StatusCode),
	}

	if span := req.GetSpan(); span != nil {
		jc, ok := span.Context().(jaeger.SpanContext)
		if ok {
			fields = append(fields,
				zap.String(TraceSpanKey, jc.SpanID().String()),
				zap.String(TraceIDKey, jc.TraceID().String()),
				zap.Bool(TraceSampledKey, jc.IsSampled()),
			)
		}
	}

	for k, v := range res.Headers() {
		if len(v) > 0 {
			fields = append(fields, zap.String(
				fmt.Sprintf("%s-%s", logFieldEndpointResponseHeaderPrefix, k),
				strings.Join(v, ", "),
			))
		}
	}

	if res.Err != nil {
		fields = append(fields, zap.Error(res.Err))

		cause := errors.Cause(res.Err)
		if cause != nil && cause != res.Err {
			fields = append(fields, zap.NamedError("errorCause", cause))
		}
	}

	return fields
}

// SendErrorString helper to send an error string
func (res *ServerHTTPResponse) SendErrorString(
	statusCode int, errMsg string,
) {
	res.WriteJSONBytes(statusCode, nil,
		[]byte(`{"error":"`+errMsg+`"}`),
	)
}

// SendError helper to send an server error message, propagates underlying cause to logs etc.
func (res *ServerHTTPResponse) SendError(
	statusCode int, errMsg string, errCause error,
) {
	res.Err = errCause
	res.WriteJSONBytes(statusCode, nil,
		[]byte(`{"error":"`+errMsg+`"}`),
	)
}

// WriteBytes writes a byte[] slice that is valid Response
func (res *ServerHTTPResponse) WriteBytes(
	statusCode int, headers Header, bytes []byte,
) {
	if headers != nil {
		for _, k := range headers.Keys() {
			v, ok := headers.Get(k)
			if ok {
				res.responseWriter.Header().Set(k, v)
			}
		}
	}

	res.pendingStatusCode = statusCode
	res.pendingBodyBytes = bytes
}

// WriteJSONBytes writes a byte[] slice that is valid json to Response
func (res *ServerHTTPResponse) WriteJSONBytes(
	statusCode int, headers Header, bytes []byte,
) {
	if headers == nil {
		headers = ServerHTTPHeader{}
	}

	headers.Add("content-type", "application/json")
	res.WriteBytes(statusCode, headers, bytes)
}

// MarshalResponseJSON serializes a json serializable into bytes
func (res *ServerHTTPResponse) MarshalResponseJSON(body interface{}) []byte {
	ctx := res.Request.Context()
	if body == nil {
		res.SendError(500, "Could not serialize json response", errors.New("No Body JSON"))
		res.contextLogger.Error(ctx, "Could not serialize nil pointer body")
		return nil
	}
	bytes, err := res.jsonWrapper.Marshal(body)
	if err != nil {
		res.SendError(500, "Could not serialize json response", err)
		res.contextLogger.Error(ctx, "Could not serialize json response", zap.Error(err))
		return nil
	}
	return bytes
}

// SendResponse sets content-type if not present and fills Response
func (res *ServerHTTPResponse) SendResponse(statusCode int, headers Header, body interface{}, bytes []byte) {
	contentTypePresent := false
	if headers != nil {
		for _, k := range headers.Keys() {
			v, ok := headers.Get(k)
			if ok {
				if k == "Content-Type" {
					contentTypePresent = true
				}
				res.responseWriter.Header().Set(k, v)
			}
		}
	}

	// Set the content-type to application/json if not already available
	if !contentTypePresent {
		res.responseWriter.Header().
			Set("content-type", "application/json")
	}
	res.pendingStatusCode = statusCode
	res.pendingBodyBytes = bytes
	res.pendingBodyObj = body
}

// WriteJSON writes a json serializable struct to Response
func (res *ServerHTTPResponse) WriteJSON(
	statusCode int, headers Header, body interface{},
) {
	bytes := res.MarshalResponseJSON(body)
	if bytes == nil {
		return
	}
	res.SendResponse(statusCode, headers, body, bytes)
}

// PeekBody allows for inspecting a key path inside the body
// that is not flushed yet. This is useful for response middlewares
// that want to inspect the response body.
func (res *ServerHTTPResponse) PeekBody(
	keys ...string,
) ([]byte, jsonparser.ValueType, error) {
	value, valueType, _, err := jsonparser.Get(
		res.pendingBodyBytes, keys...,
	)
	if err != nil {
		return nil, -1, err
	}

	return value, valueType, nil
}

// Flush will write the body to the response. Before flush is called
// the body is pending. A pending body allows a response middleware to
// write a different body.
func (res *ServerHTTPResponse) flush(ctx context.Context) {
	if res.flushed {
		/* coverage ignore next line */
		res.contextLogger.Error(ctx,
			"Flushed a server response multiple times",
			zap.String("path", res.Request.URL.Path),
		)
		/* coverage ignore next line */
		return
	}

	res.flushed = true
	res.writeHeader(res.pendingStatusCode)
	if _, noContent := noContentStatusCodes[res.pendingStatusCode]; !noContent {
		res.writeBytes(res.pendingBodyBytes)
	}
	res.finish(ctx)
}

func (res *ServerHTTPResponse) writeHeader(statusCode int) {
	res.StatusCode = statusCode
	res.responseWriter.WriteHeader(statusCode)
}

// WriteBytes writes raw bytes to output
func (res *ServerHTTPResponse) writeBytes(bytes []byte) {
	_, err := res.responseWriter.Write(bytes)
	if err != nil {
		/* coverage ignore next line */
		res.contextLogger.Error(res.Request.Context(),
			"Could not write string to resp body",
			zap.Error(err),
			zap.String("bytesLength", strconv.Itoa(len(bytes))),
		)
	}
}

// GetPendingResponse lets you read the pending body bytes, obj and status code
// which isn't sent back yet.
func (res *ServerHTTPResponse) GetPendingResponse() ([]byte, int) {
	return res.pendingBodyBytes, res.pendingStatusCode
}

// GetPendingResponseObject lets you read the pending body object
// which isn't sent back yet.
func (res *ServerHTTPResponse) GetPendingResponseObject() interface{} {
	return res.pendingBodyObj
}

// Headers returns the underlying http response's headers
func (res *ServerHTTPResponse) Headers() http.Header {
	return res.responseWriter.Header()
}
