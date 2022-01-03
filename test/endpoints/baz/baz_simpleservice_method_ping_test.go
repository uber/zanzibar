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

package baz

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"

	"github.com/stretchr/testify/assert"
	bazClient "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBazBase "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/baz/base"

	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
)

var testConfig = map[string]interface{}{
	"clients.baz.serviceName": "bazService",
}
var testOptions = &testGateway.Options{
	KnownTChannelBackends: []string{"baz"},
	TestBinary:            util.DefaultMainFile("example-gateway"),
	ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
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
			ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		},
		exampleGateway.CreateGateway,
	)
	if err != nil {
		b.Error("got bootstrap err: " + err.Error())
		return
	}

	err = gateway.TChannelBackends()["baz"].Register(
		"baz", "ping", "SimpleService::ping",
		bazClient.NewSimpleServicePingHandler(ping),
	)
	if err != nil {
		b.Error("got register err: " + err.Error())
		return
	}

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

	err = gateway.TChannelBackends()["baz"].Register(
		"baz", "ping", "SimpleService::ping",
		bazClient.NewSimpleServicePingHandler(fakePing),
	)
	assert.NoError(t, err)

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
	gateway, err := testGateway.CreateGateway(t, testConfig, &testGateway.Options{
		KnownTChannelBackends: []string{"baz"},
		TestBinary:            util.DefaultMainFile("example-gateway"),
		ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
	})
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

	err = gateway.TChannelBackends()["baz"].Register(
		"baz", "ping", "SimpleService::ping",
		bazClient.NewSimpleServicePingHandler(fakePing),
	)
	assert.NoError(t, err)

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

	assert.Len(t, gateway.Logs("info", "Started Example-gateway"), 1)
	assert.Len(t, gateway.Logs("info", "Failed after non-retriable error."), 1)
	assert.Len(t, gateway.Logs("warn", "Client failure: TChannel client call returned error"), 1)
	assert.Len(t, gateway.Logs("warn", "Finished an incoming server HTTP request with 500 status code"), 1)
	assert.Len(t, gateway.Logs("warn", "Failed to send outgoing client TChannel request"), 1)
	assert.Len(t, gateway.Logs("warn", "Client failure: could not make client request"), 1)

	logLines := gateway.Logs("warn", "Client failure: could not make client request")
	assert.Equal(t, 1, len(logLines))

	logObj := logLines[0]
	assert.Equal(
		t,
		"tchannel error ErrCodeUnexpected: Server Error",
		logObj["error"].(string),
	)
}
