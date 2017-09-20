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
	"go.uber.org/zap"
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
	loggers           map[string]*zap.Logger
}

type tchannelCall struct {
	client        *tchannelClient
	call          *tchannel.OutboundCall
	methodName    string
	serviceMethod string
	success       bool
	startTime     time.Time
	finishTime    time.Time
	reqBody       []byte
	resBody       []byte
	reqHeaders    map[string]string
	resHeaders    map[string]string
	logger        *zap.Logger
}

// NewTChannelClient returns a tchannelClient that makes calls over the given tchannel to the given thrift service.
func NewTChannelClient(
	ch *tchannel.Channel,
	logger *zap.Logger,
	opt *TChannelClientOption,
) TChannelClient {
	loggers := make(map[string]*zap.Logger, len(opt.MethodNames))
	for serviceMethod, methodName := range opt.MethodNames {
		loggers[serviceMethod] = logger.With(
			zap.String("clientID", opt.ClientID),
			zap.String("methodName", methodName),
			zap.String("target-service", opt.ServiceName),
			zap.String("target-endpoint", serviceMethod),
		)
	}

	return &tchannelClient{
		ch:                ch,
		sc:                ch.GetSubChannel(opt.ServiceName),
		serviceName:       opt.ServiceName,
		clientID:          opt.ClientID,
		methodNames:       opt.MethodNames,
		timeout:           opt.Timeout,
		timeoutPerAttempt: opt.TimeoutPerAttempt,
		routingKey:        opt.RoutingKey,
		loggers:           loggers,
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
		client:        c,
		methodName:    c.methodNames[serviceMethod],
		serviceMethod: serviceMethod,
		reqHeaders:    reqHeaders,
		logger:        c.loggers[serviceMethod],
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
			call.logger.Error("Could not begin outbound request", zap.Error(cerr))
			return errors.Wrapf(
				err, "Could not begin outbound %s.%s (%s %s) request",
				call.client.clientID, call.methodName, call.client.serviceName, call.serviceMethod,
			)
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
		call.logger.Error("Could not make outbound request", zap.Error(err))
		// Do not wrap system errors.
		if _, ok := err.(tchannel.SystemError); ok {
			return call.success, nil, err
		}
		return call.success, nil, errors.Wrapf(
			err, "Could not make outbound %s.%s (%s %s) response",
			call.client.clientID, call.methodName, call.client.serviceName, call.serviceMethod,
		)
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
		c.logger.Error("Could not create arg2writer for outbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not create arg2writer for outbound %s.%s (%s %s) request",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := WriteHeaders(twriter, reqHeaders); err != nil {
		_ = twriter.Close()
		c.logger.Error("Could not write headers for outbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not write headers for outbound %s.%s (%s %s) request",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := twriter.Close(); err != nil {
		c.logger.Error("Could not close arg2writer for outbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not close arg2writer for outbound %s.%s (%s %s) request",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}

	return nil
}

// writeReqBody writes request body to arg3
func (c *tchannelCall) writeReqBody(req RWTStruct) error {
	structWireValue, err := req.ToWire()
	if err != nil {
		c.logger.Error("Could not write request for outbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not write request for outbound %s.%s (%s %s) request",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	c.reqBody = structWireValue.GetBinary()

	twriter, err := c.call.Arg3Writer()
	if err != nil {
		c.logger.Error("Could not create arg3writer for outbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not create arg3writer for outbound %s.%s (%s %s) request",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := protocol.Binary.Encode(structWireValue, twriter); err != nil {
		_ = twriter.Close()
		c.logger.Error("Could not write request for outbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not write request for outbound %s.%s (%s %s) request",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := twriter.Close(); err != nil {
		c.logger.Error("Could not close arg3writer for outbound request", zap.Error(err))
		return errors.Wrapf(
			err, "Could not close arg3writer for outbound %s.%s (%s %s) request",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}

	return nil
}

// readResHeaders read response headers from arg2
func (c *tchannelCall) readResHeaders(response *tchannel.OutboundCallResponse) error {
	treader, err := response.Arg2Reader()
	if err != nil {
		c.logger.Error("Could not create arg2reader for outbound response", zap.Error(err))
		// Do not wrap system errors.
		if _, ok := err.(tchannel.SystemError); ok {
			return err
		}
		return errors.Wrapf(
			err, "Could not create arg2reader for outbound %s.%s (%s %s) response",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if c.resHeaders, err = ReadHeaders(treader); err != nil {
		_ = treader.Close()
		c.logger.Error("Could not read headers for outbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not read headers for outbound %s.%s (%s %s) response",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := EnsureEmpty(treader, "reading response headers"); err != nil {
		_ = treader.Close()
		c.logger.Error("Could not ensure arg2reader is empty for outbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not ensure arg2reader is empty for outbound %s.%s (%s %s) response",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := treader.Close(); err != nil {
		c.logger.Error("Could not close arg2reader for outbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not close arg2reader for outbound %s.%s (%s %s) response",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}

	// must have called Arg2Reader before calling ApplicationError
	c.success = !response.ApplicationError()
	return nil
}

// readResHeaders read response headers from arg2
func (c *tchannelCall) readResBody(response *tchannel.OutboundCallResponse, resp RWTStruct) error {
	treader, err := response.Arg3Reader()
	if err != nil {
		c.logger.Error("Could not create arg3Reader for outbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not create arg3Reader for outbound %s.%s (%s %s) response",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := ReadStruct(treader, resp); err != nil {
		_ = treader.Close()
		c.logger.Error("Could not read outbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not read outbound %s.%s (%s %s) response",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}

	structWireValue, err := resp.ToWire()
	if err != nil {
		c.logger.Error("Could not serialize response for outbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not serialize response for outbound %s.%s (%s %s) response",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	c.resBody = structWireValue.GetBinary()

	if err := EnsureEmpty(treader, "reading response body"); err != nil {
		_ = treader.Close()
		c.logger.Error("Could not ensure arg3reader is empty for outbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not ensure arg3reader is empty for outbound %s.%s (%s %s) response",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := treader.Close(); err != nil {
		c.logger.Error("Could not close arg3reader for outbound response", zap.Error(err))
		return errors.Wrapf(
			err, "Could not close arg3reader outbound %s.%s (%s %s) response",
			c.client.clientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}

	return nil
}
