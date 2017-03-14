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
	"errors"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"

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

func TestAddCredentialsMissingAuthCode(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownBackends: []string{"googleNow"},
		TestBinary: filepath.Join(
			getDirName(), "..", "..", "..",
			"examples", "example-gateway", "main.go",
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
