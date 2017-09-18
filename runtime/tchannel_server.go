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
}

// TChannelRouter handles incoming TChannel calls and routes them to the matching TChannelHandler.
type TChannelRouter struct {
	sync.RWMutex
	registrar tchannel.Registrar
	endpoints map[string]*TChannelEndpoint
	logger    *zap.Logger
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
	endpoint := TChannelEndpoint{
		EndpointID:     endpointID,
		HandlerID:      handlerID,
		Method:         method,
		handler:        h,
		postResponseCB: cb,
		Logger:         logger,
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
) error {
	var success bool
	var startTime, finishTime time.Time
	var reqBody, resBody []byte
	var reqHeaders, resHeaders map[string]string

	defer func() {
		// finish
		finishTime = time.Now()

		// write logs
		e.Logger.Info("Finished an incoming server TChannel request",
			serverTChannelLogFields(call, startTime, finishTime, reqBody, resBody, reqHeaders, resHeaders)...,
		)
	}()

	// start
	startTime = time.Now()

	// read request headers
	treader, err := call.Arg2Reader()
	if err != nil {
		e.Logger.Error("Could not create arg2reader for inbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not create arg2reader for inbound %s.%s (%s) request",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}
	reqHeaders, err = ReadHeaders(treader)
	if err != nil {
		_ = treader.Close()
		e.Logger.Error("Could not read headers for inbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not read headers for inbound %s.%s (%s) request",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}
	if err := EnsureEmpty(treader, "reading request headers"); err != nil {
		_ = treader.Close()
		e.Logger.Error("Could not ensure arg2reader is empty for inbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not ensure arg2reader is empty for inbound %s.%s (%s) request",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}
	if err := treader.Close(); err != nil {
		e.Logger.Error("Could not close arg2reader for inbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not close arg2reader for inbound %s.%s (%s) request",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}

	// read request body
	treader, err = call.Arg3Reader()
	if err != nil {
		e.Logger.Error("Could not create arg3reader for inbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not create arg3reader for inbound %s.%s (%s) request",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}
	buf := GetBuffer()
	defer PutBuffer(buf)
	if _, err := buf.ReadFrom(treader); err != nil {
		_ = treader.Close()
		e.Logger.Error("Could not read from arg3reader for inbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not read from arg3reader for inbound %s.%s (%s) request",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}
	reqBody = buf.Bytes()
	wireValue, err := protocol.Binary.Decode(bytes.NewReader(reqBody), wire.TStruct)
	if err != nil {
		e.Logger.Error("Could not decode arg3 for inbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not decode arg3 for inbound %s.%s (%s) request",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}

	// trace request
	tracer := tchannel.TracerFromRegistrar(s.registrar)
	ctx = tchannel.ExtractInboundSpan(ctx, call, reqHeaders, tracer)

	// handle request
	var resp RWTStruct
	success, resp, resHeaders, err = e.handler.Handle(ctx, reqHeaders, &wireValue)
	if e.postResponseCB != nil {
		defer e.postResponseCB(ctx, e.Method, resp)
	}
	if err != nil {
		_ = treader.Close()
		e.Logger.Error("Unexpected tchannel system error", zap.Error(err))
		return call.Response().SendSystemError(errors.New("Server Error"))
	}
	if !success {
		if err = call.Response().SetApplicationError(); err != nil {
			_ = treader.Close()
			e.Logger.Error("Could not set application error for inbound response", zap.Error(err))
			return err
		}
	}

	// close request body
	if err := EnsureEmpty(treader, "reading request body"); err != nil {
		_ = treader.Close()
		e.Logger.Error("Could not ensure arg3reader is empty for inbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not ensure arg3reader is empty for inbound %s.%s (%s) request",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}
	if err := treader.Close(); err != nil {
		e.Logger.Error("Could not close arg3reader for inbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not close arg3reader for inbound %s.%s (%s) request",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}

	// write response headers
	twriter, err := call.Response().Arg2Writer()
	if err != nil {
		e.Logger.Error("Could not create arg3writer for inbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not create arg2writer for inbound %s.%s (%s) response",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}
	if err := WriteHeaders(twriter, resHeaders); err != nil {
		_ = twriter.Close()
		e.Logger.Error("Could not write headers for inbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not write headers for inbound %s.%s (%s) response",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}
	if err := twriter.Close(); err != nil {
		e.Logger.Error("Could not close arg2writer for inbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not close arg2writer for inbound %s.%s (%s) response",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}

	// write response body
	structWireValue, err := resp.ToWire()
	if err != nil {
		// If we could not write the body then we should do something else instead.
		_ = call.Response().SendSystemError(errors.New("Server Error"))
		e.Logger.Error("Could not serialize arg3 for inbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not serialize arg3 for inbound %s.%s (%s) response",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}
	resBody = structWireValue.GetBinary()
	twriter, err = call.Response().Arg3Writer()
	if err != nil {
		e.Logger.Error("Could not create arg3writer for inbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not create arg3writer for inbound %s.%s (%s) response",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}
	err = protocol.Binary.Encode(structWireValue, twriter)
	if err != nil {
		_ = twriter.Close()
		e.Logger.Error("Could not write arg3 for inbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not write arg3 for inbound %s.%s (%s) response",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}
	if err := twriter.Close(); err != nil {
		e.Logger.Error("Could not close arg3writer for inbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not close arg3writer for inbound %s.%s (%s) response",
			e.EndpointID, e.HandlerID, e.Method,
		)
	}

	return nil
}

func getServiceMethod(method string) (string, string, bool) {
	s := string(method)
	sep := strings.Index(s, "::")
	if sep == -1 {
		return "", "", false
	}
	return s[:sep], s[sep+2:], true
}

func serverTChannelLogFields(
	call *tchannel.InboundCall,
	startTime, finishTime time.Time,
	reqBody, resBody []byte,
	reqHeaders, resHeaders map[string]string,
) []zapcore.Field {
	fields := []zapcore.Field{
		zap.String("remoteAddr", call.RemotePeer().HostPort),
		zap.String("calling-service", call.CallerName()),
		zap.Time("timestamp-started", startTime),
		zap.Time("timestamp-finished", finishTime),

		// TODO: Do not log body by default because PII and bandwidth.
		// Temporarily log during the developement cycle
		// TODO: Add a gateway level configurable body unmarshaller
		// to extract only non-PII info.
		zap.ByteString("Request Body", reqBody),
		zap.ByteString("Response Body", resBody),
	}

	for k, v := range reqHeaders {
		fields = append(fields, zap.String("Request-Header-"+k, v))
	}
	for k, v := range resHeaders {
		fields = append(fields, zap.String("Response-Header-"+k, v))
	}

	// TODO: log jaeger trace span

	return fields
}
