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
	"github.com/uber-go/zap"
	"github.com/uber/tchannel-go"

	zt "github.com/uber/zanzibar/runtime/tchannel"
)

const defaultM3MaxQueueSize = 10000
const defaultM3MaxPacketSize = 1440 // 1440kb in UDP M3MaxPacketSize
const defaultM3FlushInterval = 500 * time.Millisecond

// Clients interface is a placeholder for the generated clients
type Clients interface {
}

// Options configures the gateway
type Options struct {
	MetricsBackend tally.CachedStatsReporter
	LogWriter      zap.WriteSyncer
}

// Gateway type
type Gateway struct {
	IP          string
	Port        int32
	RealPort    int32
	RealAddr    string
	WaitGroup   *sync.WaitGroup
	Clients     Clients
	Channel     *tchannel.Channel
	Logger      zap.Logger
	MetricScope tally.Scope
	ServiceName string
	Config      *StaticConfig
	Router      *Router

	loggerFile        *os.File
	metricScopeCloser io.Closer
	metricsBackend    tally.CachedStatsReporter
	logWriter         zap.WriteSyncer
	server            *HTTPServer
	localServer       *HTTPServer
	tchannelServer    *zt.Server
	// clients?
	//	- panic ???
	//	- process reporter ?
}

// CreateGateway func
func CreateGateway(
	config *StaticConfig, opts *Options,
) (*Gateway, error) {
	var metricsBackend tally.CachedStatsReporter
	var logWriter zap.WriteSyncer
	if opts != nil && opts.MetricsBackend != nil {
		metricsBackend = opts.MetricsBackend
	}
	if opts != nil && opts.LogWriter != nil {
		logWriter = opts.LogWriter
	}

	// TODO: (lu) channel name and opts
	ch, err := tchannel.NewChannel("Gateway", nil)
	if err != nil {
		return nil, err
	}

	gateway := &Gateway{
		IP:          config.MustGetString("ip"),
		Port:        int32(config.MustGetInt("port")),
		ServiceName: config.MustGetString("serviceName"),
		WaitGroup:   &sync.WaitGroup{},
		Config:      config,
		Channel:     ch,

		logWriter:      logWriter,
		metricsBackend: metricsBackend,
	}

	gateway.Router = NewRouter(gateway)

	if err := gateway.setupLogger(config); err != nil {
		return nil, err
	}

	if err := gateway.setupMetrics(config); err != nil {
		return nil, err
	}

	if err := gateway.setupHTTPServer(); err != nil {
		return nil, err
	}

	if err := gateway.setupTChannel(config); err != nil {
		return nil, err
	}

	return gateway, nil
}

// RegisterFn type used to avoid cyclic dependencies
type RegisterFn func(gateway *Gateway, router *Router)

// Bootstrap func
func (gateway *Gateway) Bootstrap(register RegisterFn) error {
	gateway.register(register)

	_, err := gateway.localServer.JustListen()
	if err != nil {
		gateway.Logger.Error("Error listening on port",
			zap.String("error", err.Error()),
		)
		return errors.Wrap(err, "error listening on port")
	}
	if gateway.localServer.RealIP != gateway.IP {
		_, err := gateway.server.JustListen()
		if err != nil {
			gateway.Logger.Error("Error listening on port",
				zap.String("error", err.Error()),
			)
			return errors.Wrap(err, "error listening on port")
		}
	} else {
		// Do not start at the same IP
		gateway.server = gateway.localServer
	}
	gateway.RealPort = gateway.server.RealPort
	gateway.RealAddr = gateway.server.RealAddr

	gateway.WaitGroup.Add(1)
	go gateway.server.JustServe(gateway.WaitGroup)

	return nil
}

func (gateway *Gateway) register(register RegisterFn) {
	gateway.Router.RegisterRaw("GET", "/debug/pprof", pprof.Index)
	gateway.Router.RegisterRaw("GET", "/debug/pprof/cmdline", pprof.Cmdline)
	gateway.Router.RegisterRaw("GET", "/debug/pprof/profile", pprof.Profile)
	gateway.Router.RegisterRaw("GET", "/debug/pprof/symbol", pprof.Symbol)
	gateway.Router.RegisterRaw("POST", "/debug/pprof/symbol", pprof.Symbol)
	gateway.Router.RegisterRaw(
		"GET", "/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP,
	)
	gateway.Router.RegisterRaw(
		"GET", "/debug/pprof/heap", pprof.Handler("heap").ServeHTTP,
	)
	gateway.Router.RegisterRaw(
		"GET", "/debug/pprof/threadcreate",
		pprof.Handler("threadcreate").ServeHTTP,
	)
	gateway.Router.RegisterRaw(
		"GET", "/debug/pprof/block", pprof.Handler("block").ServeHTTP,
	)

	gateway.Router.Register("GET", "/health", NewEndpoint(
		gateway, "health", "health", gateway.handleHealthRequest,
	))

	register(gateway, gateway.Router)
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

	res.WriteJSONBytes(200, bytes)
}

// Close the http server
func (gateway *Gateway) Close() {
	if gateway.loggerFile != nil {
		_ = gateway.loggerFile.Sync()
		_ = gateway.loggerFile.Close()
	}

	gateway.metricsBackend.Flush()
	_ = gateway.metricScopeCloser.Close()
	if gateway.localServer != gateway.server {
		gateway.localServer.Close()
	}
	gateway.server.Close()
}

// InspectOrDie inspects the config for this gateway
func (gateway *Gateway) InspectOrDie() map[string]interface{} {
	return gateway.Config.InspectOrDie()
}

// Wait for gateway to close the server
func (gateway *Gateway) Wait() {
	gateway.WaitGroup.Wait()
}

func (gateway *Gateway) setupMetrics(config *StaticConfig) error {
	metricsType := config.MustGetString("metrics.type")

	var metricsBackend tally.CachedStatsReporter
	if metricsType == "m3" {
		if gateway.metricsBackend != nil {
			panic("expected no metrics backend in gateway.")
		}

		env := config.MustGetString("env")

		commonTags := map[string]string{"env": env}
		m3Backend, err := m3.NewReporter(m3.Options{
			HostPorts:          []string{config.MustGetString("metrics.m3.hostPort")},
			Service:            config.MustGetString("metrics.tally.service"),
			CommonTags:         commonTags,
			Env:                env,
			IncludeHost:        false,
			MaxQueueSize:       defaultM3MaxQueueSize,
			MaxPacketSizeBytes: defaultM3MaxPacketSize,
		})
		if err != nil {
			return err
		}

		metricsBackend = m3Backend
		gateway.metricsBackend = m3Backend
	} else {
		if gateway.metricsBackend == nil {
			panic("expected gateway to have MetricsBackend in opts")
		}
		metricsBackend = gateway.metricsBackend
	}

	// TODO: decide what default tags we want...
	defaultTags := map[string]string{}

	prefix := config.MustGetString("metrics.tally.service") +
		".production.all-workers"
	flushIntervalConfig := config.MustGetInt("metrics.tally.flushInterval")

	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Tags:           defaultTags,
		Prefix:         prefix,
		CachedReporter: metricsBackend,
		Separator:      tally.DefaultSeparator,
	}, time.Duration(flushIntervalConfig)*time.Millisecond)
	gateway.MetricScope = scope
	gateway.metricScopeCloser = closer

	return nil
}

func (gateway *Gateway) setupLogger(config *StaticConfig) error {
	var output zap.Option
	tempLogger := zap.New(
		zap.NewJSONEncoder(),
		zap.Output(os.Stderr),
	)

	loggerFileName := config.MustGetString("logger.fileName")
	loggerOutput := config.MustGetString("logger.output")

	if loggerFileName == "" || loggerOutput == "stdout" {
		var writer zap.WriteSyncer
		if gateway.logWriter != nil {
			writer = zap.MultiWriteSyncer(os.Stdout, gateway.logWriter)
		} else {
			writer = os.Stdout
		}

		output = zap.Output(writer)
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
			0666,
		)
		if err != nil {
			tempLogger.Error("Error opening log file",
				zap.String("error", err.Error()),
			)
			return errors.Wrap(err, "Error opening log file")
		}
		gateway.loggerFile = loggerFile
		output = zap.Output(loggerFile)
	}

	// Default to a STDOUT logger
	gateway.Logger = zap.New(
		zap.NewJSONEncoder(
			zap.RFC3339Formatter("ts"),
		),
		output,
	)
	return nil
}

func (gateway *Gateway) setupHTTPServer() error {
	listenIP, err := tchannel.ListenIP()
	if err != nil {
		return errors.Wrap(err, "error finding the best IP")
	}
	gateway.server = &HTTPServer{
		Server: &http.Server{
			Addr:    listenIP.String() + ":" + strconv.FormatInt(int64(gateway.Port), 10),
			Handler: gateway.Router,
		},
		Logger: gateway.Logger,
	}

	gateway.localServer = &HTTPServer{
		Server: &http.Server{
			Addr:    "127.0.0.1:" + strconv.FormatInt(int64(gateway.Port), 10),
			Handler: gateway.Router,
		},
		Logger: gateway.Logger,
	}
	return nil
}

func (gateway *Gateway) setupTChannel(config *StaticConfig) error {
	tchannelServer, err := NewTChannelServer(
		&TChannelServerOptions{
			ServiceName: config.MustGetString("tchannel.serviceName"),
			ProcessName: config.MustGetString("tchannel.processName"),
		}, gateway)

	if err != nil {
		return err
	}

	gateway.tchannelServer = tchannelServer
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
