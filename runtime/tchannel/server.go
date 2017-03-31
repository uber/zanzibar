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

package tchannel

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	tchan "github.com/uber/tchannel-go"
	netContext "golang.org/x/net/context"

	"github.com/pkg/errors"
	"github.com/uber-go/zap"
	"github.com/uber/zanzibar/runtime"
	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"
)

// PostResponseCB registers a callback that is run after a response has been
// completely processed (e.g. written to the channel).
// This gives the server a chance to clean up resources from the response object
type PostResponseCB func(ctx context.Context, method string, response RWTStruct)

type handler struct {
	server         TChanServer
	postResponseCB PostResponseCB
}

// Server handles incoming TChannel calls and forwards them to the matching TChanServer.
type Server struct {
	sync.RWMutex
	registrar tchan.Registrar
	logger    zap.Logger
	handlers  map[string]handler
}

// netContextServer implements the Handler interface that consumes netContext instead of stdlib context
type netContextServer struct {
	server *Server
}

func (ncs netContextServer) Handle(ctx netContext.Context, call *tchan.InboundCall) {
	ncs.server.Handle(ctx, call)
}

// NewServer returns a server that can serve thrift services over TChannel.
func NewServer(registrar tchan.Registrar, gateway *zanzibar.Gateway) *Server {
	server := &Server{
		registrar: registrar,
		logger:    gateway.Logger,
		handlers:  map[string]handler{},
	}
	return server
}

func (s *Server) register(svr TChanServer, h *handler) {
	service := svr.Service()
	s.Lock()
	s.handlers[service] = *h
	s.Unlock()

	ncs := netContextServer{server: s}
	for _, m := range svr.Methods() {
		s.registrar.Register(ncs, service+"::"+m)
	}
}

// Register registers the given TChanServer to the be called on any incoming call for its services.
func (s *Server) Register(svr TChanServer) {
	handler := &handler{server: svr}
	s.register(svr, handler)
}

// RegisterWithPostResponseCB registers the given TChanServer with a PostResponseCB function
func (s *Server) RegisterWithPostResponseCB(svr TChanServer, cb PostResponseCB) {
	handler := &handler{
		server:         svr,
		postResponseCB: cb,
	}
	s.register(svr, handler)
}

func (s *Server) onError(err error) {
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

func (s *Server) handle(ctx context.Context, handler handler, method string, call *tchan.InboundCall) error {
	serviceName := handler.server.Service()

	reader, err := call.Arg2Reader()
	if err != nil {
		return errors.Wrapf(err, "could not create arg2reader for inbound call: %s::%s", serviceName, method)
	}
	headers, err := ReadHeaders(reader)
	if err != nil {
		return errors.Wrapf(err, "could not reade headers for inbound call: %s::%s", serviceName, method)
	}
	if err := EnsureEmpty(reader, "reading request headers"); err != nil {
		return errors.Wrapf(err, "could not ensure arg2reader is empty for inbound call: %s::%s", serviceName, method)
	}

	if err := reader.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg2reader for inbound call: %s::%s", serviceName, method)
	}

	reader, err = call.Arg3Reader()
	if err != nil {
		return errors.Wrapf(err, "could not create arg3reader for inbound call: %s::%s", serviceName, method)
	}

	buf := GetBuffer()
	defer PutBuffer(buf)
	if _, err := buf.ReadFrom(reader); err != nil {
		return errors.Wrapf(err, "could not read from arg3reader for inbound call: %s::%s", serviceName, method)
	}

	tracer := tchan.TracerFromRegistrar(s.registrar)
	ctx = tchan.ExtractInboundSpan(ctx, call, headers, tracer)

	wireValue, err := protocol.Binary.Decode(bytes.NewReader(buf.Bytes()), wire.TStruct)
	if err != nil {
		return errors.Wrapf(err, "could not decode arg3 for inbound call: %s::%s", serviceName, method)
	}

	success, respHeaders, resp, err := handler.server.Handle(ctx, method, &wireValue)

	if handler.postResponseCB != nil {
		defer handler.postResponseCB(ctx, method, resp)
	}

	if err != nil {
		if er := reader.Close(); er != nil {
			return errors.Wrapf(er, "could not close arg3reader for inbound call: %s::%s", serviceName, method)
		}
		return call.Response().SendSystemError(err)
	}

	if err := EnsureEmpty(reader, "reading request body"); err != nil {
		return errors.Wrapf(err, "could not ensure arg3reader is empty for inbound call: %s::%s", serviceName, method)
	}
	if err := reader.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg3reader is empty for inbound call: %s::%s", serviceName, method)
	}

	if !success {
		if err := call.Response().SetApplicationError(); err != nil {
			return err
		}
	}

	writer, err := call.Response().Arg2Writer()
	if err != nil {
		return errors.Wrapf(err, "could not create arg2writer for inbound call response: %s::%s", serviceName, method)
	}

	if err := WriteHeaders(writer, respHeaders); err != nil {
		return errors.Wrapf(err, "could not write headers for inbound call response: %s::%s", serviceName, method)
	}
	if err := writer.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg2writer for inbound call response: %s::%s", serviceName, method)
	}

	writer, err = call.Response().Arg3Writer()
	if err != nil {
		return errors.Wrapf(err, "could not create arg3writer for inbound call response: %s::%s", serviceName, method)
	}

	err = WriteStruct(writer, resp)
	if err != nil {
		return errors.Wrapf(err, "could not write arg3 for inbound call response: %s::%s", serviceName, method)
	}

	if err := writer.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg3writer for inbound call response: %s::%s", serviceName, method)
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

// Handle handles an incoming TChannel call and forwards it to the correct handler.
func (s *Server) Handle(ctx context.Context, call *tchan.InboundCall) {
	op := call.MethodString()
	service, method, ok := getServiceMethod(op)
	if !ok {
		s.logger.Error(fmt.Sprintf("Handle got call for %s which does not match the expected call format", op))
	}

	s.RLock()
	handler, ok := s.handlers[service]
	s.RUnlock()
	if !ok {
		s.logger.Error(fmt.Sprintf("Handle got call for service %s which is not registered", service))
	}

	if err := s.handle(ctx, handler, method, call); err != nil {
		s.onError(err)
	}
}
