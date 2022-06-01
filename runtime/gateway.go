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

package zanzibar

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	metricCollector "github.com/afex/hystrix-go/hystrix/metric_collector"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/m3"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	jaegerLibTally "github.com/uber/jaeger-lib/metrics/tally"
	"github.com/uber/tchannel-go"
	"github.com/uber/zanzibar/runtime/jsonwrapper"
	"github.com/uber/zanzibar/runtime/plugins"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var levelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

var defaultShutdownPollInterval = 500 * time.Millisecond
var defaultCloseTimeout = 10000 * time.Millisecond

const (
	localhost                = "127.0.0.1"
	testenv                  = "test"
	metricsServiceFromEnvKey = "metrics.serviceNameEnv"
	serviceFromEnvKey        = "serviceNameEnv"
	skipZanzibarLogsKey      = "contextlogger.zanzibar.skiplogs"
)

// Options configures the gateway
type Options struct {
	MetricsBackend            tally.CachedStatsReporter
	MetricsDefaultBuckets     tally.Buckets // To support non-default histogram metric bucketing specific to a gateway
	LogWriter                 zapcore.WriteSyncer
	GetContextScopeExtractors func() []ContextScopeTagsExtractor
	GetContextFieldExtractors func() []ContextLogFieldsExtractor
	JSONWrapper               jsonwrapper.JSONWrapper

	// If present, request uuid is retrieved from the incoming request
	// headers using the key, and put on the context. Otherwise, a new
	// uuid is created for the incoming request.
	RequestUUIDHeaderKey string
}

// Gateway type
type Gateway struct {
	HTTPPort               int32
	TChannelPort           int32
	RealHTTPPort           int32
	RealHTTPAddr           string
	RealTChannelPort       int32
	RealTChannelAddr       string
	WaitGroup              *sync.WaitGroup
	ServerTChannel         *tchannel.Channel
	ClientTChannels        map[string]*tchannel.Channel
	ContextLogger          ContextLogger
	ContextMetrics         ContextMetrics
	ContextExtractor       ContextExtractor
	RootScope              tally.Scope
	Logger                 *zap.Logger
	ServiceName            string
	Config                 *StaticConfig
	HTTPRouter             HTTPRouter
	ServerTChannelRouter   *TChannelRouter
	TChannelSubLoggerLevel zapcore.Level
	Tracer                 opentracing.Tracer
	JSONWrapper            jsonwrapper.JSONWrapper

	// gRPC client dispatcher for gRPC client lifecycle management
	GRPCClientDispatcher *yarpc.Dispatcher

	atomLevel             *zap.AtomicLevel
	loggerFile            *os.File
	scopeCloser           io.Closer
	metricsBackend        tally.CachedStatsReporter
	runtimeMetrics        RuntimeMetricsCollector
	metricsDefaultBuckets tally.Buckets
	logEncoder            zapcore.Encoder
	logWriter             zapcore.WriteSyncer
	logWriteSyncer        zapcore.WriteSyncer
	httpServer            *HTTPServer
	localHTTPServer       *HTTPServer
	tchannelServer        *tchannel.Channel
	tracerCloser          io.Closer

	requestUUIDHeaderKey string
	isUnhealthy          bool
}

// DefaultDependencies are the common dependencies for all modules
type DefaultDependencies struct {
	// ContextExtractor extracts context for scope and logs field
	ContextExtractor ContextExtractor
	// ContextLogger is a logger with request-scoped log fields
	ContextLogger ContextLogger
	// ContextMetrics emit metrics from context
	ContextMetrics ContextMetrics

	Logger         *zap.Logger
	Scope          tally.Scope
	Tracer         opentracing.Tracer
	Config         *StaticConfig
	ServerTChannel *tchannel.Channel
	Gateway        *Gateway

	// dispatcher for managing gRPC clients
	GRPCClientDispatcher *yarpc.Dispatcher

	// JSONWrapper provides json marshaling and unmarshaling functionality
	JSONWrapper jsonwrapper.JSONWrapper
}

// CreateGateway func
func CreateGateway(
	config *StaticConfig, opts *Options,
) (*Gateway, error) {
	var metricsBackend tally.CachedStatsReporter
	var metricsDefaultBuckets tally.Buckets
	var logWriter zapcore.WriteSyncer
	var scopeTagsExtractors []ContextScopeTagsExtractor
	var logFieldsExtractors []ContextLogFieldsExtractor
	var jsonWrapper jsonwrapper.JSONWrapper

	if opts == nil {
		opts = &Options{}
	}
	if opts.MetricsBackend != nil {
		metricsBackend = opts.MetricsBackend
	}
	if opts.MetricsDefaultBuckets != nil {
		metricsDefaultBuckets = opts.MetricsDefaultBuckets
	} else {
		metricsDefaultBuckets = tally.DefaultBuckets
	}
	if opts.LogWriter != nil {
		logWriter = opts.LogWriter
	}
	if opts.JSONWrapper != nil {
		jsonWrapper = opts.JSONWrapper
	} else {
		jsonWrapper = jsonwrapper.NewDefaultJSONWrapper()
	}

	if opts.GetContextScopeExtractors != nil {
		scopeTagsExtractors = opts.GetContextScopeExtractors()
	} else {
		scopeTagsExtractors = []ContextScopeTagsExtractor{GetEndpointRequestHeadersFromCtx}
	}

	if opts.GetContextFieldExtractors != nil {
		logFieldsExtractors = opts.GetContextFieldExtractors()
	} else {
		logFieldsExtractors = []ContextLogFieldsExtractor{
			func(ctx context.Context) []zap.Field {
				reqHeaders := GetEndpointRequestHeadersFromCtx(ctx)
				fields := make([]zap.Field, 0, len(reqHeaders))
				for k, v := range reqHeaders {
					fields = append(fields, zap.String(k, v))
				}
				return fields
			},
		}
	}

	extractors := &ContextExtractors{
		ScopeTagsExtractors: scopeTagsExtractors,
		LogFieldsExtractors: logFieldsExtractors,
	}

	var service string
	if config.ContainsKey(serviceFromEnvKey) {
		service = os.Getenv(config.MustGetString(serviceFromEnvKey))
	}
	if service == "" {
		service = config.MustGetString("serviceName")
	}

	gateway := &Gateway{
		HTTPPort:              int32(config.MustGetInt("http.port")),
		TChannelPort:          int32(config.MustGetInt("tchannel.port")),
		ServiceName:           service,
		WaitGroup:             &sync.WaitGroup{},
		Config:                config,
		ContextExtractor:      extractors,
		JSONWrapper:           jsonWrapper,
		logWriter:             logWriter,
		metricsBackend:        metricsBackend,
		metricsDefaultBuckets: metricsDefaultBuckets,
		requestUUIDHeaderKey:  opts.RequestUUIDHeaderKey,
	}

	gateway.setupConfig(config)
	config.Freeze()

	// order matters for following setup method calls
	if err := gateway.setupMetrics(config); err != nil {
		return nil, err
	}

	if err := gateway.setupLogger(config); err != nil {
		return nil, err
	}

	if err := gateway.setupTracer(config); err != nil {
		return nil, err
	}

	// setup router after metrics and logs
	gateway.HTTPRouter = NewHTTPRouter(gateway)

	if err := gateway.setupHTTPServer(); err != nil {
		return nil, err
	}

	if err := gateway.setupServerTChannel(config); err != nil {
		return nil, err
	}

	// setup gRPC client dispatcher after metrics, logger and tracer
	if err := gateway.setupGRPCClientDispatcher(config); err != nil {
		return nil, err
	}

	gateway.registerPredefined()

	return gateway, nil
}

// Bootstrap func
func (gateway *Gateway) Bootstrap() error {
	env := gateway.Config.MustGetString("env")

	// start HTTP server
	gateway.RootScope.Counter("server.bootstrap").Inc(1)
	_, err := gateway.localHTTPServer.JustListen()
	if err != nil {
		gateway.Logger.Error("Error listening on port", zap.Error(err))
		return errors.Wrap(err, "error listening on port")
	}
	if gateway.localHTTPServer.RealIP != gateway.httpServer.RealIP && env != testenv {
		_, err := gateway.httpServer.JustListen()
		if err != nil {
			gateway.Logger.Error("Error listening on port", zap.Error(err))
			return errors.Wrap(err, "error listening on port")
		}
	} else {
		// Do not start at the same IP
		gateway.httpServer = gateway.localHTTPServer
	}
	gateway.RealHTTPPort = gateway.httpServer.RealPort
	gateway.RealHTTPAddr = gateway.httpServer.RealAddr

	gateway.WaitGroup.Add(1)
	go gateway.httpServer.JustServe(gateway.WaitGroup)

	if gateway.httpServer != gateway.localHTTPServer {
		gateway.WaitGroup.Add(1)
		go gateway.localHTTPServer.JustServe(gateway.WaitGroup)
	}

	// start TChannel server
	ip := localhost
	if gateway.Config.ContainsKey("tchannel.server.ip") {
		ip = gateway.Config.MustGetString("tchannel.server.ip")
	} else if env != testenv {
		tchannelIP, err := tchannel.ListenIP()
		if err != nil {
			return errors.Wrap(err, "error finding the best IP for tchannel")
		}
		ip = tchannelIP.String()
	}
	tchannelAddr := ip + ":" + strconv.Itoa(int(gateway.TChannelPort))
	ln, err := net.Listen("tcp", tchannelAddr)
	if err != nil {
		gateway.Logger.Error("Error listening tchannel port", zap.Error(err))
		return err
	}
	gateway.RealTChannelAddr = ln.Addr().String()
	gateway.RealTChannelPort = int32(ln.Addr().(*net.TCPAddr).Port)

	// tchannel serve does not block, connection handling is done in different goroutine
	err = gateway.tchannelServer.Serve(ln)
	if err != nil {
		gateway.Logger.Error("Error starting tchannel server", zap.Error(err))
		return err
	}

	gateway.RootScope.Counter("startup.success").Inc(1)

	if gateway.GRPCClientDispatcher != nil {
		err = gateway.GRPCClientDispatcher.Start()
		if err != nil {
			gateway.Logger.Error("error starting gRPC client dispatcher", zap.Error(err))
			return err
		}
	}

	return nil
}

func (gateway *Gateway) registerPredefined() {
	deps := &DefaultDependencies{
		Scope:         gateway.RootScope,
		ContextLogger: gateway.ContextLogger,
		Logger:        gateway.Logger,
		Tracer:        gateway.Tracer,
	}

	tracer := NewRouterEndpoint(
		gateway.ContextExtractor, deps,
		"health", "health",
		gateway.handleHealthRequest,
	)
	_ = gateway.HTTPRouter.Handle("GET", "/health", http.HandlerFunc(tracer.HandleRequest))
}

func (gateway *Gateway) handleHealthRequest(
	ctx context.Context,
	req *ServerHTTPRequest,
	res *ServerHTTPResponse,
) context.Context {
	if gateway.isUnhealthy {
		message := "Unhealthy, from " + gateway.ServiceName
		bytes := []byte(
			"{\"ok\":false,\"message\":\"" + message + "\"}\n",
		)
		res.WriteJSONBytes(503, nil, bytes)
		return ctx
	}
	message := "Healthy, from " + gateway.ServiceName
	bytes := []byte(
		"{\"ok\":true,\"message\":\"" + message + "\"}\n",
	)

	res.WriteJSONBytes(200, nil, bytes)
	return ctx
}

// Shutdown starts the graceful shutdown, blocks until it is complete
func (gateway *Gateway) Shutdown() {
	// stop accepting incoming requests as soon as shutdown signal is received.
	gateway.isUnhealthy = true

	var swg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), gateway.ShutdownTimeout())
	defer cancel()

	ec := make(chan error, 4)

	if gateway.localHTTPServer != gateway.httpServer {
		swg.Add(1)
		go func() {
			defer swg.Done()
			if err := gateway.localHTTPServer.Shutdown(ctx); err != nil {
				ec <- errors.Wrap(err, "error shutting down local http server")
			}
		}()
	}

	// shutdown http server
	swg.Add(1)
	go func() {
		defer swg.Done()
		if err := gateway.httpServer.Shutdown(ctx); err != nil {
			ec <- errors.Wrap(err, "error shutting down http server")
		}
	}()

	// shutdown tchannel server
	swg.Add(1)
	go func() {
		defer swg.Done()
		if err := gateway.shutdownTChannelServerAndClients(ctx); err != nil {
			ec <- errors.Wrap(err, "error shutting down tchannel server or clients")
		}
	}()

	// wait for servers to shutdown before stopping GRPCClientDispatcher
	swg.Wait()

	// stop all grpc clients
	if gateway.GRPCClientDispatcher != nil {
		swg.Add(1)
		go func() {
			defer swg.Done()
			if err := gateway.GRPCClientDispatcher.Stop(); err != nil {
				ec <- errors.Wrap(err, "error stopping gRPC client dispatcher")
			}
		}()
	}

	swg.Wait()

	select {
	case err := <-ec:
		// close ec so that the range ec will not block forever
		close(ec)
		errs := make([]string, 0, cap(ec))
		errs = append(errs, err.Error())
		for e := range ec {
			errs = append(errs, e.Error())
		}
		gateway.Logger.Error(fmt.Sprintf(
			"%d errors when shutting down the servers: %s",
			len(errs), strings.Join(errs, ";")),
		)
		gateway.RootScope.Counter("shutdown.failure").Inc(1)
	default:
		gateway.Logger.Info("servers are shut down gracefully")
		gateway.RootScope.Counter("shutdown.success").Inc(1)
	}

	_ = gateway.tracerCloser.Close()
	gateway.metricsBackend.Flush()
	_ = gateway.scopeCloser.Close()

	// close log files as the last step
	if gateway.loggerFile != nil {
		_ = gateway.loggerFile.Sync()
		_ = gateway.loggerFile.Close()
	}

	// stop collecting runtime metrics
	if gateway.runtimeMetrics != nil {
		gateway.runtimeMetrics.Stop()
	}
}

// Close shuts down the servers and returns immediately
func (gateway *Gateway) Close() {
	if gateway.localHTTPServer != gateway.httpServer {
		gateway.localHTTPServer.Close()
	}
	gateway.httpServer.Close()
	gateway.tchannelServer.Close()

	_ = gateway.tracerCloser.Close()
	gateway.metricsBackend.Flush()
	_ = gateway.scopeCloser.Close()

	// close log files as the last step
	if gateway.loggerFile != nil {
		_ = gateway.loggerFile.Sync()
		_ = gateway.loggerFile.Close()
	}

	// stop collecting runtime metrics
	if gateway.runtimeMetrics != nil {
		gateway.runtimeMetrics.Stop()
	}
}

// ShutdownTimeout returns the shutdown configured timeout, which default to 10s.
func (gateway *Gateway) ShutdownTimeout() time.Duration {
	if gateway.Config.ContainsKey("shutdown.timeout") {
		return time.Duration(gateway.Config.MustGetInt("shutdown.timeout")) * time.Millisecond
	}
	return defaultCloseTimeout
}

// InspectOrDie inspects the config for this gateway
func (gateway *Gateway) InspectOrDie() map[string]interface{} {
	return gateway.Config.InspectOrDie()
}

// Wait for gateway to close the server
func (gateway *Gateway) Wait() {
	gateway.WaitGroup.Wait()
}

func (gateway *Gateway) setupConfig(config *StaticConfig) {
	useDC := config.MustGetBoolean("useDatacenter")

	if useDC {
		dcFile := config.MustGetString("datacenterFile")
		bytes, err := ioutil.ReadFile(dcFile)
		if err != nil {
			panic("expected datacenterFile: " + dcFile + " to exist")
		}

		config.SetSeedOrDie("datacenter", string(bytes))
	} else {
		config.SetConfigValueOrDie("datacenter", []byte("unknown"), "string")
	}
}

func (gateway *Gateway) setupMetrics(config *StaticConfig) (err error) {
	metricsType := config.MustGetString("metrics.type")

	var service string
	if config.ContainsKey(metricsServiceFromEnvKey) {
		service = os.Getenv(config.MustGetString(metricsServiceFromEnvKey))
	}
	if service == "" {
		service = config.MustGetString("metrics.serviceName")
	}

	env := config.MustGetString("env")
	if metricsType == "m3" {
		if gateway.metricsBackend != nil {
			panic("expected no metrics backend in gateway.")
		}

		opts := m3.Options{
			HostPorts:          []string{config.MustGetString("metrics.m3.hostPort")},
			Service:            service,
			Env:                env,
			MaxQueueSize:       int(config.MustGetInt("metrics.m3.maxQueueSize")),
			MaxPacketSizeBytes: int32(config.MustGetInt("metrics.m3.maxPacketSizeBytes")),
		}
		if gateway.metricsBackend, err = m3.NewReporter(opts); err != nil {
			return err
		}
	} else if gateway.metricsBackend == nil {
		panic("expected gateway to have MetricsBackend in opts")
	}

	defaultTags := map[string]string{
		"env":     env,
		"service": service,
		"dc":      gateway.Config.MustGetString("datacenter"),
	}
	if config.MustGetBoolean("metrics.m3.includeHost") {
		defaultTags["host"] = GetHostname()
	}

	// Adds in any env variable variables specified in config
	envVarsToTagInRootScope := []string{}
	config.MustGetStruct("envVarsToTagInRootScope", &envVarsToTagInRootScope)
	for _, envVarName := range envVarsToTagInRootScope {
		envVarValue := os.Getenv(envVarName)
		defaultTags[envVarName] = envVarValue
	}
	gateway.RootScope, gateway.scopeCloser = tally.NewRootScope(
		tally.ScopeOptions{
			Tags:            defaultTags,
			CachedReporter:  gateway.metricsBackend,
			DefaultBuckets:  gateway.metricsDefaultBuckets,
			Separator:       tally.DefaultSeparator,
			SanitizeOptions: &m3.DefaultSanitizerOpts,
		},
		time.Duration(config.MustGetInt("metrics.flushInterval"))*time.Millisecond,
	)

	gateway.ContextMetrics = NewContextMetrics(gateway.RootScope)
	// start collecting runtime metrics
	collectInterval := time.Duration(config.MustGetInt("metrics.runtime.collectInterval")) * time.Millisecond
	runtimeMetricsOpts := RuntimeMetricsOptions{
		EnableCPUMetrics: config.MustGetBoolean("metrics.runtime.enableCPUMetrics"),
		EnableMemMetrics: config.MustGetBoolean("metrics.runtime.enableMemMetrics"),
		EnableGCMetrics:  config.MustGetBoolean("metrics.runtime.enableGCMetrics"),
		CollectInterval:  collectInterval,
	}

	// runtime metrics should always have host tag
	runtimeScope := gateway.RootScope
	if _, ok := defaultTags["host"]; !ok {
		runtimeScope = gateway.RootScope.Tagged(map[string]string{
			"host": GetHostname(),
		})
	}
	gateway.runtimeMetrics = StartRuntimeMetricsCollector(
		runtimeMetricsOpts,
		runtimeScope,
	)

	//Initialize M3Collector for hystrix metrics
	c := plugins.InitializeM3Collector(gateway.RootScope)
	metricCollector.Registry.Register(c.NewM3Collector)

	return nil
}

func (gateway *Gateway) setupLogger(config *StaticConfig) error {
	var output zapcore.WriteSyncer
	logEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	logLevel := zap.InfoLevel
	if config.ContainsKey("logger.level") {
		levelString := config.MustGetString("logger.level")
		var ok bool
		logLevel, ok = levelMap[levelString]
		if !ok {
			return errors.Errorf("unknown log level for gateway logger: %s", levelString)
		}
	}

	tempLogger := zap.New(
		zapcore.NewCore(
			logEncoder,
			os.Stderr,
			logLevel,
		),
	)

	loggerFileName := config.MustGetString("logger.fileName")
	loggerOutput := config.MustGetString("logger.output")

	if loggerFileName == "" || loggerOutput == "stdout" {
		var writer zapcore.WriteSyncer
		if gateway.logWriter != nil {
			writer = zap.CombineWriteSyncers(os.Stdout, gateway.logWriter)
		} else {
			writer = os.Stdout
		}

		output = writer
	} else {
		err := os.MkdirAll(filepath.Dir(loggerFileName), 0777)
		if err != nil {
			tempLogger.Error("Error creating log directory", zap.Error(err))
			return errors.Wrap(err, "Error creating log directory")
		}

		loggerFile, err := os.OpenFile(
			loggerFileName,
			os.O_APPEND|os.O_WRONLY|os.O_CREATE,
			0644,
		)
		if err != nil {
			tempLogger.Error("Error opening log file", zap.Error(err))
			return errors.Wrap(err, "Error opening log file")
		}
		gateway.loggerFile = loggerFile
		if gateway.logWriter != nil {
			writer := zap.CombineWriteSyncers(loggerFile, gateway.logWriter)
			output = writer
		} else {
			output = loggerFile
		}
	}

	atomLevel := zap.NewAtomicLevelAt(logLevel)
	zapLogger := zap.New(zapcore.NewCore(
		logEncoder,
		output,
		atomLevel,
	))

	gateway.atomLevel = &atomLevel
	gateway.logEncoder = logEncoder
	gateway.logWriteSyncer = output

	// Default to a STDOUT logger
	gateway.Logger = zapLogger.With(
		zap.String("zone", gateway.Config.MustGetString("datacenter")),
		zap.String("env", gateway.Config.MustGetString("env")),
		zap.String("hostname", GetHostname()),
		zap.String("service", gateway.Config.MustGetString("serviceName")),
		zap.Int("pid", os.Getpid()),
	)

	gateway.ContextLogger = NewContextLogger(gateway.Logger)

	if config.ContainsKey(skipZanzibarLogsKey) {
		skipZanzibarLogs := config.MustGetBoolean(skipZanzibarLogsKey)
		gateway.ContextLogger.SetSkipZanzibarLogs(skipZanzibarLogs)
	}

	return nil
}

// SubLogger returns a sub logger clone with given name and log level.
func (gateway *Gateway) SubLogger(name string, level zapcore.Level) *zap.Logger {
	newCore := zapcore.NewCore(
		gateway.logEncoder.Clone(),
		gateway.logWriteSyncer,
		level,
	)
	return gateway.Logger.With(
		zap.String("subLogger", name),
	).WithOptions(
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return newCore
		}),
	)
}

func (gateway *Gateway) initJaegerConfig(config *StaticConfig) *jaegerConfig.Configuration {
	return &jaegerConfig.Configuration{
		ServiceName: config.MustGetString("serviceName"),
		Disabled:    config.MustGetBoolean("jaeger.disabled"),
		Reporter: &jaegerConfig.ReporterConfig{
			LocalAgentHostPort:  config.MustGetString("jaeger.reporter.hostport"),
			BufferFlushInterval: time.Duration(config.MustGetInt("jaeger.reporter.flush.milliseconds")) * time.Millisecond,
		},
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  config.MustGetString("jaeger.sampler.type"),
			Param: config.MustGetFloat("jaeger.sampler.param"),
		},
	}
}

func (gateway *Gateway) setupTracer(config *StaticConfig) error {
	levelString := gateway.Config.MustGetString("subLoggerLevel.jaeger")
	level, ok := levelMap[levelString]
	if !ok {
		return errors.Errorf("unknown sub logger level for jaeger tracer: %s", levelString)
	}

	opts := []jaegerConfig.Option{
		// TChannel logger implements jaeger logger interface
		jaegerConfig.Logger(NewTChannelLogger(gateway.SubLogger("jaeger", level))),
		jaegerConfig.Metrics(jaegerLibTally.Wrap(gateway.RootScope)),
	}
	jc := gateway.initJaegerConfig(config)

	tracer, closer, err := jc.NewTracer(opts...)
	if err != nil {
		return errors.Wrapf(err, "error initializing Jaeger tracer client")
	}
	opentracing.SetGlobalTracer(tracer)
	gateway.Tracer = tracer
	gateway.tracerCloser = closer
	return nil
}

func (gateway *Gateway) setupHTTPServer() error {
	levelString := gateway.Config.MustGetString("subLoggerLevel.http")
	level, ok := levelMap[levelString]
	if !ok {
		return errors.Errorf("unknown sub logger level for http server: %s", levelString)
	}
	httpLogger := gateway.SubLogger("http", level)

	listenIP, err := tchannel.ListenIP()
	if err != nil {
		return errors.Wrap(err, "error finding the best IP")
	}
	gateway.httpServer = &HTTPServer{
		Server: &http.Server{
			Addr:    listenIP.String() + ":" + strconv.FormatInt(int64(gateway.HTTPPort), 10),
			Handler: gateway.HTTPRouter,
		},
		Logger: httpLogger,
	}

	gateway.localHTTPServer = &HTTPServer{
		Server: &http.Server{
			Addr:    "127.0.0.1:" + strconv.FormatInt(int64(gateway.HTTPPort), 10),
			Handler: gateway.HTTPRouter,
		},
		Logger: httpLogger,
	}
	return nil
}

func (gateway *Gateway) setupServerTChannel(config *StaticConfig) error {
	serviceName := config.MustGetString("tchannel.serviceName")
	processName := config.MustGetString("tchannel.processName")
	levelString := gateway.Config.MustGetString("subLoggerLevel.tchannel")
	level, ok := levelMap[levelString]
	if !ok {
		return errors.Errorf("unknown sub logger level for tchannel server: %s", levelString)
	}
	gateway.TChannelSubLoggerLevel = level

	channel, err := tchannel.NewChannel(
		serviceName,
		&tchannel.ChannelOptions{
			ProcessName: processName,
			Tracer:      gateway.Tracer,
			Logger:      NewTChannelLogger(gateway.SubLogger("tchannel", level)),
			StatsReporter: NewTChannelStatsReporter(
				gateway.RootScope,
			),
		})

	if err != nil {
		return errors.Errorf("Error creating top channel:\n%s", err)
	}

	gateway.ServerTChannel = channel
	gateway.tchannelServer = gateway.ServerTChannel
	gateway.ServerTChannelRouter = NewTChannelRouter(channel, gateway)
	// client tchannels are created explicitly for each client if "dedicated.tchannel.client: true"
	gateway.ClientTChannels = make(map[string]*tchannel.Channel)
	return nil
}

// SetupClientTChannel sets up a dedicated tchannel for each client with a given service name
// If multiple backends with the same service name exist (for e.g. presentation service), then
// all of them would receive the same channel. The method is exported because it is called from
// the generated clients if "dedicated.tchannel.client: true" else server tchannel is reused
func (gateway *Gateway) SetupClientTChannel(config *StaticConfig, serviceName string) *tchannel.Channel {
	if ch, ok := gateway.ClientTChannels[serviceName]; ok {
		gateway.Logger.Info(fmt.Sprintf("returning already initialised TChannel client for [%v]", serviceName))
		return ch
	}
	processName := config.MustGetString("tchannel.processName")
	level := gateway.TChannelSubLoggerLevel

	channel, err := tchannel.NewChannel(
		// when specifying the service name for the channel, we reuse the server service
		// name else calls from other unauthorised sources may be blocked
		config.MustGetString("tchannel.serviceName"),
		&tchannel.ChannelOptions{
			ProcessName:   processName,
			Tracer:        gateway.Tracer,
			Logger:        NewTChannelLogger(gateway.SubLogger("tchannel", level)),
			StatsReporter: NewTChannelStatsReporter(gateway.RootScope),
		})

	scope := gateway.RootScope.Tagged(map[string]string{
		"client": serviceName,
	})
	if err != nil {
		scope.Gauge("tchannel.client.running").Update(0)
		gateway.Logger.Info(fmt.Sprintf("Failed to initiate dedicated TChannel client for [%v]", serviceName))
	} else {
		gateway.Logger.Info(fmt.Sprintf("Dedicated TChannel client initiated for client [%v]", serviceName))
		scope.Gauge("tchannel.client.running").Update(1)
	}
	gateway.ClientTChannels[serviceName] = channel
	return channel
}

func (gateway *Gateway) setupGRPCClientDispatcher(config *StaticConfig) error {
	ip := config.MustGetString("sidecarRouter.default.grpc.ip")
	port := config.MustGetInt("sidecarRouter.default.grpc.port")
	address := fmt.Sprintf("%s:%d", ip, port)

	clientServiceNameMapping := make(map[string]string)
	config.MustGetStruct("grpc.clientServiceNameMapping", &clientServiceNameMapping)
	if len(clientServiceNameMapping) == 0 {
		return nil
	}

	unaryOutbound := grpc.NewTransport(
		grpc.Logger(gateway.Logger),
		grpc.Tracer(gateway.Tracer),
	).NewSingleOutbound(address)
	outbounds := make(yarpc.Outbounds, len(clientServiceNameMapping))
	for key, value := range clientServiceNameMapping {
		outbounds[key] = transport.Outbounds{
			ServiceName: value,
			Unary:       unaryOutbound,
		}
	}

	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name:      config.MustGetString("serviceName"),
		Outbounds: outbounds,
		Logging: yarpc.LoggingConfig{
			Zap: gateway.Logger,
			// TODO: set proper extractors
		},
		Metrics: yarpc.MetricsConfig{
			// TODO: contextual scope
			Tally: gateway.RootScope,
		},
	})
	gateway.GRPCClientDispatcher = dispatcher
	return nil
}

// GetDirnameFromRuntimeCaller will compute the current dirname
// if passed a filename from runtime.Caller(0). This is useful
// for doing __dirname/__FILE__ for golang.
func GetDirnameFromRuntimeCaller(file string) string {
	dirname := filepath.Dir(file)

	// Strip _obj dirs generated by test -cover ...
	if filepath.Base(dirname) == "_obj" {
		dirname = filepath.Dir(dirname)
	}

	// Strip _obj_test in go test -cover
	if filepath.Base(dirname) == "_obj_test" {
		dirname = filepath.Dir(dirname)
	}

	// go test -cover does weird folder stuff
	if filepath.Base(dirname) == "_test" {
		dirname = filepath.Dir(dirname)
	}

	// if filepath then we are done, otherwise its go package name
	if filepath.IsAbs(dirname) {
		return dirname
	}

	// If dirname is not absolute then its a package name...
	return filepath.Join(os.Getenv("GOPATH"), "src", dirname)
}

// GetHostname returns hostname
func GetHostname() string {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}
	return host
}

// shutdownTChannelServerAndClients gracefully shuts down the tchannel server, blocks until the shutdown is
// complete or the timeout has reached if there is one associated with the given context
// It also shuts down all the dedicated client tchannel connections on a best effort basis
func (gateway *Gateway) shutdownTChannelServerAndClients(ctx context.Context) error {
	shutdownPollInterval := defaultShutdownPollInterval
	if gateway.Config.ContainsKey("shutdown.pollInterval") {
		shutdownPollInterval = time.Duration(gateway.Config.MustGetInt("shutdown.pollInterval")) * time.Millisecond
	}
	ticker := time.NewTicker(shutdownPollInterval)
	defer ticker.Stop()

	gateway.Logger.Info("Closing the TChannel server")
	gateway.tchannelServer.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if gateway.tchannelServer.Closed() {
				gateway.Logger.Info("TChannel server closed successfully")

				for serviceName, clientTchannel := range gateway.ClientTChannels {
					go func(service string, ch *tchannel.Channel) {
						gateway.Logger.Info(fmt.Sprintf("Closing TChannel client for [%v]", service))
						ch.Close()
						gateway.RootScope.Tagged(map[string]string{
							"client": service,
						}).Gauge("tchannel.client.running").Update(0)

					}(serviceName, clientTchannel)
				}
			} else {
				gateway.Logger.Info(fmt.Sprintf("Failed to close TChannel server within %v ms", shutdownPollInterval))
			}
			return nil
		}
	}
}
