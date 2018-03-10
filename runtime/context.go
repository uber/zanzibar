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
)

type contextFieldKey string

const (
	requestUUIDKey = contextFieldKey("requestUUID")
)

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
