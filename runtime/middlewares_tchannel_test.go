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

package zanzibar_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/thriftrw/protocol/stream"

	"github.com/golang/mock/gomock"
	"github.com/mcuadros/go-jsonschema-generator"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/tchannel/baz/baz"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
	exampletchannel "github.com/uber/zanzibar/examples/example-gateway/middlewares/example_tchannel"
	zanzibar "github.com/uber/zanzibar/runtime"
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

	expectedClientReqHeaders := map[string]string{
		"x-uuid":                "uuid",
		"x-nil-response-header": "false",
	}

	ctx := context.Background()
	var result baz.SimpleService_Call_Result
	ms.MockClients().Baz.EXPECT().Call(gomock.Any(), expectedClientReqHeaders, gomock.Any(), gomock.Any()).
		Return(ctx, map[string]string{"some-res-header": "something"}, nil)

	success, resHeaders, err := ms.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	dynamicRespHeaders := []string{
		"client.response.duration",
	}
	for _, dynamicValue := range dynamicRespHeaders {
		assert.Contains(t, resHeaders, dynamicValue)
		delete(resHeaders, dynamicValue)
	}

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

	expectedClientReqHeaders = map[string]string{
		"x-uuid":                "uuid",
		"x-nil-response-header": "true",
	}

	ms.MockClients().Baz.EXPECT().Call(gomock.Any(), expectedClientReqHeaders, gomock.Any(), gomock.Any()).
		Return(ctx, map[string]string{"some-res-header": "something"}, nil)
	success, _, err = ms.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	if !assert.Error(t, err, "got tchannel error") {
		return
	}

	assert.False(t, success)
}

// Ensures that a tchannel middleware can read state from a tchannel middeware earlier in the stack.
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
	middlewareStack := zanzibar.NewTchannelStack(middles, nil)
	middlewares := middlewareStack.TchannelMiddlewares()
	assert.Equal(t, 2, len(middlewares))
	ss := zanzibar.NewTchannelSharedState(middles)
	ss.SetTchannelState(ex, "foo")
	assert.Equal(t, ss.GetTchannelState("example_tchannel").(string), "foo")
}

type countTchannelMiddleware struct {
	name       string
	reqCounter int
	resCounter int
	reqBail    bool
}

type mockTchannelHandler struct {
	name string
}

func (c *countTchannelMiddleware) HandleRequest(
	ctx context.Context,
	reqHeaders map[string]string,
	sr stream.Reader,
	shared zanzibar.TchannelSharedState,
) (context.Context, bool, error) {
	c.reqCounter++
	return ctx, c.reqBail, nil
}

func (c *countTchannelMiddleware) HandleResponse(
	ctx context.Context,
	rwt zanzibar.RWTStruct,
	shared zanzibar.TchannelSharedState,
) zanzibar.RWTStruct {
	c.resCounter++
	return rwt
}

func (c *countTchannelMiddleware) JSONSchema() *jsonschema.Document {
	return nil
}

func (c *countTchannelMiddleware) Name() string {
	return c.name
}

func (c *mockTchannelHandler) Handle(ctx context.Context, reqHeaders map[string]string, sr stream.Reader) (ctx_resp context.Context, success bool, resp zanzibar.RWTStruct, respHeaders map[string]string, err error) {
	return ctx, true, nil, map[string]string{}, err
}

// Ensures that a tchannel middleware stack can correctly return all of its handlers.
func TestTchannelMiddlewareRequestAbort(t *testing.T) {
	mid1 := &countTchannelMiddleware{
		name: "mid1",
	}
	mid2 := &countTchannelMiddleware{
		name:    "mid2",
		reqBail: true,
	}

	mockTHandler := &mockTchannelHandler{
		name: "mockTHandler",
	}

	middles := []zanzibar.MiddlewareTchannelHandle{mid1, mid2}
	middlewareStack := zanzibar.NewTchannelStack(middles, mockTHandler)
	_, _, _, _, err := middlewareStack.Handle(context.Background(), map[string]string{}, nil)
	assert.NoError(t, err)

	assert.Equal(t, mid1.reqCounter, 1)
	assert.Equal(t, mid1.resCounter, 0)
	assert.Equal(t, mid2.reqCounter, 0)
	assert.Equal(t, mid2.resCounter, 0)
}
