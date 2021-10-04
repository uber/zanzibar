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

package xyzservicegeneratedmock

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/uber/zanzibar/v1/config"
	zanzibar "github.com/uber/zanzibar/v1/runtime"

	service "github.com/uber/zanzibar/v1/examples/example-gateway/build/app/demo/services/xyz"
)

// MockService interface
type MockService interface {
	MakeHTTPRequest(
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
	MockClients() *MockClientNodes
	Server() *zanzibar.Gateway
	Start()
	Stop()
}

type mockService struct {
	started        bool
	server         *zanzibar.Gateway
	ctrl           *gomock.Controller
	mockClients    *MockClientNodes
	httpClient     *http.Client
	tChannelClient zanzibar.TChannelCaller
}

// MustCreateTestService creates a new MockService, panics if it fails doing so.
// Optional testConfigPaths specifies runtime config files used in tests, it
// should be paths that are relative to "$GOPATH/src".
// If testConfigPaths is absent, a default test config file is used.
// The default test config file is chosen base on existence in order below:
// - "../../config/test.yaml" where current dir is the dir of service-config.yaml for the mocked service
// - "config/test.yaml" where current dir is the project root
func MustCreateTestService(t *testing.T, testConfigPaths ...string) MockService {
	if len(testConfigPaths) == 0 {
		configPath := filepath.Join("src", "github.com/uber/zanzibar/v1/examples/example-gateway/config/test.yaml")
		defaultPath := filepath.Join(os.Getenv("GOPATH"), configPath)

		// This is a temporary solution for running tests using bazel
		// see https://docs.bazel.build/versions/master/test-encyclopedia.html for relevant env vars
		// TODO: need long term solution to avoid hardcoding bazel specifics
		bazelPath := filepath.Join(os.Getenv("TEST_SRCDIR"), os.Getenv("TEST_WORKSPACE"), configPath)

		testConfigPaths = append(
			testConfigPaths,
			defaultPath,
			bazelPath,
		)
	}
	c := config.NewRuntimeConfigOrDie(testConfigPaths, nil)

	server, err := zanzibar.CreateGateway(c, nil)
	if err != nil {
		panic(err)
	}

	ctrl := gomock.NewController(t)
	_, dependencies, mockNodes := InitializeDependenciesMock(server, ctrl)
	registerErr := service.RegisterDeps(server, dependencies)
	if registerErr != nil {
		panic(registerErr)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   false,
			MaxIdleConns:        500,
			MaxIdleConnsPerHost: 500,
		},
	}

	// TChannel clients must have a timeout defined, so defining a long timeout that should not realistically
	// be hit in tests.
	timeout := time.Duration(5) * time.Minute
	timeoutPerAttempt := time.Duration(1) * time.Minute

	tchannelClient := zanzibar.NewRawTChannelClient(
		server.ServerTChannel,
		server.ContextLogger,
		server.RootScope,
		&zanzibar.TChannelClientOption{
			ServiceName:       server.ServiceName,
			ClientID:          "TestClient",
			Timeout:           timeout,
			TimeoutPerAttempt: timeoutPerAttempt,
		},
	)

	return &mockService{
		server:         server,
		ctrl:           ctrl,
		mockClients:    mockNodes,
		httpClient:     httpClient,
		tChannelClient: tchannelClient,
	}
}

// Server returns the mock server
func (m *mockService) Server() *zanzibar.Gateway {
	return m.server
}

// Start starts the mock server, panics if fails doing so
func (m *mockService) Start() {
	if err := m.server.Bootstrap(); err != nil {
		panic(err)
	}
	m.started = true
}

// Stop stops the mock server
func (m *mockService) Stop() {
	// m.ctrl.Finish() calls runtime.Goexit() on errors
	// put it in defer so cleanup is always done
	defer func() {
		m.server.Shutdown()
		m.started = false
	}()
	m.ctrl.Finish()
}

// MockClients returns the mock clients
func (m *mockService) MockClients() *MockClientNodes {
	return m.mockClients
}

// MakeHTTPRequest makes a HTTP request to the mock server
func (m *mockService) MakeHTTPRequest(
	method string,
	url string,
	headers map[string]string,
	body io.Reader,
) (*http.Response, error) {
	if !m.started {
		return nil, errors.New("mock server is not started")
	}

	client := m.httpClient

	fullURL := "http://" + m.server.RealHTTPAddr + url

	req, err := http.NewRequest(method, fullURL, body)
	for headerName, headerValue := range headers {
		req.Header.Set(headerName, headerValue)
	}

	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

// MakeTChannelRequest makes a TChannel request to the mock server
func (m *mockService) MakeTChannelRequest(
	ctx context.Context,
	thriftService string,
	method string,
	headers map[string]string,
	req, res zanzibar.RWTStruct,
) (bool, map[string]string, error) {
	if !m.started {
		return false, nil, errors.New("mock server is not started")
	}

	sc := m.server.ServerTChannel.GetSubChannel(m.server.ServiceName)
	sc.Peers().Add(m.server.RealTChannelAddr)
	return m.tChannelClient.Call(ctx, thriftService, method, headers, req, res)
}
