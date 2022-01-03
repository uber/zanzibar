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

package testbackend

import (
	"net/http"
	"strconv"
	"sync"

	zanzibar "github.com/uber/zanzibar/runtime"
	zrouter "github.com/uber/zanzibar/runtime/router"
	"go.uber.org/zap"
)

// TestHTTPBackend will pretend to be a http backend
type TestHTTPBackend struct {
	Server    *zanzibar.HTTPServer
	IP        string
	Port      int32
	RealPort  int32
	RealAddr  string
	WaitGroup *sync.WaitGroup
	router    *zrouter.Router
}

// BuildHTTPBackends returns a map of backends based on config
func BuildHTTPBackends(
	cfg map[string]interface{}, knownHTTPBackends []string,
) (map[string]*TestHTTPBackend, error) {
	n := len(knownHTTPBackends)
	result := make(map[string]*TestHTTPBackend, n)

	for i := 0; i < n; i++ {
		backend := CreateHTTPBackend(0)
		err := backend.Bootstrap()
		if err != nil {
			return nil, err
		}

		fieldName := knownHTTPBackends[i]
		result[fieldName] = backend
		cfg["clients."+fieldName+".ip"] = "127.0.0.1"
		cfg["clients."+fieldName+".port"] = int64(backend.RealPort)
	}

	return result, nil
}

// Bootstrap creates a backend for testing
func (backend *TestHTTPBackend) Bootstrap() error {
	_, err := backend.Server.JustListen()
	if err != nil {
		return err
	}

	backend.RealPort = backend.Server.RealPort
	backend.RealAddr = backend.Server.RealAddr

	backend.WaitGroup.Add(1)
	go backend.Server.JustServe(backend.WaitGroup)
	return nil
}

// HandleFunc registers funcs
func (backend *TestHTTPBackend) HandleFunc(
	method string, path string, handler http.HandlerFunc,
) {
	_ = backend.router.Handle(method, path, handler)
}

// Close ...
func (backend *TestHTTPBackend) Close() {
	backend.Server.Close()
	backend.WaitGroup.Wait()
}

// Wait ...
func (backend *TestHTTPBackend) Wait() {
	backend.WaitGroup.Wait()
}

// CreateHTTPBackend creates a HTTP backend for testing
func CreateHTTPBackend(port int32) *TestHTTPBackend {
	backend := &TestHTTPBackend{
		IP:        "127.0.0.1",
		Port:      port,
		WaitGroup: &sync.WaitGroup{},
		router: &zrouter.Router{
			HandleMethodNotAllowed: true,
		},
	}

	testLogger := zap.NewNop()
	backend.Server = &zanzibar.HTTPServer{
		Server: &http.Server{
			Addr:    backend.IP + ":" + strconv.Itoa(int(port)),
			Handler: backend.router,
		},
		Logger: testLogger,
	}

	return backend
}
