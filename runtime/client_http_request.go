// Copyright (c) 2023 Uber Technologies, Inc.
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
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/uber/zanzibar/v2/runtime/jsonwrapper"

	"go.uber.org/zap"
)

var metricNormalizer = strings.NewReplacer("::", "--")

// ClientHTTPRequest is the struct for making a single client request using an outbound http client.
type ClientHTTPRequest struct {
	ClientID               string
	ClientTargetEndpoint   string
	MethodName             string
	Metrics                ContextMetrics
	client                 *HTTPClient
	httpReq                *http.Request
	res                    *ClientHTTPResponse
	started                bool
	startTime              time.Time
	Logger                 *zap.Logger
	ContextLogger          ContextLogger
	rawBody                []byte
	defaultHeaders         map[string]string
	ctx                    context.Context
	jsonWrapper            jsonwrapper.JSONWrapper
	timeoutAndRetryOptions *TimeoutAndRetryOptions
}

// NewClientHTTPRequest allocates a ClientHTTPRequest. The ctx parameter is the context associated with the outbound requests.
func NewClientHTTPRequest(
	ctx context.Context,
	clientID string,
	clientMethod string,
	clientTargetEndpoint string,
	client *HTTPClient,
) *ClientHTTPRequest {
	scopeTags := map[string]string{
		scopeTagClientMethod:    clientMethod,
		scopeTagClient:          clientID,
		scopeTagsTargetEndpoint: metricNormalizer.Replace(clientTargetEndpoint),
	}

	ctx = WithScopeTagsDefault(ctx, scopeTags, client.contextMetrics.Scope())
	req := &ClientHTTPRequest{
		ClientID:             clientID,
		MethodName:           clientMethod,
		ClientTargetEndpoint: clientTargetEndpoint,
		Metrics:              client.contextMetrics,
		client:               client,
		ContextLogger:        client.ContextLogger,
		defaultHeaders:       client.DefaultHeaders,
		ctx:                  ctx,
		jsonWrapper:          client.JSONWrapper,
	}

	req.res = NewClientHTTPResponse(req)
	req.timeoutAndRetryOptions = GetTimeoutAndRetryOptions(ctx)
	req.start()
	return req
}

// Start the request, do some metrics book keeping
func (req *ClientHTTPRequest) start() {
	if req.started {
		/* coverage ignore next line */
		req.ContextLogger.ErrorZ(req.ctx, "Cannot start ClientHTTPRequest twice")
		/* coverage ignore next line */
		return
	}
	req.started = true
	req.startTime = time.Now()
}

// CheckHeaders verifies that the outbound request contains required headers
func (req *ClientHTTPRequest) CheckHeaders(expected []string) error {
	if req.httpReq == nil {
		/* coverage ignore next line */
		panic("must call `req.WriteJSON()` before `req.CheckHeaders()`")
	}

	actualHeaders := req.httpReq.Header
	missingHeaders := make([]string, 0)
	for _, headerName := range expected {
		// headerName is case insensitive, http.Header Get canonicalize the key
		headerValue := actualHeaders.Get(headerName)
		if headerValue == "" {
			missingHeaders = append(missingHeaders, headerName)
		}
	}
	if len(missingHeaders) > 0 {
		return errors.New("missing mandatory headers: " + strings.Join(missingHeaders, ","))
	}
	return nil
}

// WriteJSON materialize the HTTP request with given method, url, headers and body.
func (req *ClientHTTPRequest) WriteJSON(
	method, url string,
	headers map[string]string,
	body interface{},
) error {
	var rawBody []byte
	if body != nil {
		var err error
		rawBody, err = req.jsonWrapper.Marshal(body)
		if err != nil {
			return errors.Wrapf(
				err, "Could not serialize %s.%s request object",
				req.ClientID, req.MethodName,
			)
		}
	}

	return req.WriteBytes(method, url, headers, rawBody)
}

// WriteBytes materialize the HTTP request with given method, url, headers and body.
// Body is assumed to be a byte array.s
func (req *ClientHTTPRequest) WriteBytes(
	method, url string,
	headers map[string]string,
	rawBody []byte,
) error {
	var httpReq *http.Request
	var httpErr error

	if rawBody != nil {
		req.rawBody = rawBody
		httpReq, httpErr = http.NewRequest(method, url, bytes.NewReader(rawBody))
	} else {
		httpReq, httpErr = http.NewRequest(method, url, nil)
	}

	if httpErr != nil {
		return errors.Wrapf(
			httpErr, "Could not create outbound %s.%s request",
			req.ClientID, req.MethodName,
		)
	}

	// Using `Add` over `Set` intentionally, allowing us to create a list
	// of headerValues for a given key.
	for headerKey, headerValue := range req.defaultHeaders {
		httpReq.Header.Set(headerKey, headerValue)
	}

	for k := range headers {
		httpReq.Header.Set(k, headers[k])
	}

	req.httpReq = httpReq
	return nil
}

// Do will send the request out.
func (req *ClientHTTPRequest) Do() (*ClientHTTPResponse, error) {
	opName := fmt.Sprintf("%s.%s(%s)", req.ClientID, req.MethodName, req.ClientTargetEndpoint)
	urlTag := opentracing.Tag{Key: "URL", Value: req.httpReq.URL}
	methodTag := opentracing.Tag{Key: "Method", Value: req.httpReq.Method}
	span, ctx := opentracing.StartSpanFromContext(req.ctx, opName, urlTag, methodTag)
	err := req.InjectSpanToHeader(span, opentracing.HTTPHeaders)
	if err != nil {
		/* coverage ignore next line */
		req.ContextLogger.ErrorZ(req.ctx, "Fail to inject span to headers", zap.Error(err))
		/* coverage ignore next line */
		return nil, err
	}
	var retryCount int64 = 1
	var res *http.Response

	// when timeoutAndRetryOptions per request is not configured, use default client level timeout
	if req.timeoutAndRetryOptions == nil || req.timeoutAndRetryOptions.MaxAttempts == 0 {
		res, err = req.client.Client.Do(req.httpReq.WithContext(ctx))
	} else {
		res, retryCount, err = req.executeDoWithRetry(ctx) // new code for retry and timeout per ep level
	}

	span.Finish()

	AppendLogFieldsToContext(req.ctx, zap.Int64(fmt.Sprintf(logFieldClientAttempts, req.ClientID), retryCount))
	if err != nil {
		return nil, errors.Wrapf(err, "errors while making outbound %s.%s request", req.ClientID, req.MethodName)
	}

	// emit metrics
	req.Metrics.IncCounter(req.ctx, clientRequest, retryCount)

	req.res.setRawHTTPResponse(res)

	// generate events
	if GetToCapture(ctx) {
		// ReadAll, caches bytes internally so multiple calls will return the same data
		rspBytes, err := req.res.ReadAll()
		if err == nil {
			event := &HTTPOutgoingEvent{
				ClientID:       req.ClientID,
				ClientEndpoint: req.ClientTargetEndpoint,
				HTTPCapture: HTTPCapture{
					ReqURL:        req.httpReq.URL.String(),
					ReqMethod:     req.httpReq.Method,
					ReqHeaders:    req.httpReq.Header.Clone(), // cloning headers maybe inefficient
					ReqBody:       req.rawBody,
					RspStatusCode: res.StatusCode,
					RspHeaders:    res.Header.Clone(), // cloning headers maybe inefficient
					RspBody:       rspBytes,
				},
			}

			if ec := GetEventContainer(ctx); ec != nil {
				ec.events = append(ec.events, event)
			}
		}
	}

	return req.res, nil
}

// executeDoWithRetry will execute executeDo with retries
func (req *ClientHTTPRequest) executeDoWithRetry(ctx context.Context) (*http.Response, int64, error) {
	var err error
	var res *http.Response
	var retryCount int64 = 0

	for i := 0; i < req.timeoutAndRetryOptions.MaxAttempts; i++ {
		retryCount++
		res, err = req.executeDo(ctx)

		if err == nil {
			return res, retryCount, nil
		}

		var shouldRetry = false
		// if attempts are pending, wait for backoff duration before next attempt
		if i+1 < req.timeoutAndRetryOptions.MaxAttempts {
			shouldRetry = req.client.CheckRetry(ctx, req.timeoutAndRetryOptions, res, err)
		}

		// TODO (future releases) - make retry conditional, inspect error/response and then retry

		// reassign body
		if req.rawBody != nil && len(req.rawBody) > 0 {
			req.httpReq.Body = io.NopCloser(bytes.NewBuffer(req.rawBody))
		}

		// Break loop if no retries
		if !shouldRetry {
			break
		}
	}
	return nil, retryCount, err
}

// executeDo will send the request out with a timeout
func (req *ClientHTTPRequest) executeDo(ctx context.Context) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, req.timeoutAndRetryOptions.RequestTimeoutPerAttemptInMs)
	defer cancel()
	res, err := req.client.Client.Do(req.httpReq.WithContext(ctx))
	// when no error, read body and capture before closing the connection
	if err == nil {
		req.res.setRawHTTPResponse(res)
		_, err = req.res.ReadAll()
		if err != nil {
			return nil, err
		}
	}
	return res, err
}

// InjectSpanToHeader will inject span to request header
// This method is current used for unit tests
// TODO: we need to set source and test code as same pkg name which would makes UTs easier
func (req *ClientHTTPRequest) InjectSpanToHeader(span opentracing.Span, format interface{}) error {
	carrier := opentracing.HTTPHeadersCarrier(req.httpReq.Header)
	if err := span.Tracer().Inject(span.Context(), format, carrier); err != nil {
		req.ContextLogger.ErrorZ(req.ctx, "Failed to inject tracing span.", zap.Error(err))
		return err
	}

	return nil
}
