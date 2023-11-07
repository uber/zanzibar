// Copyright (c) 2023 Uber Technologies, Inc.
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
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/tchannel-go"
	"github.com/uber/zanzibar/v2/runtime/jsonwrapper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestCreateGatewayLoggingConfig(t *testing.T) {
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
		"logger.fileName":                    "",
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
		"grpc.clientServiceNameMapping":      map[string]string{"test": "test"},
		"serviceName":                        "not-overridden",
		"metrics.serviceName":                "not-overridden",
		"serviceNameEnv":                     "TEST",
		"metrics.serviceNameEnv":             "TEST",
		"http.notFoundHandler.custom":        true,
		"http.handleMethodNotAllowed":        true,
	})

	var metricsBackend tally.CachedStatsReporter
	opts := &Options{
		GetContextScopeExtractors: nil,
		GetContextFieldExtractors: nil,
		JSONWrapper:               jsonwrapper.NewDefaultJSONWrapper(),
		MetricsBackend:            metricsBackend,
		NotFoundHandler: func(gateway *Gateway) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {}
		},
	}

	if err := os.Setenv("TEST", "overridden"); err != nil {
		t.Errorf("no error expected but got %v", err)
	}
	g1, err := CreateGateway(cfg, opts)
	assert.Nil(t, err)
	assert.Equal(t, g1.ServiceName, "overridden")
	if err = os.Unsetenv("TEST"); err != nil {
		t.Errorf("no error expected but got %v", err)
	}
}

func TestCreateGatewayBadLoggingConfig(t *testing.T) {
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

func TestGatewaySetupClientTChannelWhenServiceNameAlreadyExists(t *testing.T) {
	cfg := NewStaticConfigOrDie(nil, map[string]interface{}{})
	g := Gateway{
		ClientTChannels: map[string]*tchannel.Channel{
			"service-foo": {},
		},
		Logger: zap.NewNop(),
	}
	ch := g.SetupClientTChannel(cfg, "service-foo")
	assert.Equal(t, ch, &tchannel.Channel{})
}

func TestGatewaySetupClientTChannel(t *testing.T) {
	cfg := NewStaticConfigOrDie(nil, map[string]interface{}{
		"tchannel.processName": "test-proc",
		"tchannel.serviceName": "test-gateway",
	})
	g := Gateway{
		TChannelSubLoggerLevel: zapcore.ErrorLevel,
		RootScope:              tally.NoopScope,
		ClientTChannels:        map[string]*tchannel.Channel{},
		Logger:                 zap.NewNop(),
		logEncoder:             zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
	}

	ch := g.SetupClientTChannel(cfg, "service-foo")
	assert.Equal(t, ch, g.ClientTChannels["service-foo"])
	assert.NotEqual(t, ch, &tchannel.Channel{})

	gauge := g.RootScope.Tagged(map[string]string{
		"client": "service-foo",
	}).Gauge("tchannel.client.running")
	assert.NotNil(t, gauge)
}

func TestGatewaySetupServerTChannelThrowsErrorWhenLoggerLevelIsIncorrect(t *testing.T) {
	cfg := NewStaticConfigOrDie(nil, map[string]interface{}{
		"tchannel.serviceName":    "test-gateway",
		"tchannel.processName":    "proc",
		"subLoggerLevel.tchannel": "non-compliant",
	})
	g := Gateway{
		Config: cfg,
	}
	err := g.setupServerTChannel(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown sub logger level for tchannel server: non-compliant")
}

func TestGatewaySetupServerTChannelCanSpecifyServerIPThroughConfig(t *testing.T) {
	cfg := NewStaticConfigOrDie(nil, map[string]interface{}{
		"tchannel.serviceName":    "test-g",
		"tchannel.processName":    "proc",
		"subLoggerLevel.tchannel": "error",
		"env":                     "test-bootstrap",
		"tchannel.server.ip":      "127.0.0.1",
	})

	s := &HTTPServer{Server: &http.Server{Addr: "127.0.0.1:0"}}
	g := Gateway{
		Config:          cfg,
		RootScope:       tally.NoopScope,
		localHTTPServer: s,
		httpServer:      s,
		WaitGroup:       &sync.WaitGroup{},
		Logger:          zap.NewNop(),
		logEncoder:      zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
	}

	err := g.setupServerTChannel(cfg)
	assert.NoError(t, err)

	err = g.Bootstrap()
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1", strings.Split(g.RealTChannelAddr, ":")[0])
}

func TestGatewaySetupServerTChannelWithShutdown(t *testing.T) {
	cfg := NewStaticConfigOrDie(nil, map[string]interface{}{
		"tchannel.serviceName":    "test-gateway",
		"tchannel.processName":    "proc",
		"subLoggerLevel.tchannel": "error",
	})
	g := Gateway{
		Config:     cfg,
		Logger:     zap.NewNop(),
		logEncoder: zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
	}
	err := g.setupServerTChannel(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, g.ServerTChannel)
	assert.NotNil(t, g.ClientTChannels)
	assert.False(t, g.tchannelServer.Closed())
	assert.NotNil(t, g.ServerTChannelRouter)
	assert.Equal(t, g.tchannelServer, g.ServerTChannel)

	// now shut down the tchannel server
	err = g.shutdownTChannelServerAndClients(context.Background())
	assert.NoError(t, err)
	assert.True(t, g.tchannelServer.Closed())
}

func TestGatewayWithTracerOverride(t *testing.T) {
	tracer := opentracing.NoopTracer{}
	tracerCloser := io.NopCloser(nil)
	rawCfgMap := map[string]interface{}{
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
		"jaeger.disabled":                    false,
		"jaeger.reporter.flush.milliseconds": 10000,
		"jaeger.reporter.hostport":           "localhost:6831",
		"jaeger.sampler.param":               0,
		"jaeger.sampler.type":                "const",
		"logger.fileName":                    "",
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
		"grpc.clientServiceNameMapping":      map[string]string{"test": "test"},
		"serviceName":                        "not-overridden",
		"metrics.serviceName":                "not-overridden",
		"serviceNameEnv":                     "TEST",
		"metrics.serviceNameEnv":             "TEST",
		"http.notFoundHandler.custom":        true,
		"http.handleMethodNotAllowed":        true,
	}

	var metricsBackend tally.CachedStatsReporter
	opts := &Options{
		GetContextScopeExtractors: nil,
		GetContextFieldExtractors: nil,
		JSONWrapper:               jsonwrapper.NewDefaultJSONWrapper(),
		MetricsBackend:            metricsBackend,
		NotFoundHandler: func(gateway *Gateway) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {}
		},
		TracerProvider: func(gateway *Gateway) (opentracing.Tracer, io.Closer, error) {
			return tracer, tracerCloser, nil
		},
	}

	t.Run("without config", func(t *testing.T) {
		cfg := NewStaticConfigOrDie(nil, rawCfgMap)
		g, err := CreateGateway(cfg, opts)
		assert.Nil(t, err)
		assert.NotEqual(t, tracer, g.Tracer)
		assert.NotEqual(t, tracerCloser, g.tracerCloser)
		_, ok := g.Tracer.(*jaeger.Tracer)
		assert.True(t, ok)
	})

	t.Run("with config and value being false", func(t *testing.T) {
		rawCfgMap["jaeger.tracer.custom"] = false
		cfg := NewStaticConfigOrDie(nil, rawCfgMap)
		g, err := CreateGateway(cfg, opts)
		assert.Nil(t, err)
		assert.NotEqual(t, tracer, g.Tracer)
		assert.NotEqual(t, tracerCloser, g.tracerCloser)
		_, ok := g.Tracer.(*jaeger.Tracer)
		assert.True(t, ok)
	})

	t.Run("with config", func(t *testing.T) {
		rawCfgMap["jaeger.tracer.custom"] = true
		cfg := NewStaticConfigOrDie(nil, rawCfgMap)
		g, err := CreateGateway(cfg, opts)
		assert.Nil(t, err)
		assert.Equal(t, tracer, g.Tracer)
		assert.Equal(t, tracerCloser, g.tracerCloser)
	})
}

func TestGatewayWithEventHandler(t *testing.T) {

	rawCfgMap := map[string]interface{}{}
	var metricsBackend tally.CachedStatsReporter

	opts := &Options{
		GetContextScopeExtractors: nil,
		GetContextFieldExtractors: nil,
		JSONWrapper:               jsonwrapper.NewDefaultJSONWrapper(),
		MetricsBackend:            metricsBackend,
		NotFoundHandler: func(gateway *Gateway) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {}
		},
	}

	t.Run("without event handler", func(t *testing.T) {
		cfg := NewStaticConfigOrDie(nil, rawCfgMap)
		g, err := CreateGateway(cfg, opts)
		assert.Nil(t, err)
		assert.NotEqual(t, NoOpEventHandler, g.EventHandler)
	})

	t.Run("with event handler", func(t *testing.T) {
		eventHandlerFn := func(_ []Event) error {
			return nil
		}

		enableEventGenFn := func(_, _ string) bool {
			return false
		}
		opts.EventProvider = func(gateway *Gateway) (EnableEventGenFn, EventHandlerFn) {
			return enableEventGenFn, eventHandlerFn
		}
		cfg := NewStaticConfigOrDie(nil, rawCfgMap)
		g, err := CreateGateway(cfg, opts)
		assert.Nil(t, err)
		assert.Equal(t, eventHandlerFn, g.EventHandler)
		assert.Equal(t, enableEventGenFn, g.EnableEventGen)

	})
}
