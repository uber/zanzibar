// Copyright (c) 2024 Uber Technologies, Inc.
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

// Context Variables
const (
	// ToCapture set to true if events have to be captured
	ToCapture = "to_capture"
)

const (
	EventThriftOutgoing = "event-thrift-outgoing"
	EventGRPCOutgoing   = "event-grpc-outgoing"
	EventHTTPIncoming   = "event-http-incoming"
	EventHTTPOutgoing   = "event-http-outgoing"
)

type EventHandlerFn func([]Event) error
type EnableEventGenFn func(string, string) bool

type Event interface {
	Name() string
}

// EventContainer holds generated events.
type EventContainer struct {
	Events []Event
}

// ThriftOutgoingEvent captures request and response data
type ThriftOutgoingEvent struct {
	MethodName  string
	ServiceName string

	ReqHeaders map[string]string
	Req        RWTStruct

	RspHeaders map[string]string
	Rsp        RWTStruct

	Success bool
}

func (tce *ThriftOutgoingEvent) Name() string {
	return EventThriftOutgoing
}

// GRPCOutgoingEvent captures request and response data
type GRPCOutgoingEvent struct {
	MethodName  string
	ServiceName string

	ReqHeaders map[string]string
	Req        []byte

	RspHeaders map[string]string
	Rsp        []byte

	Success bool
}

func (gce *GRPCOutgoingEvent) Name() string {
	return EventGRPCOutgoing
}

// HTTPIncomingEvent captures incoming request and response data received
type HTTPIncomingEvent struct {
	EndpointName string // optional
	HandlerName  string // optional

	HTTPCapture
}

func (hce *HTTPIncomingEvent) Name() string {
	return EventHTTPIncoming
}

// HTTPOutgoingEvent captures incoming request and response data received
type HTTPOutgoingEvent struct {
	ClientID       string // optional
	ClientEndpoint string // optional

	HTTPCapture
}

func (hce *HTTPOutgoingEvent) Name() string {
	return EventHTTPOutgoing
}

// HTTPCapture captures request and response data
type HTTPCapture struct {
	ReqURL     string
	ReqMethod  string
	ReqHeaders map[string][]string
	ReqBody    []byte

	RspStatusCode int
	RspMethod     string
	RspHeaders    map[string][]string
	RspBody       []byte
}

// NoOpEventHandler ignored events
func NoOpEventHandler(events []Event) error {
	return nil
}

// NoOpEventGen will not sample
func NoOpEventGen(_, _ string) bool {
	return false
}