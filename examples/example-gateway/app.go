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

package app

import (
	"net/textproto"

	"go.uber.org/zap"

	"context"

	zanzibar "github.com/uber/zanzibar/runtime"
)

// AppOptions defines the custom application func
var AppOptions = &zanzibar.Options{
	GetContextScopeExtractors: getContextScopeTagExtractors,
	GetContextFieldExtractors: getContextLogFieldExtractors,
}

func getContextScopeTagExtractors() []zanzibar.ContextScopeTagsExtractor {
	extractors := []zanzibar.ContextScopeTagsExtractor{
		getRequestTags,
	}

	return extractors
}

func getContextLogFieldExtractors() []zanzibar.ContextLogFieldsExtractor {
	extractors := []zanzibar.ContextLogFieldsExtractor{
		getRequestFields,
	}

	return extractors
}

func getRequestTags(ctx context.Context) map[string]string {
	tags := map[string]string{}
	headers := zanzibar.GetEndpointRequestHeadersFromCtx(ctx)
	tags["regionname"] = headers["Regionname"]
	tags["device"] = headers["Device"]
	tags["deviceversion"] = headers["Deviceversion"]

	return tags
}

func getRequestFields(ctx context.Context) []zap.Field {
	var fields []zap.Field
	headers := zanzibar.GetEndpointRequestHeadersFromCtx(ctx)

	for k, v := range headers {
		if textproto.CanonicalMIMEHeaderKey("x-token") == textproto.CanonicalMIMEHeaderKey(k) {
			continue
		}
		fields = append(fields, zap.String(k, v))
	}
	return fields
}
