package zanzibar

import "go.uber.org/zap"

const (
	logFieldErrorLocation = "errorLocation"
	logFieldErrorType     = "errorType"
)

// ErrorBuilder provides useful functions to use Error.
type ErrorBuilder interface {
	Error(err error, errType ErrorType) Error
	LogFieldErrorLocation(err error) zap.Field
	LogFieldErrorType(err error) zap.Field
}

// NewErrorBuilder creates an instance of ErrorBuilder.
// Input module id is used as error location for Errors
// created by this builder.
//
// PseudoErrLocation is prefixed with "~" to identify
// logged error that is not created in the present module.
func NewErrorBuilder(moduleClassName, moduleName string) ErrorBuilder {
	return zErrorBuilder{
		errLocation:       moduleClassName + "::" + moduleName,
		pseudoErrLocation: "~" + moduleClassName + "::" + moduleName,
	}
}

type zErrorBuilder struct {
	errLocation, pseudoErrLocation string
}

type zError struct {
	error
	errLocation string
	errType     ErrorType
}

var _ Error = (*zError)(nil)
var _ ErrorBuilder = (*zErrorBuilder)(nil)

func (zb zErrorBuilder) Error(err error, errType ErrorType) Error {
	return zError{
		error:       err,
		errLocation: zb.errLocation,
		errType:     errType,
	}
}

func (zb zErrorBuilder) toError(err error) Error {
	if zerr, ok := err.(Error); ok {
		return zerr
	}
	return zError{
		error:       err,
		errLocation: zb.pseudoErrLocation,
	}
}

func (zb zErrorBuilder) LogFieldErrorLocation(err error) zap.Field {
	zerr := zb.toError(err)
	return zap.String(logFieldErrorLocation, zerr.ErrorLocation())
}

func (zb zErrorBuilder) LogFieldErrorType(err error) zap.Field {
	zerr := zb.toError(err)
	return zap.String(logFieldErrorType, zerr.ErrorType().String())
}

func (e zError) Unwrap() error {
	return e.error
}

func (e zError) ErrorLocation() string {
	return e.errLocation
}

func (e zError) ErrorType() ErrorType {
	return e.errType
}
