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

// TODO: (lu) to be generated

package baz

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/test/lib/test_gateway"

	bazServer "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
)

var testPingCounter int

func ping(ctx context.Context, reqHeaders map[string]string) (*baz.BazResponse, map[string]string, error) {
	testPingCounter++
	res := baz.BazResponse{
		Message: "pong",
	}
	return &res, nil, nil

}

func TestPingSuccessfulRequestOKResponse(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"clients.baz.serviceName": "Qux",
	}, &testGateway.Options{
		KnownTChannelBackends: []string{"baz"},
		TestBinary: filepath.Join(
			getDirName(), "..", "..", "build", "main.go",
		),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	gateway.TChannelBackends()["baz"].Register(
		"SimpleService",
		"Ping",
		bazServer.NewSimpleServicePingHandler(ping),
	)

	headers := map[string]string{}

	res, err := gateway.MakeRequest(
		"GET",
		"/baz/ping",
		headers,
		bytes.NewReader([]byte(`{}`)),
	)

	if !assert.NoError(t, err, "got http error") {
		return
	}

	defer func() { _ = res.Body.Close() }()
	data, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "failed to read response body") {
		return
	}

	assert.Equal(t, 1, testPingCounter)
	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, `{"message":"pong"}`, string(data))
}
