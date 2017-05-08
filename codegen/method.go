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

// MethodSpec specifies all needed parts to generate code for a method in service.
type MethodSpec struct {
	Name       string
	HTTPMethod string
	// Used by edge gateway to generate endpoint.
	EndpointName string
	HTTPPath     string
	PathSegments []PathSegment
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
	ConvertRequestLines []string
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
	thriftService string,
) (*MethodSpec, error) {
	method := &MethodSpec{}
	method.CompiledThriftSpec = funcSpec
	var err error
	var ok bool
	method.Name = funcSpec.MethodName()
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

	err = method.setExceptions(thriftFile, funcSpec.ResultSpec, packageHelper)
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
		return nil, errors.Errorf(
			"invalid annotation: HTTP GET method with body type",
		)
	}

	var httpPath string
	if httpPath, ok = funcSpec.Annotations[antHTTPPath]; !ok {
		return nil, errors.Errorf(
			"missing anotation '%s' for HTTP path", antHTTPPath,
		)
	}
	method.setHTTPPath(httpPath, funcSpec)

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
		ms.RequestType, err = packageHelper.TypeFullName(
			curThriftFile, funcSpec.ArgsSpec[0].Type,
		)
	} else {
		ms.RequestBoxed = false

		goPackageName, err := packageHelper.TypePackageName(curThriftFile)
		if err == nil {
			ms.RequestType = goPackageName + "." +
				ms.ThriftService + "_" + strings.Title(ms.Name) + "_Args"
		}
	}
	if err != nil {
		return errors.Wrap(err, "failed to set request type")
	}
	return nil
}

func isStructType(spec compile.TypeSpec) bool {
	spec = compile.RootTypeSpec(spec)
	_, isStruct := spec.(*compile.StructSpec)
	return isStruct
}

func (ms *MethodSpec) setResponseType(curThriftFile string, respSpec *compile.ResultSpec, packageHelper *PackageHelper) error {
	if respSpec == nil {
		ms.ResponseType = ""
		return nil
	}
	typeName, err := packageHelper.TypeFullName(curThriftFile, respSpec.ReturnType)
	if err != nil {
		return errors.Wrap(err, "failed to get response type")
	}
	ms.ResponseType = typeName
	return nil
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
		typeName, err := h.TypeFullName(curThriftFile, e.Type)
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

		if seenStatusCodes[code] {
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
	fields compile.FieldGroup, paramName string, prefix string,
) (string, bool) {
	for i := 0; i < len(fields); i++ {
		field := fields[i]

		if param, ok := field.Annotations[antHTTPRef]; ok {
			if param == "params."+paramName[1:] {
				return prefix + field.Name, true
			}
		}

		switch t := field.Type.(type) {
		case *compile.BinarySpec:
		case *compile.StringSpec:
		case *compile.BoolSpec:
		case *compile.DoubleSpec:
		case *compile.I8Spec:
		case *compile.I16Spec:
		case *compile.I32Spec:
		case *compile.I64Spec:
		case *compile.StructSpec:
			path, ok := findParamsAnnotation(
				t.Fields, paramName, field.Name+".",
			)
			if ok {
				return path, true
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

	return "", false
}

func (ms *MethodSpec) setHTTPPath(
	httpPath string, funcSpec *compile.FunctionSpec,
) {
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
					structType.Fields, segment, "",
				)
			} else {
				fieldSelect, ok = findParamsAnnotation(
					compile.FieldGroup(funcSpec.ArgsSpec), segment, "",
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
	clientModule *ModuleSpec, clientService string, clientMethod string,
) error {
	var downstreamService *ServiceSpec
	for _, service := range clientModule.Services {
		if service.Name == clientService {
			downstreamService = service
			break
		}
	}
	if downstreamService == nil {
		return errors.Errorf(
			"Downstream service '%s' is not found in '%s'",
			clientService, clientModule.ThriftFile,
		)
	}
	var downstreamMethod *MethodSpec
	for _, method := range downstreamService.Methods {
		if method.Name == clientMethod {
			downstreamMethod = method
			break
		}
	}
	if downstreamMethod == nil {
		return errors.Errorf(
			"\n Downstream method '%s' is not found in '%s'",
			clientMethod, clientModule.ThriftFile,
		)
	}
	// Remove irrelevant services and methods.
	ms.Downstream = clientModule
	ms.DownstreamService = clientService
	ms.DownstreamMethod = downstreamMethod
	return nil
}

func createTypeConverter(
	fromFields []*compile.FieldSpec,
	toFields []*compile.FieldSpec,
	lines []string,
	h *PackageHelper,
) ([]string, error) {
	for i := 0; i < len(toFields); i++ {
		toField := toFields[i]

		var fromField *compile.FieldSpec
		for j := 0; j < len(fromFields); j++ {
			if fromFields[j].Name == toField.Name {
				fromField = fromFields[j]
				break
			}
		}

		if fromField == nil {
			return nil, errors.Errorf(
				"cannot map by name for the field %s",
				toField.Name,
			)
		}

		line := "out." + strings.Title(toField.Name) + " = "

		// Override thrift type names to avoid naming collisions between endpoint
		// and client types.
		switch toField.Type.(type) {
		case *compile.BoolSpec:
			line += "bool"
		case *compile.I8Spec:
			line += "int8"
		case *compile.I16Spec:
			line += "int16"
		case *compile.I32Spec:
			line += "int32"
		case *compile.I64Spec:
			line += "int64"
		case *compile.DoubleSpec:
			line += "float64"
		case *compile.StringSpec:
			line += "string"
		case *compile.BinarySpec:
			line += "[]byte"
		default:
			pkgName, err := h.TypePackageName(toField.Type.ThriftFile())
			if err != nil {
				return nil, err
			}
			line += "(*" + pkgName + "." + toField.Type.ThriftName() + ")"
		}

		line += "(in." + strings.Title(fromField.Name) + ")"

		lines = append(lines, line)
	}

	return lines, nil
}

func (ms *MethodSpec) setRequestFieldMap(
	funcSpec *compile.FunctionSpec,
	downstreamSpec *compile.FunctionSpec,
	h *PackageHelper,
) error {
	// TODO(sindelar): Iterate over fields that are structs (for foo/bar examples).
	lines := []string{}

	// Add type checking and conversion, custom mapping
	structType := compile.FieldGroup(funcSpec.ArgsSpec)
	downstreamStructType := compile.FieldGroup(downstreamSpec.ArgsSpec)

	lines, err := createTypeConverter(
		structType, downstreamStructType, lines, h,
	)
	if err != nil {
		return err
	}

	ms.ConvertRequestLines = lines
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
