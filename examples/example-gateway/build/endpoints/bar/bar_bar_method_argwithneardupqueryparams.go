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

package barendpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	zanzibar "github.com/uber/zanzibar/v2/runtime"
	"go.uber.org/thriftrw/ptr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	workflow "github.com/uber/zanzibar/v2/examples/example-gateway/build/endpoints/bar/workflow"
	endpointsIDlEndpointsBarBar "github.com/uber/zanzibar/v2/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/bar/bar"

	defaultExample "github.com/uber/zanzibar/v2/examples/example-gateway/middlewares/default/default_example"
	defaultExample2 "github.com/uber/zanzibar/v2/examples/example-gateway/middlewares/default/default_example2"

	module "github.com/uber/zanzibar/v2/examples/example-gateway/build/endpoints/bar/module"
)

// BarArgWithNearDupQueryParamsHandler is the handler for "/bar/argWithNearDupQueryParams"
type BarArgWithNearDupQueryParamsHandler struct {
	Dependencies *module.Dependencies
	endpoint     *zanzibar.RouterEndpoint
}

// NewBarArgWithNearDupQueryParamsHandler creates a handler
func NewBarArgWithNearDupQueryParamsHandler(deps *module.Dependencies) *BarArgWithNearDupQueryParamsHandler {
	handler := &BarArgWithNearDupQueryParamsHandler{
		Dependencies: deps,
	}
	handler.endpoint = zanzibar.NewRouterEndpoint(
		deps.Default.ContextExtractor, deps.Default,
		"bar", "argWithNearDupQueryParams",
		zanzibar.NewStack([]zanzibar.MiddlewareHandle{
			deps.Middleware.DefaultExample2.NewMiddlewareHandle(
				defaultExample2.Options{},
			),
			deps.Middleware.DefaultExample.NewMiddlewareHandle(
				defaultExample.Options{},
			),
		}, handler.HandleRequest).Handle,
	)

	return handler
}

// Register adds the http handler to the gateway's http router
func (h *BarArgWithNearDupQueryParamsHandler) Register(g *zanzibar.Gateway) error {
	return g.HTTPRouter.Handle(
		"GET", "/bar/argWithNearDupQueryParams",
		http.HandlerFunc(h.endpoint.HandleRequest),
	)
}

// HandleRequest handles "/bar/argWithNearDupQueryParams".
func (h *BarArgWithNearDupQueryParamsHandler) HandleRequest(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
) context.Context {
	defer func() {
		if r := recover(); r != nil {
			stacktrace := string(debug.Stack())
			e := errors.Errorf("enpoint panic: %v, stacktrace: %v", r, stacktrace)
			ctx = h.Dependencies.Default.ContextLogger.ErrorZ(
				ctx,
				"Endpoint failure: endpoint panic",
				zap.Error(e),
				zap.String("stacktrace", stacktrace))

			h.Dependencies.Default.ContextMetrics.IncCounter(ctx, zanzibar.MetricEndpointPanics, 1)
			res.SendError(502, "Unexpected workflow panic, recovered at endpoint.", nil)
		}
	}()

	var requestBody endpointsIDlEndpointsBarBar.Bar_ArgWithNearDupQueryParams_Args

	oneNameOk := req.CheckQueryValue("oneName")
	if !oneNameOk {
		return ctx
	}
	oneNameQuery, ok := req.GetQueryValue("oneName")
	if !ok {
		return ctx
	}
	requestBody.One = oneNameQuery

	oneName1Ok := req.HasQueryValue("one_name")
	if oneName1Ok {
		oneName1Query, ok := req.GetQueryInt32("one_name")
		if !ok {
			return ctx
		}
		requestBody.Two = ptr.Int32(oneName1Query)
	}

	oneNamEOk := req.HasQueryValue("One_NamE")
	if oneNamEOk {
		oneNamEQuery, ok := req.GetQueryValue("One_NamE")
		if !ok {
			return ctx
		}
		requestBody.Three = ptr.String(oneNamEQuery)
	}

	oneName2Ok := req.HasQueryValue("one-Name")
	if oneName2Ok {
		oneName2Query, ok := req.GetQueryValue("one-Name")
		if !ok {
			return ctx
		}
		requestBody.Four = ptr.String(oneName2Query)
	}

	// log endpoint request to downstream services
	if ce := h.Dependencies.Default.ContextLogger.Check(zapcore.DebugLevel, "stub"); ce != nil {
		var zfields []zapcore.Field
		zfields = append(zfields, zap.String("body", fmt.Sprintf("%s", req.GetRawBody())))
		for _, k := range req.Header.Keys() {
			if val, ok := req.Header.Get(k); ok {
				zfields = append(zfields, zap.String(k, val))
			}
		}
		ctx = h.Dependencies.Default.ContextLogger.DebugZ(ctx, "endpoint request to downstream", zfields...)
	}

	w := workflow.NewBarArgWithNearDupQueryParamsWorkflow(h.Dependencies)
	if span := req.GetSpan(); span != nil {
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	ctx, response, cliRespHeaders, err := w.Handle(ctx, req.Header, &requestBody)

	// log downstream response to endpoint
	if ce := h.Dependencies.Default.ContextLogger.Check(zapcore.DebugLevel, "stub"); ce != nil {
		var zfields []zapcore.Field
		if body, err := json.Marshal(response); err == nil {
			zfields = append(zfields, zap.String("body", fmt.Sprintf("%s", body)))
		}
		if cliRespHeaders != nil {
			for _, k := range cliRespHeaders.Keys() {
				if val, ok := cliRespHeaders.Get(k); ok {
					zfields = append(zfields, zap.String(k, val))
				}
			}
		}
		if traceKey, ok := req.Header.Get("x-trace-id"); ok {
			zfields = append(zfields, zap.String("x-trace-id", traceKey))
		}
		ctx = h.Dependencies.Default.ContextLogger.DebugZ(ctx, "downstream service response", zfields...)
	}
	// map useful client response headers to server response
	if cliRespHeaders != nil {
		if val, ok := cliRespHeaders.Get(zanzibar.ClientResponseDurationKey); ok {
			if duration, err := time.ParseDuration(val); err == nil {
				res.DownstreamFinishTime = duration
			}
			cliRespHeaders.Unset(zanzibar.ClientResponseDurationKey)
		}
		if val, ok := cliRespHeaders.Get(zanzibar.ClientTypeKey); ok {
			res.ClientType = val
			cliRespHeaders.Unset(zanzibar.ClientTypeKey)
		}
	}

	if err != nil {
		res.SendError(500, "Unexpected server error", err)
		return ctx

	}

	res.WriteJSON(200, cliRespHeaders, response)
	return ctx
}
