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
	started     bool
	startTime   time.Time
	client      *HTTPClient
	httpRequest *http.Request
	res         *ClientHTTPResponse

	ClientName string
	MethodName string
	Logger     *zap.Logger
}

// NewClientHTTPRequest allocates a ClientHTTPRequest
func NewClientHTTPRequest(
	clientName string, methodName string,
	client *HTTPClient,
) *ClientHTTPRequest {
	req := &ClientHTTPRequest{
		Logger: client.Logger,
		client: client,
	}

	req.res = NewClientHTTPResponse(req)

	req.start(clientName, methodName)
	return req
}

// Start the request, do some metrics book keeping
func (req *ClientHTTPRequest) start(
	clientName string, methodName string,
) {
	if req.started {
		/* coverage ignore next line */
		req.Logger.Error(
			"Cannot start ClientHTTPRequest twice",
			zap.String("methodName", methodName),
			zap.String("clientName", clientName),
		)
		/* coverage ignore next line */
		return
	}

	req.ClientName = clientName
	req.MethodName = methodName

	req.started = true
	req.startTime = time.Now()
}

// WriteJSON will send a json http request out.
func (req *ClientHTTPRequest) WriteJSON(
	method string, url string, headers map[string]string, body json.Marshaler,
) error {
	var httpReq *http.Request
	var httpErr error
	if body != nil {
		rawBody, err := body.MarshalJSON()
		if err != nil {
			req.Logger.Error("Could not serialize client json request",
				zap.String("error", err.Error()),
			)
			return errors.Wrapf(err,
				"Could not serialize json for client: %s", req.ClientName,
			)
		}

		httpReq, httpErr = http.NewRequest(
			method, url, bytes.NewReader(rawBody),
		)
	} else {
		httpReq, httpErr = http.NewRequest(method, url, nil)
	}

	if httpErr != nil {
		req.Logger.Error("Could not make outbound request",
			zap.String("error", httpErr.Error()),
		)
		return errors.Wrapf(httpErr,
			"Could not make outbound request for client: %s",
			req.ClientName,
		)
	}

	for k := range headers {
		httpReq.Header.Add(k, headers[k])
	}

	req.httpRequest = httpReq
	req.httpRequest.Header.Set("Content-Type", "application/json")
	return nil
}

// Do will send the request out.
func (req *ClientHTTPRequest) Do(
	ctx context.Context,
) (*ClientHTTPResponse, error) {
	res, err := req.client.Client.Do(
		req.httpRequest.WithContext(ctx),
	)

	if err != nil {
		return nil, err
	}

	req.res.setRawHTTPResponse(res)
	return req.res, nil
}
