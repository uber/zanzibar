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
	"github.com/uber-go/tally"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextFieldKey string

// ContextScopeTagsExtractor defines func where extracts tags from context
type ContextScopeTagsExtractor func(context.Context) map[string]string

const (
	endpointKey           = contextFieldKey("endpoint")
	requestUUIDKey        = contextFieldKey("requestUUID")
	routingDelegateKey    = contextFieldKey("rd")
	endpointRequestHeader = contextFieldKey("endpointRequestHeader")
	requestLogFields      = contextFieldKey("requestLogFields")
	requestScopeFields    = contextFieldKey("requestScopeFields")
)

const (
	logFieldRequestMethod       = "method"
	logFieldRequestURL          = "url"
	logFieldRequestStartTime    = "timestamp-started"
	logFieldRequestFinishedTime = "timestamp-finished"
	logFieldRequestHeaderPrefix = "Request-Header"
	logFieldResponseStatusCode  = "statusCode"
	logFieldRequestUUID         = "requestUUID"
	logFieldEndpointID          = "endpointID"
	logFieldHandlerID           = "handlerID"
)

const (
	scopeFieldRequestMethod = "method"
	scopeFieldEndpoint      = "endpointid"
	scopeFieldHandler       = "handlerid"
)

const (
	// EndpointScope defines the name of endpoint scope
	EndpointScope = "endpoint"

	// ClientScope defines the name of client scope
	ClientScope = "client"
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

// WithEndpointRequestHeadersField adds the endpoint request header information in the
// request context.
func WithEndpointRequestHeadersField(ctx context.Context, requestHeaders map[string]string) context.Context {
	headers := GetEndpointRequestHeadersFromCtx(ctx)
	for k, v := range requestHeaders {
		headers[k] = v
	}

	return context.WithValue(ctx, endpointRequestHeader, headers)
}

// GetEndpointRequestHeadersFromCtx returns the endpoint request headers, if it exists on context
func GetEndpointRequestHeadersFromCtx(ctx context.Context) map[string]string {
	requestHeaders := make(map[string]string)
	if val := ctx.Value(endpointRequestHeader); val != nil {
		headers, _ := val.(map[string]string)
		for k, v := range headers {
			requestHeaders[k] = v
		}
	}

	return requestHeaders
}

// withRequestFields annotates zanzibar request context to context.Context. In
// future, we can use a request context struct to add more context in terms of
// request handler, etc if need be.
func withRequestFields(ctx context.Context) context.Context {
	reqUUID := uuid.NewUUID()
	ctx = context.WithValue(ctx, requestUUIDKey, reqUUID)
	ctx = WithLogFields(ctx, zap.String(logFieldRequestUUID, reqUUID.String()))
	return ctx
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

// WithScopeFields returns a new context with the given scope fields attached to context.Context
func WithScopeFields(ctx context.Context, newFields map[string]string) (context.Context, map[string]string) {
	fields := GetScopeFieldsFromCtx(ctx)
	for k, v := range newFields {
		fields[k] = v
	}

	return context.WithValue(ctx, requestScopeFields, fields), fields
}

// GetScopeFieldsFromCtx returns the tag info extracted from context.
func GetScopeFieldsFromCtx(ctx context.Context) map[string]string {
	fields := make(map[string]string)
	if val := ctx.Value(requestScopeFields); val != nil {
		headers, _ := val.(map[string]string)
		for k, v := range headers {
			fields[k] = v
		}
	}

	return fields
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

// ContextExtractor is a extractor that extracts some log fields from the context
type ContextExtractor interface {
	ExtractScopeTags(ctx context.Context) map[string]string
}

// AddContextScopeTagsExtractor added a scope tags extractor to contextExtractor.
func (c *ContextExtractors) AddContextScopeTagsExtractor(extractors ...ContextScopeTagsExtractor) {
	c.contextScopeExtractors = extractors
}

// MakeContextExtractor returns a extractor that extracts log fields a context.
func (c *ContextExtractors) MakeContextExtractor() ContextExtractor {
	return &ContextExtractors{
		contextScopeExtractors: c.contextScopeExtractors,
	}
}

// ContextExtractors warps extractors for context
type ContextExtractors struct {
	contextScopeExtractors []ContextScopeTagsExtractor
}

// ExtractScopeTags extracts scope fields from a context into a tag.
func (c *ContextExtractors) ExtractScopeTags(ctx context.Context) map[string]string {
	tags := make(map[string]string)
	for _, fn := range c.contextScopeExtractors {
		sc := fn(ctx)
		for k, v := range sc {
			tags[k] = v
		}
	}

	return tags
}

// ContextLogger is a logger that extracts some log fields from the context before passing through to underlying zap logger.
type ContextLogger interface {
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Panic(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)

	// Other utility methods on the logger
	Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry
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

func (c *contextLogger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	return c.log.Check(lvl, msg)
}

// ContextMetrics emits metrics
type ContextMetrics struct {
	scope                  tally.Scope
	InboundHTTPMetrics     *InboundHTTPMetrics
	EndpointMetrics        *EndpointMetrics
	InboundTChannelMetrics *InboundTChannelMetrics
}

// NewContextMetrics emits metrics.
func NewContextMetrics(scope tally.Scope) *ContextMetrics {
	return &ContextMetrics{
		scope: scope,
	}
}

// MakeInboundHTTPMetrics add tags to scope and create inbound http metrics
func (c *ContextMetrics) MakeInboundHTTPMetrics(tags map[string]string) {
	scope := c.scope.Tagged(tags)
	c.InboundHTTPMetrics = NewInboundHTTPMetrics(scope)
}

// MakeEndpointMetrics add tags to scope and create endpoint metrics
func (c *ContextMetrics) MakeEndpointMetrics(tags map[string]string) {
	scope := c.scope.Tagged(tags)
	c.EndpointMetrics = NewEndpointMetrics(scope)
}

// MakeInboundTChannelMetrics add tags to scope and create inbound tchannel metrics
func (c *ContextMetrics) MakeInboundTChannelMetrics(tags map[string]string) {
	scope := c.scope.Tagged(tags)
	c.InboundTChannelMetrics = NewInboundTChannelMetrics(scope)
}
