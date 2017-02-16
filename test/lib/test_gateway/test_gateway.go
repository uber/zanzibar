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
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally/m3"
	"github.com/uber/zanzibar/test/lib/test_backend"
	"github.com/uber/zanzibar/test/lib/test_m3_server"
)

// TestGateway interface
type TestGateway interface {
	MakeRequest(
		method string, url string, body io.Reader,
	) (*http.Response, error)
	Backends() map[string]*testBackend.TestBackend
	GetPort() int
	Close()
}

// ChildProcessGateway for testing
type ChildProcessGateway struct {
	cmd            *exec.Cmd
	binaryFileInfo *testBinaryInfo
	jsonLines      []string
	test           *testing.T
	opts           *Options
	httpClient     *http.Client
	m3Server       *testM3Server.FakeM3Server
	backends       map[string]*testBackend.TestBackend

	M3Service        *testM3Server.FakeM3Service
	MetricsWaitGroup sync.WaitGroup
	RealAddr         string
	RealHost         string
	RealPort         int
}

// Options used to create TestGateway
type Options struct {
	LogWhitelist  map[string]bool
	KnownBackends []string
	CountMetrics  bool
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
		false, countMetrics, metrics.Compact,
	)
	gateway.M3Service = gateway.m3Server.Service
	go gateway.m3Server.Serve()
}

// CreateGateway bootstrap gateway for testing
func CreateGateway(
	t *testing.T, config map[string]interface{}, opts *Options,
) (TestGateway, error) {
	if config == nil {
		config = map[string]interface{}{}
	}
	if opts == nil {
		opts = &Options{}
	}

	backends, err := testBackend.BuildBackends(config, opts.KnownBackends)
	if err != nil {
		return nil, err
	}

	testGateway := &ChildProcessGateway{
		test: t, opts: opts,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
			},
		},
		backends: backends,
	}

	testGateway.setupMetrics(t, opts)

	if _, contains := config["port"]; !contains {
		config["port"] = 0
	}

	config["ip"] = "127.0.0.1"
	config["tchannel.serviceName"] = "test-gateway"
	config["tchannel.processName"] = "test-gateway"
	config["metrics.m3.hostPort"] = testGateway.m3Server.Addr
	config["metrics.tally.service"] = "test-gateway"
	config["metrics.tally.flushInterval"] = 10
	config["metrics.m3.flushInterval"] = 10
	config["logger.output"] = "stdout"

	err = testGateway.createAndSpawnChild(config)
	if err != nil {
		return nil, err
	}

	return testGateway, nil
}

// MakeRequest helper
func (gateway *ChildProcessGateway) MakeRequest(
	method string, url string, body io.Reader,
) (*http.Response, error) {
	client := gateway.httpClient

	fullURL := "http://" + gateway.RealAddr + url

	req, err := http.NewRequest(method, fullURL, body)

	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

// Backends returns the backends
func (gateway *ChildProcessGateway) Backends() map[string]*testBackend.TestBackend {
	return gateway.backends
}

// GetPort ...
func (gateway *ChildProcessGateway) GetPort() int {
	return gateway.RealPort
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
		if gateway.opts == nil || !gateway.opts.LogWhitelist[msg] {
			assert.Fail(gateway.test,
				"Got unexpected error log from example-gateway:", line,
			)
		}
	}
}
