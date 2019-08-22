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
	"context"
	"strings"
	"time"

	"go.uber.org/yarpc/yarpcerrors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// YARPCClientOpts used to configure various client options.
type YARPCClientOpts struct {
	ServiceName            string
	ClientID               string
	MethodNames            map[string]string
	Loggers                map[string]*zap.Logger
	Metrics                ContextMetrics
	ContextExtractor       ContextExtractor
	RoutingKey             string
	RequestUUIDHeaderKey   string
	CircuitBreakerDisabled bool
	Timeout                time.Duration
	ScopeTags              map[string]map[string]string
}

// NewYARPCClientOpts creates a new instance of YARPCClientOpts.
func NewYARPCClientOpts(
	logger *zap.Logger,
	metrics ContextMetrics,
	contextExtractor ContextExtractor,
	methodNames map[string]string,
	clientID, serviceName, routingKey, requestUUIDHeaderKey string,
	circuitBreakerDisabled bool,
	timeoutInMS int,
) *YARPCClientOpts {
	numMethods := len(methodNames)
	loggers := make(map[string]*zap.Logger, numMethods)
	for serviceMethod, methodName := range methodNames {
		loggers[serviceMethod] = logger.With(
			zap.String(logFieldClientID, clientID),
			zap.String(logFieldClientMethod, methodName),
			zap.String(logFieldClientService, serviceName),
		)
	}
	scopeTags := make(map[string]map[string]string)
	for serviceMethod, methodName := range methodNames {
		scopeTags[serviceMethod] = map[string]string{
			scopeTagClient:         clientID,
			scopeTagClientMethod:   methodName,
			scopeTagsTargetService: serviceName,
		}
	}
	return &YARPCClientOpts{
		ServiceName:            serviceName,
		ClientID:               clientID,
		MethodNames:            methodNames,
		Loggers:                loggers,
		Metrics:                metrics,
		ContextExtractor:       contextExtractor,
		RoutingKey:             routingKey,
		RequestUUIDHeaderKey:   requestUUIDHeaderKey,
		CircuitBreakerDisabled: circuitBreakerDisabled,
		Timeout:                time.Duration(timeoutInMS) * time.Millisecond,
		ScopeTags:              scopeTags,
	}
}

// YARPCClientCallHelper is used to track internal state of logging and metrics.
type YARPCClientCallHelper interface {
	// Start method should be used just before calling the actual YARPC client method call.
	Start()
	// Finish method should be used right after the actual call to YARPC client method.
	Finish(ctx context.Context, err error) context.Context
}

type callHelper struct {
	startTime  time.Time
	finishTime time.Time
	logger     *zap.Logger
	metrics    ContextMetrics
	extractor  ContextExtractor
}

// NewYARPCClientCallHelper used to initialize a helper that will
// be used to track logging and metric for a YARPC Client call.
func NewYARPCClientCallHelper(ctx context.Context, serviceMethod string, opts *YARPCClientOpts) (context.Context, YARPCClientCallHelper) {
	ctx = WithScopeTags(ctx, opts.ScopeTags[serviceMethod])
	return ctx, &callHelper{
		logger:    opts.Loggers[serviceMethod],
		metrics:   opts.Metrics,
		extractor: opts.ContextExtractor,
	}
}

// Start method should be used just before calling the actual YARPC client method call.
// This method starts a timer used for metric.
func (c *callHelper) Start() {
	c.startTime = time.Now()
}

// Finish method should be used right after the actual call to YARPC client method.
// This method emits latency and error metric as well as logging in case of error.
func (c *callHelper) Finish(ctx context.Context, err error) context.Context {
	c.finishTime = time.Now()
	c.metrics.RecordTimer(ctx, clientLatency, c.finishTime.Sub(c.startTime))
	fields := []zapcore.Field{
		zap.Time(logFieldRequestStartTime, c.startTime),
		zap.Time(logFieldRequestFinishedTime, c.finishTime),
	}
	ctx = WithEndpointRequestHeadersField(ctx, map[string]string{})
	if c.extractor != nil {
		fields = append(fields, c.extractor.ExtractLogFields(ctx)...)
	}
	fields = append(fields, GetLogFieldsFromCtx(ctx)...)
	if err != nil {
		if yarpcerrors.IsStatus(err) {
			yarpcErr := yarpcerrors.FromError(err)
			errCode := strings.Builder{}
			errCode.WriteString("client.errors.")
			errCode.WriteString(yarpcErr.Code().String())
			c.metrics.IncCounter(ctx, errCode.String(), 1)

			fields = append(fields, zap.Int("code", int(yarpcErr.Code())))
			fields = append(fields, zap.String("message", yarpcErr.Message()))
			fields = append(fields, zap.String("name", yarpcErr.Name()))
		} else {
			fields = append(fields, zap.Error(err))
		}
		c.metrics.IncCounter(ctx, "client.errors", 1)
		c.logger.Warn("Failed to send outgoing client gRPC request", fields...)
		return ctx
	}
	c.logger.Info("Finished an outgoing client gRPC request", fields...)
	c.metrics.IncCounter(ctx, "client.success", 1)
	return ctx
}
