# Zanzibar Thrift file semantics

This document defines the semantics of a thrift file.

# HTTP + JSON

## Structs and types

The structs defined in the thrift file are serialized into
JSON. When parsing, any fields on the wire not defined
in the thrift struct are ignored.

Each go struct has a Pascal name field based on either
the thrift field name or the thrift method argument name.
On the wire, in JSON format, all JSON objects have the fieldName
that matches 1:1 to the thrift field name / thrift argument name.

This means each Go struct has a json annotation to specify what
the go struct field serializes into/out of in the JSON wire format.

## JSON semantics

Thrift contains structs with nested types. Each thrift method
has zero or more arguments, these arguments can be optional
or required.

Each thrift method effectively takes a single struct with
zero or more fields on it.

For the JSON request body the arguments of a thrift method must be
represented as a JSON object on the wire, the argument names
are the field names in the JSON object and the argument values
are the field types in the JSON object on the wire.

For the JSON response body, the body is the return type of the
thrift method. In theory this can be literally a boolean or a number
but it's strongly recommended that all thrift methods return objects
so that you can add extra optional fields in the future.

JSON representations:

### `bool`

A thrift bool is a JSON boolean

### `byte`

A thrift byte is a JSON number

### `i16`, `i32`

A thrift `i16`, `i32` is a JSON number

### `i64`

A thrift `i64` is dependent on the [`js.type` thrift annotation](https://github.com/thriftrw/thriftrw-node#i64)

Without annotations its serialized as an array of 8 numbers.

 - Buffer -> 8 byte numbers in an array, e.g. [0, 255, 1, 2, 3, 4, 5, 6]
 - Date -> ISO Date string, e.g. 2016-05-23T22:03:11.618Z
 - Long -> object with low, high and unsigned fields,
	e.g. { low: -1, high: 2147483647, unsigned: false }

### `double`

A thrift `double` is a number on the wire.

### `binary`

A thrift `binary` is an array of numbers, one number for each byte (0-255).

### `string`

A thrift `string` is a JSON string

### `struct`

A thrift struct consists of zero or more fields. 
A thrift struct is a JSON object on the wire with N fields
based on the field names in the struct.

If a field is `optional` then the fieldName can be either
a JSON `null`, a missing field, or the value of the type.

If a field is `required` then the fieldName on the JSON
object MUST exist and must be the correct type.

### `list<t1>`

A thrift `list<t1>` is a JSON array containing only the type t1.

### `set<t1>`

A thrift `set<t1>` is a JSON array containing only the type t1.

### `map<t1,t2>`

A thrift `map<t1, t2>` is a JSON object. `t1` must be `string`.
Each key in the JSON object has a value that must be only `t2`.

### `enum`

A thrift `enum` is a JSON string. The string value must be one
of the enum names defined in the thrift `enum` declaration.


## Annotations

### `zanzibar.http.method`

required. Annotation on thrift method

The HTTP method to use, this is mandatory and can be
"GET", "POST", "DELETE", "UPDATE", "PATCH"

If the method is GET then the function arguments
must have zanzibar.http.ref annotations remapping arguments
to query parameters, params or headers.

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
	
	For a method that is annotated by `zanzibar.http.method = "GET"`,
	its arguments are by default implicitly read/written into query
	parameters in the URL. Also, the query annotation currently only
	works for methods with `zanzibar.http.method = "GET"` annotation.
	It is useful to rename the field name in the URL query, e.g.,
	`1. required string latitude (zanzibar.http.ref = "query.lat")`.
    
	If the annotation is on a field of a struct and that struct is
	a method argument, the URL query name will be prefixed with the
	struct's field name plus ".".

	Following types of method params are supported in query params:

	- bool
	- i8
	- i16
	- i32
	- i64
	- double
	- string
	- list of bool, i8, i16, i32, i64, double or string
	- struct with fields of bool, i8, i16, i32, i64, double, string, or list of bool, i8, i16, i32, i64, double or string

    If the annotation is on a field of type list, the URL query name/value pair 
    can be repeated several times with '&' to send multiple values for a 
    list type.

 	If the annotation is on a field of a struct and that struct is
	a method argument, the URL query name will be prefixed with the
	struct's field name plus ".".

	For types beyond the above supported ones, http `POST` should be
	used, and the method params should go into the request body instead
	of url params.

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

### `zanzibar.validation.type`

optional. 

This annotation allows the JSON parser to be lax and 
parse multiple types into a single thrift field.

For example :

 - `optional double a (zanzibar.validation.type = "string,number")`
 - `optional string a (zanzibar.validation.type = "string,number")`
 - `optional i32 a (zanzibar.validation.type = "string,number")`
 - `optional bool a (zanzibar.validation.type = "boolean,number")`
 - `optional i32 a (zanzibar.validation.type = "boolean,number")`
 - `optional bool a (zanzibar.validation.type = "string,boolean")`

 The coercions available are :

 - `double` may be parsed from a string containing a number
 - `string` may be parsed from a number into a string
 - `i32` may be parsed from a string containing an integer
 - `bool` may be parsed from a number, 0 is false, positive is true
 - `i32` may be parsed from a boolean, false is 0, true is 1
 - `bool` may be parsed from a string `"false"` is false, `"true"` is true

###

# TChannel + Thrift

The TChannel + Thrift semantics are thoroughly documented
in https://github.com/uber/tchannel/blob/master/docs/thrift.md
