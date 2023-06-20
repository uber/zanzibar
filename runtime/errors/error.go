package errors

import "fmt"

type ErrorType int

const (
	ClientException ErrorType = iota + 1
	TChannelError
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

func NewZErrorFactory(moduleClass, moduleName string) zErrorFactory {
	return zErrorFactory{
		errLocation: moduleClass + "::" + moduleName,
	}
}

func (factory zErrorFactory) ZError(err error, errType ErrorType) ZError {
	return &zError{
		error:         err,
		errorType:     errType,
		errorLocation: factory.errLocation,
	}
}
