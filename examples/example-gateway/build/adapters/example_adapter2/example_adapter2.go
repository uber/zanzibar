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

package exampleadapter2adapter

import (
	handle "github.com/uber/zanzibar/examples/example-gateway/adapters/example_adapter2"
	module "github.com/uber/zanzibar/examples/example-gateway/build/adapters/example_adapter2/module"
	zanzibar "github.com/uber/zanzibar/runtime"
)

// Adapter is a container for module.Deps and factory for AdapterHandle
type Adapter struct {
	Deps *module.Dependencies
}

// NewAdapter is a factory method for the struct
func NewAdapter(deps *module.Dependencies) Adapter {
	return Adapter{
		Deps: deps,
	}
}

// NewAdapterHandle calls back to the custom adapter to build an AdapterHandle
func (m *Adapter) NewAdapterHandle() zanzibar.AdapterHandle {
	return handle.NewAdapter(m.Deps)
}
