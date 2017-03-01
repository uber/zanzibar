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
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/uber/zanzibar/examples/example-gateway/clients"
	"github.com/uber/zanzibar/examples/example-gateway/endpoints"
	"github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/test_backend"
	"github.com/uber/zanzibar/test/lib/test_gateway"
)

// BenchGateway for testing
type BenchGateway struct {
	ActualGateway *zanzibar.Gateway

	backends   map[string]*testBackend.TestBackend
	httpClient *http.Client
}

func getProjectDir() string {
	goPath := os.Getenv("GOPATH")
	return path.Join(goPath, "src", "github.com", "uber", "zanzibar")
}

// CreateGateway bootstrap gateway for testing
func CreateGateway(
	seedConfig map[string]interface{}, opts *testGateway.Options,
) (testGateway.TestGateway, error) {
	if seedConfig == nil {
		seedConfig = map[string]interface{}{}
	}
	if opts == nil {
		opts = &testGateway.Options{}
	}

	backends, err := testBackend.BuildBackends(seedConfig, opts.KnownBackends)
	if err != nil {
		return nil, err
	}

	seedConfig["port"] = int64(0)
	seedConfig["tchannel.serviceName"] = "bench-gateway"
	seedConfig["tchannel.processName"] = "bench-gateway"
	seedConfig["metrics.tally.service"] = "bench-gateway"

	benchGateway := &BenchGateway{
		httpClient: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
			},
		},
		backends: backends,
	}

	config := zanzibar.NewStaticConfigOrDie([]string{
		filepath.Join(getProjectDir(), "config", "production.json"),
		filepath.Join(
			getProjectDir(),
			"examples",
			"example-gateway",
			"config",
			"production.json",
		),
	}, seedConfig)

	clients := clients.CreateClients(config)

	gateway, err := zanzibar.CreateGateway(config, &zanzibar.Options{
		Clients: clients,
	})
	if err != nil {
		return nil, err
	}
	benchGateway.ActualGateway = gateway
	err = gateway.Bootstrap(endpoints.Register)
	if err != nil {
		return nil, err
	}

	return benchGateway, nil
}

// GetPort ...
func (gateway *BenchGateway) GetPort() int {
	return int(gateway.ActualGateway.RealPort)
}

// Backends ...
func (gateway *BenchGateway) Backends() map[string]*testBackend.TestBackend {
	return gateway.backends
}

// MakeRequest helper
func (gateway *BenchGateway) MakeRequest(
	method string, url string, body io.Reader,
) (*http.Response, error) {
	client := gateway.httpClient

	fullURL := "http://" + gateway.ActualGateway.RealAddr + url

	req, err := http.NewRequest(method, fullURL, body)

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
