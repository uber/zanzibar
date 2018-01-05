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
	"github.com/stretchr/testify/assert"

	"testing"
	"time"
)

func TestNilCallReferenceForLogger(t *testing.T) {

	headers := map[string]string{
		"header-key": "header-value",
	}
	outboundCall := &tchannelOutboundCall{
		client:        nil,
		call:          nil,
		methodName:    "GET",
		serviceMethod: "Test",
		success:       false,
		startTime:     time.Now(),
		finishTime:    time.Now(),
		reqHeaders:    headers,
		resHeaders:    headers,
		logger:        nil,
		metrics:       nil,
	}
	fields := outboundCall.logFields()

	// one field for each of the the:
	// timestamp-started, timestamp-finished, remoteAddr, requestHeader, responseHeader
	assert.Len(t, fields, 5)
	assert.Equal(t, fields[0].Key, "remoteAddr")
	// nil call should cause remoteAddr to be set to unknown
	assert.Equal(t, fields[0].String, "unknown")
}
