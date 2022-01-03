// Copyright (c) 2022 Uber Technologies, Inc.
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
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber-go/tally"
	"github.com/uber/zanzibar/runtime/jsonwrapper"
	"go.uber.org/zap"
)

// ServerHTTPRequest struct manages server http request
type ServerHTTPRequest struct {
	httpRequest *http.Request
	res         *ServerHTTPResponse
	startTime   time.Time
	started     bool
	tracer      opentracing.Tracer
	span        opentracing.Span
	queryValues url.Values
	parseFailed bool
	rawBody     []byte

	EndpointName string
	HandlerName  string
	URL          *url.URL
	Method       string
	Params       url.Values
	Header       Header

	// logger logs entries with default fields that contains request meta info
	contextLogger ContextLogger
	// scope emit metrics with default tags that contains request meta info
	scope       tally.Scope
	jsonWrapper jsonwrapper.JSONWrapper
}

// NewServerHTTPRequest is helper function to alloc ServerHTTPRequest
func NewServerHTTPRequest(
	w http.ResponseWriter,
	r *http.Request,
	params url.Values,
	endpoint *RouterEndpoint,
) *ServerHTTPRequest {
	ctx := r.Context()

	// put request log fields on context
	logFields := []zap.Field{
		zap.String(logFieldEndpointID, endpoint.EndpointName),
		zap.String(logFieldEndpointHandler, endpoint.HandlerName),
		zap.String(logFieldRequestURL, r.URL.Path),
	}

	// put request scope tags on context
	scopeTags := map[string]string{
		scopeTagEndpoint: endpoint.EndpointName,
		scopeTagHandler:  endpoint.HandlerName,
		scopeTagProtocol: scopeTagHTTP,
	}
	if endpoint.contextExtractor != nil {
		headers := map[string]string{}

		for k, v := range r.Header {
			// TODO: this 0th element logic is probably not correct
			headers[k] = v[0]
		}
		ctx = WithEndpointRequestHeadersField(ctx, headers)
		for k, v := range endpoint.contextExtractor.ExtractScopeTags(ctx) {
			scopeTags[k] = v
		}

		logFields = append(logFields, endpoint.contextExtractor.ExtractLogFields(ctx)...)
	}

	// Overriding the api-environment and default to production
	apiEnvironment := GetAPIEnvironment(endpoint, r)
	scopeTags[scopeTagsAPIEnvironment] = apiEnvironment
	logFields = append(logFields, zap.String(apienvironmentKey, apiEnvironment))

	// Overriding the environment for shadow requests
	if endpoint.config != nil {
		if endpoint.config.ContainsKey("service.shadow.env.override.enable") &&
			endpoint.config.MustGetBoolean("service.shadow.env.override.enable") &&
			endpoint.config.ContainsKey("shadowRequestHeader") &&
			r.Header.Get(endpoint.config.MustGetString("shadowRequestHeader")) != "" {
			scopeTags[environmentKey] = shadowEnvironment
			logFields = append(logFields, zap.String(environmentKey, shadowEnvironment))
		}
	}

	ctx = WithScopeTags(ctx, scopeTags)
	ctx = WithLogFields(ctx, logFields...)

	httpRequest := r.WithContext(ctx)

	scope := endpoint.scope.Tagged(scopeTags)
	logger := endpoint.contextLogger

	req := &ServerHTTPRequest{
		httpRequest:   httpRequest,
		queryValues:   nil,
		tracer:        endpoint.tracer,
		EndpointName:  endpoint.EndpointName,
		HandlerName:   endpoint.HandlerName,
		URL:           httpRequest.URL,
		Method:        httpRequest.Method,
		Params:        params,
		Header:        NewServerHTTPHeader(r.Header),
		contextLogger: logger,
		scope:         scope,
		jsonWrapper:   endpoint.JSONWrapper,
	}

	req.res = NewServerHTTPResponse(w, req)
	req.start()
	return req
}

// GetAPIEnvironment returns the api environment for a given request.
// By default, the api environment is set to production. However, there may be
// use cases where a different environment may be required for monitoring purposes.
// This may be overridden by a non-empty environment value in the request header.
func GetAPIEnvironment(endpoint *RouterEndpoint, r *http.Request) string {
	apiEnvironment := apiEnvironmentDefault
	if endpoint.config != nil &&
		endpoint.config.ContainsKey("apiEnvironmentHeader") &&
		r.Header.Get(endpoint.config.MustGetString("apiEnvironmentHeader")) != "" {
		apiEnvironment = r.Header.Get(endpoint.config.MustGetString("apiEnvironmentHeader"))
	}
	return apiEnvironment
}

// Context returns the request's context.
func (req *ServerHTTPRequest) Context() context.Context {
	return req.httpRequest.Context()
}

// StartTime returns the request's start time.
func (req *ServerHTTPRequest) StartTime() time.Time {
	return req.startTime
}

// start the request, emit metrics etc
func (req *ServerHTTPRequest) start() {
	if req.started {
		/* coverage ignore next line */
		req.contextLogger.Error(req.Context(),
			"Cannot start ServerHTTPRequest twice",
			zap.String("path", req.URL.Path),
		)
		/* coverage ignore next line */
		return
	}
	req.started = true
	req.startTime = time.Now()

	// emit request count
	req.scope.Counter(endpointRequest).Inc(1)

	if req.tracer != nil {
		opName := fmt.Sprintf("%s.%s", req.EndpointName, req.HandlerName)
		urlTag := opentracing.Tag{Key: "URL", Value: req.URL}
		MethodTag := opentracing.Tag{Key: "Method", Value: req.Method}
		carrier := opentracing.HTTPHeadersCarrier(req.httpRequest.Header)
		spanContext, err := req.tracer.Extract(opentracing.HTTPHeaders, carrier)
		var span opentracing.Span
		if err != nil {
			if err != opentracing.ErrSpanContextNotFound {
				/* coverage ignore next line */
				req.contextLogger.WarnZ(req.Context(), "Error Extracting Trace Headers", zap.Error(err))
			}
			span = req.tracer.StartSpan(opName, urlTag, MethodTag)
		} else {
			span = req.tracer.StartSpan(opName, urlTag, MethodTag, ext.RPCServerOption(spanContext))
		}
		req.span = span
	}
}

// CheckHeaders verifies that request contains required headers.
func (req *ServerHTTPRequest) CheckHeaders(headers []string) bool {
	for _, headerName := range headers {
		_, ok := req.Header.Get(headerName)
		if !ok {
			req.contextLogger.WarnZ(req.Context(), "Got request without mandatory header",
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

// PeekBody allows for inspecting a key path inside the body
// that is not flushed yet. This is useful for response middlewares
// that want to inspect the response body.
func (req *ServerHTTPRequest) PeekBody(
	keys ...string,
) ([]byte, jsonparser.ValueType, error) {
	value, valueType, _, err := jsonparser.Get(
		req.rawBody, keys...,
	)

	if err != nil {
		return nil, -1, err
	}

	return value, valueType, nil
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
		req.contextLogger.WarnZ(req.Context(), "Got request with invalid query string", zap.Error(err))

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

// SetQueryValue will set the value of the query parameter, replacing any existing value
// We only work with one value, and not a list (in keeping with other url.Values methods)
func (req *ServerHTTPRequest) SetQueryValue(key string, value string) {
	if req.queryValues == nil {
		req.queryValues = make(url.Values)
	}

	req.queryValues.Set(key, value)
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

	req.LogAndSendQueryError(err, "bool", key, value)
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
		req.LogAndSendQueryError(err, "int8", key, value)
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
		req.LogAndSendQueryError(err, "int16", key, value)
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
		req.LogAndSendQueryError(err, "int32", key, value)
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
		req.LogAndSendQueryError(err, "int64", key, value)
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
		req.LogAndSendQueryError(err, "float64", key, value)
		return 0, false
	}

	return number, true
}

// -- Query params as  lists --

// GetQueryBoolList will return a query param as a list of boolean
func (req *ServerHTTPRequest) GetQueryBoolList(key string) ([]bool, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	values := req.queryValues[key]
	ret := make([]bool, len(values))
	for i, value := range values {
		if value == "true" {
			ret[i] = true
		} else if value == "false" {
			ret[i] = false
		} else {
			err := &strconv.NumError{
				Func: "ParseBool",
				Num:  value,
				Err:  strconv.ErrSyntax,
			}
			req.LogAndSendQueryError(err, "bool", key, value)
			return nil, false
		}
	}

	return ret, true
}

// GetQueryInt8List will return a query params as list of int8
func (req *ServerHTTPRequest) GetQueryInt8List(key string) ([]int8, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	values := req.queryValues[key]
	ret := make([]int8, len(values))
	for i, value := range values {
		number, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			req.LogAndSendQueryError(err, "int8", key, value)
			return nil, false
		}
		ret[i] = int8(number)
	}
	return ret, true
}

// GetQueryInt16List will return a query params as list of int16
func (req *ServerHTTPRequest) GetQueryInt16List(key string) ([]int16, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	values := req.queryValues[key]
	ret := make([]int16, len(values))
	for i, value := range values {
		number, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			req.LogAndSendQueryError(err, "int16", key, value)
			return nil, false
		}
		ret[i] = int16(number)
	}
	return ret, true
}

// GetQueryInt32List will return a query params as list of int32
func (req *ServerHTTPRequest) GetQueryInt32List(key string) ([]int32, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	values := req.queryValues[key]
	ret := make([]int32, len(values))
	for i, value := range values {
		number, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			req.LogAndSendQueryError(err, "int32", key, value)
			return nil, false
		}
		ret[i] = int32(number)
	}
	return ret, true
}

// GetQueryInt64List will return a query params as list of int64
func (req *ServerHTTPRequest) GetQueryInt64List(key string) ([]int64, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	values := req.queryValues[key]
	ret := make([]int64, len(values))
	for i, value := range values {
		number, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			req.LogAndSendQueryError(err, "int64", key, value)
			return nil, false
		}
		ret[i] = number
	}
	return ret, true
}

// GetQueryFloat64List will return a query params as list of float64
func (req *ServerHTTPRequest) GetQueryFloat64List(key string) ([]float64, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	values := req.queryValues[key]
	ret := make([]float64, len(values))
	for i, value := range values {
		number, err := strconv.ParseFloat(value, 64)
		if err != nil {
			req.LogAndSendQueryError(err, "float64", key, value)
			return nil, false
		}
		ret[i] = number
	}
	return ret, true
}

// GetQueryValues will return all query parameters for key.
func (req *ServerHTTPRequest) GetQueryValues(key string) ([]string, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	return req.queryValues[key], true
}

// GetQueryValueList will return all query parameters for key.
func (req *ServerHTTPRequest) GetQueryValueList(key string) ([]string, bool) {
	return req.GetQueryValues(key)
}

// -- Query param as set --

/**
 * A set of bools does not make sense and is unimplemented
 * Also, in every use-case for a gateway, a set implemented as a map is not very useful, instead one where
 * it is a list with no duplicates is. Therefore the implementation picks that approach.
 */

// The "value" in the map representation of a set datastructure
var _nullVal = struct{}{}

// GetQueryInt8Set will return a query params as set of int8 (implemented as a deduped slice)
func (req *ServerHTTPRequest) GetQueryInt8Set(key string) ([]int8, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	values := req.queryValues[key]
	set := make(map[int8]struct{}, len(values))
	for _, value := range values {
		number, err := strconv.ParseInt(value, 0, 8)
		if err != nil {
			req.LogAndSendQueryError(err, "int8", key, value)
			return nil, false
		}
		set[int8(number)] = _nullVal
	}
	ret := make([]int8, len(set))
	i := 0
	for item := range set {
		ret[i] = item
		i++
	}
	return ret, true
}

// GetQueryInt16Set will return a query params as set of int16
func (req *ServerHTTPRequest) GetQueryInt16Set(key string) ([]int16, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	values := req.queryValues[key]
	set := make(map[int16]struct{}, len(values))
	for _, value := range values {
		number, err := strconv.ParseInt(value, 0, 16)
		if err != nil {
			req.LogAndSendQueryError(err, "int16", key, value)
			return nil, false
		}
		set[int16(number)] = _nullVal
	}
	ret := make([]int16, len(set))
	i := 0
	for item := range set {
		ret[i] = item
		i++
	}
	return ret, true
}

// GetQueryInt32Set will return a query params as set of int32
func (req *ServerHTTPRequest) GetQueryInt32Set(key string) ([]int32, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	values := req.queryValues[key]
	set := make(map[int32]struct{}, len(values))
	for _, value := range values {
		number, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			req.LogAndSendQueryError(err, "int32", key, value)
			return nil, false
		}
		set[int32(number)] = _nullVal
	}
	ret := make([]int32, len(set))
	i := 0
	for item := range set {
		ret[i] = item
		i++
	}
	return ret, true
}

// GetQueryInt64Set will return a query params as set of int64
func (req *ServerHTTPRequest) GetQueryInt64Set(key string) ([]int64, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	values := req.queryValues[key]
	set := make(map[int64]struct{}, len(values))
	for _, value := range values {
		number, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			req.LogAndSendQueryError(err, "int64", key, value)
			return nil, false
		}
		set[number] = _nullVal
	}
	ret := make([]int64, len(set))
	i := 0
	for item := range set {
		ret[i] = item
		i++
	}
	return ret, true
}

// GetQueryFloat64Set will return a query params as set of float64
func (req *ServerHTTPRequest) GetQueryFloat64Set(key string) ([]float64, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	values := req.queryValues[key]
	set := make(map[float64]struct{}, len(values))
	for _, value := range values {
		number, err := strconv.ParseFloat(value, 64)
		if err != nil {
			req.LogAndSendQueryError(err, "float64", key, value)
			return nil, false
		}
		set[number] = _nullVal
	}
	ret := make([]float64, len(set))
	i := 0
	for item := range set {
		ret[i] = item
		i++
	}
	return ret, true
}

// GetQueryValueSet will return all query parameters for key as a set
func (req *ServerHTTPRequest) GetQueryValueSet(key string) ([]string, bool) {
	success := req.parseQueryValues()
	if !success {
		return nil, false
	}

	set := make(map[string]struct{}, len(req.queryValues[key]))
	for _, v := range req.queryValues[key] {
		set[v] = _nullVal
	}
	ret := make([]string, len(set))
	i := 0
	for item := range set {
		ret[i] = item
		i++
	}
	return ret, true
}

// HasQueryPrefix will check if any query param starts with key.
func (req *ServerHTTPRequest) HasQueryPrefix(prefix string) bool {
	success := req.parseQueryValues()
	if !success {
		return false
	}

	for key := range req.queryValues {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	return false
}

// CheckQueryValue will check for a required query param.
func (req *ServerHTTPRequest) CheckQueryValue(key string) bool {
	success := req.parseQueryValues()
	if !success {
		return false
	}

	values := req.queryValues[key]
	if len(values) == 0 {
		req.contextLogger.WarnZ(req.Context(), "Got request with missing query string value",
			zap.String("expectedKey", key),
		)
		if !req.parseFailed {
			req.res.SendErrorString(
				400, "Could not parse query string",
			)
			req.parseFailed = true
		}
		return false
	}

	return true
}

// HasQueryValue will return bool if the query param exists.
func (req *ServerHTTPRequest) HasQueryValue(key string) bool {
	success := req.parseQueryValues()
	if !success {
		return false
	}

	values := req.queryValues[key]
	if len(values) == 0 {
		return false
	}

	return true
}

// ReadAndUnmarshalBody will try to unmarshal into struct or fail
func (req *ServerHTTPRequest) ReadAndUnmarshalBody(
	body interface{},
) bool {
	rawBody, success := req.ReadAll()
	if !success {
		return false
	}
	return req.UnmarshalBody(body, rawBody)
}

// GetRawBody returns raw body of request
func (req *ServerHTTPRequest) GetRawBody() []byte {
	return req.rawBody
}

// ReadAll helper to read entire body
func (req *ServerHTTPRequest) ReadAll() ([]byte, bool) {
	if req.rawBody != nil {
		return req.rawBody, true
	}
	rawBody, err := ioutil.ReadAll(req.httpRequest.Body)
	if err != nil {
		req.contextLogger.ErrorZ(req.Context(), "Could not read request body", zap.Error(err))
		if !req.parseFailed {
			req.res.SendError(500, "Could not read request body", err)
			req.parseFailed = true
		}
		return nil, false
	}
	req.rawBody = rawBody
	return rawBody, true
}

// UnmarshalBody helper to unmarshal body into struct
func (req *ServerHTTPRequest) UnmarshalBody(
	body interface{}, rawBody []byte,
) bool {
	err := req.jsonWrapper.Unmarshal(rawBody, body)
	if err != nil {
		req.contextLogger.WarnZ(req.Context(), "Could not parse json", zap.Error(err))
		if !req.parseFailed {
			req.res.SendError(400, "Could not parse json: "+err.Error(), err)
			req.parseFailed = true
		}
		return false
	}

	return true
}

// ReplaceBody replaces the raw request body with given body and updates the request content-length header accordingly.
// This method is only supposed to be used in middlewares where request body needs to be modified.
// The encoding of the body should stay the same.
func (req *ServerHTTPRequest) ReplaceBody(body []byte) {
	// Replace the cached body bytes and fix dependent header
	req.rawBody = body
	if _, ok := req.Header.Get("Content-Length"); ok {
		req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	}
}

// GetSpan returns the http request span
func (req *ServerHTTPRequest) GetSpan() opentracing.Span {
	return req.span
}

// LogAndSendQueryError handles parse failure of query params by logging the issue and returning a 400 to the requestor
func (req *ServerHTTPRequest) LogAndSendQueryError(err error, expected, key, value string) {
	req.contextLogger.WarnZ(req.Context(), "Got request with invalid query string types",
		zap.String("expected", expected),
		zap.String("actual", value),
		zap.String("key", key),
		zap.Error(err),
	)
	if !req.parseFailed {
		req.res.SendError(400, "Could not parse query string", err)
		req.parseFailed = true
	}
}
