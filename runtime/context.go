// Copyright (c) 2023 Uber Technologies, Inc.
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
	"sync"
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
	endpointKey            = contextFieldKey("endpoint")
	requestUUIDKey         = contextFieldKey("requestUUID")
	routingDelegateKey     = contextFieldKey("rd")
	shardKey               = contextFieldKey("sk")
	endpointRequestHeader  = contextFieldKey("endpointRequestHeader")
	scopeTags              = contextFieldKey("scopeTags")
	ctxLogCounterName      = contextFieldKey("ctxLogCounter")
	ctxLogLevel            = contextFieldKey("ctxLogLevel")
	ctxTimeoutRetryOptions = contextFieldKey("trOptions")
	safeFieldsKey          = contextFieldKey("safeFields")
)

const (
	// thrift service::method of endpoint thrift spec
	logFieldRequestMethod      = "endpointThriftMethod"
	logFieldRequestHTTPMethod  = "method"
	logFieldRequestPathname    = "pathname"
	logFieldRequestRemoteAddr  = "remoteAddr"
	logFieldRequestHost        = "host"
	logFieldResponseStatusCode = "statusCode"
	logFieldRequestUUID        = "requestUUID"
	logFieldEndpointID         = "endpointID"
	logFieldEndpointHandler    = "endpointHandler"
	logFieldClientStatusCode   = "client_status_code"
	logFieldClientRemoteAddr   = "client_remote_addr"

	logFieldClientRequestHeaderPrefix    = "Client-Req-Header"
	logFieldClientResponseHeaderPrefix   = "Client-Res-Header"
	logFieldEndpointResponseHeaderPrefix = "Res-Header"

	// LogFieldErrorLocation is field name to log error location.
	LogFieldErrorLocation = "error_location"

	// LogFieldErrorType is field name to log error type.
	LogFieldErrorType = "error_type"
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

// scopeData - scope and tags
type scopeData struct {
	tags  map[string]string
	scope tally.Scope // optional - may not be used by zanzibar users who use
}

// WithTimeAndRetryOptions returns a context with timeout and retry options.
func WithTimeAndRetryOptions(ctx context.Context, tro *TimeoutAndRetryOptions) context.Context {
	return context.WithValue(ctx, ctxTimeoutRetryOptions, tro)
}

// GetTimeoutAndRetryOptions returns timeout and retry options stored in the context
func GetTimeoutAndRetryOptions(ctx context.Context) *TimeoutAndRetryOptions {
	if val := ctx.Value(ctxTimeoutRetryOptions); val != nil {
		tro, _ := val.(*TimeoutAndRetryOptions)
		return tro
	}
	return nil
}

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
	requestHeaders := make(map[string]string, 0)
	if val := ctx.Value(endpointRequestHeader); val != nil {
		headers, _ := val.(map[string]string)
		requestHeaders = make(map[string]string, len(headers))
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

type safeFields struct {
	mu     sync.Mutex
	fields []zap.Field
}

// WithSafeLogFields initiates empty safeFields in the context.
func WithSafeLogFields(ctx context.Context) context.Context {
	return context.WithValue(ctx, safeFieldsKey, &safeFields{})
}

func (sf *safeFields) append(fields []zap.Field) {
	sf.mu.Lock()
	sf.fields = append(sf.fields, fields...)
	sf.mu.Unlock()
}

func (sf *safeFields) getFields() []zap.Field {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.fields
}

func getSafeFieldsFromContext(ctx context.Context) *safeFields {
	v := ctx.Value(safeFieldsKey).(*safeFields)
	if v != nil {
		return v
	}
	return &safeFields{}
}

// WithLogFields returns a new context with the given log fields attached to context.Context
//
// Deprecated: Use ContextLogger.Append instead.
func WithLogFields(ctx context.Context, newFields ...zap.Field) context.Context {
	sf := getSafeFieldsFromContext(ctx)
	sf.append(newFields)
	return context.WithValue(ctx, safeFieldsKey, sf)
}

// GetLogFieldsFromCtx returns the log fields attached to the context.Context
func GetLogFieldsFromCtx(ctx context.Context) []zap.Field {
	if ctx != nil {
		v := ctx.Value(safeFieldsKey)
		if v != nil {
			return v.(*safeFields).getFields()
		}
	}
	return []zap.Field{}
}

// WithScopeTags adds tags to context without updating the scope
func WithScopeTags(ctx context.Context, fields map[string]string) context.Context {
	return WithScopeTagsDefault(ctx, fields, nil)
}

// WithScopeTagsDefault returns a new context with the given scope tags attached to context.Context
// This operation adds the fields and updates scope in context.
// defaultScope when scopeData is stored for the first time
func WithScopeTagsDefault(ctx context.Context, fields map[string]string, defaultScope tally.Scope) context.Context {
	sd, ok := ctx.Value(scopeTags).(*scopeData)
	if !ok {
		m := make(map[string]string)
		sd = &scopeData{tags: merge(m, fields)}
	}

	// assign a scope if one hasn't been sent
	if sd.scope == nil {
		sd.scope = defaultScope
	}

	// modify the scope if non nil
	if sd.scope != nil {
		sd.scope = sd.scope.Tagged(fields)
	}

	return context.WithValue(ctx, scopeTags, sd)
}

func merge(m1, m2 map[string]string) map[string]string {
	out := make(map[string]string, len(m1)+len(m2))
	for k, v := range m1 {
		out[k] = v
	}
	for k, v := range m2 {
		out[k] = v
	}
	return out
}

// GetScopeTagsFromCtx returns the tag info extracted from context.
func GetScopeTagsFromCtx(ctx context.Context) map[string]string {
	tags := make(map[string]string, 0)
	if val := ctx.Value(scopeTags); val != nil {
		sd, _ := val.(*scopeData)

		// copy
		tags = make(map[string]string, len(sd.tags))
		for k, v := range sd.tags {
			tags[k] = v
		}
	}

	return tags
}

// getScope returns the scope stored in the context or returns the default scope when not found
func getScope(ctx context.Context, defScope tally.Scope) tally.Scope {
	sd, ok := ctx.Value(scopeTags).(*scopeData)
	if !ok {
		return defScope
	}

	if sd.scope == nil { //backward compatibility - scope was not calculated
		return defScope.Tagged(sd.tags)
	}
	return sd.scope
}

func accumulateLogFields(ctx context.Context, newFields []zap.Field) []zap.Field {
	previousFields := GetLogFieldsFromCtx(ctx)
	previousFieldsLen := len(previousFields)
	accumulatedFields := make([]zap.Field, previousFieldsLen, previousFieldsLen+len(newFields))
	copy(accumulatedFields, previousFields)
	return append(accumulatedFields, newFields...)
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

	// Other utility methods on the logger
	Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry

	SetSkipZanzibarLogs(bool)

	// Append appends the fields to the context.
	Append(ctx context.Context, fields ...zap.Field)
}

// NewContextLogger returns a logger that extracts log fields a context before passing through to underlying zap logger.
func NewContextLogger(log *zap.Logger) ContextLogger {
	return &contextLogger{
		log:              log,
		skipZanzibarLogs: false,
	}
}

func (c *contextLogger) SetSkipZanzibarLogs(skipZanzibarLogs bool) {
	c.skipZanzibarLogs = skipZanzibarLogs
}

type contextLogger struct {
	log              *zap.Logger
	skipZanzibarLogs bool
}

func (c *contextLogger) Append(ctx context.Context, fields ...zap.Field) {
	v := getSafeFieldsFromContext(ctx)
	v.append(fields)
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
		ctx = GetAccumulatedLogContext(ctx, c, msg, zapcore.DebugLevel, userFields...)
	} else {
		c.log.Debug(msg, accumulateLogFields(ctx, userFields)...)
	}
	return ctx
}

func (c *contextLogger) ErrorZ(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	if c.skipZanzibarLogs {
		ctx = GetAccumulatedLogContext(ctx, c, msg, zapcore.ErrorLevel, userFields...)
	} else {
		c.log.Error(msg, accumulateLogFields(ctx, userFields)...)
	}
	return ctx
}

func (c *contextLogger) InfoZ(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	if c.skipZanzibarLogs {
		ctx = GetAccumulatedLogContext(ctx, c, msg, zapcore.InfoLevel, userFields...)
	} else {
		c.log.Info(msg, accumulateLogFields(ctx, userFields)...)
	}
	return ctx
}

func (c *contextLogger) PanicZ(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	if c.skipZanzibarLogs {
		ctx = GetAccumulatedLogContext(ctx, c, msg, zapcore.PanicLevel, userFields...)
	} else {
		c.log.Panic(msg, accumulateLogFields(ctx, userFields)...)
	}
	return ctx
}

func (c *contextLogger) WarnZ(ctx context.Context, msg string, userFields ...zap.Field) context.Context {
	if c.skipZanzibarLogs {
		ctx = GetAccumulatedLogContext(ctx, c, msg, zapcore.WarnLevel, userFields...)
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
// This interface is now deprecated.
type ContextMetrics interface {
	Scope() tally.Scope
	IncCounter(ctx context.Context, name string, value int64)
	RecordTimer(ctx context.Context, name string, d time.Duration)
	RecordHistogramDuration(ctx context.Context, name string, d time.Duration)
}

// contextMetrics is a container that holds the initial value of tally.Scope and to be used when a context is
// yet to be created.
//
// Note: current code passes contextMetrics to various pieces of code even though ctx is available. This wiring
// should be deprecated in the future.
type contextMetrics struct {
	scope tally.Scope
}

// NewContextMetrics create ContextMetrics to emit metrics with tags extracted from context.
func NewContextMetrics(scope tally.Scope) ContextMetrics {
	return &contextMetrics{
		scope: scope,
	}
}

// Scope retrieves the scope stored within context metrics
func (c *contextMetrics) Scope() tally.Scope {
	return c.scope
}

// IncCounter increments the counter with current tags from context
func (c *contextMetrics) IncCounter(ctx context.Context, name string, value int64) {
	getScope(ctx, c.scope).Counter(name).Inc(value)
}

// RecordTimer records the duration with current tags from context
func (c *contextMetrics) RecordTimer(ctx context.Context, name string, d time.Duration) {
	getScope(ctx, c.scope).Timer(name).Record(d)
}

// RecordHistogramDuration records the duration with current tags from context in a histogram
func (c *contextMetrics) RecordHistogramDuration(ctx context.Context, name string, d time.Duration) {
	getScope(ctx, c.scope).Histogram(name, tally.DefaultBuckets).RecordDuration(d)
}

// GetAccumulatedLogContext returns accumulated log context
func GetAccumulatedLogContext(ctx context.Context, c *contextLogger, msg string, logLevel zapcore.Level, userFields ...zap.Field) context.Context {
	if !c.log.Core().Enabled(logLevel) {
		return ctx
	}
	return accumulateLogMsgAndFieldsInContext(ctx, msg, userFields, logLevel)
}
