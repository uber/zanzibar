// Copyright (c) 2017 Uber Technologies, Inc.
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

package repository

import "fmt"

// RequestError defines the errors for invalid requests.
type RequestError struct {
	FieldName RequestField
	Cause     error
}

// RequestField is a field name in a request.
type RequestField string

const (
	// ClientID field.
	ClientID RequestField = "client_id"
	// ClientType field.
	ClientType RequestField = "client_type"
)

// NewRequestError returns an error for invalid request.
func NewRequestError(fieldName RequestField, cause error) error {
	return &RequestError{
		FieldName: fieldName,
		Cause:     cause,
	}
}

// Error implements the error interface.
func (r *RequestError) Error() string {
	return fmt.Sprintf("invalid request for field %q: %s", r.FieldName, r.Cause.Error())
}

// JSON serializes the error into JSON.
func (r *RequestError) JSON() string {
	return fmt.Sprintf("{\"field_name\": %q, \"cause\": %q}", r.FieldName, r.Cause.Error())
}
