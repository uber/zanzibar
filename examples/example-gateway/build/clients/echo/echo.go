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

package echoclient

import (
	"context"

	"github.com/isopropylcyanide/hystrix-go/hystrix"
	"go.uber.org/yarpc"

	module "github.com/uber/zanzibar/examples/example-gateway/build/clients/echo/module"
	gen "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/echo"

	zanzibar "github.com/uber/zanzibar/runtime"
)

// Client defines echo client interface.
type Client interface {
	EchoEcho(
		ctx context.Context,
		request *gen.Request,
		opts ...yarpc.CallOption,
	) (*gen.Response, error)
}

// echoClient is the gRPC client for downstream service.
type echoClient struct {
	echoClient gen.EchoYARPCClient
	opts       *zanzibar.GRPCClientOpts
}

// NewClient returns a new gRPC client for service echo
func NewClient(deps *module.Dependencies) Client {
	oc := deps.Default.GRPCClientDispatcher.MustOutboundConfig("echo")
	var routingKey string
	if deps.Default.Config.ContainsKey("clients.echo.routingKey") {
		routingKey = deps.Default.Config.MustGetString("clients.echo.routingKey")
	}
	var requestUUIDHeaderKey string
	if deps.Default.Config.ContainsKey("clients.echo.requestUUIDHeaderKey") {
		requestUUIDHeaderKey = deps.Default.Config.MustGetString("clients.echo.requestUUIDHeaderKey")
	}
	timeoutInMS := int(deps.Default.Config.MustGetInt("clients.echo.timeout"))
	methodNames := map[string]string{
		"Echo::Echo": "EchoEcho",
	}

	// circuitBreakerDisabled sets whether circuit-breaker should be disabled
	circuitBreakerDisabled := false
	if deps.Default.Config.ContainsKey("clients.echo.circuitBreakerDisabled") {
		circuitBreakerDisabled = deps.Default.Config.MustGetBoolean("clients.echo.circuitBreakerDisabled")
	}
	if !circuitBreakerDisabled {
		for _, methodName := range methodNames {
			circuitBreakerName := "echo" + "-" + methodName
			configureCircuitBreaker(deps, timeoutInMS, circuitBreakerName)
		}
	}

	return &echoClient{
		echoClient: gen.NewEchoYARPCClient(oc),
		opts: zanzibar.NewGRPCClientOpts(
			deps.Default.ContextLogger,
			deps.Default.ContextMetrics,
			deps.Default.ContextExtractor,
			methodNames,
			"echo",
			routingKey,
			requestUUIDHeaderKey,
			circuitBreakerDisabled,
			timeoutInMS,
		),
	}
}

func configureCircuitBreaker(deps *module.Dependencies, timeoutVal int, circuitBreakerName string) {
	// sleepWindowInMilliseconds sets the amount of time, after tripping the circuit,
	// to reject requests before allowing attempts again to determine if the circuit should again be closed
	sleepWindowInMilliseconds := 5000
	if deps.Default.Config.ContainsKey("clients.echo.sleepWindowInMilliseconds") {
		sleepWindowInMilliseconds = int(deps.Default.Config.MustGetInt("clients.echo.sleepWindowInMilliseconds"))
	}
	// maxConcurrentRequests sets how many requests can be run at the same time, beyond which requests are rejected
	maxConcurrentRequests := 20
	if deps.Default.Config.ContainsKey("clients.echo.maxConcurrentRequests") {
		maxConcurrentRequests = int(deps.Default.Config.MustGetInt("clients.echo.maxConcurrentRequests"))
	}
	// errorPercentThreshold sets the error percentage at or above which the circuit should trip open
	errorPercentThreshold := 20
	if deps.Default.Config.ContainsKey("clients.echo.errorPercentThreshold") {
		errorPercentThreshold = int(deps.Default.Config.MustGetInt("clients.echo.errorPercentThreshold"))
	}
	// requestVolumeThreshold sets a minimum number of requests that will trip the circuit in a rolling window of 10s
	// For example, if the value is 20, then if only 19 requests are received in the rolling window of 10 seconds
	// the circuit will not trip open even if all 19 failed.
	requestVolumeThreshold := 20
	if deps.Default.Config.ContainsKey("clients.echo.requestVolumeThreshold") {
		requestVolumeThreshold = int(deps.Default.Config.MustGetInt("clients.echo.requestVolumeThreshold"))
	}
	hystrix.ConfigureCommand(circuitBreakerName, hystrix.CommandConfig{
		MaxConcurrentRequests:  maxConcurrentRequests,
		ErrorPercentThreshold:  errorPercentThreshold,
		SleepWindow:            sleepWindowInMilliseconds,
		RequestVolumeThreshold: requestVolumeThreshold,
		Timeout:                timeoutVal,
	})
}

// EchoEcho is a client RPC call for method Echo::Echo.
func (e *echoClient) EchoEcho(
	ctx context.Context,
	request *gen.Request,
	opts ...yarpc.CallOption,
) (*gen.Response, error) {
	var result *gen.Response
	var err error

	ctx, callHelper := zanzibar.NewGRPCClientCallHelper(ctx, "Echo::Echo", e.opts)

	if e.opts.RoutingKey != "" {
		opts = append(opts, yarpc.WithRoutingKey(e.opts.RoutingKey))
	}
	if e.opts.RequestUUIDHeaderKey != "" {
		reqUUID := zanzibar.RequestUUIDFromCtx(ctx)
		if reqUUID != "" {
			opts = append(opts, yarpc.WithHeader(e.opts.RequestUUIDHeaderKey, reqUUID))
		}
	}
	ctx, cancel := context.WithTimeout(ctx, e.opts.Timeout)
	defer cancel()

	runFunc := e.echoClient.Echo
	callHelper.Start()
	if e.opts.CircuitBreakerDisabled {
		result, err = runFunc(ctx, request, opts...)
	} else {
		circuitBreakerName := "echo" + "-" + "EchoEcho"
		err = hystrix.DoC(ctx, circuitBreakerName, func(ctx context.Context) error {
			result, err = runFunc(ctx, request, opts...)
			return err
		}, nil)
	}
	callHelper.Finish(ctx, err)

	return result, err
}
