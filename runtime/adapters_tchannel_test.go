// Copyright (c) 2019 Uber Technologies, Inc.
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
	"github.com/mcuadros/go-jsonschema-generator"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/examples/example-gateway/adapters/example_adapter_tchannel"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/tchannel/baz/baz"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
	"github.com/uber/zanzibar/runtime"
	"go.uber.org/thriftrw/wire"
)

// Ensures that an adapter stack can correctly return all of its handlers.
func TestTchannelAdapterHandlers(t *testing.T) {
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

// Ensures that a tchannel adapter can read state from a tchannel adapter earlier in the stack.
func TestTchannelAdapterSharedStateSet(t *testing.T) {
	ex := exampleadaptertchannel.NewAdapter(
		nil,
		exampleadaptertchannel.Options{},
	)

	ex2 := exampleadaptertchannel.NewAdapter(
		nil,
		exampleadaptertchannel.Options{},
	)

	adapterHandles := []zanzibar.AdapterTchannelHandle{ex, ex2}
	executionStack := zanzibar.NewExecutionTchannelStack(adapterHandles, nil, nil)
	adapters := executionStack.TchannelAdapters()
	assert.Equal(t, 2, len(adapters))
	ss := zanzibar.NewTchannelSharedState(adapters, nil)
	ss.SetTchannelAdapterState(ex, "foo")
	assert.Equal(t, ss.GetTchannelState("example_adapter_tchannel").(string), "foo")
}

type countTchannelAdapter struct {
	name       string
	reqCounter int
	resCounter int
	reqBail    bool
}

type mockTchannelAdapterHandler struct {
	name string
}

func (c *countTchannelAdapter) HandleRequest(
	ctx context.Context,
	reqHeaders map[string]string,
	wireValue *wire.Value,
	shared zanzibar.TchannelSharedState,
) (bool, error) {
	c.reqCounter++
	return c.reqBail, nil
}

func (c *countTchannelAdapter) HandleResponse(
	ctx context.Context,
	rwt zanzibar.RWTStruct,
	shared zanzibar.TchannelSharedState,
) zanzibar.RWTStruct {
	c.resCounter++
	return rwt
}

func (c *countTchannelAdapter) JSONSchema() *jsonschema.Document {
	return nil
}

func (c *countTchannelAdapter) Name() string {
	return c.name
}

func (c *mockTchannelAdapterHandler) Handle(ctx context.Context,
	reqHeaders map[string]string,
	wireValue *wire.Value,
) (success bool, resp zanzibar.RWTStruct, respHeaders map[string]string, err error) {
	return true, nil, map[string]string{}, err
}

// Ensures that a tchannel adapter stack can correctly return all of its handlers.
func TestTchannelAdapterRequestAbort(t *testing.T) {
	adapter1 := &countTchannelAdapter{
		name: "adapter1",
	}
	adapter2 := &countTchannelAdapter{
		name:    "adapter2",
		reqBail: true,
	}

	mockTHandler := &mockTchannelAdapterHandler{
		name: "mockTHandler",
	}

	adapterHandles := []zanzibar.AdapterTchannelHandle{adapter1, adapter2}
	executionStack := zanzibar.NewExecutionTchannelStack(adapterHandles, nil, mockTHandler)
	_, _, _, err := executionStack.Handle(context.Background(), map[string]string{}, nil)
	assert.NoError(t, err)

	assert.Equal(t, adapter1.reqCounter, 1)
	assert.Equal(t, adapter1.resCounter, 1)
	assert.Equal(t, adapter2.reqCounter, 0)
	assert.Equal(t, adapter2.resCounter, 0)
}
