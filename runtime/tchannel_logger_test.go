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

package zanzibar_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	zanzibar "github.com/uber/zanzibar/runtime"

	"github.com/uber/tchannel-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestTChannelLoggerEnabled(t *testing.T) {
	t2z := map[tchannel.LogLevel]zapcore.Level{
		tchannel.LogLevelDebug: zap.DebugLevel,
		tchannel.LogLevelInfo:  zap.InfoLevel,
		tchannel.LogLevelWarn:  zap.WarnLevel,
		tchannel.LogLevelError: zap.ErrorLevel,
		tchannel.LogLevelFatal: zap.FatalLevel,
	}
	for tlevel := tchannel.LogLevelFatal; tlevel > tchannel.LogLevelAll; tlevel-- {
		withLogger(t2z[tlevel], func(logger tchannel.Logger, logs *observer.ObservedLogs) {
			for l := tchannel.LogLevelAll; l <= tchannel.LogLevelFatal; l++ {
				assert.Equal(t, tlevel <= l, logger.Enabled(l), "levelLogger.Enabled(%v) at %v", l, tlevel)
			}
		})
	}
}
func TestTChannelLoggerLeveledMethods(t *testing.T) {
	withLogger(zap.DebugLevel, func(logger tchannel.Logger, logs *observer.ObservedLogs) {
		tests := []struct {
			method        func(string)
			methodf       func(string, ...interface{})
			expectedLevel zapcore.Level
		}{
			{method: logger.Debug, expectedLevel: zap.DebugLevel},
			{method: logger.Info, expectedLevel: zap.InfoLevel},
			{method: logger.Warn, expectedLevel: zap.WarnLevel},
			{method: logger.Error, expectedLevel: zap.ErrorLevel},
			{methodf: logger.Infof, expectedLevel: zap.InfoLevel},
			{methodf: logger.Debugf, expectedLevel: zap.DebugLevel},
		}
		for i, tt := range tests {
			if tt.method == nil {
				tt.methodf("%s", "")
			} else {
				tt.method("")
			}
			output := logs.AllUntimed()
			assert.Equal(t, i+1, len(output), "Unexpected number of logs.")
			assert.Equal(t, 0, len(output[i].Context), "Unexpected context on first log.")
			assert.Equal(
				t,
				zapcore.Entry{Level: tt.expectedLevel},
				output[i].Entry,
				"Unexpected output from %s-level logger method.", tt.expectedLevel)
		}
	})
}

func TestTChannelLoggerWithFields(t *testing.T) {
	withLogger(zap.DebugLevel, func(logger tchannel.Logger, logs *observer.ObservedLogs) {
		assert.Nil(t, logger.Fields(), "Fields() always return nil")

		fields := []tchannel.LogField{
			{Key: "a", Value: "foo"},
			{Key: "b", Value: 42},
		}
		expectedFields := []zapcore.Field{}
		for _, f := range fields {
			expectedFields = append(expectedFields, zap.Any(f.Key, f.Value))
		}

		lwf := logger.WithFields(fields...)
		lwf.Info("hi")
		log := logs.AllUntimed()[0]

		assert.Nil(t, lwf.Fields(), "Fields() always return nil")
		assert.Equal(
			t,
			expectedFields,
			log.Context,
			"Unexpected output from %s-level logger method.", zap.InfoLevel,
		)
	})

}

func withLogger(l zapcore.LevelEnabler, f func(tchannel.Logger, *observer.ObservedLogs)) {
	core, logs := observer.New(l)
	logger := zanzibar.NewTChannelLogger(zap.New(core))
	f(logger, logs)
}
