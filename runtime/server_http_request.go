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

	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

// ServerHTTPRequest struct manages request
type ServerHTTPRequest struct {
	httpRequest *http.Request
	res         *ServerHTTPResponse
	gateway     *Gateway
	started     bool
	startTime   time.Time
	metrics     *EndpointMetrics
	queryValues url.Values
	parseFailed bool

	Logger *zap.Logger
	Scope  tally.Scope

	EndpointName string
	HandlerName  string
	URL          *url.URL
	Method       string
	Params       httprouter.Params
	Header       Header
}

// NewServerHTTPRequest is helper function to alloc ServerHTTPRequest
func NewServerHTTPRequest(
	w http.ResponseWriter, r *http.Request,
	params httprouter.Params, endpoint *RouterEndpoint,
) *ServerHTTPRequest {
	req := &ServerHTTPRequest{
		gateway:     endpoint.gateway,
		httpRequest: r,
		queryValues: nil,

		Logger: endpoint.gateway.Logger,
		Scope:  endpoint.gateway.MetricScope,

		URL:     r.URL,
		Method:  r.Method,
		Params:  params,
		Header:  NewServerHTTPHeader(r.Header),
		metrics: &endpoint.metrics,
	}
	req.res = NewServerHTTPResponse(w, req)

	req.start(endpoint.EndpointName, endpoint.HandlerName)

	return req
}

// start the request, do some metrics etc
func (req *ServerHTTPRequest) start(endpoint string, handler string) {
	if req.started {
		/* coverage ignore next line */
		req.Logger.Error(
			"Cannot start ServerHTTPRequest twice",
			zap.String("path", req.URL.Path),
		)
		/* coverage ignore next line */
		return
	}

	req.EndpointName = endpoint
	req.HandlerName = handler
	req.started = true
	req.startTime = time.Now()

	req.metrics.requestRecvd.Inc(1)
}

// CheckHeaders verifies that request contains required headers.
func (req *ServerHTTPRequest) CheckHeaders(headers []string) bool {
	for _, headerName := range headers {
		headerValue := req.httpRequest.Header.Get(headerName)
		if headerValue == "" {
			req.Logger.Warn("Got request without mandatory header",
				zap.String("headerName", headerName),
			)

			if !req.parseFailed {
				req.res.SendErrorString(
					400, "Missing mandatory header: "+headerName,
				)
				req.parseFailed = true
			}

			return false
		}

	}
	return true
}

func (req *ServerHTTPRequest) parseQueryValues() bool {
	if req.parseFailed {
		return false
	}

	if req.queryValues != nil {
		return true
	}

	values, err := url.ParseQuery(req.httpRequest.URL.RawQuery)
	if err != nil {
		req.Logger.Warn("Got request with invalid query string",
			zap.String("error", err.Error()),
		)

		if !req.parseFailed {
			req.res.SendErrorString(
				400, "Could not parse query string",
			)
			req.parseFailed = true
		}
		return false
	}

	req.queryValues = values
	return true
}

// GetQueryValue will return the first query parameter for key or empty string
func (req *ServerHTTPRequest) GetQueryValue(key string) (string, bool) {
	success := req.parseQueryValues()
	if !success {
		return "", false
	}

	return req.queryValues.Get(key), true
}

// GetQueryBool will return a query param as a boolean
func (req *ServerHTTPRequest) GetQueryBool(key string) (bool, bool) {
	success := req.parseQueryValues()
	if !success {
		return false, false
	}

	value := req.queryValues.Get(key)
	if value == "true" {
		return true, true
	} else if value == "false" {
		return false, true
	}

	err := &strconv.NumError{
		Func: "ParseBool",
		Num:  value,
		Err:  strconv.ErrSyntax,
	}

	req.Logger.Warn("Got request with invalid query string types",
		zap.String("expected", "bool"),
		zap.String("actual", value),
		zap.String("error", err.Error()),
	)
	if !req.parseFailed {
		req.res.SendErrorString(
			400, "Could not parse query string",
		)
		req.parseFailed = true
	}
	return false, false
}

// GetQueryInt8 will return a query params as int8
func (req *ServerHTTPRequest) GetQueryInt8(key string) (int8, bool) {
	success := req.parseQueryValues()
	if !success {
		return 0, false
	}

	value := req.queryValues.Get(key)
	number, err := strconv.ParseInt(value, 10, 8)
	if err != nil {
		req.Logger.Warn("Got request with invalid query string types",
			zap.String("expected", "int8"),
			zap.String("actual", value),
			zap.String("error", err.Error()),
		)
		if !req.parseFailed {
			req.res.SendErrorString(
				400, "Could not parse query string",
			)
			req.parseFailed = true
		}
		return 0, false
	}

	return int8(number), true
}

// GetQueryInt16 will return a query params as int16
func (req *ServerHTTPRequest) GetQueryInt16(key string) (int16, bool) {
	success := req.parseQueryValues()
	if !success {
		return 0, false
	}

	value := req.queryValues.Get(key)
	number, err := strconv.ParseInt(value, 10, 16)
	if err != nil {
		req.Logger.Warn("Got request with invalid query string types",
			zap.String("expected", "int16"),
			zap.String("actual", value),
			zap.String("error", err.Error()),
		)
		if !req.parseFailed {
			req.res.SendErrorString(
				400, "Could not parse query string",
			)
			req.parseFailed = true
		}
		return 0, false
	}

	return int16(number), true
}

// GetQueryInt32 will return a query params as int32
func (req *ServerHTTPRequest) GetQueryInt32(key string) (int32, bool) {
	success := req.parseQueryValues()
	if !success {
		return 0, false
	}

	value := req.queryValues.Get(key)
	number, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		req.Logger.Warn("Got request with invalid query string types",
			zap.String("expected", "int32"),
			zap.String("actual", value),
			zap.String("error", err.Error()),
		)
		if !req.parseFailed {
			req.res.SendErrorString(
				400, "Could not parse query string",
			)
			req.parseFailed = true
		}
		return 0, false
	}

	return int32(number), true
}

// GetQueryInt64 will return a query param as int64
func (req *ServerHTTPRequest) GetQueryInt64(key string) (int64, bool) {
	success := req.parseQueryValues()
	if !success {
		return 0, false
	}

	value := req.queryValues.Get(key)
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		req.Logger.Warn("Got request with invalid query string types",
			zap.String("expected", "int64"),
			zap.String("actual", value),
			zap.String("error", err.Error()),
		)
		if !req.parseFailed {
			req.res.SendErrorString(
				400, "Could not parse query string",
			)
			req.parseFailed = true
		}
		return 0, false
	}

	return number, true
}

// GetQueryFloat64 will return query param key as float64
func (req *ServerHTTPRequest) GetQueryFloat64(key string) (float64, bool) {
	success := req.parseQueryValues()
	if !success {
		return 0, false
	}

	value := req.queryValues.Get(key)
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		req.Logger.Warn("Got request with invalid query string types",
			zap.String("expected", "float64"),
			zap.String("actual", value),
			zap.String("error", err.Error()),
		)
		if !req.parseFailed {
			req.res.SendErrorString(
				400, "Could not parse query string",
			)
			req.parseFailed = true
		}
		return 0, false
	}

	return number, true
}

// GetQueryValues will return all query parameters for key.
func (req *ServerHTTPRequest) GetQueryValues(key string) ([]string, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	return req.queryValues[key], true
}

// ReadAndUnmarshalBody will try to unmarshal into struct or fail
func (req *ServerHTTPRequest) ReadAndUnmarshalBody(
	body json.Unmarshaler,
) bool {
	rawBody, success := req.ReadAll()
	if !success {
		return false
	}

	return req.UnmarshalBody(body, rawBody)
}

// ReadAll helper to read entire body
func (req *ServerHTTPRequest) ReadAll() ([]byte, bool) {
	rawBody, err := ioutil.ReadAll(req.httpRequest.Body)
	if err != nil {
		req.Logger.Error("Could not ReadAll() body",
			zap.String("error", err.Error()),
		)
		if !req.parseFailed {
			req.res.SendErrorString(500, "Could not ReadAll() body")
			req.parseFailed = true
		}

		return nil, false
	}

	return rawBody, true
}

// UnmarshalBody helper to unmarshal body into struct
func (req *ServerHTTPRequest) UnmarshalBody(
	body json.Unmarshaler, rawBody []byte,
) bool {
	err := body.UnmarshalJSON(rawBody)
	if err != nil {
		req.Logger.Warn("Could not parse json",
			zap.String("error", err.Error()),
		)
		if !req.parseFailed {
			req.res.SendErrorString(400, "Could not parse json: "+err.Error())
			req.parseFailed = true
		}
		return false
	}

	return true
}
