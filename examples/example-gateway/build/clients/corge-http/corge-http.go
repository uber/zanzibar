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

package corgehttpclient

import (
	"context"
	"fmt"
	"net/textproto"
	"regexp"
	"time"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/uber/zanzibar/v2/config"
	zanzibar "github.com/uber/zanzibar/v2/runtime"
	"github.com/uber/zanzibar/v2/runtime/jsonwrapper"

	module "github.com/uber/zanzibar/v2/examples/example-gateway/build/clients/corge-http/module"
	clientsIDlClientsCorgeCorge "github.com/uber/zanzibar/v2/examples/example-gateway/build/gen-code/clients-idl/clients/corge/corge"
)

// CircuitBreakerConfigKey is key value for qps level to circuit breaker parameters mapping
const CircuitBreakerConfigKey = "circuitbreaking-configurations"

// Client defines corge-http client interface.
type Client interface {
	HTTPClient() *zanzibar.HTTPClient

	EchoString(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsIDlClientsCorgeCorge.Corge_EchoString_Args,
	) (context.Context, string, map[string]string, error)
	NoContent(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsIDlClientsCorgeCorge.Corge_NoContent_Args,
	) (context.Context, map[string]string, error)
	NoContentNoException(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsIDlClientsCorgeCorge.Corge_NoContentNoException_Args,
	) (context.Context, map[string]string, error)
	CorgeNoContentOnException(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsIDlClientsCorgeCorge.Corge_NoContentOnException_Args,
	) (context.Context, *clientsIDlClientsCorgeCorge.Foo, map[string]string, error)
}

// corgeHTTPClient is the http client.
type corgeHTTPClient struct {
	clientID                  string
	httpClient                *zanzibar.HTTPClient
	jsonWrapper               jsonwrapper.JSONWrapper
	circuitBreakerDisabled    bool
	requestUUIDHeaderKey      string
	requestProcedureHeaderKey string

	calleeHeader  string
	callerHeader  string
	callerName    string
	calleeName    string
	altRoutingMap map[string]map[string]string
}

// NewClient returns a new http client.
func NewClient(deps *module.Dependencies) Client {
	ip := deps.Default.Config.MustGetString("sidecarRouter.default.http.ip")
	port := deps.Default.Config.MustGetInt("sidecarRouter.default.http.port")
	callerHeader := deps.Default.Config.MustGetString("sidecarRouter.default.http.callerHeader")
	calleeHeader := deps.Default.Config.MustGetString("sidecarRouter.default.http.calleeHeader")
	callerName := deps.Default.Config.MustGetString("serviceName")
	calleeName := deps.Default.Config.MustGetString("clients.corge-http.serviceName")

	var altServiceDetail = config.AlternateServiceDetail{}
	if deps.Default.Config.ContainsKey("clients.corge-http.alternates") {
		deps.Default.Config.MustGetStruct("clients.corge-http.alternates", &altServiceDetail)
	}

	baseURL := fmt.Sprintf("http://%s:%d", ip, port)
	timeoutVal := int(deps.Default.Config.MustGetInt("clients.corge-http.timeout"))
	timeout := time.Millisecond * time.Duration(
		timeoutVal,
	)
	defaultHeaders := make(map[string]string)
	if deps.Default.Config.ContainsKey("http.defaultHeaders") {
		deps.Default.Config.MustGetStruct("http.defaultHeaders", &defaultHeaders)
	}
	if deps.Default.Config.ContainsKey("clients.corge-http.defaultHeaders") {
		deps.Default.Config.MustGetStruct("clients.corge-http.defaultHeaders", &defaultHeaders)
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
	if deps.Default.Config.ContainsKey("clients.corge-http.followRedirect") {
		followRedirect = deps.Default.Config.MustGetBoolean("clients.corge-http.followRedirect")
	}

	methodNames := map[string]string{
		"EchoString":                "Corge::echoString",
		"NoContent":                 "Corge::noContent",
		"NoContentNoException":      "Corge::noContentNoException",
		"CorgeNoContentOnException": "Corge::noContentOnException",
	}
	// circuitBreakerDisabled sets whether circuit-breaker should be disabled
	circuitBreakerDisabled := false
	if deps.Default.Config.ContainsKey("clients.corge-http.circuitBreakerDisabled") {
		circuitBreakerDisabled = deps.Default.Config.MustGetBoolean("clients.corge-http.circuitBreakerDisabled")
	}

	//get mapping of client method and it's timeout
	//if mapping is not provided, use client's timeout for all methods
	clientMethodTimeoutMapping := make(map[string]int64)
	if deps.Default.Config.ContainsKey("clients.corge-http.methodTimeoutMapping") {
		deps.Default.Config.MustGetStruct("clients.corge-http.methodTimeoutMapping", &clientMethodTimeoutMapping)
	} else {
		//override the client overall-timeout with the client's method level timeout
		for methodName := range methodNames {
			clientMethodTimeoutMapping[methodName] = int64(timeoutVal)
		}
	}

	qpsLevels := map[string]string{
		"corge-http-CorgeNoContentOnException": "default",
		"corge-http-EchoString":                "default",
		"corge-http-NoContent":                 "default",
		"corge-http-NoContentNoException":      "default",
	}
	if !circuitBreakerDisabled {
		for methodName, methodTimeout := range clientMethodTimeoutMapping {
			circuitBreakerName := "corge-http" + "-" + methodName
			qpsLevel := "default"
			if level, ok := qpsLevels[circuitBreakerName]; ok {
				qpsLevel = level
			}
			configureCircuitBreaker(deps, int(methodTimeout), circuitBreakerName, qpsLevel)
		}
	}

	return &corgeHTTPClient{
		clientID:      "corge-http",
		callerHeader:  callerHeader,
		calleeHeader:  calleeHeader,
		callerName:    callerName,
		calleeName:    calleeName,
		altRoutingMap: initializeAltRoutingMap(altServiceDetail),
		httpClient: zanzibar.NewHTTPClientContext(
			deps.Default.ContextLogger, deps.Default.ContextMetrics, deps.Default.JSONWrapper,
			"corge-http",
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

func initializeAltRoutingMap(altServiceDetail config.AlternateServiceDetail) map[string]map[string]string {
	// The goal is to support for each header key, multiple values that point to different services
	routingMap := make(map[string]map[string]string)
	for _, alt := range altServiceDetail.RoutingConfigs {
		if headerValueToServiceMap, ok := routingMap[textproto.CanonicalMIMEHeaderKey(alt.HeaderName)]; ok {
			headerValueToServiceMap[alt.HeaderValue] = alt.ServiceName
		} else {
			routingMap[textproto.CanonicalMIMEHeaderKey(alt.HeaderName)] = map[string]string{alt.HeaderValue: alt.ServiceName}
		}
	}
	return routingMap
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
	if deps.Default.Config.ContainsKey("clients.corge-http.sleepWindowInMilliseconds") {
		sleepWindowInMilliseconds = int(deps.Default.Config.MustGetInt("clients.corge-http.sleepWindowInMilliseconds"))
	}
	if deps.Default.Config.ContainsKey("clients.corge-http.maxConcurrentRequests") {
		maxConcurrentRequests = int(deps.Default.Config.MustGetInt("clients.corge-http.maxConcurrentRequests"))
	}
	if deps.Default.Config.ContainsKey("clients.corge-http.errorPercentThreshold") {
		errorPercentThreshold = int(deps.Default.Config.MustGetInt("clients.corge-http.errorPercentThreshold"))
	}
	if deps.Default.Config.ContainsKey("clients.corge-http.requestVolumeThreshold") {
		requestVolumeThreshold = int(deps.Default.Config.MustGetInt("clients.corge-http.requestVolumeThreshold"))
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
func (c *corgeHTTPClient) HTTPClient() *zanzibar.HTTPClient {
	return c.httpClient
}

// EchoString calls "/echo/string" endpoint.
func (c *corgeHTTPClient) EchoString(
	ctx context.Context,
	headers map[string]string,
	r *clientsIDlClientsCorgeCorge.Corge_EchoString_Args,
) (context.Context, string, map[string]string, error) {
	reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
	if headers == nil {
		headers = make(map[string]string)
	}
	if reqUUID != "" {
		headers[c.requestUUIDHeaderKey] = reqUUID
	}
	if c.requestProcedureHeaderKey != "" {
		headers[c.requestProcedureHeaderKey] = "Corge::echoString"
	}

	var defaultRes string
	req := zanzibar.NewClientHTTPRequest(ctx, c.clientID, "EchoString", "Corge::echoString", c.httpClient)

	headers[c.callerHeader] = c.callerName

	// Set the service name if dynamic routing header is present
	for routeHeaderKey, routeMap := range c.altRoutingMap {
		if headerVal, ok := headers[routeHeaderKey]; ok {
			for routeRegex, altServiceName := range routeMap {
				//if headerVal matches routeRegex regex, set the alternative service name
				if matchFound, _ := regexp.MatchString(routeRegex, headerVal); matchFound {
					headers[c.calleeHeader] = altServiceName
					break
				}
			}
		}
	}

	// If serviceName was not set in the dynamic routing section above, set as the default
	if _, ok := headers[c.calleeHeader]; !ok {
		headers[c.calleeHeader] = c.calleeName
	}

	// Generate full URL.
	fullURL := c.httpClient.BaseURL + "/echo" + "/string"

	err := req.WriteJSON("POST", fullURL, headers, r)
	if err != nil {
		return ctx, defaultRes, nil, err
	}

	var res *zanzibar.ClientHTTPResponse
	if c.circuitBreakerDisabled {
		res, err = req.Do()
	} else {
		// We want hystrix ckt-breaker to count errors only for system issues
		var clientErr error
		circuitBreakerName := "corge-http" + "-" + "EchoString"
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
		return ctx, defaultRes, nil, err
	}

	respHeaders := make(map[string]string)
	for k := range res.Header {
		respHeaders[k] = res.Header.Get(k)
	}

	defer func() {
		respHeaders[zanzibar.ClientResponseDurationKey] = res.Duration.String()
	}()

	res.CheckOKResponse([]int{200})

	switch res.StatusCode {
	case 200:
		var responseBody string
		rawBody, err := res.ReadAll()
		if err != nil {
			return ctx, defaultRes, respHeaders, err
		}
		err = res.UnmarshalBody(&responseBody, rawBody)
		if err != nil {
			return ctx, defaultRes, respHeaders, err
		}

		return ctx, responseBody, respHeaders, nil
	default:
		_, err = res.ReadAll()
		if err != nil {
			return ctx, defaultRes, respHeaders, err
		}
	}

	return ctx, defaultRes, respHeaders, &zanzibar.UnexpectedHTTPError{
		StatusCode: res.StatusCode,
		RawBody:    res.GetRawBody(),
	}
}

// NoContent calls "/echo/no-content" endpoint.
func (c *corgeHTTPClient) NoContent(
	ctx context.Context,
	headers map[string]string,
	r *clientsIDlClientsCorgeCorge.Corge_NoContent_Args,
) (context.Context, map[string]string, error) {
	reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
	if headers == nil {
		headers = make(map[string]string)
	}
	if reqUUID != "" {
		headers[c.requestUUIDHeaderKey] = reqUUID
	}
	if c.requestProcedureHeaderKey != "" {
		headers[c.requestProcedureHeaderKey] = "Corge::noContent"
	}

	req := zanzibar.NewClientHTTPRequest(ctx, c.clientID, "NoContent", "Corge::noContent", c.httpClient)

	headers[c.callerHeader] = c.callerName

	// Set the service name if dynamic routing header is present
	for routeHeaderKey, routeMap := range c.altRoutingMap {
		if headerVal, ok := headers[routeHeaderKey]; ok {
			for routeRegex, altServiceName := range routeMap {
				//if headerVal matches routeRegex regex, set the alternative service name
				if matchFound, _ := regexp.MatchString(routeRegex, headerVal); matchFound {
					headers[c.calleeHeader] = altServiceName
					break
				}
			}
		}
	}

	// If serviceName was not set in the dynamic routing section above, set as the default
	if _, ok := headers[c.calleeHeader]; !ok {
		headers[c.calleeHeader] = c.calleeName
	}

	// Generate full URL.
	fullURL := c.httpClient.BaseURL + "/echo" + "/no-content"

	err := req.WriteJSON("POST", fullURL, headers, r)
	if err != nil {
		return ctx, nil, err
	}

	var res *zanzibar.ClientHTTPResponse
	if c.circuitBreakerDisabled {
		res, err = req.Do()
	} else {
		// We want hystrix ckt-breaker to count errors only for system issues
		var clientErr error
		circuitBreakerName := "corge-http" + "-" + "NoContent"
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
		return ctx, nil, err
	}

	respHeaders := make(map[string]string)
	for k := range res.Header {
		respHeaders[k] = res.Header.Get(k)
	}

	defer func() {
		respHeaders[zanzibar.ClientResponseDurationKey] = res.Duration.String()
	}()

	res.CheckOKResponse([]int{204, 304})

	switch res.StatusCode {
	case 204:

		return ctx, respHeaders, nil
	case 304:

		return ctx, respHeaders, &clientsIDlClientsCorgeCorge.NotModified{}

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

// NoContentNoException calls "/echo/no-content-no-exception" endpoint.
func (c *corgeHTTPClient) NoContentNoException(
	ctx context.Context,
	headers map[string]string,
	r *clientsIDlClientsCorgeCorge.Corge_NoContentNoException_Args,
) (context.Context, map[string]string, error) {
	reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
	if headers == nil {
		headers = make(map[string]string)
	}
	if reqUUID != "" {
		headers[c.requestUUIDHeaderKey] = reqUUID
	}
	if c.requestProcedureHeaderKey != "" {
		headers[c.requestProcedureHeaderKey] = "Corge::noContentNoException"
	}

	req := zanzibar.NewClientHTTPRequest(ctx, c.clientID, "NoContentNoException", "Corge::noContentNoException", c.httpClient)

	headers[c.callerHeader] = c.callerName

	// Set the service name if dynamic routing header is present
	for routeHeaderKey, routeMap := range c.altRoutingMap {
		if headerVal, ok := headers[routeHeaderKey]; ok {
			for routeRegex, altServiceName := range routeMap {
				//if headerVal matches routeRegex regex, set the alternative service name
				if matchFound, _ := regexp.MatchString(routeRegex, headerVal); matchFound {
					headers[c.calleeHeader] = altServiceName
					break
				}
			}
		}
	}

	// If serviceName was not set in the dynamic routing section above, set as the default
	if _, ok := headers[c.calleeHeader]; !ok {
		headers[c.calleeHeader] = c.calleeName
	}

	// Generate full URL.
	fullURL := c.httpClient.BaseURL + "/echo" + "/no-content-no-exception"

	err := req.WriteJSON("POST", fullURL, headers, r)
	if err != nil {
		return ctx, nil, err
	}

	var res *zanzibar.ClientHTTPResponse
	if c.circuitBreakerDisabled {
		res, err = req.Do()
	} else {
		// We want hystrix ckt-breaker to count errors only for system issues
		var clientErr error
		circuitBreakerName := "corge-http" + "-" + "NoContentNoException"
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
		return ctx, nil, err
	}

	respHeaders := make(map[string]string)
	for k := range res.Header {
		respHeaders[k] = res.Header.Get(k)
	}

	defer func() {
		respHeaders[zanzibar.ClientResponseDurationKey] = res.Duration.String()
	}()

	res.CheckOKResponse([]int{204})

	switch res.StatusCode {
	case 204:
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

// CorgeNoContentOnException calls "/echo/no-content-on-exception" endpoint.
func (c *corgeHTTPClient) CorgeNoContentOnException(
	ctx context.Context,
	headers map[string]string,
	r *clientsIDlClientsCorgeCorge.Corge_NoContentOnException_Args,
) (context.Context, *clientsIDlClientsCorgeCorge.Foo, map[string]string, error) {
	reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
	if headers == nil {
		headers = make(map[string]string)
	}
	if reqUUID != "" {
		headers[c.requestUUIDHeaderKey] = reqUUID
	}
	if c.requestProcedureHeaderKey != "" {
		headers[c.requestProcedureHeaderKey] = "Corge::noContentOnException"
	}

	var defaultRes *clientsIDlClientsCorgeCorge.Foo
	req := zanzibar.NewClientHTTPRequest(ctx, c.clientID, "CorgeNoContentOnException", "Corge::noContentOnException", c.httpClient)

	headers[c.callerHeader] = c.callerName

	// Set the service name if dynamic routing header is present
	for routeHeaderKey, routeMap := range c.altRoutingMap {
		if headerVal, ok := headers[routeHeaderKey]; ok {
			for routeRegex, altServiceName := range routeMap {
				//if headerVal matches routeRegex regex, set the alternative service name
				if matchFound, _ := regexp.MatchString(routeRegex, headerVal); matchFound {
					headers[c.calleeHeader] = altServiceName
					break
				}
			}
		}
	}

	// If serviceName was not set in the dynamic routing section above, set as the default
	if _, ok := headers[c.calleeHeader]; !ok {
		headers[c.calleeHeader] = c.calleeName
	}

	// Generate full URL.
	fullURL := c.httpClient.BaseURL + "/echo" + "/no-content-on-exception"

	err := req.WriteJSON("POST", fullURL, headers, r)
	if err != nil {
		return ctx, defaultRes, nil, err
	}

	var res *zanzibar.ClientHTTPResponse
	if c.circuitBreakerDisabled {
		res, err = req.Do()
	} else {
		// We want hystrix ckt-breaker to count errors only for system issues
		var clientErr error
		circuitBreakerName := "corge-http" + "-" + "CorgeNoContentOnException"
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
		return ctx, defaultRes, nil, err
	}

	respHeaders := make(map[string]string)
	for k := range res.Header {
		respHeaders[k] = res.Header.Get(k)
	}

	defer func() {
		respHeaders[zanzibar.ClientResponseDurationKey] = res.Duration.String()
	}()

	res.CheckOKResponse([]int{200, 304})

	switch res.StatusCode {
	case 200:
		var responseBody clientsIDlClientsCorgeCorge.Foo
		rawBody, err := res.ReadAll()
		if err != nil {
			return ctx, defaultRes, respHeaders, err
		}
		err = res.UnmarshalBody(&responseBody, rawBody)
		if err != nil {
			return ctx, defaultRes, respHeaders, err
		}

		return ctx, &responseBody, respHeaders, nil

	case 304:

		return ctx, defaultRes, respHeaders, &clientsIDlClientsCorgeCorge.NotModified{}

	default:
		_, err = res.ReadAll()
		if err != nil {
			return ctx, defaultRes, respHeaders, err
		}
	}

	return ctx, defaultRes, respHeaders, &zanzibar.UnexpectedHTTPError{
		StatusCode: res.StatusCode,
		RawBody:    res.GetRawBody(),
	}
}
