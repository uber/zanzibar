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

package bar_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"

	"net"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/test/lib/test_gateway"
)

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(file)
}

func TestBarNormalFailingJSONInBackend(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary: filepath.Join(
			getDirName(), "..", "..", "..",
			"examples", "example-gateway", "build", "main.go",
		),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte("bad bytes")); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"POST", "/bar/bar-path", nil,
		bytes.NewReader([]byte(`{"stringField":"foo"}`)),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "500 Internal Server Error", res.Status)
	assert.Equal(t, 1, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t,
		"{\"error\":\"could not make client request:: "+
			"Could not parse client(bar) json: "+
			"parse error: "+
			"invalid character 'b' after top-level value "+
			"near offset 0 of 'bad bytes'\"}",
		string(respBytes),
	)
}

func TestBarNormalMalformedClientResponseReadAll(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		LogWhitelist: map[string]bool{
			"Could not ReadAll() client body": true,
		},
		TestBinary: filepath.Join(
			getDirName(), "..", "..", "..",
			"examples", "example-gateway", "build", "main.go",
		),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].Server.ConnState =
		func(conn net.Conn, state http.ConnState) {
			_, _ = conn.Write([]byte(
				"HTTP/1.1 500 Internal Server Error\n" +
					"Content-Length: 12\n" +
					"\n" +
					"abc\n"))
			_ = conn.Close()
		}

	res, err := gateway.MakeRequest(
		"POST", "/bar/bar-path", nil,
		bytes.NewReader([]byte(`{"stringField":"foo"}`)),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "500 Internal Server Error", res.Status)
	assert.Equal(t, 0, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Contains(t, string(respBytes),
		"Could not read client(bar) response body",
	)
}
