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

	netContext "golang.org/x/net/context"
)

// TChannelClientOption is used when creating a new TChannelClient
type TChannelClientOption struct {
	ServiceName       string
	Timeout           time.Duration
	TimeoutPerAttempt time.Duration
}

// TChannelClient implements TChanClient and makes outgoing Thrift calls.
type TChannelClient struct {
	ch                *tchannel.Channel
	sc                *tchannel.SubChannel
	serviceName       string
	timeout           time.Duration
	timeoutPerAttempt time.Duration
}

// NewTChannelClient returns a TChannelClient that makes calls over the given tchannel to the given thrift service.
func NewTChannelClient(ch *tchannel.Channel, opt *TChannelClientOption) TChanClient {
	client := &TChannelClient{
		ch:                ch,
		sc:                ch.GetSubChannel(opt.ServiceName),
		serviceName:       opt.ServiceName,
		timeout:           opt.Timeout,
		timeoutPerAttempt: opt.TimeoutPerAttempt,
	}
	return client
}

func (c *TChannelClient) writeArgs(call *tchannel.OutboundCall, headers map[string]string, req RWTStruct) error {
	writer, err := call.Arg2Writer()
	if err != nil {
		return errors.Wrapf(err, "could not create arg2writer for outbound call %s: ", c.serviceName)
	}
	headers = tchannel.InjectOutboundSpan(call.Response(), headers)
	if err := WriteHeaders(writer, headers); err != nil {
		return errors.Wrapf(err, "could not write headers for outbound call %s: ", c.serviceName)
	}
	if err := writer.Close(); err != nil {
		return errors.Wrapf(err, "could not close arg2writer for outbound call %s: ", c.serviceName)
	}

	writer, err = call.Arg3Writer()
	if err != nil {
		return errors.Wrapf(err, "could not create arg3writer for outbound call %s: ", c.serviceName)
	}

	if err := WriteStruct(writer, req); err != nil {
		return errors.Wrapf(err, "could not write request for outbound call %s: ", c.serviceName)
	}

	return writer.Close()
}

// readResponse reads the response struct into resp, and returns:
// (response headers, whether there was an application error, unexpected error).
func (c *TChannelClient) readResponse(response *tchannel.OutboundCallResponse, resp RWTStruct) (map[string]string, bool, error) {
	reader, err := response.Arg2Reader()
	if err != nil {
		return nil, false, errors.Wrapf(err, "could not create arg2reader for outbound call response: %s", c.serviceName)
	}

	headers, err := ReadHeaders(reader)
	if err != nil {
		return nil, false, errors.Wrapf(err, "could not read headers for outbound call response: %s", c.serviceName)
	}

	if err := EnsureEmpty(reader, "reading response headers"); err != nil {
		return nil, false, errors.Wrapf(err, "could not ensure arg2reader is empty for outbound call response: %s", c.serviceName)
	}

	if err := reader.Close(); err != nil {
		return nil, false, errors.Wrapf(err, "could not close arg2reader for outbound call response: %s", c.serviceName)
	}

	success := !response.ApplicationError()
	reader, err = response.Arg3Reader()
	if err != nil {
		return headers, success, errors.Wrapf(err, "could not create arg3Reader for outbound call response: %s", c.serviceName)
	}

	if err := ReadStruct(reader, resp); err != nil {
		return headers, success, errors.Wrapf(err, "could not read outbound call response: %s", c.serviceName)
	}

	if err := EnsureEmpty(reader, "reading response body"); err != nil {
		return nil, false, errors.Wrapf(err, "could not ensure arg3reader is empty for outbound call response: %s", c.serviceName)
	}

	return headers, success, reader.Close()
}

// Call makes a RPC call to the given service.
func (c *TChannelClient) Call(ctx context.Context, thriftService, methodName string, reqHeaders map[string]string, req, resp RWTStruct) (map[string]string, bool, error) {
	var respHeaders map[string]string
	var isOK bool

	retryOpts := &tchannel.RetryOptions{
		TimeoutPerAttempt: c.timeoutPerAttempt,
	}
	ctx, cancel := tchannel.NewContextBuilder(c.timeout).
		SetParentContext(ctx).
		SetRetryOptions(retryOpts).
		Build()
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

		respHeaders, isOK, err = c.readResponse(call.Response(), resp)
		return err
	})
	if err != nil {
		return nil, false, errors.Wrapf(err, "could not make outbound call: %s", c.serviceName)
	}

	return respHeaders, isOK, nil
}
