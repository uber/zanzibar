package errors

import (
	"fmt"

	"go.uber.org/zap"
)

type ErrorType int

const (
	ClientException ErrorType = iota + 1
	TChannelError
)

const (
	logFieldErrorLocation = "errorLocation"
	logFieldErrorType     = "errorType"
)

func (t ErrorType) String() string {
	switch t {
	case ClientException:
		return "ClientException"
	case TChannelError:
		return "TChannelError"
	case 0:
		return ""
	default:
		return fmt.Sprintf("Unknown(ErrorType=%d)", t)
	}
}

type ZError interface {
	error
	ErrorLocation() string
	ErrorType() ErrorType
}

type ZErrorFactory interface {
	ZError(err error, errType ErrorType) ZError
	LogFieldErrorLocation(err error) zap.Field
	LogFieldErrorType(err error) zap.Field
}

type zError struct {
	error
	errorType     ErrorType
	errorLocation string
}

func (z zError) ErrorLocation() string {
	return z.errorLocation
}

func (z zError) ErrorType() ErrorType {
	return z.errorType
}

type zErrorFactory struct {
	errLocation string
}

func NewZErrorFactory(moduleClass, moduleName string) ZErrorFactory {
	return zErrorFactory{
		errLocation: moduleClass + "::" + moduleName,
	}
}

func (factory zErrorFactory) ZError(err error, errType ErrorType) ZError {
	return zError{
		error:         err,
		errorType:     errType,
		errorLocation: factory.errLocation,
	}
}

// toZError casts input to ZError if possible, otherwise creates new ZError
// using "~" prefix (denotes pseudo) to the factory location, since actual error source may not
// be the same.
func (factory zErrorFactory) toZError(err error) ZError {
	if zerr, ok := err.(ZError); ok {
		return zerr
	}
	return zError{
		error:         err,
		errorLocation: "~" + factory.errLocation,
	}
}

func (factory zErrorFactory) LogFieldErrorLocation(err error) zap.Field {
	zerr := factory.toZError(err)
	return zap.String(logFieldErrorLocation, zerr.ErrorLocation())
}

func (factory zErrorFactory) LogFieldErrorType(err error) zap.Field {
	zerr := factory.toZError(err)
	return zap.String(logFieldErrorType, zerr.ErrorType().String())
}
