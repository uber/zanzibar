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
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestWithEndpointField(t *testing.T) {
	expected := "someEndpoint"
	ctx := WithEndpointField(context.TODO(), expected)

	ek := ctx.Value(endpointKey)
	endpoint, ok := ek.(string)

	assert.True(t, ok)
	assert.Equal(t, endpoint, expected)
}

func TestGetRequestEndpointFromCtx(t *testing.T) {
	expected := "someEndpoint"
	ctx := WithEndpointField(context.TODO(), expected)
	endpoint := GetRequestEndpointFromCtx(ctx)
	assert.Equal(t, expected, endpoint)

	expected = ""
	ctx = context.TODO()
	endpoint = GetRequestEndpointFromCtx(ctx)
	assert.Equal(t, expected, endpoint)
}

func TestWithEndpointRequestHeadersField(t *testing.T) {
	expected := map[string]string{"region": "san_francisco", "dc": "sjc1"}
	ctx := WithEndpointRequestHeadersField(context.TODO(), expected)
	rh := ctx.Value(endpointRequestHeader)
	requestHeaders, ok := rh.(map[string]string)

	assert.True(t, ok)
	assert.Equal(t, requestHeaders, expected)
}

func TestGetEndpointRequestHeadersFromCtx(t *testing.T) {
	expected := map[string]string{"region": "san_francisco", "dc": "sjc1"}
	headers := map[string]string{"region": "san_francisco", "dc": "sjc1"}
	ctx := WithEndpointRequestHeadersField(context.TODO(), headers)
	requestHeaders := GetEndpointRequestHeadersFromCtx(ctx)
	assert.Equal(t, expected, requestHeaders)

	expected = map[string]string{}
	ctx = context.TODO()
	requestHeaders = GetEndpointRequestHeadersFromCtx(ctx)
	assert.Equal(t, expected, requestHeaders)
}

func TestWithScopeTags(t *testing.T) {
	expected := map[string]string{"endpoint": "tincup", "handler": "exchange"}
	ctx := WithScopeTagsDefault(context.TODO(), expected, tally.NoopScope)
	rs := ctx.Value(scopeTags)
	sd, ok := rs.(*scopeData)

	assert.True(t, ok)
	assert.Equal(t, expected, sd.tags)
}

func TestGetScopeTagsFromCtx(t *testing.T) {
	expected := map[string]string{"endpoint": "tincup", "handler": "exchange"}
	scopeTags := map[string]string{"endpoint": "tincup", "handler": "exchange"}
	ctx := WithScopeTagsDefault(context.TODO(), scopeTags, tally.NoopScope)
	scopes := GetScopeTagsFromCtx(ctx)
	assert.Equal(t, expected, scopes)

	expected = map[string]string{}
	ctx = context.TODO()
	scopes = GetScopeTagsFromCtx(ctx)
	assert.Equal(t, expected, scopes)
}

// TestGetScopeFromCtx tests the common case where a default scope is available
func TestGetScopeFromCtx(t *testing.T) {

	const counterName = "cn"

	// create root scope
	root := tally.NewTestScope("", nil)

	// add scope tags to context
	tags := map[string]string{"endpoint": "tincup", "handler": "exchange"}
	ctx := WithScopeTagsDefault(context.TODO(), tags, root)
	scope := getScope(ctx, root)

	// increment a counter
	scope.Counter(counterName).Inc(1)

	// check if the counter has the right tags
	ts := scope.(tally.TestScope)
	ss := ts.Snapshot()

	assert.Equal(t, 1, len(ss.Counters()))
	// examine only the first counter
	for _, counterSnapshot := range ss.Counters() {
		assert.Equal(t, tags, counterSnapshot.Tags())
		break
	}
}

// TestContextMetrics_NoDefault_getScope tests the scenario where
//  1. a default scope is not available when tags are assigned to a context
//  2. getScope chooses the default context
func TestGetScopeFromCtx_NoDefault_getScope(t *testing.T) {

	const counterName = "cn"

	root := tally.NoopScope
	tags := map[string]string{"endpoint": "tincup", "handler": "exchange"}

	ctx := WithScopeTags(context.TODO(), tags)
	scope := getScope(ctx, root)
	scope.Counter(counterName).Inc(1)

	ts := scope.(tally.TestScope)
	ss := ts.Snapshot()

	assert.Equal(t, 1, len(ss.Counters()))
	// examine only the first counter
	for _, counterSnapshot := range ss.Counters() {
		assert.Equal(t, tags, counterSnapshot.Tags())
		break
	}
}

// TestContextMetrics_NoDefault tests the scenario where
//  1. a default scope is not available
//  2. methods in contextMetrics pick the scope stored within it
func TestContextMetrics_NoDefault(t *testing.T) {

	const counterName = "cn"

	// create root scope
	root := tally.NewTestScope("", nil)

	// create context metrics
	cm := NewContextMetrics(root)
	tags := map[string]string{"endpoint": "tincup", "handler": "exchange"}

	// add tags to context withot
	ctx := WithScopeTags(context.TODO(), tags)

	// increment a counter
	cm.IncCounter(ctx, counterName, 1)

	// check if the counter is generated with the right tags
	scope := getScope(ctx, root)
	ts := scope.(tally.TestScope)
	ss := ts.Snapshot()
	assert.Equal(t, 1, len(ss.Counters()))
	// examine only the first counter
	for _, counterSnapshot := range ss.Counters() {
		assert.Equal(t, tags, counterSnapshot.Tags())
		break
	}
}

// TestContextMetrics_NoDefault_TagDivergence tests the scenario where
//  1. a default scope is not available
//  2. tags stored in the root scope and the context diverge
func TestContextMetrics_NoDefault_TagDivergence(t *testing.T) {

	const counterName = "cn"

	// step: create a root scope with tags
	rootTags := map[string]string{"tag": "t3"}
	root := tally.NewTestScope("", rootTags)

	// step: create context metrics
	cm := NewContextMetrics(root)

	// step: assign new tags to context without a default scope
	tags := map[string]string{"endpoint": "tincup", "handler": "exchange"}
	ctx := WithScopeTags(context.TODO(), tags)

	// step: increment counter
	cm.IncCounter(ctx, counterName, 1)

	// step: check if the counter is generated with the tags from root and context
	expected := merge(rootTags, tags)
	scope := getScope(ctx, root)
	ts := scope.(tally.TestScope)
	ss := ts.Snapshot()
	assert.Equal(t, 1, len(ss.Counters()))

	// step: examine only the first counter
	for _, counterSnapshot := range ss.Counters() {
		assert.Equal(t, expected, counterSnapshot.Tags())
		break
	}
}

// TestContextMetrics_ScopeDivergence checks for the case where
// 1. the root scope and the scope within the context has diverged
func TestContextMetrics_ScopeDivergence(t *testing.T) {

	const counterName = "cn"

	// step: create a root scope with tags and assign it to metrics
	rootTags := map[string]string{"tag": "t3"}
	root := tally.NewTestScope("", rootTags)
	cm := NewContextMetrics(root)

	// step: create another root scope with tags and assign it to context
	ctxTags := map[string]string{"endpoint": "tincup", "handler": "exchange"}
	ctxRoot := tally.NewTestScope("", ctxTags)
	ctx := WithScopeTagsDefault(context.TODO(), ctxTags, ctxRoot)

	// step: increment counter
	cm.IncCounter(ctx, counterName, 1)

	// step: check if the counter is generated with the context tgs
	scope := getScope(ctx, root)
	ts := scope.(tally.TestScope)
	ss := ts.Snapshot()
	assert.Equal(t, 1, len(ss.Counters()))

	// step: examine only the first counter
	for _, counterSnapshot := range ss.Counters() {
		assert.Equal(t, ctxTags, counterSnapshot.Tags())
		break
	}
}

func TestWithRequestFields(t *testing.T) {
	uid := uuid.New()
	ctx := withRequestUUID(context.TODO(), uid)

	u := ctx.Value(requestUUIDKey)

	assert.NotNil(t, ctx)
	assert.Equal(t, uid, u)
}

func TestGetRequestUUIDFromCtx(t *testing.T) {
	uid := uuid.New()
	ctx := withRequestUUID(context.TODO(), uid)

	requestUUID := RequestUUIDFromCtx(ctx)

	assert.NotNil(t, ctx)
	assert.Equal(t, uid, requestUUID)

	// Test Default Scenario where no uuid exists in the context
	requestUUID = RequestUUIDFromCtx(context.TODO())
	assert.Equal(t, "", requestUUID)
}

func TestWithRoutingDelegate(t *testing.T) {
	expected := "somewhere"
	ctx := WithRoutingDelegate(context.TODO(), expected)
	rd := ctx.Value(routingDelegateKey)
	routingDelegate, ok := rd.(string)

	assert.True(t, ok)
	assert.Equal(t, routingDelegate, expected)
}

func TestGetRoutingDelegateFromCtx(t *testing.T) {
	expected := "somewhere"
	ctx := WithRoutingDelegate(context.TODO(), expected)
	rd := GetRoutingDelegateFromCtx(ctx)

	assert.Equal(t, expected, rd)
}

func TestWithShardKey(t *testing.T) {
	expected := "myshardkey"
	ctx := WithShardKey(context.TODO(), expected)
	sk := ctx.Value(shardKey)
	shardKey, ok := sk.(string)

	assert.True(t, ok)
	assert.Equal(t, shardKey, expected)
}

func TestGetShardKeyFromCtx(t *testing.T) {
	expected := "myshardkey"
	ctx := WithShardKey(context.TODO(), expected)
	sk := GetShardKeyFromCtx(ctx)

	assert.Equal(t, expected, sk)
}

func TestContextLogger(t *testing.T) {
	zapLoggerCore, logs := observer.New(zap.DebugLevel)
	zapLogger := zap.New(zapLoggerCore)
	contextLogger := NewContextLogger(zapLogger)
	ctx := context.Background()
	ctxWithField := WithLogFields(ctx, zap.String("ctxField", "ctxValue"))

	var logMessages []observer.LoggedEntry

	contextLogger.Debug(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 1)
	assert.Equal(t, zap.DebugLevel, logMessages[0].Level)
	assert.Equal(t, logMessages[0].Context[0].Key, "ctxField")
	assert.Equal(t, logMessages[0].Context[0].String, "ctxValue")
	assert.Equal(t, logMessages[0].Context[1].Key, "argField")
	assert.Equal(t, logMessages[0].Context[1].String, "argValue")

	contextLogger.Info(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 1)
	assert.Equal(t, zap.InfoLevel, logMessages[0].Level)

	contextLogger.Warn(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 1)
	assert.Equal(t, zap.WarnLevel, logMessages[0].Level)

	contextLogger.Error(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 1)
	assert.Equal(t, zap.ErrorLevel, logMessages[0].Level)
}

func TestContextLogger_DefaultZ(t *testing.T) {
	zapLoggerCore, logs := observer.New(zap.DebugLevel)
	zapLogger := zap.New(zapLoggerCore)
	contextLogger := NewContextLogger(zapLogger)
	ctx := context.Background()
	ctxWithField := WithLogFields(ctx, zap.String("ctxField", "ctxValue"))

	var logMessages []observer.LoggedEntry

	contextLogger.DebugZ(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 1)
	assert.Equal(t, zap.DebugLevel, logMessages[0].Level)
	assert.Equal(t, logMessages[0].Context[0].Key, "ctxField")
	assert.Equal(t, logMessages[0].Context[0].String, "ctxValue")
	assert.Equal(t, logMessages[0].Context[1].Key, "argField")
	assert.Equal(t, logMessages[0].Context[1].String, "argValue")

	contextLogger.InfoZ(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 1)
	assert.Equal(t, zap.InfoLevel, logMessages[0].Level)

	contextLogger.WarnZ(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 1)
	assert.Equal(t, zap.WarnLevel, logMessages[0].Level)

	contextLogger.ErrorZ(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 1)
	assert.Equal(t, zap.ErrorLevel, logMessages[0].Level)
}

func TestContextLogger_SkipZanzibarLogsZ(t *testing.T) {
	zapLoggerCore, logs := observer.New(zap.DebugLevel)
	zapLogger := zap.New(zapLoggerCore)
	contextLogger := NewContextLogger(zapLogger)
	contextLogger.SetSkipZanzibarLogs(true)
	ctx := context.Background()
	ctxWithField := WithLogFields(ctx, zap.String("ctxField", "ctxValue"))

	var logMessages []observer.LoggedEntry

	contextLogger.DebugZ(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 0)

	contextLogger.InfoZ(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 0)

	contextLogger.WarnZ(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 0)

	contextLogger.ErrorZ(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 0)

	contextLogger.PanicZ(ctxWithField, "msg", zap.String("argField", "argValue"))
	logMessages = logs.TakeAll()
	assert.Len(t, logMessages, 0)
}

func TestContextLoggerPanic(t *testing.T) {
	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()

	zapNop := zap.NewNop()

	contextLogger := NewContextLogger(zapNop)
	ctx := context.Background()

	contextLogger.Panic(ctx, "msg", zap.String("argField", "argValue"))
}

func TestContextLoggerPanic_DefaultZ(t *testing.T) {
	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()

	zapNop := zap.NewNop()

	contextLogger := NewContextLogger(zapNop)
	ctx := context.Background()

	contextLogger.PanicZ(ctx, "msg", zap.String("argField", "argValue"))
}

func TestExtractScopeTag(t *testing.T) {
	headers := map[string]string{"x-uber-region-id": "san_francisco"}
	ctx := WithEndpointRequestHeadersField(context.TODO(), headers)
	contextScopeExtractors := []ContextScopeTagsExtractor{func(ctx context.Context) map[string]string {
		headers := GetEndpointRequestHeadersFromCtx(ctx)
		return map[string]string{"region-id": headers["x-uber-region-id"]}
	}}

	expected := map[string]string{"region-id": "san_francisco"}
	extractors := &ContextExtractors{
		ScopeTagsExtractors: contextScopeExtractors,
	}

	tags := extractors.ExtractScopeTags(ctx)
	assert.Equal(t, tags, expected)
}

func TestExtractLogField(t *testing.T) {
	headers := map[string]string{"x-uber-region-id": "san_francisco"}
	ctx := WithEndpointRequestHeadersField(context.TODO(), headers)
	contextLogFieldsExtractor := []ContextLogFieldsExtractor{func(ctx context.Context) []zap.Field {
		var fields []zap.Field
		headers := GetEndpointRequestHeadersFromCtx(ctx)
		fields = append(fields, zap.String("region-id", headers["x-uber-region-id"]))
		return fields
	}}

	var expected []zap.Field
	expected = append(expected, zap.String("region-id", "san_francisco"))
	extractors := &ContextExtractors{
		LogFieldsExtractors: contextLogFieldsExtractor,
	}
	fields := extractors.ExtractLogFields(ctx)
	assert.Equal(t, expected, fields)
}

func TestAccumulateLogMsgAndFieldsInContext(t *testing.T) {
	ctx := accumulateLogMsgAndFieldsInContext(context.TODO(), "message1",
		[]zap.Field{zap.String("ctxField1", "ctxFieldValue1")}, zapcore.ErrorLevel)
	ctx = accumulateLogMsgAndFieldsInContext(ctx, "message2",
		[]zap.Field{zap.String("ctxField1", "ctxFieldValue2")}, zapcore.ErrorLevel)
	logFields := GetLogFieldsFromCtx(ctx)
	assert.Equal(t, []zap.Field{
		zap.String("msg1", "message1"),
		zap.String("ctxField1", "ctxFieldValue1"),
		zap.String("msg2", "message2"),
		zap.String("ctxField1", "ctxFieldValue2"),
	}, logFields)
}

func TestAccumulateLogMsgAndFieldsInContextWithLogLevel(t *testing.T) {
	ctx := accumulateLogMsgAndFieldsInContext(context.TODO(), "message",
		[]zap.Field{}, zapcore.DebugLevel)
	logLevel := ctx.Value(ctxLogLevel).(zapcore.Level)
	assert.Equal(t, zapcore.DebugLevel, logLevel)
	ctx = accumulateLogMsgAndFieldsInContext(ctx, "message",
		[]zap.Field{}, zapcore.ErrorLevel)
	logLevel = ctx.Value(ctxLogLevel).(zapcore.Level)
	assert.Equal(t, zapcore.ErrorLevel, logLevel)
	ctx = accumulateLogMsgAndFieldsInContext(ctx, "message",
		[]zap.Field{}, zapcore.DebugLevel)
	logLevel = ctx.Value(ctxLogLevel).(zapcore.Level)
	assert.Equal(t, zapcore.ErrorLevel, logLevel)
	ctx = accumulateLogMsgAndFieldsInContext(ctx, "message",
		[]zap.Field{}, zapcore.InfoLevel)
	logLevel = ctx.Value(ctxLogLevel).(zapcore.Level)
	assert.Equal(t, zapcore.ErrorLevel, logLevel)
	ctx = accumulateLogMsgAndFieldsInContext(ctx, "message",
		[]zap.Field{}, zapcore.WarnLevel)
	logLevel = ctx.Value(ctxLogLevel).(zapcore.Level)
	assert.Equal(t, zapcore.ErrorLevel, logLevel)
	ctx = accumulateLogMsgAndFieldsInContext(ctx, "message",
		[]zap.Field{}, zapcore.DPanicLevel)
	logLevel = ctx.Value(ctxLogLevel).(zapcore.Level)
	assert.Equal(t, zapcore.DPanicLevel, logLevel)
	ctx = accumulateLogMsgAndFieldsInContext(ctx, "message",
		[]zap.Field{}, zapcore.PanicLevel)
	logLevel = ctx.Value(ctxLogLevel).(zapcore.Level)
	assert.Equal(t, zapcore.PanicLevel, logLevel)
	ctx = accumulateLogMsgAndFieldsInContext(ctx, "message",
		[]zap.Field{}, zapcore.FatalLevel)
	logLevel = ctx.Value(ctxLogLevel).(zapcore.Level)
	assert.Equal(t, zapcore.FatalLevel, logLevel)
	ctx = accumulateLogMsgAndFieldsInContext(ctx, "message",
		[]zap.Field{}, zapcore.PanicLevel)
	logLevel = ctx.Value(ctxLogLevel).(zapcore.Level)
	assert.Equal(t, zapcore.FatalLevel, logLevel)
}

func TestGetCtxLogLevelOrDebugLevelFromCtx(t *testing.T) {
	ctx := accumulateLogMsgAndFieldsInContext(context.TODO(), "message",
		[]zap.Field{}, zapcore.DebugLevel)
	logLevel := GetCtxLogLevelOrDebugLevelFromCtx(ctx)
	logCounter := GetCtxLogCounterFromCtx(ctx)
	assert.Equal(t, zapcore.DebugLevel, logLevel)
	assert.Equal(t, 1, logCounter)
	ctx = accumulateLogMsgAndFieldsInContext(ctx, "message",
		[]zap.Field{}, zapcore.ErrorLevel)
	logCounter = GetCtxLogCounterFromCtx(ctx)
	logLevel = GetCtxLogLevelOrDebugLevelFromCtx(ctx)
	assert.Equal(t, zapcore.ErrorLevel, logLevel)
	assert.Equal(t, 2, logCounter)
	ctx = context.TODO()
	logCounter = GetCtxLogCounterFromCtx(ctx)
	logLevel = GetCtxLogLevelOrDebugLevelFromCtx(ctx)
	assert.Equal(t, zapcore.DebugLevel, logLevel)
	assert.Equal(t, 0, logCounter)
}

func TestLogLevelWithLogZ(t *testing.T) {
	zapLoggerCore, _ := observer.New(zap.WarnLevel)
	zapLogger := zap.New(zapLoggerCore)
	contextLogger := NewContextLogger(zapLogger)
	contextLogger.SetSkipZanzibarLogs(true)
	ctx := context.Background()

	// debug logs should not be added as debugLevel < warnLevel
	ctx = contextLogger.DebugZ(ctx, "msg", zap.String("argField", "argValue"))
	logs := GetLogFieldsFromCtx(ctx)
	assert.Len(t, logs, 0)

	// info logs should not be added as infoLevel < warnLevel
	ctx = contextLogger.InfoZ(ctx, "msg", zap.String("argField", "argValue"))
	logs = GetLogFieldsFromCtx(ctx)
	assert.Len(t, logs, 0)

	// warn logs should be added
	ctx = context.Background()
	ctx = contextLogger.WarnZ(ctx, "msg", zap.String("argField", "argValue"))
	logs = GetLogFieldsFromCtx(ctx)
	assert.Len(t, logs, 2)

	// error logs should be added as errorLevel > warnLevel
	ctx = context.Background()
	ctx = contextLogger.ErrorZ(ctx, "msg", zap.String("argField", "argValue"))
	logs = GetLogFieldsFromCtx(ctx)
	assert.Len(t, logs, 2)

	// panic logs should be added as panicLevel > warnLevel
	ctx = context.Background()
	ctx = contextLogger.PanicZ(ctx, "msg", zap.String("argField", "argValue"))
	logs = GetLogFieldsFromCtx(ctx)
	assert.Len(t, logs, 2)
}

func TestAccumulateLogField(t *testing.T) {
	ctxFields := []zapcore.Field{zap.String("ctxField1", "ctxValue1"), zap.String("ctxField2", "ctxValue2")}
	ctx := context.Background()
	// append fields in a way s.t. len(log fields) != cap(log fields)
	ctxWithField := WithLogFields(ctx, ctxFields...)
	ctxWithField = WithLogFields(ctxWithField, zap.String("ctxField3", "ctxValue3"))

	fields1 := accumulateLogFields(ctxWithField, []zapcore.Field{zap.String("argField", "one")})
	fields2 := accumulateLogFields(ctxWithField, []zapcore.Field{zap.String("argField", "two")})

	assert.Len(t, fields1, 4)
	assert.Len(t, fields2, 4)

	// concurrent logs shouldn't affect each other
	assert.Equal(t, "one", fields1[3].String)
	assert.Equal(t, "two", fields2[3].String)
}

func TestTimeoutAndRetry(t *testing.T) {
	tro := BuildTimeoutAndRetryConfig(1, 1, 1, 1)
	ctx := context.Background()
	ctx = WithTimeAndRetryOptions(ctx, tro)
	tro1 := GetTimeoutAndRetryOptions(ctx)

	assert.Equal(t, tro.BackOffTimeAcrossRetriesInMs, tro1.BackOffTimeAcrossRetriesInMs)
	assert.Equal(t, tro.OverallTimeoutInMs, tro1.OverallTimeoutInMs)
	assert.Equal(t, tro.RequestTimeoutPerAttemptInMs, tro1.RequestTimeoutPerAttemptInMs)
	assert.Equal(t, tro.MaxAttempts, tro1.MaxAttempts)
}

func TestTimeoutAndRetry_NotSet(t *testing.T) {
	tro := GetTimeoutAndRetryOptions(context.Background())
	assert.Equal(t, tro, (*TimeoutAndRetryOptions)(nil))
}

func TestSafeLogFields(t *testing.T) {
	ctx := context.Background()
	ctx = WithSafeLogFields(ctx)
	AppendLogFieldsToContext(ctx, zap.String("foo", "bar"))
	AppendLogFieldsToContext(ctx, zap.String("hello", "world"))
	fields := GetLogFieldsFromCtx(ctx)

	assert.Len(t, fields, 2)
	assert.Equal(t, zap.String("foo", "bar"), fields[0])
	assert.Equal(t, zap.String("hello", "world"), fields[1])
}

func TestAppendLogFieldsToContext(t *testing.T) {
	condition := "when context doesn't contain safeLogFields container"
	t.Run(condition, func(t *testing.T) {
		ctx := context.Background()
		AppendLogFieldsToContext(ctx, zap.String("foo", "bar"))
		fields := GetLogFieldsFromCtx(ctx)
		assert.NotNil(t, fields)
		assert.Len(t, fields, 0)
	})
}
