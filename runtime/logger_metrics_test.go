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

package zanzibar_test

import (
	"context"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"testing"
)

func newTempLogger(scope tally.TestScope) *zap.Logger {
	if scope == nil {
		scope = tally.NewTestScope("test", nil)
	}
	return zap.New(
		zanzibar.NewInstrumentedZapCore(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
				zapcore.AddSync(ioutil.Discard),
				zap.DebugLevel,
			),
			scope,
		),
	)
}

func TestLoggingZapCore(t *testing.T) {
	metricsScope := tally.NewTestScope("test", nil)
	tempLogger := newTempLogger(metricsScope)

	tempLogger.Debug("debug msg")
	tempLogger.Info("info msg")
	tempLogger.Warn("warn msg")
	tempLogger.Error("error msg")
	tempLogger.DPanic("dpanic msg")

	assert.Panics(t, func() {
		tempLogger.Panic("panic msg")
	})

	snapshot := metricsScope.Snapshot()
	counters := snapshot.Counters()

	assert.Equal(t, 7, len(counters))

	expectedKeys := []string{
		"test.zap.logged.debug+",
		"test.zap.logged.info+",
		"test.zap.logged.warn+",
		"test.zap.logged.error+",
		"test.zap.logged.dpanic+",
		"test.zap.logged.panic+",
		"test.zap.logged.fatal+",
	}

	for _, key := range expectedKeys {
		assert.Contains(t, counters, key, "should contain %s", key)
	}
}

func BenchmarkLogger(b *testing.B) {
	testCases := []struct {
		label            string
		perRequestLogger bool
		logPerRequest    int
		numFieldPerLog   int
	}{
		{"per endpoint, light log", false, 1, 1},
		{"per endpoint, medium log", false, 10, 50},
		{"per endpoint, heavy log", false, 20, 100},
		{"per request, light log", true, 1, 1},
		{"per request, medium log", true, 10, 50},
		{"per request, heavy log", true, 20, 100},
	}

	getLogFields := func(ctx context.Context, n int) []zap.Field {
		zfields := []zap.Field{}
		UUID := ctx.Value("reqUUID").(uuid.UUID).String()
		for i := 0; i < n; i++ {
			zfields = append(zfields, zap.String("reqUUID", UUID))
		}
		return zfields
	}

	for _, tt := range testCases {
		b.Run(tt.label, func(b *testing.B) { // per test cases
			logger := newTempLogger(nil)
			ctx := context.WithValue(context.Background(), "reqUUID", uuid.NewUUID())
			zfields := getLogFields(ctx, tt.numFieldPerLog)

			for i := 0; i < b.N; i++ { // per request
				if tt.perRequestLogger {
					logger = logger.With(zfields...)
					logger.Info("test-msg")
				} else {
					for j := 0; j < tt.logPerRequest; j++ {
						zfields := getLogFields(ctx, tt.numFieldPerLog)
						logger.Info("test-msg", zfields...)
					}
				}
			}
		})
	}
}
