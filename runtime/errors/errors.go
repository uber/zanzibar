package errors

import (
	"context"

	"github.com/uber/tchannel-go"
	"go.uber.org/yarpc/yarpcerrors"
)

const _statusCodeAnnotationKey = "rpc.code"

var (
	// _tchannelCodeToCode maps TChannel SystemErrCodes to their corresponding Code.
	_tchannelCodeToCode = map[tchannel.SystemErrCode]yarpcerrors.Code{
		tchannel.ErrCodeTimeout:    yarpcerrors.CodeDeadlineExceeded,
		tchannel.ErrCodeCancelled:  yarpcerrors.CodeCancelled,
		tchannel.ErrCodeBusy:       yarpcerrors.CodeResourceExhausted,
		tchannel.ErrCodeDeclined:   yarpcerrors.CodeUnavailable,
		tchannel.ErrCodeUnexpected: yarpcerrors.CodeInternal,
		tchannel.ErrCodeBadRequest: yarpcerrors.CodeInvalidArgument,
		tchannel.ErrCodeNetwork:    yarpcerrors.CodeUnavailable,
		tchannel.ErrCodeProtocol:   yarpcerrors.CodeInternal,
	}
)

func IsSystemError(err error) bool {
	if err == nil {
		return false
	}
	_, isSysErr := err.(tchannel.SystemError)
	return isSysErr
}

func SetContextSystemErrorCode(ctx context.Context, err error) context.Context {
	if ctx != nil && err != nil {
		if systemErr, ok := err.(tchannel.SystemError); ok {
			if code, ok := _tchannelCodeToCode[systemErr.Code()]; ok {
				ctx = SetContextStatusCode(ctx, code)
			} else {
				// same as yarpc-go https://github.com/yarpc/yarpc-go/blob/d33ff85d687eb11de3324507ffdc817a39001b3f/transport/tchannel/error.go#L67C39-L67C51
				ctx = SetContextStatusCode(ctx, yarpcerrors.CodeInternal)
			}
		}
	}
	return ctx
}

func SetContextStatusCode(ctx context.Context, code yarpcerrors.Code) context.Context {
	if ctx != nil {
		if statusCode := ctx.Value(_statusCodeAnnotationKey); statusCode == nil {
			ctx = context.WithValue(ctx, _statusCodeAnnotationKey, code)
		}
	}
	return ctx
}
