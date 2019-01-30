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
	"net/http"
	"testing"

	"github.com/mcuadros/go-jsonschema-generator"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/examples/example-gateway/adapters/example_adapter"
	"github.com/uber/zanzibar/examples/example-gateway/adapters/example_adapter2"
	"github.com/uber/zanzibar/examples/example-gateway/build/adapters/example_adapter/module"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	"github.com/uber/zanzibar/runtime"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
)

// Ensures that an adapter stack can correctly return all of its handlers.
func TestAdapterHandlers(t *testing.T) {
	ex := exampleadapter.NewAdapter(
		&module.Dependencies{
			Default: &zanzibar.DefaultDependencies{},
		},
		exampleadapter.Options{},
	)

	adapterHandles := []zanzibar.AdapterHandle{ex}
	executionStack := zanzibar.NewExecutionStack(adapterHandles, nil, noopHandlerFn)

	// Verify the custom adapter has been added.
	adapters := executionStack.Adapters()
	assert.Equal(t, 1, len(adapters))

	// Run the zanzibar.HandleFn of composed adapters
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}

	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo",
		http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			executionStack.Handle,
		).HandleRequest),
	)
	assert.NoError(t, err)
	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

type countAdapter struct {
	name       string
	reqCounter int
	resCounter int
	reqBail    bool
	resBail    bool
}

func (c *countAdapter) HandleRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
	shared zanzibar.SharedState,
) bool {
	c.reqCounter++
	if !c.reqBail {
		res.WriteJSONBytes(200, nil, []byte(""))
	}

	return !c.reqBail
}

func (c *countAdapter) HandleResponse(
	ctx context.Context,
	res *zanzibar.ServerHTTPResponse,
	shared zanzibar.SharedState,
) {
	c.resCounter++
}

func (c *countAdapter) JSONSchema() *jsonschema.Document {
	return nil
}

func (c *countAdapter) Name() string {
	return c.name
}

// Ensures that an adapter stack can correctly return all of its handlers.
func TestAdapterRequestAbort(t *testing.T) {
	adapter1 := &countAdapter{
		name: "adapter1",
	}
	adapter2 := &countAdapter{
		name:    "adapter2",
		reqBail: true,
	}
	adapter3 := &countAdapter{
		name: "adapter3",
	}

	adapterHandles := []zanzibar.AdapterHandle{adapter1, adapter2, adapter3}
	executionStack := zanzibar.NewExecutionStack(adapterHandles, nil, noopHandlerFn)

	// Verify the custom middleware has been added.
	adapters := executionStack.Adapters()
	assert.Equal(t, 3, len(adapters))

	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}

	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo",
		http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			executionStack.Handle,
		).HandleRequest),
	)
	assert.NoError(t, err)
	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	assert.Equal(t, adapter1.reqCounter, 1)
	assert.Equal(t, adapter1.resCounter, 1)
	assert.Equal(t, adapter2.reqCounter, 1)
	assert.Equal(t, adapter2.resCounter, 1)
	assert.Equal(t, adapter3.reqCounter, 0)
	assert.Equal(t, adapter3.resCounter, 0)
}

// Ensures that an adapter stack can correctly return all of its handlers.
func TestAdapterResponseAbort(t *testing.T) {
	adapter1 := &countAdapter{
		name: "adapter1",
	}
	adapter2 := &countAdapter{
		name: "adapter2",
	}
	adapter3 := &countAdapter{
		name: "adapter3",
	}

	adapterHandles := []zanzibar.AdapterHandle{adapter1, adapter2, adapter3}
	executionStack := zanzibar.NewExecutionStack(adapterHandles, nil, noopHandlerFn)

	// Verify the custom adapter has been added.
	adapters := executionStack.Adapters()
	assert.Equal(t, 3, len(adapters))

	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}

	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo",
		http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			executionStack.Handle,
		).HandleRequest),
	)
	assert.NoError(t, err)
	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	assert.Equal(t, adapter1.reqCounter, 1)
	assert.Equal(t, adapter1.resCounter, 1)
	assert.Equal(t, adapter2.reqCounter, 1)
	assert.Equal(t, adapter2.resCounter, 1)
	assert.Equal(t, adapter3.reqCounter, 1)
	assert.Equal(t, adapter3.resCounter, 1)
}

// Ensures that an adapter can read state from an adapter earlier in the stack.
func TestAdapterSharedStates(t *testing.T) {
	ex := exampleadapter.NewAdapter(
		nil,
		exampleadapter.Options{},
	)
	ex2 := exampleadapter2.NewAdapter(
		nil,
		exampleadapter2.Options{},
	)

	adapterHandles := []zanzibar.AdapterHandle{ex, ex2}
	executionStack := zanzibar.NewExecutionStack(adapterHandles, nil, noopHandlerFn)

	// Verify the custom adapter has been added.
	adapters := executionStack.Adapters()
	assert.Equal(t, 2, len(adapters))

	// Run the zanzibar.HandleFn of composed adapters.
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)
	deps := &zanzibar.DefaultDependencies{
		Scope:         bgateway.ActualGateway.RootScope,
		Logger:        bgateway.ActualGateway.Logger,
		ContextLogger: bgateway.ActualGateway.ContextLogger,
		Tracer:        bgateway.ActualGateway.Tracer,
	}

	err = bgateway.ActualGateway.HTTPRouter.Handle(
		"GET", "/foo",
		http.HandlerFunc(zanzibar.NewRouterEndpoint(
			bgateway.ActualGateway.ContextExtractor,
			deps,
			"foo", "foo",
			executionStack.Handle,
		).HandleRequest),
	)
	assert.NoError(t, err)
	resp, err := gateway.MakeRequest("GET", "/foo", nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

// Ensures that an adapter can read state from an adapter earlier in the stack.
func TestAdapterSharedStateSet(t *testing.T) {
	ex := exampleadapter.NewAdapter(
		nil,
		exampleadapter.Options{},
	)
	ex2 := exampleadapter2.NewAdapter(
		nil,
		exampleadapter2.Options{},
	)

	adapterHandles := []zanzibar.AdapterHandle{ex, ex2}

	ss := zanzibar.NewSharedState(adapterHandles, nil)

	ss.SetAdapterState(ex, "foo")
	assert.Equal(t, ss.GetState("example_adapter").(string), "foo")
}
