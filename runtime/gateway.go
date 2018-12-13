// Copyright (c) 2019 Uber Technologies, Inc.
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
	"go.uber.org/config"
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
	GetContextFieldExtractors func() []ContextLogFieldsExtractor
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
	Logger           *zap.Logger
	ServiceName      string
	Config           config.Provider
	HTTPRouter       HTTPRouter
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
	//	- process reporter ?
}

// DefaultDependencies are the common dependencies for all modules
type DefaultDependencies struct {
	// ContextExtractor extracts context for scope and logs field
	ContextExtractor ContextExtractor
	// ContextLogger is a logger with request-scoped log fields
	ContextLogger ContextLogger
	// ContextMetrics emit metrics from context
	ContextMetrics ContextMetrics

	Logger  *zap.Logger
	Scope   tally.Scope
	Tracer  opentracing.Tracer
	Config  config.Provider
	Channel *tchannel.Channel
}

type subloggerConfig struct{
	Jaeger string `yaml:"jaeger"`
	HTTP string `yaml:"http"`
	TChannel string `yaml:"tchannel"`
}

// CreateGateway func
func CreateGateway(
	config config.Provider, opts *Options,
) (*Gateway, error) {
	var metricsBackend tally.CachedStatsReporter
	var logWriter zapcore.WriteSyncer
	var scopeExtractors []ContextScopeTagsExtractor
	var fieldExtractors []ContextLogFieldsExtractor
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

	if opts.GetContextFieldExtractors != nil {
		fieldExtractors = opts.GetContextFieldExtractors()

		for _, fieldExtractor := range fieldExtractors {
			contextExtractors.AddContextLogFieldsExtractor(fieldExtractor)
		}
	}

	var (
		httpPort, tchannelPort int32
		serviceName string
	)

	config.Get("http.port").Populate(&httpPort)
	config.Get("tchannel.port").Populate(&tchannelPort)
	config.Get("serviceName").Populate(&serviceName)

	gateway := &Gateway{
		HTTPPort:         httpPort,
		TChannelPort:     tchannelPort,
		ServiceName:      serviceName,
		WaitGroup:        &sync.WaitGroup{},
		Config:           config,
		ContextExtractor: contextExtractors.MakeContextExtractor(),
		logWriter:        logWriter,
		metricsBackend:   metricsBackend,
	}

	// order matters for following setup method calls
	if err := gateway.setupMetrics(); err != nil {
		return nil, err
	}

	if err := gateway.setupLogger(); err != nil {
		return nil, err
	}

	var subloggerCfg *subloggerConfig
	config.Get("subLoggerLevel").Populate(&subloggerCfg)

	if err := gateway.setupTracer(subloggerCfg); err != nil {
		return nil, err
	}

	// setup router after metrics and logs
	gateway.HTTPRouter = NewHTTPRouter(gateway)

	if err := gateway.setupHTTPServer(subloggerCfg); err != nil {
		return nil, err
	}

	if err := gateway.setupTChannel(subloggerCfg); err != nil {
		return nil, err
	}

	gateway.registerPredefined()

	return gateway, nil
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

func (gateway *Gateway) registerPredefined() {
	gateway.HTTPRouter.Handle("GET", "/debug/pprof", http.HandlerFunc(pprof.Index))
	gateway.HTTPRouter.Handle("GET", "/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	gateway.HTTPRouter.Handle("GET", "/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	gateway.HTTPRouter.Handle("GET", "/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	gateway.HTTPRouter.Handle("POST", "/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	gateway.HTTPRouter.Handle(
		"GET", "/debug/pprof/goroutine", pprof.Handler("goroutine"),
	)
	gateway.HTTPRouter.Handle(
		"GET", "/debug/pprof/heap", pprof.Handler("heap"),
	)
	gateway.HTTPRouter.Handle(
		"GET", "/debug/pprof/threadcreate", pprof.Handler("threadcreate"),
	)
	gateway.HTTPRouter.Handle(
		"GET", "/debug/pprof/block", pprof.Handler("block"),
	)
	gateway.HTTPRouter.Handle("GET", "/debug/loglevel", gateway.atomLevel)
	gateway.HTTPRouter.Handle("PUT", "/debug/loglevel", gateway.atomLevel)

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
	gateway.HTTPRouter.Handle("GET", "/health", http.HandlerFunc(tracer.HandleRequest))
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
	var shutdownTimeout time.Duration

	gateway.Config.Get("shutdown.timeout").Populate(&shutdownTimeout)

	if shutdownTimeout == time.Duration(0) {
		return defaultCloseTimeout
	}

	return shutdownTimeout
}

// Wait for gateway to close the server
func (gateway *Gateway) Wait() {
	gateway.WaitGroup.Wait()
}

type metricsConfig struct {
	Type string `yaml:"type"`
	ServiceName string `yaml:"serviceName"`
	M3 m3metricsConfig `yaml:"m3"`
	Runtime metricsRuntimeConfig `yaml:"runtime"`
	FlushInterval time.Duration `yaml:"flushDuration"`
}

type m3metricsConfig struct {
	HostPort string `yaml:"hostPort"`
	MaxQueueSize int `yaml:"maxQueueSize"`
	MaxPacketSizeBytes int32 `yaml:"maxPacketSizeBytes"`
}

type metricsRuntimeConfig struct {
	CollectInterval time.Duration `yaml:"collectInterval"`
	EnableCPUMetrics bool `yaml:"enableCPUMetrics"`
	EnableMemMetrics bool `yaml:"enableMemMetrics"`
	EnableGCMetrics bool `yaml:"enableGCMetrics"`
}

func (gateway *Gateway) setupMetrics() (err error) {
	metricsCfg := &metricsConfig{}
	gateway.Config.Get("metrics").Populate(&metricsCfg)
	var env, dc string
	gateway.Config.Get("env").Populate(&env)
	gateway.Config.Get("datacenter").Populate(&dc)

	if metricsCfg.Type == "m3" {
		if gateway.metricsBackend != nil {
			panic("expected no metrics backend in gateway.")
		}

		// TODO: Why aren't common tags emitted?
		// NewReporter adds 'env' and 'service' common tags; and no 'host' tag.
		commonTags := map[string]string{}
		opts := m3.Options{
			HostPorts:          []string{metricsCfg.M3.HostPort},
			Service:            metricsCfg.ServiceName,
			Env:                env,
			CommonTags:         commonTags,
			IncludeHost:        false,
			MaxQueueSize:       metricsCfg.M3.MaxQueueSize,
			MaxPacketSizeBytes: metricsCfg.M3.MaxPacketSizeBytes,
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
		"service": metricsCfg.ServiceName,
		"host":    GetHostname(),
		"dc":      dc,
	}

	// Adds in any env variable variables specified in config
	var envVarsToTagInRootScope []string
	gateway.Config.Get("envVarsToTagInRootScope").Populate(&envVarsToTagInRootScope)
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
		metricsCfg.FlushInterval,
	)

	gateway.ContextMetrics = NewContextMetrics(gateway.RootScope)
	// start collecting runtime metrics
	collectInterval := metricsCfg.Runtime.CollectInterval
	runtimeMetricsOpts := RuntimeMetricsOptions{
		EnableCPUMetrics: metricsCfg.Runtime.EnableCPUMetrics,
		EnableMemMetrics: metricsCfg.Runtime.EnableMemMetrics,
		EnableGCMetrics:  metricsCfg.Runtime.EnableGCMetrics,
		CollectInterval:  collectInterval,
	}
	gateway.runtimeMetrics = StartRuntimeMetricsCollector(
		runtimeMetricsOpts,
		gateway.RootScope,
	)

	return nil
}

type loggerConfig struct {
	FileName string `yaml:"fileName"`
	Output string `yaml:"output"`
}

func (gateway *Gateway) setupLogger() error {
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

	var loggerConfig loggerConfig
	gateway.Config.Get("logger").Populate(&loggerConfig)
	loggerFileName := loggerConfig.FileName
	loggerOutput := loggerConfig.Output

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

	var env, dc, serviceName string
	gateway.Config.Get("env").Populate(&env)
	gateway.Config.Get("datacenter").Populate(&dc)
	gateway.Config.Get("serviceName").Populate(&serviceName)

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
		zap.String("zone", dc),
		zap.String("env", env),
		zap.String("hostname", GetHostname()),
		zap.String("service", serviceName),
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

type _jaegerConfig struct {
	Disabled bool `yaml:"disabled"`
	Reporter struct {
		HostPort string `yaml:"hostport"`
		Flush struct {
			Milliseconds time.Duration `yaml:"milliseconds"`
		} `yaml:"flush"`
	} `yaml:"reporter"`
	Sampler struct {
		Type string `yaml:"type"`
		Param float64 `yaml:"param"`
	} `yaml:"sampler"`
}

func (gateway *Gateway) initJaegerConfig() *jaegerConfig.Configuration {
	var serviceName string

	gateway.Config.Get("serviceName").Populate(&serviceName)

	var cfg _jaegerConfig
	gateway.Config.Get("jaeger").Populate(&cfg)

	return &jaegerConfig.Configuration{
		ServiceName: serviceName,
		Disabled:    cfg.Disabled,
		Reporter: &jaegerConfig.ReporterConfig{
			LocalAgentHostPort:  cfg.Reporter.HostPort,
			BufferFlushInterval: cfg.Reporter.Flush.Milliseconds,
		},
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  cfg.Sampler.Type,
			Param: cfg.Sampler.Param,
		},
	}
}

func (gateway *Gateway) setupTracer(subconfig *subloggerConfig) error {
	levelString := subconfig.Jaeger
	level, ok := levelMap[levelString]
	if !ok {
		return errors.Errorf("unknown sub logger level for jaeger tracer: %s", levelString)
	}

	opts := []jaegerConfig.Option{
		// TChannel logger implements jaeger logger interface
		jaegerConfig.Logger(NewTChannelLogger(gateway.SubLogger("jaeger", level))),
		jaegerConfig.Metrics(jaegerLibTally.Wrap(gateway.RootScope)),
	}
	jc := gateway.initJaegerConfig()

	tracer, closer, err := jc.NewTracer(opts...)
	if err != nil {
		return errors.Wrapf(err, "error initializing Jaeger tracer client")
	}
	opentracing.SetGlobalTracer(tracer)
	gateway.Tracer = tracer
	gateway.tracerCloser = closer
	return nil
}

func (gateway *Gateway) setupHTTPServer(subconfig *subloggerConfig) error {
	levelString := subconfig.HTTP
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

func (gateway *Gateway) setupTChannel(subloggerCfg *subloggerConfig) error {
	var tchannelCfg *TChannelConfig

	err := gateway.Config.Get("tchannel").Populate(&tchannelCfg)
	if err != nil {
		return err
	}

	serviceName := tchannelCfg.ServiceName
	processName := tchannelCfg.ProcessName
	levelString := subloggerCfg.TChannel
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
	var shutdownPollInterval time.Duration
	gateway.Config.Get("shutdown.pollInterval").Populate(&shutdownPollInterval)

	if shutdownPollInterval == time.Duration(0) {
		shutdownPollInterval = defaultShutdownPollInterval
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
