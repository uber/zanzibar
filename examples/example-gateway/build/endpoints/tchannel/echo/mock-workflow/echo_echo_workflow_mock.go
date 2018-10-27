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

package mockechoworkflow

import (
	"testing"

	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/tchannel/echo/module"
	workflow "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/tchannel/echo/workflow"
	echoendpointstatic "github.com/uber/zanzibar/examples/example-gateway/endpoints/tchannel/echo"
)

// NewEchoEchoWorkflowMock creates a workflow with mock clients
func NewEchoEchoWorkflowMock(t *testing.T) (workflow.EchoEchoWorkflow, *MockNodes) {
	mockNodes := &MockNodes{}

	initializedDefaultDependencies := &zanzibar.DefaultDependencies{
		Logger: zap.NewNop(),
	}
	initializedDefaultDependencies.ContextLogger = zanzibar.NewContextLogger(initializedDefaultDependencies.Logger)

	w := echoendpointstatic.NewEchoEchoWorkflow(
		&module.Dependencies{
			Default: initializedDefaultDependencies,
		},
	)

	return w, mockNodes
}
