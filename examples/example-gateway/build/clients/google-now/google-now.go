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

package googlenowclient

import (
	"context"
	"fmt"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"go.uber.org/zap"

	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/runtime/jsonwrapper"

	module "github.com/uber/zanzibar/examples/example-gateway/build/clients/google-now/module"
	clientsIDlClientsGooglenowGooglenow "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/googlenow/googlenow"
)

// CircuitBreakerConfigKey is key value for qps level to circuit breaker parameters mapping
const CircuitBreakerConfigKey = "circuitbreaking-configurations"

var logFieldErrLocation = zanzibar.LogFieldErrorLocation("client::google-now")

// Client defines google-now client interface.
type Client interface {
	HTTPClient() *zanzibar.HTTPClient
	AddCredentials(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsIDlClientsGooglenowGooglenow.GoogleNowService_AddCredentials_Args,
	) (context.Context, map[string]string, error)
	CheckCredentials(
		ctx context.Context,
		reqHeaders map[string]string,
	) (context.Context, map[string]string, error)
}

// googleNowClient is the http client.
type googleNowClient struct {
	clientID                  string
	httpClient                *zanzibar.HTTPClient
	jsonWrapper               jsonwrapper.JSONWrapper
	circuitBreakerDisabled    bool
	requestUUIDHeaderKey      string
	requestProcedureHeaderKey string
}

// NewClient returns a new http client.
func NewClient(deps *module.Dependencies) Client {
	ip := deps.Default.Config.MustGetString("clients.google-now.ip")
	port := deps.Default.Config.MustGetInt("clients.google-now.port")
	baseURL := fmt.Sprintf("http://%s:%d", ip, port)
	timeoutVal := int(deps.Default.Config.MustGetInt("clients.google-now.timeout"))
	timeout := time.Millisecond * time.Duration(
		timeoutVal,
	)
	defaultHeaders := make(map[string]string)
	if deps.Default.Config.ContainsKey("http.defaultHeaders") {
		deps.Default.Config.MustGetStruct("http.defaultHeaders", &defaultHeaders)
	}
	if deps.Default.Config.ContainsKey("clients.google-now.defaultHeaders") {
		deps.Default.Config.MustGetStruct("clients.google-now.defaultHeaders", &defaultHeaders)
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
	if deps.Default.Config.ContainsKey("clients.google-now.followRedirect") {
		followRedirect = deps.Default.Config.MustGetBoolean("clients.google-now.followRedirect")
	}

	methodNames := map[string]string{
		"AddCredentials":   "GoogleNowService::addCredentials",
		"CheckCredentials": "GoogleNowService::checkCredentials",
	}
	// circuitBreakerDisabled sets whether circuit-breaker should be disabled
	circuitBreakerDisabled := false
	if deps.Default.Config.ContainsKey("clients.google-now.circuitBreakerDisabled") {
		circuitBreakerDisabled = deps.Default.Config.MustGetBoolean("clients.google-now.circuitBreakerDisabled")
	}

	//get mapping of client method and it's timeout
	//if mapping is not provided, use client's timeout for all methods
	clientMethodTimeoutMapping := make(map[string]int64)
	if deps.Default.Config.ContainsKey("clients.google-now.methodTimeoutMapping") {
		deps.Default.Config.MustGetStruct("clients.google-now.methodTimeoutMapping", &clientMethodTimeoutMapping)
	} else {
		//override the client overall-timeout with the client's method level timeout
		for methodName := range methodNames {
			clientMethodTimeoutMapping[methodName] = int64(timeoutVal)
		}
	}

	qpsLevels := map[string]string{
		"google-now-AddCredentials":   "2",
		"google-now-CheckCredentials": "1",
	}
	if !circuitBreakerDisabled {
		for methodName, methodTimeout := range clientMethodTimeoutMapping {
			circuitBreakerName := "google-now" + "-" + methodName
			qpsLevel := "default"
			if level, ok := qpsLevels[circuitBreakerName]; ok {
				qpsLevel = level
			}
			configureCircuitBreaker(deps, int(methodTimeout), circuitBreakerName, qpsLevel)
		}
	}

	return &googleNowClient{
		clientID: "google-now",
		httpClient: zanzibar.NewHTTPClientContext(
			deps.Default.ContextLogger, deps.Default.ContextMetrics, deps.Default.JSONWrapper,
			"google-now",
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

// CircuitBreakerConfig is used for storing the circuit breaker parameters for each qps level
type CircuitBreakerConfig struct {
	Parameters map[string]map[string]int
}

func configureCircuitBreaker(deps *module.Dependencies, timeoutVal int, circuitBreakerName string, qpsLevel string) {
	// sleepWindowInMilliseconds sets the amount of time, after tripping the circuit,
	// to reject requests before allowing attempts again to determine if the circuit should again be closed
	sleepWindowInMilliseconds := 5000
	// maxConcurrentRequests sets how many requests can be run at the same time, beyond which requests are rejected
	maxConcurrentRequests := 20
	// errorPercentThreshold sets the error percentage at or above which the circuit should trip open
	errorPercentThreshold := 20
	// requestVolumeThreshold sets a minimum number of requests that will trip the circuit in a rolling window of 10s
	// For example, if the value is 20, then if only 19 requests are received in the rolling window of 10 seconds
	// the circuit will not trip open even if all 19 failed.
	requestVolumeThreshold := 20
	// parses circuit breaker configurations
	if deps.Default.Config.ContainsKey(CircuitBreakerConfigKey) {
		var config CircuitBreakerConfig
		deps.Default.Config.MustGetStruct(CircuitBreakerConfigKey, &config)
		parameters := config.Parameters
		// first checks if level exists in configurations then assigns parameters
		// if "default" qps level assigns default parameters from circuit breaker configurations
		if settings, ok := parameters[qpsLevel]; ok {
			if sleep, ok := settings["sleepWindowInMilliseconds"]; ok {
				sleepWindowInMilliseconds = sleep
			}
			if max, ok := settings["maxConcurrentRequests"]; ok {
				maxConcurrentRequests = max
			}
			if errorPercent, ok := settings["errorPercentThreshold"]; ok {
				errorPercentThreshold = errorPercent
			}
			if reqVolThreshold, ok := settings["requestVolumeThreshold"]; ok {
				requestVolumeThreshold = reqVolThreshold
			}
		}
	}
	// client settings override parameters
	if deps.Default.Config.ContainsKey("clients.google-now.sleepWindowInMilliseconds") {
		sleepWindowInMilliseconds = int(deps.Default.Config.MustGetInt("clients.google-now.sleepWindowInMilliseconds"))
	}
	if deps.Default.Config.ContainsKey("clients.google-now.maxConcurrentRequests") {
		maxConcurrentRequests = int(deps.Default.Config.MustGetInt("clients.google-now.maxConcurrentRequests"))
	}
	if deps.Default.Config.ContainsKey("clients.google-now.errorPercentThreshold") {
		errorPercentThreshold = int(deps.Default.Config.MustGetInt("clients.google-now.errorPercentThreshold"))
	}
	if deps.Default.Config.ContainsKey("clients.google-now.requestVolumeThreshold") {
		requestVolumeThreshold = int(deps.Default.Config.MustGetInt("clients.google-now.requestVolumeThreshold"))
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
func (c *googleNowClient) HTTPClient() *zanzibar.HTTPClient {
	return c.httpClient
}

// AddCredentials calls "/add-credentials" endpoint.
func (c *googleNowClient) AddCredentials(
	ctx context.Context,
	headers map[string]string,
	r *clientsIDlClientsGooglenowGooglenow.GoogleNowService_AddCredentials_Args,
) (context.Context, map[string]string, error) {
	reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
	if headers == nil {
		headers = make(map[string]string)
	}
	if reqUUID != "" {
		headers[c.requestUUIDHeaderKey] = reqUUID
	}
	if c.requestProcedureHeaderKey != "" {
		headers[c.requestProcedureHeaderKey] = "GoogleNowService::addCredentials"
	}

	req := zanzibar.NewClientHTTPRequest(ctx, c.clientID, "AddCredentials", "GoogleNowService::addCredentials", c.httpClient)

	// Generate full URL.
	fullURL := c.httpClient.BaseURL + "/add-credentials"

	err := req.WriteJSON("POST", fullURL, headers, r)
	if err != nil {
		zanzibar.AppendLogFieldsToContext(ctx, zap.String("error", fmt.Sprintf("error creating outbound http request: %s", err)), logFieldErrLocation)
		return ctx, nil, err
	}

	headerErr := req.CheckHeaders([]string{"x-uuid"})
	if headerErr != nil {
		zanzibar.AppendLogFieldsToContext(ctx, zap.Error(headerErr), logFieldErrLocation)
		return ctx, nil, headerErr
	}

	var res *zanzibar.ClientHTTPResponse
	if c.circuitBreakerDisabled {
		res, err = req.Do()
	} else {
		// We want hystrix ckt-breaker to count errors only for system issues
		var clientErr error
		circuitBreakerName := "google-now" + "-" + "AddCredentials"
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
		zanzibar.AppendLogFieldsToContext(ctx, zap.String("error", fmt.Sprintf("error making http call: %s", err)), logFieldErrLocation)
		return ctx, nil, err
	}

	respHeaders := make(map[string]string)
	for k := range res.Header {
		respHeaders[k] = res.Header.Get(k)
	}

	defer func() {
		respHeaders[zanzibar.ClientResponseDurationKey] = res.Duration.String()
	}()
	// TODO(jakev): verify mandatory response headers

	res.CheckOKResponse([]int{202})

	switch res.StatusCode {
	case 202:
		_, err = res.ReadAll()
		if err != nil {
			return ctx, respHeaders, err
		}
		return ctx, respHeaders, nil
	default:
		_, err = res.ReadAll()
		if err != nil {
			return ctx, respHeaders, err
		}
	}

	return ctx, respHeaders, &zanzibar.UnexpectedHTTPError{
		StatusCode: res.StatusCode,
		RawBody:    res.GetRawBody(),
	}
}

// CheckCredentials calls "/check-credentials" endpoint.
func (c *googleNowClient) CheckCredentials(
	ctx context.Context,
	headers map[string]string,
) (context.Context, map[string]string, error) {
	reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
	if headers == nil {
		headers = make(map[string]string)
	}
	if reqUUID != "" {
		headers[c.requestUUIDHeaderKey] = reqUUID
	}
	if c.requestProcedureHeaderKey != "" {
		headers[c.requestProcedureHeaderKey] = "GoogleNowService::checkCredentials"
	}

	req := zanzibar.NewClientHTTPRequest(ctx, c.clientID, "CheckCredentials", "GoogleNowService::checkCredentials", c.httpClient)

	// Generate full URL.
	fullURL := c.httpClient.BaseURL + "/check-credentials"

	err := req.WriteJSON("POST", fullURL, headers, nil)
	if err != nil {
		zanzibar.AppendLogFieldsToContext(ctx, zap.String("error", fmt.Sprintf("error creating outbound http request: %s", err)), logFieldErrLocation)
		return ctx, nil, err
	}

	headerErr := req.CheckHeaders([]string{"x-uuid"})
	if headerErr != nil {
		zanzibar.AppendLogFieldsToContext(ctx, zap.Error(headerErr), logFieldErrLocation)
		return ctx, nil, headerErr
	}

	var res *zanzibar.ClientHTTPResponse
	if c.circuitBreakerDisabled {
		res, err = req.Do()
	} else {
		// We want hystrix ckt-breaker to count errors only for system issues
		var clientErr error
		circuitBreakerName := "google-now" + "-" + "CheckCredentials"
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
		zanzibar.AppendLogFieldsToContext(ctx, zap.String("error", fmt.Sprintf("error making http call: %s", err)), logFieldErrLocation)
		return ctx, nil, err
	}

	respHeaders := make(map[string]string)
	for k := range res.Header {
		respHeaders[k] = res.Header.Get(k)
	}

	defer func() {
		respHeaders[zanzibar.ClientResponseDurationKey] = res.Duration.String()
	}()
	// TODO(jakev): verify mandatory response headers

	res.CheckOKResponse([]int{202})

	switch res.StatusCode {
	case 202:
		_, err = res.ReadAll()
		if err != nil {
			return ctx, respHeaders, err
		}
		return ctx, respHeaders, nil
	default:
		_, err = res.ReadAll()
		if err != nil {
			return ctx, respHeaders, err
		}
	}

	return ctx, respHeaders, &zanzibar.UnexpectedHTTPError{
		StatusCode: res.StatusCode,
		RawBody:    res.GetRawBody(),
	}
}
