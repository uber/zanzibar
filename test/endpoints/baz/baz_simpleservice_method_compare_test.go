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
	"errors"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
	"github.com/uber/zanzibar/test/lib/test_gateway"

	bazClient "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBazBase "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/base"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
)

var testCompareCounter int

func compare(
	ctx context.Context, reqHeaders map[string]string, args *baz.SimpleService_Compare_Args,
) (*clientsBazBase.BazResponse, map[string]string, error) {
	testCompareCounter++
	r1 := args.Arg1
	r2 := args.Arg2
	if r1.B1 && r1.S2 == "hello" && r1.I3 == 42 && r2.B1 && r2.S2 == "hola" && r2.I3 == 42 {
		return &clientsBazBase.BazResponse{
			Message: "different",
		}, nil, nil
	}
	return nil, nil, errors.New("Wrong Args")

}

func BenchmarkCompare(b *testing.B) {
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
		"compare",
		bazClient.NewSimpleServiceCompareHandler(compare),
	)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := gateway.MakeRequest(
				"POST", "/baz/compare", nil,
				bytes.NewReader([]byte(`{"arg1":{"b1":true,"s2":"hello","i3":42},"arg2":{"b1":true,"s2":"hola","i3":42}}`)),
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

func TestCompare(t *testing.T) {
	gateway, err := testGateway.CreateGateway(t, testConfig, testOptions)
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	callCounter := 0

	fakeCompare := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *baz.SimpleService_Compare_Args,
	) (*clientsBazBase.BazResponse, map[string]string, error) {
		callCounter++

		return &clientsBazBase.BazResponse{
			Message: "a message",
		}, nil, nil
	}

	gateway.TChannelBackends()["baz"].Register(
		"SimpleService",
		"compare",
		bazClient.NewSimpleServiceCompareHandler(fakeCompare),
	)

	res, err := gateway.MakeRequest(
		"POST", "/baz/compare", nil,
		bytes.NewBuffer([]byte(`{
			"arg1":{ "b1":true,"s2":"a","i3":1 },
			"arg2":{ "b1":true,"s2":"a","i3":1 }
		}`)),
	)

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

func TestCompareInvalidArgs(t *testing.T) {
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
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	callCounter := 0

	fakeCompare := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *baz.SimpleService_Compare_Args,
	) (*clientsBazBase.BazResponse, map[string]string, error) {
		callCounter++

		return &clientsBazBase.BazResponse{
			Message: "a message",
		}, nil, nil
	}

	gateway.TChannelBackends()["baz"].Register(
		"SimpleService",
		"compare",
		bazClient.NewSimpleServiceCompareHandler(fakeCompare),
	)

	bgateway := gateway.(*benchGateway.BenchGateway)
	bazClient := bazClient.NewClient(bgateway.ActualGateway)

	res, _, err := bazClient.Compare(
		context.Background(),
		nil,
		&baz.SimpleService_Compare_Args{},
	)

	assert.NotNil(t, err)
	assert.Contains(t,
		err.Error(),
		"field Arg1 of SimpleService_Compare_Args is required",
	)

	assert.Nil(t, res)
}
