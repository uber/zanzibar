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
	"testing"

	"net"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

func TestBarNormalFailingJSONInBackend(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
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
		bytes.NewReader([]byte(`{
			"request":{"stringField":"foo","boolField":true,"binaryField":"aGVsbG8=","timestamp":123,"enumField":0,"i64TestField":{"i64Field":123,"timestamp":1257894000000,"longField":1234}}
		}`)),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}
	assert.Equal(t, "500 Internal Server Error", res.Status)
	assert.Equal(t, 1, counter)

	res, err = gateway.MakeRequest(
		"POST", "/bar/bar-path", nil,
		bytes.NewReader([]byte(`{
			"request":{"stringField":"foo","boolField":true,"binaryField":"aGVsbG8=","timestamp":123,"enumField":"APPLE","i64TestField":{"i64Field":123,"timestamp":1257894000000,"longField":1234}}
		}`)),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "500 Internal Server Error", res.Status)
	assert.Equal(t, 2, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, string(respBytes),
		`{"error":"Unexpected server error"}`)
}

func TestBarNormalMalformedClientResponseReadAll(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		LogWhitelist: map[string]bool{
			"Could not read response body":    true,
			"Could not make outbound request": true,
		},
		TestBinary:  util.DefaultMainFile("example-gateway"),
		ConfigFiles: util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	endpoints := map[string]string{
		"/bar/bar-path": `{
			"request":{"stringField":"foo","boolField":true,"binaryField":"aGVsbG8=","timestamp":123,"enumField":0,"i64TestField":{"i64Field":123,"timestamp":1257894000000,"longField":1234}}
		}`,
		"/bar/arg-not-struct-path": `{"request":"foo"}`,
	}

	for k, v := range endpoints {
		gateway.HTTPBackends()["bar"].Server.ConnState =
			func(conn net.Conn, state http.ConnState) {
				_, _ = conn.Write([]byte(
					"HTTP/1.1 200 OK\n" +
						"Content-Length: 12\n" +
						"\n" +
						"abc\n"))
				_ = conn.Close()
			}

		res, err := gateway.MakeRequest(
			"POST", k, nil,
			bytes.NewReader([]byte(v)),
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

		assert.Equal(t, string(respBytes),
			`{"error":"Unexpected server error"}`)
	}
}

func TestBarExceptionCode(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			bytes, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t,
				[]byte(`{"request":{"stringField":"foo","boolField":true,"binaryField":"AAD//w==","timestamp":"2010-11-04T18:06:40Z","enumField":"APPLE","i64TestField":{"i64Field":123,"timestamp":"2009-11-10T23:00:00Z","longField":{"high":0,"low":1234}}}}`),
				bytes,
			)
			w.WriteHeader(403)
			if _, err := w.Write([]byte(`{"stringField":"foo"}`)); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"POST", "/bar/bar-path", nil,
		bytes.NewReader([]byte(`{
			"request":{"stringField":"foo","boolField":true,"binaryField":"AAD//w==","timestamp":1288894000000,"enumField":0,"i64TestField":{"i64Field":123,"timestamp":1257894000000,"longField":1234}}
		}`)),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "403 Forbidden", res.Status)
	assert.Equal(t, 1, counter)
}

func TestMalformedBarExceptionCode(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
			if _, err := w.Write([]byte("")); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"POST", "/bar/bar-path", nil,
		bytes.NewReader([]byte(`{
			"request":{"stringField":"foo","boolField":true,"binaryField":"aGVsbG8=","timestamp":123,"enumField":0,"i64TestField":{"i64Field":123,"timestamp":1257894000000,"longField":1234}}
		}`)),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "500 Internal Server Error", res.Status)
	assert.Equal(t, 1, counter)
}

func TestBarExceptionInvalidStatusCode(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(402)
			if _, err := w.Write([]byte("{}")); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"POST", "/bar/bar-path", nil,
		bytes.NewReader([]byte(`{
			"request":{"stringField":"foo","boolField":true,"binaryField":"aGVsbG8=","timestamp":123,"enumField":0,"i64TestField":{"i64Field":123,"timestamp":1257894000000,"longField":1234}}
		}`)),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "500 Internal Server Error", res.Status)
	assert.Equal(t, 1, counter)
}
