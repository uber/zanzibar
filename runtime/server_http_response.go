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
)

// ServerHTTPResponse struct manages request
type ServerHTTPResponse struct {
	responseWriter    http.ResponseWriter
	Request           *ServerHTTPRequest
	gateway           *Gateway
	finishTime        time.Time
	finished          bool
	metrics           *EndpointMetrics
	flushed           bool
	pendingBodyBytes  []byte
	pendingBodyObj    interface{}
	pendingStatusCode int

	StatusCode int
}

// NewServerHTTPResponse is helper function to alloc ServerHTTPResponse
func NewServerHTTPResponse(
	w http.ResponseWriter, req *ServerHTTPRequest,
) *ServerHTTPResponse {
	res := &ServerHTTPResponse{
		gateway:        req.gateway,
		Request:        req,
		responseWriter: w,
		StatusCode:     200,
		metrics:        req.metrics,
	}

	return res
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

	counter := res.metrics.statusCodes[res.StatusCode]
	if counter == nil {
		res.Request.Logger.Error(
			"Could not emit statusCode metric",
			zap.Int("UnexpectedStatusCode", res.StatusCode),
		)
	} else {
		counter.Inc(1)
	}

	res.metrics.requestLatency.Record(
		res.finishTime.Sub(res.Request.startTime),
	)
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
	statusCode int, headers ServerHeader, bytes []byte,
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
	res.responseWriter.Header().
		Set("content-type", "application/json")

	res.pendingStatusCode = statusCode
	res.pendingBodyBytes = bytes
}

// WriteJSON writes a json serializable struct to Response
func (res *ServerHTTPResponse) WriteJSON(
	statusCode int, headers ServerHeader, body json.Marshaler,
) {
	if body == nil {
		res.SendErrorString(500, "Could not serialize json response")
		res.Request.Logger.Error("Could not serialize nil pointer body")
		return
	}

	bytes, err := body.MarshalJSON()
	if err != nil {
		res.SendErrorString(500, "Could not serialize json response")
		res.Request.Logger.Error("Could not serialize json response",
			zap.String("error", err.Error()),
		)
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
			zap.String("error", err.Error()),
		)
	}
}
