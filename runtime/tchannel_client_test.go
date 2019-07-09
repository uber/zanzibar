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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.Len(t, fields, 4)
	assert.Equal(t, fields[0].Key, "remoteAddr")
	// nil call should cause remoteAddr to be set to unknown
	assert.Equal(t, fields[0].String, "unknown")
	assert.Equal(t, fields[3].Key, "foo")
	assert.Equal(t, fields[3].String, "bar")
}
