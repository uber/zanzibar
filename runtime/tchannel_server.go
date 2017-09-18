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

	"github.com/pkg/errors"
	"github.com/uber/tchannel-go"
	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/zap"
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
func NewTChannelRouter(registrar tchannel.Registrar, logger *zap.Logger) *TChannelRouter {
	return &TChannelRouter{
		registrar: registrar,
		endpoints: map[string]*TChannelEndpoint{},
		logger:    logger,
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
	endpoint := TChannelEndpoint{
		EndpointID:     endpointID,
		HandlerID:      handlerID,
		Method:         method,
		handler:        h,
		postResponseCB: cb,
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
	treader, err := call.Arg2Reader()
	if err != nil {
		return errors.Wrapf(err, "could not create arg2reader for inbound call: %s", e.Method)
	}

	headers, err := ReadHeaders(treader)
	if err != nil {
		_ = treader.Close()

		return errors.Wrapf(err, "could not reade headers for inbound call: %s", e.Method)
	}
	if err := EnsureEmpty(treader, "reading request headers"); err != nil {
		_ = treader.Close()

		return errors.Wrapf(err, "could not ensure arg2reader is empty for inbound call: %s", e.Method)
	}
	if err := treader.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg2reader for inbound call: %s", e.Method)
	}

	treader, err = call.Arg3Reader()
	if err != nil {
		return errors.Wrapf(err, "could not create arg3reader for inbound call: %s", e.Method)
	}
	buf := GetBuffer()
	defer PutBuffer(buf)
	if _, err := buf.ReadFrom(treader); err != nil {
		_ = treader.Close()

		return errors.Wrapf(err, "could not read from arg3reader for inbound call: %s", e.Method)
	}

	tracer := tchannel.TracerFromRegistrar(s.registrar)
	ctx = tchannel.ExtractInboundSpan(ctx, call, headers, tracer)

	wireValue, err := protocol.Binary.Decode(bytes.NewReader(buf.Bytes()), wire.TStruct)
	if err != nil {
		return errors.Wrapf(err, "could not decode arg3 for inbound call: %s", e.Method)
	}

	success, resp, respHeaders, err := e.handler.Handle(ctx, headers, &wireValue)

	if e.postResponseCB != nil {
		defer e.postResponseCB(ctx, e.Method, resp)
	}

	if err != nil {
		if er := treader.Close(); er != nil {
			return errors.Wrapf(er, "could not close arg3reader for inbound call: %s", e.Method)
		}
		s.logger.Error("Unexpected tchannel system error",
			zap.String("method", e.Method),
			zap.String("error", err.Error()),
		)
		return call.Response().SendSystemError(errors.New("Server Error"))
	}

	if err := EnsureEmpty(treader, "reading request body"); err != nil {
		_ = treader.Close()

		return errors.Wrapf(err, "could not ensure arg3reader is empty for inbound call: %s", e.Method)
	}
	if err := treader.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg3reader is empty for inbound call: %s", e.Method)
	}

	structWireValue, err := resp.ToWire()
	if err != nil {
		// If we could not write the body then we should do something else
		// instead.

		_ = call.Response().SendSystemError(errors.New("Server Error"))

		return errors.Wrapf(err, "could not serialize arg3 for inbound call response: %s", e.Method)
	}

	if !success {
		if err := call.Response().SetApplicationError(); err != nil {
			return err
		}
	}

	twriter, err := call.Response().Arg2Writer()
	if err != nil {
		return errors.Wrapf(err, "could not create arg2writer for inbound call response: %s", e.Method)
	}

	if err := WriteHeaders(twriter, respHeaders); err != nil {
		_ = twriter.Close()

		return errors.Wrapf(err, "could not write headers for inbound call response: %s", e.Method)
	}
	if err := twriter.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg2writer for inbound call response: %s", e.Method)
	}

	twriter, err = call.Response().Arg3Writer()
	if err != nil {
		return errors.Wrapf(err, "could not create arg3writer for inbound call response: %s", e.Method)
	}

	err = protocol.Binary.Encode(structWireValue, twriter)
	if err != nil {
		_ = twriter.Close()

		return errors.Wrapf(err, "could not write arg3 for inbound call response: %s", e.Method)
	}

	if err := twriter.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg3writer for inbound call response: %s", e.Method)
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
