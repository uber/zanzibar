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

package exampletchannel

import (
	"context"
	"errors"
	"github.com/mcuadros/go-jsonschema-generator"
	"github.com/uber/zanzibar/examples/example-gateway/build/middlewares/example_tchannel/module"
	"github.com/uber/zanzibar/runtime"
	"go.uber.org/thriftrw/wire"
)

type exampleTchannelMiddleware struct {
	deps    *module.Dependencies
	options Options
}

// FooResult is an example of ThriftRW
type FooResult struct {
	Success string `json:"success,omitempty"`
}

// ToWire translates a FooResult struct into a Thrift-level intermediate
// representation. This intermediate representation may be serialized
// into bytes using a ThriftRW protocol implementation.
func (v FooResult) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		w      wire.Value
		err    error
	)

	w, err = wire.NewValueString(v.Success), error(nil)
	if err != nil {
		return w, err
	}
	fields[0] = wire.Field{ID: 1, Value: w}

	return wire.NewValueStruct(wire.Struct{Fields: fields[:1]}), nil
}

// FromWire deserializes a FooResult struct from its Thrift-level
// representation. The Thrift-level representation may be obtained
// from a ThriftRW protocol implementation.
func (v FooResult) FromWire(w wire.Value) error {
	var err error

	stringFieldIsSet := false

	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.Success, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				stringFieldIsSet = true
			}
		}
	}

	if !stringFieldIsSet {
		return errors.New("field StringField of Success is required")
	}

	return nil
}

// Options for middleware configuration
type Options struct {
	Foo string `json:"Foo"` // string to verify in shared
}

// MiddlewareState accessible by other middlewares and endpoint handler
// though the context object.
type MiddlewareState struct{}

// NewMiddleware creates a new middleware that executes the next middleware
// after performing it's operations.
func NewMiddleware(
	deps *module.Dependencies,
	options Options,
) zanzibar.MiddlewareTchannelHandle {
	return &exampleTchannelMiddleware{
		deps:    deps,
		options: options,
	}
}

// HandleRequest handles the requests before calling lower level middlewares.
func (m *exampleTchannelMiddleware) HandleRequest(
	ctx context.Context,
	reqHeaders map[string]string,
	wireValue *wire.Value,
	shared zanzibar.TchannelSharedState,
) bool {
	return true
}

func (m *exampleTchannelMiddleware) HandleResponse(
	ctx context.Context,
	wireValue *wire.Value,
	shared zanzibar.TchannelSharedState,
) zanzibar.RWTStruct {
	var res FooResult
	res.Success = "Foo_Success"
	return res
}

// JSONSchema returns a schema definition of the configuration options for a middlware
func (m *exampleTchannelMiddleware) JSONSchema() *jsonschema.Document {
	s := &jsonschema.Document{}
	s.Read(&Options{})
	return s
}

func (m *exampleTchannelMiddleware) Name() string {
	return "example_tchannel"
}
