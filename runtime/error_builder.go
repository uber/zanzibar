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

package zanzibar

// ErrorBuilder wraps input error into Error.
type ErrorBuilder interface {
	Error(err error, errType ErrorType) error

	Rebuild(zErr Error, err error) error
}

// NewErrorBuilder creates an instance of ErrorBuilder.
// Input module id is used as error location for Errors
// created by this builder.
func NewErrorBuilder(moduleClassName, moduleName string) ErrorBuilder {
	return errorBuilder{
		errLocation: moduleClassName + "::" + moduleName,
	}
}

type errorBuilder struct {
	errLocation string
}

type wrappedError struct {
	error
	errLocation string
	errType     ErrorType
}

var _ Error = (*wrappedError)(nil)
var _ ErrorBuilder = (*errorBuilder)(nil)

func (eb errorBuilder) Error(err error, errType ErrorType) error {
	return wrappedError{
		error:       err,
		errLocation: eb.errLocation,
		errType:     errType,
	}
}

func (eb errorBuilder) Rebuild(zErr Error, err error) error {
	return wrappedError{
		error:       err,
		errLocation: zErr.ErrorLocation(),
		errType:     zErr.ErrorType(),
	}
}

func (e wrappedError) Unwrap() error {
	return e.error
}

func (e wrappedError) ErrorLocation() string {
	return e.errLocation
}

func (e wrappedError) ErrorType() ErrorType {
	return e.errType
}
