// Copyright (c) 2018 Uber Technologies, Inc.
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

import "fmt"

// RequestError defines the errors for invalid requests.
type RequestError struct {
	FieldName RequestField
	Cause     error
}

// RequestField is a field name in a request.
type RequestField string

// unique camel-cased key naming with module name + field name
// comment annotates jq style field path in json payload from UI
// json tag used by UI follows previous convention
// TODO: fix mixed-case-naming introduced in early commits
const (
	// ClientsType: .client_updates[idx].type
	ClientsType RequestField = "client_type"
	// ClientsServiceName: .client_updates[idx].serviceName
	ClientsServiceName RequestField = "service_name"
	// ClientsIP: .client_updates[idx].ip
	ClientsIP RequestField = "ip"
	// ClientsPort: .client_updates[idx].port
	ClientsPort RequestField = "port"
	// ClientsThriftFile: .client_updates[idx].thriftFile
	ClientsThriftFile RequestField = "ThriftFile"
	// ClientsExposedMethods: .client_updates[idx].exposedMethods
	ClientsExposedMethods RequestField = "exposedMethods"
	// EndpointsThriftFile: .endpoint_updates[idx].thriftFile
	EndpointsThriftFile RequestField = "ThriftFile"
	// EndpointsClientID: .endpoint_updates[idx].ClientId
	EndpointsClientID RequestField = "client_id"
	// EndpointThriftMethodName: .endpoint_updates[idx].ClientMethod
	EndpointThriftMethodName RequestField = "thriftMethodName"
	// Filename: .managed_thrift_files[idx].filename
	Filename RequestField = "filename"
	// ThriftFiles: .thrift_files
	ThriftFiles RequestField = "thrift_files"
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
