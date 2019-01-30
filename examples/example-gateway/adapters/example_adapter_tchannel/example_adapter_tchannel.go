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

package exampleadaptertchannel

import (
	"context"
	"github.com/mcuadros/go-jsonschema-generator"
	"github.com/uber/zanzibar/examples/example-gateway/build/adapters/example_adapter_tchannel/module"
	"github.com/uber/zanzibar/runtime"
	"go.uber.org/thriftrw/wire"
)

type exampleAdapterTchannel struct {
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
) zanzibar.AdapterTchannelHandle {
	return &exampleAdapterTchannel{
		deps:    deps,
		options: options,
	}
}

// HandleRequest handles the requests before calling lower level adapters.
func (m *exampleAdapterTchannel) HandleRequest(
	ctx context.Context,
	reqHeaders map[string]string,
	wireValue *wire.Value,
	shared zanzibar.TchannelSharedState,
) (bool, error) {
	return true, nil
}

func (m *exampleAdapterTchannel) HandleResponse(
	ctx context.Context,
	rwt zanzibar.RWTStruct,
	shared zanzibar.TchannelSharedState,
) zanzibar.RWTStruct {
	return rwt
}

// JSONSchema returns a schema definition of the configuration options for an adapter
func (m *exampleAdapterTchannel) JSONSchema() *jsonschema.Document {
	s := &jsonschema.Document{}
	s.Read(&Options{})
	return s
}

func (m *exampleAdapterTchannel) Name() string {
	return "example_adapter_tchannel"
}
