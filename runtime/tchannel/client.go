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

package tchannel

import (
	"github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/thrift"
	"golang.org/x/net/context"
)

// client implements TChanClient and makes outgoing Thrift calls.
type Client struct {
	ch          *tchannel.Channel
	sc          *tchannel.SubChannel
	serviceName string
}

// NewClient returns a Client that makes calls over the given tchannel to the given Hyperbahn service.
func NewClient(ch *tchannel.Channel, serviceName string) TChanClient {
	client := &Client{
		ch:          ch,
		sc:          ch.GetSubChannel(serviceName),
		serviceName: serviceName,
	}
	return client
}

func writeArgs(call *tchannel.OutboundCall, headers map[string]string, req RWTStruct) error {
	writer, err := call.Arg2Writer()
	if err != nil {
		return err
	}
	headers = tchannel.InjectOutboundSpan(call.Response(), headers)
	if err := thrift.WriteHeaders(writer, headers); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	writer, err = call.Arg3Writer()
	if err != nil {
		return err
	}

	if err := WriteStruct(writer, req); err != nil {
		return err
	}

	return writer.Close()
}

// readResponse reads the response struct into resp, and returns:
// (response headers, whether there was an application error, unexpected error).
func readResponse(response *tchannel.OutboundCallResponse, resp RWTStruct) (map[string]string, bool, error) {
	reader, err := response.Arg2Reader()
	if err != nil {
		return nil, false, err
	}

	headers, err := thrift.ReadHeaders(reader)
	if err != nil {
		return nil, false, err
	}

	if err := EnsureEmpty(reader, "reading response headers"); err != nil {
		return nil, false, err
	}

	if err := reader.Close(); err != nil {
		return nil, false, err
	}

	success := !response.ApplicationError()
	reader, err = response.Arg3Reader()
	if err != nil {
		return headers, success, err
	}

	if err := ReadStruct(reader, resp); err != nil {
		return headers, success, err
	}

	if err := EnsureEmpty(reader, "reading response body"); err != nil {
		return nil, false, err
	}

	return headers, success, reader.Close()
}

func (c *Client) Call(ctx thrift.Context, thriftService, methodName string, req, resp RWTStruct) (bool, error) {
	var (
		headers = ctx.Headers()

		respHeaders map[string]string
		isOK        bool
	)

	err := c.ch.RunWithRetry(ctx, func(ctx context.Context, rs *tchannel.RequestState) error {
		respHeaders, isOK = nil, false

		call, err := c.sc.BeginCall(ctx, thriftService+"::"+methodName, &tchannel.CallOptions{
			Format:       tchannel.Thrift,
			RequestState: rs,
		})
		if err != nil {
			return err
		}

		if err := writeArgs(call, headers, req); err != nil {
			return err
		}

		respHeaders, isOK, err = readResponse(call.Response(), resp)
		return err
	})
	if err != nil {
		return false, err
	}

	ctx.SetResponseHeaders(respHeaders)
	return isOK, nil
}
