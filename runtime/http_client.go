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
	"net/http"
	"strconv"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

// HTTPClient defines a http client.
type HTTPClient struct {
	Client         *http.Client
	BaseURL        string
	DefaultHeaders map[string]string
	loggers        map[string]*zap.Logger
	ContextMetrics ContextMetrics
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
	logger *zap.Logger,
	scope tally.Scope,
	clientID string,
	methodNames []string,
	baseURL string,
	defaultHeaders map[string]string,
	timeout time.Duration,
) *HTTPClient {
	loggers := make(map[string]*zap.Logger, len(methodNames))

	for _, methodName := range methodNames {
		loggers[methodName] = logger.With(
			zap.String("clientID", clientID),
			zap.String("methodName", methodName),
		)
	}
	return &HTTPClient{
		Client: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
			},
			Timeout: timeout,
		},
		BaseURL:        baseURL,
		DefaultHeaders: defaultHeaders,
		loggers:        loggers,
		ContextMetrics: NewContextMetrics(scope),
	}
}

// NewHTTPClientContext will allocate a http client.
func NewHTTPClientContext(
	logger *zap.Logger,
	ContextMetrics ContextMetrics,
	clientID string,
	methodNames []string,
	baseURL string,
	defaultHeaders map[string]string,
	timeout time.Duration,
) *HTTPClient {
	loggers := make(map[string]*zap.Logger, len(methodNames))

	for _, methodName := range methodNames {
		loggers[methodName] = logger.With(
			zap.String("clientID", clientID),
			zap.String("methodName", methodName),
		)
	}
	return &HTTPClient{
		Client: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
			},
			Timeout: timeout,
		},
		BaseURL:        baseURL,
		DefaultHeaders: defaultHeaders,
		loggers:        loggers,
		ContextMetrics: ContextMetrics,
	}
}
