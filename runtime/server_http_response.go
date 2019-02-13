// Copyright (c) 2019 Uber Technologies, Inc.
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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"github.com/uber-go/tally"
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
	}
}

// finish will handle final logic, like metrics
func (res *ServerHTTPResponse) finish() {
	if !res.Request.started {
		/* coverage ignore next line */
		res.logger.Error(
			"Forgot to start server response",
			zap.String("path", res.Request.URL.Path),
		)
		/* coverage ignore next line */
		return
	}
	if res.finished {
		/* coverage ignore next line */
		res.logger.Error(
			"Finished an server response twice",
			zap.String("path", res.Request.URL.Path),
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
	tagged.Timer(endpointLatency).Record(res.finishTime.Sub(res.Request.startTime))
	if !known {
		res.logger.Error(
			"Could not emit statusCode metric",
			zap.Int("UnknownStatusCode", res.StatusCode),
		)
	} else {
		tagged.Counter(endpointStatus).Inc(1)
	}
	if !known || res.StatusCode >= 400 && res.StatusCode < 600 {
		tagged.Counter(endpointAppErrors).Inc(1)

	}

	span := res.Request.GetSpan()
	if span != nil {
		span.Finish()
	}

	res.logger.Info(
		"Finished an incoming server HTTP request",
		serverHTTPLogFields(res.Request, res)...,
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

	if res.err != nil {
		fields = append(fields, zap.Error(res.err))

		cause := errors.Cause(res.err)
		if cause != nil && cause != res.err {
			fields = append(fields, zap.NamedError("errorCause", cause))
		}
	}

	for k, v := range req.httpRequest.Header {
		if len(v) > 0 {
			fields = append(fields, zap.String("Request-Header-"+k, v[0]))
		}
	}
	for k, v := range req.res.responseWriter.Header() {
		if len(v) > 0 {
			fields = append(fields, zap.String("Response-Header-"+k, v[0]))
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

// WriteJSONBytes writes a byte[] slice that is valid json to Response
func (res *ServerHTTPResponse) WriteJSONBytes(
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

	res.responseWriter.Header().Set("content-type", "application/json")

	res.pendingStatusCode = statusCode
	res.pendingBodyBytes = bytes
}

// WriteJSON writes a json serializable struct to Response
func (res *ServerHTTPResponse) WriteJSON(
	statusCode int, headers Header, body json.Marshaler,
) {
	if body == nil {
		res.SendError(500, "Could not serialize json response", errors.New("No Body JSON"))
		res.logger.Error("Could not serialize nil pointer body")
		return
	}

	bytes, err := body.MarshalJSON()
	if err != nil {
		res.SendError(500, "Could not serialize json response", err)
		res.logger.Error("Could not serialize json response", zap.Error(err))
		return
	}

	if headers != nil {
		for _, k := range headers.Keys() {
			v, ok := headers.Get(k)
			if ok {
				res.responseWriter.Header().Set(k, v)
			}
		}
	}

	res.responseWriter.Header().
		Set("content-type", "application/json")

	res.pendingStatusCode = statusCode
	res.pendingBodyBytes = bytes
	res.pendingBodyObj = body
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
func (res *ServerHTTPResponse) flush() {
	if res.flushed {
		/* coverage ignore next line */
		res.logger.Error(
			"Flushed a server response twice",
			zap.String("path", res.Request.URL.Path),
		)
		/* coverage ignore next line */
		return
	}

	res.flushed = true
	res.writeHeader(res.pendingStatusCode)
	res.writeBytes(res.pendingBodyBytes)
	res.finish()
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

// ResponseHeaders returns the underlying http response's headers
func (res *ServerHTTPResponse) ResponseHeaders() http.Header {
	return res.responseWriter.Header()
}
