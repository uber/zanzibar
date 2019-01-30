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

package mockpanictchannelworkflow

import (
	exampleadaptertchanneladaptergenerated "github.com/uber/zanzibar/examples/example-gateway/build/adapters/example_adapter_tchannel"
	bazclientgenerated "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	bazclientgeneratedmock "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz/mock-client"
)

// MockClientNodes contains mock client dependencies for the panicTChannel endpoint module
type MockClientNodes struct {
	Baz *bazclientgeneratedmock.MockClient
}

// clientDependenciesNodes contains client dependencies
type clientDependenciesNodes struct {
	Baz bazclientgenerated.Client
}

// adapterDependenciesNodes contains adapter dependencies
type adapterDependenciesNodes struct {
	ExampleAdapterTchannel exampleadaptertchanneladaptergenerated.Adapter
}
