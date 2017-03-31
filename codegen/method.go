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

// MethodSpec specifies all needed parts to generate code for a method in service.
type MethodSpec struct {
	Name       string
	HTTPMethod string
	// Used by edge gateway to generate endpoint.
	EndpointName string
	HTTPPath     string
	PathSegments []PathSegment
	// Headers needed, generated from "zanzibar.http.headers"
	Headers             []string
	RequestType         string
	ResponseType        string
	OKStatusCode        []StatusCode
	ExceptionStatusCode []StatusCode
	// Additional struct generated from the bundle of request args.
	RequestBoxed  bool
	RequestStruct []StructSpec
	// The triftrw compiled spec, used to extract type information
	CompiledThriftSpec *compile.FunctionSpec
	// The downstream service method set by endpoint config
	Downstream *ModuleSpec
	// the downstream service name
	DownstreamService string
	// The downstream methdo spec for the endpoint
	DownstreamMethod *MethodSpec
	// A map from upstream to downstream field names in the requests.
	RequestFieldMap map[string]string
	// A map from upstream to field names to downstream types in the requests.
	RequestTypeMap map[string]string
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
	antHTTPHeaders     = "zanzibar.http.headers"
	antHTTPRef         = "zanzibar.http.ref"
	antMeta            = "zanzibar.meta"
	antHandler         = "zanzibar.handler"
)

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
		ms.RequestType, err = packageHelper.TypeFullName(curThriftFile, funcSpec.ArgsSpec[0].Type)
	} else {
		ms.RequestBoxed = false
		ms.RequestType, err = ms.newRequestType(curThriftFile, funcSpec, packageHelper)
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

func (ms *MethodSpec) newRequestType(curThriftFile string, f *compile.FunctionSpec, h *PackageHelper) (string, error) {
	requestType := strings.Title(f.Name) + "HTTPRequest"
	ms.RequestStruct = make([]StructSpec, len(f.ArgsSpec))
	for i, arg := range f.ArgsSpec {
		typeName, err := h.TypeFullName(curThriftFile, arg.Type)
		if err != nil {
			return "", errors.Wrap(err, "failed to generate new request type")
		}
		if isStructType(arg.Type) {
			typeName = "*" + typeName
		}
		ms.RequestStruct[i] = StructSpec{
			Type:        typeName,
			Name:        arg.Name,
			Annotations: arg.Annotations,
		}
	}
	return requestType, nil
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
	scode := strings.Split(statusCode, ",")
	ms.OKStatusCode = make([]StatusCode, len(scode))
	var err error
	for i, c := range scode {
		ms.OKStatusCode[i].Code, err = strconv.Atoi(c)
		if err != nil {
			return errors.Wrapf(err, "failed to parse the annotation %s for ok response status")
		}
	}
	return nil
}

func (ms *MethodSpec) setExceptionStatusCode(resultSpec *compile.ResultSpec) error {
	ms.ExceptionStatusCode = make([]StatusCode, len(resultSpec.Exceptions))
	for i, e := range resultSpec.Exceptions {
		code, err := strconv.Atoi(e.Annotations[antHTTPStatus])
		if err != nil {
			return errors.Wrapf(err, "cannot parse the annotation %s for exception %s", antHTTPStatus, e.Name)
		}
		ms.ExceptionStatusCode[i] = StatusCode{
			Code:    code,
			Message: e.Name,
		}
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
			"Downstream method (%s) is not found: %s",
			clientModule.ThriftFile, clientMethod,
		)
	}
	// Remove irrelevant services and methods.
	ms.Downstream = clientModule
	ms.DownstreamService = clientService
	ms.DownstreamMethod = downstreamMethod
	return nil
}

func (ms *MethodSpec) setRequestFieldMap(
	funcSpec *compile.FunctionSpec,
	downstreamSpec *compile.FunctionSpec,
) error {
	// TODO(sindelar): Iterate over fields that are structs (for foo/bar examples).
	ms.RequestFieldMap = map[string]string{}
	ms.RequestTypeMap = map[string]string{}

	structType := compile.FieldGroup(funcSpec.ArgsSpec)

	for i := 0; i < len(structType); i++ {
		field := structType[i]
		// Add type checking and conversion, custom mapping
		downstreamStructType := compile.FieldGroup(downstreamSpec.ArgsSpec)
		var downstreamField *compile.FieldSpec
		for j := 0; j < len(downstreamStructType); j++ {
			if downstreamStructType[j].Name == field.Name {
				downstreamField = downstreamStructType[j]
				break
			}
		}

		if downstreamField == nil {
			return errors.Errorf(
				"cannot map by name for the field %s to type: %s",
				field.Name, downstreamSpec.Name,
			)
		}
		ms.RequestFieldMap[field.Name] = field.Name

		// Override thrift type names to avoid naming collisions between endpoint
		// and client types.
		switch field.Type.(type) {
		case *compile.BoolSpec, *compile.I8Spec, *compile.I16Spec, *compile.I32Spec,
			*compile.I64Spec, *compile.DoubleSpec, *compile.StringSpec:
			ms.RequestTypeMap[field.Name] = field.Type.ThriftName()
		default:
			thriftPkgNameParts := strings.Split(field.Type.ThriftFile(), "/")
			thriftPkgName := thriftPkgNameParts[len(thriftPkgNameParts)-2]
			ms.RequestTypeMap[field.Name] = "(*clientType" + strings.Title(thriftPkgName) + "." + field.Type.ThriftName() + ")"
		}
	}
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
