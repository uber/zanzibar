package zanzibar

import (
	"go.uber.org/thriftrw/wire"
)

const (
	ToCapture = "to-capture"
)

type ThriftCapture struct {
	ID string

	MethodName  string
	ServiceName string

	ReqHeaders map[string]string
	ReqBody    wire.Value

	RspHeaders map[string]string
	RspBody    wire.Value
}

type HTTPCapture struct {
	ID string

	URL string

	ReqHeaders map[string][]string
	ReqBody    []byte

	RspHeaders map[string][]string
	RspBody    []byte
}

//type CaptureHTTPRequest struct {
//}
//
//type CaptureHTTP func(payload any, request ServerHTTPRequest, response ServerHTTPResponse)
//type CaptureThrift func(payload any)
//
//// DefaultCaptureHTTP is a dummy function for capturing http request and response
//func DefaultCaptureHTTP(payload any, request ServerHTTPRequest, response ServerHTTPResponse) {
//	fmt.Printf("Capturing: %s", request.URL)
//}
//
//// DefaultCaptureThrift is a dummy function for capturing thrift request and response
//func DefaultCaptureThrift(payload any) {
//	fmt.Printf("Capturing Thrift")
//}
//
//func capturePayload(ctx context.Context) any {
//	return ctx.Value(ToCapture)
//}
