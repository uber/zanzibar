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

	"time"

	"github.com/uber-go/tally/m3"
	"github.com/uber/zanzibar/examples/example-gateway/clients"
	config "github.com/uber/zanzibar/examples/example-gateway/config"
	endpoints "github.com/uber/zanzibar/examples/example-gateway/endpoints"
	zanzibar "github.com/uber/zanzibar/runtime"
)

const defaultM3MaxQueueSize = 10000
const defaultM3MaxPacketSize = 1440 // 1440kb in UDP M3MaxPacketSize
const defaultM3FlushInterval = 500 * time.Millisecond

// BenchGateway for testing
type BenchGateway struct {
	ActualGateway *zanzibar.Gateway
	httpClient    *http.Client
}

// CreateGateway bootstrap gateway for testing
func CreateGateway(config *config.Config) (*BenchGateway, error) {
	config.IP = "127.0.0.1"
	config.Port = 0
	config.Metrics.M3.HostPort = "127.0.0.1:8053"
	config.Metrics.Tally.Service = "bench-example-gateway"
	config.Metrics.M3.FlushInterval = 500 * time.Millisecond
	config.Metrics.Tally.FlushInterval = 1 * time.Second

	benchGateway := &BenchGateway{
		httpClient: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
			},
		},
	}

	clientOpts := &clients.Options{}
	clientOpts.Contacts.IP = config.Clients.Contacts.IP
	clientOpts.Contacts.Port = config.Clients.Contacts.Port
	clientOpts.GoogleNow.IP = config.Clients.GoogleNow.IP
	clientOpts.GoogleNow.Port = config.Clients.GoogleNow.Port
	clients := clients.CreateClients(clientOpts)

	m3FlushIntervalConfig := config.Metrics.M3.FlushInterval
	var m3FlushInterval time.Duration
	if m3FlushIntervalConfig == 0 {
		m3FlushInterval = defaultM3FlushInterval
	} else {
		m3FlushInterval = m3FlushIntervalConfig
	}

	commonTags := map[string]string{"env": "bench"}
	m3Backend, err := metrics.NewM3Backend(
		config.Metrics.M3.HostPort,
		config.Metrics.Tally.Service,
		commonTags, // default tags
		false,      // include host
		defaultM3MaxQueueSize,
		defaultM3MaxPacketSize,
		m3FlushInterval,
	)
	if err != nil {
		return nil, err
	}

	gateway, err := zanzibar.CreateGateway(&zanzibar.Options{
		IP:   config.IP,
		Port: config.Port,
		Logger: zanzibar.LoggerOptions{
			FileName: config.Logger.FileName,
		},
		Metrics: zanzibar.MetricsOptions{
			FlushInterval: config.Metrics.Tally.FlushInterval,
			Service:       config.Metrics.Tally.Service,
		},

		Clients:        clients,
		MetricsBackend: m3Backend,
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
