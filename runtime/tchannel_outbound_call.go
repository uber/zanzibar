// Copyright (c) 2022 Uber Technologies, Inc.
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
	"fmt"
	"time"

	stream "go.uber.org/thriftrw/protocol/stream"

	"go.uber.org/thriftrw/protocol/binary"

	"github.com/pkg/errors"
	"github.com/uber/tchannel-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type tchannelOutboundCall struct {
	client        *TChannelClient
	call          *tchannel.OutboundCall
	methodName    string
	serviceMethod string
	success       bool
	startTime     time.Time
	finishTime    time.Time
	duration      time.Duration
	reqHeaders    map[string]string
	resHeaders    map[string]string
	contextLogger ContextLogger
	metrics       ContextMetrics
}

func (c *tchannelOutboundCall) start() {
	c.startTime = time.Now()
}

func (c *tchannelOutboundCall) finish(ctx context.Context, err error) {
	c.finishTime = time.Now()

	// emit metrics
	if err != nil {
		errCause := tchannel.GetSystemErrorCode(errors.Cause(err))
		scopeTags := map[string]string{scopeTagError: errCause.MetricsKey()}
		ctx = WithScopeTags(ctx, scopeTags)
		c.metrics.IncCounter(ctx, clientSystemErrors, 1)
	} else if !c.success {
		c.metrics.IncCounter(ctx, clientAppErrors, 1)
	} else {
		c.metrics.IncCounter(ctx, clientSuccess, 1)
	}
	delta := c.finishTime.Sub(c.startTime)
	c.metrics.RecordTimer(ctx, clientLatency, delta)
	c.metrics.RecordHistogramDuration(ctx, clientLatencyHist, delta)
	c.duration = delta

	// write logs
	fields := c.logFields(ctx)
	if err == nil {
		c.contextLogger.Debug(ctx, "Finished an outgoing client TChannel request", fields...)
	} else {
		fields = append(fields, zap.Error(err))
		c.contextLogger.Warn(ctx, "Failed to send outgoing client TChannel request", fields...)
	}
}

func (c *tchannelOutboundCall) logFields(ctx context.Context) []zapcore.Field {
	var hostPort string
	if c.call != nil {
		hostPort = c.call.RemotePeer().HostPort
	} else {
		hostPort = "unknown"
	}
	fields := []zapcore.Field{
		zap.String("remoteAddr", hostPort),
		zap.Time("timestamp-started", c.startTime),
		zap.Time("timestamp-finished", c.finishTime),
	}

	headers := map[string]string{}
	for k, v := range c.reqHeaders {
		s := fmt.Sprintf("%s-%s", logFieldClientRequestHeaderPrefix, k)
		headers[s] = v
	}

	for k, v := range c.resHeaders {
		s := fmt.Sprintf("%s-%s", logFieldClientResponseHeaderPrefix, k)
		headers[s] = v
	}

	// If an extractor function is provided, use it, else copy all the headers
	if c.client != nil && c.client.contextExtractor != nil {
		ctx = WithEndpointRequestHeadersField(ctx, headers)
		fields = append(fields, c.client.contextExtractor.ExtractLogFields(ctx)...)
	} else {
		for k, v := range headers {
			fields = append(fields, zap.String(k, v))
		}
	}

	fields = append(fields, GetLogFieldsFromCtx(ctx)...)
	return fields
}

// writeReqHeaders writes request headers to arg2
func (c *tchannelOutboundCall) writeReqHeaders(reqHeaders map[string]string) error {
	c.reqHeaders = reqHeaders

	twriter, err := c.call.Arg2Writer()
	if err != nil {
		return errors.Wrapf(
			err, "Could not create arg2writer for outbound %s.%s (%s %s) request",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := WriteHeaders(twriter, reqHeaders); err != nil {
		_ = twriter.Close()
		return errors.Wrapf(
			err, "Could not write headers for outbound %s.%s (%s %s) request",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := twriter.Close(); err != nil {
		return errors.Wrapf(
			err, "Could not close arg2writer for outbound %s.%s (%s %s) request",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}

	return nil
}

// writeReqBody writes request body to arg3
func (c *tchannelOutboundCall) writeReqBody(ctx context.Context, req RWTStruct) error {
	twriter, err := c.call.Arg3Writer()
	if err != nil {
		return errors.Wrapf(
			err, "Could not create arg3writer for outbound %s.%s (%s %s) request",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	sw := binary.Default.Writer(twriter)
	defer func(sw stream.Writer) {
		e := sw.Close()
		if e != nil {
			err = errors.Wrapf(e, "Could not close stream writer for outbound %s.%s (%s %s) request",
				c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod)
		}
	}(sw)

	if err := req.Encode(sw); err != nil {
		_ = twriter.Close()
		return errors.Wrapf(
			err, "Could not write request for outbound %s.%s (%s %s) request",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := twriter.Close(); err != nil {
		return errors.Wrapf(
			err, "Could not close arg3writer for outbound %s.%s (%s %s) request",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}

	// request sent when arg3writer is closed
	c.metrics.IncCounter(ctx, clientRequest, 1)
	return nil
}

// readResHeaders read response headers from arg2
func (c *tchannelOutboundCall) readResHeaders(response *tchannel.OutboundCallResponse) error {
	treader, err := response.Arg2Reader()
	if err != nil {
		// Do not wrap system errors.
		if _, ok := err.(tchannel.SystemError); ok {
			return err
		}
		return errors.Wrapf(
			err, "Could not create arg2reader for outbound %s.%s (%s %s) response",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if c.resHeaders, err = ReadHeaders(treader); err != nil {
		_ = treader.Close()
		return errors.Wrapf(
			err, "Could not read headers for outbound %s.%s (%s %s) response",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := EnsureEmpty(treader, "reading response headers"); err != nil {
		_ = treader.Close()
		return errors.Wrapf(
			err, "Could not ensure arg2reader is empty for outbound %s.%s (%s %s) response",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := treader.Close(); err != nil {
		return errors.Wrapf(
			err, "Could not close arg2reader for outbound %s.%s (%s %s) response",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}

	// must have called Arg2Reader before calling ApplicationError
	c.success = !response.ApplicationError()
	return nil
}

// readResBody read response body from arg3
func (c *tchannelOutboundCall) readResBody(ctx context.Context, response *tchannel.OutboundCallResponse, resp RWTStruct) error {
	treader, err := response.Arg3Reader()
	if err != nil {
		return errors.Wrapf(
			err, "Could not create arg3Reader for outbound %s.%s (%s %s) response",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := ReadStruct(treader, resp); err != nil {
		_ = treader.Close()
		c.metrics.IncCounter(ctx, clientTchannelUnmarshalError, 1)
		return errors.Wrapf(
			err, "Could not read outbound %s.%s (%s %s) response",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := EnsureEmpty(treader, "reading response body"); err != nil {
		_ = treader.Close()
		return errors.Wrapf(
			err, "Could not ensure arg3reader is empty for outbound %s.%s (%s %s) response",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}
	if err := treader.Close(); err != nil {
		return errors.Wrapf(
			err, "Could not close arg3reader outbound %s.%s (%s %s) response",
			c.client.ClientID, c.methodName, c.client.serviceName, c.serviceMethod,
		)
	}

	return nil
}
