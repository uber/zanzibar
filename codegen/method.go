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
	ParamName      string
	Required       bool
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
	ParseQueryParamGoStatements []string

	// Statements for writing query parameters
	WriteQueryParamGoStatements []string

	// Statements for reading request headers
	ReqHeaderGoStatements []string

	// Statements for reading request headers for clients
	ReqClientHeaderGoStatements []string

	// ResHeaderFields is a map of header name to a golang
	// field accessor expression used to read fields out
	// of the response body and place them into response headers
	ResHeaderFields map[string]HeaderFieldInfo

	// ReqHeaders needed, generated from "zanzibar.http.reqHeaders"
	ReqHeaders []string
	// ResHeaders needed, generated from "zanzibar.http.resHeaders"
	ResHeaders []string

	RequestType       string
	ShortRequestType  string
	ResponseType      string
	ShortResponseType string
	OKStatusCode      StatusCode
	Exceptions        []ExceptionSpec
	ExceptionsIndex   map[string]ExceptionSpec
	ValidStatusCodes  []int
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

	// Statements for reading data out of url params (server)
	RequestParamGoStatements []string
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
	antHTTPMethod     = "zanzibar.http.method"
	antHTTPPath       = "zanzibar.http.path"
	antHTTPStatus     = "zanzibar.http.status"
	antHTTPReqHeaders = "zanzibar.http.reqHeaders"
	antHTTPResHeaders = "zanzibar.http.resHeaders"
	antHTTPRef        = "zanzibar.http.ref"
	antMeta           = "zanzibar.meta"
	antHandler        = "zanzibar.handler"

	// AntHTTPReqDefBoxed annotates a method so that the genereted method takes
	// generated argument directly instead of a struct that warps the argument.
	// The annotated method should have one and only one argument.
	AntHTTPReqDefBoxed = "zanzibar.http.req.def"
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
		return nil, errors.Errorf("missing annotation '%s' for HTTP method", antHTTPMethod)
	}

	method.EndpointName = funcSpec.Annotations[antHandler]

	err = method.setOKStatusCode(funcSpec.Annotations[antHTTPStatus])
	if err != nil {
		return nil, err
	}

	method.setValidStatusCodes()

	if method.HTTPMethod == "GET" && method.RequestType != "" {
		if method.IsEndpoint {
			err := method.setParseQueryParamStatements(funcSpec, packageHelper)
			if err != nil {
				return nil, err
			}
		} else {
			err := method.setWriteQueryParamStatements(funcSpec, packageHelper)
			if err != nil {
				return nil, err
			}
		}
	}

	var httpPath string
	if httpPath, ok = funcSpec.Annotations[antHTTPPath]; !ok {
		return nil, errors.Errorf(
			"missing annotation '%s' for HTTP path", antHTTPPath,
		)
	}
	method.setHTTPPath(httpPath, funcSpec)

	err = method.setRequestParamFields(funcSpec, packageHelper)
	if err != nil {
		return nil, err
	}

	err = method.setEndpointRequestHeaderFields(funcSpec, packageHelper)
	if err != nil {
		return nil, err
	}
	err = method.setClientRequestHeaderFields(funcSpec, packageHelper)
	if err != nil {
		return nil, err
	}
	method.setResponseHeaderFields(funcSpec)

	return method, nil
}

func scanForNonParams(funcSpec *compile.FunctionSpec) bool {
	hasNonParams := false

	visitor := func(
		goPrefix string, thriftPrefix string, field *compile.FieldSpec,
	) bool {
		realType := compile.RootTypeSpec(field.Type)
		// ignore nested structs
		if _, ok := realType.(*compile.StructSpec); ok {
			return false
		}

		param, ok := field.Annotations[antHTTPRef]
		if !ok || param[0:6] != "params" {
			hasNonParams = true
			return true
		}

		return false
	}
	walkFieldGroups(compile.FieldGroup(funcSpec.ArgsSpec), visitor)
	return hasNonParams
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
		if err == nil && IsStructType(funcSpec.ArgsSpec[0].Type) {
			ms.ShortRequestType = ms.RequestType
			ms.RequestType = "*" + ms.RequestType
		}
	} else {
		ms.RequestBoxed = false

		goPackageName, err := packageHelper.TypePackageName(curThriftFile)
		if err == nil {
			ms.ShortRequestType = goPackageName + "." +
				ms.ThriftService + "_" + strings.Title(ms.Name) + "_Args"
			ms.RequestType = "*" + ms.ShortRequestType
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
	ms.ShortResponseType = typeName
	if IsStructType(respSpec.ReturnType) {
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

	switch compile.RootTypeSpec(respSpec.ReturnType).(type) {
	case *compile.BoolSpec, *compile.I8Spec, *compile.I16Spec, *compile.I32Spec,
		*compile.I64Spec, *compile.DoubleSpec, *compile.StringSpec, *compile.EnumSpec:
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
) (string, bool, bool) {
	var identifier string
	var required bool
	visitor := func(
		goPrefix string, thriftPrefix string, field *compile.FieldSpec,
	) bool {
		if param, ok := field.Annotations[antHTTPRef]; ok {
			if param == "params."+paramName[1:] {
				identifier = goPrefix + "." + pascalCase(field.Name)
				required = field.Required
				return true
			}
		}

		return false
	}
	walkFieldGroups(fields, visitor)

	if identifier == "" {
		return "", required, false
	}

	return identifier, required, true
}

func (ms *MethodSpec) setRequestParamFields(
	funcSpec *compile.FunctionSpec, packageHelper *PackageHelper,
) error {
	statements := LineBuilder{}

	seenStructs, err := findStructs(funcSpec, packageHelper)
	if err != nil {
		return err
	}

	for _, segment := range ms.PathSegments {
		if segment.Type != "param" {
			continue
		}

		for seenStruct, typeName := range seenStructs {
			if strings.HasPrefix(segment.BodyIdentifier, seenStruct) {
				statements.appendf("if requestBody%s == nil {",
					seenStruct,
				)
				statements.appendf("\trequestBody%s = &%s{}",
					seenStruct, typeName,
				)
				statements.append("}")
			}
		}

		if segment.Required {
			statements.appendf("requestBody%s = req.Params.ByName(%q)",
				segment.BodyIdentifier, segment.ParamName,
			)
		} else {
			statements.appendf(
				"requestBody%s = ptr.String(req.Params.ByName(%q))",
				segment.BodyIdentifier,
				segment.ParamName,
			)
		}
	}

	ms.RequestParamGoStatements = statements.GetLines()

	return nil
}

func findStructs(
	funcSpec *compile.FunctionSpec, packageHelper *PackageHelper,
) (map[string]string, error) {
	fields := compile.FieldGroup(funcSpec.ArgsSpec)
	var seenStructs = map[string]string{}
	var finalError error

	visitor := func(
		goPrefix string, thriftPrefix string, field *compile.FieldSpec,
	) bool {
		realType := compile.RootTypeSpec(field.Type)
		longFieldName := goPrefix + "." + pascalCase(field.Name)

		if _, ok := realType.(*compile.StructSpec); ok {
			typeName, err := GoType(packageHelper, realType)
			if err != nil {
				finalError = err
				return true
			}

			seenStructs[longFieldName] = typeName
		}

		return false
	}
	walkFieldGroups(fields, visitor)

	if finalError != nil {
		return nil, finalError
	}

	return seenStructs, nil
}

func (ms *MethodSpec) setEndpointRequestHeaderFields(
	funcSpec *compile.FunctionSpec, packageHelper *PackageHelper,
) error {
	fields := compile.FieldGroup(funcSpec.ArgsSpec)
	// ms.ReqHeaderFields = map[string]HeaderFieldInfo{}

	statements := LineBuilder{}

	var finalError error
	var seenHeaders bool
	var headersMap = map[string]int{}
	var seenOptStructs = map[string]string{}

	// Scan for all annotations
	visitor := func(
		goPrefix string, thriftPrefix string, field *compile.FieldSpec,
	) bool {
		realType := compile.RootTypeSpec(field.Type)
		longFieldName := goPrefix + "." + pascalCase(field.Name)

		// If the type is a struct then we cannot really do anything
		if _, ok := realType.(*compile.StructSpec); ok {
			// if a field is a struct then we must do a nil check

			typeName, err := GoType(packageHelper, realType)
			if err != nil {
				finalError = err
				return true
			}

			if field.Required {
				statements.appendf("if requestBody%s == nil {", longFieldName)
				statements.appendf("\trequestBody%s = &%s{}",
					longFieldName, typeName,
				)
				statements.append("}")
			} else {
				seenOptStructs[longFieldName] = typeName
			}

			return false
		}

		if param, ok := field.Annotations[antHTTPRef]; ok {
			if param[0:8] == "headers." {
				headerName := param[8:]
				camelHeaderName := camelCase(headerName)
				bodyIdentifier := goPrefix + "." + pascalCase(field.Name)

				seenCount := headersMap[camelHeaderName]
				var variableName string
				if seenCount > 0 {
					variableName = camelHeaderName + "No" +
						strconv.Itoa(seenCount) + "Value"
				} else {
					variableName = camelHeaderName + "Value"
				}
				headersMap[camelHeaderName] = seenCount + 1

				if field.Required {
					statements.appendf("%s, _ := req.Header.Get(%q)",
						variableName, headerName,
					)

					for seenStruct, typeName := range seenOptStructs {
						if strings.HasPrefix(longFieldName, seenStruct) {
							statements.appendf("if requestBody%s == nil {",
								seenStruct,
							)
							statements.appendf("\trequestBody%s = &%s{}",
								seenStruct, typeName,
							)
							statements.append("}")
						}
					}

					statements.appendf("requestBody%s = %s",
						bodyIdentifier, variableName,
					)
				} else {
					statements.appendf("%s, %sExists := req.Header.Get(%q)",
						variableName, variableName, headerName,
					)
					statements.appendf("if %sExists {", variableName)

					for seenStruct, typeName := range seenOptStructs {
						if strings.HasPrefix(longFieldName, seenStruct) {
							statements.appendf("\tif requestBody%s == nil {",
								seenStruct,
							)
							statements.appendf("\t\trequestBody%s = &%s{}",
								seenStruct, typeName,
							)
							statements.append("\t}")
						}
					}

					statements.appendf("\trequestBody%s = ptr.String(%s)",
						bodyIdentifier, variableName,
					)
					statements.append("}")
				}

				seenHeaders = true
			}
		}
		return false
	}
	walkFieldGroups(fields, visitor)

	if finalError != nil {
		return finalError
	}

	if seenHeaders {
		ms.ReqHeaderGoStatements = statements.GetLines()
	}
	return nil
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
	visitor := func(
		goPrefix string, thriftPrefix string, field *compile.FieldSpec,
	) bool {
		if param, ok := field.Annotations[antHTTPRef]; ok {
			if param[0:8] == "headers." {
				headerName := param[8:]
				ms.ResHeaderFields[headerName] = HeaderFieldInfo{
					FieldIdentifier: goPrefix + "." + pascalCase(field.Name),
					IsPointer:       !field.Required,
				}
			}
		}
		return false
	}
	walkFieldGroups(fields, visitor)
}

func (ms *MethodSpec) setClientRequestHeaderFields(
	funcSpec *compile.FunctionSpec, packageHelper *PackageHelper,
) error {
	fields := compile.FieldGroup(funcSpec.ArgsSpec)

	statements := LineBuilder{}
	var finalError error
	var seenOptStructs = map[string]string{}

	// Scan for all annotations
	visitor := func(
		goPrefix string, thriftPrefix string, field *compile.FieldSpec,
	) bool {
		realType := compile.RootTypeSpec(field.Type)
		longFieldName := goPrefix + "." + pascalCase(field.Name)

		// If the type is a struct then we cannot really do anything
		if _, ok := realType.(*compile.StructSpec); ok {
			// if a field is a struct then we must do a nil check
			typeName, err := GoType(packageHelper, realType)
			if err != nil {
				finalError = err
				return true
			}
			seenOptStructs[longFieldName] = typeName
			return false
		}

		if param, ok := field.Annotations[antHTTPRef]; ok {
			if param[0:8] == "headers." {
				headerName := param[8:]
				bodyIdentifier := goPrefix + "." + pascalCase(field.Name)
				var headerNameValuePair string
				if field.Required {
					headerNameValuePair = "headers[%q]= r%s"
				} else {
					headerNameValuePair = "headers[%q]= *r%s"
				}
				if len(seenOptStructs) == 0 {
					statements.appendf(headerNameValuePair,
						headerName, bodyIdentifier,
					)
				} else {
					closeFunction := ""
					for seenStruct := range seenOptStructs {
						if strings.HasPrefix(longFieldName, seenStruct) {
							statements.appendf("if r%s != nil {", seenStruct)
							closeFunction = closeFunction + "}"
						}
					}
					statements.appendf(headerNameValuePair,
						headerName, bodyIdentifier,
					)
					statements.append(closeFunction)
				}
			}
		}
		return false
	}
	walkFieldGroups(fields, visitor)
	if finalError != nil {
		return finalError
	}

	ms.ReqClientHeaderGoStatements = statements.GetLines()
	return nil
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
			var required bool
			var ok bool
			if ms.RequestBoxed {
				// Boxed requests mean first arg is struct
				structType := funcSpec.ArgsSpec[0].Type.(*compile.StructSpec)
				fieldSelect, required, ok = findParamsAnnotation(
					structType.Fields, segment,
				)
			} else {
				fieldSelect, required, ok = findParamsAnnotation(
					compile.FieldGroup(funcSpec.ArgsSpec), segment,
				)
			}

			if !ok {
				panic("cannot find params: " + segment)
			}
			ms.PathSegments[i].BodyIdentifier = fieldSelect
			ms.PathSegments[i].ParamName = segment[1:]
			ms.PathSegments[i].Required = required
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
	reqTransforms map[string]FieldMapperEntry,
	respTransforms map[string]FieldMapperEntry,
	h *PackageHelper,
	downstreamMethod *MethodSpec,
) error {
	// TODO(sindelar): Iterate over fields that are structs (for foo/bar examples).

	// Add type checking and conversion, custom mapping
	structType := compile.FieldGroup(funcSpec.ArgsSpec)
	downstreamStructType := compile.FieldGroup(downstreamSpec.ArgsSpec)

	typeConverter := NewTypeConverter(h, RequestHelper{
		RequestSuffix:     "ClientRequest",
		RequestInputType:  ms.RequestType,
		RequestOutputType: downstreamMethod.RequestType,
		ResponseType:      downstreamMethod.ShortRequestType,
		OutputMethodName:  ms.Name,
	})

	err := typeConverter.GenStructConverter(structType, downstreamStructType, reqTransforms, false)
	if err != nil {
		return err
	}
	ms.ConvertRequestGoStatements = typeConverter.GetLines()

	// TODO: support non-struct return types
	respType := funcSpec.ResultSpec.ReturnType
	downstreamRespType := downstreamSpec.ResultSpec.ReturnType

	if respType == nil || downstreamRespType == nil {
		return nil
	}

	respConverter := NewTypeConverter(h, RequestHelper{
		RequestSuffix:     "ClientResponse",
		RequestInputType:  downstreamMethod.ResponseType,
		RequestOutputType: ms.ResponseType,
		ResponseType:      ms.ShortResponseType,
		OutputMethodName:  ms.Name,
	})

	var respFields, downstreamRespFields []*compile.FieldSpec
	var isPrimitive bool
	switch respType.(type) {
	case
		*compile.BoolSpec,
		*compile.I8Spec,
		*compile.I16Spec,
		*compile.I32Spec,
		*compile.EnumSpec,
		*compile.I64Spec,
		*compile.DoubleSpec,
		*compile.StringSpec:

		isPrimitive = true
	default:
		respFields = respType.(*compile.StructSpec).Fields
		downstreamRespFields = downstreamRespType.(*compile.StructSpec).Fields
		isPrimitive = false
	}
	err = respConverter.GenStructConverter(downstreamRespFields, respFields, respTransforms, isPrimitive)
	if err != nil {
		return err
	}
	ms.ConvertResponseGoStatements = respConverter.GetLines()

	return nil
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
		panic(fmt.Sprintf(
			"Unknown type (%T) %v for query string parameter",
			typeSpec, typeSpec,
		))
	}

	return queryMethod
}

func getQueryEncodeExpression(
	typeSpec compile.TypeSpec, valueName string,
) string {
	var encodeExpression string

	switch typeSpec.(type) {
	case *compile.BoolSpec:
		encodeExpression = "strconv.FormatBool(%s)"
	case *compile.I8Spec:
		encodeExpression = "strconv.Itoa(int(%s))"
	case *compile.I16Spec:
		encodeExpression = "strconv.Itoa(int(%s))"
	case *compile.I32Spec:
		encodeExpression = "strconv.Itoa(int(%s))"
	case *compile.I64Spec:
		encodeExpression = "strconv.FormatInt(%s, 10)"
	case *compile.DoubleSpec:
		encodeExpression = "strconv.FormatFloat(%s, 'G', -1, 64)"
	case *compile.StringSpec:
		encodeExpression = "%s"
	default:
		panic(fmt.Sprintf(
			"Unknown type (%T) %v for query string parameter",
			typeSpec, typeSpec,
		))
	}

	return fmt.Sprintf(encodeExpression, valueName)
}

func (ms *MethodSpec) setWriteQueryParamStatements(
	funcSpec *compile.FunctionSpec, packageHelper *PackageHelper,
) error {
	var statements LineBuilder
	var hasQueryFields bool
	var stack = []string{}

	visitor := func(
		goPrefix string, thriftPrefix string, field *compile.FieldSpec,
	) bool {
		realType := compile.RootTypeSpec(field.Type)
		longFieldName := goPrefix + "." + pascalCase(field.Name)

		if len(stack) > 0 {
			if !strings.HasPrefix(longFieldName, stack[len(stack)-1]) {
				stack = stack[:len(stack)-1]
				statements.append("}")
			}
		}

		if _, ok := realType.(*compile.StructSpec); ok {
			// If a field is a struct then skip

			if field.Required {
				statements.appendf("if r%s == nil {", longFieldName)
				// TODO: generate correct number of nils...
				statements.append("\treturn nil, nil, errors.New(")
				statements.appendf("\t\t\"The field %s is required\",",
					longFieldName,
				)
				statements.append("\t)")
				statements.appendf("}")
			} else {
				stack = append(stack, longFieldName)

				statements.appendf("if r%s != nil {", longFieldName)
			}

			return false
		}

		httpRefAnnotation := field.Annotations[antHTTPRef]
		if httpRefAnnotation != "" && !strings.HasPrefix(httpRefAnnotation, "query") {
			return false
		}

		longQueryName := getLongQueryName(field, thriftPrefix)
		identifierName := camelCase(longQueryName) + "Query"

		if !hasQueryFields {
			statements.append("queryValues := &url.Values{}")
			hasQueryFields = true
		}

		if field.Required {
			encodeExpr := getQueryEncodeExpression(
				realType, "r"+longFieldName,
			)

			statements.appendf("%s := %s",
				identifierName, encodeExpr,
			)
			statements.appendf("queryValues.Set(\"%s\", %s)",
				longQueryName, identifierName,
			)
		} else {
			encodeExpr := getQueryEncodeExpression(
				realType, "*r"+longFieldName,
			)

			statements.appendf("if r%s != nil {", longFieldName)
			statements.appendf("\t%s := %s",
				identifierName, encodeExpr,
			)
			statements.appendf("\tqueryValues.Set(\"%s\", %s)",
				longQueryName, identifierName,
			)
			statements.append("}")
		}

		return false
	}
	walkFieldGroups(compile.FieldGroup(funcSpec.ArgsSpec), visitor)

	for i := 0; i < len(stack); i++ {
		statements.append("}")
	}

	if hasQueryFields {
		statements.append("fullURL += \"?\" + queryValues.Encode()")
	}
	ms.WriteQueryParamGoStatements = statements.GetLines()
	return nil
}

func (ms *MethodSpec) setParseQueryParamStatements(
	funcSpec *compile.FunctionSpec, packageHelper *PackageHelper,
) error {
	// If a thrift field has a http.ref annotation then we
	// should not read this field from query parameters.
	var statements LineBuilder

	var finalError error
	var stack = []string{}

	visitor := func(
		goPrefix string, thriftPrefix string, field *compile.FieldSpec,
	) bool {
		realType := compile.RootTypeSpec(field.Type)
		longFieldName := goPrefix + "." + pascalCase(field.Name)
		longQueryName := getLongQueryName(field, thriftPrefix)

		if len(stack) > 0 {
			if !strings.HasPrefix(longFieldName, stack[len(stack)-1]) {
				stack = stack[:len(stack)-1]
				statements.append("}")
			}
		}

		// If the type is a struct then we cannot really do anything
		if _, ok := realType.(*compile.StructSpec); ok {
			// if a field is a struct then we must do a nil check

			typeName, err := GoType(packageHelper, realType)
			if err != nil {
				finalError = err
				return true
			}

			if !field.Required {
				stack = append(stack, longFieldName)

				statements.appendf(
					"if req.HasQueryPrefix(%q) || requestBody%s != nil {",
					longQueryName,
					longFieldName,
				)
			}

			statements.appendf("if requestBody%s == nil {", longFieldName)
			statements.appendf("\trequestBody%s = &%s{}",
				longFieldName, typeName,
			)
			statements.append("}")

			return false
		}

		identifierName := camelCase(longQueryName) + "Query"

		httpRefAnnotation := field.Annotations[antHTTPRef]
		if httpRefAnnotation != "" && !strings.HasPrefix(httpRefAnnotation, "query") {
			return false
		}

		okIdentifierName := camelCase(longQueryName) + "Ok"
		if field.Required {
			statements.appendf("%s := req.CheckQueryValue(%q)",
				okIdentifierName, longQueryName,
			)
			statements.appendf("if !%s {", okIdentifierName)
			statements.append("\treturn")
			statements.append("}")
		} else {
			statements.appendf("%s := req.HasQueryValue(%q)",
				okIdentifierName, longQueryName,
			)
			statements.appendf("if %s {", okIdentifierName)
		}

		queryMethodName := getQueryMethodForType(realType)
		pointerMethod := pointerMethodType(realType)

		statements.appendf("%s, ok := req.%s(%q)",
			identifierName, queryMethodName, longQueryName,
		)

		statements.append("if !ok {")
		statements.append("\treturn")
		statements.append("}")

		if field.Required {
			statements.appendf("requestBody%s = %s",
				longFieldName, identifierName,
			)
		} else {
			statements.appendf("\trequestBody%s = ptr.%s(%s)",
				longFieldName, pointerMethod, identifierName,
			)
			statements.append("}")
		}

		// new line after block.
		statements.append("")
		return false
	}
	walkFieldGroups(compile.FieldGroup(funcSpec.ArgsSpec), visitor)

	for i := 0; i < len(stack); i++ {
		statements.append("}")
	}

	if finalError != nil {
		return finalError
	}

	ms.ParseQueryParamGoStatements = statements.GetLines()
	return nil
}

func getLongQueryName(field *compile.FieldSpec, thriftPrefix string) string {
	var longQueryName string

	queryName := field.Name
	queryAnnotation := field.Annotations[antHTTPRef]
	if strings.HasPrefix(queryAnnotation, "query.") {
		// len("query.") == 6
		queryName = queryAnnotation[6:]
	}

	if thriftPrefix == "" {
		longQueryName = queryName
	} else if thriftPrefix[0] == '.' {
		longQueryName = thriftPrefix[1:] + "." + queryName
	} else {
		longQueryName = thriftPrefix + "." + queryName
	}

	return longQueryName
}

func isRequestBoxed(f *compile.FunctionSpec) bool {
	boxed, ok := f.Annotations[AntHTTPReqDefBoxed]
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
