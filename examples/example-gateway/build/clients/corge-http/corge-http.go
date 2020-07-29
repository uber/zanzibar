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
	"regexp"
	"time"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/uber/zanzibar/config"
	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/runtime/jsonwrapper"

	module "github.com/uber/zanzibar/examples/example-gateway/build/clients/corge-http/module"
	clientsIDlClientsCorgeCorge "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/corge/corge"
)

// Client defines corge-http client interface.
type Client interface {
	HTTPClient() *zanzibar.HTTPClient

	EchoString(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsIDlClientsCorgeCorge.Corge_EchoString_Args,
	) (string, map[string]string, error)
	NoContent(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsIDlClientsCorgeCorge.Corge_NoContent_Args,
	) (map[string]string, error)
	NoContentNoException(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsIDlClientsCorgeCorge.Corge_NoContentNoException_Args,
	) (map[string]string, error)
	CorgeNoContentOnException(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsIDlClientsCorgeCorge.Corge_NoContentOnException_Args,
	) (*clientsIDlClientsCorgeCorge.Foo, map[string]string, error)
}

// corgeHTTPClient is the http client.
type corgeHTTPClient struct {
	clientID               string
	httpClient             *zanzibar.HTTPClient
	jsonWrapper            jsonwrapper.JSONWrapper
	circuitBreakerDisabled bool
	requestUUIDHeaderKey   string

	calleeHeader        string
	callerHeader        string
	callerName          string
	calleeName          string
	alternateRoutingMap map[string]map[string]string
}

// NewClient returns a new http client.
func NewClient(deps *module.Dependencies) Client {
	ip := deps.Default.Config.MustGetString("sidecarRouter.default.http.ip")
	port := deps.Default.Config.MustGetInt("sidecarRouter.default.http.port")
	callerHeader := deps.Default.Config.MustGetString("sidecarRouter.default.http.callerHeader")
	calleeHeader := deps.Default.Config.MustGetString("sidecarRouter.default.http.calleeHeader")
	callerName := deps.Default.Config.MustGetString("serviceName")
	calleeName := deps.Default.Config.MustGetString("clients.corge-http.serviceName")

	var alternateServiceDetail = config.AlternateServiceDetail{}
	if deps.Default.Config.ContainsKey("clients.corge-http.alternates") {
		var alternateServiceDetail config.AlternateServiceDetail
		deps.Default.Config.MustGetStruct("clients.corge-http.alternates", &alternateServiceDetail)
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
	followRedirect := true
	if deps.Default.Config.ContainsKey("clients.corge-http.followRedirect") {
		followRedirect = deps.Default.Config.MustGetBoolean("clients.corge-http.followRedirect")
	}

	circuitBreakerDisabled := configureCicruitBreaker(deps, timeoutVal)

	return &corgeHTTPClient{
		clientID:            "corge-http",
		callerHeader:        callerHeader,
		calleeHeader:        calleeHeader,
		callerName:          callerName,
		calleeName:          calleeName,
		alternateRoutingMap: initializeAlternateRoutingMap(alternateServiceDetail),
		httpClient: zanzibar.NewHTTPClientContext(
			deps.Default.Logger, deps.Default.ContextMetrics, deps.Default.JSONWrapper,
			"corge-http",
			map[string]string{
				"EchoString":                "Corge::echoString",
				"NoContent":                 "Corge::noContent",
				"NoContentNoException":      "Corge::noContentNoException",
				"CorgeNoContentOnException": "Corge::noContentOnException",
			},
			baseURL,
			defaultHeaders,
			timeout,
			followRedirect,
		),
		circuitBreakerDisabled: circuitBreakerDisabled,
		requestUUIDHeaderKey:   requestUUIDHeaderKey,
	}
}

func initializeAlternateRoutingMap(altServiceDetail config.AlternateServiceDetail) map[string]map[string]string {
	routingMap := make(map[string]map[string]string)
	for _, alt := range altServiceDetail.RoutingConfigs {
		if headerValueToServiceMap, ok := routingMap[alt.HeaderName]; ok {
			headerValueToServiceMap[alt.HeaderValue] = alt.ServiceName
		} else {
			routingMap[alt.HeaderName] = map[string]string{alt.HeaderValue: alt.ServiceName}
		}
	}
	return routingMap
}
func configureCicruitBreaker(deps *module.Dependencies, timeoutVal int) bool {
	// circuitBreakerDisabled sets whether circuit-breaker should be disabled
	circuitBreakerDisabled := false
	if deps.Default.Config.ContainsKey("clients.corge-http.circuitBreakerDisabled") {
		circuitBreakerDisabled = deps.Default.Config.MustGetBoolean("clients.corge-http.circuitBreakerDisabled")
	}
	// sleepWindowInMilliseconds sets the amount of time, after tripping the circuit,
	// to reject requests before allowing attempts again to determine if the circuit should again be closed
	sleepWindowInMilliseconds := 5000
	if deps.Default.Config.ContainsKey("clients.corge-http.sleepWindowInMilliseconds") {
		sleepWindowInMilliseconds = int(deps.Default.Config.MustGetInt("clients.corge-http.sleepWindowInMilliseconds"))
	}
	// maxConcurrentRequests sets how many requests can be run at the same time, beyond which requests are rejected
	maxConcurrentRequests := 20
	if deps.Default.Config.ContainsKey("clients.corge-http.maxConcurrentRequests") {
		maxConcurrentRequests = int(deps.Default.Config.MustGetInt("clients.corge-http.maxConcurrentRequests"))
	}
	// errorPercentThreshold sets the error percentage at or above which the circuit should trip open
	errorPercentThreshold := 20
	if deps.Default.Config.ContainsKey("clients.corge-http.errorPercentThreshold") {
		errorPercentThreshold = int(deps.Default.Config.MustGetInt("clients.corge-http.errorPercentThreshold"))
	}
	// requestVolumeThreshold sets a minimum number of requests that will trip the circuit in a rolling window of 10s
	// For example, if the value is 20, then if only 19 requests are received in the rolling window of 10 seconds
	// the circuit will not trip open even if all 19 failed.
	requestVolumeThreshold := 20
	if deps.Default.Config.ContainsKey("clients.corge-http.requestVolumeThreshold") {
		requestVolumeThreshold = int(deps.Default.Config.MustGetInt("clients.corge-http.requestVolumeThreshold"))
	}
	if !circuitBreakerDisabled {
		hystrix.ConfigureCommand("corge-http", hystrix.CommandConfig{
			MaxConcurrentRequests:  maxConcurrentRequests,
			ErrorPercentThreshold:  errorPercentThreshold,
			SleepWindow:            sleepWindowInMilliseconds,
			RequestVolumeThreshold: requestVolumeThreshold,
			Timeout:                timeoutVal,
		})
	}
	return circuitBreakerDisabled
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
) (string, map[string]string, error) {
	reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
	if reqUUID != "" {
		if headers == nil {
			headers = make(map[string]string)
		}
		headers[c.requestUUIDHeaderKey] = reqUUID
	}

	var defaultRes string
	req := zanzibar.NewClientHTTPRequest(ctx, c.clientID, "EchoString", "Corge::echoString", c.httpClient)

	headers[c.callerHeader] = c.callerName

	// Set the service name if dynamic routing header is present
	for headerKey, headerVal := range headers {
		if routeMap, ok := c.alternateRoutingMap[headerKey]; ok {
			for routeRegex, altServiceName := range routeMap {
				//if headerVal matches routeRegex regex, set the alternative service name
				if matchFound, _ := regexp.MatchString(routeRegex, headerVal); matchFound {
					headers[c.calleeHeader] = altServiceName
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
		return defaultRes, nil, err
	}

	var res *zanzibar.ClientHTTPResponse
	if c.circuitBreakerDisabled {
		res, err = req.Do()
	} else {
		// We want hystrix ckt-breaker to count errors only for system issues
		var clientErr error
		err = hystrix.DoC(ctx, "corge-http", func(ctx context.Context) error {
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

// NoContent calls "/echo/no-content" endpoint.
func (c *corgeHTTPClient) NoContent(
	ctx context.Context,
	headers map[string]string,
	r *clientsIDlClientsCorgeCorge.Corge_NoContent_Args,
) (map[string]string, error) {
	reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
	if reqUUID != "" {
		if headers == nil {
			headers = make(map[string]string)
		}
		headers[c.requestUUIDHeaderKey] = reqUUID
	}

	req := zanzibar.NewClientHTTPRequest(ctx, c.clientID, "NoContent", "Corge::noContent", c.httpClient)

	headers[c.callerHeader] = c.callerName

	// Set the service name if dynamic routing header is present
	for headerKey, headerVal := range headers {
		if routeMap, ok := c.alternateRoutingMap[headerKey]; ok {
			for routeRegex, altServiceName := range routeMap {
				//if headerVal matches routeRegex regex, set the alternative service name
				if matchFound, _ := regexp.MatchString(routeRegex, headerVal); matchFound {
					headers[c.calleeHeader] = altServiceName
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
		return nil, err
	}

	var res *zanzibar.ClientHTTPResponse
	if c.circuitBreakerDisabled {
		res, err = req.Do()
	} else {
		// We want hystrix ckt-breaker to count errors only for system issues
		var clientErr error
		err = hystrix.DoC(ctx, "corge-http", func(ctx context.Context) error {
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
		return nil, err
	}

	respHeaders := make(map[string]string)
	for k := range res.Header {
		respHeaders[k] = res.Header.Get(k)
	}

	res.CheckOKResponse([]int{204, 304})

	switch res.StatusCode {
	case 204:

		return respHeaders, nil
	case 304:

		return respHeaders, &clientsIDlClientsCorgeCorge.NotModified{}

	default:
		_, err = res.ReadAll()
		if err != nil {
			return respHeaders, err
		}
	}

	return respHeaders, &zanzibar.UnexpectedHTTPError{
		StatusCode: res.StatusCode,
		RawBody:    res.GetRawBody(),
	}
}

// NoContentNoException calls "/echo/no-content-no-exception" endpoint.
func (c *corgeHTTPClient) NoContentNoException(
	ctx context.Context,
	headers map[string]string,
	r *clientsIDlClientsCorgeCorge.Corge_NoContentNoException_Args,
) (map[string]string, error) {
	reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
	if reqUUID != "" {
		if headers == nil {
			headers = make(map[string]string)
		}
		headers[c.requestUUIDHeaderKey] = reqUUID
	}

	req := zanzibar.NewClientHTTPRequest(ctx, c.clientID, "NoContentNoException", "Corge::noContentNoException", c.httpClient)

	headers[c.callerHeader] = c.callerName

	// Set the service name if dynamic routing header is present
	for headerKey, headerVal := range headers {
		if routeMap, ok := c.alternateRoutingMap[headerKey]; ok {
			for routeRegex, altServiceName := range routeMap {
				//if headerVal matches routeRegex regex, set the alternative service name
				if matchFound, _ := regexp.MatchString(routeRegex, headerVal); matchFound {
					headers[c.calleeHeader] = altServiceName
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
		return nil, err
	}

	var res *zanzibar.ClientHTTPResponse
	if c.circuitBreakerDisabled {
		res, err = req.Do()
	} else {
		// We want hystrix ckt-breaker to count errors only for system issues
		var clientErr error
		err = hystrix.DoC(ctx, "corge-http", func(ctx context.Context) error {
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
		return nil, err
	}

	respHeaders := make(map[string]string)
	for k := range res.Header {
		respHeaders[k] = res.Header.Get(k)
	}

	res.CheckOKResponse([]int{204})

	switch res.StatusCode {
	case 204:
		return respHeaders, nil
	default:
		_, err = res.ReadAll()
		if err != nil {
			return respHeaders, err
		}
	}

	return respHeaders, &zanzibar.UnexpectedHTTPError{
		StatusCode: res.StatusCode,
		RawBody:    res.GetRawBody(),
	}
}

// CorgeNoContentOnException calls "/echo/no-content-on-exception" endpoint.
func (c *corgeHTTPClient) CorgeNoContentOnException(
	ctx context.Context,
	headers map[string]string,
	r *clientsIDlClientsCorgeCorge.Corge_NoContentOnException_Args,
) (*clientsIDlClientsCorgeCorge.Foo, map[string]string, error) {
	reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
	if reqUUID != "" {
		if headers == nil {
			headers = make(map[string]string)
		}
		headers[c.requestUUIDHeaderKey] = reqUUID
	}

	var defaultRes *clientsIDlClientsCorgeCorge.Foo
	req := zanzibar.NewClientHTTPRequest(ctx, c.clientID, "CorgeNoContentOnException", "Corge::noContentOnException", c.httpClient)

	headers[c.callerHeader] = c.callerName

	// Set the service name if dynamic routing header is present
	for headerKey, headerVal := range headers {
		if routeMap, ok := c.alternateRoutingMap[headerKey]; ok {
			for routeRegex, altServiceName := range routeMap {
				//if headerVal matches routeRegex regex, set the alternative service name
				if matchFound, _ := regexp.MatchString(routeRegex, headerVal); matchFound {
					headers[c.calleeHeader] = altServiceName
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
		return defaultRes, nil, err
	}

	var res *zanzibar.ClientHTTPResponse
	if c.circuitBreakerDisabled {
		res, err = req.Do()
	} else {
		// We want hystrix ckt-breaker to count errors only for system issues
		var clientErr error
		err = hystrix.DoC(ctx, "corge-http", func(ctx context.Context) error {
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

	res.CheckOKResponse([]int{200, 304})

	switch res.StatusCode {
	case 200:
		var responseBody clientsIDlClientsCorgeCorge.Foo
		rawBody, err := res.ReadAll()
		if err != nil {
			return defaultRes, respHeaders, err
		}
		err = res.UnmarshalBody(&responseBody, rawBody)
		if err != nil {
			return defaultRes, respHeaders, err
		}

		return &responseBody, respHeaders, nil

	case 304:

		return defaultRes, respHeaders, &clientsIDlClientsCorgeCorge.NotModified{}

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
