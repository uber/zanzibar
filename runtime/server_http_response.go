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
	"encoding/json"
	"net/http"
	"time"

	"github.com/buger/jsonparser"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ServerHTTPResponse struct manages request
type ServerHTTPResponse struct {
	responseWriter    http.ResponseWriter
	Request           *ServerHTTPRequest
	flushed           bool
	finished          bool
	finishTime        time.Time
	pendingBodyBytes  []byte
	pendingBodyObj    interface{}
	pendingStatusCode int
	StatusCode        int
	logger            *zap.Logger
}

// NewServerHTTPResponse is helper function to alloc ServerHTTPResponse
func NewServerHTTPResponse(
	w http.ResponseWriter,
	req *ServerHTTPRequest,
) *ServerHTTPResponse {
	return &ServerHTTPResponse{
		Request:        req,
		responseWriter: w,
		StatusCode:     200,
	}
}

// finish will handle final logic, like metrics
func (res *ServerHTTPResponse) finish() {
	if !res.Request.started {
		/* coverage ignore next line */
		res.Request.Logger.Error(
			"Forgot to start server response",
			zap.String("path", res.Request.URL.Path),
		)
		/* coverage ignore next line */
		return
	}
	if res.finished {
		/* coverage ignore next line */
		res.Request.Logger.Error(
			"Finished an server response twice",
			zap.String("path", res.Request.URL.Path),
		)
		/* coverage ignore next line */
		return
	}
	res.finished = true
	res.finishTime = time.Now()

	// emit metrics
	res.Request.metrics.Latency.Record(res.finishTime.Sub(res.Request.startTime))
	_, known := knownStatusCodes[res.StatusCode]
	if !known {
		res.Request.Logger.Error(
			"Could not emit statusCode metric",
			zap.Int("UnknownStatusCode", res.StatusCode),
		)
	} else {
		res.Request.metrics.Status[res.StatusCode].Inc(1)
	}
	if !known || res.StatusCode >= 400 && res.StatusCode < 600 {
		res.Request.metrics.Errors.Inc(1)
	} else {
		res.Request.metrics.Success.Inc(1)
	}

	// write logs
	res.Request.Logger.Debug(
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

		// TODO: Do not log body by default because PII and bandwidth.
		// Temporarily log during the development cycle
		// TODO: Add a gateway level configurable body unmarshaller
		// to extract only non-PII info.
		zap.ByteString("Request Body", req.RawBody),
		zap.ByteString("Response Body", res.pendingBodyBytes),
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

	// TODO: log jaeger trace span

	return fields
}

// SendErrorString helper to send an error string
func (res *ServerHTTPResponse) SendErrorString(
	statusCode int, err string,
) {
	res.Request.Logger.Warn(
		"Sending error for endpoint request",
		zap.String("error", err),
		zap.String("path", res.Request.URL.Path),
	)

	res.WriteJSONBytes(statusCode, nil,
		[]byte(`{"error":"`+err+`"}`),
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

	// TODO: mark header as pending ?
	res.responseWriter.Header().Set("content-type", "application/json")

	res.pendingStatusCode = statusCode
	res.pendingBodyBytes = bytes
}

// WriteJSON writes a json serializable struct to Response
func (res *ServerHTTPResponse) WriteJSON(
	statusCode int, headers Header, body json.Marshaler,
) {
	if body == nil {
		res.SendErrorString(500, "Could not serialize json response")
		res.Request.Logger.Error("Could not serialize nil pointer body")
		return
	}

	bytes, err := body.MarshalJSON()
	if err != nil {
		res.SendErrorString(500, "Could not serialize json response")
		res.Request.Logger.Error("Could not serialize json response", zap.Error(err))
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

	// TODO: mark header as pending ?
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
		res.Request.Logger.Error(
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
		res.Request.Logger.Error("Could not write string to resp body",
			zap.Error(err),
		)
	}
}
