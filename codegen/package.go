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
	// The root directory containing thrift files.
	thriftRootDir string
	// The root directory just for the gateway thrift files.
	gatewayThriftRootDir string
	// Namespace under thrift folder
	gatewayThriftNamespace string
	// The root directory where all files of go types are generated.
	typeFileRootDir string
	// The directory to put the generated service code.
	targetGenDir string
}

// NewPackageHelper creates a package helper.
func NewPackageHelper(
	thriftRootDir string,
	typeFileRootDir string,
	targetGenDir string,
	gatewayThriftRootDir string,
) (*PackageHelper, error) {
	genDir, err := filepath.Abs(targetGenDir)
	if err != nil {
		return nil, errors.Errorf("%s is not valid path: %s", targetGenDir, err)
	}

	gatewayThriftRootDir = path.Clean(gatewayThriftRootDir)
	idlIndex := strings.Index(gatewayThriftRootDir, "idl/") + 4
	gatewayThriftNamespace := gatewayThriftRootDir[idlIndex:]

	genDirIndex := strings.Index(genDir, gatewayThriftNamespace)
	if genDirIndex == -1 {
		return nil, errors.Errorf(
			"gatewayThriftNamespace (%s) must be inside targetGenDir (%s)",
			gatewayThriftNamespace,
			genDir,
		)
	}

	return &PackageHelper{
		thriftRootDir:          path.Clean(thriftRootDir),
		typeFileRootDir:        typeFileRootDir,
		gatewayThriftRootDir:   gatewayThriftRootDir,
		gatewayThriftNamespace: gatewayThriftNamespace,
		targetGenDir:           genDir,
	}, nil
}

// TypeImportPath returns the Go import path for types defined in a thrift file.
func (p PackageHelper) TypeImportPath(thrift string) (string, error) {
	if !strings.HasSuffix(thrift, ".thrift") {
		return "", errors.Errorf("file %s is not .thrift", thrift)
	}
	idx := strings.Index(thrift, p.thriftRootDir)
	if idx == -1 {
		return "", errors.Errorf("file %s is not in thrift dir", thrift)
	}
	return path.Join(
		p.gatewayThriftNamespace,
		p.typeFileRootDir,
		thrift[idx+len(p.thriftRootDir):len(thrift)-7],
	), nil
}

// GoGatewayPackageName returns the name of the gateway package
func (p PackageHelper) GoGatewayPackageName() string {
	nsIndex := strings.Index(p.targetGenDir, p.gatewayThriftNamespace)

	return path.Join(
		p.gatewayThriftNamespace,
		p.targetGenDir[nsIndex+len(p.gatewayThriftNamespace):],
	)
}

// PackageGenPath returns the Go package path for generated code from a thrift file.
func (p PackageHelper) PackageGenPath(thrift string) (string, error) {
	if !strings.HasSuffix(thrift, ".thrift") {
		return "", errors.Errorf("file %s is not .thrift", thrift)
	}
	root := path.Clean(p.gatewayThriftRootDir)
	idx := strings.Index(thrift, root)
	if idx == -1 {
		return "", errors.Errorf("file %s is not in thrift dir", thrift)
	}

	nsIndex := strings.Index(p.targetGenDir, p.gatewayThriftNamespace)

	return path.Join(
		p.gatewayThriftNamespace,
		p.targetGenDir[nsIndex+len(p.gatewayThriftNamespace):],
		filepath.Dir(thrift[idx+len(root):]),
	), nil
}

// TypePackageName returns the package name that defines the type.
func (p PackageHelper) TypePackageName(thrift string) (string, error) {
	if !strings.HasSuffix(thrift, ".thrift") {
		return "", errors.Errorf("file %s is not .thrift", thrift)
	}
	file := path.Base(thrift)
	return file[:len(file)-7], nil
}

func (p PackageHelper) getRelativeFileName(thrift string) (string, error) {
	if !strings.HasSuffix(thrift, ".thrift") {
		return "", errors.Errorf("file %s is not .thrift", thrift)
	}
	root := path.Clean(p.gatewayThriftRootDir)
	idx := strings.Index(thrift, root)
	if idx == -1 {
		return "", errors.Errorf("file %s is not in thrift dir %s", thrift, p.thriftRootDir)
	}
	return thrift[idx+len(root):], nil
}

// TargetClientPath returns the path for generated file for services in a thrift file.
func (p PackageHelper) TargetClientPath(thrift string) (string, error) {
	fileName, err := p.getRelativeFileName(thrift)
	if err != nil {
		return "", err
	}

	goFile := strings.Replace(fileName, ".thrift", ".go", -1)
	return path.Join(p.targetGenDir, goFile), nil
}

// TargetClientsInitPath returns where the clients init should go
func (p PackageHelper) TargetClientsInitPath() string {
	return path.Join(p.targetGenDir, "clients", "clients.go")
}

// TargetStructPath returns the path for any structs needed for
// a generated client based on thrift file.
func (p PackageHelper) TargetStructPath(thrift string) (string, error) {
	fileName, err := p.getRelativeFileName(thrift)
	if err != nil {
		return "", err
	}
	goFile := strings.Replace(fileName, ".thrift", "_structs.go", -1)
	return path.Join(p.targetGenDir, goFile), nil
}

// TargetEndpointPath returns the path for the endpoint handler based
// on the thrift file and method name
func (p PackageHelper) TargetEndpointPath(
	thrift, serviceName, methodName string,
) (string, error) {
	fileName, err := p.getRelativeFileName(thrift)
	if err != nil {
		return "", err
	}

	fileEnding := "_" + strings.ToLower(serviceName) + "_method_" + strings.ToLower(methodName) + ".go"
	goFile := strings.Replace(fileName, ".thrift", fileEnding, -1)
	return path.Join(p.targetGenDir, goFile), nil
}

// TargetEndpointTestPath returns the path for the endpoint test based
// on the thrift file and method name
func (p PackageHelper) TargetEndpointTestPath(
	thrift, serviceName, methodName string,
) (string, error) {
	fileName, err := p.getRelativeFileName(thrift)
	if err != nil {
		return "", err
	}

	fileEnding := "_" + strings.ToLower(serviceName) + "_method_" + strings.ToLower(methodName) + "_test.go"
	goFile := strings.Replace(fileName, ".thrift", fileEnding, -1)
	return path.Join(p.targetGenDir, goFile), nil
}

// TargetEndpointPath

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
