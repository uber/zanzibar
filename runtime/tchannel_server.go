// Copyright (c) 2017 Uber Technologies, Inc.
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
	EndpointID     string
	HandlerID      string
	Method         string
	handler        TChannelHandler
	postResponseCB PostResponseCB
	Logger         *zap.Logger
	metrics        *InboundTChannelMetrics
}

type tchannelCall struct {
	endpoint   *TChannelEndpoint
	call       *tchannel.InboundCall
	success    bool
	startTime  time.Time
	finishTime time.Time
	reqBody    []byte
	resBody    []byte
	reqHeaders map[string]string
	resHeaders map[string]string
}

// TChannelRouter handles incoming TChannel calls and routes them to the matching TChannelHandler.
type TChannelRouter struct {
	sync.RWMutex
	registrar tchannel.Registrar
	endpoints map[string]*TChannelEndpoint
	logger    *zap.Logger
	scope     tally.Scope
}

// netContextRouter implements the Handler interface that consumes netContext instead of stdlib context
type netContextRouter struct {
	router *TChannelRouter
}

func (ncr netContextRouter) Handle(ctx netContext.Context, call *tchannel.InboundCall) {
	ncr.router.Handle(ctx, call)
}

// NewTChannelRouter returns a TChannel router that can serve thrift services over TChannel.
func NewTChannelRouter(registrar tchannel.Registrar, g *Gateway) *TChannelRouter {
	return &TChannelRouter{
		registrar: registrar,
		endpoints: map[string]*TChannelEndpoint{},
		logger:    g.Logger,
		scope:     g.AllHostScope,
	}
}

// Register registers the given TChannelHandler to be called on an incoming call for its method.
// "service" is the thrift service name as in the thrift definition.
func (s *TChannelRouter) Register(
	endpointID, handlerID, method string,
	h TChannelHandler,
) *TChannelEndpoint {
	return s.RegisterWithPostResponseCB(endpointID, handlerID, method, h, nil)
}

// RegisterWithPostResponseCB registers the given TChannelHandler with a PostResponseCB function
func (s *TChannelRouter) RegisterWithPostResponseCB(
	endpointID, handlerID, method string,
	h TChannelHandler,
	cb PostResponseCB,
) *TChannelEndpoint {
	logger := s.logger.With(
		zap.String("endpointID", endpointID),
		zap.String("handlerID", endpointID),
		zap.String("method", method),
	)
	scope := s.scope.Tagged(map[string]string{
		"endpoint": endpointID,
		"handler":  handlerID,
		"method":   method,
	})
	endpoint := TChannelEndpoint{
		EndpointID:     endpointID,
		HandlerID:      handlerID,
		Method:         method,
		handler:        h,
		postResponseCB: cb,
		Logger:         logger,
		metrics:        NewInboundTChannelMetrics(scope),
	}
	s.register(&endpoint)
	return &endpoint
}

func (s *TChannelRouter) register(e *TChannelEndpoint) {
	s.Lock()
	s.endpoints[e.Method] = e
	s.Unlock()

	ncr := netContextRouter{router: s}
	s.registrar.Register(ncr, e.Method)
}

// Handle handles an incoming TChannel call and forwards it to the correct handler.
func (s *TChannelRouter) Handle(ctx context.Context, call *tchannel.InboundCall) {
	method := call.MethodString()
	if _, _, ok := getServiceMethod(method); !ok {
		s.logger.Error(fmt.Sprintf("Handle got call for %s which does not match the expected call format", method))
		return
	}

	s.RLock()
	e, ok := s.endpoints[method]
	s.RUnlock()
	if !ok {
		s.logger.Error(fmt.Sprintf("Handle got call for %s which is not registered", method))
		return
	}

	if err := s.handle(ctx, e, call); err != nil {
		s.onError(err)
	}
}

func getServiceMethod(method string) (string, string, bool) {
	s := string(method)
	sep := strings.Index(s, "::")
	if sep == -1 {
		return "", "", false
	}
	return s[:sep], s[sep+2:], true
}

func (s *TChannelRouter) onError(err error) {
	if tchannel.GetSystemErrorCode(err) == tchannel.ErrCodeTimeout {
		s.logger.Warn("Thrift server timeout", zap.Error(err))
	} else {
		s.logger.Error("Thrift server error.", zap.Error(err))
	}
}

func (s *TChannelRouter) handle(
	ctx context.Context,
	e *TChannelEndpoint,
	call *tchannel.InboundCall,
) (err error) {
	c := tchannelCall{
		endpoint: e,
		call:     call,
	}
	defer func() { c.finish(err) }()
	c.start()

	if err := c.readReqHeaders(); err != nil {
		return err
	}
	wireValue, err := c.readReqBody()
	if err != nil {
		return err
	}

	// trace request
	tracer := tchannel.TracerFromRegistrar(s.registrar)
	ctx = tchannel.ExtractInboundSpan(ctx, call, c.reqHeaders, tracer)

	resp, err := c.handle(ctx, &wireValue)
	if err != nil {
		return err
	}
	if err = c.writeResHeaders(); err != nil {
		return err
	}
	if err = c.writeResBody(resp); err != nil {
		return err
	}

	return err
}

func (c *tchannelCall) start() {
	c.startTime = time.Now()
}

func (c *tchannelCall) finish(err error) {
	c.finishTime = time.Now()

	// emit metrics
	if err != nil {
		c.endpoint.metrics.SystemErrors.Inc(1)
	} else if !c.success {
		c.endpoint.metrics.AppErrors.Inc(1)
	} else {
		c.endpoint.metrics.Success.Inc(1)
	}
	c.endpoint.metrics.Latency.Record(c.finishTime.Sub(c.startTime))

	// write logs
	c.endpoint.Logger.Info("Finished an incoming server TChannel request", c.logFields()...)

}

func (c *tchannelCall) logFields() []zapcore.Field {
	fields := []zapcore.Field{
		zap.String("remoteAddr", c.call.RemotePeer().HostPort),
		zap.String("calling-service", c.call.CallerName()),
		zap.Time("timestamp-started", c.startTime),
		zap.Time("timestamp-finished", c.finishTime),

		// TODO: Do not log body by default because PII and bandwidth.
		// Temporarily log during the developement cycle
		// TODO: Add a gateway level configurable body unmarshaller
		// to extract only non-PII info.
		zap.ByteString("Request Body", c.reqBody),
		zap.ByteString("Response Body", c.resBody),
	}

	for k, v := range c.reqHeaders {
		fields = append(fields, zap.String("Request-Header-"+k, v))
	}
	for k, v := range c.resHeaders {
		fields = append(fields, zap.String("Response-Header-"+k, v))
	}

	// TODO: log jaeger trace span

	return fields
}

// readReqHeaders reads request headers from arg2
func (c *tchannelCall) readReqHeaders() error {
	treader, err := c.call.Arg2Reader()
	if err != nil {
		c.endpoint.Logger.Error("Could not create arg2reader for inbound request", zap.Error(err))
		return errors.Wrapf(err, "Could not create arg2reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	c.reqHeaders, err = ReadHeaders(treader)
	if err != nil {
		_ = treader.Close()
		c.endpoint.Logger.Error("Could not read headers for inbound request", zap.Error(err))
		return errors.Wrapf(err, "Could not read headers for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err := EnsureEmpty(treader, "reading request headers"); err != nil {
		_ = treader.Close()
		c.endpoint.Logger.Error("Could not ensure arg2reader is empty for inbound request", zap.Error(err))
		return errors.Wrapf(err, "Could not ensure arg2reader is empty for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err := treader.Close(); err != nil {
		c.endpoint.Logger.Error("Could not close arg2reader for inbound request", zap.Error(err))
		return errors.Wrapf(err, "Could not close arg2reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	return nil
}

// readReqBody reads request body from arg3
func (c *tchannelCall) readReqBody() (wireValue wire.Value, err error) {
	treader, err := c.call.Arg3Reader()
	if err != nil {
		c.endpoint.Logger.Error("Could not create arg3reader for inbound request", zap.Error(err))
		err = errors.Wrapf(err, "Could not create arg3reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}
	buf := GetBuffer()
	defer PutBuffer(buf)
	if _, err = buf.ReadFrom(treader); err != nil {
		_ = treader.Close()
		c.endpoint.Logger.Error("Could not read from arg3reader for inbound request", zap.Error(err))
		err = errors.Wrapf(err, "Could not read from arg3reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}
	c.reqBody = buf.Bytes()
	wireValue, err = protocol.Binary.Decode(bytes.NewReader(c.reqBody), wire.TStruct)
	if err != nil {
		c.endpoint.Logger.Error("Could not decode arg3 for inbound request", zap.Error(err))
		err = errors.Wrapf(err, "Could not decode arg3 for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}
	if err = EnsureEmpty(treader, "reading request body"); err != nil {
		_ = treader.Close()
		c.endpoint.Logger.Error("Could not ensure arg3reader is empty for inbound request", zap.Error(err))
		err = errors.Wrapf(err, "Could not ensure arg3reader is empty for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}
	if err = treader.Close(); err != nil {
		c.endpoint.Logger.Error("Could not close arg3reader for inbound request", zap.Error(err))
		err = errors.Wrapf(err, "Could not close arg3reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}

	// request received when arg3reader is closed
	c.endpoint.metrics.Recvd.Inc(1)
	return
}

// handle tchannel server endpoint call
func (c *tchannelCall) handle(ctx context.Context, wireValue *wire.Value) (resp RWTStruct, err error) {
	c.success, resp, c.resHeaders, err = c.endpoint.handler.Handle(ctx, c.reqHeaders, wireValue)
	if c.endpoint.postResponseCB != nil {
		defer c.endpoint.postResponseCB(ctx, c.endpoint.Method, resp)
	}
	if err != nil {
		c.endpoint.Logger.Error("Unexpected tchannel system error", zap.Error(err))
		err = c.call.Response().SendSystemError(errors.New("Server Error"))
		return
	}
	if !c.success {
		if err = c.call.Response().SetApplicationError(); err != nil {
			c.endpoint.Logger.Error("Could not set application error for inbound response", zap.Error(err))
			return
		}
	}
	return
}

// writeResHeaders writes response headers to arg2
func (c *tchannelCall) writeResHeaders() error {
	twriter, err := c.call.Response().Arg2Writer()
	if err != nil {
		c.endpoint.Logger.Error("Could not create arg3writer for inbound response", zap.Error(err))
		return errors.Wrapf(err, "Could not create arg2writer for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err = WriteHeaders(twriter, c.resHeaders); err != nil {
		_ = twriter.Close()
		c.endpoint.Logger.Error("Could not write headers for inbound response", zap.Error(err))
		return errors.Wrapf(err, "Could not write headers for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err = twriter.Close(); err != nil {
		c.endpoint.Logger.Error("Could not close arg2writer for inbound response", zap.Error(err))
		return errors.Wrapf(err, "Could not close arg2writer for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	return nil
}

// writeResBody writes response body to arg3
func (c *tchannelCall) writeResBody(resp RWTStruct) error {
	structWireValue, err := resp.ToWire()
	if err != nil {
		// If we could not write the body then we should do something else instead.
		_ = c.call.Response().SendSystemError(errors.New("Server Error"))
		c.endpoint.Logger.Error("Could not serialize arg3 for inbound response", zap.Error(err))
		return errors.Wrapf(err, "Could not serialize arg3 for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	c.resBody = structWireValue.GetBinary()
	twriter, err := c.call.Response().Arg3Writer()
	if err != nil {
		c.endpoint.Logger.Error("Could not create arg3writer for inbound response", zap.Error(err))
		return errors.Wrapf(err, "Could not create arg3writer for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	err = protocol.Binary.Encode(structWireValue, twriter)
	if err != nil {
		_ = twriter.Close()
		c.endpoint.Logger.Error("Could not write arg3 for inbound response", zap.Error(err))
		return errors.Wrapf(err, "Could not write arg3 for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err = twriter.Close(); err != nil {
		c.endpoint.Logger.Error("Could not close arg3writer for inbound response", zap.Error(err))
		return errors.Wrapf(err, "Could not close arg3writer for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	return nil
}
