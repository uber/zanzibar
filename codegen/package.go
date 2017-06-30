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
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

// PackageHelper manages the mapping from thrift file to generated type code and service code.
type PackageHelper struct {
	// The project package name
	packageRoot string
	// The absolute root dir path for all configs, i.e., clients, endpoints, idl, etc
	configRoot string
	// The absolute root directory containing thrift files
	thriftRootDir string
	// The go package name of where all the generated structs are
	genCodePackage string
	// The absolute directory to put the generated service code
	targetGenDir string
	// The go package name where all the generated code is
	goGatewayNamespace string
	// The root directory for the gateway test config files
	testConfigsRootDir string
	// String containing copyright header to add to generated code.
	copyrightHeader string
	// The middlewares available for the endpoints
	middlewareSpecs map[string]*MiddlewareSpec
}

// NewPackageHelper creates a package helper.
func NewPackageHelper(
	packageRoot string,
	configRoot string,
	middlewareConfig string,
	relThriftRootDir string,
	genCodePackage string,
	relTargetGenDir string,
	copyrightHeader string,
) (*PackageHelper, error) {
	absConfigRoot, err := filepath.Abs(configRoot)
	if err != nil {
		return nil, errors.Errorf(
			"%s is not valid path: %s", configRoot, err,
		)
	}

	goGatewayNamespace := path.Join(packageRoot, relTargetGenDir)

	middlewareSpecs, err := parseMiddlewareConfig(middlewareConfig, absConfigRoot)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Cannot load middlewares:")
	}

	p := &PackageHelper{
		packageRoot:        packageRoot,
		configRoot:         absConfigRoot,
		thriftRootDir:      filepath.Join(absConfigRoot, relThriftRootDir),
		genCodePackage:     genCodePackage,
		goGatewayNamespace: goGatewayNamespace,
		targetGenDir:       filepath.Join(absConfigRoot, relTargetGenDir),
		copyrightHeader:    copyrightHeader,
		middlewareSpecs:    middlewareSpecs,
	}
	return p, nil
}

// PackageRoot returns the service's root package name
func (p PackageHelper) PackageRoot() string {
	return p.packageRoot
}

// ConfigRoot returns the service's absolute config root path
func (p PackageHelper) ConfigRoot() string {
	return p.configRoot
}

// MiddlewareSpecs returns a map of middlewares available
func (p PackageHelper) MiddlewareSpecs() map[string]*MiddlewareSpec {
	return p.middlewareSpecs
}

// TypeImportPath returns the Go import path for types defined in a thrift file.
func (p PackageHelper) TypeImportPath(thrift string) (string, error) {
	if !strings.HasSuffix(thrift, ".thrift") {
		return "", errors.Errorf("file %s is not .thrift", thrift)
	}

	idx := strings.Index(thrift, p.thriftRootDir)
	if idx == -1 {
		return "", errors.Errorf(
			"file %s is not in thrift dir (%s)",
			thrift, p.thriftRootDir,
		)
	}
	return path.Join(
		p.genCodePackage,
		thrift[idx+len(p.thriftRootDir):len(thrift)-7],
	), nil
}

// GoGatewayPackageName returns the name of the gateway package
func (p PackageHelper) GoGatewayPackageName() string {
	return p.goGatewayNamespace
}

// ThriftIDLPath returns the file path to the thrift idl folder
func (p PackageHelper) ThriftIDLPath() string {
	return p.thriftRootDir
}

// CodeGenTargetPath returns the file path where the code should
// be generated.
func (p PackageHelper) CodeGenTargetPath() string {
	return p.targetGenDir
}

// TypePackageName returns the package name that defines the type.
func (p PackageHelper) TypePackageName(thrift string) (string, error) {
	if !strings.HasSuffix(thrift, ".thrift") {
		return "", errors.Errorf("file %s is not .thrift", thrift)
	}
	idx := strings.Index(thrift, p.thriftRootDir)
	if idx == -1 {
		return "", errors.Errorf(
			"file %s is not in thrift dir (%s)",
			thrift, p.thriftRootDir,
		)
	}

	// Strip the leading / and strip the .thrift on the end.
	thriftSegment := thrift[idx+len(p.thriftRootDir)+1 : len(thrift)-7]

	thriftPackageName := strings.Replace(thriftSegment, "/", "_", -1)
	return camelCase(thriftPackageName), nil
}

func (p PackageHelper) getRelativeFileName(thrift string) (string, error) {
	if !strings.HasSuffix(thrift, ".thrift") {
		return "", errors.Errorf("file %s is not .thrift", thrift)
	}
	idx := strings.Index(thrift, p.thriftRootDir)
	if idx == -1 {
		return "", errors.Errorf(
			"file %s is not in thrift dir (%s)",
			thrift, p.thriftRootDir,
		)
	}
	return thrift[idx+len(p.thriftRootDir):], nil
}

// TargetClientsInitPath returns where the clients init should go
func (p PackageHelper) TargetClientsInitPath() string {
	return path.Join(p.targetGenDir, "clients", "clients.go")
}

// TargetEndpointsRegisterPath returns where the endpoints register file
// should be written to
func (p PackageHelper) TargetEndpointsRegisterPath() string {
	return path.Join(p.targetGenDir, "endpoints", "register.go")
}

// EndpointTestConfigPath returns the path for the endpoint test configs
func (p PackageHelper) EndpointTestConfigPath(
	serviceName, methodName string,
) string {
	fileName := strings.ToLower(methodName) + "_test.json"
	return path.Join(p.testConfigsRootDir, strings.ToLower(serviceName), fileName)
}

// TypeFullName returns the referred Go type name in generated code.
func (p PackageHelper) TypeFullName(typeSpec compile.TypeSpec) (string, error) {
	if typeSpec == nil {
		return "", nil
	}
	return goType(p, typeSpec)
}

func goType(p PackageHelper, spec compile.TypeSpec) (string, error) {
	switch s := spec.(type) {
	case *compile.BoolSpec:
		return "bool", nil
	case *compile.I8Spec:
		return "int8", nil
	case *compile.I16Spec:
		return "int16", nil
	case *compile.I32Spec:
		return "int32", nil
	case *compile.I64Spec:
		return "int64", nil
	case *compile.DoubleSpec:
		return "float64", nil
	case *compile.StringSpec:
		return "string", nil
	case *compile.BinarySpec:
		return "[]byte", nil
	case *compile.MapSpec:
		k, err := goReferenceType(p, s.KeySpec)
		if err != nil {
			return "", err
		}
		v, err := goReferenceType(p, s.ValueSpec)
		if err != nil {
			return "", err
		}
		if !isHashable(s.KeySpec) {
			return fmt.Sprintf("[]struct{Key %s; Value %s}", k, v), nil
		}
		return fmt.Sprintf("map[%s]%s", k, v), nil
	case *compile.ListSpec:
		v, err := goReferenceType(p, s.ValueSpec)
		if err != nil {
			return "", err
		}
		return "[]" + v, nil
	case *compile.SetSpec:
		v, err := goReferenceType(p, s.ValueSpec)
		if err != nil {
			return "", err
		}
		if !isHashable(s.ValueSpec) {
			return fmt.Sprintf("[]%s", v), nil
		}
		return fmt.Sprintf("map[%s]struct{}", v), nil
	case *compile.EnumSpec, *compile.StructSpec, *compile.TypedefSpec:
		return goCustomType(p, spec)
	default:
		panic(fmt.Sprintf("Unknown type (%T) %v", spec, spec))
	}
}

func goReferenceType(p PackageHelper, spec compile.TypeSpec) (string, error) {
	t, err := goType(p, spec)
	if err != nil {
		return "", err
	}

	if isStructType(spec) {
		t = "*" + t
	}

	return t, nil
}

func goCustomType(p PackageHelper, spec compile.TypeSpec) (string, error) {
	f := spec.ThriftFile()
	if f == "" {
		return "", fmt.Errorf("goCustomType called with native type (%T) %v", spec, spec)
	}

	pkg, err := p.TypePackageName(f)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get package for custom type (%T) %v", spec, spec)
	}

	return pkg + "." + pascalCase(spec.ThriftName()), nil
}

func isStructType(spec compile.TypeSpec) bool {
	spec = compile.RootTypeSpec(spec)
	_, isStruct := spec.(*compile.StructSpec)
	return isStruct
}

// isHashable returns true if the given type is considered hashable by thriftrw.
//
// Only primitive types, enums, and typedefs of other hashable types are considered hashable.
func isHashable(t compile.TypeSpec) bool {
	return isPrimitiveType(t)
}

// isPrimitiveType returns true if the given type is a primitive type.
// Primitive types, enums, and typedefs of primitive types are considered primitive.
//
// binary is not considered a primitive type because it is represented as []byte in Go.
func isPrimitiveType(spec compile.TypeSpec) bool {
	spec = compile.RootTypeSpec(spec)
	switch spec.(type) {
	case *compile.BoolSpec, *compile.I8Spec, *compile.I16Spec, *compile.I32Spec,
		*compile.I64Spec, *compile.DoubleSpec, *compile.StringSpec:
		return true
	}

	_, isEnum := spec.(*compile.EnumSpec)
	return isEnum
}
