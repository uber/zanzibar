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

package testadapter1

import (
	"context"
	"github.com/mcuadros/go-jsonschema-generator"

	"github.com/uber/zanzibar/examples/example-gateway/build/adapters/test_adapter1/module"
	"github.com/uber/zanzibar/runtime"
)

type test1Adapter struct {
	deps    *module.Dependencies
	options Options
}

// Options for adapter configuration
type Options struct{}

// NewAdapter creates a new adapter that executes the next adapter
// after performing it's operations.
func NewAdapter(
	deps *module.Dependencies,
	options Options,
) zanzibar.AdapterHandle {
	return &test1Adapter{
		deps:    deps,
		options: options,
	}
}

// HandleRequest handles the requests before calling lower level adapters.
func (m *test1Adapter) HandleRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
) bool {
	return true
}

func (m *test1Adapter) HandleResponse(
	ctx context.Context,
	res *zanzibar.ServerHTTPResponse,
) {
}

// JSONSchema returns a schema definition of the configuration options for an adapter
func (m *test1Adapter) JSONSchema() *jsonschema.Document {
	s := &jsonschema.Document{}
	s.Read(&Options{})
	return s
}

func (m *test1Adapter) Name() string {
	return "test_adapter1"
}
