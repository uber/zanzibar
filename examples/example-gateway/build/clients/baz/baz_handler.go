// Package bazClient is generated code used to make or handle TChannel calls using Thrift.
package bazClient

import (
	"context"
	"errors"

	"github.com/uber/zanzibar/runtime"

	clientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
)

// SimpleServiceHandler is the handler struct for TChannel service SimpleService.
type SimpleServiceHandler struct {
	CallFunc         SimpleServiceCallFunc
	SimpleFunc       SimpleServiceSimpleFunc
	SimpleFutureFunc SimpleServiceSimpleFutureFunc
}

// SimpleServiceCallFunc ...
type SimpleServiceCallFunc func(
	context.Context, map[string]string, *clientsBazBaz.SimpleService_Call_Args,
) (*clientsBazBaz.BazResponse, map[string]string, error)

// SimpleServiceSimpleFunc ...
type SimpleServiceSimpleFunc func(
	context.Context, map[string]string,
) (map[string]string, error)

// SimpleServiceSimpleFutureFunc ...
type SimpleServiceSimpleFutureFunc func(
	context.Context, map[string]string,
) (map[string]string, error)

// NewServerWithSimpleServiceCall creates a TChanServer with given handler function registered.
func NewServerWithSimpleServiceCall(arg SimpleServiceCallFunc) zanzibar.TChanServer {
	return NewSimpleServiceServer(&SimpleServiceHandler{
		CallFunc: arg,
	})
}

// NewServerWithSimpleServiceSimple creates a TChanServer with given handler function registered.
func NewServerWithSimpleServiceSimple(arg SimpleServiceSimpleFunc) zanzibar.TChanServer {
	return NewSimpleServiceServer(&SimpleServiceHandler{
		SimpleFunc: arg,
	})
}

// NewServerWithSimpleServiceSimpleFuture creates a TChanServer with given handler function registered.
func NewServerWithSimpleServiceSimpleFuture(arg SimpleServiceSimpleFutureFunc) zanzibar.TChanServer {
	return NewSimpleServiceServer(&SimpleServiceHandler{
		SimpleFutureFunc: arg,
	})
}

// Call ...
func (h *SimpleServiceHandler) Call(
	ctx context.Context, reqHeaders map[string]string, args *clientsBazBaz.SimpleService_Call_Args,
) (*clientsBazBaz.BazResponse, map[string]string, error) {
	if h.CallFunc == nil {
		return nil, nil, errors.New("CallFunc is not defined")
	}
	return h.CallFunc(ctx, reqHeaders, args)
}

// Simple ...
func (h *SimpleServiceHandler) Simple(
	ctx context.Context, reqHeaders map[string]string,
) (map[string]string, error) {
	if h.SimpleFunc == nil {
		return nil, errors.New("SimpleFunc is not defined")
	}
	return h.SimpleFunc(ctx, reqHeaders)
}

// SimpleFuture ...
func (h *SimpleServiceHandler) SimpleFuture(
	ctx context.Context, reqHeaders map[string]string,
) (map[string]string, error) {
	if h.SimpleFutureFunc == nil {
		return nil, errors.New("SimpleFutureFunc is not defined")
	}
	return h.SimpleFutureFunc(ctx, reqHeaders)
}
