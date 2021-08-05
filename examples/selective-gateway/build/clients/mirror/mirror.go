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

package mirrorclient

import (
	"context"

	"github.com/afex/hystrix-go/hystrix"
	"go.uber.org/yarpc"

	module "github.com/uber/zanzibar/examples/selective-gateway/build/clients/mirror/module"
	gen "github.com/uber/zanzibar/examples/selective-gateway/build/proto-gen/clients/mirror"

	zanzibar "github.com/uber/zanzibar/runtime"
)

// CircuitBreakerConfigKey is key value for qps level to circuit breaker parameters mapping
const CircuitBreakerConfigKey = "circuitbreaking-configurations"

// Client defines mirror client interface.
type Client interface {
	MirrorMirror(
		ctx context.Context,
		request *gen.Request,
		opts ...yarpc.CallOption,
	) (*gen.Response, error)

	MirrorInternalMirror(
		ctx context.Context,
		request *gen.InternalRequest,
		opts ...yarpc.CallOption,
	) (*gen.InternalResponse, error)
}

// mirrorClient is the gRPC client for downstream service.
type mirrorClient struct {
	mirrorClient         gen.MirrorYARPCClient
	mirrorInternalClient gen.MirrorInternalYARPCClient
	opts                 *zanzibar.GRPCClientOpts
}

// NewClient returns a new gRPC client for service mirror
func NewClient(deps *module.Dependencies) Client {
	oc := deps.Default.GRPCClientDispatcher.MustOutboundConfig("mirror")
	var routingKey string
	if deps.Default.Config.ContainsKey("clients.mirror.routingKey") {
		routingKey = deps.Default.Config.MustGetString("clients.mirror.routingKey")
	}
	var requestUUIDHeaderKey string
	if deps.Default.Config.ContainsKey("clients.mirror.requestUUIDHeaderKey") {
		requestUUIDHeaderKey = deps.Default.Config.MustGetString("clients.mirror.requestUUIDHeaderKey")
	}
	timeoutInMS := int(deps.Default.Config.MustGetInt("clients.mirror.timeout"))
	methodNames := map[string]string{
		"Mirror::Mirror":         "MirrorMirror",
		"MirrorInternal::Mirror": "MirrorInternalMirror",
	}

	qpsLevels := map[string]string{
		"mirror-MirrorInternalMirror": "default",
		"mirror-MirrorMirror":         "default",
	}

	// circuitBreakerDisabled sets whether circuit-breaker should be disabled
	circuitBreakerDisabled := false
	if deps.Default.Config.ContainsKey("clients.mirror.circuitBreakerDisabled") {
		circuitBreakerDisabled = deps.Default.Config.MustGetBoolean("clients.mirror.circuitBreakerDisabled")
	}
	if !circuitBreakerDisabled {
		for _, methodName := range methodNames {
			circuitBreakerName := "mirror" + "-" + methodName
			qpsLevel := "default"
			if level, ok := qpsLevels[circuitBreakerName]; ok {
				qpsLevel = level
			}
			configureCircuitBreaker(deps, timeoutInMS, circuitBreakerName, qpsLevel)
		}
	}

	return &mirrorClient{
		mirrorClient:         gen.NewMirrorYARPCClient(oc),
		mirrorInternalClient: gen.NewMirrorInternalYARPCClient(oc),
		opts: zanzibar.NewGRPCClientOpts(
			deps.Default.ContextLogger,
			deps.Default.ContextMetrics,
			deps.Default.ContextExtractor,
			methodNames,
			"mirror",
			routingKey,
			requestUUIDHeaderKey,
			circuitBreakerDisabled,
			timeoutInMS,
		),
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
	if deps.Default.Config.ContainsKey("clients.mirror.sleepWindowInMilliseconds") {
		sleepWindowInMilliseconds = int(deps.Default.Config.MustGetInt("clients.mirror.sleepWindowInMilliseconds"))
	}
	if deps.Default.Config.ContainsKey("clients.mirror.maxConcurrentRequests") {
		maxConcurrentRequests = int(deps.Default.Config.MustGetInt("clients.mirror.maxConcurrentRequests"))
	}
	if deps.Default.Config.ContainsKey("clients.mirror.errorPercentThreshold") {
		errorPercentThreshold = int(deps.Default.Config.MustGetInt("clients.mirror.errorPercentThreshold"))
	}
	if deps.Default.Config.ContainsKey("clients.mirror.requestVolumeThreshold") {
		requestVolumeThreshold = int(deps.Default.Config.MustGetInt("clients.mirror.requestVolumeThreshold"))
	}
	hystrix.ConfigureCommand(circuitBreakerName, hystrix.CommandConfig{
		MaxConcurrentRequests:  maxConcurrentRequests,
		ErrorPercentThreshold:  errorPercentThreshold,
		SleepWindow:            sleepWindowInMilliseconds,
		RequestVolumeThreshold: requestVolumeThreshold,
		Timeout:                timeoutVal,
	})
}

// MirrorMirror is a client RPC call for method Mirror::Mirror.
func (e *mirrorClient) MirrorMirror(
	ctx context.Context,
	request *gen.Request,
	opts ...yarpc.CallOption,
) (*gen.Response, error) {
	var result *gen.Response
	var err error

	ctx, callHelper := zanzibar.NewGRPCClientCallHelper(ctx, "Mirror::Mirror", e.opts)

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

	runFunc := e.mirrorClient.Mirror
	callHelper.Start()
	if e.opts.CircuitBreakerDisabled {
		result, err = runFunc(ctx, request, opts...)
	} else {
		circuitBreakerName := "mirror" + "-" + "MirrorMirror"
		err = hystrix.DoC(ctx, circuitBreakerName, func(ctx context.Context) error {
			result, err = runFunc(ctx, request, opts...)
			return err
		}, nil)
	}
	callHelper.Finish(ctx, err)

	return result, err
}

// MirrorInternalMirror is a client RPC call for method MirrorInternal::Mirror.
func (e *mirrorClient) MirrorInternalMirror(
	ctx context.Context,
	request *gen.InternalRequest,
	opts ...yarpc.CallOption,
) (*gen.InternalResponse, error) {
	var result *gen.InternalResponse
	var err error

	ctx, callHelper := zanzibar.NewGRPCClientCallHelper(ctx, "MirrorInternal::Mirror", e.opts)

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

	runFunc := e.mirrorInternalClient.Mirror
	callHelper.Start()
	if e.opts.CircuitBreakerDisabled {
		result, err = runFunc(ctx, request, opts...)
	} else {
		circuitBreakerName := "mirror" + "-" + "MirrorInternalMirror"
		err = hystrix.DoC(ctx, circuitBreakerName, func(ctx context.Context) error {
			result, err = runFunc(ctx, request, opts...)
			return err
		}, nil)
	}
	callHelper.Finish(ctx, err)

	return result, err
}
