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

package zanzibar

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go"
	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	netContext "golang.org/x/net/context"
)

// PostResponseCB registers a callback that is run after a response has been
// completely processed (e.g. written to the channel).
// This gives the server a chance to clean up resources from the response object
type PostResponseCB func(ctx context.Context, method string, response RWTStruct)

// TChannelEndpoint handles tchannel requests
type TChannelEndpoint struct {
	TChannelHandler

	EndpointID string
	HandlerID  string
	Method     string
	Scope      tally.Scope

	// Deprecated: Use contextLogger instead.
	Logger *zap.Logger

	contextLogger  ContextLogger
	ContextMetrics ContextMetrics
	callback       PostResponseCB
}

type tchannelInboundCall struct {
	ctx        context.Context
	endpoint   *TChannelEndpoint
	call       *tchannel.InboundCall
	success    bool
	responded  bool
	startTime  time.Time
	finishTime time.Time
	reqHeaders map[string]string
	resHeaders map[string]string
}

// TChannelRouter handles incoming TChannel calls and routes them to the matching TChannelHandler.
type TChannelRouter struct {
	sync.RWMutex
	registrar tchannel.Registrar
	endpoints map[string]*TChannelEndpoint
	logger    ContextLogger
	metrics   ContextMetrics
	extractor ContextExtractor
}

// netContextRouter implements the Handler interface that consumes netContext instead of stdlib context
type netContextRouter struct {
	router *TChannelRouter
}

func (ncr netContextRouter) Handle(ctx netContext.Context, call *tchannel.InboundCall) {
	ncr.router.Handle(ctx, call)
}

// NewTChannelEndpoint creates a new tchannel endpoint to handle an incoming
// call for its method.
func NewTChannelEndpoint(
	logger *zap.Logger,
	scope tally.Scope,
	endpointID, handlerID, method string,
	handler TChannelHandler,
) *TChannelEndpoint {
	return NewTChannelEndpointWithPostResponseCB(
		logger, scope,
		endpointID, handlerID, method,
		handler, nil,
	)
}

// NewTChannelEndpointWithPostResponseCB creates a new tchannel endpoint,
// with or without a post response callback function.
func NewTChannelEndpointWithPostResponseCB(
	logger *zap.Logger,
	scope tally.Scope,
	endpointID, handlerID, method string,
	handler TChannelHandler,
	callback PostResponseCB,
) *TChannelEndpoint {
	contextLogger := NewContextLogger(logger)

	return &TChannelEndpoint{
		TChannelHandler: handler,
		EndpointID:      endpointID,
		HandlerID:       handlerID,
		Method:          method,
		callback:        callback,
		Logger:          logger,
		Scope:           scope,
		contextLogger:   contextLogger,
	}
}

// NewTChannelRouter returns a TChannel router that can serve thrift services over TChannel.
func NewTChannelRouter(registrar tchannel.Registrar, g *Gateway) *TChannelRouter {
	return &TChannelRouter{
		registrar: registrar,
		endpoints: map[string]*TChannelEndpoint{},
		logger:    g.ContextLogger,
		metrics:   g.ContextMetrics,
		extractor: g.ContextExtractor,
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
	e.ContextMetrics = s.metrics
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
		s.logger.Error(ctx, "Handle got call for which does not match the expected call format", zap.String(logFieldRequestMethod, method))
		return
	}

	ctx = withRequestFields(ctx)

	s.RLock()
	e, ok := s.endpoints[method]
	s.RUnlock()
	if !ok {
		s.logger.Error(ctx, "Handle got call for method which is not registered",
			zap.String(logFieldRequestMethod, method),
		)
		return
	}

	ctx = WithEndpointField(ctx, e.EndpointID)
	ctx = WithLogFields(ctx,
		zap.String(logFieldEndpointID, e.EndpointID),
		zap.String(logFieldHandlerID, e.HandlerID),
		zap.String(logFieldRequestMethod, e.Method),
	)

	scopeTags := map[string]string{
		scopeTagEndpoint:       e.EndpointID,
		scopeTagHandler:        e.HandlerID,
		scopeTagEndpointMethod: e.Method,
		scopeTagProtocal:       scopeTagTChannel,
	}

	ctx = WithScopeTags(ctx, scopeTags)

	var err error
	errc := make(chan error, 1)
	c := tchannelInboundCall{
		endpoint: e,
		call:     call,
		ctx:      ctx,
	}

	c.start()
	go func() { errc <- s.handle(ctx, &c) }()
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

	c.finish(ctx, err)
}

func (s *TChannelRouter) handle(
	ctx context.Context,
	c *tchannelInboundCall,
) (err error) {
	// read request
	if err = c.readReqHeaders(ctx); err != nil {
		return err
	}

	scopeTags := make(map[string]string)
	c.ctx = WithEndpointRequestHeadersField(c.ctx, c.reqHeaders)
	for k, v := range s.extractor.ExtractScopeTags(c.ctx) {
		scopeTags[k] = v
	}

	c.ctx = WithScopeTags(c.ctx, scopeTags)
	c.endpoint.ContextMetrics.IncCounter(c.ctx, endpointRequest, 1)
	wireValue, err := c.readReqBody(ctx)
	if err != nil {
		return err
	}

	// trace request
	tracer := tchannel.TracerFromRegistrar(s.registrar)
	ctx = tchannel.ExtractInboundSpan(ctx, c.call, c.reqHeaders, tracer)

	// handle request
	resp, err := c.handle(ctx, &wireValue)
	if err != nil {
		return err
	}

	// write response
	if err = c.writeResHeaders(ctx); err != nil {
		return err
	}
	if err = c.writeResBody(ctx, resp); err != nil {
		return err
	}

	return err
}

func (c *tchannelInboundCall) start() {
	c.startTime = time.Now()
	c.endpoint.ContextMetrics.IncCounter(c.ctx, endpointRequest, 1)
}

func (c *tchannelInboundCall) finish(ctx context.Context, err error) {
	c.finishTime = time.Now()

	// emit metrics
	if err != nil {
		errCause := tchannel.GetSystemErrorCode(errors.Cause(err))
		scopeTags := map[string]string{scopeTagError: errCause.MetricsKey()}
		c.ctx = WithScopeTags(c.ctx, scopeTags)
		c.endpoint.ContextMetrics.IncCounter(c.ctx, endpointSystemErrors, 1)
	} else if !c.success {
		c.endpoint.ContextMetrics.IncCounter(c.ctx, endpointAppErrors, 1)
	} else {
		c.endpoint.ContextMetrics.IncCounter(c.ctx, endpointSuccess, 1)
	}
	c.endpoint.ContextMetrics.RecordTimer(c.ctx, endpointLatency, c.finishTime.Sub(c.startTime))

	// write logs
	if err == nil {
		c.endpoint.contextLogger.Info(c.ctx, "Finished an incoming server TChannel request", c.logFields()...)
	} else {
		c.endpoint.contextLogger.Warn(c.ctx, "Failed to serve incoming TChannel request", zap.Error(err))
	}
}

func (c *tchannelInboundCall) logFields() []zapcore.Field {
	fields := []zapcore.Field{
		zap.String("remoteAddr", c.call.RemotePeer().HostPort),
		zap.String("calling-service", c.call.CallerName()),
		zap.Time("timestamp-started", c.startTime),
		zap.Time("timestamp-finished", c.finishTime),
	}

	for k, v := range c.reqHeaders {
		fields = append(fields, zap.String("Request-Header-"+k, v))
	}
	for k, v := range c.resHeaders {
		fields = append(fields, zap.String("Response-Header-"+k, v))
	}

	return fields
}

// readReqHeaders reads request headers from arg2
func (c *tchannelInboundCall) readReqHeaders(ctx context.Context) error {
	// fail fast if timed out
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		return context.DeadlineExceeded
	}

	treader, err := c.call.Arg2Reader()
	if err != nil {
		return errors.Wrapf(err, "Could not create arg2reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	c.reqHeaders, err = ReadHeaders(treader)
	if err != nil {
		_ = treader.Close()
		return errors.Wrapf(err, "Could not read headers for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err := EnsureEmpty(treader, "reading request headers"); err != nil {
		_ = treader.Close()
		return errors.Wrapf(err, "Could not ensure arg2reader is empty for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err := treader.Close(); err != nil {
		return errors.Wrapf(err, "Could not close arg2reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}

	return nil
}

// readReqBody reads request body from arg3
func (c *tchannelInboundCall) readReqBody(ctx context.Context) (wireValue wire.Value, err error) {
	// fail fast if timed out
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		err = context.DeadlineExceeded
		return
	}

	treader, err := c.call.Arg3Reader()
	if err != nil {
		err = errors.Wrapf(err, "Could not create arg3reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}
	buf := GetBuffer()
	defer PutBuffer(buf)
	if _, err = buf.ReadFrom(treader); err != nil {
		_ = treader.Close()
		err = errors.Wrapf(err, "Could not read from arg3reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}
	wireValue, err = protocol.Binary.Decode(bytes.NewReader(buf.Bytes()), wire.TStruct)
	if err != nil {
		c.endpoint.contextLogger.Warn(ctx, "Could not decode arg3 for inbound request", zap.Error(err))
		err = errors.Wrapf(err, "Could not decode arg3 for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}
	if err = EnsureEmpty(treader, "reading request body"); err != nil {
		_ = treader.Close()
		err = errors.Wrapf(err, "Could not ensure arg3reader is empty for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}
	if err = treader.Close(); err != nil {
		err = errors.Wrapf(err, "Could not close arg3reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}

	return
}

// handle tchannel server endpoint call
func (c *tchannelInboundCall) handle(ctx context.Context, wireValue *wire.Value) (resp RWTStruct, err error) {
	// fail fast if timed out
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		err = context.DeadlineExceeded
		return
	}

	c.success, resp, c.resHeaders, err = c.endpoint.Handle(ctx, c.reqHeaders, wireValue)
	if c.endpoint.callback != nil {
		defer c.endpoint.callback(ctx, c.endpoint.Method, resp)
	}
	if err != nil {
		c.endpoint.contextLogger.Warn(c.ctx, "Unexpected tchannel system error", zap.Error(err))
		if er := c.call.Response().SendSystemError(errors.New("Server Error")); er != nil {
			c.endpoint.contextLogger.Warn(c.ctx, "Error sending server error response", zap.Error(er))
		}
		return
	}
	if !c.success {
		if err = c.call.Response().SetApplicationError(); err != nil {
			c.endpoint.contextLogger.Warn(c.ctx, "Could not set application error for inbound response", zap.Error(err))
			return
		}
	}
	return
}

// writeResHeaders writes response headers to arg2
func (c *tchannelInboundCall) writeResHeaders(ctx context.Context) error {
	// fail fast if timed out
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		return context.DeadlineExceeded
	}

	twriter, err := c.call.Response().Arg2Writer()
	if err != nil {
		return errors.Wrapf(err, "Could not create arg2writer for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err = WriteHeaders(twriter, c.resHeaders); err != nil {
		_ = twriter.Close()
		return errors.Wrapf(err, "Could not write headers for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err = twriter.Close(); err != nil {
		return errors.Wrapf(err, "Could not close arg2writer for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	return nil
}

// writeResBody writes response body to arg3
func (c *tchannelInboundCall) writeResBody(ctx context.Context, resp RWTStruct) error {
	// fail fast if timed out
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		return context.DeadlineExceeded
	}

	structWireValue, err := resp.ToWire()
	if err != nil {
		if er := c.call.Response().SendSystemError(errors.New("Server Error")); er != nil {
			c.endpoint.contextLogger.Warn(c.ctx, "Error sending server error response", zap.Error(er))
		}
		return errors.Wrapf(err, "Could not serialize arg3 for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}

	twriter, err := c.call.Response().Arg3Writer()
	if err != nil {
		return errors.Wrapf(err, "Could not create arg3writer for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	err = protocol.Binary.Encode(structWireValue, twriter)
	if err != nil {
		_ = twriter.Close()
		return errors.Wrapf(err, "Could not write arg3 for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	c.responded = true
	if err = twriter.Close(); err != nil {
		return errors.Wrapf(err, "Could not close arg3writer for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	return nil
}
