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
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/thriftrw/ptr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	workflow "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/bar/workflow"
	endpointsIDlEndpointsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/bar/bar"

	defaultExample "github.com/uber/zanzibar/examples/example-gateway/middlewares/default/default_example"
	defaultExample2 "github.com/uber/zanzibar/examples/example-gateway/middlewares/default/default_example2"

	module "github.com/uber/zanzibar/examples/example-gateway/build/endpoints/bar/module"
)

// BarArgWithManyQueryParamsHandler is the handler for "/bar/argWithManyQueryParams"
type BarArgWithManyQueryParamsHandler struct {
	Dependencies *module.Dependencies
	endpoint     *zanzibar.RouterEndpoint
}

// NewBarArgWithManyQueryParamsHandler creates a handler
func NewBarArgWithManyQueryParamsHandler(deps *module.Dependencies) *BarArgWithManyQueryParamsHandler {
	handler := &BarArgWithManyQueryParamsHandler{
		Dependencies: deps,
	}
	handler.endpoint = zanzibar.NewRouterEndpoint(
		deps.Default.ContextExtractor, deps.Default,
		"bar", "argWithManyQueryParams",
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
func (h *BarArgWithManyQueryParamsHandler) Register(g *zanzibar.Gateway) error {
	return g.HTTPRouter.Handle(
		"GET", "/bar/argWithManyQueryParams",
		http.HandlerFunc(h.endpoint.HandleRequest),
	)
}

// HandleRequest handles "/bar/argWithManyQueryParams".
func (h *BarArgWithManyQueryParamsHandler) HandleRequest(
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
				zap.String("stacktrace", stacktrace),
				zap.String("endpoint", h.endpoint.EndpointName))

			h.Dependencies.Default.ContextMetrics.IncCounter(ctx, zanzibar.MetricEndpointPanics, 1)
			res.SendError(502, "Unexpected workflow panic, recovered at endpoint.", nil)
		}
	}()

	var requestBody endpointsIDlEndpointsBarBar.Bar_ArgWithManyQueryParams_Args

	aStrOk := req.CheckQueryValue("aStr")
	if !aStrOk {
		return ctx
	}
	aStrQuery, ok := req.GetQueryValue("aStr")
	if !ok {
		return ctx
	}
	requestBody.AStr = aStrQuery

	anOptStrOk := req.HasQueryValue("anOptStr")
	if anOptStrOk {
		anOptStrQuery, ok := req.GetQueryValue("anOptStr")
		if !ok {
			return ctx
		}
		requestBody.AnOptStr = ptr.String(anOptStrQuery)
	}

	aBoolOk := req.CheckQueryValue("aBool")
	if !aBoolOk {
		return ctx
	}
	aBoolQuery, ok := req.GetQueryBool("aBool")
	if !ok {
		return ctx
	}
	requestBody.ABool = aBoolQuery

	anOptBoolOk := req.HasQueryValue("anOptBool")
	if anOptBoolOk {
		anOptBoolQuery, ok := req.GetQueryBool("anOptBool")
		if !ok {
			return ctx
		}
		requestBody.AnOptBool = ptr.Bool(anOptBoolQuery)
	}

	aInt8Ok := req.CheckQueryValue("aInt8")
	if !aInt8Ok {
		return ctx
	}
	aInt8Query, ok := req.GetQueryInt8("aInt8")
	if !ok {
		return ctx
	}
	requestBody.AInt8 = aInt8Query

	anOptInt8Ok := req.HasQueryValue("anOptInt8")
	if anOptInt8Ok {
		anOptInt8Query, ok := req.GetQueryInt8("anOptInt8")
		if !ok {
			return ctx
		}
		requestBody.AnOptInt8 = ptr.Int8(anOptInt8Query)
	}

	aInt16Ok := req.CheckQueryValue("aInt16")
	if !aInt16Ok {
		return ctx
	}
	aInt16Query, ok := req.GetQueryInt16("aInt16")
	if !ok {
		return ctx
	}
	requestBody.AInt16 = aInt16Query

	anOptInt16Ok := req.HasQueryValue("anOptInt16")
	if anOptInt16Ok {
		anOptInt16Query, ok := req.GetQueryInt16("anOptInt16")
		if !ok {
			return ctx
		}
		requestBody.AnOptInt16 = ptr.Int16(anOptInt16Query)
	}

	aInt32Ok := req.CheckQueryValue("aInt32")
	if !aInt32Ok {
		return ctx
	}
	aInt32Query, ok := req.GetQueryInt32("aInt32")
	if !ok {
		return ctx
	}
	requestBody.AInt32 = aInt32Query

	anOptInt32Ok := req.HasQueryValue("anOptInt32")
	if anOptInt32Ok {
		anOptInt32Query, ok := req.GetQueryInt32("anOptInt32")
		if !ok {
			return ctx
		}
		requestBody.AnOptInt32 = ptr.Int32(anOptInt32Query)
	}

	aInt64Ok := req.CheckQueryValue("aInt64")
	if !aInt64Ok {
		return ctx
	}
	aInt64Query, ok := req.GetQueryInt64("aInt64")
	if !ok {
		return ctx
	}
	requestBody.AInt64 = aInt64Query

	anOptInt64Ok := req.HasQueryValue("anOptInt64")
	if anOptInt64Ok {
		anOptInt64Query, ok := req.GetQueryInt64("anOptInt64")
		if !ok {
			return ctx
		}
		requestBody.AnOptInt64 = ptr.Int64(anOptInt64Query)
	}

	aFloat64Ok := req.CheckQueryValue("aFloat64")
	if !aFloat64Ok {
		return ctx
	}
	aFloat64Query, ok := req.GetQueryFloat64("aFloat64")
	if !ok {
		return ctx
	}
	requestBody.AFloat64 = aFloat64Query

	anOptFloat64Ok := req.HasQueryValue("anOptFloat64")
	if anOptFloat64Ok {
		anOptFloat64Query, ok := req.GetQueryFloat64("anOptFloat64")
		if !ok {
			return ctx
		}
		requestBody.AnOptFloat64 = ptr.Float64(anOptFloat64Query)
	}

	aUUIDOk := req.CheckQueryValue("aUUID")
	if !aUUIDOk {
		return ctx
	}
	aUUIDQuery, ok := req.GetQueryValue("aUUID")
	if !ok {
		return ctx
	}
	requestBody.AUUID = (endpointsIDlEndpointsBarBar.UUID)(aUUIDQuery)

	anOptUUIDOk := req.HasQueryValue("anOptUUID")
	if anOptUUIDOk {
		anOptUUIDQuery, ok := req.GetQueryValue("anOptUUID")
		if !ok {
			return ctx
		}
		requestBody.AnOptUUID = (*endpointsIDlEndpointsBarBar.UUID)(ptr.String(string(anOptUUIDQuery)))
	}

	aListUUIDOk := req.CheckQueryValue("aListUUID")
	if !aListUUIDOk {
		return ctx
	}
	aListUUIDQuery, ok := req.GetQueryValueList("aListUUID")
	if !ok {
		return ctx
	}
	aListUUIDQueryFinal := make([]endpointsIDlEndpointsBarBar.UUID, len(aListUUIDQuery))
	for i, v := range aListUUIDQuery {
		aListUUIDQueryFinal[i] = endpointsIDlEndpointsBarBar.UUID(v)
	}
	requestBody.AListUUID = aListUUIDQueryFinal

	anOptListUUIDOk := req.HasQueryValue("anOptListUUID")
	if anOptListUUIDOk {
		anOptListUUIDQuery, ok := req.GetQueryValueList("anOptListUUID")
		if !ok {
			return ctx
		}
		anOptListUUIDQueryFinal := make([]endpointsIDlEndpointsBarBar.UUID, len(anOptListUUIDQuery))
		for i, v := range anOptListUUIDQuery {
			anOptListUUIDQueryFinal[i] = endpointsIDlEndpointsBarBar.UUID(v)
		}
		requestBody.AnOptListUUID = anOptListUUIDQueryFinal
	}

	aStringListOk := req.CheckQueryValue("aStringList")
	if !aStringListOk {
		return ctx
	}
	aStringListQuery, ok := req.GetQueryValueList("aStringList")
	if !ok {
		return ctx
	}
	requestBody.AStringList = (endpointsIDlEndpointsBarBar.StringList)(aStringListQuery)

	anOptStringListOk := req.HasQueryValue("anOptStringList")
	if anOptStringListOk {
		anOptStringListQuery, ok := req.GetQueryValueList("anOptStringList")
		if !ok {
			return ctx
		}
		requestBody.AnOptStringList = (endpointsIDlEndpointsBarBar.StringList)(anOptStringListQuery)
	}

	aUUIDListOk := req.CheckQueryValue("aUUIDList")
	if !aUUIDListOk {
		return ctx
	}
	aUUIDListQuery, ok := req.GetQueryValueList("aUUIDList")
	if !ok {
		return ctx
	}
	aUUIDListQueryFinal := make([]endpointsIDlEndpointsBarBar.UUID, len(aUUIDListQuery))
	for i, v := range aUUIDListQuery {
		aUUIDListQueryFinal[i] = endpointsIDlEndpointsBarBar.UUID(v)
	}
	requestBody.AUUIDList = (endpointsIDlEndpointsBarBar.UUIDList)(aUUIDListQueryFinal)

	anOptUUIDListOk := req.HasQueryValue("anOptUUIDList")
	if anOptUUIDListOk {
		anOptUUIDListQuery, ok := req.GetQueryValueList("anOptUUIDList")
		if !ok {
			return ctx
		}
		anOptUUIDListQueryFinal := make([]endpointsIDlEndpointsBarBar.UUID, len(anOptUUIDListQuery))
		for i, v := range anOptUUIDListQuery {
			anOptUUIDListQueryFinal[i] = endpointsIDlEndpointsBarBar.UUID(v)
		}
		requestBody.AnOptUUIDList = (endpointsIDlEndpointsBarBar.UUIDList)(anOptUUIDListQueryFinal)
	}

	aTsOk := req.CheckQueryValue("aTs")
	if !aTsOk {
		return ctx
	}
	aTsQuery, ok := req.GetQueryInt64("aTs")
	if !ok {
		return ctx
	}
	requestBody.ATs = (endpointsIDlEndpointsBarBar.Timestamp)(aTsQuery)

	anOptTsOk := req.HasQueryValue("anOptTs")
	if anOptTsOk {
		anOptTsQuery, ok := req.GetQueryInt64("anOptTs")
		if !ok {
			return ctx
		}
		requestBody.AnOptTs = (*endpointsIDlEndpointsBarBar.Timestamp)(ptr.Int64(int64(anOptTsQuery)))
	}

	aReqDemoOk := req.CheckQueryValue("aReqDemo")
	if !aReqDemoOk {
		return ctx
	}
	var aReqDemoQuery endpointsIDlEndpointsBarBar.DemoType
	_tmpaReqDemoQuery, ok := req.GetQueryValue("aReqDemo")
	if ok {
		if err := aReqDemoQuery.UnmarshalText([]byte(_tmpaReqDemoQuery)); err != nil {
			req.LogAndSendQueryError(err, "enum", "aReqDemo", _tmpaReqDemoQuery)
			ok = false
		}
	}
	if !ok {
		return ctx
	}
	requestBody.AReqDemo = (endpointsIDlEndpointsBarBar.DemoType)(aReqDemoQuery)

	anOptFruitOk := req.HasQueryValue("anOptFruit")
	if anOptFruitOk {
		var anOptFruitQuery endpointsIDlEndpointsBarBar.Fruit
		_tmpanOptFruitQuery, ok := req.GetQueryValue("anOptFruit")
		if ok {
			if err := anOptFruitQuery.UnmarshalText([]byte(_tmpanOptFruitQuery)); err != nil {
				req.LogAndSendQueryError(err, "enum", "anOptFruit", _tmpanOptFruitQuery)
				ok = false
			}
		}
		if !ok {
			return ctx
		}
		requestBody.AnOptFruit = (*endpointsIDlEndpointsBarBar.Fruit)(ptr.Int32(int32(anOptFruitQuery)))
	}

	aReqFruitsOk := req.CheckQueryValue("aReqFruits")
	if !aReqFruitsOk {
		return ctx
	}
	aReqFruitsQuery, ok := req.GetQueryValueList("aReqFruits")
	if !ok {
		return ctx
	}
	aReqFruitsQueryFinal := make([]endpointsIDlEndpointsBarBar.Fruit, len(aReqFruitsQuery))
	for i, v := range aReqFruitsQuery {
		var _tmpv endpointsIDlEndpointsBarBar.Fruit
		if err := _tmpv.UnmarshalText([]byte(v)); err != nil {
			req.LogAndSendQueryError(err, "enum", "aReqFruits", v)
			return ctx
		}
		aReqFruitsQueryFinal[i] = endpointsIDlEndpointsBarBar.Fruit(_tmpv)
	}
	requestBody.AReqFruits = aReqFruitsQueryFinal

	anOptDemosOk := req.HasQueryValue("anOptDemos")
	if anOptDemosOk {
		anOptDemosQuery, ok := req.GetQueryValueList("anOptDemos")
		if !ok {
			return ctx
		}
		anOptDemosQueryFinal := make([]endpointsIDlEndpointsBarBar.DemoType, len(anOptDemosQuery))
		for i, v := range anOptDemosQuery {
			var _tmpv endpointsIDlEndpointsBarBar.DemoType
			if err := _tmpv.UnmarshalText([]byte(v)); err != nil {
				req.LogAndSendQueryError(err, "enum", "anOptDemos", v)
				return ctx
			}
			anOptDemosQueryFinal[i] = endpointsIDlEndpointsBarBar.DemoType(_tmpv)
		}
		requestBody.AnOptDemos = anOptDemosQueryFinal
	}

	// log endpoint request to downstream services
	if ce := h.Dependencies.Default.ContextLogger.Check(zapcore.DebugLevel, "stub"); ce != nil {
		zfields := []zapcore.Field{
			zap.String("endpoint", h.endpoint.EndpointName),
		}
		zfields = append(zfields, zap.String("body", fmt.Sprintf("%s", req.GetRawBody())))
		for _, k := range req.Header.Keys() {
			if val, ok := req.Header.Get(k); ok {
				zfields = append(zfields, zap.String(k, val))
			}
		}
		ctx = h.Dependencies.Default.ContextLogger.DebugZ(ctx, "endpoint request to downstream", zfields...)
	}

	w := workflow.NewBarArgWithManyQueryParamsWorkflow(h.Dependencies)
	if span := req.GetSpan(); span != nil {
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	ctx, response, cliRespHeaders, err := w.Handle(ctx, req.Header, &requestBody)

	// log downstream response to endpoint
	if ce := h.Dependencies.Default.ContextLogger.Check(zapcore.DebugLevel, "stub"); ce != nil {
		zfields := []zapcore.Field{
			zap.String("endpoint", h.endpoint.EndpointName),
		}
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
		if zErr, ok := err.(zanzibar.Error); ok {
			err = zErr.Unwrap()
		}

		res.SendError(500, "Unexpected server error", err)
		return ctx

	}

	res.WriteJSON(200, cliRespHeaders, response)
	return ctx
}
