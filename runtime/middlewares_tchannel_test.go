// Copyright (c) 2018 Uber Technologies, Inc.
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

package zanzibar_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
	"github.com/uber/zanzibar/examples/example-gateway/middlewares/example_tchannel"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/tchannel/baz/baz"
	"github.com/uber/zanzibar/runtime"
)

// Ensures that a middleware stack can correctly return all of its handlers.
func TestTchannelHandlers(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	// fixtures
	reqHeaders := map[string]string{
		"x-token":               "token",
		"x-uuid":                "uuid",
		"x-nil-response-header": "false",
	}
	args := &baz.SimpleService_Call_Args{
		Arg: &baz.BazRequest{
			B1: true,
			S2: "hello",
			I3: 42,
		},
	}
	expectedHeaders := map[string]string{
		"some-res-header": "something",
	}

	ctx := context.Background()
	var result baz.SimpleService_Call_Result
	ms.MockClients().Baz.EXPECT().Call(gomock.Any(), reqHeaders, gomock.Any()).
		Return(map[string]string{"some-res-header": "something"}, nil)

	success, resHeaders, err := ms.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result,
	)
	if !assert.NoError(t, err, "got tchannel error") {
		return
	}
	assert.True(t, success)
	assert.Equal(t, expectedHeaders, resHeaders)

	reqHeaders = map[string]string{
		"x-token":               "token",
		"x-uuid":                "uuid",
		"x-nil-response-header": "true",
	}

	ms.MockClients().Baz.EXPECT().Call(gomock.Any(), reqHeaders, gomock.Any()).
		Return(map[string]string{"some-res-header": "something"}, nil)
	success, resHeaders, err = ms.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result,
	)
	if !assert.Error(t, err, "got tchannel error") {
		return
	}

	assert.False(t, success)
	assert.Nil(t, resHeaders)
}

// Ensures that a middleware can read state from a middeware earlier in the stack.
func TestTchannelMiddlewareSharedStateSet(t *testing.T) {
	ex := exampletchannel.NewMiddleware(
		nil,
		exampletchannel.Options{
			Foo: "test_state",
		},
	)

	exTchannel := exampletchannel.NewMiddleware(
		nil,
		exampletchannel.Options{
			Foo: "foo",
		},
	)

	middles := []zanzibar.MiddlewareTchannelHandle{ex, exTchannel}

	ss := zanzibar.NewTchannelSharedState(middles)

	ss.SetTchannelState(ex, "foo")
	assert.Equal(t, ss.GetTchannelState("example_tchannel").(string), "foo")
}
