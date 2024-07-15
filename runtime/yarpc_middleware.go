package zanzibar

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"go.uber.org/yarpc/api/middleware"
	"go.uber.org/yarpc/api/transport"
)

// NewCaptureOutboundMiddleware captures outbound rpc calls
func NewCaptureOutboundMiddleware() middleware.UnaryOutbound {
	return &captureOutboundMiddleware{}
}

type captureOutboundMiddleware struct {
}

func (m *captureOutboundMiddleware) Call(
	ctx context.Context,
	req *transport.Request,
	next transport.UnaryOutbound,
) (*transport.Response, error) {
	captureEnabled := isCaptureEnabled(ctx, req)
	var event *GRPCOutgoingEvent
	var err error

	if req != nil && captureEnabled {
		event, err = prepareRequest(req)
		if err != nil || event == nil {
			captureEnabled = false
		}
	}

	// call next middleware
	resp, clientErr := next.Call(ctx, req)

	// if capture at this request is still enabled process response and store it in receiveInteraction
	if captureEnabled && resp != nil && event != nil {
		err = prepareResponse(req, resp, event)
		if err != nil {
			return resp, clientErr
		}
		if ec := GetEventContainer(ctx); ec != nil {
			ec.Events = append(ec.Events, event)
		}
	}
	return resp, clientErr
}

func isCaptureEnabled(ctx context.Context, req *transport.Request) bool {
	if GetToCapture(ctx) && req != nil && req.Encoding == "grpc" {
		return true
	}
	return false
}

func prepareRequest(request *transport.Request) (*GRPCOutgoingEvent, error) {
	if request.Body == nil {
		return nil, fmt.Errorf("req.Body is nil for %s::%s", request.Service, request.Procedure)
	}
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	clonedHeaders := cloneMap(request.Headers.OriginalItems())
	return &GRPCOutgoingEvent{
		ServiceName: request.Service,
		MethodName:  request.Procedure,
		ReqHeaders:  clonedHeaders,
		Req:         bodyBytes,
	}, nil
}

func prepareResponse(req *transport.Request, resp *transport.Response, event *GRPCOutgoingEvent) error {
	if resp.Body == nil {
		return fmt.Errorf("resp.Body is nil for %s::%s", req.Service, req.Procedure)
	}
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// close body before swapping reader
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	resp.Body = io.NopCloser(bytes.NewReader(responseBytes))
	event.Rsp = responseBytes
	event.RspHeaders = cloneMap(resp.Headers.Items())
	event.Success = !resp.ApplicationError
	return nil
}

func cloneMap(src map[string]string) map[string]string {
	clone := make(map[string]string)
	for key, val := range src {
		clone[key] = val
	}
	return clone
}
