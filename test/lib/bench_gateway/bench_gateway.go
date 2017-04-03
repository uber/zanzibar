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

package benchGateway

import (
	"io"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"bytes"

	"encoding/json"

	"github.com/uber-go/zap"
	"github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/test_backend"
	"github.com/uber/zanzibar/test/lib/test_gateway"
)

// BenchGateway for testing
type BenchGateway struct {
	ActualGateway *zanzibar.Gateway

	backendsHTTP map[string]*testBackend.TestHTTPBackend
	logBytes     *bytes.Buffer
	readLogs     bool
	errorLogs    map[string][]string
	httpClient   *http.Client
}

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(file)
}

func getZanzibarDirName() string {
	return filepath.Join(getDirName(), "..", "..", "..")
}

// CreateGateway bootstrap gateway for testing
func CreateGateway(
	seedConfig map[string]interface{},
	opts *testGateway.Options,
	createClients func(config *zanzibar.StaticConfig, gateway *zanzibar.Gateway) interface{},
	regEndpoints func(g *zanzibar.Gateway, router *zanzibar.Router),
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

	seedConfig["port"] = int64(0)

	if _, ok := seedConfig["tchannel.serviceName"]; !ok {
		seedConfig["tchannel.serviceName"] = "bench-gateway"
	}
	seedConfig["tchannel.processName"] = "bench-gateway"
	seedConfig["metrics.tally.service"] = "bench-gateway"
	seedConfig["logger.output"] = "stdout"

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
		},
		backendsHTTP: backendsHTTP,
		logBytes:     bytes.NewBuffer(nil),

		readLogs:  false,
		errorLogs: map[string][]string{},
	}

	config := zanzibar.NewStaticConfigOrDie([]string{
		filepath.Join(getZanzibarDirName(), "config", "production.json"),
		filepath.Join(
			getDirName(),
			"..",
			"..",
			"..",
			"examples",
			"example-gateway",
			"config",
			"production.json",
		),
	}, seedConfig)

	gateway, err := zanzibar.CreateGateway(config, &zanzibar.Options{
		LogWriter: zap.AddSync(benchGateway.logBytes),
	})
	if err != nil {
		return nil, err
	}
	gateway.Clients = createClients(config, gateway)

	benchGateway.ActualGateway = gateway
	err = gateway.Bootstrap(regEndpoints)
	if err != nil {
		return nil, err
	}

	return benchGateway, nil
}

// GetPort ...
func (gateway *BenchGateway) GetPort() int {
	return int(gateway.ActualGateway.RealPort)
}

// GetErrorLogs ...
func (gateway *BenchGateway) GetErrorLogs() map[string][]string {
	if gateway.readLogs {
		return gateway.errorLogs
	}

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

		msgLogs := gateway.errorLogs[msg]
		if msgLogs == nil {
			msgLogs = []string{line}
		} else {
			msgLogs = append(msgLogs, line)
		}
		gateway.errorLogs[msg] = msgLogs
	}

	gateway.readLogs = true
	return gateway.errorLogs
}

// HTTPBackends returns the HTTP backends of the gateway
func (gateway *BenchGateway) HTTPBackends() map[string]*testBackend.TestHTTPBackend {
	return gateway.backendsHTTP
}

// MakeRequest helper
func (gateway *BenchGateway) MakeRequest(
	method string, url string, headers map[string]string, body io.Reader,
) (*http.Response, error) {
	client := gateway.httpClient

	fullURL := "http://" + gateway.ActualGateway.RealAddr + url

	req, err := http.NewRequest(method, fullURL, body)
	for headerName, headerValue := range headers {
		req.Header.Set(headerName, headerValue)
	}

	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

// Close test gateway
func (gateway *BenchGateway) Close() {
	gateway.ActualGateway.Close()
	gateway.ActualGateway.Wait()
}
