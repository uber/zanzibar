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
	ClientID          string
	MethodNames       map[string]string
	Timeout           time.Duration
	TimeoutPerAttempt time.Duration
	RoutingKey        *string
}

// tchannelClient implements TChannelClient and makes outgoing Thrift calls.
type tchannelClient struct {
	ch                *tchannel.Channel
	sc                *tchannel.SubChannel
	serviceName       string
	clientID          string
	methodNames       map[string]string
	timeout           time.Duration
	timeoutPerAttempt time.Duration
	routingKey        *string
}

type tchannelCall struct {
	client     *tchannelClient
	call       *tchannel.OutboundCall
	success    bool
	startTime  time.Time
	finishTime time.Time
	reqBody    []byte
	resBody    []byte
	reqHeaders map[string]string
	resHeaders map[string]string
}

// NewTChannelClient returns a tchannelClient that makes calls over the given tchannel to the given thrift service.
func NewTChannelClient(ch *tchannel.Channel, opt *TChannelClientOption) TChannelClient {
	return &tchannelClient{
		ch:                ch,
		sc:                ch.GetSubChannel(opt.ServiceName),
		serviceName:       opt.ServiceName,
		clientID:          opt.ClientID,
		methodNames:       opt.MethodNames,
		timeout:           opt.Timeout,
		timeoutPerAttempt: opt.TimeoutPerAttempt,
		routingKey:        opt.RoutingKey,
	}
}

// Call makes a RPC call to the given service.
func (c *tchannelClient) Call(
	ctx context.Context,
	thriftService, methodName string,
	reqHeaders map[string]string,
	req, resp RWTStruct,
) (success bool, resHeaders map[string]string, err error) {
	retryOpts := tchannel.RetryOptions{
		TimeoutPerAttempt: c.timeoutPerAttempt,
	}
	ctxBuilder := tchannel.NewContextBuilder(c.timeout).
		SetParentContext(ctx).
		SetRetryOptions(&retryOpts)
	if c.routingKey != nil {
		ctxBuilder.SetRoutingKey(*c.routingKey)
	}
	ctx, cancel := ctxBuilder.Build()
	defer cancel()

	serviceMethod := thriftService + "::" + methodName

	call := &tchannelCall{
		client:     c,
		reqHeaders: reqHeaders,
	}
	defer call.finish()
	call.start()

	err = c.ch.RunWithRetry(ctx, func(ctx netContext.Context, rs *tchannel.RequestState) (cerr error) {
		call.resHeaders, call.success = nil, false

		call.call, cerr = c.sc.BeginCall(ctx, serviceMethod, &tchannel.CallOptions{
			Format:       tchannel.Thrift,
			RequestState: rs,
		})
		if cerr != nil {
			return errors.Wrapf(err, "could not begin outbound call: %s", c.serviceName)
		}

		// trace request
		reqHeaders = tchannel.InjectOutboundSpan(call.call.Response(), reqHeaders)

		if cerr := call.writeReqHeaders(reqHeaders); cerr != nil {
			return cerr
		}
		if cerr := call.writeReqBody(req); cerr != nil {
			return cerr
		}

		response := call.call.Response()
		if cerr = call.readResHeaders(response); cerr != nil {
			return cerr
		}
		if cerr = call.readResBody(response, resp); cerr != nil {
			return cerr
		}

		return cerr
	})

	if err != nil {
		// Do not wrap system errors.
		if _, ok := err.(tchannel.SystemError); ok {
			return call.success, nil, err
		}
		return call.success, nil, errors.Wrapf(err, "could not make outbound call: %s", c.serviceName)
	}

	return call.success, call.resHeaders, err
}

func (c *tchannelCall) start() {
	c.startTime = time.Now()
}

func (c *tchannelCall) finish() {
	c.finishTime = time.Now()
}

// writeReqHeaders writes request headers to arg2
func (c *tchannelCall) writeReqHeaders(reqHeaders map[string]string) error {
	c.reqHeaders = reqHeaders

	twriter, err := c.call.Arg2Writer()
	if err != nil {
		return errors.Wrapf(err, "could not create arg2writer for outbound call %s: ", c.client.serviceName)
	}

	if err := WriteHeaders(twriter, reqHeaders); err != nil {
		_ = twriter.Close()
		return errors.Wrapf(err, "could not write headers for outbound call %s: ", c.client.serviceName)
	}

	if err := twriter.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg2writer for outbound call %s: ", c.client.serviceName)
	}

	return nil
}

// writeReqBody writes request body to arg3
func (c *tchannelCall) writeReqBody(req RWTStruct) error {
	structWireValue, err := req.ToWire()
	if err != nil {
		return errors.Wrapf(err, "could not write request for outbound call %s: ", c.client.serviceName)
	}
	c.reqBody = structWireValue.GetBinary()

	twriter, err := c.call.Arg3Writer()
	if err != nil {
		return errors.Wrapf(err, "could not create arg3writer for outbound call %s: ", c.client.serviceName)
	}

	if err := protocol.Binary.Encode(structWireValue, twriter); err != nil {
		_ = twriter.Close()
		return errors.Wrapf(err, "could not write request for outbound call %s: ", c.client.serviceName)
	}

	if err := twriter.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg3writer for outbound request %s: ", c.client.serviceName)
	}

	return nil
}

// readResHeaders read response headers from arg2
func (c *tchannelCall) readResHeaders(response *tchannel.OutboundCallResponse) error {
	treader, err := response.Arg2Reader()
	if err != nil {
		// Do not wrap system errors.
		if _, ok := err.(tchannel.SystemError); ok {
			return err
		}
		return errors.Wrapf(err, "could not create arg2reader for outbound call response: %s", c.client.serviceName)
	}

	if c.resHeaders, err = ReadHeaders(treader); err != nil {
		_ = treader.Close()
		return errors.Wrapf(err, "could not read headers for outbound call response: %s", c.client.serviceName)
	}

	if err := EnsureEmpty(treader, "reading response headers"); err != nil {
		_ = treader.Close()
		return errors.Wrapf(err, "could not ensure arg2reader is empty for outbound call response: %s", c.client.serviceName)
	}

	if err := treader.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg2reader for outbound call response: %s", c.client.serviceName)
	}

	// must have called Arg2Reader before calling ApplicationError
	c.success = !response.ApplicationError()
	return nil
}

// readResHeaders read response headers from arg2
func (c *tchannelCall) readResBody(response *tchannel.OutboundCallResponse, resp RWTStruct) error {
	treader, err := response.Arg3Reader()
	if err != nil {
		return errors.Wrapf(err, "could not create arg3Reader for outbound call response: %s", c.client.serviceName)
	}

	if err := ReadStruct(treader, resp); err != nil {
		_ = treader.Close()
		return errors.Wrapf(err, "could not read outbound call response: %s", c.client.serviceName)
	}

	structWireValue, err := resp.ToWire()
	if err != nil {
		return errors.Wrapf(err, "could not write response for outbound call response: %s: ", c.client.serviceName)
	}
	c.resBody = structWireValue.GetBinary()

	if err := EnsureEmpty(treader, "reading response body"); err != nil {
		_ = treader.Close()
		return errors.Wrapf(err, "could not ensure arg3reader is empty for outbound call response: %s", c.client.serviceName)
	}

	if err := treader.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg3reader for outbound call response: %s", c.client.serviceName)
	}

	return nil
}
