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
	"net/http"
	"strconv"
	"time"

	"github.com/uber-go/tally"
	"github.com/uber/zanzibar/runtime/jsonwrapper"
)

// CheckRetry specifies a policy for handling retries. It is called
// following each request with the response and error values returned by
// the http.Client. If CheckRetry returns false, the Client stops retrying
// and returns the response to the caller. If CheckRetry returns an error,
// that error value is returned in lieu of the error from the request.
type CheckRetry func(ctx context.Context, timeoutAndRetryOptions *TimeoutAndRetryOptions, resp *http.Response, err error) bool

// HTTPClient defines a http client.
type HTTPClient struct {
	Client         *http.Client
	BaseURL        string
	DefaultHeaders map[string]string
	JSONWrapper    jsonwrapper.JSONWrapper
	ContextLogger  ContextLogger
	contextMetrics ContextMetrics
	CheckRetry     CheckRetry
}

// UnexpectedHTTPError defines an error for HTTP
type UnexpectedHTTPError struct {
	StatusCode int
	RawBody    []byte
}

func (rawErr *UnexpectedHTTPError) Error() string {
	return "Unexpected http client response (" +
		strconv.Itoa(rawErr.StatusCode) + ")"
}

// NewHTTPClient is deprecated, use NewHTTPClientContext instead
func NewHTTPClient(
	contextLogger ContextLogger,
	scope tally.Scope,
	jsonWrapper jsonwrapper.JSONWrapper,
	clientID string,
	methodToTargetEndpoint map[string]string,
	baseURL string,
	defaultHeaders map[string]string,
	timeout time.Duration,
) *HTTPClient {
	return NewHTTPClientContext(
		contextLogger,
		NewContextMetrics(scope),
		jsonWrapper,
		clientID,
		methodToTargetEndpoint,
		baseURL,
		defaultHeaders,
		timeout,
		true,
	)
}

// NewHTTPClientContext will allocate a http client.
func NewHTTPClientContext(
	contextLogger ContextLogger,
	ContextMetrics ContextMetrics,
	jsonWrapper jsonwrapper.JSONWrapper,
	clientID string,
	methodToTargetEndpoint map[string]string,
	baseURL string,
	defaultHeaders map[string]string,
	timeout time.Duration,
	followRedirect bool,
) *HTTPClient {

	var checkRedirect func(req *http.Request, via []*http.Request) error
	if !followRedirect {
		checkRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return &HTTPClient{
		Client: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
			},
			Timeout:       timeout,
			CheckRedirect: checkRedirect,
		},
		BaseURL:        baseURL,
		DefaultHeaders: defaultHeaders,
		ContextLogger:  contextLogger,
		contextMetrics: ContextMetrics,
		JSONWrapper:    jsonWrapper,
		CheckRetry:     DefaultRetryPolicy,
	}
}

// DefaultRetryPolicy allows retries for any type of server error
func DefaultRetryPolicy(ctx context.Context, timeoutAndRetryOptions *TimeoutAndRetryOptions, resp *http.Response, err error) bool {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false
	}
	//wait for the backoff time
	timer := time.NewTimer(timeoutAndRetryOptions.BackOffTimeAcrossRetriesInMs)
	select {
	case <-timer.C:
	}

	return true
}
