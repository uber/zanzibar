// Copyright (c) 2020 Uber Technologies, Inc.
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

package trafficshadow_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

func TestTrafficShadowHelloForClient1(t *testing.T) {
	 testTrafficShadowHello(t, false)
}

func TestTrafficShadowHelloForClient2(t *testing.T) {
	testTrafficShadowHello(t, true)
}

func testTrafficShadowHello(t *testing.T, callShadowClient bool) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"trafficshadowclient1", "trafficshadowclient2"},
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	if callShadowClient {
		gateway.HTTPBackends()["trafficshadowclient2"].HandleFunc(
			"GET", "/trafficshadow/hello",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				if _, err := w.Write([]byte(`{
				"resField2": "ashish2"
			}`)); err != nil {
					t.Fatal("can't write fake response")
				}
				counter++
			},
		)
	} else {
		gateway.HTTPBackends()["trafficshadowclient1"].HandleFunc(
			"GET", "/trafficshadow/hello",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				if _, err := w.Write([]byte(`{
				"resField1": "ashish1"
			}`)); err != nil {
					t.Fatal("can't write fake response")
				}
				counter++
			},
		)
	}

	var res *http.Response
	if callShadowClient {
		res, _ = gateway.MakeRequest(
			"GET",
			"/trafficshadow/hello",
			map[string]string{
				"X-Uber-Shadow-Client": "sample_val",
			}, nil,
		)
	} else {
		res, _ = gateway.MakeRequest(
			"GET",
			"/trafficshadow/hello",
			nil, nil,
		)
	}

	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	if callShadowClient {
		assert.Equal(t, `{"resField":"ashish2"}`, string(respBytes))
	} else {
		assert.Equal(t, `{"resField":"ashish1"}`, string(respBytes))
	}
}
