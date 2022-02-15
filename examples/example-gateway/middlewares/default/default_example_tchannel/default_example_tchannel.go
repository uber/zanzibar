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

package defaultexampletchannel

import (
	"context"

	"go.uber.org/thriftrw/protocol/stream"

	"github.com/mcuadros/go-jsonschema-generator"
	"github.com/uber/zanzibar/examples/example-gateway/build/middlewares/default/default_example_tchannel/module"
	zanzibar "github.com/uber/zanzibar/runtime"
)

type defaultExampleTchannelMiddleware struct {
	deps    *module.Dependencies
	options Options
}

// Options for middleware configuration
type Options struct{}

// NewMiddleware creates a new middleware that executes the next middleware
// after performing it's operations.
func NewMiddleware(
	deps *module.Dependencies,
	options Options,
) zanzibar.MiddlewareTchannelHandle {
	return &defaultExampleTchannelMiddleware{
		deps:    deps,
		options: options,
	}
}

// HandleRequest handles the requests before calling lower level middlewares.
func (m *defaultExampleTchannelMiddleware) HandleRequest(
	ctx context.Context,
	reqHeaders map[string]string,
	sr stream.Reader,
	shared zanzibar.TchannelSharedState,
) (context.Context, bool, error) {
	return ctx, true, nil
}

func (m *defaultExampleTchannelMiddleware) HandleResponse(
	ctx context.Context,
	rwt zanzibar.RWTStruct,
	shared zanzibar.TchannelSharedState,
) zanzibar.RWTStruct {
	return rwt
}

// JSONSchema returns a schema definition of the configuration options for a middlware
func (m *defaultExampleTchannelMiddleware) JSONSchema() *jsonschema.Document {
	s := &jsonschema.Document{}
	s.Read(&Options{})
	return s
}

func (m *defaultExampleTchannelMiddleware) Name() string {
	return "default_example_tchannel"
}
