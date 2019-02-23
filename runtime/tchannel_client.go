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
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go"
	"go.uber.org/zap"
	netContext "golang.org/x/net/context"
)

// TChannelClientOption is used when creating a new TChannelClient
type TChannelClientOption struct {
	ServiceName       string
	ClientID          string
	Timeout           time.Duration
	TimeoutPerAttempt time.Duration
	RoutingKey        *string

	// MethodNames is a map from "ThriftService::method" to "ZanzibarMethodName",
	// where ThriftService and method are from the service's Thrift IDL, and
	// ZanzibarMethodName is the public method name exposed on the Zanzibar-generated
	// client, from the zanzibar configuration. For example, if a client named FooClient
	// has a methodMap of map[string]string{"Foo::bar":"Bar"}, then one can do
	// `FooClient.Bar()` to issue a RPC to Thrift service `Foo`'s `bar` method.
	MethodNames map[string]string

	// An alternate subchannel that can optionally be used to make a TChannel call
	// instead; e.g. can allow the service to be overridden when a "X-Zanzibar-Use-Staging"
	// header is present
	AltSubchannelName string

	// the header key that is used together with the request uuid on context to
	// form a header when sending the request to downstream, e.g. "x-request-uuid"
	RequestUUIDHeaderKey string
}

// TChannelClient implements TChannelCaller and makes outgoing Thrift calls.
type TChannelClient struct {
	ClientID string
	Loggers  map[string]*zap.Logger

	ch                *tchannel.Channel
	sc                *tchannel.SubChannel
	scAlt             *tchannel.SubChannel
	serviceName       string
	methodNames       map[string]string
	timeout           time.Duration
	timeoutPerAttempt time.Duration
	routingKey        *string
	metrics           ContextMetrics

	requestUUIDHeaderKey string
}

// NewTChannelClient is deprecated, use NewTChannelClientContext instead
func NewTChannelClient(
	ch *tchannel.Channel,
	logger *zap.Logger,
	scope tally.Scope,
	opt *TChannelClientOption,
) *TChannelClient {
	return NewTChannelClientContext(
		ch,
		logger,
		NewContextMetrics(scope),
		opt,
	)
}

// NewTChannelClientContext returns a TChannelClient that makes calls over the given tchannel to the given thrift service.
func NewTChannelClientContext(
	ch *tchannel.Channel,
	logger *zap.Logger,
	metrics ContextMetrics,
	opt *TChannelClientOption,
) *TChannelClient {
	numMethods := len(opt.MethodNames)
	loggers := make(map[string]*zap.Logger, numMethods)

	for serviceMethod, methodName := range opt.MethodNames {
		loggers[serviceMethod] = logger.With(
			zap.String("clientID", opt.ClientID),
			zap.String("methodName", methodName),
			zap.String("serviceName", opt.ServiceName),
			zap.String("serviceMethod", serviceMethod),
		)
	}

	client := &TChannelClient{
		ch:                ch,
		sc:                ch.GetSubChannel(opt.ServiceName),
		serviceName:       opt.ServiceName,
		ClientID:          opt.ClientID,
		methodNames:       opt.MethodNames,
		timeout:           opt.Timeout,
		timeoutPerAttempt: opt.TimeoutPerAttempt,
		routingKey:        opt.RoutingKey,
		Loggers:           loggers,
		metrics:           metrics,

		requestUUIDHeaderKey: opt.RequestUUIDHeaderKey,
	}
	if opt.AltSubchannelName != "" {
		client.scAlt = ch.GetSubChannel(opt.AltSubchannelName)
	}

	return client
}

// Call makes a RPC call to the given service.
func (c *TChannelClient) Call(
	ctx context.Context,
	thriftService, methodName string,
	reqHeaders map[string]string,
	req, resp RWTStruct,
) (success bool, resHeaders map[string]string, err error) {
	serviceMethod := thriftService + "::" + methodName
	scopeTags := map[string]string{
		scopeTagClient:          c.ClientID,
		scopeTagClientMethod:    methodName,
		scopeTagsTargetService:  c.serviceName,
		scopeTagsTargetEndpoint: serviceMethod,
	}

	ctx = WithScopeTags(ctx, scopeTags)
	call := &tchannelOutboundCall{
		client:        c,
		methodName:    c.methodNames[serviceMethod],
		serviceMethod: serviceMethod,
		reqHeaders:    reqHeaders,
		logger:        c.Loggers[serviceMethod],
		metrics:       c.metrics,
	}

	return c.call(ctx, call, reqHeaders, req, resp, false)
}

// CallThruAltChannel makes a RPC call using a configured alternate channel
func (c *TChannelClient) CallThruAltChannel(
	ctx context.Context,
	thriftService, methodName string,
	reqHeaders map[string]string,
	req, resp RWTStruct,
) (success bool, resHeaders map[string]string, err error) {
	serviceMethod := thriftService + "::" + methodName
	scopeTags := map[string]string{
		scopeTagClient:          c.ClientID,
		scopeTagClientMethod:    methodName,
		scopeTagsTargetService:  c.serviceName,
		scopeTagsTargetEndpoint: serviceMethod,
	}

	ctx = WithScopeTags(ctx, scopeTags)
	call := &tchannelOutboundCall{
		client:        c,
		methodName:    c.methodNames[serviceMethod],
		serviceMethod: serviceMethod,
		reqHeaders:    reqHeaders,
		logger:        c.Loggers[serviceMethod],
		metrics:       c.metrics,
	}

	return c.call(ctx, call, reqHeaders, req, resp, true)
}

func (c *TChannelClient) call(
	ctx context.Context,
	call *tchannelOutboundCall,
	reqHeaders map[string]string,
	req, resp RWTStruct,
	useAltSubchannel bool,
) (success bool, resHeaders map[string]string, err error) {
	defer func() { call.finish(ctx, err) }()
	call.start()

	reqUUID := RequestUUIDFromCtx(ctx)
	if reqUUID != "" {
		reqHeaders[c.requestUUIDHeaderKey] = reqUUID
	}

	retryOpts := tchannel.RetryOptions{
		TimeoutPerAttempt: c.timeoutPerAttempt,
	}
	ctxBuilder := tchannel.NewContextBuilder(c.timeout).
		SetParentContext(ctx).
		SetRetryOptions(&retryOpts)
	if c.routingKey != nil {
		ctxBuilder.SetRoutingKey(*c.routingKey)
	}
	rd := GetRoutingDelegateFromCtx(ctx)
	if rd != "" {
		ctxBuilder.SetRoutingDelegate(rd)
	}
	ctx, cancel := ctxBuilder.Build()
	defer cancel()

	err = c.ch.RunWithRetry(ctx, func(ctx netContext.Context, rs *tchannel.RequestState) (cerr error) {
		call.resHeaders, call.success = nil, false

		sc := c.sc
		if useAltSubchannel {
			if c.scAlt == nil {
				return errors.Errorf("alternate subchannel not configured for %s", call.client.ClientID)
			}
			sc = c.scAlt
		}

		call.call, cerr = sc.BeginCall(ctx, call.serviceMethod, &tchannel.CallOptions{
			Format:       tchannel.Thrift,
			RequestState: rs,
		})
		if cerr != nil {
			return errors.Wrapf(
				err, "Could not begin outbound %s.%s (%s %s) request",
				call.client.ClientID, call.methodName, call.client.serviceName, call.serviceMethod,
			)
		}

		// trace request
		reqHeaders = tchannel.InjectOutboundSpan(call.call.Response(), reqHeaders)

		if cerr := call.writeReqHeaders(reqHeaders); cerr != nil {
			return cerr
		}
		if cerr := call.writeReqBody(ctx, req); cerr != nil {
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
		return call.success, nil, errors.Wrapf(
			err, "Could not make outbound %s.%s (%s %s) response",
			call.client.ClientID, call.methodName, call.client.serviceName, call.serviceMethod,
		)
	}

	return call.success, call.resHeaders, err
}
