# Zanzibar Thrift file semantics

This document defines the semantics of a thrift file.

# HTTP + JSON

## Structs and types

The structs defined in the thrift file are serialized into
JSON. When parsing, any fields on the wire not defined
in the thrift struct are ignored.

## Annotations

### `zanzibar.http.method`

required. Annotation on thrift method

The HTTP method to use, this is mandatory and can be
"GET", "POST", "DELETE", "UPDATE", "PATCH"

If the method is GET then the function arguments
must have http.ref annotations remapping arguments
to query parameters or params.

### `zanzibar.http.path`

required. Annotation on thrift method

The HTTP path necessary to send a request. This HTTP
path may contain parameter segments.

### `zanzibar.http.status`

required. Annotation on thrift method or exception

The HTTP status code for the thrift function return
value. This can be set on both the thrift function
and the exceptions thrown by a thrift function

### `zanzibar.http.ref`

optional. Annotation on thrift struct field or function argument

Zanzibar allows for customizing how a body struct is
parsed from the HTTP+JSON request on the wire ( and 
how its serialized for the client ).

 - `headers.{{$headerName}}` means that this field is
	not in the JSON body and is read/written to the HTTP 
	headers.
 - `params.{{$paramName}}` means that this field is not
	in the JSON body and is instead read/written to a named 
	parameter in the URL path.
 - `query.{{$queryName}}` means that this field is not 
	in the JSON body and is instead read/written to a query
	parameter in the URL.
 - `body.{{$fieldName}}` means that this field comes from 
	a different field in the body. The fieldName is absolute
	from the root of the body JSON object.

### `zanzibar.http.headerGroups`

optional. Annotation on thrift method

`headerGroups` is a comma seperated list of struct names
which define mandatory headers to apply to this method.

### `zanzibar.http.reqHeaders`

optional. Annotation on thrift method

The list of required headers on the http request.

### `zanzibar.http.resHeaders`

optional. Annotation on thrift method

The list of required headers on the http response.

# TChannel + Thrift

The TChannel + Thrift semantics are thoroughly documented
in https://github.com/uber/tchannel/blob/master/docs/thrift.md
