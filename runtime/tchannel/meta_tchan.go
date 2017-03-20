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
	"errors"
	"fmt"

	"github.com/uber/tchannel-go/thrift"

	// meta_tchan.go in package tchannel instead of meta to avoid circular dependency
	gen "github.com/uber/zanzibar/runtime/tchannel/gen-code/meta"
	"go.uber.org/thriftrw/wire"
)

// Interfaces for the service and client for the services defined in the IDL.

// tchanMeta is the interface that defines the server handler and client interface.
type tchanMeta interface {
	Health(ctx thrift.Context) (*gen.HealthStatus, error)
	VersionInfo(ctx thrift.Context) (*gen.VersionInfo, error)
}

// Implementation of a client and service handler.

type tchanMetaClient struct {
	thriftService string
	client        TChanClient
}

func newTChanMetaClient(client TChanClient) tchanMeta {
	return &tchanMetaClient{
		"Meta",
		client,
	}
}

func (c *tchanMetaClient) Health(ctx thrift.Context) (*gen.HealthStatus, error) {
	var resp gen.Meta_Health_Result
	args := gen.Meta_Health_Args{}
	success, err := c.client.Call(ctx, c.thriftService, "health", &args, &resp)
	if err == nil && !success {
		err = errors.New("received no result or unknown exception for health")
	}
	if err != nil {
		return nil, err
	}

	return gen.Meta_Health_Helper.UnwrapResponse(&resp)
}

func (c *tchanMetaClient) VersionInfo(ctx thrift.Context) (*gen.VersionInfo, error) {
	var resp gen.Meta_VersionInfo_Result
	args := gen.Meta_VersionInfo_Args{}
	success, err := c.client.Call(ctx, c.thriftService, "versionInfo", &args, &resp)
	if err == nil && !success {
		err = errors.New("received no result or unknown exception for versionInfo")
	}
	if err != nil {
		return nil, err
	}

	return gen.Meta_VersionInfo_Helper.UnwrapResponse(&resp)
}

type tchanMetaServer struct {
	handler tchanMeta
}

func newTChanMetaServer(handler tchanMeta) TChanServer {
	return &tchanMetaServer{
		handler,
	}
}

func (s *tchanMetaServer) Service() string {
	return "Meta"
}

func (s *tchanMetaServer) Methods() []string {
	return []string{
		"health",
		"thriftIDL",
		"versionInfo",
	}
}

func (s *tchanMetaServer) Handle(ctx thrift.Context, methodName string, wireValue wire.Value) (bool, RWTStruct, error) {
	switch methodName {
	case "health":
		return s.handleHealth(ctx, wireValue)
	case "versionInfo":
		return s.handleVersionInfo(ctx, wireValue)
	default:
		return false, nil, fmt.Errorf("method %v not found in service %v", methodName, s.Service())
	}
}

func (s *tchanMetaServer) handleHealth(ctx thrift.Context, wireValue wire.Value) (bool, RWTStruct, error) {
	var req gen.Meta_Health_Args
	var res gen.Meta_Health_Result

	if err := req.FromWire(wireValue); err != nil {
		return false, nil, err
	}

	r, err := s.handler.Health(ctx)

	if err != nil {
		return false, nil, err
	}
	res.Success = r

	return err == nil, &res, nil
}

func (s *tchanMetaServer) handleVersionInfo(ctx thrift.Context, wireValue wire.Value) (bool, RWTStruct, error) {
	var req gen.Meta_VersionInfo_Args
	var res gen.Meta_VersionInfo_Result

	if err := req.FromWire(wireValue); err != nil {
		return false, nil, err
	}

	r, err := s.handler.VersionInfo(ctx)

	if err != nil {
		return false, nil, err
	}
	res.Success = r

	return err == nil, &res, nil
}
