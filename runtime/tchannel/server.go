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
	"log"
	"strings"
	"sync"

	tchan "github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/thrift"

	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"
	"golang.org/x/net/context"
)

type handler struct {
	server         TChanServer
	postResponseCB PostResponseCB
}

// Server handles incoming TChannel calls and forwards them to the matching TChanServer.
type Server struct {
	sync.RWMutex
	registrar tchan.Registrar
	log       tchan.Logger
	handlers  map[string]handler
	ctxFn     func(ctx context.Context, method string, headers map[string]string) thrift.Context
}

// NewServer returns a server that can serve thrift services over TChannel.
func NewServer(registrar tchan.Registrar) *Server {
	server := &Server{
		registrar: registrar,
		log:       registrar.Logger(),
		handlers:  map[string]handler{},
		ctxFn:     defaultContextFn,
	}
	return server
}

// Register registers the given TChanServer to the be called on any incoming call for its services.
func (s *Server) Register(svr TChanServer, opts ...RegisterOption) {
	service := svr.Service()
	handler := &handler{server: svr}
	for _, opt := range opts {
		opt.Apply(handler)
	}

	s.Lock()
	s.handlers[service] = *handler
	s.Unlock()

	for _, m := range svr.Methods() {
		s.registrar.Register(s, service+"::"+m)
	}
}

// SetContextFn sets the function used to convert a context.Context to a thrift.Context.
func (s *Server) SetContextFn(f func(ctx context.Context, method string, headers map[string]string) thrift.Context) {
	s.ctxFn = f
}

func (s *Server) onError(err error) {
	if tchan.GetSystemErrorCode(err) == tchan.ErrCodeTimeout {
		s.log.Warn("Thrift server timeout:" + err.Error())
	} else {
		s.log.WithFields(tchan.ErrField(err)).Error("Thrift server error.")
	}
}

func defaultContextFn(ctx context.Context, method string, headers map[string]string) thrift.Context {
	return thrift.WithHeaders(ctx, headers)
}

func (s *Server) handle(origCtx context.Context, handler handler, method string, call *tchan.InboundCall) error {
	reader, err := call.Arg2Reader()
	if err != nil {
		return err
	}
	headers, err := ReadHeaders(reader)
	if err != nil {
		return err
	}
	if err := EnsureEmpty(reader, "reading request headers"); err != nil {
		return err
	}

	if err := reader.Close(); err != nil {
		return err
	}

	reader, err = call.Arg3Reader()
	if err != nil {
		return err
	}

	buf := GetBuffer()
	defer PutBuffer(buf)
	if _, err := buf.ReadFrom(reader); err != nil {
		return err
	}

	tracer := tchan.TracerFromRegistrar(s.registrar)
	origCtx = tchan.ExtractInboundSpan(origCtx, call, headers, tracer)
	ctx := s.ctxFn(origCtx, method, headers)

	wireValue, err := protocol.Binary.Decode(bytes.NewReader(buf.Bytes()), wire.TStruct)
	if err != nil {
		return err
	}

	// TODO: (lu) pass wireValue pointer
	success, resp, err := handler.server.Handle(ctx, method, wireValue)

	if handler.postResponseCB != nil {
		defer handler.postResponseCB(ctx, method, resp)
	}

	if err != nil {
		if er := reader.Close(); er != nil {
			return er
		}
		return call.Response().SendSystemError(err)
	}

	if err := EnsureEmpty(reader, "reading request body"); err != nil {
		return err
	}
	if err := reader.Close(); err != nil {
		return err
	}

	if !success {
		if err := call.Response().SetApplicationError(); err != nil {
			return err
		}
	}

	writer, err := call.Response().Arg2Writer()
	if err != nil {
		return err
	}

	if err := WriteHeaders(writer, ctx.ResponseHeaders()); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	writer, err = call.Response().Arg3Writer()
	if err != nil {
		return err
	}

	err = WriteStruct(writer, resp)
	if err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
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
		log.Fatalf("Handle got call for %s which does not match the expected call format", op)
	}

	s.RLock()
	handler, ok := s.handlers[service]
	s.RUnlock()
	if !ok {
		log.Fatalf("Handle got call for service %v which is not registered", service)
	}

	if err := s.handle(ctx, handler, method, call); err != nil {
		s.onError(err)
	}
}
