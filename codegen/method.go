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

package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

// PathSegment represents a part of the http path.
type PathSegment struct {
	Type           string
	Text           string
	BodyIdentifier string
}

// ExceptionSpec contains information about thrift exceptions
type ExceptionSpec struct {
	StructSpec

	StatusCode StatusCode
}

// HeaderFieldInfo contains information about where to store
// the string from headers into the request/response body.
type HeaderFieldInfo struct {
	FieldIdentifier string
	IsPointer       bool
}

// MethodSpec specifies all needed parts to generate code for a method in service.
type MethodSpec struct {
	Name       string
	HTTPMethod string
	// Used by edge gateway to generate endpoint.
	EndpointName string
	HTTPPath     string
	PathSegments []PathSegment
	IsEndpoint   bool

	// Statements for reading query parameters.
	QueryParamGoStatements []string

	// ReqHeaderFields is a map of "header name" to
	// a golang field accessor expression like ".Foo.Bar"
	// Use to place request headers in the body
	ReqHeaderFields map[string]HeaderFieldInfo

	// ResHeaderFields is a map of header name to a golang
	// field accessor expression used to read fields out
	// of the response body and place them into response headers
	ResHeaderFields map[string]HeaderFieldInfo

	// ReqHeaders needed, generated from "zanzibar.http.reqHeaders"
	ReqHeaders []string
	// ResHeaders needed, generated from "zanzibar.http.resHeaders"
	ResHeaders []string

	RequestType      string
	ResponseType     string
	OKStatusCode     StatusCode
	Exceptions       []ExceptionSpec
	ExceptionsIndex  map[string]ExceptionSpec
	ValidStatusCodes []int
	// Additional struct generated from the bundle of request args.
	RequestBoxed bool
	// Thrift service name the method belongs to.
	ThriftService string
	// The thriftrw-generated go package name
	GenCodePkgName string
	// Whether the method needs annotation or not.
	WantAnnot bool
	// The thriftrw compiled spec, used to extract type information
	CompiledThriftSpec *compile.FunctionSpec
	// The downstream service method set by endpoint config
	Downstream *ModuleSpec
	// the downstream service name
	DownstreamService string
	// The downstream method spec for the endpoint
	DownstreamMethod *MethodSpec

	// Statements for converting request types
	ConvertRequestGoStatements []string

	// Statements for converting response types
	ConvertResponseGoStatements []string
}

// StructSpec specifies a Go struct to be generated.
type StructSpec struct {
	Type        string
	Name        string
	Annotations map[string]string
}

// StatusCode is for http status code with exception message.
type StatusCode struct {
	Code    int
	Message string
}

const (
	antHTTPMethod      = "zanzibar.http.method"
	antHTTPPath        = "zanzibar.http.path"
	antHTTPStatus      = "zanzibar.http.status"
	antHTTPReqDefBoxed = "zanzibar.http.req.def"
	antHTTPReqHeaders  = "zanzibar.http.reqHeaders"
	antHTTPResHeaders  = "zanzibar.http.resHeaders"
	antHTTPRef         = "zanzibar.http.ref"
	antMeta            = "zanzibar.meta"
	antHandler         = "zanzibar.handler"
)

// NewMethod creates new method specification.
func NewMethod(
	thriftFile string,
	funcSpec *compile.FunctionSpec,
	packageHelper *PackageHelper,
	wantAnnot bool,
	isEndpoint bool,
	thriftService string,
) (*MethodSpec, error) {
	method := &MethodSpec{}
	method.CompiledThriftSpec = funcSpec
	var err error
	var ok bool
	method.Name = funcSpec.MethodName()
	method.IsEndpoint = isEndpoint
	method.WantAnnot = wantAnnot
	method.ThriftService = thriftService

	method.GenCodePkgName, err = packageHelper.TypePackageName(thriftFile)
	if err != nil {
		return nil, err
	}

	err = method.setResponseType(thriftFile, funcSpec.ResultSpec, packageHelper)
	if err != nil {
		return nil, err
	}

	err = method.setRequestType(thriftFile, funcSpec, packageHelper)
	if err != nil {
		return nil, err
	}

	err = method.setExceptions(thriftFile, isEndpoint, funcSpec.ResultSpec, packageHelper)
	if err != nil {
		return nil, err
	}

	method.ReqHeaders = headers(funcSpec.Annotations[antHTTPReqHeaders])
	method.ResHeaders = headers(funcSpec.Annotations[antHTTPResHeaders])

	if !wantAnnot {
		return method, nil
	}

	if method.HTTPMethod, ok = funcSpec.Annotations[antHTTPMethod]; !ok {
		return nil, errors.Errorf("missing anotation '%s' for HTTP method", antHTTPMethod)
	}

	method.EndpointName = funcSpec.Annotations[antHandler]

	err = method.setOKStatusCode(funcSpec.Annotations[antHTTPStatus])
	if err != nil {
		return nil, err
	}

	method.setValidStatusCodes()

	if method.HTTPMethod == "GET" && method.RequestType != "" {
		if !method.IsEndpoint {
			return nil, errors.Errorf(
				"invalid annotation: HTTP GET method cannot have a body",
			)
		}

		err := method.setQueryParamStatements(funcSpec)
		if err != nil {
			return nil, err
		}
	}

	var httpPath string
	if httpPath, ok = funcSpec.Annotations[antHTTPPath]; !ok {
		return nil, errors.Errorf(
			"missing anotation '%s' for HTTP path", antHTTPPath,
		)
	}
	method.setHTTPPath(httpPath, funcSpec)

	method.setRequestHeaderFields(funcSpec)
	method.setResponseHeaderFields(funcSpec)

	return method, nil
}

// setRequestType sets the request type of the method specification. If the
// "zanzibar.http.req.def.boxed" is true, then the first parameter will be used as
// the request body; otherwise a new struct is generated to bundle the request
// parameters as http body and the name of the struct will be returned.
func (ms *MethodSpec) setRequestType(curThriftFile string, funcSpec *compile.FunctionSpec, packageHelper *PackageHelper) error {
	if len(funcSpec.ArgsSpec) == 0 {
		ms.RequestType = ""
		return nil
	}
	var err error
	if isRequestBoxed(funcSpec) {
		ms.RequestBoxed = true
		ms.RequestType, err = packageHelper.TypeFullName(funcSpec.ArgsSpec[0].Type)
		if err == nil && isStructType(funcSpec.ArgsSpec[0].Type) {
			ms.RequestType = "*" + ms.RequestType
		}
	} else {
		ms.RequestBoxed = false

		goPackageName, err := packageHelper.TypePackageName(curThriftFile)
		if err == nil {
			ms.RequestType = "*" + goPackageName + "." +
				ms.ThriftService + "_" + strings.Title(ms.Name) + "_Args"
		}
	}
	if err != nil {
		return errors.Wrap(err, "failed to set request type")
	}
	return nil
}

func (ms *MethodSpec) setResponseType(curThriftFile string, respSpec *compile.ResultSpec, packageHelper *PackageHelper) error {
	if respSpec == nil {
		ms.ResponseType = ""
		return nil
	}
	typeName, err := packageHelper.TypeFullName(respSpec.ReturnType)
	if isStructType(respSpec.ReturnType) {
		typeName = "*" + typeName
	}
	if err != nil {
		return errors.Wrap(err, "failed to get response type")
	}
	ms.ResponseType = typeName
	return nil
}

// RefResponse prepends the response variable with '&' if it is not of reference type
// It is used to construct the `Success` field of the `$service_$method_Result` struct
// generated by thriftrw, which is always of reference type.
func (ms *MethodSpec) RefResponse(respVar string) string {
	respSpec := ms.CompiledThriftSpec.ResultSpec
	if respSpec == nil || respSpec.ReturnType == nil {
		return respVar
	}

	switch compile.RootTypeSpec(respSpec.ReturnType).ThriftName() {
	case "bool", "byte", "i16", "i32", "i64", "double", "string":
		return "&" + respVar
	default:
		return respVar
	}
}

func (ms *MethodSpec) setOKStatusCode(statusCode string) error {
	if statusCode == "" {
		return errors.Errorf("no http OK status code set by annotation '%s' ", antHTTPStatus)
	}

	code, err := strconv.Atoi(statusCode)
	if err != nil {
		return errors.Wrapf(err,
			"Could not parse status code annotation (%s) for ok response",
			statusCode,
		)
	}
	ms.OKStatusCode = StatusCode{
		Code: code,
	}

	return nil
}

func (ms *MethodSpec) setValidStatusCodes() {
	ms.ValidStatusCodes = make([]int, len(ms.Exceptions)+1)

	ms.ValidStatusCodes[0] = ms.OKStatusCode.Code
	for i := 0; i < len(ms.Exceptions); i++ {
		ms.ValidStatusCodes[i+1] = ms.Exceptions[i].StatusCode.Code
	}
}

func (ms *MethodSpec) setExceptions(
	curThriftFile string,
	isEndpoint bool,
	resultSpec *compile.ResultSpec,
	h *PackageHelper,
) error {
	seenStatusCodes := map[int]bool{
		ms.OKStatusCode.Code: true,
	}
	ms.Exceptions = make([]ExceptionSpec, len(resultSpec.Exceptions))
	ms.ExceptionsIndex = make(
		map[string]ExceptionSpec, len(resultSpec.Exceptions),
	)

	for i, e := range resultSpec.Exceptions {
		typeName, err := h.TypeFullName(e.Type)
		if err != nil {
			return errors.Wrapf(
				err,
				"cannot resolve type full name for %s for exception %s",
				e.Type,
				e.Name,
			)
		}

		if !ms.WantAnnot {
			exception := ExceptionSpec{
				StructSpec: StructSpec{
					Type: typeName,
					Name: e.Name,
				},
			}
			ms.Exceptions[i] = exception
			ms.ExceptionsIndex[e.Name] = exception
			continue
		}

		code, err := strconv.Atoi(e.Annotations[antHTTPStatus])
		if err != nil {
			return errors.Wrapf(
				err,
				"cannot parse the annotation %s for exception %s", antHTTPStatus, e.Name,
			)
		}

		if seenStatusCodes[code] && !isEndpoint {
			return errors.Wrapf(
				err,
				"cannot have duplicate status code %s for exception %s",
				antHTTPStatus,
				e.Name,
			)
		}
		seenStatusCodes[code] = true

		exception := ExceptionSpec{
			StructSpec: StructSpec{
				Type:        typeName,
				Name:        e.Name,
				Annotations: e.Annotations,
			},
			StatusCode: StatusCode{
				Code:    code,
				Message: e.Name,
			},
		}
		ms.Exceptions[i] = exception
		ms.ExceptionsIndex[e.Name] = exception
	}
	return nil
}

func findParamsAnnotation(
	fields compile.FieldGroup, paramName string,
) (string, bool) {
	var identifier string
	visitor := func(prefix string, field *compile.FieldSpec) bool {
		if param, ok := field.Annotations[antHTTPRef]; ok {
			if param == "params."+paramName[1:] {
				identifier = prefix + "." + strings.Title(field.Name)
				return true
			}
		}

		return false
	}
	walkFieldGroups(fields, visitor)

	if identifier == "" {
		return "", false
	}

	return identifier, true
}

func (ms *MethodSpec) setRequestHeaderFields(
	funcSpec *compile.FunctionSpec,
) {
	fields := compile.FieldGroup(funcSpec.ArgsSpec)
	ms.ReqHeaderFields = map[string]HeaderFieldInfo{}

	// Scan for all annotations
	visitor := func(prefix string, field *compile.FieldSpec) bool {
		if param, ok := field.Annotations[antHTTPRef]; ok {
			if param[0:8] == "headers." {
				headerName := param[8:]
				ms.ReqHeaderFields[headerName] = HeaderFieldInfo{
					FieldIdentifier: prefix + "." + strings.Title(field.Name),
					IsPointer:       !field.Required,
				}
			}
		}
		return false
	}
	walkFieldGroups(fields, visitor)
}

func (ms *MethodSpec) setResponseHeaderFields(
	funcSpec *compile.FunctionSpec,
) {
	structType, ok := funcSpec.ResultSpec.ReturnType.(*compile.StructSpec)
	// If the result is not a struct then there are zero response header
	// annotations.
	if !ok {
		return
	}

	fields := structType.Fields
	ms.ResHeaderFields = map[string]HeaderFieldInfo{}

	// Scan for all annotations
	visitor := func(prefix string, field *compile.FieldSpec) bool {
		if param, ok := field.Annotations[antHTTPRef]; ok {
			if param[0:8] == "headers." {
				headerName := param[8:]
				ms.ResHeaderFields[headerName] = HeaderFieldInfo{
					FieldIdentifier: prefix + "." + strings.Title(field.Name),
					IsPointer:       !field.Required,
				}
			}
		}
		return false
	}
	walkFieldGroups(fields, visitor)
}

func (ms *MethodSpec) setHTTPPath(httpPath string, funcSpec *compile.FunctionSpec) {
	ms.HTTPPath = httpPath

	segments := strings.Split(httpPath[1:], "/")
	ms.PathSegments = make([]PathSegment, len(segments))
	for i := 0; i < len(segments); i++ {
		segment := segments[i]

		if segment == "" || segment[0] != ':' {
			ms.PathSegments[i].Type = "static"
			ms.PathSegments[i].Text = segment
		} else {
			ms.PathSegments[i].Type = "param"

			var fieldSelect string
			var ok bool
			if ms.RequestBoxed {
				// Boxed requests mean first arg is struct
				structType := funcSpec.ArgsSpec[0].Type.(*compile.StructSpec)
				fieldSelect, ok = findParamsAnnotation(
					structType.Fields, segment,
				)
			} else {
				fieldSelect, ok = findParamsAnnotation(
					compile.FieldGroup(funcSpec.ArgsSpec), segment,
				)
			}

			if !ok {
				panic("cannot find params: " + segment)
			}
			ms.PathSegments[i].BodyIdentifier = fieldSelect
		}
	}
}

func (ms *MethodSpec) setDownstream(
	clientModule *ModuleSpec, clientThriftService, clientThriftMethod string,
) error {
	var downstreamService *ServiceSpec
	for _, service := range clientModule.Services {
		if service.Name == clientThriftService {
			downstreamService = service
			break
		}
	}
	if downstreamService == nil {
		return errors.Errorf(
			"Downstream service '%s' is not found in '%s'",
			clientThriftService, clientModule.ThriftFile,
		)
	}
	var downstreamMethod *MethodSpec
	for _, method := range downstreamService.Methods {
		if method.Name == clientThriftMethod {
			downstreamMethod = method
			break
		}
	}
	if downstreamMethod == nil {
		return errors.Errorf(
			"\n Downstream method '%s' is not found in '%s'",
			clientThriftMethod, clientModule.ThriftFile,
		)
	}
	// Remove irrelevant services and methods.
	ms.Downstream = clientModule
	ms.DownstreamService = clientThriftService
	ms.DownstreamMethod = downstreamMethod
	return nil
}

func (ms *MethodSpec) setTypeConverters(
	funcSpec *compile.FunctionSpec,
	downstreamSpec *compile.FunctionSpec,
	h *PackageHelper,
) error {
	// TODO(sindelar): Iterate over fields that are structs (for foo/bar examples).

	// Add type checking and conversion, custom mapping
	structType := compile.FieldGroup(funcSpec.ArgsSpec)
	downstreamStructType := compile.FieldGroup(downstreamSpec.ArgsSpec)

	typeConverter := &TypeConverter{
		LineBuilder: LineBuilder{},
		Helper:      h,
	}

	err := typeConverter.GenStructConverter(structType, downstreamStructType)
	if err != nil {
		return err
	}

	ms.ConvertRequestGoStatements = typeConverter.GetLines()

	// TODO: support non-struct return types
	respType := funcSpec.ResultSpec.ReturnType
	downstreamRespType := funcSpec.ResultSpec.ReturnType

	if respType == nil || downstreamRespType == nil {
		return nil
	}

	respFields := respType.(*compile.StructSpec).Fields
	downstreamRespFields := downstreamRespType.(*compile.StructSpec).Fields

	respConverter := &TypeConverter{
		LineBuilder: LineBuilder{},
		Helper:      h,
	}

	err = respConverter.GenStructConverter(downstreamRespFields, respFields)
	if err != nil {
		return err
	}

	ms.ConvertResponseGoStatements = respConverter.GetLines()

	return nil
}

func pointerMethodType(typeSpec compile.TypeSpec) string {
	var pointerMethod string

	switch typeSpec.(type) {
	case *compile.BoolSpec:
		pointerMethod = "Bool"
	case *compile.I8Spec:
		pointerMethod = "Int8"
	case *compile.I16Spec:
		pointerMethod = "Int16"
	case *compile.I32Spec:
		pointerMethod = "Int32"
	case *compile.I64Spec:
		pointerMethod = "Int64"
	case *compile.DoubleSpec:
		pointerMethod = "Float64"
	case *compile.StringSpec:
		pointerMethod = "String"
	default:
		panic(fmt.Sprintf("Unknown type (%T) %v", typeSpec, typeSpec))
	}

	return pointerMethod
}

func getQueryMethodForType(typeSpec compile.TypeSpec) string {
	var queryMethod string

	switch typeSpec.(type) {
	case *compile.BoolSpec:
		queryMethod = "GetQueryBool"
	case *compile.I8Spec:
		queryMethod = "GetQueryInt8"
	case *compile.I16Spec:
		queryMethod = "GetQueryInt16"
	case *compile.I32Spec:
		queryMethod = "GetQueryInt32"
	case *compile.I64Spec:
		queryMethod = "GetQueryInt64"
	case *compile.DoubleSpec:
		queryMethod = "GetQueryFloat64"
	case *compile.StringSpec:
		queryMethod = "GetQueryValue"
	default:
		panic(fmt.Sprintf("Unknown type (%T) %v", typeSpec, typeSpec))
	}

	return queryMethod
}

func (ms *MethodSpec) setQueryParamStatements(
	funcSpec *compile.FunctionSpec,
) error {
	// If a thrift field has a http.ref annotation then we
	// should not read this field from query parameters.
	statements := LineBuilder{}
	structType := compile.FieldGroup(funcSpec.ArgsSpec)

	for _, field := range structType {
		realType := compile.RootTypeSpec(field.Type)
		fieldName := field.Name
		identifierName := camelCase(fieldName) + "Query"

		httpRefAnnotation := field.Annotations[antHTTPRef]
		if httpRefAnnotation != "" {
			continue
		}

		okIdentifierName := camelCase(fieldName) + "Ok"
		if field.Required {
			statements.appendf("%s := req.CheckQueryValue(%q)",
				okIdentifierName, fieldName,
			)
			statements.appendf("if !%s {", okIdentifierName)
			statements.append("\treturn")
			statements.append("}")
		} else {
			statements.appendf("%s := req.HasQueryValue(%q)",
				okIdentifierName, fieldName,
			)
			statements.appendf("if %s {", okIdentifierName)
		}

		pointerMethod := pointerMethodType(realType)
		queryMethodName := getQueryMethodForType(realType)

		statements.appendf("%s, ok := req.%s(%q)",
			identifierName, queryMethodName, fieldName,
		)

		statements.append("if !ok {")
		statements.append("\treturn")
		statements.append("}")

		if field.Required {
			statements.appendf("requestBody.%s = %s",
				strings.Title(field.Name), identifierName,
			)
		} else {
			statements.appendf("\trequestBody.%s = ptr.%s(%s)",
				strings.Title(field.Name), pointerMethod, identifierName,
			)
			statements.append("}")
		}

		// new line after block.
		statements.append("")
	}

	ms.QueryParamGoStatements = statements.GetLines()

	return nil
}

func isRequestBoxed(f *compile.FunctionSpec) bool {
	boxed, ok := f.Annotations[antHTTPReqDefBoxed]
	if ok && boxed == "true" {
		return true
	}
	return false
}

func headers(annotation string) []string {
	if annotation == "" {
		return nil
	}
	return strings.Split(annotation, ",")
}

func walkFieldGroups(
	fields compile.FieldGroup,
	visitField func(string, *compile.FieldSpec) bool,
) bool {
	return walkFieldGroupsInternal("", fields, visitField)
}

func walkFieldGroupsInternal(
	prefix string, fields compile.FieldGroup,
	visitField func(string, *compile.FieldSpec) bool,
) bool {
	for i := 0; i < len(fields); i++ {
		field := fields[i]

		bail := visitField(prefix, field)
		if bail {
			return true
		}

		realType := compile.RootTypeSpec(field.Type)
		switch t := realType.(type) {
		case *compile.BinarySpec:
		case *compile.StringSpec:
		case *compile.BoolSpec:
		case *compile.DoubleSpec:
		case *compile.I8Spec:
		case *compile.I16Spec:
		case *compile.I32Spec:
		case *compile.I64Spec:
		case *compile.EnumSpec:
		case *compile.StructSpec:
			bail := walkFieldGroupsInternal(
				prefix+"."+strings.Title(field.Name),
				t.Fields,
				visitField,
			)
			if bail {
				return true
			}
		case *compile.SetSpec:
			// TODO: implement
		case *compile.MapSpec:
			// TODO: implement
		case *compile.ListSpec:
			// TODO: implement
		default:
			panic("unknown Spec")
		}
	}

	return false
}
