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

	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	bazServer "github.com/uber/zanzibar/examples/example-gateway/clients/baz"
)

var testCallCounter int

func call(
	ctx context.Context, reqHeaders map[string]string, args *baz.SimpleService_Call_Args,
) (*baz.BazResponse, map[string]string, error) {
	testCallCounter++
	var resp *baz.BazResponse
	r := args.Arg
	if r.B1 && r.S2 == "hello" && r.I3 == 42 {
		resp = &baz.BazResponse{
			Message: "yo",
		}
	}
	return resp, nil, nil

}

func TestCallSuccessfulRequestOKResponse(t *testing.T) {
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

	gateway.TChannelBackends()["baz"].Register(bazServer.WithCall(call))

	headers := map[string]string{}

	res, err := gateway.MakeRequest(
		"POST",
		"/baz/call-path",
		headers,
		bytes.NewReader([]byte(`{"arg":{"b1":true,"s2":"hello","i3":42}}`)),
	)

	if !assert.NoError(t, err, "got http error") {
		return
	}

	defer func() { _ = res.Body.Close() }()
	data, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "failed to read response body") {
		return
	}

	assert.Equal(t, 1, testCallCounter)
	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, `{"message":"yo"}`, string(data))
}
