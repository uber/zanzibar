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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ClientHTTPResponse is the struct managing the client response
// when making outbound http calls.
type ClientHTTPResponse struct {
	req              *ClientHTTPRequest
	finishTime       time.Time
	finished         bool
	rawResponse      *http.Response
	rawResponseBytes []byte
	StatusCode       int
	Header           http.Header
}

// NewClientHTTPResponse allocates a client http response object
// to track http response.
func NewClientHTTPResponse(
	req *ClientHTTPRequest,
) *ClientHTTPResponse {
	return &ClientHTTPResponse{
		req: req,
	}
}

func (res *ClientHTTPResponse) setRawHTTPResponse(httpRes *http.Response) {
	res.rawResponse = httpRes
	res.StatusCode = httpRes.StatusCode
	res.Header = httpRes.Header
}

// ReadAll reads bytes from response.
func (res *ClientHTTPResponse) ReadAll() ([]byte, error) {
	rawBody, err := ioutil.ReadAll(res.rawResponse.Body)

	cerr := res.rawResponse.Body.Close()
	if cerr != nil {
		/* coverage ignore next line */
		res.req.Logger.Error("Could not close response body", zap.Error(cerr))
	}

	if err != nil {
		res.req.Logger.Error("Could not read response body", zap.Error(err))
		res.finish()
		return nil, errors.Wrapf(
			err, "Could not read %s.%s response body",
			res.req.ClientID, res.req.MethodName,
		)
	}

	res.rawResponseBytes = rawBody
	res.finish()
	return rawBody, nil
}

// GetRawBody returns the body as byte array if it has been read.
func (res *ClientHTTPResponse) GetRawBody() []byte {
	return res.rawResponseBytes
}

// ReadAndUnmarshalBody will try to unmarshal non pointer value or fail
func (res *ClientHTTPResponse) ReadAndUnmarshalBody(v interface{}) error {
	rawBody, err := res.ReadAll()
	if err != nil {
		/* coverage ignore next line */
		return err
	}

	err = json.Unmarshal(rawBody, v)
	if err != nil {
		res.req.Logger.Warn("Could not parse response json", zap.Error(err))
		return errors.Wrapf(
			err, "Could not parse %s.%s response json",
			res.req.ClientID, res.req.MethodName,
		)
	}

	return nil
}

// CheckOKResponse checks if the status code is OK.
func (res *ClientHTTPResponse) CheckOKResponse(okResponses []int) {
	for _, okResponse := range okResponses {
		if res.rawResponse.StatusCode == okResponse {
			return
		}
	}

	res.req.Logger.Warn("Unknown response status code",
		zap.Int("UnknownStatusCode", res.rawResponse.StatusCode),
	)
}

// finish will handle final logic, like metrics
func (res *ClientHTTPResponse) finish() {
	if !res.req.started {
		/* coverage ignore next line */
		res.req.Logger.Error("Forgot to start client request")
		/* coverage ignore next line */
		return
	}
	if res.finished {
		/* coverage ignore next line */
		res.req.Logger.Error("Finished a client response twice")
		/* coverage ignore next line */
		return
	}
	res.finished = true
	res.finishTime = time.Now()

	// emit metrics
	res.req.client.ContextMetrics.RecordTimer(res.req.ctx, clientLatency, res.finishTime.Sub(res.req.startTime))
	_, known := knownStatusCodes[res.StatusCode]
	if !known {
		res.req.Logger.Error(
			"Could not emit statusCode metric",
			zap.Int("UnknownStatusCode", res.StatusCode),
		)
	} else {
		scopeTags := map[string]string{scopeTagStatus: fmt.Sprintf("%d", res.StatusCode)}
		res.req.ctx = WithScopeTags(res.req.ctx, scopeTags)
		res.req.client.ContextMetrics.IncCounter(res.req.ctx, clientStatus, 1)
	}
	if !known || res.StatusCode >= 400 && res.StatusCode < 600 {
		res.req.client.ContextMetrics.IncCounter(res.req.ctx, clientErrors, 1)
	}

	// write logs
	res.req.ContextLogger.Info(
		res.req.ctx,
		"Finished an outgoing client HTTP request",
		clientHTTPLogFields(res.req, res)...,
	)
}

func clientHTTPLogFields(req *ClientHTTPRequest, res *ClientHTTPResponse) []zapcore.Field {
	fields := []zapcore.Field{
		zap.Time(logFieldRequestFinishedTime, res.finishTime),
		zap.Int(logFieldResponseStatusCode, res.StatusCode),
	}

	for k, v := range req.res.Header {
		if len(v) > 0 {
			fields = append(fields, zap.String("Response-Header-"+k, v[0]))
		}
	}

	// TODO: log jaeger trace span

	return fields
}
