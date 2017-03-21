// TODO: this file should be generated thriftrw-gen, which is yet to be implemented.

// Package baz is generated code used to make or handle TChannel calls using Thrift.
package baz

import (
	"errors"
	"fmt"

	"github.com/uber/tchannel-go/thrift"
	"github.com/uber/zanzibar/runtime/tchannel"
	"go.uber.org/thriftrw/wire"
)

// Interfaces for the service and client for the services defined in the IDL.

// TChanSecondService is the interface that defines the server handler and client interface.
type TChanSecondService interface {
	Echo(ctx thrift.Context, arg string) (string, error)
}

// TChanSimpleService is the interface that defines the server handler and client interface.
type TChanSimpleService interface {
	Call(ctx thrift.Context, arg *Data) (*Data, error)
	Simple(ctx thrift.Context) error
	SimpleFuture(ctx thrift.Context) error
}

// Implementation of a client and service handler.

type tchanSecondServiceClient struct {
	thriftService string
	client        tchannel.TChanClient
}

func NewTChanSecondServiceInheritedClient(thriftService string, client tchannel.TChanClient) *tchanSecondServiceClient {
	return &tchanSecondServiceClient{
		thriftService,
		client,
	}
}

// NewTChanSecondServiceClient creates a client that can be used to make remote calls.
func NewTChanSecondServiceClient(client tchannel.TChanClient) TChanSecondService {
	return NewTChanSecondServiceInheritedClient("SecondService", client)
}

func (c *tchanSecondServiceClient) Echo(ctx thrift.Context, arg string) (string, error) {
	var resp SecondService_Echo_Result
	args := SecondService_Echo_Args{
		Arg: &arg,
	}
	success, err := c.client.Call(ctx, c.thriftService, "Echo", &args, &resp)
	if err == nil && !success {
		err = errors.New("received no result or unknown exception for Echo")
	}
	if err != nil {
		return "", err
	}

	return SecondService_Echo_Helper.UnwrapResponse(&resp)
}

type tchanSecondServiceServer struct {
	handler TChanSecondService
}

// NewTChanSecondServiceServer wraps a handler for TChanSecondService so it can be
// registered with a tchannel.Server.
func NewTChanSecondServiceServer(handler TChanSecondService) tchannel.TChanServer {
	return &tchanSecondServiceServer{
		handler,
	}
}

func (s *tchanSecondServiceServer) Service() string {
	return "SecondService"
}

func (s *tchanSecondServiceServer) Methods() []string {
	return []string{
		"Echo",
	}
}

func (s *tchanSecondServiceServer) Handle(ctx thrift.Context, methodName string, wireValue wire.Value) (bool, tchannel.RWTStruct, error) {
	switch methodName {
	case "Echo":
		return s.handleEcho(ctx, wireValue)

	default:
		return false, nil, fmt.Errorf("method %v not found in service %v", methodName, s.Service())
	}
}

func (s *tchanSecondServiceServer) handleEcho(ctx thrift.Context, wireValue wire.Value) (bool, tchannel.RWTStruct, error) {
	var req SecondService_Echo_Args
	var res SecondService_Echo_Result

	if err := req.FromWire(wireValue); err != nil {
		return false, nil, err
	}

	r, err := s.handler.Echo(ctx, *req.Arg)

	if err != nil {
		return false, nil, err
	} else {
		res.Success = &r
	}

	return err == nil, &res, nil
}

type tchanSimpleServiceClient struct {
	thriftService string
	client        tchannel.TChanClient
}

func NewTChanSimpleServiceInheritedClient(thriftService string, client tchannel.TChanClient) *tchanSimpleServiceClient {
	return &tchanSimpleServiceClient{
		thriftService,
		client,
	}
}

// NewTChanSimpleServiceClient creates a client that can be used to make remote calls.
func NewTChanSimpleServiceClient(client tchannel.TChanClient) TChanSimpleService {
	return NewTChanSimpleServiceInheritedClient("SimpleService", client)
}

func (c *tchanSimpleServiceClient) Call(ctx thrift.Context, arg *Data) (*Data, error) {
	var resp SimpleService_Call_Result
	args := SimpleService_Call_Args{
		Arg: arg,
	}
	success, err := c.client.Call(ctx, c.thriftService, "Call", &args, &resp)
	if err == nil && !success {
		err = errors.New("received no result or unknown exception for Call")
	}
	if err != nil {
		return nil, err
	}

	return SimpleService_Call_Helper.UnwrapResponse(&resp)
}

func (c *tchanSimpleServiceClient) Simple(ctx thrift.Context) error {
	var resp SimpleService_Simple_Result
	args := SimpleService_Simple_Args{}
	success, err := c.client.Call(ctx, c.thriftService, "Simple", &args, &resp)
	if err == nil && !success {
		switch {
		case resp.SimpleErr != nil:
			err = resp.SimpleErr
		default:
			err = errors.New("received no result or unknown exception for Simple")
		}
	}

	return err
}

func (c *tchanSimpleServiceClient) SimpleFuture(ctx thrift.Context) error {
	var resp SimpleService_SimpleFuture_Result
	args := SimpleService_SimpleFuture_Args{}
	success, err := c.client.Call(ctx, c.thriftService, "SimpleFuture", &args, &resp)
	if err == nil && !success {
		switch {
		case resp.SimpleErr != nil:
			err = resp.SimpleErr
		case resp.NewErr != nil:
			err = resp.NewErr
		default:
			err = errors.New("received no result or unknown exception for SimpleFuture")
		}
	}

	return err
}

type tchanSimpleServiceServer struct {
	handler TChanSimpleService
}

// NewTChanSimpleServiceServer wraps a handler for TChanSimpleService so it can be
// registered with a tchannel.Server.
func NewTChanSimpleServiceServer(handler TChanSimpleService) tchannel.TChanServer {
	return &tchanSimpleServiceServer{
		handler,
	}
}

func (s *tchanSimpleServiceServer) Service() string {
	return "SimpleService"
}

func (s *tchanSimpleServiceServer) Methods() []string {
	return []string{
		"Call",
		"Simple",
		"SimpleFuture",
	}
}

func (s *tchanSimpleServiceServer) Handle(ctx thrift.Context, methodName string, wireValue wire.Value) (bool, tchannel.RWTStruct, error) {
	switch methodName {
	case "Call":
		return s.handleCall(ctx, wireValue)
	case "Simple":
		return s.handleSimple(ctx, wireValue)
	case "SimpleFuture":
		return s.handleSimpleFuture(ctx, wireValue)

	default:
		return false, nil, fmt.Errorf("method %v not found in service %v", methodName, s.Service())
	}
}

func (s *tchanSimpleServiceServer) handleCall(ctx thrift.Context, wireValue wire.Value) (bool, tchannel.RWTStruct, error) {
	var req SimpleService_Call_Args
	var res SimpleService_Call_Result

	if err := req.FromWire(wireValue); err != nil {
		return false, nil, err
	}

	r, err := s.handler.Call(ctx, req.Arg)

	if err != nil {
		return false, nil, err
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}

func (s *tchanSimpleServiceServer) handleSimple(ctx thrift.Context, wireValue wire.Value) (bool, tchannel.RWTStruct, error) {
	var req SimpleService_Simple_Args
	var res SimpleService_Simple_Result

	if err := req.FromWire(wireValue); err != nil {
		return false, nil, err
	}

	err := s.handler.Simple(ctx)

	if err != nil {
		switch v := err.(type) {
		case *SimpleErr:
			if v == nil {
				return false, nil, errors.New("Handler for simpleErr returned non-nil error type *SimpleErr but nil value")
			}
			res.SimpleErr = v
		default:
			return false, nil, err
		}
	} else {
	}

	return err == nil, &res, nil
}

func (s *tchanSimpleServiceServer) handleSimpleFuture(ctx thrift.Context, wireValue wire.Value) (bool, tchannel.RWTStruct, error) {
	var req SimpleService_SimpleFuture_Args
	var res SimpleService_SimpleFuture_Result

	if err := req.FromWire(wireValue); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.SimpleFuture(ctx)

	if err != nil {
		switch v := err.(type) {
		case *SimpleErr:
			if v == nil {
				return false, nil, errors.New("Handler for simpleErr returned non-nil error type *SimpleErr but nil value")
			}
			res.SimpleErr = v
		case *NewErr:
			if v == nil {
				return false, nil, errors.New("Handler for newErr returned non-nil error type *NewErr_ but nil value")
			}
			res.NewErr = v
		default:
			return false, nil, err
		}
	} else {
	}

	return err == nil, &res, nil
}
