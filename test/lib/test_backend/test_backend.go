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

package testBackend

import (
	"sync"

	"net/http"

	"strconv"

	"github.com/julienschmidt/httprouter"
	zap "github.com/uber-go/zap"
	"github.com/uber/zanzibar/runtime"
)

// TestBackend will pretend to be a http backend
type TestBackend struct {
	Server    *zanzibar.HTTPServer
	IP        string
	Port      int32
	RealPort  int32
	RealAddr  string
	WaitGroup *sync.WaitGroup
	router    *httprouter.Router
}

// BuildBackends returns a map of backends based on config
func BuildBackends(
	cfg map[string]interface{}, knownBackends []string,
) (map[string]*TestBackend, error) {
	n := len(knownBackends)
	result := make(map[string]*TestBackend, n)

	for i := 0; i < n; i++ {
		backend := CreateBackend(0)
		err := backend.Bootstrap()
		if err != nil {
			return nil, err
		}

		fieldName := knownBackends[i]
		result[fieldName] = backend
		cfg["clients."+fieldName+".ip"] = "127.0.0.1"
		cfg["clients."+fieldName+".port"] = int64(backend.RealPort)
	}

	return result, nil
}

// Bootstrap creates a backend for testing
func (backend *TestBackend) Bootstrap() error {
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
func (backend *TestBackend) HandleFunc(
	method string, path string, handler http.HandlerFunc,
) {
	backend.router.HandlerFunc(method, path, handler)
}

// Close ...
func (backend *TestBackend) Close() {
	backend.Server.Close()
	backend.WaitGroup.Wait()
}

// Wait ...
func (backend *TestBackend) Wait() {
	backend.WaitGroup.Wait()
}

// CreateBackend creates a backend for testing
func CreateBackend(port int32) *TestBackend {
	backend := &TestBackend{
		IP:        "127.0.0.1",
		Port:      port,
		WaitGroup: &sync.WaitGroup{},
		router: &httprouter.Router{
			HandleMethodNotAllowed: true,
		},
	}

	testLogger := zap.New(zap.NewJSONEncoder())

	backend.Server = &zanzibar.HTTPServer{
		Server: &http.Server{
			Addr:    backend.IP + ":" + strconv.Itoa(int(port)),
			Handler: backend.router,
		},
		Logger: testLogger,
	}

	return backend
}
