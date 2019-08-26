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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"go.uber.org/yarpc/yarpcerrors"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

const (
	serviceName            = "Echo"
	clientID               = "Echo"
	methodName             = "Echo"
	routingKey             = "routingKey"
	requestUUIDHeaderKey   = "reqID"
	circuitBreakerDisabled = false
	timeoutInMS            = 10
	serviceMethod          = "Echo::Echo"
)

var (
	scopeExtractor = func(ctx context.Context) map[string]string {
		tags := map[string]string{}
		headers := GetEndpointRequestHeadersFromCtx(ctx)
		tags["regionname"] = headers["Regionname"]
		tags["device"] = headers["Device"]
		tags["deviceversion"] = headers["Deviceversion"]

		return tags
	}
	logFieldsExtractors = func(ctx context.Context) []zap.Field {
		reqHeaders := GetEndpointRequestHeadersFromCtx(ctx)
		fields := make([]zap.Field, 0, len(reqHeaders))
		for k, v := range reqHeaders {
			fields = append(fields, zap.String(k, v))
		}
		return fields
	}
	logger     = zap.NewNop()
	metrics    = NewContextMetrics(tally.NoopScope)
	extractors = &ContextExtractors{
		ScopeTagsExtractors: []ContextScopeTagsExtractor{scopeExtractor},
		LogFieldsExtractors: []ContextLogFieldsExtractor{logFieldsExtractors},
	}
	methodNames = map[string]string{
		serviceMethod: methodName,
	}
	expectedTimeout = time.Duration(timeoutInMS) * time.Millisecond
	expectedLoggers = map[string]*zap.Logger{
		serviceMethod: logger,
	}
	expectedScopeTags = map[string]map[string]string{
		serviceMethod: {
			scopeTagClient:         clientID,
			scopeTagClientMethod:   methodName,
			scopeTagsTargetService: serviceName,
		},
	}
)

func TestNewGRPCClientOpts(t *testing.T) {
	actual := NewGRPCClientOpts(
		logger,
		metrics,
		extractors,
		methodNames,
		clientID,
		serviceName,
		routingKey,
		requestUUIDHeaderKey,
		circuitBreakerDisabled,
		timeoutInMS,
	)
	expected := &GRPCClientOpts{
		serviceName,
		clientID,
		methodNames,
		expectedLoggers,
		metrics,
		extractors,
		routingKey,
		requestUUIDHeaderKey,
		circuitBreakerDisabled,
		expectedTimeout,
		expectedScopeTags,
	}
	assert.Equal(t, expected, actual)
}

func TestGRPCCallHelper(t *testing.T) {
	ctx := context.Background()
	opts := NewGRPCClientOpts(
		logger,
		metrics,
		extractors,
		methodNames,
		clientID,
		serviceName,
		routingKey,
		requestUUIDHeaderKey,
		circuitBreakerDisabled,
		timeoutInMS,
	)
	_, actual := NewGRPCClientCallHelper(ctx, serviceMethod, opts)
	expected := &callHelper{
		logger:    expectedLoggers[serviceMethod],
		metrics:   metrics,
		extractor: extractors,
	}
	assert.Equal(t, expected, actual)
}

func testCallHelper(t *testing.T, err error) {
	helper := &callHelper{
		logger:    logger,
		metrics:   metrics,
		extractor: extractors,
	}

	assert.Zero(t, helper.startTime, "startTime not initialized to zero")
	assert.Zero(t, helper.finishTime, "finishTime not initialized to zero")
	helper.Start()
	assert.NotZero(t, helper.startTime, "startTime didn't update after calling Start()")
	assert.Zero(t, helper.finishTime, "finishTime update after calling Start()")

	// Adding sleep just to make sure startTime and finishTime are never same.
	time.Sleep(10 * time.Millisecond)

	ctx := context.Background()
	helper.Finish(ctx, err)
	assert.NotZero(t, helper.startTime, "startTime initialized to zero calling Finish()")
	assert.NotZero(t, helper.finishTime, "finishTime initialized to zero calling Finish()")
}

func TestCallHelper(t *testing.T) {
	testCallHelper(t, nil)
	testCallHelper(t, errors.New("mock error"))
	testCallHelper(t, yarpcerrors.Newf(1, "CodeCancelled"))
}
