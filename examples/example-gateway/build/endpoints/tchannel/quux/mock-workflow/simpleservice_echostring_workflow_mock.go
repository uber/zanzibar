// Code generated by zanzibar
// @generated

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

package mocktchannelquuxworkflow

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/uber-go/tally"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"

	bazclientgeneratedmock "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz/mock-client"
	quuxclientgeneratedmock "github.com/uber/zanzibar/examples/example-gateway/build/clients/quux/mock-client"
	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/tchannel/quux/module"
	workflow "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/tchannel/quux/workflow"
	defaultexamplemiddlewaregenerated "github.com/uber/zanzibar/examples/example-gateway/build/middlewares/default/default_example"
	defaultexamplemiddlewaremodule "github.com/uber/zanzibar/examples/example-gateway/build/middlewares/default/default_example/module"
	defaultexample2middlewaregenerated "github.com/uber/zanzibar/examples/example-gateway/build/middlewares/default/default_example2"
	defaultexample2middlewaremodule "github.com/uber/zanzibar/examples/example-gateway/build/middlewares/default/default_example2/module"
	defaultexampletchannelmiddlewaregenerated "github.com/uber/zanzibar/examples/example-gateway/build/middlewares/default/default_example_tchannel"
	defaultexampletchannelmiddlewaremodule "github.com/uber/zanzibar/examples/example-gateway/build/middlewares/default/default_example_tchannel/module"
	fixturequuxclientstatic "github.com/uber/zanzibar/examples/example-gateway/clients/quux/fixture"
	quuxendpointstatic "github.com/uber/zanzibar/examples/example-gateway/endpoints/tchannel/quux"
)

// NewSimpleServiceEchoStringWorkflowMock creates a workflow with mock clients
func NewSimpleServiceEchoStringWorkflowMock(t *testing.T) (workflow.SimpleServiceEchoStringWorkflow, *MockClientNodes) {
	ctrl := gomock.NewController(t)

	initializedDefaultDependencies := &zanzibar.DefaultDependencies{
		Logger: zap.NewNop(),
		Scope:  tally.NewTestScope("", make(map[string]string)),
	}
	initializedDefaultDependencies.ContextLogger = zanzibar.NewContextLogger(initializedDefaultDependencies.Logger)
	initializedDefaultDependencies.ContextExtractor = &zanzibar.ContextExtractors{}

	initializedClientDependencies := &clientDependenciesNodes{}
	mockClientNodes := &MockClientNodes{
		Baz:  bazclientgeneratedmock.NewMockClient(ctrl),
		Quux: quuxclientgeneratedmock.New(ctrl, fixturequuxclientstatic.Fixture),
	}
	initializedClientDependencies.Baz = mockClientNodes.Baz
	initializedClientDependencies.Quux = mockClientNodes.Quux

	initializedMiddlewareDependencies := &middlewareDependenciesNodes{}

	initializedMiddlewareDependencies.DefaultExample = defaultexamplemiddlewaregenerated.NewMiddleware(&defaultexamplemiddlewaremodule.Dependencies{
		Default: initializedDefaultDependencies,
		Client: &defaultexamplemiddlewaremodule.ClientDependencies{
			Baz: initializedClientDependencies.Baz,
		},
	})
	initializedMiddlewareDependencies.DefaultExample2 = defaultexample2middlewaregenerated.NewMiddleware(&defaultexample2middlewaremodule.Dependencies{
		Default: initializedDefaultDependencies,
		Client: &defaultexample2middlewaremodule.ClientDependencies{
			Baz: initializedClientDependencies.Baz,
		},
	})
	initializedMiddlewareDependencies.DefaultExampleTchannel = defaultexampletchannelmiddlewaregenerated.NewMiddleware(&defaultexampletchannelmiddlewaremodule.Dependencies{
		Default: initializedDefaultDependencies,
	})

	w := quuxendpointstatic.NewSimpleServiceEchoStringWorkflow(
		&module.Dependencies{
			Default: initializedDefaultDependencies,
			Client: &module.ClientDependencies{
				Quux: initializedClientDependencies.Quux,
			},
			Middleware: &module.MiddlewareDependencies{
				DefaultExample:         initializedMiddlewareDependencies.DefaultExample,
				DefaultExample2:        initializedMiddlewareDependencies.DefaultExample2,
				DefaultExampleTchannel: initializedMiddlewareDependencies.DefaultExampleTchannel,
			},
		},
	)

	return w, mockClientNodes
}
