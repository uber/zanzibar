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

package benchgateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/uber/zanzibar/config"
	"github.com/uber/zanzibar/runtime"
	testBackend "github.com/uber/zanzibar/test/lib/test_backend"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"go.uber.org/zap/zapcore"
)

// BenchGateway for testing
type BenchGateway struct {
	ActualGateway *zanzibar.Gateway
	Dependencies  interface{}

	backendsHTTP     map[string]*testBackend.TestHTTPBackend
	backendsTChannel map[string]*testBackend.TestTChannelBackend
	logBytes         *bytes.Buffer
	readLogs         bool
	logMessages      map[string][]testGateway.LogMessage
	httpClient       *http.Client
	tchannelClient   zanzibar.TChannelCaller
}

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(file)
}

func getZanzibarDirName() string {
	return filepath.Join(getDirName(), "..", "..", "..")
}

// CreateGatewayFn creates a new gateway to benchmark
type CreateGatewayFn func(
	config *zanzibar.StaticConfig,
	opts *zanzibar.Options,
) (*zanzibar.Gateway, interface{}, error)

// CreateGateway bootstrap gateway for testing
func CreateGateway(
	seedConfig map[string]interface{},
	opts *testGateway.Options,
	createGateway CreateGatewayFn,
) (testGateway.TestGateway, error) {
	if seedConfig == nil {
		seedConfig = map[string]interface{}{}
	}
	if opts == nil {
		opts = &testGateway.Options{}
	}

	backendsHTTP, err := testBackend.BuildHTTPBackends(seedConfig, opts.KnownHTTPBackends)
	if err != nil {
		return nil, err
	}

	backendsTChannel, err := testBackend.BuildTChannelBackends(seedConfig, opts.KnownTChannelBackends)
	if err != nil {
		return nil, err
	}

	seedConfig["http.port"] = int64(0)
	seedConfig["tchannel.port"] = int64(0)
	seedConfig["envVarsToTagInRootScope"] = []string{}

	if _, ok := seedConfig["tchannel.serviceName"]; !ok {
		seedConfig["tchannel.serviceName"] = "bench-gateway"
	}
	if _, ok := seedConfig["logger.output"]; !ok {
		seedConfig["logger.output"] = "disk"
	}
	if _, ok := seedConfig["logger.fileName"]; !ok {
		seedConfig["logger.fileName"] = "zanzibar.log"
	}
	seedConfig["tchannel.processName"] = "bench-gateway"
	seedConfig["metrics.serviceName"] = "bench-gateway"

	benchGateway := &BenchGateway{
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				DisableKeepAlives:     false,
				MaxIdleConns:          50000,
				MaxIdleConnsPerHost:   50000,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   30 * time.Second,
				ExpectContinueTimeout: 30 * time.Second,
			},
			Timeout: 30 * 1000 * time.Millisecond,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		backendsHTTP:     backendsHTTP,
		backendsTChannel: backendsTChannel,
		logBytes:         bytes.NewBuffer(nil),

		readLogs:    false,
		logMessages: map[string][]testGateway.LogMessage{},
	}

	config := config.NewRuntimeConfigOrDie(opts.ConfigFiles, seedConfig)

	gateway, dependencies, err := createGateway(config, &zanzibar.Options{
		LogWriter: zapcore.AddSync(benchGateway.logBytes),
	})
	if err != nil {
		return nil, err
	}

	benchGateway.Dependencies = dependencies
	benchGateway.ActualGateway = gateway

	timeout := time.Duration(10000) * time.Millisecond
	timeoutPerAttempt := time.Duration(2000) * time.Millisecond
	if t, ok := seedConfig["tchannel.client.timeout"]; ok {
		timeout = time.Duration(t.(int)) * time.Millisecond
	}
	if t, ok := seedConfig["tchannel.client.timeoutPerAttempt"]; ok {
		timeoutPerAttempt = time.Duration(t.(int)) * time.Millisecond
	}

	benchGateway.tchannelClient = zanzibar.NewTChannelClient(
		gateway.Channel,
		gateway.Logger,
		gateway.RootScope,
		&zanzibar.TChannelClientOption{
			ServiceName:       gateway.ServiceName,
			MethodNames:       opts.TChannelClientMethods,
			Timeout:           timeout,
			TimeoutPerAttempt: timeoutPerAttempt,
		})

	err = gateway.Bootstrap()
	if err != nil {
		return nil, err
	}

	return benchGateway, nil
}

// HTTPPort ...
func (gateway *BenchGateway) HTTPPort() int {
	return int(gateway.ActualGateway.RealHTTPPort)
}

func (gateway *BenchGateway) buildLogs() {
	// Logs can be a little late...
	// So just wait a bit...
	time.Sleep(time.Millisecond * 15)

	lines := strings.Split(gateway.logBytes.String(), "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if len(line) == 0 {
			continue
		}

		lineStruct := map[string]interface{}{}
		jsonError := json.Unmarshal([]byte(line), &lineStruct)
		if jsonError != nil {
			// do not decode msg
			continue
		}

		msg := lineStruct["msg"].(string)

		msgLogs := gateway.logMessages[msg]
		if msgLogs == nil {
			msgLogs = []testGateway.LogMessage{lineStruct}
		} else {
			msgLogs = append(msgLogs, lineStruct)
		}
		gateway.logMessages[msg] = msgLogs
	}

	gateway.readLogs = true
}

// Logs ...
func (gateway *BenchGateway) Logs(
	level string, msg string,
) []testGateway.LogMessage {
	if !gateway.readLogs {
		gateway.buildLogs()
	}

	lines := gateway.logMessages[msg]
	for _, line := range lines {
		if line["level"].(string) != level {
			return nil
		}
	}

	return lines
}

// AllLogs ...
func (gateway *BenchGateway) AllLogs() map[string][]testGateway.LogMessage {
	if !gateway.readLogs {
		gateway.buildLogs()
	}

	return gateway.logMessages
}

// HTTPBackends returns the HTTP backends of the gateway
func (gateway *BenchGateway) HTTPBackends() map[string]*testBackend.TestHTTPBackend {
	return gateway.backendsHTTP
}

// TChannelBackends returns the TChannel backends of the gateway
func (gateway *BenchGateway) TChannelBackends() map[string]*testBackend.TestTChannelBackend {
	return gateway.backendsTChannel
}

// MakeRequest helper
func (gateway *BenchGateway) MakeRequest(
	method string, url string, headers map[string]string, body io.Reader,
) (*http.Response, error) {
	client := gateway.httpClient

	fullURL := "http://" + gateway.ActualGateway.RealHTTPAddr + url

	req, err := http.NewRequest(method, fullURL, body)
	for headerName, headerValue := range headers {
		req.Header.Set(headerName, headerValue)
	}

	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

// MakeTChannelRequest helper
func (gateway *BenchGateway) MakeTChannelRequest(
	ctx context.Context,
	thriftService string,
	method string,
	headers map[string]string,
	req, res zanzibar.RWTStruct,
) (bool, map[string]string, error) {
	sc := gateway.ActualGateway.Channel.GetSubChannel(gateway.ActualGateway.ServiceName)
	sc.Peers().Add(gateway.ActualGateway.RealTChannelAddr)

	return gateway.tchannelClient.Call(ctx, thriftService, method, headers, req, res)
}

// Close test gateway
func (gateway *BenchGateway) Close() {
	gateway.ActualGateway.Close()
	gateway.ActualGateway.Wait()
}
