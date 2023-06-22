// Copyright (c) 2023 Uber Technologies, Inc.
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

package errors

import (
	"fmt"

	"go.uber.org/zap"
)

type ErrorType int

const (
	// ClientException are client defined exceptions defined in the client thrifts.
	ClientException ErrorType = iota + 1
	// TChannelError are errors of type tchannel.SystemError.
	TChannelError
	// ClientError are errors from client such as undefined exceptions, incomplete response.
	ClientError
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
	case ClientError:
		return "ClientError"
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
