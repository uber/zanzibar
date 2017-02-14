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

package startGatewayTest

import (
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"syscall"
	"testing"
	"time"

	"github.com/uber-go/tally/m3"
	"github.com/uber-go/zap"

	"github.com/uber/zanzibar/examples/example-gateway/clients"
	"github.com/uber/zanzibar/examples/example-gateway/config"
	"github.com/uber/zanzibar/examples/example-gateway/endpoints"
	"github.com/uber/zanzibar/runtime"
	"gopkg.in/yaml.v2"
)

const defaultM3MaxQueueSize = 10000
const defaultM3MaxPacketSize = 1440 // 1440kb in UDP M3MaxPacketSize
const defaultM3FlushInterval = 500 * time.Millisecond

var cachedServer *zanzibar.Gateway

func TestMain(m *testing.M) {
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
		_ = <-sigs

		if cachedServer != nil {
			cachedServer.Close()
		}
	}()
}

func loadConfig(config *config.Config) error {
	configRoot := os.Getenv("CONFIG_DIR")
	filePath := path.Join(configRoot, "production.yaml")

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	return nil
}

func TestStartGateway(t *testing.T) {
	gatewayConfig := &config.Config{}
	tempLogger := zap.New(
		zap.NewJSONEncoder(),
		zap.Output(os.Stderr),
	)

	if err := loadConfig(gatewayConfig); err != nil {
		tempLogger.Error("Error initializing configuration",
			zap.String("error", err.Error()),
		)
		// ?
		return
	}

	clientOpts := &clients.Options{}
	clientOpts.Contacts.IP = gatewayConfig.Clients.Contacts.IP
	clientOpts.Contacts.Port = gatewayConfig.Clients.Contacts.Port
	clientOpts.GoogleNow.IP = gatewayConfig.Clients.GoogleNow.IP
	clientOpts.GoogleNow.Port = gatewayConfig.Clients.GoogleNow.Port
	clients := clients.CreateClients(clientOpts)

	m3FlushIntervalConfig := gatewayConfig.Metrics.M3.FlushInterval
	var m3FlushInterval time.Duration
	if m3FlushIntervalConfig == 0 {
		m3FlushInterval = defaultM3FlushInterval
	} else {
		m3FlushInterval = m3FlushIntervalConfig
	}

	commonTags := map[string]string{"env": "test"}
	m3Backend, err := metrics.NewM3Backend(
		gatewayConfig.Metrics.M3.HostPort,
		gatewayConfig.Metrics.Tally.Service,
		commonTags, // default tags
		false,      // include host
		defaultM3MaxQueueSize,
		defaultM3MaxPacketSize,
		m3FlushInterval,
	)
	if err != nil {
		tempLogger.Error("Error initializing m3backend",
			zap.String("error", err.Error()),
		)
		// ?
		return
	}

	server, err := zanzibar.CreateGateway(&zanzibar.Options{
		IP:   gatewayConfig.IP,
		Port: gatewayConfig.Port,
		Logger: zanzibar.LoggerOptions{
			FileName: gatewayConfig.Logger.FileName,
		},
		Metrics: zanzibar.MetricsOptions{
			FlushInterval: gatewayConfig.Metrics.Tally.FlushInterval,
			Service:       gatewayConfig.Metrics.Tally.Service,
		},
		TChannel: zanzibar.TChannelOptions{
			ServiceName: gatewayConfig.TChannel.ServiceName,
			ProcessName: gatewayConfig.TChannel.ProcessName,
		},

		Clients:        clients,
		MetricsBackend: m3Backend,
	})
	if err != nil {
		tempLogger.Error(
			"Failed to CreateGateway in TestStartGateway()",
			zap.String("error", err.Error()),
		)
		// ?
		return
	}

	cachedServer = server
	err = server.Bootstrap(endpoints.Register)
	if err != nil {
		tempLogger.Error(
			"Failed to Bootstrap in TestStartGateway()",
			zap.String("error", err.Error()),
		)
		// ?
		return
	}
	server.Logger.Info("Started Gateway",
		zap.String("realAddr", server.RealAddr),
		zap.Object("config", gatewayConfig),
	)

	server.Wait()
}
