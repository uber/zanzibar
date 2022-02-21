// Copyright (c) 2022 Uber Technologies, Inc.
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

package zanzibar

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	stream "go.uber.org/thriftrw/protocol/stream"

	"github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"github.com/uber-go/tally"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/tchannel-go"
	"go.uber.org/zap"
	netContext "golang.org/x/net/context"
)

// PostResponseCB registers a callback that is run after a response has been
// completely processed (e.g. written to the channel).
// This gives the server a chance to clean up resources from the response object
type PostResponseCB func(ctx context.Context, method string, response RWTStruct)

// TChannelEndpoint wraps over a TChannelHandler and can be registered to a TChannelRouter
// to handle tchannel inbound call. It only has one Handle method which is delegated to the
// embedded TChannelHandler.
type TChannelEndpoint struct {
	TChannelHandler

	EndpointID string
	HandlerID  string
	Method     string

	callback PostResponseCB
}

// TChannelRouter handles incoming TChannel calls and routes them to the matching TChannelHandler.
type TChannelRouter struct {
	sync.RWMutex
	registrar     tchannel.Registrar
	endpoints     map[string]*TChannelEndpoint
	contextLogger ContextLogger
	scope         tally.Scope
	extractor     ContextExtractor

	requestUUIDHeaderKey string
}

// netContextRouter implements the Handle interface that consumes netContext instead of stdlib context
type netContextRouter struct {
	router *TChannelRouter
}

func (ncr netContextRouter) Handle(ctx netContext.Context, call *tchannel.InboundCall) {
	ncr.router.Handle(ctx, call)
}

// NewTChannelEndpoint creates a new tchannel endpoint to handle an incoming
// call for its method.
func NewTChannelEndpoint(
	endpointID, handlerID, method string,
	handler TChannelHandler,
) *TChannelEndpoint {
	return NewTChannelEndpointWithPostResponseCB(
		endpointID, handlerID, method,
		handler, nil,
	)
}

// NewTChannelEndpointWithPostResponseCB creates a new tchannel endpoint,
// with or without a post response callback function.
func NewTChannelEndpointWithPostResponseCB(
	endpointID, handlerID, method string,
	handler TChannelHandler,
	callback PostResponseCB,
) *TChannelEndpoint {
	return &TChannelEndpoint{
		TChannelHandler: handler,
		EndpointID:      endpointID,
		HandlerID:       handlerID,
		Method:          method,
		callback:        callback,
	}
}

// NewTChannelRouter returns a TChannel router that can serve thrift services over TChannel.
func NewTChannelRouter(registrar tchannel.Registrar, g *Gateway) *TChannelRouter {
	return &TChannelRouter{
		registrar:     registrar,
		endpoints:     map[string]*TChannelEndpoint{},
		contextLogger: g.ContextLogger,
		scope:         g.RootScope,
		extractor:     g.ContextExtractor,

		requestUUIDHeaderKey: g.requestUUIDHeaderKey,
	}
}

// Register registers the given TChannelEndpoint.
func (s *TChannelRouter) Register(e *TChannelEndpoint) error {
	s.RLock()
	if _, ok := s.endpoints[e.Method]; ok {
		s.RUnlock()
		return fmt.Errorf("handler for '%s' is already registered", e.Method)
	}
	s.RUnlock()
	s.Lock()
	s.endpoints[e.Method] = e
	s.Unlock()

	ncr := netContextRouter{router: s}
	s.registrar.Register(ncr, e.Method)
	return nil
}

// Handle handles an incoming TChannel call and forwards it to the correct handler.
func (s *TChannelRouter) Handle(ctx context.Context, call *tchannel.InboundCall) {
	method := call.MethodString()
	if sep := strings.Index(method, "::"); sep == -1 {
		s.contextLogger.Error(ctx, "Handle got call for which does not match the expected call format", zap.String(logFieldRequestMethod, method))
		return
	}

	s.RLock()
	e, ok := s.endpoints[method]
	s.RUnlock()
	if !ok {
		s.contextLogger.Error(ctx, "Handle got call for method which is not registered",
			zap.String(logFieldRequestMethod, method),
		)
		return
	}

	// put log fields on the context
	logFields := []zap.Field{
		zap.String(logFieldEndpointID, e.EndpointID),
		zap.String(logFieldEndpointHandler, e.HandlerID),
		zap.String(logFieldRequestMethod, e.Method),
	}
	ctx = WithLogFields(ctx, logFields...)

	// put scope tags on the context
	scopeTags := map[string]string{
		scopeTagEndpoint:       e.EndpointID,
		scopeTagHandler:        e.HandlerID,
		scopeTagEndpointMethod: e.Method,
		scopeTagProtocol:       scopeTagTChannel,
	}
	ctx = WithScopeTags(ctx, scopeTags)

	var err error
	c := &tchannelInboundCall{
		call:          call,
		endpoint:      e,
		contextLogger: s.contextLogger,
		scope:         s.scope.Tagged(scopeTags),
	}

	c.start()
	ctx, err = s.handleHeader(ctx, c)
	defer func() { c.finish(ctx, err) }()
	if err != nil {
		return
	}

	errc := make(chan error, 1)
	go func() { errc <- s.handleBody(ctx, c) }()
	select {
	case <-ctx.Done():
		err = ctx.Err()
		if err == context.Canceled {
			// check if context was Canceled due to handle response
			if c.responded {
				err = <-errc
			}
		}
	case err = <-errc:
	}
}

func (s *TChannelRouter) handleHeader(
	ctx context.Context,
	c *tchannelInboundCall,
) (context.Context, error) {
	if err := c.readReqHeaders(ctx); err != nil {
		return ctx, err
	}

	reqUUID, ok := c.reqHeaders[s.requestUUIDHeaderKey]
	if !ok {
		reqUUID = uuid.New()
	}
	ctx = withRequestUUID(ctx, reqUUID)

	// put request headers on context so that user-provided extractor
	// functions can choose to have certain headers as metric tags or
	// log fields
	ctx = WithEndpointRequestHeadersField(ctx, c.reqHeaders)

	// use user-provided extractor function to decide metric tags
	scopeTags := make(map[string]string)
	for k, v := range s.extractor.ExtractScopeTags(ctx) {
		scopeTags[k] = v
	}
	ctx = WithScopeTags(ctx, scopeTags)
	if len(scopeTags) != 0 {
		c.scope = c.scope.Tagged(scopeTags)
	}

	// use user-provided extractor function to decide log fields
	logFields := s.extractor.ExtractLogFields(ctx)
	logFields = append(logFields, zap.String(logFieldRequestUUID, reqUUID))
	ctx = WithLogFields(ctx, logFields...)

	return ctx, nil
}

func (s *TChannelRouter) handleBody(
	ctx context.Context,
	c *tchannelInboundCall,
) (err error) {
	sr, err := c.readReqBody(ctx)
	defer func(sr stream.Reader) {
		e := sr.Close()
		if e != nil {
			err = errors.Wrapf(e, "Could not close stream writer for outbound %s.%s (%s) response",
				c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method)
		}
	}(sr)
	if err != nil {
		return err
	}

	// trace request
	tracer := tchannel.TracerFromRegistrar(s.registrar)
	ctx = tchannel.ExtractInboundSpan(ctx, c.call, c.reqHeaders, tracer)

	if s := opentracing.SpanFromContext(ctx); s != nil {
		if jaegerCtx, ok := s.Context().(jaeger.SpanContext); ok {
			logFields := make([]zap.Field, 3)
			logFields[0] = zap.String(TraceIDKey, jaegerCtx.TraceID().String())
			logFields[1] = zap.String(TraceSpanKey, jaegerCtx.SpanID().String())
			logFields[2] = zap.Bool(TraceSampledKey, jaegerCtx.IsSampled())
			ctx = WithLogFields(ctx, logFields...)
		}
	}

	// handle request
	resp, err := c.handle(ctx, sr)
	if err != nil {
		return err
	}

	// TODO: put response headers on ctx for final metrics and logs
	// ctx = WithEndpointResponseHeadersField(ctx, c.resHeaders)

	// write response
	if err = c.writeResHeaders(ctx); err != nil {
		return err
	}
	if err = c.writeResBody(ctx, resp); err != nil {
		return err
	}

	return err
}
