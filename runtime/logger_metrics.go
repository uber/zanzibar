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

package zanzibar

import (
	"github.com/uber-go/tally"
	"go.uber.org/zap/zapcore"
)

type loggingZapCore struct {
	zapcore.Core

	allHostsScope tally.Scope

	debugCounter  tally.Counter
	infoCounter   tally.Counter
	warnCounter   tally.Counter
	errorCounter  tally.Counter
	dPanicCounter tally.Counter
	panicCounter  tally.Counter
	fatalCounter  tally.Counter
}

// NewLoggingZapCore will return a zapcore.Core that emits
// "logged" metrics, one counter for each level.
func NewLoggingZapCore(
	core zapcore.Core,
	allHostsScope tally.Scope,
) zapcore.Core {
	return &loggingZapCore{
		Core:          core,
		allHostsScope: allHostsScope,

		debugCounter:  allHostsScope.Counter("zap.logged.debug"),
		infoCounter:   allHostsScope.Counter("zap.logged.info"),
		warnCounter:   allHostsScope.Counter("zap.logged.warn"),
		errorCounter:  allHostsScope.Counter("zap.logged.error"),
		dPanicCounter: allHostsScope.Counter("zap.logged.dpanic"),
		panicCounter:  allHostsScope.Counter("zap.logged.panic"),
		fatalCounter:  allHostsScope.Counter("zap.logged.fatal"),
	}
}

func copyLoggingZapCore(
	core zapcore.Core,
	lcore *loggingZapCore,
) zapcore.Core {
	return &loggingZapCore{
		Core:          core,
		allHostsScope: lcore.allHostsScope,
		debugCounter:  lcore.debugCounter,
		infoCounter:   lcore.infoCounter,
		warnCounter:   lcore.warnCounter,
		errorCounter:  lcore.errorCounter,
		dPanicCounter: lcore.dPanicCounter,
		panicCounter:  lcore.panicCounter,
		fatalCounter:  lcore.fatalCounter,
	}
}

func (lcore *loggingZapCore) With(fields []zapcore.Field) zapcore.Core {
	newCore := lcore.Core.With(fields)
	return copyLoggingZapCore(newCore, lcore)
}

func (lcore *loggingZapCore) Check(
	ent zapcore.Entry, ce *zapcore.CheckedEntry,
) *zapcore.CheckedEntry {
	if lcore.Enabled(ent.Level) {
		return ce.AddCore(ent, lcore)
	}
	return ce
}

func (lcore *loggingZapCore) Write(
	entry zapcore.Entry, fields []zapcore.Field,
) error {
	switch entry.Level {
	case zapcore.DebugLevel:
		lcore.debugCounter.Inc(1)
	case zapcore.InfoLevel:
		lcore.infoCounter.Inc(1)
	case zapcore.WarnLevel:
		lcore.warnCounter.Inc(1)
	case zapcore.ErrorLevel:
		lcore.errorCounter.Inc(1)
	case zapcore.DPanicLevel:
		lcore.dPanicCounter.Inc(1)
	case zapcore.PanicLevel:
		lcore.panicCounter.Inc(1)
	case zapcore.FatalLevel:
		lcore.fatalCounter.Inc(1)
	default:
		// noop
	}

	return lcore.Core.Write(entry, fields)
}
