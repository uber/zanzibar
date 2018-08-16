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
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWithEndpointFields(t *testing.T) {
	expected := "someEndpoint"
	ctx := withEndpointField(context.TODO(), expected)

	ek := ctx.Value(endpointKey)
	endpoint, ok := ek.(string)

	assert.True(t, ok)
	assert.Equal(t, endpoint, expected)
}

func TestWithRequestFields(t *testing.T) {
	ctx := withRequestFields(context.TODO())

	u := ctx.Value(requestUUIDKey)
	u1, ok := u.(uuid.UUID)

	assert.NotNil(t, ctx)
	assert.NotNil(t, u)
	assert.NotNil(t, u1)
	assert.True(t, ok)
}

func TestGetRequestUUIDFromCtx(t *testing.T) {
	ctx := withRequestFields(context.TODO())

	requestUUID := GetRequestUUIDFromCtx(ctx)

	assert.NotNil(t, ctx)
	assert.NotNil(t, requestUUID)

	// Test Default Scenario where no uuid exists in the context
	requestUUID = GetRequestUUIDFromCtx(context.TODO())
	assert.Nil(t, requestUUID)
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
