// Copyright (c) 2017 Uber Technologies, Inc.
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
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber-go/tally"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLoggingZapCore(t *testing.T) {
	metricsScope := tally.NewTestScope("test", nil)

	tempLogger := zap.New(
		zanzibar.NewLoggingZapCore(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
				zapcore.AddSync(ioutil.Discard),
				zap.InfoLevel,
			),
			metricsScope,
		),
	)

	tempLogger.Debug("debug msg")
	tempLogger.Info("info msg")
	tempLogger.Warn("warn msg")
	tempLogger.Error("error msg")

	assert.Panics(t, func() {
		tempLogger.DPanic("dpanic msg")
		tempLogger.Panic("panic msg")
		tempLogger.Fatal("fatal msg")
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
