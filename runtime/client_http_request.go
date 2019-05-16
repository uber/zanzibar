// Copyright (c) 2019 Uber Technologies, Inc.
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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var metricNormalizer = strings.NewReplacer("::", "--")

// ClientHTTPRequest is the struct for making a single client request using an outbound http client.
type ClientHTTPRequest struct {
	ClientID             string
	ClientTargetEndpoint string
	MethodName           string
	client               *HTTPClient
	httpReq              *http.Request
	res                  *ClientHTTPResponse
	started              bool
	startTime            time.Time
	Logger               *zap.Logger
	ContextLogger        ContextLogger
	rawBody              []byte
	defaultHeaders       map[string]string
	ctx                  context.Context
	metrics              ContextMetrics
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

	ctx = WithScopeTags(ctx, scopeTags)
	req := &ClientHTTPRequest{
		ClientID:             clientID,
		MethodName:           clientMethod,
		ClientTargetEndpoint: clientTargetEndpoint,
		client:               client,
		Logger:               client.loggers[clientMethod],
		ContextLogger:        NewContextLogger(client.loggers[clientMethod]),
		defaultHeaders:       client.DefaultHeaders,
		ctx:                  ctx,
		metrics:              client.contextMetrics,
	}
	req.res = NewClientHTTPResponse(req)
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

// CheckHeaders verifies that the outbound request contains required headers
func (req *ClientHTTPRequest) CheckHeaders(expected []string) error {
	if req.httpReq == nil {
		/* coverage ignore next line */
		panic("must call `req.WriteJSON()` before `req.CheckHeaders()`")
	}

	actualHeaders := req.httpReq.Header

	for _, headerName := range expected {
		// headerName is case insensitive, http.Header Get canonicalize the key
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

// WriteJSON materialize the HTTP request with given method, url, headers and body.
// Body is serialized to JSON bytes using the stdlib json.Marshal function, which
// calls the `MarshalJSON` method if the body implements the Marshaler interface.
func (req *ClientHTTPRequest) WriteJSON(
	method, url string,
	headers map[string]string,
	body interface{},
) error {
	var httpReq *http.Request
	var httpErr error
	if body != nil {
		rawBody, err := json.Marshal(body)
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

	// Using `Add` over `Set` intentionally, allowing us to create a list
	// of headerValues for a given key.
	for headerKey, headerValue := range req.defaultHeaders {
		httpReq.Header.Set(headerKey, headerValue)
	}

	for k := range headers {
		httpReq.Header.Set(k, headers[k])
	}

	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	req.httpReq = httpReq
	req.ctx = WithLogFields(req.ctx,
		zap.String(logFieldClientHTTPMethod, method),
		zap.String(logFieldRequestURL, url),
		zap.Time(logFieldRequestStartTime, req.startTime),
	)

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
		req.Logger.Error("Fail to inject span to headers", zap.Error(err))
		/* coverage ignore next line */
		return nil, err
	}

	res, err := req.client.Client.Do(req.httpReq.WithContext(ctx))
	span.Finish()
	if err != nil {
		req.Logger.Error("Could not make outbound request", zap.Error(err))
		return nil, err
	}

	// emit metrics
	req.metrics.IncCounter(req.ctx, clientRequest, 1)

	req.res.setRawHTTPResponse(res)
	return req.res, nil
}

// InjectSpanToHeader will inject span to request header
// This method is current used for unit tests
// TODO: we need to set source and test code as same pkg name which would makes UTs easier
func (req *ClientHTTPRequest) InjectSpanToHeader(span opentracing.Span, format interface{}) error {
	carrier := opentracing.HTTPHeadersCarrier(req.httpReq.Header)
	if err := span.Tracer().Inject(span.Context(), format, carrier); err != nil {
		req.Logger.Error("Failed to inject tracing span.", zap.Error(err))
		return err
	}

	return nil
}
