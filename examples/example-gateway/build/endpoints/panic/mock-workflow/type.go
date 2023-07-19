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

package mockpanicworkflow

import (
	bazclientgenerated "github.com/uber/zanzibar/v2/examples/example-gateway/build/clients/baz"
	bazclientgeneratedmock "github.com/uber/zanzibar/v2/examples/example-gateway/build/clients/baz/mock-client"
	multiclientgenerated "github.com/uber/zanzibar/v2/examples/example-gateway/build/clients/multi"
	multiclientgeneratedmock "github.com/uber/zanzibar/v2/examples/example-gateway/build/clients/multi/mock-client"
	defaultexamplemiddlewaregenerated "github.com/uber/zanzibar/v2/examples/example-gateway/build/middlewares/default/default_example"
	defaultexample2middlewaregenerated "github.com/uber/zanzibar/v2/examples/example-gateway/build/middlewares/default/default_example2"
	defaultexampletchannelmiddlewaregenerated "github.com/uber/zanzibar/v2/examples/example-gateway/build/middlewares/default/default_example_tchannel"
)

// MockClientNodes contains mock client dependencies for the panic endpoint module
type MockClientNodes struct {
	Baz   *bazclientgeneratedmock.MockClient
	Multi *multiclientgeneratedmock.MockClient
}

// clientDependenciesNodes contains client dependencies
type clientDependenciesNodes struct {
	Baz   bazclientgenerated.Client
	Multi multiclientgenerated.Client
}

// middlewareDependenciesNodes contains middleware dependencies
type middlewareDependenciesNodes struct {
	DefaultExample         defaultexamplemiddlewaregenerated.Middleware
	DefaultExample2        defaultexample2middlewaregenerated.Middleware
	DefaultExampleTchannel defaultexampletchannelmiddlewaregenerated.Middleware
}
