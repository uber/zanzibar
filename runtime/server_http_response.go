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
	"io"
	"net/http"

	"time"

	"github.com/uber-go/zap"
)

// ServerHTTPResponse struct manages request
type ServerHTTPResponse struct {
	responseWriter http.ResponseWriter
	req            *ServerHTTPRequest
	gateway        *Gateway
	finishTime     time.Time
	finished       bool
	metrics        *EndpointMetrics

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
		res.req.Logger.Error(
			"Forgot to start incoming request",
			zap.String("path", res.req.URL.Path),
		)
		return
	}
	if res.finished {
		res.req.Logger.Error(
			"Finished an incoming request twice",
			zap.String("path", res.req.URL.Path),
		)
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

	res.writeHeader(statusCode)
	res.writeString(err)

	res.finish()
}

// CopyJSON will copy json bytes from a Reader
func (res *ServerHTTPResponse) CopyJSON(statusCode int, src io.Reader) {
	res.responseWriter.Header().Set("content-type", "application/json")
	res.writeHeader(statusCode)
	_, err := io.Copy(res.responseWriter, src)
	if err != nil {
		res.req.Logger.Error("Could not copy bytes",
			zap.String("error", err.Error()),
		)
	}

	res.finish()
}

// WriteJSONBytes writes a byte[] slice that is valid json to Response
func (res *ServerHTTPResponse) WriteJSONBytes(
	statusCode int, bytes []byte,
) {
	res.responseWriter.Header().Set("content-type", "application/json")
	res.writeHeader(statusCode)
	res.writeBytes(bytes)

	res.finish()
}

// WriteJSON writes a json serializable struct to Response
func (res *ServerHTTPResponse) WriteJSON(
	statusCode int, body json.Marshaler,
) {
	bytes, err := body.MarshalJSON()
	if err != nil {
		res.SendErrorString(500, "Could not serialize json response")
		res.req.Logger.Error("Could not serialize json response",
			zap.String("error", err.Error()),
		)
		return
	}

	res.responseWriter.Header().Set("content-type", "application/json")
	res.writeHeader(statusCode)
	res.writeBytes(bytes)

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
		res.req.Logger.Error("Could not write string to resp body",
			zap.String("error", err.Error()),
		)
	}
}

// WriteHeader writes the header to http respnse.
func (res *ServerHTTPResponse) WriteHeader(statusCode int) {
	res.writeHeader(statusCode)
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
