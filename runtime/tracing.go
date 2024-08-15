// Copyright (c) 2024 Uber Technologies, Inc.
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
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	tracingComponentTag = opentracing.Tag{Key: string(ext.Component), Value: "zanzibar"}
)

// The span context struct in otel is not exportable, an this is the way to access fields in span context
// https://github.com/open-telemetry/opentelemetry-go/blob/v1.16.0/bridge/opentracing/README.md#extended-functionality
type otelSpanContextProvider interface {
	TraceID() trace.TraceID
	SpanID() trace.SpanID
	IsSampled() bool
}

// extractSpanLogFields extracts zap log fields from an opentracing span
func extractSpanLogFields(span opentracing.Span) []zap.Field {
	var fields []zap.Field
	if span != nil {
		if jc, ok := span.Context().(jaeger.SpanContext); ok {
			fields = append(fields,
				zap.String(TraceIDKey, jc.TraceID().String()),
				zap.String(TraceSpanKey, jc.SpanID().String()),
				zap.Bool(TraceSampledKey, jc.IsSampled()),
			)
		} else if oc, ok := span.Context().(otelSpanContextProvider); ok {
			fields = append(fields,
				zap.String(TraceIDKey, oc.TraceID().String()),
				zap.String(TraceSpanKey, oc.SpanID().String()),
				zap.Bool(TraceSampledKey, oc.IsSampled()),
			)
		}
	}
	return fields
}

// updateClientSpanWithError updates a client span with the error tags and logs
func updateClientSpanWithError(span opentracing.Span, res *http.Response, err error) {
	if span == nil {
		return
	}

	if res != nil {
		ext.HTTPStatusCode.Set(span, uint16(res.StatusCode))
		if res.StatusCode >= 400 {
			ext.Error.Set(span, true)
		}
	}
	if err != nil {
		ext.Error.Set(span, true)
		ext.LogError(span, err)
	}
}
