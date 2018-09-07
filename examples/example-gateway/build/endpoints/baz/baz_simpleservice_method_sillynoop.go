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

package bazendpoint

import (
	"context"
	"runtime/debug"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	workflow "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/baz/workflow"
	endpointsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/baz/baz"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/baz/module"
)

// SimpleServiceSillyNoopHandler is the handler for "/baz/silly-noop"
type SimpleServiceSillyNoopHandler struct {
	Dependencies  *module.Dependencies
	endpoint      *zanzibar.RouterEndpoint
	endpointScope tally.Scope
}

// NewSimpleServiceSillyNoopHandler creates a handler
func NewSimpleServiceSillyNoopHandler(deps *module.Dependencies) *SimpleServiceSillyNoopHandler {
	handler := &SimpleServiceSillyNoopHandler{
		Dependencies: deps,
	}
	handler.endpoint = zanzibar.NewRouterEndpoint(
		deps.Default.Logger, deps.Default.Scope, deps.Default.Tracer,
		"baz", "sillyNoop",
		handler.HandleRequest,
	)
	handler.endpointScope = deps.Default.Scope.Tagged(map[string]string{
		"endpoint": handler.endpoint.EndpointName,
	})
	return handler
}

// Register adds the http handler to the gateway's http router
func (h *SimpleServiceSillyNoopHandler) Register(g *zanzibar.Gateway) error {
	return g.HTTPRouter.Register(
		"GET", "/baz/silly-noop",
		h.endpoint,
	)
}

// HandleRequest handles "/baz/silly-noop".
func (h *SimpleServiceSillyNoopHandler) HandleRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
) {
	defer func() {
		if r := recover(); r != nil {
			stacktrace := string(debug.Stack())
			e := errors.Errorf("enpoint panic: %v, stacktrace: %v", r, stacktrace)
			h.Dependencies.Default.ContextLogger.Error(
				ctx,
				"endpoint panic",
				zap.Error(e),
				zap.String("stacktrace", stacktrace),
				zap.String("endpoint", h.endpoint.EndpointName))

			h.endpointScope.Counter("endpoint.panic").Inc(1)
			res.SendError(502, "Unexpected workflow panic, recovered at endpoint.", nil)
		}
	}()

	// log endpoint request to downstream services
	if ce := h.Dependencies.Default.ContextLogger.Check(zapcore.DebugLevel, "stub"); ce != nil {
		zfields := []zapcore.Field{
			zap.String("endpoint", h.endpoint.EndpointName),
		}
		for _, k := range req.Header.Keys() {
			if val, ok := req.Header.Get(k); ok {
				zfields = append(zfields, zap.String(k, val))
			}
		}
		h.Dependencies.Default.ContextLogger.Debug(ctx, "endpoint request to downstream", zfields...)
	}

	w := workflow.NewSimpleServiceSillyNoopWorkflow(h.Dependencies)
	if span := req.GetSpan(); span != nil {
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	cliRespHeaders, err := w.Handle(ctx, req.Header)
	if err != nil {
		switch errValue := err.(type) {

		case *endpointsBazBaz.AuthErr:
			res.WriteJSON(
				403, cliRespHeaders, errValue,
			)
			return

		case *endpointsBazBaz.ServerErr:
			res.WriteJSON(
				500, cliRespHeaders, errValue,
			)
			return

		default:
			res.SendError(500, "Unexpected server error", err)
			return
		}

	}

	res.WriteJSONBytes(204, cliRespHeaders, nil)
}
