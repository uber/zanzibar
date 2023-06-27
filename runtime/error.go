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

// ErrorType is used for error grouping.
type ErrorType int

const (
	// TChannelError are errors of type tchannel.SystemError
	TChannelError ErrorType = iota + 1

	// ClientException are client exceptions defined in the
	// client IDL.
	ClientException

	// BadResponse are errors reading client response such
	// as undefined exceptions, empty response.
	BadResponse
)

//go:generate stringer -type=ErrorType

// Error is a wrapper on go error to provide meta fields about
// error that are logged to improve error debugging, as well
// as to facilitate error grouping and filtering in logs.
type Error interface {
	error

	// ErrorLocation is the module identifier that produced
	// the error. It should be one of client, middleware, endpoint.
	ErrorLocation() string

	// ErrorType is used for error grouping.
	ErrorType() ErrorType

	// Unwrap returns wrapped error.
	Unwrap() error
}
