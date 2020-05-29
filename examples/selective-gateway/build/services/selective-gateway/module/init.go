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

package module

import (
	echoclientgenerated "github.com/uber/zanzibar/examples/selective-gateway/build/clients/echo"
	echoclientmodule "github.com/uber/zanzibar/examples/selective-gateway/build/clients/echo/module"
	bounceendpointgenerated "github.com/uber/zanzibar/examples/selective-gateway/build/endpoints/bounce"
	bounceendpointmodule "github.com/uber/zanzibar/examples/selective-gateway/build/endpoints/bounce/module"

	zanzibar "github.com/uber/zanzibar/runtime"
)

// DependenciesTree contains all deps for this service.
type DependenciesTree struct {
	Client   *ClientDependenciesNodes
	Endpoint *EndpointDependenciesNodes
}

// ClientDependenciesNodes contains client dependencies
type ClientDependenciesNodes struct {
	Echo echoclientgenerated.Client
}

// EndpointDependenciesNodes contains endpoint dependencies
type EndpointDependenciesNodes struct {
	Bounce bounceendpointgenerated.Endpoint
}

// InitializeDependencies fully initializes all dependencies in the dep tree
// for the selective-gateway service
func InitializeDependencies(
	g *zanzibar.Gateway,
) (*DependenciesTree, *Dependencies) {
	tree := &DependenciesTree{}

	initializedDefaultDependencies := &zanzibar.DefaultDependencies{
		Logger:           g.Logger,
		ContextExtractor: g.ContextExtractor,
		ContextLogger:    g.ContextLogger,
		ContextMetrics:   zanzibar.NewContextMetrics(g.RootScope),
		Scope:            g.RootScope,
		Tracer:           g.Tracer,
		Config:           g.Config,
		Channel:          g.Channel,

		GRPCClientDispatcher: g.GRPCClientDispatcher,
		JSONWrapper:          g.JSONWrapper,
	}

	initializedClientDependencies := &ClientDependenciesNodes{}
	tree.Client = initializedClientDependencies
	initializedClientDependencies.Echo = echoclientgenerated.NewClient(&echoclientmodule.Dependencies{
		Default: initializedDefaultDependencies,
		Client:  &echoclientmodule.ClientDependencies{},
	})

	initializedEndpointDependencies := &EndpointDependenciesNodes{}
	tree.Endpoint = initializedEndpointDependencies
	initializedEndpointDependencies.Bounce = bounceendpointgenerated.NewEndpoint(&bounceendpointmodule.Dependencies{
		Default: initializedDefaultDependencies,
		Client: &bounceendpointmodule.ClientDependencies{
			Echo: initializedClientDependencies.Echo,
		},
	})

	dependencies := &Dependencies{
		Default: initializedDefaultDependencies,
		Endpoint: &EndpointDependencies{
			Bounce: initializedEndpointDependencies.Bounce,
		},
	}

	return tree, dependencies
}
