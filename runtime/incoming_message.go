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
	"io/ioutil"
	"net/http"
	"net/url"

	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/uber-go/zap"
)

var newLineBytes = []byte("\n")

// IncomingMessage struct manages request/response
type IncomingMessage struct {
	responseWriter http.ResponseWriter
	httpRequest    *http.Request
	gateway        *Gateway
	started        bool
	startTime      time.Time
	finishTime     time.Time
	finished       bool
	metrics        *EndpointMetrics

	EndpointName string
	HandlerName  string
	StatusCode   int
	URL          *url.URL
	Method       string
	Params       httprouter.Params
	Header       http.Header
}

// NewIncomingMessage is helper function to alloc IncomingMessage
func NewIncomingMessage(
	w http.ResponseWriter, r *http.Request,
	params httprouter.Params, endpoint *Endpoint,
) *IncomingMessage {
	inc := &IncomingMessage{
		gateway:        endpoint.gateway,
		responseWriter: w,
		httpRequest:    r,
		URL:            r.URL,
		StatusCode:     200,
		Method:         r.Method,
		Params:         params,
		Header:         r.Header,
		metrics:        &endpoint.metrics,
	}

	inc.Start(endpoint.EndpointName, endpoint.HandlerName)

	return inc
}

// finish will handle final logic, like metrics
func (inc *IncomingMessage) finish() {
	if !inc.started {
		inc.gateway.Logger.Error(
			"Forgot to start incoming request",
			zap.String("path", inc.URL.Path),
		)
		return
	}
	if inc.finished {
		inc.gateway.Logger.Error(
			"Finished an incoming request twice",
			zap.String("path", inc.URL.Path),
		)
		return
	}

	inc.finished = true
	inc.finishTime = time.Now()

	counter := inc.metrics.statusCodes[inc.StatusCode]
	if counter == nil {
		inc.gateway.Logger.Error(
			"Could not emit statusCode metric",
			zap.Int("UnexpectedStatusCode", inc.StatusCode),
		)
	} else {
		counter.Inc(1)
	}

	inc.metrics.requestLatency.Record(inc.finishTime.Sub(inc.startTime))
}

// Start the request, do some metrics etc
func (inc *IncomingMessage) Start(endpoint string, handler string) {
	if inc.started {
		inc.gateway.Logger.Error(
			"Cannot start IncomingMessage twice",
			zap.String("path", inc.URL.Path),
		)
		return
	}

	inc.EndpointName = endpoint
	inc.HandlerName = handler
	inc.started = true
	inc.startTime = time.Now()

	inc.metrics.requestRecvd.Inc(1)
}

// SendError helper to send an error
func (inc *IncomingMessage) SendError(statusCode int, err error) {
	inc.SendErrorString(statusCode, err.Error())
}

// SendErrorString helper to send an error string
func (inc *IncomingMessage) SendErrorString(statusCode int, err string) {
	inc.gateway.Logger.Warn(
		"Sending error for endpoint request",
		zap.String("error", err),
		zap.String("path", inc.URL.Path),
	)

	inc.writeHeader(statusCode)
	inc.writeString(err)

	inc.finish()
}

// CopyJSON will copy json bytes from a Reader
func (inc *IncomingMessage) CopyJSON(statusCode int, src io.Reader) {
	inc.responseWriter.Header().Set("content-type", "application/json")
	inc.writeHeader(statusCode)
	_, err := io.Copy(inc.responseWriter, src)
	if err != nil {
		inc.gateway.Logger.Error("Could not copy bytes",
			zap.String("error", err.Error()),
		)
	}

	inc.finish()
}

// WriteJSONBytes writes a byte[] slice that is valid json to Response
func (inc *IncomingMessage) WriteJSONBytes(statusCode int, bytes []byte) {
	inc.responseWriter.Header().Set("content-type", "application/json")
	inc.writeHeader(statusCode)
	inc.writeBytes(bytes)

	inc.finish()
}

// WriteJSON writes a json serializable struct to Response
func (inc *IncomingMessage) WriteJSON(statusCode int, body json.Marshaler) {
	bytes, err := body.MarshalJSON()
	if err != nil {
		inc.SendErrorString(500, "Could not serialize json response")
		inc.gateway.Logger.Error("Could not serialize json response",
			zap.String("error", err.Error()),
		)
		return
	}

	inc.responseWriter.Header().Set("content-type", "application/json")
	inc.writeHeader(statusCode)
	inc.writeBytes(bytes)
	inc.writeBytes(newLineBytes)

	inc.finish()
}

func (inc *IncomingMessage) writeHeader(statusCode int) {
	inc.StatusCode = statusCode
	inc.responseWriter.WriteHeader(statusCode)
}

// WriteBytes writes raw bytes to output
func (inc *IncomingMessage) writeBytes(bytes []byte) {
	_, err := inc.responseWriter.Write(bytes)
	if err != nil {
		inc.gateway.Logger.Error("Could not write string to resp body",
			zap.String("error", err.Error()),
		)
	}
}

// WriteHeader writes the header to http respnse.
func (inc *IncomingMessage) WriteHeader(statusCode int) {
	inc.writeHeader(statusCode)
}

// WriteString helper just writes a string to the response
func (inc *IncomingMessage) writeString(text string) {
	inc.writeBytes([]byte(text))
}

// NotFound helper to make request NotFound
func (inc *IncomingMessage) NotFound() {
	http.NotFound(inc.responseWriter, inc.httpRequest)
	// A NotFound request is not started...
	// TODO: inc.finish()
}

// ReadAll helper to read entire body
func (inc *IncomingMessage) ReadAll() ([]byte, bool) {
	rawBody, err := ioutil.ReadAll(inc.httpRequest.Body)
	if err != nil {
		inc.SendErrorString(500, "Could not ReadAll() body")
		inc.gateway.Logger.Error("Could not ReadAll() body",
			zap.String("error", err.Error()),
		)
		return nil, false
	}

	return rawBody, true
}

// UnmarshalBody helper to unmarshal body into struct
func (inc *IncomingMessage) UnmarshalBody(
	body json.Unmarshaler, rawBody []byte,
) bool {
	err := body.UnmarshalJSON(rawBody)
	if err != nil {
		inc.SendErrorString(400, "Could not parse json: "+err.Error())
		inc.gateway.Logger.Warn("Could not parse json",
			zap.String("error", err.Error()),
		)
		return false
	}

	return true
}

// IsOKResponse checks if the status code is OK.
func (inc *IncomingMessage) IsOKResponse(statusCode int, okResponses []int) bool {
	for _, r := range okResponses {
		if statusCode == r {
			return true
		}
	}
	return false
}
