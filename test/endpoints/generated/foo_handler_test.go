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

package generated_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	assert "github.com/stretchr/testify/assert"
	config "github.com/uber/zanzibar/examples/example-gateway/config"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testBackend "github.com/uber/zanzibar/test/lib/test_backend"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
)

var benchBytes = []byte("{\"authCode\":\"abcdef\"}")

type testCase struct {
	Counter      int
	IsBench      bool
	Backend      *testBackend.TestBackend
	TestGateway  *testGateway.TestGateway
	BenchGateway *benchGateway.BenchGateway
}

func (testCase *testCase) Close() {
	testCase.Backend.Close()

	if testCase.IsBench {
		testCase.BenchGateway.Close()
	} else {
		testCase.TestGateway.Close()
	}
}

func newTestCase(t *testing.T, isBench bool) (*testCase, error) {
	testCase := &testCase{
		IsBench: isBench,
	}

	testCase.Backend = testBackend.CreateBackend(0)
	err := testCase.Backend.Bootstrap()
	if err != nil {
		return nil, err
	}

	addCredentials := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte("{\"statusCode\":200}")); err != nil {
			t.Fatal("can't write fake response")
		}
		testCase.Counter++
	}
	testCase.Backend.HandleFunc("POST", "/add-credentials", addCredentials)

	config := &config.Config{}
	config.Clients.GoogleNow.IP = "127.0.0.1"
	config.Clients.GoogleNow.Port = testCase.Backend.RealPort

	if testCase.IsBench {
		gateway, err := benchGateway.CreateGateway(config)
		if err != nil {
			return nil, err
		}
		testCase.BenchGateway = gateway
	} else {
		gateway, err := testGateway.CreateGateway(t, config, nil)
		if err != nil {
			return nil, err
		}
		testCase.TestGateway = gateway
	}
	return testCase, nil
}

func BenchmarkRtnowAddCredentials(b *testing.B) {
	testCase, err := newTestCase(nil, true)
	if err != nil {
		b.Error("got bootstrap err: " + err.Error())
		return
	}

	b.ResetTimer()

	// b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := testCase.BenchGateway.MakeRequest(
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
		}
	})

	b.StopTimer()
	testCase.Close()
	b.StartTimer()
}

func TestAddCredentials(t *testing.T) {
	testCase, err := newTestCase(t, false)
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer testCase.Close()

	assert.NotNil(t, testCase.TestGateway, "gateway exists")

	res, err := testCase.TestGateway.MakeRequest(
		"POST", "/googlenow/add-credentials", bytes.NewReader(benchBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, testCase.Counter)
}
