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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber/tchannel-go"
	"go.uber.org/zap"
)

func TestNilCallReferenceForLogger(t *testing.T) {
	headers := map[string]string{
		"header-key": "header-value",
	}
	staticTestTime := time.Unix(1500000000, 0)
	outboundCall := &tchannelOutboundCall{
		methodName:    "Get",
		serviceMethod: "Test",
		startTime:     staticTestTime,
		finishTime:    staticTestTime,
		reqHeaders:    headers,
		resHeaders:    headers,
	}
	ctx := context.TODO()
	ctx = WithLogFields(ctx, zap.String("foo", "bar"))

	fields := outboundCall.logFields(ctx)

	// one field for each of the:
	// timestamp-started, timestamp-finished, remoteAddr, requestHeader, responseHeader
	assert.Len(t, fields, 6)

	var addr, reqKey, resKey, foo bool
	for _, f := range fields {
		switch f.Key {
		case "remoteAddr":
			assert.Equal(t, f.String, "unknown")
			addr = true
		case "Client-Req-Header-header-key":
			assert.Equal(t, f.String, "header-value")
			reqKey = true
		case "Client-Res-Header-header-key":
			assert.Equal(t, f.String, "header-value")
			resKey = true
		case "foo":
			assert.Equal(t, f.String, "bar")
			foo = true
		}
	}
	assert.True(t, addr, "remoteAddr key not present")
	assert.True(t, reqKey, "Client-Req-Header-header-key key not present")
	assert.True(t, resKey, "Client-Res-Header-header-key key not present")
	assert.True(t, foo, "foo key not present")
}

func TestMaxAttempts(t *testing.T) {
	methodName := map[string]string{
		"methodKey": "methodValue",
	}
	tChannelClient := &TChannelClient{
		maxAttempts: 2,
		serviceName: "test",
		methodNames: methodName,
		timeout:     1,
	}
	ctx := context.TODO()
	retryOpts := tchannel.RetryOptions{
		MaxAttempts: 2,
	}
	contextBuilder := tchannel.NewContextBuilder(tChannelClient.timeout).SetParentContext(ctx).SetRetryOptions(&retryOpts)
	maxAttempts := contextBuilder.RetryOptions.MaxAttempts
	assert.Equal(t, 2, maxAttempts)
}

func TestMaxAttemptsDefault(t *testing.T) {
	methodName := map[string]string{
		"methodKey": "methodValue",
	}
	tChannelClient := &TChannelClient{
		maxAttempts: 2,
		serviceName: "test",
		methodNames: methodName,
		timeout:     1,
	}
	ctx := context.TODO()
	retryOpts := tchannel.RetryOptions{}
	contextBuilder := tchannel.NewContextBuilder(tChannelClient.timeout).SetParentContext(ctx).SetRetryOptions(&retryOpts)
	maxAttempts := contextBuilder.RetryOptions.MaxAttempts
	assert.Equal(t, maxAttempts, 0)
}

func TestMaxAttemptFieldWithNoSet(t *testing.T) {
	methodName := map[string]string{
		"methodKey": "methodValue",
	}
	tChannelClient := &TChannelClient{
		serviceName: "test",
		methodNames: methodName,
		timeout:     1,
	}
	ctx := context.TODO()
	retryOpts := tchannel.RetryOptions{}
	contextBuilder := tchannel.NewContextBuilder(tChannelClient.timeout).SetParentContext(ctx).SetRetryOptions(&retryOpts)
	maxAttempts := contextBuilder.RetryOptions.MaxAttempts
	assert.Equal(t, maxAttempts, 0)
}
