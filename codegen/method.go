// Copyright (c) 2019 Uber Technologies, Inc.
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
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

const (
	antHTTPMethod     = "%s.http.method"
	antHTTPPath       = "%s.http.path"
	antHTTPStatus     = "%s.http.status"
	antHTTPReqHeaders = "%s.http.reqHeaders"
	antHTTPResHeaders = "%s.http.resHeaders"
	antHTTPRef        = "%s.http.ref"
	antMeta           = "%s.meta"
	antHandler        = "%s.handler"

	// AntHTTPReqDefBoxed annotates a method so that the genereted method takes
	// generated argument directly instead of a struct that warps the argument.
	// The annotated method should have one and only one argument.
	AntHTTPReqDefBoxed = "%s.http.req.def"
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
	annotations  annotations
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

	RequestType            string
	ShortRequestType       string
	ResponseType           string
	ShortResponseType      string
	OKStatusCode           StatusCode
	Exceptions             []ExceptionSpec
	ExceptionsByStatusCode map[int][]ExceptionSpec
	ExceptionsIndex        map[string]ExceptionSpec
	ValidStatusCodes       []int
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

	// Statements for propagating headers to client requests
	PropagateHeadersGoStatements []string

	// Statements for reading data out of url params (server)
	RequestParamGoStatements []string
}

type annotations struct {
	HTTPMethod      string
	HTTPPath        string
	HTTPStatus      string
	HTTPReqHeaders  string
	HTTPResHeaders  string
	HTTPRef         string
	Meta            string
	Handler         string
	HTTPReqDefBoxed string
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

// NewMethod creates new method specification.
func NewMethod(
	thriftFile string,
	funcSpec *compile.FunctionSpec,
	packageHelper *PackageHelper,
	wantAnnot bool,
	isEndpoint bool,
	thriftService string,
) (*MethodSpec, error) {
	var (
		err    error
		ok     bool
		ant    = packageHelper.annotationPrefix
		method = &MethodSpec{}
	)
	method.CompiledThriftSpec = funcSpec
	method.Name = funcSpec.MethodName()
	method.IsEndpoint = isEndpoint
	method.WantAnnot = wantAnnot
	method.ThriftService = thriftService
	method.annotations = annotations{
		HTTPMethod:      fmt.Sprintf(antHTTPMethod, ant),
		HTTPPath:        fmt.Sprintf(antHTTPPath, ant),
		HTTPStatus:      fmt.Sprintf(antHTTPStatus, ant),
		HTTPReqHeaders:  fmt.Sprintf(antHTTPReqHeaders, ant),
		HTTPResHeaders:  fmt.Sprintf(antHTTPResHeaders, ant),
		HTTPRef:         fmt.Sprintf(antHTTPRef, ant),
		Meta:            fmt.Sprintf(antMeta, ant),
		Handler:         fmt.Sprintf(antHandler, ant),
		HTTPReqDefBoxed: fmt.Sprintf(AntHTTPReqDefBoxed, ant),
	}

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

	method.ReqHeaders = headers(funcSpec.Annotations[method.annotations.HTTPReqHeaders])
	method.ResHeaders = headers(funcSpec.Annotations[method.annotations.HTTPResHeaders])

	if !wantAnnot {
		return method, nil
	}

	if method.HTTPMethod, ok = funcSpec.Annotations[method.annotations.HTTPMethod]; !ok {
		return nil, errors.Errorf("missing annotation '%s' for HTTP method", method.annotations.HTTPMethod)
	}

	method.EndpointName = funcSpec.Annotations[method.annotations.Handler]

	err = method.setOKStatusCode(funcSpec.Annotations[method.annotations.HTTPStatus])
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
	if httpPath, ok = funcSpec.Annotations[method.annotations.HTTPPath]; !ok {
		return nil, errors.Errorf(
			"missing annotation '%s' for HTTP path", method.annotations.HTTPPath,
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

func (ms *MethodSpec) scanForNonParams(funcSpec *compile.FunctionSpec) bool {
	hasNonParams := false

	visitor := func(
		goPrefix string, thriftPrefix string, field *compile.FieldSpec,
	) bool {
		realType := compile.RootTypeSpec(field.Type)
		// ignore nested structs
		if _, ok := realType.(*compile.StructSpec); ok {
			return false
		}

		param, ok := field.Annotations[ms.annotations.HTTPRef]
		if !ok || strings.HasPrefix(param, "params") {
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
	if ms.isRequestBoxed(funcSpec) {
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
		return errors.Errorf("no http OK status code set by annotation '%s' ", ms.annotations.HTTPStatus)
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
	ms.ValidStatusCodes = []int{
		ms.OKStatusCode.Code,
	}

	for code := range ms.ExceptionsByStatusCode {
		ms.ValidStatusCodes = append(ms.ValidStatusCodes, code)
	}

	// Prevents non-deterministic builds
	sort.Ints(ms.ValidStatusCodes)
}

func (ms *MethodSpec) setExceptions(
	curThriftFile string,
	isEndpoint bool,
	resultSpec *compile.ResultSpec,
	h *PackageHelper,
) error {
	ms.Exceptions = make([]ExceptionSpec, len(resultSpec.Exceptions))
	ms.ExceptionsIndex = make(
		map[string]ExceptionSpec, len(resultSpec.Exceptions),
	)
	ms.ExceptionsByStatusCode = map[int][]ExceptionSpec{}

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
			if _, exists := ms.ExceptionsByStatusCode[exception.StatusCode.Code]; !exists {
				ms.ExceptionsByStatusCode[exception.StatusCode.Code] = []ExceptionSpec{}
			}
			ms.ExceptionsByStatusCode[exception.StatusCode.Code] = append(
				ms.ExceptionsByStatusCode[exception.StatusCode.Code],
				exception,
			)
			continue
		}

		code, err := strconv.Atoi(e.Annotations[ms.annotations.HTTPStatus])
		if err != nil {
			return errors.Wrapf(
				err,
				"cannot parse the annotation %s for exception %s", ms.annotations.HTTPStatus, e.Name,
			)
		}

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
		if _, exists := ms.ExceptionsByStatusCode[exception.StatusCode.Code]; !exists {
			ms.ExceptionsByStatusCode[exception.StatusCode.Code] = []ExceptionSpec{}
		}
		ms.ExceptionsByStatusCode[exception.StatusCode.Code] = append(
			ms.ExceptionsByStatusCode[exception.StatusCode.Code],
			exception,
		)
	}
	return nil
}

func (ms *MethodSpec) findParamsAnnotation(
	fields compile.FieldGroup, paramName string,
) (string, bool, bool) {
	var identifier string
	var required bool
	visitor := func(
		goPrefix string, thriftPrefix string, field *compile.FieldSpec,
	) bool {
		if param, ok := field.Annotations[ms.annotations.HTTPRef]; ok {
			if param == "params."+paramName[1:] {
				identifier = goPrefix + "." + PascalCase(field.Name)
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
			statements.appendf("requestBody%s = req.Params.Get(%q)",
				segment.BodyIdentifier, segment.ParamName,
			)
		} else {
			statements.appendf(
				"requestBody%s = ptr.String(req.Params.Get(%q))",
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
		longFieldName := goPrefix + "." + PascalCase(field.Name)

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
		longFieldName := goPrefix + "." + PascalCase(field.Name)

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

		if param, ok := field.Annotations[ms.annotations.HTTPRef]; ok {
			if strings.HasPrefix(param, "headers.") {
				headerName := param[8:]
				camelHeaderName := CamelCase(headerName)

				fieldThriftType, err := GoType(packageHelper, field.Type)
				if err != nil {
					finalError = err
					return true
				}

				bodyIdentifier := goPrefix + "." + PascalCase(field.Name)

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

					statements.appendf("requestBody%s = %s(%s)",
						bodyIdentifier, fieldThriftType, variableName,
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

					switch fieldThriftType {
					case "string":
						statements.appendf("\trequestBody%s = ptr.String(%s)",
							bodyIdentifier, variableName,
						)
					case "int64":
						statements.appendf("body, _ := strconv.ParseInt(%s, 10, 64)",
							variableName,
						)
						statements.appendf("requestBody%s = &body", bodyIdentifier)
					case "bool":
						statements.appendf("body, _ := strconv.ParseBool(%s)",
							variableName,
						)
						statements.appendf("requestBody%s = &body", bodyIdentifier)
					case "float64":
					case "float32":
						statements.appendf("body, _ := strconv.ParseFloat(%s, 64)",
							variableName,
						)
						statements.appendf("requestBody%s = &body", bodyIdentifier)
					default:
						statements.appendf("body := %s(%s)",
							fieldThriftType, variableName,
						)
						statements.appendf("requestBody%s = &body", bodyIdentifier)

					}
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
		if param, ok := field.Annotations[ms.annotations.HTTPRef]; ok {
			if strings.HasPrefix(param, "headers.") {
				headerName := param[8:]
				ms.ResHeaderFields[headerName] = HeaderFieldInfo{
					FieldIdentifier: goPrefix + "." + PascalCase(field.Name),
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
		longFieldName := goPrefix + "." + PascalCase(field.Name)

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

		if param, ok := field.Annotations[ms.annotations.HTTPRef]; ok {
			if strings.HasPrefix(param, "headers.") {
				headerName := param[8:]
				bodyIdentifier := goPrefix + "." + PascalCase(field.Name)
				var headerNameValuePair string
				if field.Required {
					// Note header values are always string
					headerNameValuePair = "headers[%q]= string(r%s)"
				} else {
					headerNameValuePair = "headers[%q]= string(*r%s)"
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
				fieldSelect, required, ok = ms.findParamsAnnotation(
					structType.Fields, segment,
				)
			} else {
				fieldSelect, required, ok = ms.findParamsAnnotation(
					compile.FieldGroup(funcSpec.ArgsSpec), segment,
				)
			}

			if !ok {
				panic(fmt.Sprintf("cannot find params: %s for http path %s", segment, httpPath))
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

func (ms *MethodSpec) setHeaderPropagator(
	reqHeaders []string,
	downstreamSpec *compile.FunctionSpec,
	headersPropagate map[string]FieldMapperEntry,
	h *PackageHelper,
	downstreamMethod *MethodSpec,
) error {
	downstreamStructType := compile.FieldGroup(downstreamSpec.ArgsSpec)
	hp := NewHeaderPropagator(h)
	hp.append(
		"func propagateHeaders",
		PascalCase(ms.Name),
		"ClientRequests(in ",
		downstreamMethod.RequestType,
		", headers zanzibar.Header) ",
		downstreamMethod.RequestType,
		"{",
	)

	hp.append("if in == nil {")
	hp.append(fmt.Sprintf(`in = %s{}`, strings.Replace(downstreamMethod.RequestType, "*", "&", 1)))
	hp.append("}")

	err := hp.Propagate(reqHeaders, downstreamStructType, headersPropagate)
	if err != nil {
		return err
	}
	hp.append("return in")
	hp.append("}")
	ms.PropagateHeadersGoStatements = hp.GetLines()
	return nil
}

func (ms *MethodSpec) setTypeConverters(
	funcSpec *compile.FunctionSpec,
	downstreamSpec *compile.FunctionSpec,
	reqTransforms map[string]FieldMapperEntry,
	headersPropagate map[string]FieldMapperEntry,
	respTransforms map[string]FieldMapperEntry,
	h *PackageHelper,
	downstreamMethod *MethodSpec,
) error {
	// TODO(sindelar): Iterate over fields that are structs (for foo/bar examples).

	// Add type checking and conversion, custom mapping
	structType := compile.FieldGroup(funcSpec.ArgsSpec)
	downstreamStructType := compile.FieldGroup(downstreamSpec.ArgsSpec)

	typeConverter := NewTypeConverter(h, headersPropagate)

	typeConverter.append(
		"func convertTo",
		PascalCase(ms.Name),
		"ClientRequest(in ", ms.RequestType, ") ", downstreamMethod.RequestType, "{")

	typeConverter.append("out := &", downstreamMethod.ShortRequestType, "{}\n")

	err := typeConverter.GenStructConverter(structType, downstreamStructType, reqTransforms)
	if err != nil {
		return err
	}
	typeConverter.append("\nreturn out")
	typeConverter.append("}")
	ms.ConvertRequestGoStatements = typeConverter.GetLines()

	// TODO: support non-struct return types
	respType := funcSpec.ResultSpec.ReturnType

	downstreamRespType := downstreamMethod.CompiledThriftSpec.ResultSpec.ReturnType

	if respType == nil || downstreamRespType == nil {
		return nil
	}

	respConverter := NewTypeConverter(h, nil)

	respConverter.append(
		"func convert",
		PascalCase(ms.DownstreamService), PascalCase(ms.Name),
		"ClientResponse(in ", downstreamMethod.ResponseType, ") ", ms.ResponseType, "{")
	var respFields, downstreamRespFields []*compile.FieldSpec
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

		respConverter.append("out", " := in\t\n")
	default:
		// default as struct
		respFields = respType.(*compile.StructSpec).Fields
		downstreamRespFields = downstreamRespType.(*compile.StructSpec).Fields
		respConverter.append("out", " := ", "&", ms.ShortResponseType, "{}\t\n")
		err = respConverter.GenStructConverter(downstreamRespFields, respFields, respTransforms)
		if err != nil {
			return err
		}
	}
	respConverter.append("\nreturn out \t}")
	ms.ConvertResponseGoStatements = respConverter.GetLines()

	return nil
}

func getQueryMethodForPrimitiveType(typeSpec compile.TypeSpec) string {
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
	case *compile.EnumSpec:
		queryMethod = "GetQueryInt32"
	case *compile.I64Spec:
		queryMethod = "GetQueryInt64"
	case *compile.DoubleSpec:
		queryMethod = "GetQueryFloat64"
	case *compile.StringSpec:
		queryMethod = "GetQueryValue"
	default:
		panic(fmt.Sprintf(
			"Unsupported type (%T) for %s as query string parameter",
			typeSpec, typeSpec.ThriftName(),
		))
	}

	return queryMethod
}

func getQueryMethodForType(typeSpec compile.TypeSpec) string {
	var queryMethod string

	switch t := typeSpec.(type) {
	case *compile.ListSpec:
		queryMethod = getQueryMethodForPrimitiveType(compile.RootTypeSpec(t.ValueSpec)) + "List"
	case *compile.SetSpec:
		queryMethod = getQueryMethodForPrimitiveType(compile.RootTypeSpec(t.ValueSpec)) + "Set"
	default:
		queryMethod = getQueryMethodForPrimitiveType(typeSpec)
	}

	return queryMethod
}

func getQueryEncodeExprPrimitive(typeSpec compile.TypeSpec) string {
	var encodeExpression string

	_, isTypedef := typeSpec.(*compile.TypedefSpec)

	switch compile.RootTypeSpec(typeSpec).(type) {
	case *compile.BoolSpec:
		if isTypedef {
			encodeExpression = "strconv.FormatBool(bool(%s))"
		} else {
			encodeExpression = "strconv.FormatBool(%s)"
		}
	case *compile.I8Spec:
		encodeExpression = "strconv.Itoa(int(%s))"
	case *compile.I16Spec:
		encodeExpression = "strconv.Itoa(int(%s))"
	case *compile.I32Spec:
		encodeExpression = "strconv.Itoa(int(%s))"
	case *compile.I64Spec:
		if isTypedef {
			encodeExpression = "strconv.FormatInt(int64(%s), 10)"
		} else {
			encodeExpression = "strconv.FormatInt(%s, 10)"
		}
	case *compile.DoubleSpec:
		if isTypedef {
			encodeExpression = "strconv.FormatFloat(float64(%s), 'G', -1, 64)"
		} else {
			encodeExpression = "strconv.FormatFloat(%s, 'G', -1, 64)"
		}
	case *compile.StringSpec:
		if isTypedef {
			encodeExpression = "string(%s)"
		} else {
			encodeExpression = "%s"
		}
	case *compile.EnumSpec:
		encodeExpression = "strconv.Itoa(int(%s))"
	default:
		panic(fmt.Sprintf(
			"Unsupported type (%T) for %s as query string parameter",
			typeSpec, typeSpec.ThriftName(),
		))
	}
	return encodeExpression
}

func getQueryEncodeExpression(typeSpec compile.TypeSpec, valueName string) string {
	var encodeExpression string

	switch t := compile.RootTypeSpec(typeSpec).(type) {
	case *compile.ListSpec:
		encodeExpression = getQueryEncodeExprPrimitive(t.ValueSpec)
	case *compile.SetSpec:
		encodeExpression = getQueryEncodeExprPrimitive(t.ValueSpec)
	default:
		encodeExpression = getQueryEncodeExprPrimitive(typeSpec)
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
		longFieldName := goPrefix + "." + PascalCase(field.Name)

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

		httpRefAnnotation := field.Annotations[ms.annotations.HTTPRef]
		if httpRefAnnotation != "" && !strings.HasPrefix(httpRefAnnotation, "query") {
			return false
		}

		longQueryName := ms.getLongQueryName(field, thriftPrefix)
		identifierName := CamelCase(longQueryName) + "Query"
		_, isList := realType.(*compile.ListSpec)
		_, isSet := realType.(*compile.SetSpec)

		if !hasQueryFields {
			statements.append("queryValues := &url.Values{}")
			hasQueryFields = true
		}

		if field.Required {
			if isList {
				encodeExpr := getQueryEncodeExpression(field.Type, "value")
				statements.appendf("for _, value := range %s {", "r"+longFieldName)
				statements.appendf("\tqueryValues.Add(\"%s\", %s)", longQueryName, encodeExpr)
				statements.append("}")
			} else if isSet {
				encodeExpr := getQueryEncodeExpression(field.Type, "value")
				statements.appendf("for value := range %s {", "r"+longFieldName)
				statements.appendf("\tqueryValues.Add(\"%s\", %s)", longQueryName, encodeExpr)
				statements.append("}")
			} else {
				encodeExpr := getQueryEncodeExpression(field.Type, "r"+longFieldName)
				statements.appendf("%s := %s", identifierName, encodeExpr)
				statements.appendf("queryValues.Set(\"%s\", %s)", longQueryName, identifierName)
			}
		} else {
			statements.appendf("if r%s != nil {", longFieldName)
			if isList {
				encodeExpr := getQueryEncodeExpression(field.Type, "value")
				statements.appendf("for _, value := range %s {", "r"+longFieldName)
				statements.appendf("\tqueryValues.Add(\"%s\", %s)", longQueryName, encodeExpr)
				statements.append("}")
			} else if isSet {
				encodeExpr := getQueryEncodeExpression(field.Type, "value")
				statements.appendf("for value := range %s {", "r"+longFieldName)
				statements.appendf("\tqueryValues.Add(\"%s\", %s)", longQueryName, encodeExpr)
				statements.append("}")
			} else {
				encodeExpr := getQueryEncodeExpression(field.Type, "*r"+longFieldName)
				statements.appendf("\t%s := %s", identifierName, encodeExpr)
				statements.appendf("\tqueryValues.Set(\"%s\", %s)", longQueryName, identifierName)
			}
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
		longFieldName := goPrefix + "." + PascalCase(field.Name)
		longQueryName := ms.getLongQueryName(field, thriftPrefix)

		if len(stack) > 0 {
			if !strings.HasPrefix(longFieldName, stack[len(stack)-1]) {
				stack = stack[:len(stack)-1]
				statements.append("}")
			}
		}

		var err error
		var typedef string
		if _, ok := field.Type.(*compile.TypedefSpec); ok {
			typedef, err = GoType(packageHelper, field.Type)
			if err != nil {
				finalError = err
				return true
			}
		}

		if _, ok := field.Type.(*compile.EnumSpec); ok {
			typedef, err = GoType(packageHelper, field.Type)
			if err != nil {
				finalError = err
				return true
			}
		}

		var aggrValueTypedef string
		var isList, isSet bool
		switch t := realType.(type) {
		case *compile.ListSpec:
			isList = true
			if _, ok := t.ValueSpec.(*compile.TypedefSpec); ok {
				aggrValueTypedef, err = GoType(packageHelper, t.ValueSpec)
				if err != nil {
					finalError = err
					return true
				}
			}
		case *compile.SetSpec:
			isSet = true
			if _, ok := t.ValueSpec.(*compile.TypedefSpec); ok {
				aggrValueTypedef, err = GoType(packageHelper, t.ValueSpec)
				if err != nil {
					finalError = err
					return true
				}
			}
		case *compile.StructSpec:
			// If the type is a struct then we cannot really do anything

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
		identifierName := CamelCase(longQueryName) + "Query"

		httpRefAnnotation := field.Annotations[ms.annotations.HTTPRef]
		if httpRefAnnotation != "" && !strings.HasPrefix(httpRefAnnotation, "query") {
			return false
		}

		okIdentifierName := CamelCase(longQueryName) + "Ok"
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

		statements.appendf("%s, ok := req.%s(%q)",
			identifierName, queryMethodName, longQueryName,
		)

		statements.append("if !ok {")
		statements.append("\treturn")
		statements.append("}")

		target := identifierName

		indent := ""
		if !field.Required {
			indent += "\t"
		}

		// if field is a list and list value is typedef, list values must be converted first
		if aggrValueTypedef != "" {
			target = fmt.Sprintf("%sFinal", identifierName)
			if isList {
				statements.appendf(
					"%s%s := make([]%s, len(%s))",
					indent, target, aggrValueTypedef, identifierName,
				)
				statements.appendf("%sfor i, v := range %s {", indent, identifierName)
				statements.appendf("%s%s[i] = %s(v)", indent+"\t", target, aggrValueTypedef)
				statements.appendf("%s}", indent)
			} else if isSet {
				statements.appendf(
					"%s%s := make(map[%s]struct{}, len(%s))",
					indent, target, aggrValueTypedef, identifierName,
				)
				statements.appendf("%sfor _, v := range %s {", indent, identifierName)
				statements.appendf("%s%s[%s(v)] = struct{}{}", indent+"\t", target, aggrValueTypedef)
				statements.appendf("%s}", indent)
			}
		}

		if field.Required || isList || isSet {
			if typedef != "" {
				statements.appendf("%srequestBody%s = %s(%s)", indent, longFieldName, typedef, target)
			} else {
				statements.appendf("%srequestBody%s = %s", indent, longFieldName, target)
			}

		} else {
			target = fmt.Sprintf("ptr.%s(%s)", pointerMethodType(realType), identifierName)
			if typedef != "" {
				statements.appendf("%srequestBody%s = (*%s)(%s)", indent, longFieldName, typedef, target)
			} else {
				statements.appendf("%srequestBody%s = %s", indent, longFieldName, target)
			}
		}

		if !field.Required {
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

func (ms *MethodSpec) getLongQueryName(field *compile.FieldSpec, thriftPrefix string) string {
	var longQueryName string

	queryName := field.Name
	queryAnnotation := field.Annotations[ms.annotations.HTTPRef]
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

func (ms *MethodSpec) isRequestBoxed(f *compile.FunctionSpec) bool {
	boxed, ok := f.Annotations[ms.annotations.HTTPReqDefBoxed]
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
