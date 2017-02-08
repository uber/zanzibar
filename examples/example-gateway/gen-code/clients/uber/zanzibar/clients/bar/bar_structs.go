package barClient

import (
	"github.com/uber/zanzibar/examples/example-gateway/gen-code/uber/zanzibar/clients/bar/bar"
	"github.com/uber/zanzibar/examples/example-gateway/gen-code/uber/zanzibar/clients/foo/foo"
)

// argNotStructHTTPRequest is the http body type for endpoint argNotStruct.
type argNotStructHTTPRequest struct {
	Request string
}

// barHTTPRequest is the http body type for endpoint bar.
type barHTTPRequest struct {
	Request bar.BarRequest
}

// tooManyArgsHTTPRequest is the http body type for endpoint tooManyArgs.
type tooManyArgsHTTPRequest struct {
	Request bar.BarRequest
	Foo     foo.FooStruct
}
