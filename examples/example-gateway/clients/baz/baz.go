// TODO: (lu) generate

package bazClient

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/uber/tchannel-go"
	"github.com/uber/zanzibar/runtime"

	zt "github.com/uber/zanzibar/runtime/tchannel"

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
	serviceName := "SimpleService"

	sc := gateway.Channel.GetSubChannel(serviceName)

	ip := config.MustGetString("clients.baz.ip")
	port := config.MustGetInt("clients.baz.port")
	sc.Peers().Add(ip + ":" + strconv.Itoa(int(port)))

	client := zt.NewClient(gateway.Channel, serviceName)

	// TODO: (lu) maybe set these at per method level
	timeout := time.Duration(config.MustGetInt("clients.baz.timeout")) * time.Millisecond
	timeoutPerAttempt := time.Duration(config.MustGetInt("clients.baz.timeoutPerAttempt")) * time.Millisecond

	return &BazClient{
		thriftService:     serviceName,
		client:            client,
		timeout:           timeout,
		timeoutPerAttempt: timeoutPerAttempt,
	}
}

// BazClient is the client to talk to SimpleService backend.
type BazClient struct {
	thriftService string
	client        zt.TChanClient

	timeout           time.Duration
	timeoutPerAttempt time.Duration
}

// Call ...
func (c *BazClient) Call(ctx context.Context, reqHeaders map[string]string, args *baz.SimpleService_Call_Args) (map[string]string, *baz.BazResponse, error) {
	var result baz.SimpleService_Call_Result

	retryOpts := &tchannel.RetryOptions{
		TimeoutPerAttempt: c.timeoutPerAttempt,
	}
	ctx, cancel := tchannel.NewContextBuilder(c.timeout).
		SetParentContext(ctx).
		SetRetryOptions(retryOpts).
		Build()
	defer cancel()

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

	retryOpts := &tchannel.RetryOptions{
		TimeoutPerAttempt: c.timeoutPerAttempt,
	}
	ctx, cancel := tchannel.NewContextBuilder(c.timeout).
		SetParentContext(ctx).
		SetRetryOptions(retryOpts).
		Build()
	defer cancel()

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

	retryOpts := &tchannel.RetryOptions{
		TimeoutPerAttempt: c.timeoutPerAttempt,
	}
	ctx, cancel := tchannel.NewContextBuilder(c.timeout).
		SetParentContext(ctx).
		SetRetryOptions(retryOpts).
		Build()
	defer cancel()

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
