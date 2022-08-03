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

package testgateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/m3"
	"github.com/uber/jaeger-client-go/testutils"
	"github.com/uber/tchannel-go"
	"go.uber.org/zap"

	"github.com/uber/zanzibar/config"
	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib"
	testBackend "github.com/uber/zanzibar/test/lib/test_backend"
	testM3Server "github.com/uber/zanzibar/test/lib/test_m3_server"
)

// TestGateway interface
type TestGateway interface {
	MakeRequest(
		method string,
		url string,
		headers map[string]string,
		body io.Reader,
	) (*http.Response, error)
	// MakeRequestWithHeaderValues is an alternate version of `MakeRequest` that uses `zanzibar.Header`
	// instead of a `map[string]string` to represent headers. This allows us to fetch multiple values
	// for a given header key.
	MakeRequestWithHeaderValues(
		method string,
		url string,
		headers zanzibar.Header,
		body io.Reader,
	) (*http.Response, error)
	MakeTChannelRequest(
		ctx context.Context,
		thriftService string,
		method string,
		headers map[string]string,
		req, resp zanzibar.RWTStruct,
		timeoutAndRetryOptions *zanzibar.TimeoutAndRetryOptions,
	) (bool, map[string]string, error)
	HTTPBackends() map[string]*testBackend.TestHTTPBackend
	TChannelBackends() map[string]*testBackend.TestTChannelBackend
	HTTPPort() int
	Logs(level string, msg string) []LogMessage
	// AllLogs() returns a map of msg to a list of LogMessage
	AllLogs() map[string][]LogMessage

	Close()
	Config() *zanzibar.StaticConfig
}

// LogMessage is a json log record parsed into map.
type LogMessage map[string]interface{}

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
	logMessages      map[string][]LogMessage
	channel          *tchannel.Channel
	serviceName      string
	startTime        time.Time
	endTime          time.Time

	HTTPClient       *http.Client
	JaegerAgent      *testutils.MockAgent
	TChannelClient   zanzibar.TChannelCaller
	M3Service        *testM3Server.FakeM3Service
	MetricsWaitGroup lib.WaitAtLeast
	RealHTTPAddr     string
	RealHTTPHost     string
	RealHTTPPort     int
	RealTChannelAddr string
	RealTChannelHost string
	RealTChannelPort int
	ContextExtractor zanzibar.ContextExtractor
	ContextMetrics   zanzibar.ContextMetrics
	ContextLogger    zanzibar.ContextLogger
	staticConfig     *zanzibar.StaticConfig
}

// Options used to create TestGateway
type Options struct {
	TestBinary            string
	ConfigFiles           []string
	LogWhitelist          map[string]bool
	KnownHTTPBackends     []string
	KnownTChannelBackends []string
	CountMetrics          bool
	// If MaxMetrics is set we only collect the first N metrics upto
	// the max metrics amount.
	MaxMetrics            int
	EnableRuntimeMetrics  bool
	JaegerDisable         bool
	JaegerFlushMillis     int64
	TChannelClientMethods map[string]string
	Backends              []*testBackend.TestTChannelBackend
}

func (gateway *ChildProcessGateway) setupMetrics(
	t *testing.T,
	opts *Options,
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
	gateway.M3Service.MaxMetrics = opts.MaxMetrics
	go gateway.m3Server.Serve()
}

func (gateway *ChildProcessGateway) setupTracing() {
	agent, err := testutils.StartMockAgent()
	if err != nil {
		panic("unable to start mock jaeger agent")
	}
	gateway.JaegerAgent = agent
}

// CreateGateway bootstrap gateway for testing
func CreateGateway(
	t *testing.T,
	conf map[string]interface{},
	opts *Options,
) (TestGateway, error) {
	startTime := time.Now()

	composedConfig := map[string]interface{}{}

	for k, v := range conf {
		composedConfig[k] = v
	}

	if opts == nil {
		panic("opts in test.CreateGateway() mandatory")
	}
	if opts.TestBinary == "" {
		panic("opts.TestBinary in test.CreateGateway() mandatory")
	}

	backendsHTTP, err := testBackend.BuildHTTPBackends(composedConfig, opts.KnownHTTPBackends)
	if err != nil {
		return nil, err
	}

	staticConf := config.NewRuntimeConfigOrDie(opts.ConfigFiles, map[string]interface{}{})
	backendsTChannel, err := testBackend.BuildTChannelBackends(composedConfig, opts.KnownTChannelBackends,
		staticConf, opts.Backends)
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

	timeout := time.Duration(10000) * time.Millisecond
	timeoutPerAttempt := time.Duration(2000) * time.Millisecond
	if t, ok := conf["tchannel.client.timeout"]; ok {
		timeout = time.Duration(t.(int)) * time.Millisecond
	}
	if t, ok := conf["tchannel.client.timeoutPerAttempt"]; ok {
		timeoutPerAttempt = time.Duration(t.(int)) * time.Millisecond
	}

	scopeExtractor := func(ctx context.Context) map[string]string {
		tags := map[string]string{}
		headers := zanzibar.GetEndpointRequestHeadersFromCtx(ctx)
		tags["regionname"] = headers["Regionname"]
		tags["device"] = headers["Device"]
		tags["deviceversion"] = headers["Deviceversion"]

		return tags
	}

	logFieldsExtractors := func(ctx context.Context) []zap.Field {
		reqHeaders := zanzibar.GetEndpointRequestHeadersFromCtx(ctx)
		fields := make([]zap.Field, 0, len(reqHeaders))
		for k, v := range reqHeaders {
			fields = append(fields, zap.String(k, v))
		}
		return fields
	}

	extractors := &zanzibar.ContextExtractors{
		ScopeTagsExtractors: []zanzibar.ContextScopeTagsExtractor{scopeExtractor},
		LogFieldsExtractors: []zanzibar.ContextLogFieldsExtractor{logFieldsExtractors},
	}

	tchannelClient := zanzibar.NewTChannelClientContext(
		channel,
		zanzibar.NewContextLogger(zap.NewNop()),
		zanzibar.NewContextMetrics(tally.NoopScope),
		extractors,
		&zanzibar.TChannelClientOption{
			ServiceName:       serviceName,
			MethodNames:       opts.TChannelClientMethods,
			Timeout:           timeout,
			TimeoutPerAttempt: timeoutPerAttempt,
		},
	)

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
		logMessages:      map[string][]LogMessage{},
		backendsHTTP:     backendsHTTP,
		backendsTChannel: backendsTChannel,
		MetricsWaitGroup: lib.WaitAtLeast{},
		ContextExtractor: extractors,
		ContextMetrics:   zanzibar.NewContextMetrics(tally.NoopScope),
		staticConfig:     staticConf,
	}

	testGateway.setupMetrics(t, opts)

	if _, contains := composedConfig["http.port"]; !contains {
		composedConfig["http.port"] = 0
	}

	if _, contains := composedConfig["tchannel.port"]; !contains {
		composedConfig["tchannel.port"] = 0
	}

	if opts.JaegerFlushMillis >= 0 {
		composedConfig["jaeger.reporter.flush.milliseconds"] = opts.JaegerFlushMillis
	} else {
		composedConfig["jaeger.reporter.flush.milliseconds"] = 10000
	}
	composedConfig["jaeger.sampler.type"] = "const"

	if opts.JaegerDisable {
		composedConfig["jaeger.disabled"] = true
		composedConfig["jaeger.reporter.hostport"] = "localhost:6381"
		composedConfig["jaeger.sampler.param"] = 0
	} else {
		testGateway.setupTracing()
		composedConfig["jaeger.disabled"] = false
		composedConfig["jaeger.reporter.hostport"] = testGateway.JaegerAgent.SpanServerAddr()
		composedConfig["jaeger.sampler.param"] = 1
	}

	composedConfig["tchannel.serviceName"] = serviceName
	composedConfig["tchannel.processName"] = serviceName
	composedConfig["metrics.serviceName"] = serviceName
	composedConfig["metrics.flushInterval"] = 10
	composedConfig["metrics.m3.hostPort"] = testGateway.m3Server.Addr
	composedConfig["metrics.runtime.enableCPUMetrics"] = opts.EnableRuntimeMetrics
	composedConfig["metrics.runtime.enableMemMetrics"] = opts.EnableRuntimeMetrics
	composedConfig["metrics.runtime.enableGCMetrics"] = opts.EnableRuntimeMetrics
	composedConfig["metrics.runtime.collectInterval"] = 10
	composedConfig["logger.output"] = "stdout"
	composedConfig["logger.level"] = "debug"
	composedConfig["env"] = "test"

	err = testGateway.createAndSpawnChild(
		opts.TestBinary,
		opts.ConfigFiles,
		composedConfig,
	)

	if err != nil {
		return nil, err
	}

	return testGateway, nil
}

// Config returns static config loaded from file + seed config
func (gateway *ChildProcessGateway) Config() *zanzibar.StaticConfig {
	return gateway.staticConfig
}

// MakeRequest helper
func (gateway *ChildProcessGateway) MakeRequest(
	method string,
	url string,
	headers map[string]string,
	body io.Reader,
) (*http.Response, error) {
	client := gateway.HTTPClient

	fullURL := "http://" + gateway.RealHTTPAddr + url

	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}

	for headerName, headerValue := range headers {
		req.Header.Set(headerName, headerValue)
	}

	return client.Do(req)
}

// MakeRequestWithHeaderValues helper
func (gateway *ChildProcessGateway) MakeRequestWithHeaderValues(
	method string,
	url string,
	headers zanzibar.Header,
	body io.Reader,
) (*http.Response, error) {
	client := gateway.HTTPClient

	fullURL := "http://" + gateway.RealHTTPAddr + url

	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}

	// For each key, fetch every disparate header value and add
	// it to the test gateway request.
	keys := headers.Keys()
	for _, key := range keys {
		if values, found := headers.Values(key); found {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
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
	timeoutAndRetryOptions *zanzibar.TimeoutAndRetryOptions,
) (bool, map[string]string, error) {
	sc := gateway.channel.GetSubChannel(gateway.serviceName)
	sc.Peers().Add(gateway.RealTChannelAddr)

	return gateway.TChannelClient.Call(ctx, thriftService, method, headers, req, res, timeoutAndRetryOptions)
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

// Logs ...
func (gateway *ChildProcessGateway) Logs(
	level string,
	msg string,
) []LogMessage {
	// Logs can be a little late...
	// So just wait a bit...
	time.Sleep(time.Millisecond * 15)

	lines := gateway.logMessages[msg]
	for _, line := range lines {
		if line["level"].(string) != level {
			return nil
		}
	}

	return lines
}

// AllLogs ...
func (gateway *ChildProcessGateway) AllLogs() map[string][]LogMessage {
	// Logs can be a little late...
	// So just wait a bit...
	time.Sleep(time.Millisecond * 15)

	return gateway.logMessages
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
