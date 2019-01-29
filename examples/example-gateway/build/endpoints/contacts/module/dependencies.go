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
	testadapter1adaptergenerated "github.com/uber/zanzibar/examples/example-gateway/build/adapters/test_adapter1"
	testadapter2adaptergenerated "github.com/uber/zanzibar/examples/example-gateway/build/adapters/test_adapter2"
	contactsclientgenerated "github.com/uber/zanzibar/examples/example-gateway/build/clients/contacts"

	zanzibar "github.com/uber/zanzibar/runtime"
)

// Dependencies contains dependencies for the contacts endpoint module
type Dependencies struct {
	Default *zanzibar.DefaultDependencies
	Adapter *AdapterDependencies
	Client  *ClientDependencies
}

// AdapterDependencies contains adapter dependencies
type AdapterDependencies struct {
	TestAdapter1 testadapter1adaptergenerated.Adapter
	TestAdapter2 testadapter2adaptergenerated.Adapter
}

// ClientDependencies contains client dependencies
type ClientDependencies struct {
	Contacts contactsclientgenerated.Client
}
