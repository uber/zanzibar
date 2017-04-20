// Package bazClient is generated code used to make or handle TChannel calls using Thrift.
package bazClient

import (
	"context"
	"errors"
	"fmt"

	"github.com/uber/zanzibar/runtime"
	"go.uber.org/thriftrw/wire"

	clientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
)

// SimpleServiceServer is the TChannel backend for SimpleService service.
type SimpleServiceServer struct {
	handler TChanBaz
}

// NewSimpleServiceServer wraps a handler for Baz so it can be registered with a thrift server.
func NewSimpleServiceServer(handler TChanBaz) zanzibar.TChanServer {
	return &SimpleServiceServer{
		handler,
	}
}

// Service returns the service name.
func (s *SimpleServiceServer) Service() string {
	return "SimpleService"
}

// Methods returns the method names handled by this server.
func (s *SimpleServiceServer) Methods() []string {
	return []string{
		"Call",
		"Simple",
		"SimpleFuture",
	}
}

// Handle dispatches a method call to corresponding handler.
func (s *SimpleServiceServer) Handle(
	ctx context.Context,
	methodName string,
	reqHeaders map[string]string,
	wireValue *wire.Value,
) (bool, zanzibar.RWTStruct, map[string]string, error) {
	switch methodName {
	case "Call":
		return s.handleCall(ctx, reqHeaders, wireValue)
	case "Simple":
		return s.handleSimple(ctx, reqHeaders, wireValue)
	case "SimpleFuture":
		return s.handleSimpleFuture(ctx, reqHeaders, wireValue)
	default:
		return false, nil, nil, fmt.Errorf(
			"method %v not found in service %v", methodName, s.Service(),
		)
	}
}

func (s *SimpleServiceServer) handleCall(
	ctx context.Context,
	reqHeaders map[string]string,
	wireValue *wire.Value,
) (bool, zanzibar.RWTStruct, map[string]string, error) {
	var req clientsBazBaz.SimpleService_Call_Args
	var res clientsBazBaz.SimpleService_Call_Result

	if err := req.FromWire(*wireValue); err != nil {
		return false, nil, nil, err
	}
	r, respHeaders, err := s.handler.Call(ctx, reqHeaders, &req)

	if err != nil {
		return false, nil, nil, err
	}
	res.Success = r

	return err == nil, &res, respHeaders, nil
}

func (s *SimpleServiceServer) handleSimple(
	ctx context.Context,
	reqHeaders map[string]string,
	wireValue *wire.Value,
) (bool, zanzibar.RWTStruct, map[string]string, error) {
	var req clientsBazBaz.SimpleService_Simple_Args
	var res clientsBazBaz.SimpleService_Simple_Result

	if err := req.FromWire(*wireValue); err != nil {
		return false, nil, nil, err
	}
	respHeaders, err := s.handler.Simple(ctx, reqHeaders)

	if err != nil {
		switch v := err.(type) {
		case *clientsBazBaz.SimpleErr:
			if v == nil {
				return false, nil, nil, errors.New(
					"Handler for SimpleErr returned non-nil error type *SimpleErr but nil value",
				)
			}
			res.SimpleErr = v
		default:
			return false, nil, nil, err
		}
	}

	return err == nil, &res, respHeaders, nil
}

func (s *SimpleServiceServer) handleSimpleFuture(
	ctx context.Context,
	reqHeaders map[string]string,
	wireValue *wire.Value,
) (bool, zanzibar.RWTStruct, map[string]string, error) {
	var req clientsBazBaz.SimpleService_SimpleFuture_Args
	var res clientsBazBaz.SimpleService_SimpleFuture_Result

	if err := req.FromWire(*wireValue); err != nil {
		return false, nil, nil, err
	}
	respHeaders, err := s.handler.SimpleFuture(ctx, reqHeaders)

	if err != nil {
		switch v := err.(type) {
		case *clientsBazBaz.SimpleErr:
			if v == nil {
				return false, nil, nil, errors.New(
					"Handler for SimpleErr returned non-nil error type *SimpleErr but nil value",
				)
			}
			res.SimpleErr = v
		case *clientsBazBaz.NewErr:
			if v == nil {
				return false, nil, nil, errors.New(
					"Handler for NewErr returned non-nil error type *NewErr but nil value",
				)
			}
			res.NewErr = v
		default:
			return false, nil, nil, err
		}
	}

	return err == nil, &res, respHeaders, nil
}
