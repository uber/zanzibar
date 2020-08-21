// Copyright (c) 2020 Uber Technologies, Inc.
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
	"github.com/uber/zanzibar/runtime/jsonwrapper"
	"os"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
)

func TestCreatGatewayLoggingConfig(t *testing.T) {
	cfg := NewStaticConfigOrDie(nil, map[string]interface{}{
		"logger.fileName": "foober",
		"logger.output":   "",
		"env":             "local",
		"datacenter":      "phx2",
		"serviceName":     "foozabar",
		"logger.level":    "fatal",
	})

	g := Gateway{
		Config: cfg,
	}

	err := g.setupLogger(cfg)
	assert.NoError(t, err)
	assert.Equal(t, zap.NewAtomicLevelAt(zap.FatalLevel), *g.atomLevel)
}

func TestGetServiceNameFromEnv(t *testing.T) {
	cfg := NewStaticConfigOrDie(nil, map[string]interface{}{
		"logger.level":                       "fatal",
		"http.port":                          int64(1234),
		"tchannel.port":                      int64(5678),
		"metrics.flushInterval":              1000,
		"metrics.runtime.collectInterval":    1000,
		"metrics.runtime.enableCPUMetrics":   false,
		"metrics.runtime.enableMemMetrics":   false,
		"metrics.runtime.enableGCMetrics":    false,
		"useDatacenter":                      false,
		"metrics.m3.includeHost":             false,
		"envVarsToTagInRootScope":            []string{},
		"metrics.m3.maxPacketSizeBytes":      int64(99999),
		"metrics.m3.maxQueueSize":            int64(9999),
		"metrics.m3.hostPort":                "127.0.0.1:8053",
		"metrics.type":                       "m3",
		"jaeger.disabled":                    true,
		"jaeger.reporter.flush.milliseconds": 10000,
		"jaeger.reporter.hostport":           "localhost:6831",
		"jaeger.sampler.param":               0.001,
		"jaeger.sampler.type":                "remote",
		"logger.fileName":                    "foober",
		"logger.output":                      "",
		"subLoggerLevel.jaeger":              "info",
		"subLoggerLevel.http":                "info",
		"subLoggerLevel.tchannel":            "info",
		"env":                                "local",
		"datacenter":                         "xyz1",
		"tchannel.serviceName":               "test",
		"tchannel.processName":               "test",
		"sidecarRouter.default.grpc.ip":      "127.0.0.1",
		"sidecarRouter.default.grpc.port":    4998,
		"grpc.clientServiceNameMapping":      map[string]string{"test":"test"},
		"serviceName":                        "not-overridden",
		"metrics.serviceName":                "not-overridden",
		"serviceNameEnv":                     "TEST",
		"metrics.serviceNameEnv":             "TEST",
	})

	var metricsBackend tally.CachedStatsReporter
	opts := &Options{
		GetContextScopeExtractors: nil,
		GetContextFieldExtractors: nil,
		JSONWrapper:               jsonwrapper.NewDefaultJSONWrapper(),
		MetricsBackend:            metricsBackend,
	}

	os.Setenv("TEST", "overridden")

	g1, err := CreateGateway(cfg, opts)
	assert.Nil(t, err)
	assert.Equal(t, g1.ServiceName, "overridden")
	os.Unsetenv("TEST")
}

func TestCreatGatewayBadLoggingConfig(t *testing.T) {
	cfg := NewStaticConfigOrDie(nil, map[string]interface{}{
		"logger.level": "invalid",
	})

	g := Gateway{
		Config: cfg,
	}

	err := g.setupLogger(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown log level for gateway logger: invalid")
}
