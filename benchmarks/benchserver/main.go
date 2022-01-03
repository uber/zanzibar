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

package main

import (
	"context"
	"net/http"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	baz "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/baz/baz"
	testBackend "github.com/uber/zanzibar/test/lib/test_backend"
)

const (
	httpPort     int32 = 8092
	tchannelPort int32 = 8094
)

func main() {
	var logger = zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			os.Stderr,
			zap.DebugLevel,
		),
	)

	httpBackend := serveHTTP(logger)
	_ = serverTChannel(logger)

	httpBackend.Wait()
}

func serveHTTP(logger *zap.Logger) *testBackend.TestHTTPBackend {
	httpBackend := testBackend.CreateHTTPBackend(httpPort)
	err := httpBackend.Bootstrap()
	if err != nil {
		panic(err)
	}
	handleContacts := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(202)
		_, _ = w.Write([]byte("{}"))
	}
	handleGoogleNow := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(202)
	}
	httpBackend.HandleFunc("POST", "/foo/contacts", handleContacts)
	httpBackend.HandleFunc("POST", "/add-credentials", handleGoogleNow)

	logger.Info("HTTP server listening on port & serving")

	return httpBackend
}

func serverTChannel(logger *zap.Logger) *testBackend.TestTChannelBackend {
	tchannelBackend, err := testBackend.CreateTChannelBackend(tchannelPort, "Qux")
	if err != nil {
		panic(err)
	}

	handleSimpleServiceCall := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SimpleService_Call_Args,
	) (map[string]string, error) {
		return nil, nil
	}
	simpleServiceCallHandler := baz.NewSimpleServiceCallHandler(handleSimpleServiceCall)

	// must register handler first before bootstrap
	err = tchannelBackend.Register("baz", "call", "SimpleService::call", simpleServiceCallHandler)
	if err != nil {
		panic(err)
	}

	err = tchannelBackend.Bootstrap()
	if err != nil {
		panic(err)
	}

	logger.Info("TChannel server listening on port & serving")
	return tchannelBackend
}
