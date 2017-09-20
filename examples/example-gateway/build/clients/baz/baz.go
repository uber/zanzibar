// Code generated by zanzibar
// @generated

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

package bazClient

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/uber/tchannel-go"
	"github.com/uber/zanzibar/runtime"

	clientsBazBase "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/base"
	clientsBazBaz "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/baz/baz"
)

// Client defines baz client interface.
type Client interface {
	EchoBinary(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoBinary_Args,
	) ([]byte, map[string]string, error)
	EchoBool(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoBool_Args,
	) (bool, map[string]string, error)
	EchoDouble(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoDouble_Args,
	) (float64, map[string]string, error)
	EchoEnum(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoEnum_Args,
	) (clientsBazBaz.Fruit, map[string]string, error)
	EchoI16(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoI16_Args,
	) (int16, map[string]string, error)
	EchoI32(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoI32_Args,
	) (int32, map[string]string, error)
	EchoI64(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoI64_Args,
	) (int64, map[string]string, error)
	EchoI8(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoI8_Args,
	) (int8, map[string]string, error)
	EchoString(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoString_Args,
	) (string, map[string]string, error)
	EchoStringList(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoStringList_Args,
	) ([]string, map[string]string, error)
	EchoStringMap(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoStringMap_Args,
	) (map[string]*clientsBazBase.BazResponse, map[string]string, error)
	EchoStringSet(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoStringSet_Args,
	) (map[string]struct{}, map[string]string, error)
	EchoStructList(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoStructList_Args,
	) ([]*clientsBazBase.BazResponse, map[string]string, error)
	EchoStructMap(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoStructMap_Args,
	) ([]struct {
		Key   *clientsBazBase.BazResponse
		Value string
	}, map[string]string, error)
	EchoStructSet(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoStructSet_Args,
	) ([]*clientsBazBase.BazResponse, map[string]string, error)
	EchoTypedef(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SecondService_EchoTypedef_Args,
	) (clientsBazBase.UUID, map[string]string, error)
	Call(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SimpleService_Call_Args,
	) (map[string]string, error)
	Compare(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsBazBaz.SimpleService_Compare_Args,
	) (*clientsBazBase.BazResponse, map[string]string, error)
	Ping(
		ctx context.Context,
		reqHeaders map[string]string,
	) (*clientsBazBase.BazResponse, map[string]string, error)
	DeliberateDiffNoop(
		ctx context.Context,
		reqHeaders map[string]string,
	) (map[string]string, error)
}

// NewClient returns a new TChannel client for service baz.
func NewClient(gateway *zanzibar.Gateway) Client {
	serviceName := gateway.Config.MustGetString("clients.baz.serviceName")
	var routingKey string
	if gateway.Config.ContainsKey("clients.baz.routingKey") {
		routingKey = gateway.Config.MustGetString("clients.baz.routingKey")
	}
	sc := gateway.Channel.GetSubChannel(serviceName, tchannel.Isolated)

	ip := gateway.Config.MustGetString("clients.baz.ip")
	port := gateway.Config.MustGetInt("clients.baz.port")
	sc.Peers().Add(ip + ":" + strconv.Itoa(int(port)))

	timeout := time.Millisecond * time.Duration(
		gateway.Config.MustGetInt("clients.baz.timeout"),
	)
	timeoutPerAttempt := time.Millisecond * time.Duration(
		gateway.Config.MustGetInt("clients.baz.timeoutPerAttempt"),
	)

	methodNames := map[string]string{
		"SecondService::echoBinary":     "EchoBinary",
		"SecondService::echoBool":       "EchoBool",
		"SecondService::echoDouble":     "EchoDouble",
		"SecondService::echoEnum":       "EchoEnum",
		"SecondService::echoI16":        "EchoI16",
		"SecondService::echoI32":        "EchoI32",
		"SecondService::echoI64":        "EchoI64",
		"SecondService::echoI8":         "EchoI8",
		"SecondService::echoString":     "EchoString",
		"SecondService::echoStringList": "EchoStringList",
		"SecondService::echoStringMap":  "EchoStringMap",
		"SecondService::echoStringSet":  "EchoStringSet",
		"SecondService::echoStructList": "EchoStructList",
		"SecondService::echoStructMap":  "EchoStructMap",
		"SecondService::echoStructSet":  "EchoStructSet",
		"SecondService::echoTypedef":    "EchoTypedef",
		"SimpleService::call":           "Call",
		"SimpleService::compare":        "Compare",
		"SimpleService::ping":           "Ping",
		"SimpleService::sillyNoop":      "DeliberateDiffNoop",
	}

	client := zanzibar.NewTChannelClient(
		gateway.Channel,
		gateway.Logger,
		gateway.AllHostScope,
		&zanzibar.TChannelClientOption{
			ServiceName:       serviceName,
			ClientID:          "baz",
			MethodNames:       methodNames,
			Timeout:           timeout,
			TimeoutPerAttempt: timeoutPerAttempt,
			RoutingKey:        &routingKey,
		},
	)

	return &bazClient{
		client: client,
		logger: gateway.Logger,
	}
}

// bazClient is the TChannel client for downstream service.
type bazClient struct {
	client zanzibar.TChannelClient
	logger *zap.Logger
}

// EchoBinary is a client RPC call for method "SecondService::echoBinary"
func (c *bazClient) EchoBinary(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoBinary_Args,
) ([]byte, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoBinary_Result
	var resp []byte

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoBinary", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoBinary")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoBinary_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoBool is a client RPC call for method "SecondService::echoBool"
func (c *bazClient) EchoBool(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoBool_Args,
) (bool, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoBool_Result
	var resp bool

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoBool", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoBool")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoBool_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoDouble is a client RPC call for method "SecondService::echoDouble"
func (c *bazClient) EchoDouble(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoDouble_Args,
) (float64, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoDouble_Result
	var resp float64

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoDouble", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoDouble")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoDouble_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoEnum is a client RPC call for method "SecondService::echoEnum"
func (c *bazClient) EchoEnum(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoEnum_Args,
) (clientsBazBaz.Fruit, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoEnum_Result
	var resp clientsBazBaz.Fruit

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoEnum", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoEnum")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoEnum_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoI16 is a client RPC call for method "SecondService::echoI16"
func (c *bazClient) EchoI16(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoI16_Args,
) (int16, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoI16_Result
	var resp int16

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoI16", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoI16")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoI16_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoI32 is a client RPC call for method "SecondService::echoI32"
func (c *bazClient) EchoI32(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoI32_Args,
) (int32, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoI32_Result
	var resp int32

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoI32", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoI32")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoI32_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoI64 is a client RPC call for method "SecondService::echoI64"
func (c *bazClient) EchoI64(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoI64_Args,
) (int64, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoI64_Result
	var resp int64

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoI64", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoI64")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoI64_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoI8 is a client RPC call for method "SecondService::echoI8"
func (c *bazClient) EchoI8(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoI8_Args,
) (int8, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoI8_Result
	var resp int8

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoI8", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoI8")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoI8_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoString is a client RPC call for method "SecondService::echoString"
func (c *bazClient) EchoString(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoString_Args,
) (string, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoString_Result
	var resp string

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoString", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoString")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoString_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoStringList is a client RPC call for method "SecondService::echoStringList"
func (c *bazClient) EchoStringList(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoStringList_Args,
) ([]string, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoStringList_Result
	var resp []string

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoStringList", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoStringList")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoStringList_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoStringMap is a client RPC call for method "SecondService::echoStringMap"
func (c *bazClient) EchoStringMap(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoStringMap_Args,
) (map[string]*clientsBazBase.BazResponse, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoStringMap_Result
	var resp map[string]*clientsBazBase.BazResponse

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoStringMap", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoStringMap")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoStringMap_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoStringSet is a client RPC call for method "SecondService::echoStringSet"
func (c *bazClient) EchoStringSet(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoStringSet_Args,
) (map[string]struct{}, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoStringSet_Result
	var resp map[string]struct{}

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoStringSet", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoStringSet")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoStringSet_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoStructList is a client RPC call for method "SecondService::echoStructList"
func (c *bazClient) EchoStructList(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoStructList_Args,
) ([]*clientsBazBase.BazResponse, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoStructList_Result
	var resp []*clientsBazBase.BazResponse

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoStructList", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoStructList")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoStructList_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoStructMap is a client RPC call for method "SecondService::echoStructMap"
func (c *bazClient) EchoStructMap(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoStructMap_Args,
) ([]struct {
	Key   *clientsBazBase.BazResponse
	Value string
}, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoStructMap_Result
	var resp []struct {
		Key   *clientsBazBase.BazResponse
		Value string
	}

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoStructMap", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoStructMap")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoStructMap_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoStructSet is a client RPC call for method "SecondService::echoStructSet"
func (c *bazClient) EchoStructSet(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoStructSet_Args,
) ([]*clientsBazBase.BazResponse, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoStructSet_Result
	var resp []*clientsBazBase.BazResponse

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoStructSet", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoStructSet")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoStructSet_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// EchoTypedef is a client RPC call for method "SecondService::echoTypedef"
func (c *bazClient) EchoTypedef(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SecondService_EchoTypedef_Args,
) (clientsBazBase.UUID, map[string]string, error) {
	var result clientsBazBaz.SecondService_EchoTypedef_Result
	var resp clientsBazBase.UUID

	success, respHeaders, err := c.client.Call(
		ctx, "SecondService", "echoTypedef", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for EchoTypedef")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SecondService_EchoTypedef_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// Call is a client RPC call for method "SimpleService::call"
func (c *bazClient) Call(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SimpleService_Call_Args,
) (map[string]string, error) {
	var result clientsBazBaz.SimpleService_Call_Result

	success, respHeaders, err := c.client.Call(
		ctx, "SimpleService", "call", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		case result.AuthErr != nil:
			err = result.AuthErr
		default:
			err = errors.New("bazClient received no result or unknown exception for Call")
		}
	}
	if err != nil {
		return nil, err
	}

	return respHeaders, err
}

// Compare is a client RPC call for method "SimpleService::compare"
func (c *bazClient) Compare(
	ctx context.Context,
	reqHeaders map[string]string,
	args *clientsBazBaz.SimpleService_Compare_Args,
) (*clientsBazBase.BazResponse, map[string]string, error) {
	var result clientsBazBaz.SimpleService_Compare_Result
	var resp *clientsBazBase.BazResponse

	success, respHeaders, err := c.client.Call(
		ctx, "SimpleService", "compare", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		case result.AuthErr != nil:
			err = result.AuthErr
		case result.OtherAuthErr != nil:
			err = result.OtherAuthErr
		default:
			err = errors.New("bazClient received no result or unknown exception for Compare")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SimpleService_Compare_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// Ping is a client RPC call for method "SimpleService::ping"
func (c *bazClient) Ping(
	ctx context.Context,
	reqHeaders map[string]string,
) (*clientsBazBase.BazResponse, map[string]string, error) {
	var result clientsBazBaz.SimpleService_Ping_Result
	var resp *clientsBazBase.BazResponse

	args := &clientsBazBaz.SimpleService_Ping_Args{}
	success, respHeaders, err := c.client.Call(
		ctx, "SimpleService", "ping", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		default:
			err = errors.New("bazClient received no result or unknown exception for Ping")
		}
	}
	if err != nil {
		return resp, nil, err
	}

	resp, err = clientsBazBaz.SimpleService_Ping_Helper.UnwrapResponse(&result)
	return resp, respHeaders, err
}

// DeliberateDiffNoop is a client RPC call for method "SimpleService::sillyNoop"
func (c *bazClient) DeliberateDiffNoop(
	ctx context.Context,
	reqHeaders map[string]string,
) (map[string]string, error) {
	var result clientsBazBaz.SimpleService_SillyNoop_Result

	args := &clientsBazBaz.SimpleService_SillyNoop_Args{}
	success, respHeaders, err := c.client.Call(
		ctx, "SimpleService", "sillyNoop", reqHeaders, args, &result,
	)

	if err == nil && !success {
		switch {
		case result.AuthErr != nil:
			err = result.AuthErr
		case result.ServerErr != nil:
			err = result.ServerErr
		default:
			err = errors.New("bazClient received no result or unknown exception for SillyNoop")
		}
	}
	if err != nil {
		return nil, err
	}

	return respHeaders, err
}
