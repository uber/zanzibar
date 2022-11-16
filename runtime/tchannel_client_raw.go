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

	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go"
)

const rawClient = "raw"

// RawTChannelClient is like TChannel client, but without ClientID or MethodNames.
// Its Logs and metrics are not scoped per method, and not used for generate TChannel
// clients. It is intended to be used internally to communicate with a test service,
// and the way to that is via `MakeTChannelRequest` method defined on the test
// service. The TChannel client is not exposed, so the MethodNames are less relevant.
// The only downside is that, without the map, the client logs does not have the
// method information (because there isn't a method anyway), but the Thrift service
// and method information is still there.
type RawTChannelClient struct {
	tc            *TChannelClient
	contextLogger ContextLogger
	metrics       ContextMetrics
}

// NewRawTChannelClient returns a RawTChannelClient that makes calls over the given
// tchannel to the given thrift service. There is no guarantee that the given thrift
// service and method is valid for given Channel.
// It is intended to be used internally for testing.
func NewRawTChannelClient(
	ch *tchannel.Channel,
	contextLogger ContextLogger,
	scope tally.Scope,
	opt *TChannelClientOption,
) *RawTChannelClient {

	metrics := NewContextMetrics(scope)
	return &RawTChannelClient{
		tc:            NewTChannelClientContext(ch, contextLogger, metrics, nil, opt),
		contextLogger: contextLogger,
		metrics:       metrics,
	}
}

// Call makes a RPC call to the given service.
func (r *RawTChannelClient) Call(
	ctx context.Context,
	thriftService, methodName string,
	reqHeaders map[string]string,
	req, resp RWTStruct,
	timeoutAndRetryOptions *TimeoutAndRetryOptions,
) (success bool, resHeaders map[string]string, err error) {
	serviceMethod := thriftService + "::" + methodName

	call := &tchannelOutboundCall{
		client:        r.tc,
		methodName:    serviceMethod,
		serviceMethod: serviceMethod,
		reqHeaders:    reqHeaders,
		contextLogger: r.tc.ContextLogger,
		metrics:       r.metrics,
	}

	if m, ok := r.tc.methodNames[serviceMethod]; ok {
		call.methodName = m
		call.contextLogger = r.tc.ContextLogger
		call.metrics = r.metrics
	}

	return r.tc.call(ctx, call, reqHeaders, req, resp, timeoutAndRetryOptions)
}
