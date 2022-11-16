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

package googlenow_test

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"

	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
)

var benchBytes = []byte("{\"authCode\":\"abcdef\"}")
var noAuthCodeBytes = []byte("{}")
var headers = map[string]string{
	"X-Uuid":  "uuid",
	"X-Token": "token",
}

func BenchmarkGoogleNowAddCredentials(b *testing.B) {
	gateway, err := benchGateway.CreateGateway(
		map[string]interface{}{
			"clients.baz.serviceName": "baz",
		},
		&testGateway.Options{
			KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
			KnownTChannelBackends: []string{"baz"},
			ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		},
		exampleGateway.CreateGateway,
	)
	if err != nil {
		b.Error("got bootstrap err: " + err.Error())
		return
	}

	gateway.HTTPBackends()["google-now"].HandleFunc(
		"POST", "/add-credentials", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(202)
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
				"POST", "/googlenow/add-credentials", headers,
				bytes.NewReader(benchBytes),
			)
			if err != nil {
				b.Error("got http error: " + err.Error())
				break
			}
			if res.Status != "202 Accepted" {
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

func TestAddCredentials(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"google-now"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["google-now"].HandleFunc(
		"POST", "/add-credentials", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(202)
			if _, err := w.Write([]byte("{\"statusCode\":202}")); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials", headers,
		bytes.NewReader(benchBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "202 Accepted", res.Status)
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
			"Could not read request body": true,
		},
		KnownHTTPBackends: []string{"google-now"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["google-now"].HandleFunc(
		"POST", "/add-credentials",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(202)
			if _, err := w.Write([]byte("{\"statusCode\":202}")); err != nil {
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
		"POST", "/googlenow/add-credentials", headers,
		bytes.NewReader([]byte("junk data")),
	)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, 0, counter)

	// give it enough time to pick up the logs
	time.Sleep(20 * time.Millisecond)

	logLines := gateway.Logs("error", "Could not read request body")
	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))

	lineStruct := logLines[0]
	errorField := lineStruct["error"].(string)
	assert.Equal(t, "unexpected EOF", errorField)
}

func TestGoogleNowFailJSONParsing(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"google-now"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["google-now"].HandleFunc(
		"POST", "/add-credentials", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(202)
			if _, err := w.Write([]byte("{\"statusCode\":202}")); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials", headers,
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
		"{\"error\":\"Could not parse json: "+
			"invalid character 'b' looking for beginning of value\"}",
		string(respBytes),
	)
}

// TODO: what this test even do ?
func TestAddCredentialsMissingAuthCode(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"google-now"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})

	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["google-now"].HandleFunc(
		"POST", "/add-credentials", func(w http.ResponseWriter, r *http.Request) {

			bytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal("Cannot read bytes")
			}

			if string(bytes) == `{"authCode":"abcdef"}` {
				w.WriteHeader(202)
				_, err := w.Write([]byte(`{"statusCode":202}`))
				if err != nil {
					t.Fatal("cannot write response")
				}
				counter++
			} else {
				w.WriteHeader(500)
				_, err := w.Write([]byte(`{"statusCode":500}`))
				if err != nil {
					t.Fatal("cannot write response")
				}
				counter++
			}
		},
	)

	res, err := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials", headers,
		bytes.NewReader(noAuthCodeBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "400 Bad Request", res.Status)
	assert.Equal(t, 0, counter)

	res2, err2 := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials", headers,
		bytes.NewReader(benchBytes),
	)
	if !assert.NoError(t, err2, "got http error") {
		return
	}

	assert.Equal(t, "202 Accepted", res2.Status)
	assert.Equal(t, 1, counter)
}

func TestAddCredentialsBackendDown(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"google-now"},
		LogWhitelist: map[string]bool{
			"Could not make http outbound google-now.AddCredentials request": true,
		},
		TestBinary:  util.DefaultMainFile("example-gateway"),
		ConfigFiles: util.DefaultConfigFiles("example-gateway"),
	})

	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	// Close backend
	gateway.HTTPBackends()["google-now"].Close()

	res, err := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials", headers,
		bytes.NewReader(benchBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "500 Internal Server Error", res.Status)

	bytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got bytes read error") {
		return
	}

	assert.Equal(t, string(bytes),
		`{"error":"Unexpected server error"}`)

	time.Sleep(10 * time.Millisecond)

	logLines := gateway.Logs("warn", "Client failure: could not make client request")

	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))

	lineStruct := logLines[0]
	errorMsg := lineStruct["error"].(string)
	assert.Contains(t, errorMsg, "dial tcp")
}

func TestAddCredentialsWrongStatusCode(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"google-now"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})

	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["google-now"].HandleFunc(
		"POST", "/add-credentials", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(201)
			if _, err := w.Write([]byte("{\"statusCode\":201}")); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)
	res, err := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials", headers,
		bytes.NewReader(benchBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "500 Internal Server Error", res.Status)

	bytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got bytes read error") {
		return
	}

	assert.Equal(t, `{"error":"Unexpected server error"}`, string(bytes))
	assert.Equal(t, 1, counter)
}

func TestGoogleNowMissingHeaders(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"google-now"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	res, err := gateway.MakeRequest(
		"POST", "/googlenow/check-credentials", nil,
		bytes.NewReader([]byte("bad bytes")),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "400 Bad Request", res.Status)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t,
		`{"error":"Missing mandatory header: X-Token"}`,
		string(respBytes),
	)
}

func TestAddCredentialsMissingOneHeader(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"google-now"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["google-now"].HandleFunc(
		"POST", "/add-credentials", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(202)
			if _, err := w.Write([]byte("{\"statusCode\":202}")); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"POST",
		"/googlenow/add-credentials",
		map[string]string{
			"x-uuid": "uuid",
		},
		bytes.NewReader(benchBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "400 Bad Request", res.Status)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t,
		`{"error":"Missing mandatory header: X-Token"}`,
		string(respBytes),
	)
}

func TestAddCredentialsHeaderMapping(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"google-now"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["google-now"].HandleFunc(
		"POST", "/add-credentials", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(
				t,
				"uuid",
				r.Header.Get("X-Uuid"))

			// Verify non-proxy headers aren't sent
			assert.Equal(
				t,
				"token",
				r.Header.Get("X-Token"))
			w.Header().Set("X-Uuid", "uuid")

			w.WriteHeader(202)
			if _, err := w.Write([]byte("{\"statusCode\":202}")); err != nil {
				t.Fatal("can't write fake response")
			}
			counter++
		},
	)

	res, err := gateway.MakeRequest(
		"POST",
		"/googlenow/add-credentials",
		headers,
		bytes.NewReader(benchBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "202 Accepted", res.Status)
	assert.Equal(
		t,
		"uuid",
		res.Header.Get("X-Uuid"))

	// Verify non-proxy headers aren't returned
	assert.Equal(
		t,
		"",
		res.Header.Get("X-Token"))

	assert.Equal(t, 1, counter)
}

func TestCheckCredentialsBackendDown(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"google-now"},
		LogWhitelist: map[string]bool{
			"Could not make http outbound google-now.CheckCredentials request": true,
		},
		TestBinary:  util.DefaultMainFile("example-gateway"),
		ConfigFiles: util.DefaultConfigFiles("example-gateway"),
	})

	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	// Close backend
	gateway.HTTPBackends()["google-now"].Close()

	res, err := gateway.MakeRequest(
		"POST", "/googlenow/check-credentials", headers,
		bytes.NewReader(noAuthCodeBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "500 Internal Server Error", res.Status)

	bytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got bytes read error") {
		return
	}

	assert.Equal(t, string(bytes),
		`{"error":"Unexpected server error"}`)

	time.Sleep(10 * time.Millisecond)

	logLines := gateway.Logs("warn", "Client failure: could not make client request")

	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))

	lineStruct := logLines[0]
	errorMsg := lineStruct["error"].(string)
	assert.Contains(t, errorMsg, "dial tcp")
}
