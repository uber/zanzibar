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

// OutgoingHTTPResponse struct manages request
type OutgoingHTTPResponse struct {
	responseWriter http.ResponseWriter
	req            *IncomingHTTPRequest
	gateway        *Gateway
	finishTime     time.Time
	finished       bool
	metrics        *EndpointMetrics

	StatusCode int
}

// NewOutgoingHTTPResponse is helper function to alloc OutgoingHTTPResponse
func NewOutgoingHTTPResponse(
	w http.ResponseWriter, req *IncomingHTTPRequest,
) *OutgoingHTTPResponse {
	res := &OutgoingHTTPResponse{
		gateway:        req.gateway,
		req:            req,
		responseWriter: w,
		StatusCode:     200,
		metrics:        req.metrics,
	}

	return res
}

// finish will handle final logic, like metrics
func (res *OutgoingHTTPResponse) finish() {
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
func (res *OutgoingHTTPResponse) SendError(statusCode int, err error) {
	res.SendErrorString(statusCode, err.Error())
}

// SendErrorString helper to send an error string
func (res *OutgoingHTTPResponse) SendErrorString(
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
func (res *OutgoingHTTPResponse) CopyJSON(statusCode int, src io.Reader) {
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
func (res *OutgoingHTTPResponse) WriteJSONBytes(
	statusCode int, bytes []byte,
) {
	res.responseWriter.Header().Set("content-type", "application/json")
	res.writeHeader(statusCode)
	res.writeBytes(bytes)

	res.finish()
}

// WriteJSON writes a json serializable struct to Response
func (res *OutgoingHTTPResponse) WriteJSON(
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
	res.writeBytes(newLineBytes)

	res.finish()
}

func (res *OutgoingHTTPResponse) writeHeader(statusCode int) {
	res.StatusCode = statusCode
	res.responseWriter.WriteHeader(statusCode)
}

// WriteBytes writes raw bytes to output
func (res *OutgoingHTTPResponse) writeBytes(bytes []byte) {
	_, err := res.responseWriter.Write(bytes)
	if err != nil {
		res.req.Logger.Error("Could not write string to resp body",
			zap.String("error", err.Error()),
		)
	}
}

// WriteHeader writes the header to http respnse.
func (res *OutgoingHTTPResponse) WriteHeader(statusCode int) {
	res.writeHeader(statusCode)
}

// WriteString helper just writes a string to the response
func (res *OutgoingHTTPResponse) writeString(text string) {
	res.writeBytes([]byte(text))
}

// NotFound helper to make request NotFound
func (res *OutgoingHTTPResponse) NotFound() {
	http.NotFound(res.responseWriter, res.req.httpRequest)
	// A NotFound request is not started...
	// TODO: inc.finish()
}

// IsOKResponse checks if the status code is OK.
func (res *OutgoingHTTPResponse) IsOKResponse(
	statusCode int, okResponses []int,
) bool {
	for _, r := range okResponses {
		if statusCode == r {
			return true
		}
	}
	return false
}
