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

package baz

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/uber/zanzibar/test/lib/bench_gateway"
	"github.com/uber/zanzibar/test/lib/test_gateway"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBazBase "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/base"
)

var testConfig = map[string]interface{}{
	"clients.baz.serviceName": "bazService",
}
var testOptions = &testGateway.Options{
	KnownTChannelBackends: []string{"baz"},
	TestBinary: filepath.Join(
		getDirName(), "..", "..", "..", "examples", "example-gateway",
		"build", "services", "example-gateway", "main.go",
	),
}

var testPingCounter int

func ping(
	ctx context.Context,
	reqHeaders map[string]string,
) (*clientsBazBase.BazResponse, map[string]string, error) {
	testPingCounter++
	res := clientsBazBase.BazResponse{
		Message: "pong",
	}
	return &res, nil, nil

}

func BenchmarkPing(b *testing.B) {
	gateway, err := benchGateway.CreateGateway(
		map[string]interface{}{
			"clients.baz.serviceName": "baz",
		},
		&testGateway.Options{
			KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
			KnownTChannelBackends: []string{"baz"},
		},
		nil,
	)
	if err != nil {
		b.Error("got bootstrap err: " + err.Error())
		return
	}

	gateway.TChannelBackends()["baz"].Register(
		"SimpleService",
		"ping",
		bazClient.NewSimpleServicePingHandler(ping),
	)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := gateway.MakeRequest(
				"GET", "/baz/ping", nil,
				bytes.NewReader([]byte(`{}`)),
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
				b.Error("could not read response: " + res.Status)
				break
			}
			_ = res.Body.Close()
		}
	})

	b.StopTimer()
	gateway.Close()
}

func TestPing(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, testConfig, testOptions)
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	callCounter := 0

	fakePing := func(
		ctx context.Context,
		reqHeaders map[string]string,
	) (*clientsBazBase.BazResponse, map[string]string, error) {
		callCounter++

		return &clientsBazBase.BazResponse{
			Message: "a message",
		}, nil, nil
	}

	gateway.TChannelBackends()["baz"].Register(
		"SimpleService",
		"ping",
		bazClient.NewSimpleServicePingHandler(fakePing),
	)

	res, err := gateway.MakeRequest("GET", "/baz/ping", nil, nil)

	if !assert.NoError(t, err, "got request error") {
		return
	}

	assert.Equal(t, 1, callCounter)
	assert.Equal(t, 200, res.StatusCode)

	bytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got read error") {
		return
	}

	assert.Equal(t, `{"message":"a message"}`, string(bytes))
}

func TestPingWithInvalidResponse(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, testConfig, testOptions)
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	callCounter := 0

	fakePing := func(
		ctx context.Context,
		reqHeaders map[string]string,
	) (*clientsBazBase.BazResponse, map[string]string, error) {
		callCounter++

		return nil, nil, nil
	}

	gateway.TChannelBackends()["baz"].Register(
		"SimpleService",
		"ping",
		bazClient.NewSimpleServicePingHandler(fakePing),
	)

	res, err := gateway.MakeRequest("GET", "/baz/ping", nil, nil)

	if !assert.NoError(t, err, "got request error") {
		return
	}

	assert.Equal(t, 1, callCounter)
	assert.Equal(t, 500, res.StatusCode)

	bytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got read error") {
		return
	}

	assert.Equal(t, `{"error":"Unexpected server error"}`, string(bytes))

	allLogs := gateway.AllLogs()

	assert.Equal(t, 8, len(allLogs))
	assert.Equal(t, 1, len(allLogs["Finished a downstream TChannel request"]))
	assert.Equal(t, 1, len(allLogs["Started ExampleGateway"]))
	assert.Equal(t, 1, len(allLogs["Outbound connection is active."]))
	assert.Equal(t, 1, len(allLogs["Failed after non-retriable error."]))
	assert.Equal(t, 1, len(allLogs["Could not make client request"]))
	assert.Equal(t, 1, len(allLogs["Workflow for endpoint returned error"]))
	assert.Equal(t, 1, len(allLogs["Sending error for endpoint request"]))
	assert.Equal(t, 1, len(allLogs["Finished an incoming server HTTP request"]))

	logLines := gateway.Logs("warn", "Could not make client request")
	assert.Equal(t, 1, len(logLines))

	logObj := logLines[0]
	assert.Equal(
		t,
		"tchannel error ErrCodeUnexpected: Server Error",
		logObj["error"].(string),
	)
}
