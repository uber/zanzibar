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
	"context"

	"github.com/pborman/uuid"
	"go.uber.org/zap"
)

type contextFieldKey string

const (
	endpointKey        = contextFieldKey("endpoint")
	requestUUIDKey     = contextFieldKey("requestUUID")
	routingDelegateKey = contextFieldKey("rd")
	requestLogFields   = contextFieldKey("requestLogFields")
)

const (
	logFieldRequestMethod       = "method"
	logFieldRequestURL          = "url"
	logFieldRequestStartTime    = "timestamp-started"
	logFieldRequestFinishedTime = "timestamp-finished"
	logFieldRequestHeaderPrefix = "Request-Header"
	logFieldResponseStatusCode  = "statusCode"
)

// WithEndpointField adds the endpoint information in the
// request context.
func WithEndpointField(ctx context.Context, endpoint string) context.Context {
	return context.WithValue(ctx, endpointKey, endpoint)
}

// GetRequestEndpointFromCtx returns the endpoint, if it exists on context
func GetRequestEndpointFromCtx(ctx context.Context) string {
	if val := ctx.Value(endpointKey); val != nil {
		endpoint, _ := val.(string)
		return endpoint
	}
	return ""
}

// withRequestFields annotates zanzibar request context to context.Context. In
// future, we can use a request context struct to add more context in terms of
// request handler, etc if need be.
func withRequestFields(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestUUIDKey, uuid.NewUUID())
}

// GetRequestUUIDFromCtx returns the RequestUUID, if it exists on context
// TODO: in future, we can extend this to have request object
func GetRequestUUIDFromCtx(ctx context.Context) uuid.UUID {
	if val := ctx.Value(requestUUIDKey); val != nil {
		uuid, _ := val.(uuid.UUID)
		return uuid
	}
	return nil
}

// WithRoutingDelegate adds the tchannel routing delegate information in the
// request context.
func WithRoutingDelegate(ctx context.Context, rd string) context.Context {
	return context.WithValue(ctx, routingDelegateKey, rd)
}

// GetRoutingDelegateFromCtx returns the tchannel routing delegate info
// extracted from context.
func GetRoutingDelegateFromCtx(ctx context.Context) string {
	if val := ctx.Value(routingDelegateKey); val != nil {
		rd, _ := val.(string)
		return rd
	}
	return ""
}

// WithLogFields returns a new context with the given log fields attached to context.Context
func WithLogFields(ctx context.Context, newFields ...zap.Field) context.Context {
	return context.WithValue(ctx, requestLogFields, accumulateLogFields(ctx, newFields))
}

func accumulateLogFields(ctx context.Context, newFields []zap.Field) []zap.Field {
	previousFieldsUntyped := ctx.Value(requestLogFields)
	if previousFieldsUntyped == nil {
		previousFieldsUntyped = []zap.Field{}
	}
	previousFields := previousFieldsUntyped.([]zap.Field)

	fields := make([]zap.Field, 0, len(previousFields)+len(newFields))

	for _, field := range previousFields {
		fields = append(fields, field)
	}
	for _, field := range newFields {
		fields = append(fields, field)
	}

	return fields
}

// ContextLogger is a logger that extracts some log fields from the context before passing through to underlying zap logger.
type ContextLogger interface {
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Panic(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
}

// NewContextLogger returns a logger that extracts log fields a context before passing through to underlying zap logger.
func NewContextLogger(log *zap.Logger) ContextLogger {
	return &contextLogger{
		log: log,
	}
}

type contextLogger struct {
	log *zap.Logger
}

func (c *contextLogger) Debug(ctx context.Context, msg string, userFields ...zap.Field) {
	c.log.Debug(msg, accumulateLogFields(ctx, userFields)...)
}

func (c *contextLogger) Error(ctx context.Context, msg string, userFields ...zap.Field) {
	c.log.Error(msg, accumulateLogFields(ctx, userFields)...)
}

func (c *contextLogger) Info(ctx context.Context, msg string, userFields ...zap.Field) {
	c.log.Info(msg, accumulateLogFields(ctx, userFields)...)
}

func (c *contextLogger) Panic(ctx context.Context, msg string, userFields ...zap.Field) {
	c.log.Panic(msg, accumulateLogFields(ctx, userFields)...)
}

func (c *contextLogger) Warn(ctx context.Context, msg string, userFields ...zap.Field) {
	c.log.Warn(msg, accumulateLogFields(ctx, userFields)...)
}
