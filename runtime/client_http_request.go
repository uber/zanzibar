// Copyright (c) 2018 Uber Technologies, Inc.
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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ClientHTTPRequest is the struct for making client
// requests using an outbound http client.
type ClientHTTPRequest struct {
	ClientID   string
	MethodName string
	client     *HTTPClient
	httpReq    *http.Request
	res        *ClientHTTPResponse
	started    bool
	startTime  time.Time
	Logger     *zap.Logger
	metrics    *OutboundHTTPMetrics
	rawBody    []byte
}

// NewClientHTTPRequest allocates a ClientHTTPRequest
func NewClientHTTPRequest(
	clientID, methodName string,
	client *HTTPClient,
) *ClientHTTPRequest {
	req := &ClientHTTPRequest{
		ClientID:   clientID,
		MethodName: methodName,
		client:     client,
		Logger:     client.loggers[methodName],
		metrics:    client.metrics[methodName],
	}
	req.res = NewClientHTTPResponse(req)
	for headerKey, headerValue := range client.Headers {
		req.httpReq.Header.Add(headerKey, headerValue)
	}
	req.start()
	return req
}

// Start the request, do some metrics book keeping
func (req *ClientHTTPRequest) start() {
	if req.started {
		/* coverage ignore next line */
		req.Logger.Error("Cannot start ClientHTTPRequest twice")
		/* coverage ignore next line */
		return
	}
	req.started = true
	req.startTime = time.Now()
}

// CheckHeaders verifies that the outbound request contaisn required headers
func (req *ClientHTTPRequest) CheckHeaders(expected []string) error {
	if req.httpReq == nil {
		/* coverage ignore next line */
		panic("must call `req.WriteJSON()` before `req.CheckHeaders()`")
	}

	actualHeaders := req.httpReq.Header

	for _, headerName := range expected {
		// TODO: case sensitivity ?
		headerValue := actualHeaders.Get(headerName)
		if headerValue == "" {
			req.Logger.Warn("Got outbound request without mandatory header",
				zap.String("headerName", headerName),
			)

			return errors.New("Missing mandatory header: " + headerName)
		}
	}

	return nil
}

// WriteJSON will send a json http request out.
func (req *ClientHTTPRequest) WriteJSON(
	method, url string,
	headers map[string]string,
	body json.Marshaler,
) error {
	var httpReq *http.Request
	var httpErr error
	if body != nil {
		rawBody, err := body.MarshalJSON()
		if err != nil {
			req.Logger.Error("Could not serialize request json", zap.Error(err))
			return errors.Wrapf(
				err, "Could not serialize %s.%s request json",
				req.ClientID, req.MethodName,
			)
		}
		req.rawBody = rawBody
		httpReq, httpErr = http.NewRequest(method, url, bytes.NewReader(rawBody))
	} else {
		httpReq, httpErr = http.NewRequest(method, url, nil)
	}

	if httpErr != nil {
		req.Logger.Error("Could not create outbound request", zap.Error(httpErr))
		return errors.Wrapf(
			httpErr, "Could not create outbound %s.%s request",
			req.ClientID, req.MethodName,
		)
	}

	for k := range headers {
		httpReq.Header.Add(k, headers[k])
	}
	httpReq.Header.Set("Content-Type", "application/json")

	req.httpReq = httpReq
	return nil
}

// Do will send the request out.
func (req *ClientHTTPRequest) Do(
	ctx context.Context,
) (*ClientHTTPResponse, error) {
	res, err := req.client.Client.Do(req.httpReq.WithContext(ctx))
	if err != nil {
		req.Logger.Error("Could not make outbound request", zap.Error(err))
		return nil, err
	}

	// emit metrics
	req.metrics.Sent.Inc(1)

	req.res.setRawHTTPResponse(res)
	return req.res, nil
}
