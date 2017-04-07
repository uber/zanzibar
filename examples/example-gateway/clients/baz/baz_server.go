// TODO: (lu) generate

// TODO: (lu) find a better place for this file

package bazClient

import (
	"context"
	"errors"
	"fmt"

	"github.com/uber/zanzibar/runtime"
	"go.uber.org/thriftrw/wire"

	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/clients/baz/baz"
)

// SimpleServiceServer is the Baz backend service
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
func (s *SimpleServiceServer) Handle(ctx context.Context, methodName string, reqHeaders map[string]string, wireValue *wire.Value) (bool, map[string]string, zanzibar.RWTStruct, error) {
	switch methodName {
	case "Call":
		return s.handleCall(ctx, reqHeaders, wireValue)
	case "Simple":
		return s.handleSimple(ctx, reqHeaders, wireValue)
	case "SimpleFuture":
		return s.handleSimpleFuture(ctx, reqHeaders, wireValue)

	default:
		return false, nil, nil, fmt.Errorf("method %v not found in service %v", methodName, s.Service())
	}
}

func (s *SimpleServiceServer) handleCall(ctx context.Context, reqHeaders map[string]string, wireValue *wire.Value) (bool, map[string]string, zanzibar.RWTStruct, error) {
	var req baz.SimpleService_Call_Args
	var res baz.SimpleService_Call_Result

	if err := req.FromWire(*wireValue); err != nil {
		return false, nil, nil, err
	}

	respHeaders, r, err := s.handler.Call(ctx, reqHeaders, req.Arg)

	if err != nil {
		return false, nil, nil, err
	}

	res.Success = r
	return true, respHeaders, &res, nil
}

func (s *SimpleServiceServer) handleSimple(ctx context.Context, reqHeaders map[string]string, wireValue *wire.Value) (bool, map[string]string, zanzibar.RWTStruct, error) {
	var req baz.SimpleService_Simple_Args
	var res baz.SimpleService_Simple_Result

	if err := req.FromWire(*wireValue); err != nil {
		return false, nil, nil, err
	}

	respHeaders, err := s.handler.Simple(ctx, reqHeaders)

	if err != nil {
		switch v := err.(type) {
		case *baz.SimpleErr:
			if v == nil {
				return false, nil, nil, errors.New("Handler for simpleErr returned non-nil error type *SimpleErr but nil value")
			}
			res.SimpleErr = v
		default:
			return false, nil, nil, err
		}
	}

	return err == nil, respHeaders, &res, nil
}

func (s *SimpleServiceServer) handleSimpleFuture(ctx context.Context, reqHeaders map[string]string, wireValue *wire.Value) (bool, map[string]string, zanzibar.RWTStruct, error) {
	var req baz.SimpleService_SimpleFuture_Args
	var res baz.SimpleService_SimpleFuture_Result

	if err := req.FromWire(*wireValue); err != nil {
		return false, nil, nil, err
	}

	respHeaders, err := s.handler.SimpleFuture(ctx, reqHeaders)

	if err != nil {
		switch v := err.(type) {
		case *baz.SimpleErr:
			if v == nil {
				return false, nil, nil, errors.New("Handler for simpleErr returned non-nil error type *SimpleErr but nil value")
			}
			res.SimpleErr = v
		case *baz.NewErr:
			if v == nil {
				return false, nil, nil, errors.New("Handler for newErr returned non-nil error type *NewErr_ but nil value")
			}
			res.NewErr = v
		default:
			return false, nil, nil, err
		}
	}

	return err == nil, respHeaders, &res, nil
}

// CallFunc ...
type CallFunc func(context.Context, map[string]string, *baz.BazRequest) (map[string]string, *baz.BazResponse, error)

// SimpleFunc ...
type SimpleFunc func(context.Context, map[string]string) (map[string]string, error)

// SimpleFutureFunc ...
type SimpleFutureFunc func(context.Context, map[string]string) (map[string]string, error)

// WithCall creates a TChanServer with Call handler function registered
func WithCall(call CallFunc) zanzibar.TChanServer {
	return NewSimpleServiceServer(&Handler{
		CallFunc: call,
	})
}

// WithSimple creates a TChanServer with Simple handler function registered
func WithSimple(simple SimpleFunc) zanzibar.TChanServer {
	return NewSimpleServiceServer(&Handler{
		SimpleFunc: simple,
	})
}

// WithSimpleFuture creates a TChanServer with SimpleFuture handler function registered
func WithSimpleFuture(simpleFuture SimpleFutureFunc) zanzibar.TChanServer {
	return NewSimpleServiceServer(&Handler{
		SimpleFutureFunc: simpleFuture,
	})
}

// Handler is intended as the base struct for testing
type Handler struct {
	CallFunc         CallFunc
	SimpleFunc       SimpleFunc
	SimpleFutureFunc SimpleFutureFunc
}

// Call ...
func (h *Handler) Call(ctx context.Context, reqHeaders map[string]string, r *baz.BazRequest) (map[string]string, *baz.BazResponse, error) {
	if h.CallFunc == nil {
		return nil, nil, errors.New("not implemented")
	}
	return h.CallFunc(ctx, reqHeaders, r)
}

// Simple ...
func (h *Handler) Simple(ctx context.Context, reqHeaders map[string]string) (map[string]string, error) {
	if h.SimpleFunc == nil {
		return nil, errors.New("not implemented")
	}
	return h.SimpleFunc(ctx, reqHeaders)
}

// SimpleFuture ...
func (h *Handler) SimpleFuture(ctx context.Context, reqHeaders map[string]string) (map[string]string, error) {
	if h.SimpleFutureFunc == nil {
		return nil, errors.New("not implemented")
	}
	return h.SimpleFutureFunc(ctx, reqHeaders)
}
