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
	"strconv"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextFieldKey string

// ContextScopeTagsExtractor defines func where extracts tags from context
type ContextScopeTagsExtractor func(context.Context) map[string]string

// ContextLogFieldsExtractor defines func where extracts log fields from context
type ContextLogFieldsExtractor func(context.Context) []zap.Field

const (
	endpointKey           = contextFieldKey("endpoint")
	requestUUIDKey        = contextFieldKey("requestUUID")
	routingDelegateKey    = contextFieldKey("rd")
	shardKey              = contextFieldKey("sk")
	endpointRequestHeader = contextFieldKey("endpointRequestHeader")
	requestLogFields      = contextFieldKey("requestLogFields")
	scopeTags             = contextFieldKey("scopeTags")
	ctxLogCounterName     = contextFieldKey("ctxLogCounter")
	ctxLogLevel           = contextFieldKey("ctxLogLevel")
)

const (
	// thrift service::method of endpoint thrift spec
	logFieldRequestMethod       = "endpointThriftMethod"
	logFieldRequestURL          = "url"
	logFieldRequestStartTime    = "timestamp-started"
	logFieldRequestFinishedTime = "timestamp-finished"
	logFieldResponseStatusCode  = "statusCode"
	logFieldRequestUUID         = "requestUUID"
	logFieldEndpointID          = "endpointID"
	logFieldEndpointHandler     = "endpointHandler"
	logFieldClientHTTPMethod    = "clientHTTPMethod"

	logFieldClientRequestHeaderPrefix    = "Client-Req-Header"
	logFieldClientResponseHeaderPrefix   = "Client-Res-Header"
	logFieldEndpointResponseHeaderPrefix = "Res-Header"
)

const (
	scopeTagClientMethod    = "clientmethod"
	scopeTagEndpointMethod  = "endpointmethod"
	scopeTagClient          = "clientid"
	scopeTagClientType      = "clienttype"
	scopeTagEndpoint        = "endpointid"
	scopeTagHandler         = "handlerid"
	scopeTagError           = "error"
	scopeTagMiddleWare      = "middlewarename"
	scopeTagStatus          = "status"
	scopeTagProtocol        = "protocol"
	scopeTagHTTP            = "HTTP"
	scopeTagTChannel        = "TChannel"
	scopeTagsTargetService  = "targetservice"
	scopeTagsTargetEndpoint = "targetendpoint"
	scopeTagsAPIEnvironment = "apienvironment"

	apiEnvironmentDefault = "production"
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

// withRequestUUID returns a context with request uuid.
func withRequestUUID(ctx context.Context, reqUUID string) context.Context {
	return context.WithValue(ctx, requestUUIDKey, reqUUID)
}

// RequestUUIDFromCtx returns the RequestUUID, if it exists on context
func RequestUUIDFromCtx(ctx context.Context) string {
	if val := ctx.Value(requestUUIDKey); val != nil {
		uuid, _ := val.(string)
		return uuid
	}
	return ""
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

// WithShardKey adds the tchannel shard key information in the
// request context.
func WithShardKey(ctx context.Context, sk string) context.Context {
	return context.WithValue(ctx, shardKey, sk)
}

// GetShardKeyFromCtx returns the tchannel shardkey info
// extracted from context.
func GetShardKeyFromCtx(ctx context.Context) string {
	if val := ctx.Value(shardKey); val != nil {
		sk, _ := val.(string)
		return sk
	}
	return ""
}

// WithLogFields returns a new context with the given log fields attached to context.Context
func WithLogFields(ctx context.Context, newFields ...zap.Field) context.Context {
	return context.WithValue(ctx, requestLogFields, accumulateLogFields(ctx, newFields))
}

// GetLogFieldsFromCtx returns the log fields attached to the context.Context
func GetLogFieldsFromCtx(ctx context.Context) []zap.Field {
	var fields []zap.Field
	if ctx != nil {
		v := ctx.Value(requestLogFields)
		if v != nil {
			fields = v.([]zap.Field)
		}
	}
	return fields
}

// WithScopeTags returns a new context with the given scope tags attached to context.Context
func WithScopeTags(ctx context.Context, newFields map[string]string) context.Context {
	tags := GetScopeTagsFromCtx(ctx)
	for k, v := range newFields {
		tags[k] = v
	}

	return context.WithValue(ctx, scopeTags, tags)
}

// GetScopeTagsFromCtx returns the tag info extracted from context.
func GetScopeTagsFromCtx(ctx context.Context) map[string]string {
	tags := make(map[string]string)
	if val := ctx.Value(scopeTags); val != nil {
		headers, _ := val.(map[string]string)
		for k, v := range headers {
			tags[k] = v
		}
	}

	return tags
}

func accumulateLogFields(ctx context.Context, newFields []zap.Field) []zap.Field {
	previousFields := GetLogFieldsFromCtx(ctx)
	return append(previousFields, newFields...)
}

func accumulateLogMsgAndFieldsInContext(ctx context.Context, msg string, newFields []zap.Field, logLevel zapcore.Level) context.Context {
	ctxLogCounter := 1
	v := ctx.Value(ctxLogCounterName)
	if v != nil {
		ctxLogCounter = v.(int)
		ctxLogCounter++
	}
	ctxLogLevelInterface := ctx.Value(ctxLogLevel)
	if ctxLogLevelInterface != nil {
		logLevelValue := ctxLogLevelInterface.(zapcore.Level)
		if logLevel < logLevelValue {
			logLevel = logLevelValue
		}
	}
	ctx = WithLogFields(ctx, zap.String("msg"+strconv.Itoa(ctxLogCounter), msg))
	ctx = WithLogFields(ctx, newFields...)
	ctx = context.WithValue(ctx, ctxLogCounterName, ctxLogCounter)
	ctx = context.WithValue(ctx, ctxLogLevel, logLevel)
	return ctx
}

// GetCtxLogCounterFromCtx returns ctxLogCounter value from ctx
func GetCtxLogCounterFromCtx(ctx context.Context) int {
	ctxLogCounter := 0
	v := ctx.Value(ctxLogCounterName)
	if v != nil {
		ctxLogCounter = v.(int)
	}
	return ctxLogCounter
}

// GetCtxLogLevelOrDebugLevelFromCtx returns ctxLogLevel value from ctx
func GetCtxLogLevelOrDebugLevelFromCtx(ctx context.Context) zapcore.Level {
	ctxLogLevelInterface := ctx.Value(ctxLogLevel)
	if ctxLogLevelInterface != nil {
		return ctxLogLevelInterface.(zapcore.Level)
	}
	return zapcore.DebugLevel
}

// ContextExtractor is a extractor that extracts some log fields from the context
type ContextExtractor interface {
	ExtractScopeTags(ctx context.Context) map[string]string
	ExtractLogFields(ctx context.Context) []zap.Field
}

// ContextExtractors warps extractors for context, implements ContextExtractor interface
type ContextExtractors struct {
	ScopeTagsExtractors []ContextScopeTagsExtractor
	LogFieldsExtractors []ContextLogFieldsExtractor
}

// ExtractScopeTags extracts scope fields from a context into a tag.
func (c *ContextExtractors) ExtractScopeTags(ctx context.Context) map[string]string {
	tags := make(map[string]string)
	for _, fn := range c.ScopeTagsExtractors {
		sc := fn(ctx)
		for k, v := range sc {
			tags[k] = v
		}
	}

	return tags
}

// ExtractLogFields extracts log fields from a context into a field.
func (c *ContextExtractors) ExtractLogFields(ctx context.Context) []zap.Field {
	var fields []zap.Field
	for _, fn := range c.LogFieldsExtractors {
		logFields := fn(ctx)
		fields = append(fields, logFields...)
	}

	return fields
}

// ContextLogger is a logger that extracts some log fields from the context before passing through to underlying zap logger.
// In cases it also updates the context instead of logging
type ContextLogger interface {
	Debug(ctx context.Context, msg string, fields ...zap.Field) context.Context
	Error(ctx context.Context, msg string, fields ...zap.Field) context.Context
	Info(ctx context.Context, msg string, fields ...zap.Field) context.Context
	Panic(ctx context.Context, msg string, fields ...zap.Field) context.Context
	Warn(ctx context.Context, msg string, fields ...zap.Field) context.Context

	// DebugZ skips logs, and adds to context if skipZanzibarLogs is set to true otherwise behaves as normal Debug, similarly for other XxxxZ's
	DebugZ(ctx context.Context, msg string, fields ...zap.Field) context.Context
	ErrorZ(ctx context.Context, msg string, fields ...zap.Field) context.Context
	InfoZ(ctx context.Context, msg string, fields ...zap.Field) context.Context
	PanicZ(ctx context.Context, msg string, fields ...zap.Field) context.Context
	WarnZ(ctx context.Context, msg string, fields ...zap.Field) context.Context

	//GetLogger returns the raw logger
	GetLogger() Logger

	// Other utility methods on the logger
	Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry

	SetSkipZanzibarLogs(bool)
}

// NewContextLogger returns a logger that extracts log fields a context before passing through to underlying zap logger.
func NewContextLogger(log *zap.Logger) ContextLogger {
	return &contextLogger{
		log:              log,
		skipZanzibarLogs: false,
	}
}

//GetLogger returns the logger
func (c *contextLogger) GetLogger() Logger {
	return c.log
}

func (c *contextLogger) SetSkipZanzibarLogs(skipZanzibarLogs bool) {
	c.skipZanzibarLogs = skipZanzibarLogs
}

type contextLogger struct {
	log              *zap.Logger
	skipZanzibarLogs bool
}

func (c *contextLogger) Debug(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	c.log.Debug(msg, accumulateLogFields(ctx, userFields)...)
	return ctx
}

func (c *contextLogger) Error(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	c.log.Error(msg, accumulateLogFields(ctx, userFields)...)
	return ctx
}

func (c *contextLogger) Info(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	c.log.Info(msg, accumulateLogFields(ctx, userFields)...)
	return ctx
}

func (c *contextLogger) Panic(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	c.log.Panic(msg, accumulateLogFields(ctx, userFields)...)
	return ctx
}

func (c *contextLogger) Warn(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	c.log.Warn(msg, accumulateLogFields(ctx, userFields)...)
	return ctx
}

func (c *contextLogger) DebugZ(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	if c.skipZanzibarLogs {
		ctx = accumulateLogMsgAndFieldsInContext(ctx, msg, userFields, zapcore.DebugLevel)
	} else {
		c.log.Debug(msg, accumulateLogFields(ctx, userFields)...)
	}
	return ctx
}

func (c *contextLogger) ErrorZ(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	if c.skipZanzibarLogs {
		ctx = accumulateLogMsgAndFieldsInContext(ctx, msg, userFields, zapcore.ErrorLevel)
	} else {
		c.log.Error(msg, accumulateLogFields(ctx, userFields)...)
	}
	return ctx
}

func (c *contextLogger) InfoZ(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	if c.skipZanzibarLogs {
		ctx = accumulateLogMsgAndFieldsInContext(ctx, msg, userFields, zapcore.InfoLevel)
	} else {
		c.log.Info(msg, accumulateLogFields(ctx, userFields)...)
	}
	return ctx
}

func (c *contextLogger) PanicZ(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	if c.skipZanzibarLogs {
		ctx = accumulateLogMsgAndFieldsInContext(ctx, msg, userFields, zapcore.PanicLevel)
	} else {
		c.log.Panic(msg, accumulateLogFields(ctx, userFields)...)
	}
	return ctx
}

func (c *contextLogger) WarnZ(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	if c.skipZanzibarLogs {
		ctx = accumulateLogMsgAndFieldsInContext(ctx, msg, userFields, zapcore.WarnLevel)
	} else {
		c.log.Warn(msg, accumulateLogFields(ctx, userFields)...)
	}
	return ctx
}

func (c *contextLogger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	return c.log.Check(lvl, msg)
}

// Logger is a generic logger interface that zap.Logger implements.
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry
}

// ContextMetrics emit metrics with tags extracted from context.
type ContextMetrics interface {
	IncCounter(ctx context.Context, name string, value int64)
	RecordTimer(ctx context.Context, name string, d time.Duration)
	RecordHistogramDuration(ctx context.Context, name string, d time.Duration)
}

type contextMetrics struct {
	scope tally.Scope
}

// NewContextMetrics create ContextMetrics to emit metrics with tags extracted from context.
func NewContextMetrics(scope tally.Scope) ContextMetrics {
	return &contextMetrics{
		scope: scope,
	}
}

// IncCounter increments the counter with current tags from context
func (c *contextMetrics) IncCounter(ctx context.Context, name string, value int64) {
	c.scope.Tagged(GetScopeTagsFromCtx(ctx)).Counter(name).Inc(value)
}

// RecordTimer records the duration with current tags from context
func (c *contextMetrics) RecordTimer(ctx context.Context, name string, d time.Duration) {
	c.scope.Tagged(GetScopeTagsFromCtx(ctx)).Timer(name).Record(d)
}

// RecordHistogramDuration records the duration with current tags from context in a histogram
func (c *contextMetrics) RecordHistogramDuration(ctx context.Context, name string, d time.Duration) {
	c.scope.Tagged(GetScopeTagsFromCtx(ctx)).Histogram(name, tally.DefaultBuckets).RecordDuration(d)
}
