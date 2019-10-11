// Copyright (c) 2019 Uber Technologies, Inc.
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
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go"
	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/zap"
)

type tchannelInboundCall struct {
	endpoint   *TChannelEndpoint
	call       *tchannel.InboundCall
	success    bool
	responded  bool
	startTime  time.Time
	finishTime time.Time
	reqHeaders map[string]string
	resHeaders map[string]string

	// Logger logs entries with default fields that contains request meta info
	logger Logger
	// Scope emit metrics with default tags that contains request meta info
	scope tally.Scope
}

func (c *tchannelInboundCall) start() {
	c.startTime = time.Now()
}

func (c *tchannelInboundCall) finish(ctx context.Context, err error) {
	c.finishTime = time.Now()

	if err != nil {
		errCause := tchannel.GetSystemErrorCode(errors.Cause(err))
		errTag := map[string]string{scopeTagError: errCause.MetricsKey()}
		c.scope.Tagged(errTag).Counter(endpointSystemErrors).Inc(1)
	} else if !c.success {
		// The endpoint already has emitted an app-error stat in the template
	} else {
		c.scope.Counter(endpointSuccess).Inc(1)
	}
	delta := c.finishTime.Sub(c.startTime)
	c.scope.Timer(endpointLatency).Record(delta)
	c.scope.Histogram(endpointLatencyHist, tally.DefaultBuckets).RecordDuration(delta)
	c.scope.Counter(endpointRequest).Inc(1)

	fields := c.logFields(ctx)
	if err == nil {
		c.logger.Debug("Finished an incoming server TChannel request", fields...)
	} else {
		fields = append(fields, zap.Error(err))
		c.logger.Warn("Failed to serve incoming TChannel request", fields...)
	}
}

func (c *tchannelInboundCall) logFields(ctx context.Context) []zap.Field {
	fields := []zap.Field{
		zap.String("remoteAddr", c.call.RemotePeer().HostPort),
		zap.String("calling-service", c.call.CallerName()),
		zap.Time("timestamp-started", c.startTime),
		zap.Time("timestamp-finished", c.finishTime),
	}

	for k, v := range c.resHeaders {
		fields = append(fields, zap.String(
			fmt.Sprintf("%s-%s", logFieldEndpointResponseHeaderPrefix, k), v,
		))
	}

	fields = append(fields, GetLogFieldsFromCtx(ctx)...)
	return fields
}

// readReqHeaders reads request headers from arg2
func (c *tchannelInboundCall) readReqHeaders(ctx context.Context) error {
	// fail fast if timed out
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		return context.DeadlineExceeded
	}

	treader, err := c.call.Arg2Reader()
	if err != nil {
		return errors.Wrapf(err, "Could not create arg2reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	c.reqHeaders, err = ReadHeaders(treader)
	if err != nil {
		_ = treader.Close()
		return errors.Wrapf(err, "Could not read headers for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err := EnsureEmpty(treader, "reading request headers"); err != nil {
		_ = treader.Close()
		return errors.Wrapf(err, "Could not ensure arg2reader is empty for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err := treader.Close(); err != nil {
		return errors.Wrapf(err, "Could not close arg2reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}

	return nil
}

// readReqBody reads request body from arg3
func (c *tchannelInboundCall) readReqBody(ctx context.Context) (wireValue wire.Value, err error) {
	// fail fast if timed out
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		err = context.DeadlineExceeded
		return
	}

	treader, err := c.call.Arg3Reader()
	if err != nil {
		err = errors.Wrapf(err, "Could not create arg3reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}
	buf := GetBuffer()
	defer PutBuffer(buf)
	if _, err = buf.ReadFrom(treader); err != nil {
		_ = treader.Close()
		err = errors.Wrapf(err, "Could not read from arg3reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}
	wireValue, err = protocol.Binary.Decode(bytes.NewReader(buf.Bytes()), wire.TStruct)
	if err != nil {
		c.logger.Warn("Could not decode arg3 for inbound request", zap.Error(err))
		err = errors.Wrapf(err, "Could not decode arg3 for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}
	if err = EnsureEmpty(treader, "reading request body"); err != nil {
		_ = treader.Close()
		err = errors.Wrapf(err, "Could not ensure arg3reader is empty for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}
	if err = treader.Close(); err != nil {
		err = errors.Wrapf(err, "Could not close arg3reader for inbound %s.%s (%s) request",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
		return
	}

	return
}

// handle tchannel server endpoint call
func (c *tchannelInboundCall) handle(ctx context.Context, wireValue *wire.Value) (resp RWTStruct, err error) {
	// fail fast if timed out
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		err = context.DeadlineExceeded
		return
	}

	c.success, resp, c.resHeaders, err = c.endpoint.Handle(ctx, c.reqHeaders, wireValue)
	if c.endpoint.callback != nil {
		defer c.endpoint.callback(ctx, c.endpoint.Method, resp)
	}
	if err != nil {
		c.logger.Warn("Unexpected tchannel system error", zap.Error(err))
		if er := c.call.Response().SendSystemError(errors.New("Server Error")); er != nil {
			c.logger.Warn("Error sending server error response", zap.Error(er))
		}
		return
	}
	if !c.success {
		if err = c.call.Response().SetApplicationError(); err != nil {
			c.logger.Warn("Could not set application error for inbound response", zap.Error(err))
			return
		}
	}
	return
}

// writeResHeaders writes response headers to arg2
func (c *tchannelInboundCall) writeResHeaders(ctx context.Context) error {
	// fail fast if timed out
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		return context.DeadlineExceeded
	}

	twriter, err := c.call.Response().Arg2Writer()
	if err != nil {
		return errors.Wrapf(err, "Could not create arg2writer for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err = WriteHeaders(twriter, c.resHeaders); err != nil {
		_ = twriter.Close()
		return errors.Wrapf(err, "Could not write headers for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	if err = twriter.Close(); err != nil {
		return errors.Wrapf(err, "Could not close arg2writer for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	return nil
}

// writeResBody writes response body to arg3
func (c *tchannelInboundCall) writeResBody(ctx context.Context, resp RWTStruct) error {
	// fail fast if timed out
	if deadline, ok := ctx.Deadline(); ok && time.Now().After(deadline) {
		return context.DeadlineExceeded
	}

	structWireValue, err := resp.ToWire()
	if err != nil {
		if er := c.call.Response().SendSystemError(errors.New("Server Error")); er != nil {
			c.logger.Warn("Error sending server error response", zap.Error(er))
		}
		return errors.Wrapf(err, "Could not serialize arg3 for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}

	twriter, err := c.call.Response().Arg3Writer()
	if err != nil {
		return errors.Wrapf(err, "Could not create arg3writer for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	err = protocol.Binary.Encode(structWireValue, twriter)
	if err != nil {
		_ = twriter.Close()
		return errors.Wrapf(err, "Could not write arg3 for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	c.responded = true
	if err = twriter.Close(); err != nil {
		return errors.Wrapf(err, "Could not close arg3writer for inbound %s.%s (%s) response",
			c.endpoint.EndpointID, c.endpoint.HandlerID, c.endpoint.Method,
		)
	}
	return nil
}
