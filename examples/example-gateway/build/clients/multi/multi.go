// Code generated by zanzibar
// @generated

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

package multiclient

import (
	"context"
	"fmt"
	"time"

	"github.com/afex/hystrix-go/hystrix"

	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/runtime/jsonwrapper"

	module "github.com/uber/zanzibar/examples/example-gateway/build/clients/multi/module"
)

// Client defines multi client interface.
type Client interface {
	HTTPClient() *zanzibar.HTTPClient
	HelloA(
		ctx context.Context,
		reqHeaders map[string]string,
	) (string, map[string]string, error)
	HelloB(
		ctx context.Context,
		reqHeaders map[string]string,
	) (string, map[string]string, error)
}

// multiClient is the http client.
type multiClient struct {
	clientID                  string
	httpClient                *zanzibar.HTTPClient
	jsonWrapper               jsonwrapper.JSONWrapper
	circuitBreakerDisabled    bool
	requestUUIDHeaderKey      string
	requestProcedureHeaderKey string
}

// NewClient returns a new http client.
func NewClient(deps *module.Dependencies) Client {
	ip := deps.Default.Config.MustGetString("clients.multi.ip")
	port := deps.Default.Config.MustGetInt("clients.multi.port")
	baseURL := fmt.Sprintf("http://%s:%d", ip, port)
	timeoutVal := int(deps.Default.Config.MustGetInt("clients.multi.timeout"))
	timeout := time.Millisecond * time.Duration(
		timeoutVal,
	)
	defaultHeaders := make(map[string]string)
	if deps.Default.Config.ContainsKey("http.defaultHeaders") {
		deps.Default.Config.MustGetStruct("http.defaultHeaders", &defaultHeaders)
	}
	if deps.Default.Config.ContainsKey("clients.multi.defaultHeaders") {
		deps.Default.Config.MustGetStruct("clients.multi.defaultHeaders", &defaultHeaders)
	}
	var requestUUIDHeaderKey string
	if deps.Default.Config.ContainsKey("http.clients.requestUUIDHeaderKey") {
		requestUUIDHeaderKey = deps.Default.Config.MustGetString("http.clients.requestUUIDHeaderKey")
	}
	var requestProcedureHeaderKey string
	if deps.Default.Config.ContainsKey("http.clients.requestProcedureHeaderKey") {
		requestProcedureHeaderKey = deps.Default.Config.MustGetString("http.clients.requestProcedureHeaderKey")
	}
	followRedirect := true
	if deps.Default.Config.ContainsKey("clients.multi.followRedirect") {
		followRedirect = deps.Default.Config.MustGetBoolean("clients.multi.followRedirect")
	}

	methodNames := map[string]string{
		"HelloA": "ServiceABack::hello",
		"HelloB": "ServiceBBack::hello",
	}
	// circuitBreakerDisabled sets whether circuit-breaker should be disabled
	circuitBreakerDisabled := false
	if deps.Default.Config.ContainsKey("clients.multi.circuitBreakerDisabled") {
		circuitBreakerDisabled = deps.Default.Config.MustGetBoolean("clients.multi.circuitBreakerDisabled")
	}
	if !circuitBreakerDisabled {
		for methodName := range methodNames {
			circuitBreakerName := "multi" + "-" + methodName
			configureCircuitBreaker(deps, timeoutVal, circuitBreakerName)
		}
	}

	levels := map[string]string{
		"HelloA": "1",
		"HelloB": "1",
	}

	return &multiClient{
		clientID: "multi",
		httpClient: zanzibar.NewHTTPClientContext(
			deps.Default.ContextLogger, deps.Default.ContextMetrics, deps.Default.JSONWrapper,
			"multi",
			methodNames,
			baseURL,
			defaultHeaders,
			timeout,
			followRedirect,
		),
		circuitBreakerDisabled:    circuitBreakerDisabled,
		requestUUIDHeaderKey:      requestUUIDHeaderKey,
		requestProcedureHeaderKey: requestProcedureHeaderKey,
	}
}

func configureCircuitBreaker(deps *module.Dependencies, timeoutVal int, circuitBreakerName string) {
	// sleepWindowInMilliseconds sets the amount of time, after tripping the circuit,
	// to reject requests before allowing attempts again to determine if the circuit should again be closed
	sleepWindowInMilliseconds := 5000
	if deps.Default.Config.ContainsKey("clients.multi.sleepWindowInMilliseconds") {
		sleepWindowInMilliseconds = int(deps.Default.Config.MustGetInt("clients.multi.sleepWindowInMilliseconds"))
	}
	// maxConcurrentRequests sets how many requests can be run at the same time, beyond which requests are rejected
	maxConcurrentRequests := 20
	if deps.Default.Config.ContainsKey("clients.multi.maxConcurrentRequests") {
		maxConcurrentRequests = int(deps.Default.Config.MustGetInt("clients.multi.maxConcurrentRequests"))
	}
	// errorPercentThreshold sets the error percentage at or above which the circuit should trip open
	errorPercentThreshold := 20
	if deps.Default.Config.ContainsKey("clients.multi.errorPercentThreshold") {
		errorPercentThreshold = int(deps.Default.Config.MustGetInt("clients.multi.errorPercentThreshold"))
	}
	// requestVolumeThreshold sets a minimum number of requests that will trip the circuit in a rolling window of 10s
	// For example, if the value is 20, then if only 19 requests are received in the rolling window of 10 seconds
	// the circuit will not trip open even if all 19 failed.
	requestVolumeThreshold := 20
	if deps.Default.Config.ContainsKey("clients.multi.requestVolumeThreshold") {
		requestVolumeThreshold = int(deps.Default.Config.MustGetInt("clients.multi.requestVolumeThreshold"))
	}
	hystrix.ConfigureCommand(circuitBreakerName, hystrix.CommandConfig{
		MaxConcurrentRequests:  maxConcurrentRequests,
		ErrorPercentThreshold:  errorPercentThreshold,
		SleepWindow:            sleepWindowInMilliseconds,
		RequestVolumeThreshold: requestVolumeThreshold,
		Timeout:                timeoutVal,
	})
}

// HTTPClient returns the underlying HTTP client, should only be
// used for internal testing.
func (c *multiClient) HTTPClient() *zanzibar.HTTPClient {
	return c.httpClient
}

// HelloA calls "/multi/serviceA_b/hello" endpoint.
func (c *multiClient) HelloA(
	ctx context.Context,
	headers map[string]string,
) (string, map[string]string, error) {
	reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
	if headers == nil {
		headers = make(map[string]string)
	}
	if reqUUID != "" {
		headers[c.requestUUIDHeaderKey] = reqUUID
	}
	if c.requestProcedureHeaderKey != "" {
		headers[c.requestProcedureHeaderKey] = "ServiceABack::hello"
	}

	var defaultRes string
	req := zanzibar.NewClientHTTPRequest(ctx, c.clientID, "HelloA", "ServiceABack::hello", c.httpClient)

	// Generate full URL.
	fullURL := c.httpClient.BaseURL + "/multi" + "/serviceA_b" + "/hello"

	err := req.WriteJSON("GET", fullURL, headers, nil)
	if err != nil {
		return defaultRes, nil, err
	}

	var res *zanzibar.ClientHTTPResponse
	if c.circuitBreakerDisabled {
		res, err = req.Do()
	} else {
		// We want hystrix ckt-breaker to count errors only for system issues
		var clientErr error
		circuitBreakerName := "multi" + "-" + "HelloA"
		err = hystrix.DoC(ctx, circuitBreakerName, func(ctx context.Context) error {
			res, clientErr = req.Do()
			if res != nil {
				// This is not a system error/issue. Downstream responded
				return nil
			}
			return clientErr
		}, nil)
		if err == nil {
			// ckt-breaker was ok, bubble up client error if set
			err = clientErr
		}
	}
	if err != nil {
		return defaultRes, nil, err
	}

	respHeaders := make(map[string]string)
	for k := range res.Header {
		respHeaders[k] = res.Header.Get(k)
	}

	res.CheckOKResponse([]int{200})

	switch res.StatusCode {
	case 200:
		var responseBody string
		rawBody, err := res.ReadAll()
		if err != nil {
			return defaultRes, respHeaders, err
		}
		err = res.UnmarshalBody(&responseBody, rawBody)
		if err != nil {
			return defaultRes, respHeaders, err
		}

		return responseBody, respHeaders, nil
	default:
		_, err = res.ReadAll()
		if err != nil {
			return defaultRes, respHeaders, err
		}
	}

	return defaultRes, respHeaders, &zanzibar.UnexpectedHTTPError{
		StatusCode: res.StatusCode,
		RawBody:    res.GetRawBody(),
	}
}

// HelloB calls "/multi/serviceB_b/hello" endpoint.
func (c *multiClient) HelloB(
	ctx context.Context,
	headers map[string]string,
) (string, map[string]string, error) {
	reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
	if headers == nil {
		headers = make(map[string]string)
	}
	if reqUUID != "" {
		headers[c.requestUUIDHeaderKey] = reqUUID
	}
	if c.requestProcedureHeaderKey != "" {
		headers[c.requestProcedureHeaderKey] = "ServiceBBack::hello"
	}

	var defaultRes string
	req := zanzibar.NewClientHTTPRequest(ctx, c.clientID, "HelloB", "ServiceBBack::hello", c.httpClient)

	// Generate full URL.
	fullURL := c.httpClient.BaseURL + "/multi" + "/serviceB_b" + "/hello"

	err := req.WriteJSON("GET", fullURL, headers, nil)
	if err != nil {
		return defaultRes, nil, err
	}

	var res *zanzibar.ClientHTTPResponse
	if c.circuitBreakerDisabled {
		res, err = req.Do()
	} else {
		// We want hystrix ckt-breaker to count errors only for system issues
		var clientErr error
		circuitBreakerName := "multi" + "-" + "HelloB"
		err = hystrix.DoC(ctx, circuitBreakerName, func(ctx context.Context) error {
			res, clientErr = req.Do()
			if res != nil {
				// This is not a system error/issue. Downstream responded
				return nil
			}
			return clientErr
		}, nil)
		if err == nil {
			// ckt-breaker was ok, bubble up client error if set
			err = clientErr
		}
	}
	if err != nil {
		return defaultRes, nil, err
	}

	respHeaders := make(map[string]string)
	for k := range res.Header {
		respHeaders[k] = res.Header.Get(k)
	}

	res.CheckOKResponse([]int{200})

	switch res.StatusCode {
	case 200:
		var responseBody string
		rawBody, err := res.ReadAll()
		if err != nil {
			return defaultRes, respHeaders, err
		}
		err = res.UnmarshalBody(&responseBody, rawBody)
		if err != nil {
			return defaultRes, respHeaders, err
		}

		return responseBody, respHeaders, nil
	default:
		_, err = res.ReadAll()
		if err != nil {
			return defaultRes, respHeaders, err
		}
	}

	return defaultRes, respHeaders, &zanzibar.UnexpectedHTTPError{
		StatusCode: res.StatusCode,
		RawBody:    res.GetRawBody(),
	}
}
