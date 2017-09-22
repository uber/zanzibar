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
	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go"
	"go.uber.org/thriftrw/protocol"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	metrics           map[string]*OutboundTChannelMetrics
}

type tchannelCall struct {
	client        *tchannelClient
	call          *tchannel.OutboundCall
	methodName    string
	serviceMethod string
	success       bool
	startTime     time.Time
	finishTime    time.Time
	reqHeaders    map[string]string
	resHeaders    map[string]string
	logger        *zap.Logger
	metrics       *OutboundTChannelMetrics
}

// NewTChannelClient returns a tchannelClient that makes calls over the given tchannel to the given thrift service.
func NewTChannelClient(
	ch *tchannel.Channel,
	logger *zap.Logger,
	scope tally.Scope,
	opt *TChannelClientOption,
) TChannelClient {
	loggers := make(map[string]*zap.Logger, len(opt.MethodNames))
	metrics := make(map[string]*OutboundTChannelMetrics, len(opt.MethodNames))
	for serviceMethod, methodName := range opt.MethodNames {
		loggers[serviceMethod] = logger.With(
			zap.String("clientID", opt.ClientID),
			zap.String("methodName", methodName),
			zap.String("serviceName", opt.ServiceName),
			zap.String("serviceMethod", serviceMethod),
		)
		metrics[serviceMethod] = NewOutboundTChannelMetrics(scope.Tagged(map[string]string{
			"client":          opt.ClientID,
			"method":          methodName,
			"target-service":  opt.ServiceName,
			"target-endpoint": serviceMethod,
		}))
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
		metrics:           metrics,
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
		metrics:       c.metrics[serviceMethod],
	}
	defer func() { call.finish(err) }()
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

func (c *tchannelCall) finish(err error) {
	c.finishTime = time.Now()

	// emit metrics
	if err != nil {
		c.metrics.SystemErrors.Inc(1)
	} else if !c.success {
		c.metrics.AppErrors.Inc(1)
	} else {
		c.metrics.Success.Inc(1)
	}
	c.metrics.Latency.Record(c.finishTime.Sub(c.startTime))

	// write logs
	c.logger.Info("Finished an outgoing client TChannel request", c.logFields()...)
}

func (c *tchannelCall) logFields() []zapcore.Field {
	fields := []zapcore.Field{
		zap.String("remoteAddr", c.call.RemotePeer().HostPort),
		zap.Time("timestamp-started", c.startTime),
		zap.Time("timestamp-finished", c.finishTime),
	}

	for k, v := range c.reqHeaders {
		fields = append(fields, zap.String("Request-Header-"+k, v))
	}
	for k, v := range c.resHeaders {
		fields = append(fields, zap.String("Response-Header-"+k, v))
	}

	// TODO: log jaeger trace span

	return fields
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

	// request sent when arg3writer is closed
	c.metrics.Sent.Inc(1)
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
