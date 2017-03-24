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

package google_now_test

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"encoding/json"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
	"github.com/uber/zanzibar/test/lib/test_gateway"
)

var benchBytes = []byte("{\"authCode\":\"abcdef\"}")
var noAuthCodeBytes = []byte("{}")

func BenchmarkGoogleNowAddCredentials(b *testing.B) {
	gateway, err := benchGateway.CreateGateway(nil, &testGateway.Options{
		KnownBackends: []string{"googleNow"},
	})
	if err != nil {
		b.Error("got bootstrap err: " + err.Error())
		return
	}

	gateway.Backends()["googleNow"].HandleFunc(
		"POST", "/add-credentials", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte("{\"statusCode\":200}")); err != nil {
				panic(errors.New("can't write fake response"))
			}
		},
	)

	b.ResetTimer()

	// b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := gateway.MakeRequest(
				"POST", "/googlenow/add-credentials",
				bytes.NewReader(benchBytes),
			)
			if err != nil {
				b.Error("got http error: " + err.Error())
				break
			}
			if res.Status != "200 OK" {
				b.Error("got bad status error: " + res.Status)
				break
			}

			_, err = ioutil.ReadAll(res.Body)
			if err != nil {
				b.Error("could not write response: " + res.Status)
				break
			}
			_ = res.Body.Close()
		}
	})

	b.StopTimer()
	gateway.Close()
	b.StartTimer()
}

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(file)
}

func TestAddCredentials(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownBackends: []string{"googleNow"},
		TestBinary: filepath.Join(
			getDirName(), "..", "..", "..",
			"examples", "example-gateway", "build", "main.go",
		),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.Backends()["googleNow"].HandleFunc(
		"POST", "/add-credentials", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte("{\"statusCode\":200}")); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials", bytes.NewReader(benchBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, counter)
}

type failingConn struct {
	net.Conn
}

func (conn *failingConn) Write(b []byte) (int, error) {
	b = b[0 : len(b)-1]
	n, err := conn.Conn.Write(b)

	_ = conn.Conn.Close()
	return n, err
}

func TestGoogleNowFailReadAllCall(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		LogWhitelist: map[string]bool{
			"Could not ReadAll() body": true,
		},
		KnownBackends: []string{"googleNow"},
		TestBinary: filepath.Join(
			getDirName(), "..", "..", "..",
			"examples", "example-gateway", "build", "main.go",
		),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.Backends()["googleNow"].HandleFunc(
		"POST", "/add-credentials",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte("{\"statusCode\":200}")); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	cgateway := gateway.(*testGateway.ChildProcessGateway)
	cgateway.HTTPClient.Transport = &http.Transport{
		DisableKeepAlives:   false,
		MaxIdleConns:        500,
		MaxIdleConnsPerHost: 500,
		Dial: func(network string, addr string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}

			conn, err := dialer.Dial(network, addr)
			if err != nil {
				return nil, err
			}

			return &failingConn{conn}, nil
		},
	}

	res, err := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials",
		bytes.NewReader([]byte("junk data")),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "short write")
	assert.Nil(t, res)
	assert.Equal(t, 0, counter)

	time.Sleep(10 * time.Millisecond)

	errLogs := gateway.GetErrorLogs()

	logLines := errLogs["Could not ReadAll() body"]
	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))

	line := logLines[0]
	lineStruct := map[string]interface{}{}
	jsonErr := json.Unmarshal([]byte(line), &lineStruct)
	if !assert.NoError(t, jsonErr, "cannot decode json lines") {
		return
	}

	errorField := lineStruct["error"].(string)
	assert.Equal(t, "unexpected EOF", errorField)
}

func TestGoogleNowFailJSONParsing(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownBackends: []string{"googleNow"},
		TestBinary: filepath.Join(
			getDirName(), "..", "..", "..",
			"examples", "example-gateway", "build", "main.go",
		),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.Backends()["googleNow"].HandleFunc(
		"POST", "/add-credentials", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			if _, err := w.Write([]byte("{\"statusCode\":200}")); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials",
		bytes.NewReader([]byte("bad bytes")),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "400 Bad Request", res.Status)
	assert.Equal(t, 0, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t,
		"Could not parse json: parse error: "+
			"invalid character 'b' after top-level value "+
			"near offset 0 of 'bad bytes'",
		string(respBytes),
	)
}

func TestAddCredentialsMissingAuthCode(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownBackends: []string{"googleNow"},
		TestBinary: filepath.Join(
			getDirName(), "..", "..", "..",
			"examples", "example-gateway", "build", "main.go",
		),
	})

	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.Backends()["googleNow"].HandleFunc(
		"POST", "/add-credentials", func(w http.ResponseWriter, r *http.Request) {
			if r.FormValue("authCode") != "" {
				if _, err := w.Write([]byte("{\"statusCode\":200}")); err != nil {
					t.Fatal("can't write fake response")
				}
				counter++
			} else {
				if _, err := w.Write([]byte("{\"statusCode\":500}")); err != nil {
					t.Fatal("can't write fake response")
				}
			}
		},
	)

	res, err := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials", bytes.NewReader(noAuthCodeBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 0, counter)
}

func TestAddCredentialsBackendDown(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownBackends: []string{"googleNow"},
		LogWhitelist: map[string]bool{
			"Could not make client request": true,
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

	// Close backend
	gateway.Backends()["googleNow"].Close()

	res, err := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials", bytes.NewReader(noAuthCodeBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "500 Internal Server Error", res.Status)

	bytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got bytes read error") {
		return
	}

	assert.Contains(t, string(bytes), "could not make client request")

	errorLogs := gateway.GetErrorLogs()
	logLines := errorLogs["Could not make client request"]

	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))

	line := logLines[0]
	lineStruct := map[string]interface{}{}
	jsonErr := json.Unmarshal([]byte(line), &lineStruct)
	if !assert.NoError(t, jsonErr, "cannot decode json lines") {
		return
	}

	errorMsg := lineStruct["error"].(string)
	assert.Contains(t, errorMsg, "dial tcp")
}

func TestAddCredentialsWrongStatusCode(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownBackends: []string{"googleNow"},
		TestBinary: filepath.Join(
			getDirName(), "..", "..", "..",
			"examples", "example-gateway", "build", "main.go",
		),
	})

	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.Backends()["googleNow"].HandleFunc(
		"POST", "/add-credentials", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(201)
			if _, err := w.Write([]byte("{\"statusCode\":201}")); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)
	res, err := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials", bytes.NewReader(noAuthCodeBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "201 Created", res.Status)

	bytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got bytes read error") {
		return
	}

	assert.Equal(t, "", string(bytes))
	assert.Equal(t, 1, counter)
}
