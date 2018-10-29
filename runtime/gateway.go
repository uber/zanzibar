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

package zanzibar

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/m3"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	jaegerLibTally "github.com/uber/jaeger-lib/metrics/tally"
	"github.com/uber/tchannel-go"
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

// Options configures the gateway
type Options struct {
	MetricsBackend            tally.CachedStatsReporter
	LogWriter                 zapcore.WriteSyncer
	GetContextScopeExtractors func() []ContextScopeTagsExtractor
}

// Gateway type
type Gateway struct {
	HTTPPort         int32
	TChannelPort     int32
	RealHTTPPort     int32
	RealHTTPAddr     string
	RealTChannelPort int32
	RealTChannelAddr string
	WaitGroup        *sync.WaitGroup
	Channel          *tchannel.Channel
	ContextLogger    ContextLogger
	ContextMetrics   ContextMetrics
	ContextExtractor ContextExtractor
	RootScope        tally.Scope
	ServiceName      string
	Config           *StaticConfig
	HTTPRouter       *HTTPRouter
	TChannelRouter   *TChannelRouter
	Tracer           opentracing.Tracer

	atomLevel       *zap.AtomicLevel
	loggerFile      *os.File
	scopeCloser     io.Closer
	metricsBackend  tally.CachedStatsReporter
	runtimeMetrics  RuntimeMetricsCollector
	logEncoder      zapcore.Encoder
	logWriter       zapcore.WriteSyncer
	logWriteSyncer  zapcore.WriteSyncer
	httpServer      *HTTPServer
	localHTTPServer *HTTPServer
	tchannelServer  *tchannel.Channel
	tracerCloser    io.Closer
	//	- panic ???
	//	- process reporter ?

	// Soon to be deprecated
	Logger *zap.Logger
}

// DefaultDependencies are the common dependencies for all modules
type DefaultDependencies struct {
	// Logger is a server-scoped logger
	// Deprecated: Use ContextLogger instead.
	Logger *zap.Logger
	// ContextExtractor extracts context for scope and logs field
	ContextExtractor ContextExtractor
	// ContextLogger is a logger with request-scoped log fields
	ContextLogger ContextLogger
	// ContextMetrics emit metrics from context
	ContextMetrics ContextMetrics

	Scope   tally.Scope
	Tracer  opentracing.Tracer
	Config  *StaticConfig
	Channel *tchannel.Channel
}

// CreateGateway func
func CreateGateway(
	config *StaticConfig, opts *Options,
) (*Gateway, error) {
	var metricsBackend tally.CachedStatsReporter
	var logWriter zapcore.WriteSyncer
	var scopeExtractors []ContextScopeTagsExtractor
	if opts == nil {
		opts = &Options{}
	}
	if opts.MetricsBackend != nil {
		metricsBackend = opts.MetricsBackend
	}
	if opts.LogWriter != nil {
		logWriter = opts.LogWriter
	}

	contextExtractors := &ContextExtractors{}
	if opts.GetContextScopeExtractors != nil {
		scopeExtractors = opts.GetContextScopeExtractors()

		for _, scopeExtractor := range scopeExtractors {
			contextExtractors.AddContextScopeTagsExtractor(scopeExtractor)
		}
	}

	gateway := &Gateway{
		HTTPPort:         int32(config.MustGetInt("http.port")),
		TChannelPort:     int32(config.MustGetInt("tchannel.port")),
		ServiceName:      config.MustGetString("serviceName"),
		WaitGroup:        &sync.WaitGroup{},
		Config:           config,
		ContextExtractor: contextExtractors.MakeContextExtractor(),
		logWriter:        logWriter,
		metricsBackend:   metricsBackend,
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

	if err := gateway.setupTChannel(config); err != nil {
		return nil, err
	}

	err := gateway.registerPredefined()

	return gateway, err
}

// Bootstrap func
func (gateway *Gateway) Bootstrap() error {
	// start HTTP server
	gateway.RootScope.Counter("server.bootstrap").Inc(1)
	_, err := gateway.localHTTPServer.JustListen()
	if err != nil {
		gateway.Logger.Error("Error listening on port", zap.Error(err))
		return errors.Wrap(err, "error listening on port")
	}
	if gateway.localHTTPServer.RealIP != gateway.httpServer.RealIP {
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
	tchannelIP, err := tchannel.ListenIP()
	if err != nil {
		return errors.Wrap(err, "error finding the best IP for tchannel")
	}
	tchannelAddr := tchannelIP.String() + ":" + strconv.Itoa(int(gateway.TChannelPort))
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
	return nil
}

func (gateway *Gateway) registerPredefined() error {
	gateway.HTTPRouter.RegisterRaw("GET", "/debug/pprof", pprof.Index)
	gateway.HTTPRouter.RegisterRaw("GET", "/debug/pprof/cmdline", pprof.Cmdline)
	gateway.HTTPRouter.RegisterRaw("GET", "/debug/pprof/profile", pprof.Profile)
	gateway.HTTPRouter.RegisterRaw("GET", "/debug/pprof/symbol", pprof.Symbol)
	gateway.HTTPRouter.RegisterRaw("POST", "/debug/pprof/symbol", pprof.Symbol)
	gateway.HTTPRouter.RegisterRaw(
		"GET", "/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP,
	)
	gateway.HTTPRouter.RegisterRaw(
		"GET", "/debug/pprof/heap", pprof.Handler("heap").ServeHTTP,
	)
	gateway.HTTPRouter.RegisterRaw(
		"GET", "/debug/pprof/threadcreate",
		pprof.Handler("threadcreate").ServeHTTP,
	)
	gateway.HTTPRouter.RegisterRaw(
		"GET", "/debug/pprof/block", pprof.Handler("block").ServeHTTP,
	)
	gateway.HTTPRouter.RegisterRaw("GET", "/debug/loglevel", gateway.atomLevel.ServeHTTP)
	gateway.HTTPRouter.RegisterRaw("PUT", "/debug/loglevel", gateway.atomLevel.ServeHTTP)

	err := gateway.HTTPRouter.Register("GET", "/health", NewRouterEndpointContext(
		gateway.ContextExtractor, gateway.ContextMetrics, gateway.Logger, gateway.Tracer,
		"health", "health",
		gateway.handleHealthRequest,
	))
	return err
}

func (gateway *Gateway) handleHealthRequest(
	ctx context.Context,
	req *ServerHTTPRequest,
	res *ServerHTTPResponse,
) {
	message := "Healthy, from " + gateway.ServiceName
	bytes := []byte(
		"{\"ok\":true,\"message\":\"" + message + "\"}\n",
	)

	res.WriteJSONBytes(200, nil, bytes)
}

// Shutdown starts the graceful shutdown, blocks until it is complete
func (gateway *Gateway) Shutdown() {
	var swg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), gateway.ShutdownTimeout())
	defer cancel()

	ec := make(chan error, 3)

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
		if err := gateway.shutdownTChannelServer(ctx); err != nil {
			ec <- errors.Wrap(err, "error shutting down tchannel server")
		}
	}()

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
	service := config.MustGetString("metrics.serviceName")
	env := config.MustGetString("env")

	if metricsType == "m3" {
		if gateway.metricsBackend != nil {
			panic("expected no metrics backend in gateway.")
		}

		// TODO: Why aren't common tags emitted?
		// NewReporter adds 'env' and 'service' common tags; and no 'host' tag.
		commonTags := map[string]string{}
		opts := m3.Options{
			HostPorts:          []string{config.MustGetString("metrics.m3.hostPort")},
			Service:            service,
			Env:                env,
			CommonTags:         commonTags,
			IncludeHost:        false,
			MaxQueueSize:       int(config.MustGetInt("metrics.m3.maxQueueSize")),
			MaxPacketSizeBytes: int32(config.MustGetInt("metrics.m3.maxPacketSizeBytes")),
		}
		if gateway.metricsBackend, err = m3.NewReporter(opts); err != nil {
			return err
		}
	} else if gateway.metricsBackend == nil {
		panic("expected gateway to have MetricsBackend in opts")
	}

	// TODO: Remove 'env' and 'service' default tags once they are emitted by metrics backend.
	defaultTags := map[string]string{
		"env":     env,
		"service": service,
		"host":    GetHostname(),
		"dc":      gateway.Config.MustGetString("datacenter"),
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
	gateway.runtimeMetrics = StartRuntimeMetricsCollector(
		runtimeMetricsOpts,
		gateway.RootScope,
	)

	return nil
}

func (gateway *Gateway) setupLogger(config *StaticConfig) error {
	var output zapcore.WriteSyncer
	logEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	logLevel := zap.InfoLevel
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
	prodCore := zapcore.NewCore(
		logEncoder,
		output,
		atomLevel,
	)
	zapLogger := zap.New(
		NewInstrumentedZapCore(
			prodCore, gateway.RootScope,
		),
	)

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
	return nil
}

// SubLogger returns a sub logger clone with given name and log level.
func (gateway *Gateway) SubLogger(name string, level zapcore.Level) *zap.Logger {
	newCore := NewInstrumentedZapCore(
		zapcore.NewCore(
			gateway.logEncoder.Clone(),
			gateway.logWriteSyncer,
			level,
		),
		gateway.RootScope,
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
		Disabled: config.MustGetBoolean("jaeger.disabled"),
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

	serviceName := config.MustGetString("serviceName")
	tracer, closer, err := jc.New(serviceName, opts...)
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

func (gateway *Gateway) setupTChannel(config *StaticConfig) error {
	serviceName := config.MustGetString("tchannel.serviceName")
	processName := config.MustGetString("tchannel.processName")
	levelString := gateway.Config.MustGetString("subLoggerLevel.tchannel")
	level, ok := levelMap[levelString]
	if !ok {
		return errors.Errorf("unknown sub logger level for tchannel server: %s", levelString)
	}

	channel, err := tchannel.NewChannel(
		serviceName,
		&tchannel.ChannelOptions{
			ProcessName: processName,
			Tracer:      gateway.Tracer,
			Logger:      NewTChannelLogger(gateway.SubLogger("tchannel", level)),
			StatsReporter: NewTChannelStatsReporter(
				gateway.RootScope,
			),

			//DefaultConnectionOptions: opts.DefaultConnectionOptions,
			//OnPeerStatusChanged:      opts.OnPeerStatusChanged,
			//RelayHost:                opts.RelayHost,
			//RelayLocalHandlers:       opts.RelayLocalHandlers,
			//RelayMaxTimeout:          opts.RelayMaxTimeout,
		})

	if err != nil {
		return errors.Errorf(
			"Error creating top channel:\n    %s",
			err)
	}

	gateway.Channel = channel
	gateway.tchannelServer = channel
	gateway.TChannelRouter = NewTChannelRouter(channel, gateway)

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

// shutdownTChannelServer gracefully shuts down the tchannel server, blocks until the shutdown is
// complete or the timeout has reached if there is one associated with the given context
func (gateway *Gateway) shutdownTChannelServer(ctx context.Context) error {
	shutdownPollInterval := defaultShutdownPollInterval
	if gateway.Config.ContainsKey("shutdown.pollInterval") {
		shutdownPollInterval = time.Duration(gateway.Config.MustGetInt("shutdown.pollInterval")) * time.Millisecond
	}
	ticker := time.NewTicker(shutdownPollInterval)
	defer ticker.Stop()

	gateway.tchannelServer.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if gateway.tchannelServer.Closed() {
				return nil
			}
		}
	}

}
