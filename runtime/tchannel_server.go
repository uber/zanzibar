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

	tchan "github.com/uber/tchannel-go"
	netContext "golang.org/x/net/context"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/zap"
)

// PostResponseCB registers a callback that is run after a response has been
// completely processed (e.g. written to the channel).
// This gives the server a chance to clean up resources from the response object
type PostResponseCB func(ctx context.Context, method string, response RWTStruct)

type handler struct {
	tchanHandler   TChanHandler
	postResponseCB PostResponseCB
}

// TChanRouter handles incoming TChannel calls and routes them to the matching TChanHandler.
type TChanRouter struct {
	sync.RWMutex
	registrar tchan.Registrar
	logger    *zap.Logger
	handlers  map[string]handler
}

// netContextRouter implements the Handler interface that consumes netContext instead of stdlib context
type netContextRouter struct {
	router *TChanRouter
}

func (ncr netContextRouter) Handle(ctx netContext.Context, call *tchan.InboundCall) {
	ncr.router.Handle(ctx, call)
}

// NewTChanRouter returns a TChannel router that can serve thrift services over TChannel.
func NewTChanRouter(registrar tchan.Registrar, logger *zap.Logger) *TChanRouter {
	return &TChanRouter{
		registrar: registrar,
		logger:    logger,
		handlers:  map[string]handler{},
	}
}

// Register registers the given TChanHandler to be called on an incoming call for its method.
// "service" is the thrift service name as in the thrift definition.
func (s *TChanRouter) Register(service string, method string, h TChanHandler) {
	handler := &handler{tchanHandler: h}
	s.register(service, method, handler)
}

// RegisterWithPostResponseCB registers the given TChanHandler with a PostResponseCB function
func (s *TChanRouter) RegisterWithPostResponseCB(service string, method string, h TChanHandler, cb PostResponseCB) {
	handler := &handler{
		tchanHandler:   h,
		postResponseCB: cb,
	}
	s.register(service, method, handler)
}

func (s *TChanRouter) register(service string, method string, h *handler) {
	key := service + "::" + method

	s.Lock()
	s.handlers[key] = *h
	s.Unlock()

	ncr := netContextRouter{router: s}
	s.registrar.Register(ncr, key)
}

// Handle handles an incoming TChannel call and forwards it to the correct handler.
func (s *TChanRouter) Handle(ctx context.Context, call *tchan.InboundCall) {
	op := call.MethodString()
	service, method, ok := getServiceMethod(op)
	if !ok {
		s.logger.Error(fmt.Sprintf("Handle got call for %s which does not match the expected call format", op))
	}

	s.RLock()
	handler, ok := s.handlers[op]
	s.RUnlock()
	if !ok {
		s.logger.Error(fmt.Sprintf("Handle got call for %s which is not registered", op))
	}

	if err := s.handle(ctx, handler, service, method, call); err != nil {
		s.onError(err)
	}
}

func (s *TChanRouter) onError(err error) {
	if tchan.GetSystemErrorCode(err) == tchan.ErrCodeTimeout {
		s.logger.Warn("Thrift server timeout",
			zap.String("error", err.Error()),
		)
	} else {
		s.logger.Error("Thrift server error.",
			zap.String("error", err.Error()),
		)
	}
}

func (s *TChanRouter) handle(ctx context.Context, handler handler, service string, method string, call *tchan.InboundCall) error {
	reader, err := call.Arg2Reader()
	if err != nil {
		return errors.Wrapf(err, "could not create arg2reader for inbound call: %s::%s", service, method)
	}
	headers, err := ReadHeaders(reader)
	if err != nil {
		return errors.Wrapf(err, "could not reade headers for inbound call: %s::%s", service, method)
	}
	if err := EnsureEmpty(reader, "reading request headers"); err != nil {
		return errors.Wrapf(err, "could not ensure arg2reader is empty for inbound call: %s::%s", service, method)
	}

	if err := reader.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg2reader for inbound call: %s::%s", service, method)
	}

	reader, err = call.Arg3Reader()
	if err != nil {
		return errors.Wrapf(err, "could not create arg3reader for inbound call: %s::%s", service, method)
	}

	buf := GetBuffer()
	defer PutBuffer(buf)
	if _, err := buf.ReadFrom(reader); err != nil {
		return errors.Wrapf(err, "could not read from arg3reader for inbound call: %s::%s", service, method)
	}

	tracer := tchan.TracerFromRegistrar(s.registrar)
	ctx = tchan.ExtractInboundSpan(ctx, call, headers, tracer)

	wireValue, err := protocol.Binary.Decode(bytes.NewReader(buf.Bytes()), wire.TStruct)
	if err != nil {
		return errors.Wrapf(err, "could not decode arg3 for inbound call: %s::%s", service, method)
	}

	success, resp, respHeaders, err := handler.tchanHandler.Handle(ctx, headers, &wireValue)

	if handler.postResponseCB != nil {
		defer handler.postResponseCB(ctx, method, resp)
	}

	if err != nil {
		if er := reader.Close(); er != nil {
			return errors.Wrapf(er, "could not close arg3reader for inbound call: %s::%s", service, method)
		}
		return call.Response().SendSystemError(err)
	}

	if err := EnsureEmpty(reader, "reading request body"); err != nil {
		return errors.Wrapf(err, "could not ensure arg3reader is empty for inbound call: %s::%s", service, method)
	}
	if err := reader.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg3reader is empty for inbound call: %s::%s", service, method)
	}

	if !success {
		if err := call.Response().SetApplicationError(); err != nil {
			return err
		}
	}

	writer, err := call.Response().Arg2Writer()
	if err != nil {
		return errors.Wrapf(err, "could not create arg2writer for inbound call response: %s::%s", service, method)
	}

	if err := WriteHeaders(writer, respHeaders); err != nil {
		return errors.Wrapf(err, "could not write headers for inbound call response: %s::%s", service, method)
	}
	if err := writer.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg2writer for inbound call response: %s::%s", service, method)
	}

	writer, err = call.Response().Arg3Writer()
	if err != nil {
		return errors.Wrapf(err, "could not create arg3writer for inbound call response: %s::%s", service, method)
	}

	err = WriteStruct(writer, resp)
	if err != nil {
		return errors.Wrapf(err, "could not write arg3 for inbound call response: %s::%s", service, method)
	}

	if err := writer.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg3writer for inbound call response: %s::%s", service, method)
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
