// IO is captured using Events. This file provides structures and functions to depict the IO, and capture the
// events.

package zanzibar

// Context Variables
const (
	// ToCapture set to true if events have to be captured
	ToCapture = "to_capture"
)

const (
	EventThriftCapture = "event-thrift-capture"
	EventHTTPCapture   = "event-http-capture"
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

// ThriftCaptureEvent captures request and response data
type ThriftCaptureEvent struct {
	MetaData *EventMetaData

	MethodName  string
	ServiceName string

	ReqHeaders map[string]string
	ReqBody    []byte

	RspHeaders map[string]string
	RspBody    []byte
}

func (tce *ThriftCaptureEvent) Name() string {
	return EventThriftCapture
}

// HTTPCaptureEvent captures request and response data
type HTTPCaptureEvent struct {
	MetaData *EventMetaData

	URL string

	ReqHeaders map[string][]string
	ReqBody    []byte

	RspHeaders map[string][]string
	RspBody    []byte
}

func (hce *HTTPCaptureEvent) Name() string {
	return EventHTTPCapture
}

// NoOpEventHandler ignored events
func NoOpEventHandler(events []Event) error {
	return nil
}

// NoOpEventSampler will not sample
func NoOpEventSampler(_, _ string) bool {
	return false
}
