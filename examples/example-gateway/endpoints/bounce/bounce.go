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

package bounce

import (
	"context"

	"github.com/uber/zanzibar/examples/example-gateway/build/clients/echo"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/bounce/module"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/bounce/workflow"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/echo"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/bounce/bounce"
	"github.com/uber/zanzibar/runtime"
)

// NewBounceBounceWorkflow ...
func NewBounceBounceWorkflow(deps *module.Dependencies) workflow.BounceBounceWorkflow {
	return &bounceWorkflow{
		echo: deps.Client.Echo,
	}
}

type bounceWorkflow struct {
	echo echoclient.Client
}

// Handle ...
func (w bounceWorkflow) Handle(
	ctx context.Context,
	reqHeaders zanzibar.Header,
	req *bounce.Bounce_Bounce_Args,
) (string, zanzibar.Header, error) {
	res, err := w.echo.Echo(ctx, &echo.Request{Message: req.Msg})
	if err != nil {
		return "", nil, err
	}
	return res.Message, nil, nil
}
