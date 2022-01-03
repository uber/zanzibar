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
	"strings"
	"time"

	"go.uber.org/yarpc/yarpcerrors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GRPCClientOpts used to configure various client options.
type GRPCClientOpts struct {
	ContextLogger          ContextLogger
	Metrics                ContextMetrics
	ContextExtractor       ContextExtractor
	RoutingKey             string
	RequestUUIDHeaderKey   string
	CircuitBreakerDisabled bool
	Timeout                time.Duration
	ScopeTags              map[string]map[string]string
}

// NewGRPCClientOpts creates a new instance of GRPCClientOpts.
func NewGRPCClientOpts(
	contextLogger ContextLogger,
	metrics ContextMetrics,
	contextExtractor ContextExtractor,
	methodNames map[string]string,
	clientID, routingKey, requestUUIDHeaderKey string,
	circuitBreakerDisabled bool,
	timeoutInMS int,
) *GRPCClientOpts {
	scopeTags := make(map[string]map[string]string)
	for serviceMethod, methodName := range methodNames {
		scopeTags[serviceMethod] = map[string]string{
			scopeTagClient:          clientID,
			scopeTagClientMethod:    methodName,
			scopeTagsTargetEndpoint: serviceMethod,
		}
	}
	return &GRPCClientOpts{
		ContextLogger:          contextLogger,
		Metrics:                metrics,
		ContextExtractor:       contextExtractor,
		RoutingKey:             routingKey,
		RequestUUIDHeaderKey:   requestUUIDHeaderKey,
		CircuitBreakerDisabled: circuitBreakerDisabled,
		Timeout:                time.Duration(timeoutInMS) * time.Millisecond,
		ScopeTags:              scopeTags,
	}
}

// GRPCClientCallHelper is used to track internal state of logging and metrics.
type GRPCClientCallHelper interface {
	// Start method should be used just before calling the actual gRPC client method call.
	Start()
	// Finish method should be used right after the actual call to gRPC client method.
	Finish(ctx context.Context, err error) context.Context
}

type callHelper struct {
	startTime     time.Time
	finishTime    time.Time
	contextLogger ContextLogger
	metrics       ContextMetrics
	extractor     ContextExtractor
}

// NewGRPCClientCallHelper used to initialize a helper that will
// be used to track logging and metric for a gRPC Client call.
func NewGRPCClientCallHelper(ctx context.Context, serviceMethod string, opts *GRPCClientOpts) (context.Context, GRPCClientCallHelper) {
	ctx = WithScopeTags(ctx, opts.ScopeTags[serviceMethod])
	return ctx, &callHelper{
		contextLogger: opts.ContextLogger,
		metrics:       opts.Metrics,
		extractor:     opts.ContextExtractor,
	}
}

// Start method should be used just before calling the actual gRPC client method call.
// This method starts a timer used for metric.
func (c *callHelper) Start() {
	c.startTime = time.Now()
}

// Finish method should be used right after the actual call to gRPC client method.
// This method emits latency and error metric as well as logging in case of error.
func (c *callHelper) Finish(ctx context.Context, err error) context.Context {
	c.finishTime = time.Now()
	delta := c.finishTime.Sub(c.startTime)
	c.metrics.RecordTimer(ctx, clientLatency, delta)
	c.metrics.RecordHistogramDuration(ctx, clientLatency, delta)
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
		c.contextLogger.WarnZ(ctx, "Failed to send outgoing client gRPC request", fields...)
		return ctx
	}
	c.contextLogger.DebugZ(ctx, "Finished an outgoing client gRPC request", fields...)
	c.metrics.IncCounter(ctx, "client.success", 1)
	return ctx
}
