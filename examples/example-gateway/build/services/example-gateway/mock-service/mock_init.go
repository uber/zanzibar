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

package examplegatewayservicegeneratedmock

import (
	"github.com/golang/mock/gomock"
	module "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/module"
	zanzibar "github.com/uber/zanzibar/runtime"

	exampleadapteradaptergenerated "github.com/uber/zanzibar/examples/example-gateway/build/adapters/example_adapter"
	exampleadapteradaptermodule "github.com/uber/zanzibar/examples/example-gateway/build/adapters/example_adapter/module"
	exampleadapter2adaptergenerated "github.com/uber/zanzibar/examples/example-gateway/build/adapters/example_adapter2"
	exampleadapter2adaptermodule "github.com/uber/zanzibar/examples/example-gateway/build/adapters/example_adapter2/module"
	exampleadaptertchanneladaptergenerated "github.com/uber/zanzibar/examples/example-gateway/build/adapters/example_adapter_tchannel"
	exampleadaptertchanneladaptermodule "github.com/uber/zanzibar/examples/example-gateway/build/adapters/example_adapter_tchannel/module"
	barclientgenerated "github.com/uber/zanzibar/examples/example-gateway/build/clients/bar/mock-client"
	bazclientgenerated "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz/mock-client"
	contactsclientgenerated "github.com/uber/zanzibar/examples/example-gateway/build/clients/contacts/mock-client"
	googlenowclientgenerated "github.com/uber/zanzibar/examples/example-gateway/build/clients/google-now/mock-client"
	multiclientgenerated "github.com/uber/zanzibar/examples/example-gateway/build/clients/multi/mock-client"
	quuxclientgenerated "github.com/uber/zanzibar/examples/example-gateway/build/clients/quux/mock-client"
	barendpointgenerated "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/bar"
	barendpointmodule "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/bar/module"
	bazendpointgenerated "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/baz"
	bazendpointmodule "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/baz/module"
	contactsendpointgenerated "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/contacts"
	contactsendpointmodule "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/contacts/module"
	googlenowendpointgenerated "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/googlenow"
	googlenowendpointmodule "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/googlenow/module"
	multiendpointgenerated "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/multi"
	multiendpointmodule "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/multi/module"
	panicendpointgenerated "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/panic"
	panicendpointmodule "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/panic/module"
	baztchannelendpointgenerated "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/tchannel/baz"
	baztchannelendpointmodule "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/tchannel/baz/module"
	panictchannelendpointgenerated "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/tchannel/panic"
	panictchannelendpointmodule "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/tchannel/panic/module"
	examplemiddlewaregenerated "github.com/uber/zanzibar/examples/example-gateway/build/middlewares/example"
	examplemiddlewaremodule "github.com/uber/zanzibar/examples/example-gateway/build/middlewares/example/module"
	exampletchannelmiddlewaregenerated "github.com/uber/zanzibar/examples/example-gateway/build/middlewares/example_tchannel"
	exampletchannelmiddlewaremodule "github.com/uber/zanzibar/examples/example-gateway/build/middlewares/example_tchannel/module"
	fixturecontactsclientgenerated "github.com/uber/zanzibar/examples/example-gateway/clients/contacts/fixture"
	fixturequuxclientstatic "github.com/uber/zanzibar/examples/example-gateway/clients/quux/fixture"
)

// MockClientNodes contains mock client dependencies
type MockClientNodes struct {
	Bar       *barclientgenerated.MockClient
	Baz       *bazclientgenerated.MockClient
	Contacts  *contactsclientgenerated.MockClientWithFixture
	GoogleNow *googlenowclientgenerated.MockClient
	Multi     *multiclientgenerated.MockClient
	Quux      *quuxclientgenerated.MockClientWithFixture
}

// InitializeDependenciesMock fully initializes all dependencies in the dep tree
// for the example-gateway service with leaf nodes being mocks
func InitializeDependenciesMock(
	g *zanzibar.Gateway,
	ctrl *gomock.Controller,
) (*module.DependenciesTree, *module.Dependencies, *MockClientNodes) {
	tree := &module.DependenciesTree{}

	initializedDefaultDependencies := &zanzibar.DefaultDependencies{
		ContextExtractor: g.ContextExtractor,
		ContextMetrics:   g.ContextMetrics,
		ContextLogger:    g.ContextLogger,
		Logger:           g.Logger,
		Scope:            g.RootScope,
		Config:           g.Config,
		Channel:          g.Channel,
		Tracer:           g.Tracer,
	}

	mockClientNodes := &MockClientNodes{
		Bar:       barclientgenerated.NewMockClient(ctrl),
		Baz:       bazclientgenerated.NewMockClient(ctrl),
		Contacts:  contactsclientgenerated.New(ctrl, fixturecontactsclientgenerated.Fixture),
		GoogleNow: googlenowclientgenerated.NewMockClient(ctrl),
		Multi:     multiclientgenerated.NewMockClient(ctrl),
		Quux:      quuxclientgenerated.New(ctrl, fixturequuxclientstatic.Fixture),
	}
	initializedClientDependencies := &module.ClientDependenciesNodes{}
	tree.Client = initializedClientDependencies
	initializedClientDependencies.Bar = mockClientNodes.Bar
	initializedClientDependencies.Baz = mockClientNodes.Baz
	initializedClientDependencies.Contacts = mockClientNodes.Contacts
	initializedClientDependencies.GoogleNow = mockClientNodes.GoogleNow
	initializedClientDependencies.Multi = mockClientNodes.Multi
	initializedClientDependencies.Quux = mockClientNodes.Quux

	initializedAdapterDependencies := &module.AdapterDependenciesNodes{}
	tree.Adapter = initializedAdapterDependencies
	initializedAdapterDependencies.ExampleAdapter = exampleadapteradaptergenerated.NewAdapter(&exampleadapteradaptermodule.Dependencies{
		Default: initializedDefaultDependencies,
	})
	initializedAdapterDependencies.ExampleAdapter2 = exampleadapter2adaptergenerated.NewAdapter(&exampleadapter2adaptermodule.Dependencies{
		Default: initializedDefaultDependencies,
	})
	initializedAdapterDependencies.ExampleAdapterTchannel = exampleadaptertchanneladaptergenerated.NewAdapter(&exampleadaptertchanneladaptermodule.Dependencies{
		Default: initializedDefaultDependencies,
	})

	initializedMiddlewareDependencies := &module.MiddlewareDependenciesNodes{}
	tree.Middleware = initializedMiddlewareDependencies
	initializedMiddlewareDependencies.Example = examplemiddlewaregenerated.NewMiddleware(&examplemiddlewaremodule.Dependencies{
		Default: initializedDefaultDependencies,
		Client: &examplemiddlewaremodule.ClientDependencies{
			Baz: initializedClientDependencies.Baz,
		},
	})
	initializedMiddlewareDependencies.ExampleTchannel = exampletchannelmiddlewaregenerated.NewMiddleware(&exampletchannelmiddlewaremodule.Dependencies{
		Default: initializedDefaultDependencies,
	})

	initializedEndpointDependencies := &module.EndpointDependenciesNodes{}
	tree.Endpoint = initializedEndpointDependencies
	initializedEndpointDependencies.Bar = barendpointgenerated.NewEndpoint(&barendpointmodule.Dependencies{
		Default: initializedDefaultDependencies,
		Adapter: &barendpointmodule.AdapterDependencies{
			ExampleAdapter:         initializedAdapterDependencies.ExampleAdapter,
			ExampleAdapter2:        initializedAdapterDependencies.ExampleAdapter2,
			ExampleAdapterTchannel: initializedAdapterDependencies.ExampleAdapterTchannel,
		},
		Client: &barendpointmodule.ClientDependencies{
			Bar: initializedClientDependencies.Bar,
		},
		Middleware: &barendpointmodule.MiddlewareDependencies{
			Example: initializedMiddlewareDependencies.Example,
		},
	})
	initializedEndpointDependencies.Baz = bazendpointgenerated.NewEndpoint(&bazendpointmodule.Dependencies{
		Default: initializedDefaultDependencies,
		Adapter: &bazendpointmodule.AdapterDependencies{
			ExampleAdapter:         initializedAdapterDependencies.ExampleAdapter,
			ExampleAdapter2:        initializedAdapterDependencies.ExampleAdapter2,
			ExampleAdapterTchannel: initializedAdapterDependencies.ExampleAdapterTchannel,
		},
		Client: &bazendpointmodule.ClientDependencies{
			Baz: initializedClientDependencies.Baz,
		},
	})
	initializedEndpointDependencies.BazTChannel = baztchannelendpointgenerated.NewEndpoint(&baztchannelendpointmodule.Dependencies{
		Default: initializedDefaultDependencies,
		Adapter: &baztchannelendpointmodule.AdapterDependencies{
			ExampleAdapter:         initializedAdapterDependencies.ExampleAdapter,
			ExampleAdapter2:        initializedAdapterDependencies.ExampleAdapter2,
			ExampleAdapterTchannel: initializedAdapterDependencies.ExampleAdapterTchannel,
		},
		Client: &baztchannelendpointmodule.ClientDependencies{
			Baz:  initializedClientDependencies.Baz,
			Quux: initializedClientDependencies.Quux,
		},
		Middleware: &baztchannelendpointmodule.MiddlewareDependencies{
			ExampleTchannel: initializedMiddlewareDependencies.ExampleTchannel,
		},
	})
	initializedEndpointDependencies.Contacts = contactsendpointgenerated.NewEndpoint(&contactsendpointmodule.Dependencies{
		Default: initializedDefaultDependencies,
		Adapter: &contactsendpointmodule.AdapterDependencies{
			ExampleAdapter:         initializedAdapterDependencies.ExampleAdapter,
			ExampleAdapter2:        initializedAdapterDependencies.ExampleAdapter2,
			ExampleAdapterTchannel: initializedAdapterDependencies.ExampleAdapterTchannel,
		},
		Client: &contactsendpointmodule.ClientDependencies{
			Contacts: initializedClientDependencies.Contacts,
		},
	})
	initializedEndpointDependencies.Googlenow = googlenowendpointgenerated.NewEndpoint(&googlenowendpointmodule.Dependencies{
		Default: initializedDefaultDependencies,
		Adapter: &googlenowendpointmodule.AdapterDependencies{
			ExampleAdapter:         initializedAdapterDependencies.ExampleAdapter,
			ExampleAdapter2:        initializedAdapterDependencies.ExampleAdapter2,
			ExampleAdapterTchannel: initializedAdapterDependencies.ExampleAdapterTchannel,
		},
		Client: &googlenowendpointmodule.ClientDependencies{
			GoogleNow: initializedClientDependencies.GoogleNow,
		},
	})
	initializedEndpointDependencies.Multi = multiendpointgenerated.NewEndpoint(&multiendpointmodule.Dependencies{
		Default: initializedDefaultDependencies,
		Adapter: &multiendpointmodule.AdapterDependencies{
			ExampleAdapter:         initializedAdapterDependencies.ExampleAdapter,
			ExampleAdapter2:        initializedAdapterDependencies.ExampleAdapter2,
			ExampleAdapterTchannel: initializedAdapterDependencies.ExampleAdapterTchannel,
		},
		Client: &multiendpointmodule.ClientDependencies{
			Multi: initializedClientDependencies.Multi,
		},
	})
	initializedEndpointDependencies.Panic = panicendpointgenerated.NewEndpoint(&panicendpointmodule.Dependencies{
		Default: initializedDefaultDependencies,
		Adapter: &panicendpointmodule.AdapterDependencies{
			ExampleAdapter:         initializedAdapterDependencies.ExampleAdapter,
			ExampleAdapter2:        initializedAdapterDependencies.ExampleAdapter2,
			ExampleAdapterTchannel: initializedAdapterDependencies.ExampleAdapterTchannel,
		},
		Client: &panicendpointmodule.ClientDependencies{
			Multi: initializedClientDependencies.Multi,
		},
	})
	initializedEndpointDependencies.PanicTChannel = panictchannelendpointgenerated.NewEndpoint(&panictchannelendpointmodule.Dependencies{
		Default: initializedDefaultDependencies,
		Adapter: &panictchannelendpointmodule.AdapterDependencies{
			ExampleAdapter:         initializedAdapterDependencies.ExampleAdapter,
			ExampleAdapter2:        initializedAdapterDependencies.ExampleAdapter2,
			ExampleAdapterTchannel: initializedAdapterDependencies.ExampleAdapterTchannel,
		},
		Client: &panictchannelendpointmodule.ClientDependencies{
			Baz: initializedClientDependencies.Baz,
		},
	})

	dependencies := &module.Dependencies{
		Default: initializedDefaultDependencies,
		Endpoint: &module.EndpointDependencies{
			Bar:           initializedEndpointDependencies.Bar,
			Baz:           initializedEndpointDependencies.Baz,
			BazTChannel:   initializedEndpointDependencies.BazTChannel,
			Contacts:      initializedEndpointDependencies.Contacts,
			Googlenow:     initializedEndpointDependencies.Googlenow,
			Multi:         initializedEndpointDependencies.Multi,
			Panic:         initializedEndpointDependencies.Panic,
			PanicTChannel: initializedEndpointDependencies.PanicTChannel,
		},
	}

	return tree, dependencies, mockClientNodes
}
