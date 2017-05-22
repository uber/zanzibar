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
	// The root directory containing thrift files.
	thriftRootDir string
	// Namespace under thrift folder
	gatewayNamespace string
	// The go package name of where all the generated structs are
	genCodePackage string
	// The directory to put the generated service code.
	targetGenDir string
	// The root directory for the gateway test config files.
	testConfigsRootDir string
	// String containing copyright header to add to generated code.
	copyrightHeader string
	// The middlewares available for the endpoints
	middlewareSpecs map[string]*MiddlewareSpec
}

// NewPackageHelper creates a package helper.
func NewPackageHelper(
	packageRoot string,
	configDirName string,
	middlewareConfig string,
	thriftRootDir string,
	genCodePackage string,
	targetGenDir string,
	gatewayNamespace string,
	copyrightHeader string,
) (*PackageHelper, error) {
	genDir, err := filepath.Abs(targetGenDir)
	if err != nil {
		return nil, errors.Errorf("%s is not valid path: %s", targetGenDir, err)
	}

	genDirIndex := strings.Index(genDir, gatewayNamespace)
	if genDirIndex == -1 {
		return nil, errors.Errorf(
			"gatewayNamespace (%s) must be inside targetGenDir (%s)",
			gatewayNamespace,
			genDir,
		)
	}

	middlewareSpecs := map[string]*MiddlewareSpec{}
	middleConfig := filepath.Join(
		configDirName,
		middlewareConfig,
	)
	middlewares, err := parseMiddlewareConfig(middleConfig, configDirName)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Cannot load middlewares:")
	}

	for _, mspec := range middlewares {
		middlewareSpecs[mspec.Name] = mspec
	}

	p := &PackageHelper{
		packageRoot:      packageRoot,
		thriftRootDir:    path.Clean(thriftRootDir),
		genCodePackage:   genCodePackage,
		gatewayNamespace: gatewayNamespace,
		targetGenDir:     genDir,
		copyrightHeader:  copyrightHeader,
		middlewareSpecs:  middlewareSpecs,
	}
	return p, nil
}

// PackageRoot returns the service's root package name
func (p PackageHelper) PackageRoot() string {
	return p.packageRoot
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
	nsIndex := strings.Index(p.targetGenDir, p.gatewayNamespace)
	return path.Join(
		p.gatewayNamespace,
		p.targetGenDir[nsIndex+len(p.gatewayNamespace):],
	)
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

// PackageGenPath returns the Go package path for generated code from a thrift file.
func (p PackageHelper) PackageGenPath(thrift string) (string, error) {
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

	nsIndex := strings.Index(p.targetGenDir, p.gatewayNamespace)
	return path.Join(
		p.gatewayNamespace,
		p.targetGenDir[nsIndex+len(p.gatewayNamespace):],
		filepath.Dir(thrift[idx+len(p.thriftRootDir):]),
	), nil
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

// TypeFullName returns the referred Go type name in generated code from curThriftFile.
func (p PackageHelper) TypeFullName(curThriftFile string, typeSpec compile.TypeSpec) (string, error) {
	if typeSpec == nil {
		return "", nil
	}
	tfile := typeSpec.ThriftFile()

	if tfile == "" {
		return typeSpec.ThriftName(), nil
	}

	pkg, err := p.TypePackageName(tfile)
	if err != nil {
		return "", errors.Wrap(err, "failed to get the full type")
	}
	return pkg + "." + typeSpec.ThriftName(), nil
}
