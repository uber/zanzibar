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
	"io/ioutil"
	"net/http"
	"net/url"

	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/uber-go/zap"
)

var newLineBytes = []byte("\n")

// IncomingHTTPRequest struct manages request
type IncomingHTTPRequest struct {
	httpRequest *http.Request
	res         *OutgoingHTTPResponse
	gateway     *Gateway
	started     bool
	startTime   time.Time
	metrics     *EndpointMetrics

	EndpointName string
	HandlerName  string
	URL          *url.URL
	Method       string
	Params       httprouter.Params
	Header       http.Header
}

// NewIncomingHTTPRequest is helper function to alloc IncomingHTTPRequest
func NewIncomingHTTPRequest(
	w http.ResponseWriter, r *http.Request,
	params httprouter.Params, endpoint *Endpoint,
) *IncomingHTTPRequest {
	req := &IncomingHTTPRequest{
		gateway:     endpoint.gateway,
		httpRequest: r,
		URL:         r.URL,
		Method:      r.Method,
		Params:      params,
		Header:      r.Header,
		metrics:     &endpoint.metrics,
	}
	req.res = NewOutgoingHTTPResponse(w, req)

	req.Start(endpoint.EndpointName, endpoint.HandlerName)

	return req
}

// Start the request, do some metrics etc
func (req *IncomingHTTPRequest) Start(endpoint string, handler string) {
	if req.started {
		req.gateway.Logger.Error(
			"Cannot start IncomingHTTPRequest twice",
			zap.String("path", req.URL.Path),
		)
		return
	}

	req.EndpointName = endpoint
	req.HandlerName = handler
	req.started = true
	req.startTime = time.Now()

	req.metrics.requestRecvd.Inc(1)
}

// ReadAndUnmarshalBody will try to unmarshal into struct or fail
func (req *IncomingHTTPRequest) ReadAndUnmarshalBody(
	body json.Unmarshaler,
) bool {
	rawBody, success := req.ReadAll()
	if !success {
		return false
	}

	return req.UnmarshalBody(body, rawBody)
}

// ReadAll helper to read entire body
func (req *IncomingHTTPRequest) ReadAll() ([]byte, bool) {
	rawBody, err := ioutil.ReadAll(req.httpRequest.Body)
	if err != nil {
		req.res.SendErrorString(500, "Could not ReadAll() body")
		req.gateway.Logger.Error("Could not ReadAll() body",
			zap.String("error", err.Error()),
		)
		return nil, false
	}

	return rawBody, true
}

// UnmarshalBody helper to unmarshal body into struct
func (req *IncomingHTTPRequest) UnmarshalBody(
	body json.Unmarshaler, rawBody []byte,
) bool {
	err := body.UnmarshalJSON(rawBody)
	if err != nil {
		req.res.SendErrorString(400, "Could not parse json: "+err.Error())
		req.gateway.Logger.Warn("Could not parse json",
			zap.String("error", err.Error()),
		)
		return false
	}

	return true
}
