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
	"reflect"

	"time"

	"github.com/buger/jsonparser"
	"github.com/uber-go/zap"
)

// ServerHTTPResponse struct manages request
type ServerHTTPResponse struct {
	responseWriter    http.ResponseWriter
	req               *ServerHTTPRequest
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
		req:            req,
		responseWriter: w,
		StatusCode:     200,
		metrics:        req.metrics,
	}

	return res
}

// finish will handle final logic, like metrics
func (res *ServerHTTPResponse) finish() {
	if !res.req.started {
		/* coverage ignore next line */
		res.req.Logger.Error(
			"Forgot to start incoming request",
			zap.String("path", res.req.URL.Path),
		)
		/* coverage ignore next line */
		return
	}
	if res.finished {
		/* coverage ignore next line */
		res.req.Logger.Error(
			"Finished an incoming request twice",
			zap.String("path", res.req.URL.Path),
		)
		/* coverage ignore next line */
		return
	}

	res.finished = true
	res.finishTime = time.Now()

	counter := res.metrics.statusCodes[res.StatusCode]
	if counter == nil {
		res.req.Logger.Error(
			"Could not emit statusCode metric",
			zap.Int("UnexpectedStatusCode", res.StatusCode),
		)
	} else {
		counter.Inc(1)
	}

	res.metrics.requestLatency.Record(
		res.finishTime.Sub(res.req.startTime),
	)
}

// SendError helper to send an error
func (res *ServerHTTPResponse) SendError(statusCode int, err error) {
	res.SendErrorString(statusCode, err.Error())
}

// SendErrorString helper to send an error string
func (res *ServerHTTPResponse) SendErrorString(
	statusCode int, err string,
) {
	res.req.Logger.Warn(
		"Sending error for endpoint request",
		zap.String("error", err),
		zap.String("path", res.req.URL.Path),
	)

	// TODO: mark bodyBytes ...

	res.writeHeader(statusCode)
	res.writeString(err)

	res.finish()
}

// WriteJSONBytes writes a byte[] slice that is valid json to Response
func (res *ServerHTTPResponse) WriteJSONBytes(
	statusCode int, bytes []byte,
) {
	// TODO: mark header as pending ?
	res.responseWriter.Header().
		Set("content-type", "application/json")

	res.pendingStatusCode = statusCode
	res.pendingBodyBytes = bytes
}

// WriteJSON writes a json serializable struct to Response
func (res *ServerHTTPResponse) WriteJSON(
	statusCode int, body json.Marshaler,
) {
	if body == nil {
		res.SendErrorString(500, "Could not serialize json response")
		res.req.Logger.Error("Could not serialize nil pointer body")
		return
	}

	bytes, err := body.MarshalJSON()
	if err != nil {
		res.SendErrorString(500, "Could not serialize json response")
		res.req.Logger.Error("Could not serialize json response",
			zap.String("error", err.Error()),
		)
		return
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

// PeekBodyReflection allows for inspecting a key path inside the
// body that is not flushed yet. This is useful for response middlewares
// that want to inspect the response body.
func (res *ServerHTTPResponse) PeekBodyReflection(
	keys ...string,
) interface{} {
	obj := res.pendingBodyObj

	rptr := reflect.ValueOf(obj).Elem()

	for i := 0; i < len(keys); i++ {
		key := keys[i]

		field := rptr.FieldByName(key)
		rptr = field
	}

	return rptr.Interface()
}

// Flush will write the body to the response. Before flush is called
// the body is pending. A pending body allows a response middleware to
// write a different body.
func (res *ServerHTTPResponse) Flush() {
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
		res.req.Logger.Error("Could not write string to resp body",
			zap.String("error", err.Error()),
		)
	}
}

// WriteString helper just writes a string to the response
func (res *ServerHTTPResponse) writeString(text string) {
	res.writeBytes([]byte(text))
}

// IsOKResponse checks if the status code is OK.
func (res *ServerHTTPResponse) IsOKResponse(
	statusCode int, okResponses []int,
) bool {
	for _, r := range okResponses {
		if statusCode == r {
			return true
		}
	}
	return false
}
