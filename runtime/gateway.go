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
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/m3"
	"github.com/uber/tchannel-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/opentracing/opentracing-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
)

const defaultM3MaxQueueSize = 10000
const defaultM3MaxPacketSize = 1440 // 1440kb in UDP M3MaxPacketSize
const defaultM3FlushInterval = 500 * time.Millisecond

// Options configures the gateway
type Options struct {
	MetricsBackend tally.CachedStatsReporter
	LogWriter      zapcore.WriteSyncer
}

// Gateway type
type Gateway struct {
	HTTPPort              int32
	TChannelPort          int32
	RealHTTPPort          int32
	RealHTTPAddr          string
	RealTChannelPort      int32
	RealTChannelAddr      string
	WaitGroup             *sync.WaitGroup
	Channel               *tchannel.Channel
	Logger                *zap.Logger
	RootScope             tally.Scope
	MetricsScope          tally.Scope
	PerWorkerMetricsScope tally.Scope
	ServiceName           string
	Config                *StaticConfig
	HTTPRouter            *HTTPRouter
	TChannelRouter        *TChannelRouter
	Tracer                opentracing.Tracer

	loggerFile      *os.File
	scopeCloser     io.Closer
	metricsBackend  tally.CachedStatsReporter
	runtimeMetrics  RuntimeMetricsCollector
	logWriter       zapcore.WriteSyncer
	httpServer      *HTTPServer
	localHTTPServer *HTTPServer
	tchannelServer  *tchannel.Channel
	tracerCloser    io.Closer
	//	- panic ???
	//	- process reporter ?
}

// CreateGateway func
func CreateGateway(
	config *StaticConfig, opts *Options,
) (*Gateway, error) {
	var metricsBackend tally.CachedStatsReporter
	var logWriter zapcore.WriteSyncer
	if opts != nil && opts.MetricsBackend != nil {
		metricsBackend = opts.MetricsBackend
	}
	if opts != nil && opts.LogWriter != nil {
		logWriter = opts.LogWriter
	}

	gateway := &Gateway{
		HTTPPort:     int32(config.MustGetInt("http.port")),
		TChannelPort: int32(config.MustGetInt("tchannel.port")),
		ServiceName:  config.MustGetString("serviceName"),
		WaitGroup:    &sync.WaitGroup{},
		Config:       config,

		logWriter:      logWriter,
		metricsBackend: metricsBackend,
	}

	gateway.setupConfig(config)
	config.Freeze()

	gateway.HTTPRouter = NewHTTPRouter(gateway)

	// order matters for following setup method calls
	if err := gateway.setupLogger(config); err != nil {
		return nil, err
	}

	if err := gateway.setupMetrics(config); err != nil {
		return nil, err
	}

	if err := gateway.setupTracer(config); err != nil {
		return nil, err
	}

	if err := gateway.setupHTTPServer(); err != nil {
		return nil, err
	}

	if err := gateway.setupTChannel(config); err != nil {
		return nil, err
	}

	gateway.registerPredefined()

	return gateway, nil
}

// Bootstrap func
func (gateway *Gateway) Bootstrap() error {
	// start HTTP server
	_, err := gateway.localHTTPServer.JustListen()
	if err != nil {
		gateway.Logger.Error("Error listening on port",
			zap.String("error", err.Error()),
		)
		return errors.Wrap(err, "error listening on port")
	}
	if gateway.localHTTPServer.RealIP != gateway.httpServer.RealIP {
		_, err := gateway.httpServer.JustListen()
		if err != nil {
			gateway.Logger.Error("Error listening on port",
				zap.String("error", err.Error()),
			)
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
		gateway.Logger.Error(
			"Error listening tchannel port",
			zap.String("error", err.Error()),
		)
		return err
	}
	gateway.RealTChannelAddr = ln.Addr().String()
	gateway.RealTChannelPort = int32(ln.Addr().(*net.TCPAddr).Port)

	// tchannel serve does not block, connection handling is done in different goroutine
	err = gateway.tchannelServer.Serve(ln)
	if err != nil {
		gateway.Logger.Error(
			"Error starting tchannel server",
			zap.String("error", err.Error()),
		)
		return err
	}

	return nil
}

func (gateway *Gateway) registerPredefined() {
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

	gateway.HTTPRouter.Register("GET", "/health", NewRouterEndpoint(
		gateway, "health", "health", gateway.handleHealthRequest,
	))
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

// Close the http server
func (gateway *Gateway) Close() {
	_ = gateway.tracerCloser.Close()
	gateway.metricsBackend.Flush()
	_ = gateway.scopeCloser.Close()
	if gateway.localHTTPServer != gateway.httpServer {
		gateway.localHTTPServer.Close()
	}
	gateway.httpServer.Close()
	gateway.tchannelServer.Close()

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
		config.SetSeedOrDie("datacenter", "unknown")
	}
}

func (gateway *Gateway) setupMetrics(config *StaticConfig) (err error) {
	metricsType := config.MustGetString("metrics.type")
	service := config.MustGetString("metrics.tally.service")
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
			MaxQueueSize:       defaultM3MaxQueueSize,
			MaxPacketSizeBytes: defaultM3MaxPacketSize,
		}
		if gateway.metricsBackend, err = m3.NewReporter(opts); err != nil {
			return err
		}
	} else if gateway.metricsBackend == nil {
		panic("expected gateway to have MetricsBackend in opts")
	}

	// Adding 'env' and 'service' common tags, since they are not emitted by metrics backend.
	defaultTags := map[string]string{
		"env":     env,
		"service": service,
	}
	gateway.RootScope, gateway.scopeCloser = tally.NewRootScope(
		tally.ScopeOptions{
			Tags:           defaultTags,
			Prefix:         service + ".production",
			CachedReporter: gateway.metricsBackend,
			Separator:      tally.DefaultSeparator,
		},
		time.Duration(config.MustGetInt("metrics.tally.flushInterval"))*time.Millisecond,
	)
	gateway.MetricsScope = gateway.RootScope.SubScope("all-workers")
	gateway.PerWorkerMetricsScope = gateway.RootScope.SubScope("per-worker").Tagged(map[string]string{
		"host": GetHostname(),
	})

	// start collecting runtime metrics
	collectInterval := time.Duration(config.MustGetInt("metrics.runtime.collectInterval")) * time.Millisecond
	runtimeMetricsOpts := RuntimeMetricsOptions{
		EnableCPUMetrics: config.MustGetBoolean("metrics.runtime.enableCPUMetrics"),
		EnableMemMetrics: config.MustGetBoolean("metrics.runtime.enableMemMetrics"),
		EnableGCMetrics:  config.MustGetBoolean("metrics.runtime.enableGCMetrics"),
		CollectInterval:  collectInterval,
	}
	gateway.runtimeMetrics = StartRuntimeMetricsCollector(runtimeMetricsOpts, gateway.PerWorkerMetricsScope)

	return nil
}

func (gateway *Gateway) setupLogger(config *StaticConfig) error {
	var output zapcore.WriteSyncer
	tempLogger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			os.Stderr,
			zap.InfoLevel,
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
			tempLogger.Error("Error creating log directory",
				zap.String("error", err.Error()),
			)
			return errors.Wrap(err, "Error creating log directory")
		}

		loggerFile, err := os.OpenFile(
			loggerFileName,
			os.O_APPEND|os.O_WRONLY|os.O_CREATE,
			0644,
		)
		if err != nil {
			tempLogger.Error("Error opening log file",
				zap.String("error", err.Error()),
			)
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

	zapLogger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			output,
			zap.InfoLevel,
		),
	)

	host := GetHostname()

	datacenter := gateway.Config.MustGetString("datacenter")

	// Default to a STDOUT logger
	gateway.Logger = zapLogger.With(
		zap.String("hostname", host),
		zap.Int("pid", os.Getpid()),
		zap.String("zone", datacenter),
	)
	return nil
}

func (gateway *Gateway) setupTracer(config *StaticConfig) error {
	opts := []jaegerConfig.Option{
		// TChannel logger implements jaeger logger interface
		jaegerConfig.Logger(NewTChannelLogger(gateway.Logger)),
		jaegerConfig.Metrics(NewJaegerMetricsFactory(gateway.MetricScope)),
	}
	// TODO: using default tracing configurations here, we may want to set it from static config
	jc := jaegerConfig.Configuration{}
	serviceName := config.MustGetString("tchannel.serviceName")
	closer, err := jc.InitGlobalTracer(serviceName, opts...)
	if err != nil {
		return errors.Wrapf(err, "error initializing Jaeger tracer client")
	}

	gateway.Tracer = opentracing.GlobalTracer()
	gateway.tracerCloser = closer
	return nil
}

func (gateway *Gateway) setupHTTPServer() error {
	listenIP, err := tchannel.ListenIP()
	if err != nil {
		return errors.Wrap(err, "error finding the best IP")
	}
	gateway.httpServer = &HTTPServer{
		Server: &http.Server{
			Addr:    listenIP.String() + ":" + strconv.FormatInt(int64(gateway.HTTPPort), 10),
			Handler: gateway.HTTPRouter,
		},
		Logger: gateway.Logger,
	}

	gateway.localHTTPServer = &HTTPServer{
		Server: &http.Server{
			Addr:    "127.0.0.1:" + strconv.FormatInt(int64(gateway.HTTPPort), 10),
			Handler: gateway.HTTPRouter,
		},
		Logger: gateway.Logger,
	}
	return nil
}

func (gateway *Gateway) setupTChannel(config *StaticConfig) error {
	serviceName := config.MustGetString("tchannel.serviceName")
	processName := config.MustGetString("tchannel.processName")

	subScope := gateway.MetricsScope.SubScope("tchannel")
	channel, err := tchannel.NewChannel(
		serviceName,
		&tchannel.ChannelOptions{
			ProcessName:   processName,
			Logger:        NewTChannelLogger(gateway.Logger),
			StatsReporter: NewTChannelStatsReporter(subScope),
			Tracer:        gateway.Tracer,

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
	gateway.TChannelRouter = NewTChannelRouter(channel, gateway.Logger)

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
