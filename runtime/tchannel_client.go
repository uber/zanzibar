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
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go"
	"go.uber.org/zap"
	netContext "golang.org/x/net/context"

	"github.com/uber/zanzibar/runtime/ruleengine"
)

const (
	logFieldClientID = "clientID"
	// thrift service::method of client thrift spec
	logFieldClientThriftMethod = "clientThriftMethod"
	// the backend service corresponding to the client
	logFieldClientService = "clientService"
	// the method name for a particular client method call
	logFieldClientMethod = "clientMethod"
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

	// Dynamically determine which alternate channel to call dynamically based on ruleEngine,
	// else fallback to default routing
	RuleEngine ruleengine.RuleEngine

	// list of headers which would be looked for matching a request with ruleEngine
	HeaderPatterns []string

	// the header key that is used together with the request uuid on context to
	// form a header when sending the request to downstream, e.g. "x-request-uuid"
	RequestUUIDHeaderKey string

	// AltChannelMap is a map for dynamic lookup of alternative channels
	AltChannelMap map[string]*tchannel.SubChannel
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
	shardKey          *string
	metrics           ContextMetrics
	contextExtractor  ContextExtractor

	requestUUIDHeaderKey string
	ruleEngine           ruleengine.RuleEngine
	headerPatterns       []string
	altChannelMap        map[string]*tchannel.SubChannel
}

// NewTChannelClient is deprecated, use NewTChannelClientContext instead
func NewTChannelClient(
	ch *tchannel.Channel,
	logger *zap.Logger,
	scope tally.Scope,
	contextExtractor ContextExtractor,
	opt *TChannelClientOption,
) *TChannelClient {
	return NewTChannelClientContext(
		ch,
		logger,
		NewContextMetrics(scope),
		contextExtractor,
		opt,
	)
}

// NewTChannelClientContext returns a TChannelClient that makes calls over the given tchannel to the given thrift service.
func NewTChannelClientContext(
	ch *tchannel.Channel,
	logger *zap.Logger,
	metrics ContextMetrics,
	contextExtractor ContextExtractor,
	opt *TChannelClientOption,
) *TChannelClient {
	numMethods := len(opt.MethodNames)
	loggers := make(map[string]*zap.Logger, numMethods)

	for serviceMethod, methodName := range opt.MethodNames {
		loggers[serviceMethod] = logger.With(
			zap.String(logFieldClientID, opt.ClientID),
			zap.String(logFieldClientService, opt.ServiceName),
			zap.String(logFieldClientMethod, methodName),
			zap.String(logFieldClientThriftMethod, serviceMethod),
		)
	}

	client := &TChannelClient{
		ch:                   ch,
		sc:                   ch.GetSubChannel(opt.ServiceName),
		serviceName:          opt.ServiceName,
		ClientID:             opt.ClientID,
		methodNames:          opt.MethodNames,
		timeout:              opt.Timeout,
		timeoutPerAttempt:    opt.TimeoutPerAttempt,
		routingKey:           opt.RoutingKey,
		Loggers:              loggers,
		metrics:              metrics,
		contextExtractor:     contextExtractor,
		requestUUIDHeaderKey: opt.RequestUUIDHeaderKey,
		ruleEngine:           opt.RuleEngine,
		headerPatterns:       opt.HeaderPatterns,
		altChannelMap:        opt.AltChannelMap,
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

	return c.call(ctx, call, reqHeaders, req, resp)
}

func (c *TChannelClient) call(
	ctx context.Context,
	call *tchannelOutboundCall,
	reqHeaders map[string]string,
	req, resp RWTStruct,
) (success bool, resHeaders map[string]string, err error) {
	defer func() { call.finish(ctx, err) }()
	call.start()

	reqUUID := RequestUUIDFromCtx(ctx)
	if reqUUID != "" {
		if reqHeaders == nil {
			reqHeaders = make(map[string]string)
		}
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

	sk := GetShardKeyFromCtx(ctx)
	if sk != "" {
		ctxBuilder.SetShardKey(sk)
	}

	ctx, cancel := ctxBuilder.Build()
	defer cancel()

	err = c.ch.RunWithRetry(ctx, func(ctx netContext.Context, rs *tchannel.RequestState) (cerr error) {
		call.resHeaders, call.success = nil, false

		sc, ctx := c.getDynamicChannelWithFallback(reqHeaders, c.sc, ctx)
		call.call, cerr = sc.BeginCall(ctx, call.serviceMethod, &tchannel.CallOptions{
			Format:          tchannel.Thrift,
			ShardKey:        GetShardKeyFromCtx(ctx),
			RequestState:    rs,
			RoutingDelegate: GetRoutingDelegateFromCtx(ctx),
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

// first rule match, would be the choosen channel. if nothing matches fallback to default channel
func (c *TChannelClient) getDynamicChannelWithFallback(reqHeaders map[string]string,
	sc *tchannel.SubChannel, ctx netContext.Context) (*tchannel.SubChannel, netContext.Context) {
	ch := sc
	if c.ruleEngine == nil {
		return ch, ctx
	}
	for _, headerPattern := range c.headerPatterns {
		// this header is not present, so can't match a rule
		headerPatternVal, ok := reqHeaders[headerPattern]
		if !ok {
			continue
		}
		val, match := c.ruleEngine.GetValue(headerPattern, strings.ToLower(headerPatternVal))
		// if rule doesn't match, continue with a next input
		if !match {
			continue
		}
		serviceDetails := val.([]string)
		// we know service has a channel, as this was constructed in c'tor
		ch = c.altChannelMap[serviceDetails[0]]
		if len(serviceDetails) > 1 {
			ctx = WithRoutingDelegate(ctx, serviceDetails[1])
		}
		return ch, ctx
	}
	// if nothing matches return the default channel/**/
	return ch, ctx
}
