// Code generated by zanzibar
// @generated

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

package main

import (
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var cachedServer *zanzibar.Gateway

func TestMain(m *testing.M) {
	readFlags()
	if os.Getenv("GATEWAY_RUN_CHILD_PROCESS_TEST") != "" {
		listenOnSignals()

		code := m.Run()
		os.Exit(code)
	} else {
		os.Exit(0)
	}
}

func listenOnSignals() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGUSR2)

	go func() {
		<-sigs

		if cachedServer != nil {
			cachedServer.Close()
		}
	}()
}

func TestStartGateway(t *testing.T) {
	testLogger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			os.Stderr,
			zap.DebugLevel,
		),
	)

	gateway, deps, err := createGateway()
	if err != nil {
		testLogger.Error(
			"Failed to CreateGateway in TestStartGateway()",
			zap.Error(err),
		)
		return
	}
	assert.NotNil(t, deps)

	cachedServer = gateway
	err = gateway.Bootstrap()
	if err != nil {
		testLogger.Error(
			"Failed to Bootstrap in TestStartGateway()",
			zap.Error(err),
		)
		return
	}
	logAndWait(gateway)
}

func logAndWait(server *zanzibar.Gateway) {
	server.Logger.Info("Started Example-gateway",
		zap.String("realHTTPAddr", server.RealHTTPAddr),
		zap.String("realTChannelAddr", server.RealTChannelAddr),
		zap.Any("config", server.InspectOrDie()),
	)

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		server.WaitGroup.Add(1)
		server.Shutdown()
		server.WaitGroup.Done()
	}()
	server.Wait()
}
