// IO is captured using Events. This file provides structures and functions to depict the IO, and capture the
// events.

package zanzibar

// Context Variables
const (
	// ToCapture set to true if events have to be captured
	ToCapture = "to_capture"
)

const (
	EventThriftOutgoing = "event-thrift-outgoing"
	EventHTTPIncoming   = "event-http-incoming"
	EventHTTPOutgoing   = "event-http-outgoing"
)

type EventHandlerFn func([]Event) error
type EventSamplerFn func(string, string) bool

type Event interface {
	Name() string
}

// EventContainer holds generated events.
type EventContainer struct {
	events []Event
}

// EventMetaData is information required to aid with storage and linking of different requests
type EventMetaData struct {
	// filled for http requests received by the gateway
	EndpointName string // optional
	HandlerName  string // optional
}

// ThriftOutgoingEvent captures request and response data
type ThriftOutgoingEvent struct {
	MethodName  string
	ServiceName string

	ReqHeaders map[string]string
	ReqBody    []byte

	RspHeaders map[string]string
	RspBody    []byte
}

func (tce *ThriftOutgoingEvent) Name() string {
	return EventThriftOutgoing
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

// NoOpEventSampler will not sample
func NoOpEventSampler(_, _ string) bool {
	return false
}
