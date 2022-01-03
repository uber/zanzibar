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

package testbackend

import (
	"context"
	"net"
	"os"
	"strconv"

	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/uber/zanzibar/config"
	zanzibar "github.com/uber/zanzibar/runtime"
)

// TestTChannelBackend will pretend to be a tchannel backend
type TestTChannelBackend struct {
	Channel     *tchannel.Channel
	Router      *zanzibar.TChannelRouter
	IP          string
	Port        int32
	RealPort    int32
	RealAddr    string
	ServiceName string
}

// BuildTChannelBackends returns a map of TChannel backends based on config
func BuildTChannelBackends(
	cfg map[string]interface{}, knownTChannelBackends []string, staticConfig *zanzibar.StaticConfig,
	tChannelBackends []*TestTChannelBackend,
) (map[string]*TestTChannelBackend, error) {
	n := len(knownTChannelBackends)
	result := make(map[string]*TestTChannelBackend, n)

	for clientIndex := 0; clientIndex < n; clientIndex++ {
		clientID := knownTChannelBackends[clientIndex]

		val, ok := cfg["clients."+clientID+".serviceName"]
		var serviceName string
		if !ok {
			serviceName = clientID
			cfg["clients."+clientID+".serviceName"] = serviceName
		} else {
			serviceName = val.(string)
		}
		cfg["clients."+clientID+".ip"] = "127.0.0.1"
		cfg["clients."+clientID+".timeout"] = int64(10000)
		cfg["clients."+clientID+".timeoutPerAttempt"] = int64(10000)

		backend, err := CreateTChannelBackend(int32(0), serviceName)
		if err != nil {
			return nil, err
		}
		err = backend.Bootstrap()
		if err != nil {
			return nil, err
		}
		result[clientID] = backend

		cfg["clients."+clientID+".port"] = int64(backend.RealPort)

		initializeAlternateBackends(result, clientID, tChannelBackends, cfg, staticConfig)
	}

	return result, nil
}

func initializeAlternateBackends(result map[string]*TestTChannelBackend, clientID string,
	backends []*TestTChannelBackend, cfg map[string]interface{}, staticConfig *zanzibar.StaticConfig) {

	if backends == nil {
		return
	}

	var altServiceDetail config.AlternateServiceDetail
	staticConfig.MustGetStruct("clients."+clientID+".alternates", &altServiceDetail)
	cfg["clients."+clientID+".alternates"] = &altServiceDetail

	// create backends for the same client for dynamic routing
	for backendIndex := 0; backendIndex < len(backends); backendIndex++ {
		backend := backends[backendIndex]
		altServiceDetail.ServicesDetailMap[backend.ServiceName].Port = int(backend.RealPort)
		transformedClientID := clientID + ":" + strconv.Itoa(backendIndex+1)
		result[transformedClientID] = backend
	}
}

// Bootstrap creates a backend for testing
func (backend *TestTChannelBackend) Bootstrap() error {
	addr := backend.IP + ":" + strconv.Itoa(int(backend.Port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	realAddr := ln.Addr().(*net.TCPAddr)
	backend.RealPort = int32(realAddr.Port)
	backend.RealAddr = realAddr.IP.String() + ":" + strconv.Itoa(int(backend.RealPort))

	// tchannel serve does not block, connection handling is done in different goroutine
	err = backend.Channel.Serve(ln)
	return err
}

// Register registers tchannel server handler
func (backend *TestTChannelBackend) Register(
	endpointID, handlerID, method string,
	handler zanzibar.TChannelHandler,
) error {
	return backend.Router.Register(zanzibar.NewTChannelEndpoint(
		endpointID, handlerID, method,
		handler,
	))
}

// Close closes the underlying channel
func (backend *TestTChannelBackend) Close() {
	backend.Channel.Close()
}

// CreateTChannelBackend creates a TChannel backend for testing
// "serviceName" is the service discovery name, not necessarily same as the thrift service name.
func CreateTChannelBackend(port int32, serviceName string) (*TestTChannelBackend, error) {
	backend := &TestTChannelBackend{
		IP:          "127.0.0.1",
		Port:        port,
		ServiceName: serviceName,
	}

	testLogger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			os.Stdout,
			zap.WarnLevel,
		),
	)
	testLogger = testLogger.With(
		zap.String("from", "test-tchannel-backend"),
		zap.String("test-backend", serviceName),
	)

	contextLogger := zanzibar.NewContextLogger(testLogger)

	tchannelOpts := &tchannel.ChannelOptions{
		Logger: zanzibar.NewTChannelLogger(testLogger),
	}

	channel, err := tchannel.NewChannel(serviceName, tchannelOpts)
	if err != nil {
		return nil, err
	}

	scopeExtractor := func(ctx context.Context) map[string]string {
		tags := map[string]string{}
		headers := zanzibar.GetEndpointRequestHeadersFromCtx(ctx)
		tags["regionname"] = headers["Regionname"]
		tags["device"] = headers["Device"]
		tags["deviceversion"] = headers["Deviceversion"]

		return tags
	}
	contextExtractors := &zanzibar.ContextExtractors{
		ScopeTagsExtractors: []zanzibar.ContextScopeTagsExtractor{scopeExtractor},
	}

	gateway := zanzibar.Gateway{
		Logger:           testLogger,
		RootScope:        tally.NoopScope,
		ContextExtractor: contextExtractors,
		ContextMetrics:   zanzibar.NewContextMetrics(tally.NoopScope),
		ContextLogger:    contextLogger,
	}

	backend.Channel = channel
	backend.Router = zanzibar.NewTChannelRouter(channel, &gateway)
	return backend, nil
}
