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

package testGateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally/m3"
	"github.com/uber/tchannel-go"
	"github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/test_backend"
	"github.com/uber/zanzibar/test/lib/test_m3_server"
)

// TestGateway interface
type TestGateway interface {
	MakeRequest(
		method string,
		url string,
		headers map[string]string,
		body io.Reader,
	) (*http.Response, error)
	MakeTChannelRequest(
		ctx context.Context,
		thriftService string,
		method string,
		headers map[string]string,
		req, resp zanzibar.RWTStruct,
	) (bool, map[string]string, error)
	HTTPBackends() map[string]*testBackend.TestHTTPBackend
	TChannelBackends() map[string]*testBackend.TestTChannelBackend
	HTTPPort() int
	ErrorLogs() map[string][]string

	Close()
}

// ChildProcessGateway for testing
type ChildProcessGateway struct {
	cmd              *exec.Cmd
	binaryFileInfo   *testBinaryInfo
	jsonLines        []string
	test             *testing.T
	opts             *Options
	m3Server         *testM3Server.FakeM3Server
	backendsHTTP     map[string]*testBackend.TestHTTPBackend
	backendsTChannel map[string]*testBackend.TestTChannelBackend
	errorLogs        map[string][]string
	channel          *tchannel.Channel
	serviceName      string
	startTime        time.Time
	endTime          time.Time

	HTTPClient       *http.Client
	TChannelClient   zanzibar.TChannelClient
	M3Service        *testM3Server.FakeM3Service
	MetricsWaitGroup sync.WaitGroup
	RealHTTPAddr     string
	RealHTTPHost     string
	RealHTTPPort     int
	RealTChannelAddr string
	RealTChannelHost string
	RealTChannelPort int
}

// Options used to create TestGateway
type Options struct {
	TestBinary            string
	LogWhitelist          map[string]bool
	KnownHTTPBackends     []string
	KnownTChannelBackends []string
	CountMetrics          bool
}

func (gateway *ChildProcessGateway) setupMetrics(
	t *testing.T, opts *Options,
) {
	countMetrics := false
	if opts != nil {
		countMetrics = opts.CountMetrics
	}

	gateway.m3Server = testM3Server.NewFakeM3Server(
		t, &gateway.MetricsWaitGroup,
		false, countMetrics, m3.Compact,
	)
	gateway.M3Service = gateway.m3Server.Service
	go gateway.m3Server.Serve()
}

// CreateGateway bootstrap gateway for testing
func CreateGateway(
	t *testing.T, config map[string]interface{}, opts *Options,
) (TestGateway, error) {
	startTime := time.Now()

	if config == nil {
		config = map[string]interface{}{}
	}
	if opts == nil {
		panic("opts in test.CreateGateway() mandatory")
	}
	if opts.TestBinary == "" {
		panic("opts.TestBinary in test.CreateGateway() mandatory")
	}

	backendsHTTP, err := testBackend.BuildHTTPBackends(config, opts.KnownHTTPBackends)
	if err != nil {
		return nil, err
	}

	backendsTChannel, err := testBackend.BuildTChannelBackends(config, opts.KnownTChannelBackends)
	if err != nil {
		return nil, err
	}

	tchannelOpts := &tchannel.ChannelOptions{
		Logger: tchannel.NullLogger,
	}

	serviceName := "test-gateway"
	channel, err := tchannel.NewChannel(serviceName, tchannelOpts)
	if err != nil {
		return nil, err
	}

	tchannelClient := zanzibar.NewTChannelClient(channel, &zanzibar.TChannelClientOption{
		ServiceName:       serviceName,
		Timeout:           time.Duration(1000) * time.Millisecond,
		TimeoutPerAttempt: time.Duration(100) * time.Millisecond,
	})

	testGateway := &ChildProcessGateway{
		channel:     channel,
		serviceName: serviceName,
		test:        t,
		opts:        opts,
		startTime:   startTime,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
			},
		},
		TChannelClient:   tchannelClient,
		jsonLines:        []string{},
		errorLogs:        map[string][]string{},
		backendsHTTP:     backendsHTTP,
		backendsTChannel: backendsTChannel,
	}

	testGateway.setupMetrics(t, opts)

	if _, contains := config["http.port"]; !contains {
		config["http.port"] = 0
	}

	if _, contains := config["tchannel.port"]; !contains {
		config["tchannel.port"] = 0
	}

	config["tchannel.serviceName"] = serviceName
	config["tchannel.processName"] = serviceName
	config["metrics.m3.hostPort"] = testGateway.m3Server.Addr
	config["metrics.tally.service"] = serviceName
	config["metrics.tally.flushInterval"] = 10
	config["metrics.m3.flushInterval"] = 10
	config["logger.output"] = "stdout"

	err = testGateway.createAndSpawnChild(opts.TestBinary, config)
	if err != nil {
		return nil, err
	}

	return testGateway, nil
}

// MakeRequest helper
func (gateway *ChildProcessGateway) MakeRequest(
	method string, url string, headers map[string]string, body io.Reader,
) (*http.Response, error) {
	client := gateway.HTTPClient

	fullURL := "http://" + gateway.RealHTTPAddr + url

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
func (gateway *ChildProcessGateway) MakeTChannelRequest(
	ctx context.Context,
	thriftService string,
	method string,
	headers map[string]string,
	req, res zanzibar.RWTStruct,
) (bool, map[string]string, error) {
	sc := gateway.channel.GetSubChannel(gateway.serviceName)
	sc.Peers().Add(gateway.RealTChannelAddr)

	return gateway.TChannelClient.Call(ctx, thriftService, method, headers, req, res)
}

// HTTPBackends returns the HTTP backends
func (gateway *ChildProcessGateway) HTTPBackends() map[string]*testBackend.TestHTTPBackend {
	return gateway.backendsHTTP
}

// TChannelBackends returns the TChannel backends
func (gateway *ChildProcessGateway) TChannelBackends() map[string]*testBackend.TestTChannelBackend {
	return gateway.backendsTChannel
}

// HTTPPort ...
func (gateway *ChildProcessGateway) HTTPPort() int {
	return gateway.RealHTTPPort
}

// ErrorLogs ...
func (gateway *ChildProcessGateway) ErrorLogs() map[string][]string {
	return gateway.errorLogs
}

// Close test gateway
func (gateway *ChildProcessGateway) Close() {
	if gateway.cmd != nil {
		err := syscall.Kill(gateway.cmd.Process.Pid, syscall.SIGUSR2)
		if err != nil {
			panic(err)
		}

		_ = gateway.cmd.Wait()
	}

	if gateway.binaryFileInfo != nil {
		gateway.binaryFileInfo.Cleanup()
	}

	if gateway.m3Server != nil {
		_ = gateway.m3Server.Close()
	}

	// Sanity verify jsonLines
	for _, line := range gateway.jsonLines {
		lineStruct := map[string]interface{}{}
		jsonErr := json.Unmarshal([]byte(line), &lineStruct)
		if !assert.NoError(gateway.test, jsonErr, "logs must be json") {
			return
		}

		level := lineStruct["level"].(string)
		if level != "error" {
			continue
		}

		msg := lineStruct["msg"].(string)

		if gateway.opts != nil && gateway.opts.LogWhitelist[msg] {
			continue
		} else {
			assert.Fail(gateway.test,
				"Got unexpected error log from example-gateway:", line,
			)
		}
	}

	gateway.endTime = time.Now()
}
