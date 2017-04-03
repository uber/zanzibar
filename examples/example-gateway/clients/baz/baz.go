// TODO: (lu) generate

package bazClient

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/uber/zanzibar/runtime"

	zt "github.com/uber/zanzibar/runtime/tchannel"
	"go.uber.org/thriftrw/wire"

	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/github.com/uber/zanzibar/clients/baz/baz"
)

// TChanBaz is the interface that defines the server handler and client interface.
type TChanBaz interface {
	Call(ctx context.Context, reqHeaders map[string]string, r *baz.SimpleService_Call_Args) (map[string]string, *baz.BazResponse, error)
	Simple(ctx context.Context, reqHeaders map[string]string) (map[string]string, error)
	SimpleFuture(ctx context.Context, reqHeaders map[string]string) (map[string]string, error)
}

// NewClient returns a new http client for service Bar.
func NewClient(config *zanzibar.StaticConfig, gateway *zanzibar.Gateway) *BazClient {
	// TODO: (lu) get serviceName from thrift
	serviceName := "SimpleService"

	sc := gateway.Channel.GetSubChannel(serviceName)

	ip := config.MustGetString("clients.baz.ip")
	port := config.MustGetInt("clients.baz.port")
	sc.Peers().Add(ip + ":" + strconv.Itoa(int(port)))

	client := zt.NewClient(gateway.Channel, serviceName)

	return &BazClient{
		thriftService: serviceName,
		client:        client,
	}
}

// BazClient is the client to talk to SimpleService backend.
type BazClient struct {
	thriftService string
	client        zt.TChanClient
}

// Call ...
func (c *BazClient) Call(ctx context.Context, reqHeaders map[string]string, args *baz.SimpleService_Call_Args) (map[string]string, *baz.BazResponse, error) {
	var result baz.SimpleService_Call_Result
	respHeaders, success, err := c.client.Call(ctx, c.thriftService, "Call", reqHeaders, args, &result)
	if err == nil && !success {
		err = errors.New("received no result or unknown exception for Call")
	}
	if err != nil {
		return nil, nil, err
	}

	resp, err := baz.SimpleService_Call_Helper.UnwrapResponse(&result)

	return respHeaders, resp, err
}

// Simple ...
func (c *BazClient) Simple(ctx context.Context, reqHeaders map[string]string) (map[string]string, error) {
	var result baz.SimpleService_Simple_Result
	args := baz.SimpleService_Simple_Args{}
	respHeaders, success, err := c.client.Call(ctx, c.thriftService, "Simple", reqHeaders, &args, &result)
	if err == nil && !success {
		switch {
		case result.SimpleErr != nil:
			err = result.SimpleErr
		default:
			err = errors.New("received no result or unknown exception for Simple")
		}
	}

	return respHeaders, err
}

// SimpleFuture ...
func (c *BazClient) SimpleFuture(ctx context.Context, reqHeaders map[string]string) (map[string]string, error) {
	var result baz.SimpleService_SimpleFuture_Result
	args := baz.SimpleService_SimpleFuture_Args{}
	respHeaders, success, err := c.client.Call(ctx, c.thriftService, "SimpleFuture", reqHeaders, &args, &result)
	if err == nil && !success {
		switch {
		case result.SimpleErr != nil:
			err = result.SimpleErr
		case result.NewErr != nil:
			err = result.NewErr
		default:
			err = errors.New("received no result or unknown exception for SimpleFuture")
		}
	}

	return respHeaders, err
}

// SimpleServiceServer is the Baz backend service
type SimpleServiceServer struct {
	handler TChanBaz
}

// NewSimpleServiceServer wraps a handler for Baz so it can be registered with a thrift server.
func NewSimpleServiceServer(handler TChanBaz) zt.TChanServer {
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
func (s *SimpleServiceServer) Handle(ctx context.Context, methodName string, reqHeaders map[string]string, wireValue *wire.Value) (bool, map[string]string, zt.RWTStruct, error) {
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

func (s *SimpleServiceServer) handleCall(ctx context.Context, reqHeaders map[string]string, wireValue *wire.Value) (bool, map[string]string, zt.RWTStruct, error) {
	var req baz.SimpleService_Call_Args
	var res baz.SimpleService_Call_Result

	if err := req.FromWire(*wireValue); err != nil {
		return false, nil, nil, err
	}

	respHeaders, r, err := s.handler.Call(ctx, reqHeaders, &req)

	if err != nil {
		return false, nil, nil, err
	}

	res.Success = r
	return true, respHeaders, &res, nil
}

func (s *SimpleServiceServer) handleSimple(ctx context.Context, reqHeaders map[string]string, wireValue *wire.Value) (bool, map[string]string, zt.RWTStruct, error) {
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

func (s *SimpleServiceServer) handleSimpleFuture(ctx context.Context, reqHeaders map[string]string, wireValue *wire.Value) (bool, map[string]string, zt.RWTStruct, error) {
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
