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

package bazTchannel

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/test_gateway"

	bazClient "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	clientsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
	endpointsBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/baz_tchannel/baz_tchannel"
)

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)
	return zanzibar.GetDirnameFromRuntimeCaller(file)
}

func TestCallTChannelSuccessfulRequestOKResponse(t *testing.T) {
	testCallCounter := 0

	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"clients.baz.serviceName": "bazService",
	}, &testGateway.Options{
		KnownTChannelBackends: []string{"baz"},
		TestBinary: filepath.Join(
			getDirName(), "..", "..", "..", "examples", "example-gateway",
			"build", "services", "example-gateway", "main.go",
		),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	fakeCall := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBaz.SimpleService_Call_Args,
	) (map[string]string, error) {
		testCallCounter++

		return map[string]string{
			"some-res-header": "something",
		}, nil
	}

	gateway.TChannelBackends()["baz"].Register(
		"SimpleService",
		"Call",
		bazClient.NewSimpleServiceCallHandler(fakeCall),
	)

	ctx := context.Background()
	reqHeaders := map[string]string{
		"x-token": "token",
		"x-uuid":  "uuid",
	}
	args := &endpointsBaz.SimpleService_Call_Args{
		Arg: &endpointsBaz.BazRequest{
			B1: true,
			S2: "hello",
			I3: 42,
		},
	}
	var result endpointsBaz.SimpleService_Call_Result

	success, resHeaders, err := gateway.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result,
	)

	if !assert.NoError(t, err, "got tchannel error") {
		return
	}
	expectedHeaders := map[string]string{
		"some-res-header": "something",
	}
	assert.Equal(t, 1, testCallCounter)
	assert.Equal(t, expectedHeaders, resHeaders)
	assert.True(t, success)

}
