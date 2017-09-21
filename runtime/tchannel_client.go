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

package zanzibar

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/uber/tchannel-go"

	"go.uber.org/thriftrw/protocol"
	netContext "golang.org/x/net/context"
)

// TChannelClientOption is used when creating a new tchannelClient
type TChannelClientOption struct {
	ServiceName       string
	Timeout           time.Duration
	TimeoutPerAttempt time.Duration
	RoutingKey        *string
}

// tchannelClient implements TChannelClient and makes outgoing Thrift calls.
type tchannelClient struct {
	ch                *tchannel.Channel
	sc                *tchannel.SubChannel
	serviceName       string
	timeout           time.Duration
	timeoutPerAttempt time.Duration
	routingKey        *string
}

// NewTChannelClient returns a tchannelClient that makes calls over the given tchannel to the given thrift service.
func NewTChannelClient(ch *tchannel.Channel, opt *TChannelClientOption) TChannelClient {
	client := &tchannelClient{
		ch:                ch,
		sc:                ch.GetSubChannel(opt.ServiceName),
		serviceName:       opt.ServiceName,
		timeout:           opt.Timeout,
		timeoutPerAttempt: opt.TimeoutPerAttempt,
		routingKey:        opt.RoutingKey,
	}
	return client
}

func (c *tchannelClient) writeArgs(call *tchannel.OutboundCall, headers map[string]string, req RWTStruct) error {
	structWireValue, err := req.ToWire()
	if err != nil {
		return errors.Wrapf(err, "could not write request for outbound call %s: ", c.serviceName)
	}

	twriter, err := call.Arg2Writer()
	if err != nil {
		return errors.Wrapf(err, "could not create arg2writer for outbound call %s: ", c.serviceName)
	}

	headers = tchannel.InjectOutboundSpan(call.Response(), headers)
	if err := WriteHeaders(twriter, headers); err != nil {
		_ = twriter.Close()

		return errors.Wrapf(err, "could not write headers for outbound call %s: ", c.serviceName)
	}
	if err := twriter.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg2writer for outbound call %s: ", c.serviceName)
	}

	twriter, err = call.Arg3Writer()
	if err != nil {
		return errors.Wrapf(err, "could not create arg3writer for outbound call %s: ", c.serviceName)
	}

	err = protocol.Binary.Encode(structWireValue, twriter)
	if err != nil {
		_ = twriter.Close()

		return errors.Wrapf(err, "could not write request for outbound call %s: ", c.serviceName)
	}

	return twriter.Close()
}

// readResponse reads the response struct into resp, and returns:
// (response headers, whether there was an application error, unexpected error).
func (c *tchannelClient) readResponse(response *tchannel.OutboundCallResponse, resp RWTStruct) (bool, map[string]string, error) {
	treader, err := response.Arg2Reader()
	if err != nil {
		// Do not wrap system errors.
		if _, ok := err.(tchannel.SystemError); ok {
			return false, nil, err
		}

		return false, nil, errors.Wrapf(err, "could not create arg2reader for outbound call response: %s", c.serviceName)
	}

	headers, err := ReadHeaders(treader)
	if err != nil {
		_ = treader.Close()

		return false, nil, errors.Wrapf(err, "could not read headers for outbound call response: %s", c.serviceName)
	}

	if err := EnsureEmpty(treader, "reading response headers"); err != nil {
		_ = treader.Close()

		return false, nil, errors.Wrapf(err, "could not ensure arg2reader is empty for outbound call response: %s", c.serviceName)
	}

	if err := treader.Close(); err != nil {
		return false, nil, errors.Wrapf(err, "could not close arg2reader for outbound call response: %s", c.serviceName)
	}

	success := !response.ApplicationError()
	treader, err = response.Arg3Reader()
	if err != nil {
		return success, headers, errors.Wrapf(err, "could not create arg3Reader for outbound call response: %s", c.serviceName)
	}

	if err := ReadStruct(treader, resp); err != nil {
		_ = treader.Close()

		return success, headers, errors.Wrapf(err, "could not read outbound call response: %s", c.serviceName)
	}

	if err := EnsureEmpty(treader, "reading response body"); err != nil {
		_ = treader.Close()

		return false, nil, errors.Wrapf(err, "could not ensure arg3reader is empty for outbound call response: %s", c.serviceName)
	}

	return success, headers, treader.Close()
}

// Call makes a RPC call to the given service.
func (c *tchannelClient) Call(ctx context.Context, thriftService, methodName string, reqHeaders map[string]string, req, resp RWTStruct) (bool, map[string]string, error) {
	var respHeaders map[string]string
	var isOK bool

	retryOpts := &tchannel.RetryOptions{
		TimeoutPerAttempt: c.timeoutPerAttempt,
	}
	ctxBuilder := tchannel.NewContextBuilder(c.timeout).
		SetParentContext(ctx).
		SetRetryOptions(retryOpts)
	if c.routingKey != nil {
		ctxBuilder.SetRoutingKey(*c.routingKey)
	}
	ctx, cancel := ctxBuilder.Build()
	defer cancel()

	arg1 := thriftService + "::" + methodName
	err := c.ch.RunWithRetry(ctx, func(ctx netContext.Context, rs *tchannel.RequestState) error {
		respHeaders, isOK = nil, false

		call, err := c.sc.BeginCall(ctx, arg1, &tchannel.CallOptions{
			Format:       tchannel.Thrift,
			RequestState: rs,
		})
		if err != nil {
			return errors.Wrapf(err, "could not begin outbound call: %s", c.serviceName)
		}

		if err := c.writeArgs(call, reqHeaders, req); err != nil {
			return err
		}

		isOK, respHeaders, err = c.readResponse(call.Response(), resp)
		return err
	})
	if err != nil {
		// Do not wrap system errors.
		if _, ok := err.(tchannel.SystemError); ok {
			return false, nil, err
		}

		return false, nil, errors.Wrapf(err, "could not make outbound call: %s", c.serviceName)
	}

	return isOK, respHeaders, nil
}
