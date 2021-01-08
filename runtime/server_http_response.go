// Copyright (c) 2020 Uber Technologies, Inc.
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
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"github.com/uber/zanzibar/runtime/jsonwrapper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ServerHTTPResponse struct manages server http response
type ServerHTTPResponse struct {
	Request    *ServerHTTPRequest
	StatusCode int

	responseWriter    http.ResponseWriter
	flushed           bool
	finished          bool
	finishTime        time.Time
	pendingBodyBytes  []byte
	pendingBodyObj    interface{}
	pendingStatusCode int
	logger            Logger
	scope             tally.Scope
	jsonWrapper       jsonwrapper.JSONWrapper
	err               error
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
		logger:         req.logger,
		scope:          req.scope,
		jsonWrapper:    req.jsonWrapper,
	}
}

// finish will handle final logic, like metrics
func (res *ServerHTTPResponse) finish(ctx context.Context) {
	logFields := GetLogFieldsFromCtx(ctx)
	if !res.Request.started {
		/* coverage ignore next line */
		res.logger.Error(
			"Forgot to start server response",
			append(logFields, zap.String("path", res.Request.URL.Path))...,
		)
		/* coverage ignore next line */
		return
	}
	if res.finished {
		/* coverage ignore next line */
		res.logger.Error(
			"Finished a server response multiple times",
			append(logFields, zap.String("path", res.Request.URL.Path))...,
		)
		/* coverage ignore next line */
		return
	}
	res.finished = true
	res.finishTime = time.Now()

	_, known := knownStatusCodes[res.StatusCode]
	// no need to put this tag on the context because this is the end of response life cycle
	statusTag := map[string]string{scopeTagStatus: fmt.Sprintf("%d", res.StatusCode)}
	tagged := res.scope.Tagged(statusTag)
	delta := res.finishTime.Sub(res.Request.startTime)
	tagged.Timer(endpointLatency).Record(delta)
	tagged.Histogram(endpointLatencyHist, tally.DefaultBuckets).RecordDuration(delta)
	if !known {
		res.logger.Error(
			"Unknown status code",
			append(logFields, zap.Int("UnknownStatusCode", res.StatusCode))...,
		)
	} else {
		tagged.Counter(endpointStatus).Inc(1)
	}

	logFn := res.logger.Debug
	if !known || res.StatusCode >= 400 && res.StatusCode < 600 {
		tagged.Counter(endpointAppErrors).Inc(1)
		logFn = res.logger.Warn
	}

	span := res.Request.GetSpan()
	if span != nil {
		span.Finish()
	}

	logFn(
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

	for k, v := range res.Headers() {
		if len(v) > 0 {
			fields = append(fields, zap.String(
				fmt.Sprintf("%s-%s", logFieldEndpointResponseHeaderPrefix, k),
				strings.Join(v, ", "),
			))
		}
	}

	if res.err != nil {
		fields = append(fields, zap.Error(res.err))

		cause := errors.Cause(res.err)
		if cause != nil && cause != res.err {
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
	res.err = errCause
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
	if body == nil {
		res.SendError(500, "Could not serialize json response", errors.New("No Body JSON"))
		res.logger.Error("Could not serialize nil pointer body")
		return nil
	}
	bytes, err := res.jsonWrapper.Marshal(body)
	if err != nil {
		res.SendError(500, "Could not serialize json response", err)
		res.logger.Error("Could not serialize json response", zap.Error(err))
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
		res.logger.Error(
			"Flushed a server response multiple times",
			zap.String("path", res.Request.URL.Path),
		)
		/* coverage ignore next line */
		return
	}

	res.flushed = true
	res.writeHeader(res.pendingStatusCode)
	res.writeBytes(res.pendingBodyBytes)
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
		res.logger.Error(
			"Could not write string to resp body",
			zap.Error(err),
		)
	}
}

// GetPendingResponse lets you read the pending body bytes, obj and status code
// which isn't sent back yet.
func (res *ServerHTTPResponse) GetPendingResponse() ([]byte, int) {
	return res.pendingBodyBytes, res.pendingStatusCode
}

// GetPendingResponse lets you read the pending body object
// which isn't sent back yet.
func (res *ServerHTTPResponse) GetPendingResponseObject() interface{} {
	return res.pendingBodyObj
}

// Headers returns the underlying http response's headers
func (res *ServerHTTPResponse) Headers() http.Header {
	return res.responseWriter.Header()
}
