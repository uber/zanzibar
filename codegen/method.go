// Copyright (c) 2020 Uber Technologies, Inc.
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
	antHTTPResNoBody   = "%s.http.res.body.disallow"
)

const queryAnnotationPrefix = "query."
const headerAnnotationPrefix = "headers."

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

	StatusCode       StatusCode
	IsBodyDisallowed bool
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

	// Fully qualified field type of the unboxed field
	BoxedRequestType string
	// Unboxed field name
	BoxedRequestName string
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

	// Statements for converting Clientless request types
	ConvertClientlessRequestGoStatements []string

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
	HTTPResNoBody   string
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
		HTTPResNoBody:   fmt.Sprintf(antHTTPResNoBody, ant),
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

	if method.RequestType != "" {
		hasNoBody := method.HTTPMethod == "GET"
		if method.IsEndpoint {
			err := method.setParseQueryParamStatements(funcSpec, packageHelper, hasNoBody)
			if err != nil {
				return nil, err
			}
		} else {
			err := method.setWriteQueryParamStatements(funcSpec, packageHelper, hasNoBody)
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
		ms.BoxedRequestType, err = packageHelper.TypeFullName(funcSpec.ArgsSpec[0].Type)
		ms.BoxedRequestName = PascalCase(funcSpec.ArgsSpec[0].Name)
		if err == nil && IsStructType(funcSpec.ArgsSpec[0].Type) {
			ms.BoxedRequestType = "*" + ms.BoxedRequestType
		}
	}

	goPackageName, err := packageHelper.TypePackageName(curThriftFile)
	if err == nil {
		ms.ShortRequestType = goPackageName + "." +
			ms.ThriftService + "_" + strings.Title(ms.Name) + "_Args"
		ms.RequestType = "*" + ms.ShortRequestType
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

		bodyDisallowed := ms.isBodyDisallowed(e)
		if !ms.WantAnnot {
			exception := ExceptionSpec{
				StructSpec: StructSpec{
					Type: typeName,
					Name: e.Name,
				},
				IsBodyDisallowed: bodyDisallowed,
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
			IsBodyDisallowed: bodyDisallowed,
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

	seenStructs, itrOrder, err := findStructs(funcSpec, packageHelper)
	if err != nil {
		return err
	}

	for _, segment := range ms.PathSegments {
		if segment.Type != "param" {
			continue
		}

		for _, seenStruct := range itrOrder {
			if strings.HasPrefix(segment.BodyIdentifier, seenStruct) {
				statements.appendf("if requestBody%s == nil {",
					seenStruct,
				)
				statements.appendf("\trequestBody%s = &%s{}",
					seenStruct, seenStructs[seenStruct],
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
) (map[string]string, []string, error) {
	fields := compile.FieldGroup(funcSpec.ArgsSpec)
	seenStructs := make(map[string]string)
	itrOrder := make([]string, 0)
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
			itrOrder = append(itrOrder, longFieldName)
		}

		return false
	}
	walkFieldGroups(fields, visitor)

	if finalError != nil {
		return nil, nil, finalError
	}

	return seenStructs, itrOrder, nil
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
			if strings.HasPrefix(param, headerAnnotationPrefix) {
				headerName := strings.TrimPrefix(param, headerAnnotationPrefix)
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
			if strings.HasPrefix(param, headerAnnotationPrefix) {
				headerName := strings.TrimPrefix(param, headerAnnotationPrefix)
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
	seenOptStructs := make(map[string]string)
	itrOrder := make([]string, 0)

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
			itrOrder = append(itrOrder, longFieldName)
			return false
		}

		if param, ok := field.Annotations[ms.annotations.HTTPRef]; ok {
			if strings.HasPrefix(param, headerAnnotationPrefix) {
				headerName := strings.TrimPrefix(param, headerAnnotationPrefix)
				bodyIdentifier := goPrefix + "." + PascalCase(field.Name)
				var headerNameValuePair string
				if field.Required {
					// Note header values are always string
					headerNameValuePair = "headers[%q]= string(r%s)"
				} else {
					headerNameValuePair = "headers[%q]= string(*r%s)"
				}
				if !field.Required {
					closeFunction := ""
					for _, seenStruct := range itrOrder {
						if strings.HasPrefix(longFieldName, seenStruct) {
							statements.appendf("if r%s != nil {", seenStruct)
							closeFunction = closeFunction + "}"
						}
					}

					statements.appendf("if r%s != nil {", bodyIdentifier)
					statements.appendf(headerNameValuePair, headerName, bodyIdentifier)
					statements.append("}")

					statements.append(closeFunction)
				} else {
					statements.appendf(headerNameValuePair,
						headerName, bodyIdentifier,
					)
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

			fieldSelect, required, ok = ms.findParamsAnnotation(
				compile.FieldGroup(funcSpec.ArgsSpec), segment,
			)

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

func (ms *MethodSpec) setClientlessTypeConverters(
	funcSpec *compile.FunctionSpec,
	reqTransforms map[string]FieldMapperEntry,
	headersPropagate map[string]FieldMapperEntry,
	respTransforms map[string]FieldMapperEntry,
	dummyReqTransforms map[string]FieldMapperEntry,
	h *PackageHelper,
) error {

	clientlessConverter := NewTypeConverter(h, nil)

	respType := funcSpec.ResultSpec.ReturnType

	clientlessConverter.append(
		"func convert",
		PascalCase(ms.Name),
		"DummyResponse(in ", ms.RequestType, ") ", ms.ResponseType, "{")

	structType := compile.FieldGroup(funcSpec.ArgsSpec)

	if respType == nil {
		return nil
	}

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

		// TODO: Add support for primitive type by mapping the first field from request to response
		return errors.Errorf(
			"clientless endpoints need a complex return type")
	default:
		// default as struct
		respFields := respType.(*compile.StructSpec).Fields
		clientlessConverter.append("out", " := ", "&", ms.ShortResponseType, "{}\t\n")
		err := clientlessConverter.GenStructConverter(structType, respFields, dummyReqTransforms)
		if err != nil {
			return err
		}

	}

	clientlessConverter.append("\nreturn out \t}")
	ms.ConvertClientlessRequestGoStatements = clientlessConverter.GetLines()

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
	case *compile.I64Spec:
		queryMethod = "GetQueryInt64"
	case *compile.DoubleSpec:
		queryMethod = "GetQueryFloat64"
	case *compile.StringSpec:
		queryMethod = "GetQueryValue"
	case *compile.EnumSpec:
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
	case
		*compile.I8Spec,
		*compile.I16Spec,
		*compile.I32Spec,
		*compile.I64Spec:
		encodeExpression = "strconv.FormatInt(int64(%s), 10)"
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
		encodeExpression = "(%s).String()"
	default:
		// This is intentional -- lets evaluate why we would want other types here before opening the flood gates
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

// hasQueryParams - checks to see if either this field has a query-param annotation
// or if this is a struct, some field in it has.
// Caveat is that unannotated fields are considered Query Params IF the REST method
// should not have a body (GET).  This is an existing convenience afforded to callers
func (ms *MethodSpec) hasQueryParams(field *compile.FieldSpec, defaultIsQuery bool) bool {

	httpRefAnnotation := field.Annotations[ms.annotations.HTTPRef]
	if strings.HasPrefix(httpRefAnnotation, queryAnnotationPrefix) {
		return true
	}
	// If it is a struct, recursively look to see if any of the fields are query params
	if container, ok := compile.RootTypeSpec(field.Type).(*compile.StructSpec); ok {
		visitor := func(goPrefix string, thriftPrefix string, field *compile.FieldSpec) bool {
			annotation := field.Annotations[ms.annotations.HTTPRef]
			if strings.HasPrefix(annotation, queryAnnotationPrefix) {
				return true
			}
			return annotation == "" && defaultIsQuery
		}
		return walkFieldGroups(container.Fields, visitor)
	}
	return httpRefAnnotation == "" && defaultIsQuery
}

// getContainedQueryParams - finds all query params of interest in this field
// In the case of structs, it recursively drills down
func (ms *MethodSpec) getContainedQueryParams(
	field *compile.FieldSpec, defaultIsQuery bool, defaultPrefix string) []string {
	rval := []string{}
	myDefaultParam := defaultPrefix + strings.ToLower(field.Name)
	annotation := field.Annotations[ms.annotations.HTTPRef]
	if strings.HasPrefix(annotation, queryAnnotationPrefix) {
		rval = append(rval, strings.TrimPrefix(annotation, queryAnnotationPrefix))
	} else if defaultIsQuery && annotation == "" {
		rval = append(rval, myDefaultParam)
	}
	// If it is a struct, look to see if any of the fields are query params
	if container, ok := compile.RootTypeSpec(field.Type).(*compile.StructSpec); ok {
		for _, subField := range container.Fields {
			rval = append(rval, ms.getContainedQueryParams(subField, defaultIsQuery, myDefaultParam+".")...)
		}
	}
	return rval
}

func (ms *MethodSpec) setWriteQueryParamStatements(
	funcSpec *compile.FunctionSpec, packageHelper *PackageHelper, hasNoBody bool,
) error {
	var statements LineBuilder
	var hasQueryFields bool
	var stack []string
	isVoidReturn := funcSpec.ResultSpec.ReturnType == nil

	visitor := func(
		goPrefix string, thriftPrefix string, field *compile.FieldSpec,
	) bool {
		// Skip if there are no query params in the field or its components
		if !ms.hasQueryParams(field, hasNoBody) {
			return false
		}

		realType := compile.RootTypeSpec(field.Type)
		longFieldName := goPrefix + "." + PascalCase(field.Name)

		if len(stack) > 0 {
			if !strings.HasPrefix(longFieldName, stack[len(stack)-1]) {
				stack = stack[:len(stack)-1]
				statements.append("}")
			}
		}
		if _, ok := realType.(*compile.StructSpec); ok {
			// If a field is a struct we need to look inside
			if field.Required {
				statements.appendf("if r%s == nil {", longFieldName)
				// Generate correct number of nils...
				if isVoidReturn {
					statements.append("\treturn nil, errors.New(")
				} else {
					statements.append("\treturn nil, nil, errors.New(")
				}
				statements.appendf("\t\t\"The field %s is required\",",
					longFieldName,
				)
				statements.append("\t)")
				statements.append("}")
			} else {
				stack = append(stack, longFieldName)

				statements.appendf("if r%s != nil {", longFieldName)
			}

			return false
		}

		longQueryName, shortQueryParam := ms.getQueryParamInfo(field, thriftPrefix)
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
				statements.appendf("\tqueryValues.Add(\"%s\", %s)", shortQueryParam, encodeExpr)
				statements.append("}")
			} else if isSet {
				encodeExpr := getQueryEncodeExpression(field.Type, "value")
				statements.appendf("for value := range %s {", "r"+longFieldName)
				statements.appendf("\tqueryValues.Add(\"%s\", %s)", shortQueryParam, encodeExpr)
				statements.append("}")
			} else {
				encodeExpr := getQueryEncodeExpression(field.Type, "r"+longFieldName)
				statements.appendf("%s := %s", identifierName, encodeExpr)
				statements.appendf("queryValues.Set(\"%s\", %s)", shortQueryParam, identifierName)
			}
		} else {
			statements.appendf("if r%s != nil {", longFieldName)
			if isList {
				encodeExpr := getQueryEncodeExpression(field.Type, "value")
				statements.appendf("for _, value := range %s {", "r"+longFieldName)
				statements.appendf("\tqueryValues.Add(\"%s\", %s)", shortQueryParam, encodeExpr)
				statements.append("}")
			} else if isSet {
				encodeExpr := getQueryEncodeExpression(field.Type, "value")
				statements.appendf("for value := range %s {", "r"+longFieldName)
				statements.appendf("\tqueryValues.Add(\"%s\", %s)", shortQueryParam, encodeExpr)
				statements.append("}")
			} else {
				encodeExpr := getQueryEncodeExpression(field.Type, "*r"+longFieldName)
				statements.appendf("\t%s := %s", identifierName, encodeExpr)
				statements.appendf("\tqueryValues.Set(\"%s\", %s)", shortQueryParam, identifierName)
			}
			statements.append("}")
		}
		return false
	}
	walkFieldGroups(compile.FieldGroup(funcSpec.ArgsSpec), visitor)

	for i := 0; i < len(stack)-1; i++ {
		statements.append("}")
	}

	if hasQueryFields {
		statements.append("fullURL += \"?\" + queryValues.Encode()")
	}
	if len(stack) > 0 {
		statements.append("}")
	}

	ms.WriteQueryParamGoStatements = statements.GetLines()
	return nil
}

// makeUniqIdent appends an integer to the identifier name if there is duplication already
// The reason for this is to disambiguate a query param "deviceID" from "device_ID" - yes people did do that
func makeUniqIdent(identifier string, seen map[string]int) string {
	count := seen[identifier]
	seen[identifier] = count + 1
	if count > 0 {
		return identifier + strconv.Itoa(count)
	}
	return identifier
}

func getCustomType(pkgHelper *PackageHelper, itemType compile.TypeSpec) (string, error) {
	switch itemType.(type) {
	case
		*compile.TypedefSpec,
		*compile.EnumSpec:
		return GoType(pkgHelper, itemType)
	}
	return "", nil
}

func (ms *MethodSpec) setParseQueryParamStatements(
	funcSpec *compile.FunctionSpec, packageHelper *PackageHelper, hasNoBody bool,
) error {
	// If a thrift field has a http.ref annotation then we
	// should not read this field from query parameters.
	var statements LineBuilder

	var finalError error
	var stack = []string{}
	seenIdents := map[string]int{}

	visitor := func(
		goPrefix string, thriftPrefix string, field *compile.FieldSpec,
	) bool {
		realType := compile.RootTypeSpec(field.Type)
		longFieldName := goPrefix + "." + PascalCase(field.Name)
		longQueryName, shortQueryParam := ms.getQueryParamInfo(field, thriftPrefix)

		// Skip if there are no query params in the field or its components
		if !ms.hasQueryParams(field, hasNoBody) {
			return false
		}

		if len(stack) > 0 {
			if !strings.HasPrefix(longFieldName, stack[len(stack)-1]) {
				stack = stack[:len(stack)-1]
				statements.append("}")
			}
		}

		customType, err := getCustomType(packageHelper, field.Type)
		if err != nil {
			finalError = err
			return true
		}

		var isList, isSet bool
		var customElemType string
		var isEnumElem bool
		switch t := realType.(type) {
		// Before you ask -- yes duplicated code because ValueSpec is not defined in the generic interface
		case *compile.ListSpec:
			isList = true
			customElemType, err = getCustomType(packageHelper, t.ValueSpec)
			if err != nil {
				finalError = err
				return true
			}
			_, isEnumElem = t.ValueSpec.(*compile.EnumSpec)
		case *compile.SetSpec:
			isSet = true
			customElemType, err = getCustomType(packageHelper, t.ValueSpec)
			if err != nil {
				finalError = err
				return true
			}
			_, isEnumElem = t.ValueSpec.(*compile.EnumSpec)
		case *compile.StructSpec:
			typeName, err := GoType(packageHelper, realType)
			if err != nil {
				finalError = err
				return true
			}

			if !field.Required {
				stack = append(stack, longFieldName)
				applicableQueryParams := ms.getContainedQueryParams(field, hasNoBody, "")

				statements.append("var _queryNeeded bool")
				statements.appendf("for _, _pfx := range %#v {", applicableQueryParams)
				statements.append("if _queryNeeded = req.HasQueryPrefix(_pfx); _queryNeeded {")
				statements.append("break")
				statements.append("}")
				statements.append("}")
				statements.append("if _queryNeeded {")
			}

			statements.appendf("if requestBody%s == nil {", longFieldName)
			statements.appendf("requestBody%s = &%s{}", longFieldName, typeName)
			statements.append("}")

			return false
		}
		isAggregate := isList || isSet // we do not support maps

		// For disambiguation of similar names
		baseIdent := makeUniqIdent(CamelCase(longQueryName), seenIdents)
		identifierName := baseIdent + "Query"
		okIdentifierName := baseIdent + "Ok"

		// make sure value is present
		if field.Required {
			statements.appendf("%s := req.CheckQueryValue(%q)", okIdentifierName, shortQueryParam)
			statements.appendf("if !%s {", okIdentifierName)
			statements.append("return")
			statements.append("}")
		} else {
			statements.appendf("%s := req.HasQueryValue(%q)", okIdentifierName, shortQueryParam)
			statements.appendf("if %s {", okIdentifierName)
		}

		queryRValue := fmt.Sprintf("req.%s(%q)", getQueryMethodForType(realType), shortQueryParam)

		// Transform if enum
		if _, isEnumType := field.Type.(*compile.EnumSpec); isEnumType {
			statements.appendf("var %s %s", identifierName, customType)
			tmpVar := "_tmp" + identifierName
			statements.appendf("%s, ok := %s", tmpVar, queryRValue)
			statements.append("if ok {")
			statements.appendf("if err := %s.UnmarshalText([]byte(%s)); err != nil {",
				identifierName, tmpVar)
			statements.appendf("req.LogAndSendQueryError(err, %q, %q, %s)",
				"enum", shortQueryParam, tmpVar)
			statements.append("ok = false")
			statements.append("}")
			statements.append("}")
		} else {
			statements.appendf("%s, ok := %s", identifierName, queryRValue)
		}

		statements.append("if !ok {")
		statements.append("return")
		statements.append("}")

		target := identifierName

		// If field is an "aggregate" with custom element types, we need to convert them first
		// Note that enums and typedefs are what get in here
		if customElemType != "" {
			target += "Final"
			valVar := "v"
			if isList {
				statements.appendf(
					"%s := make([]%s, len(%s))",
					target, customElemType, identifierName,
				)
				statements.appendf("for i, %s := range %s {", valVar, identifierName)
				if isEnumElem {
					tmpVar := "_tmp" + valVar
					statements.appendf("var %s %s", tmpVar, customElemType)
					statements.appendf("if err := %s.UnmarshalText([]byte(%s)); err != nil {",
						tmpVar, valVar)
					statements.appendf("req.LogAndSendQueryError(err, %q, %q, %s)",
						"enum", shortQueryParam, valVar)
					statements.append("return")
					statements.append("}")
					valVar = tmpVar
				}
				statements.appendf("%s[i] = %s(%s)", target, customElemType, valVar)
				statements.append("}")
			} else if isSet {
				statements.appendf(
					"%s := make(map[%s]struct{}, len(%s))",
					target, customElemType, identifierName,
				)
				statements.appendf("for %s := range %s {", valVar, identifierName)
				if isEnumElem {
					tmpVar := "_tmp" + valVar
					statements.appendf("var %s %s", tmpVar, customElemType)
					statements.appendf("if err := %s.UnmarshalText([]byte(%s)); err != nil {",
						tmpVar, valVar)
					statements.appendf("req.LogAndSendQueryError(err, %q, %q, %s)",
						"enum", shortQueryParam, valVar)
					statements.append("return")
					statements.append("}")
					valVar = tmpVar
				}
				statements.appendf("%s[%s(%s)] = struct{}{}", target, customElemType, valVar)
				statements.append("}")
			}
		}

		var deref string
		if !field.Required && !isAggregate {
			deref = "*"
			targetName := identifierName
			if customType != "" {
				targetName = fmt.Sprintf("%s(%s)", strings.ToLower(pointerMethodType(realType)), targetName)
			}
			target = fmt.Sprintf("ptr.%s(%s)", pointerMethodType(realType), targetName)
		}
		if customType != "" {
			target = fmt.Sprintf("(%s%s)(%s)", deref, customType, target)
		}
		statements.appendf("requestBody%s = %s", longFieldName, target)

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

// getQueryParamInfo -- returns the fully-qualified query name and the query param
// The query param is what is specified in the annotation if present, otherwise it is the same as the long query name
func (ms *MethodSpec) getQueryParamInfo(field *compile.FieldSpec, thriftPrefix string) (string, string) {
	var longQueryName, queryParam string

	queryName := field.Name
	queryAnnotation := field.Annotations[ms.annotations.HTTPRef]
	if strings.HasPrefix(queryAnnotation, queryAnnotationPrefix) {
		queryName = strings.TrimPrefix(queryAnnotation, queryAnnotationPrefix)
		queryParam = queryName
	}
	longQueryName = strings.TrimPrefix(thriftPrefix+"."+queryName, ".")
	// default the short query param to the fully qualified long path
	if queryParam == "" {
		queryParam = longQueryName
	}
	return longQueryName, queryParam
}

func (ms *MethodSpec) isRequestBoxed(f *compile.FunctionSpec) bool {
	boxed, ok := f.Annotations[ms.annotations.HTTPReqDefBoxed]
	return ok && boxed == "true"
}

func (ms *MethodSpec) isBodyDisallowed(f *compile.FieldSpec) bool {
	val, ok := f.Annotations[ms.annotations.HTTPResNoBody]
	return ok && val == "true"
}

func headers(annotation string) []string {
	if annotation == "" {
		return nil
	}
	return strings.Split(annotation, ",")
}
