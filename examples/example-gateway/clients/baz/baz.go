// TODO: (lu) generate

package bazClient

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/uber/zanzibar/runtime"

	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
)

// TChanBaz is the interface that defines the server handler and client interface.
type TChanBaz interface {
	Call(ctx context.Context, reqHeaders map[string]string, r *baz.BazRequest) (*baz.BazResponse, map[string]string, error)
	Simple(ctx context.Context, reqHeaders map[string]string) (map[string]string, error)
	SimpleFuture(ctx context.Context, reqHeaders map[string]string) (map[string]string, error)
}

// NewClient returns a new http client for service Bar.
func NewClient(gateway *zanzibar.Gateway) *BazClient {
	// this is the service discovery service name
	serviceName := gateway.Config.MustGetString("clients.baz.serviceName")
	sc := gateway.Channel.GetSubChannel(serviceName)

	ip := gateway.Config.MustGetString("clients.baz.ip")
	port := gateway.Config.MustGetInt("clients.baz.port")
	sc.Peers().Add(ip + ":" + strconv.Itoa(int(port)))

	// TODO: (lu) maybe set these at per method level
	timeout := time.Duration(
		gateway.Config.MustGetInt("clients.baz.timeout"),
	) * time.Millisecond
	timeoutPerAttempt := time.Duration(
		gateway.Config.MustGetInt("clients.baz.timeoutPerAttempt"),
	) * time.Millisecond

	client := zanzibar.NewTChannelClient(gateway.Channel,
		&zanzibar.TChannelClientOption{
			ServiceName:       serviceName,
			Timeout:           timeout,
			TimeoutPerAttempt: timeoutPerAttempt,
		},
	)

	return &BazClient{
		// this is the thrift service name, different from service discovery service name
		thriftService: "SimpleService",
		client:        client,
	}
}

// BazClient is the client to talk to SimpleService backend.
type BazClient struct {
	// TODO: (lu) refactor to get rid of this field
	thriftService string
	client        zanzibar.TChanClient
}

// Call ...
func (c *BazClient) Call(ctx context.Context, reqHeaders map[string]string, args *baz.SimpleService_Call_Args) (*baz.BazResponse, map[string]string, error) {
	var result baz.SimpleService_Call_Result

	success, respHeaders, err := c.client.Call(ctx, c.thriftService, "Call", reqHeaders, args, &result)
	if err == nil && !success {
		err = errors.New("received no result or unknown exception for Call")
	}
	if err != nil {
		return nil, nil, err
	}

	resp, err := baz.SimpleService_Call_Helper.UnwrapResponse(&result)

	return resp, respHeaders, err
}

// Simple ...
func (c *BazClient) Simple(ctx context.Context, reqHeaders map[string]string) (map[string]string, error) {
	var result baz.SimpleService_Simple_Result

	args := baz.SimpleService_Simple_Args{}
	success, respHeaders, err := c.client.Call(ctx, c.thriftService, "Simple", reqHeaders, &args, &result)
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
	success, respHeaders, err := c.client.Call(ctx, c.thriftService, "SimpleFuture", reqHeaders, &args, &result)
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
