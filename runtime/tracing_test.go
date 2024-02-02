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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
)

func TestExtractSpanLogFields(t *testing.T) {
	t.Run("nil span", func(t *testing.T) {
		assert.Empty(t, extractSpanLogFields(nil))
	})

	t.Run("jaeger span", func(t *testing.T) {
		tracer, closer, err := config.Configuration{
			ServiceName: "test",
			Sampler: &config.SamplerConfig{
				Type:  "const",
				Param: 1,
			}}.NewTracer(config.Reporter(jaeger.NewInMemoryReporter()))
		require.NoError(t, err)
		defer closer.Close()

		span := tracer.StartSpan("op")
		jSpan, ok := span.(*jaeger.Span)
		require.True(t, ok)
		fields := extractSpanLogFields(span)

		assert.Equal(t, 3, len(fields))
		assert.Equal(t, []zap.Field{
			zap.String(TraceIDKey, jSpan.SpanContext().TraceID().String()),
			zap.String(TraceSpanKey, jSpan.SpanContext().SpanID().String()),
			zap.Bool(TraceSampledKey, jSpan.SpanContext().IsSampled()),
		}, fields)
	})
}
