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
	"sort"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

// ModuleSpec collects the service specifications from thrift file.
type ModuleSpec struct {
	// Source thrift file to generate the code.
	ThriftFile string
	// Go package path of this module.
	GoPackage string
	// Go package name, generated base on module name.
	PackageName string
	// Go client file path, generated from thrift file.
	GoClientFilePath string
	// Go client structs file path, generated from thrift file.
	GoClientStructsFilePath string
	// Generated imports
	IncludedPackages []string
	Services         []*ServiceSpec
}

// ServiceSpec specifies a service.
type ServiceSpec struct {
	// Service name
	Name string
	// Source thrift file to generate the code.
	ThriftFile string
	// List of methods/endpoints of the service
	Methods []*MethodSpec
}

// NewModuleSpec returns a specification for a thrift module
func NewModuleSpec(thrift string, packageHelper *PackageHelper) (*ModuleSpec, error) {
	module, err := compile.Compile(thrift)
	if err != nil {
		return nil, errors.Wrap(err, "failed parse thrift file")
	}
	clientPath, err := packageHelper.TargetClientPath(module.ThriftPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate target client path")
	}
	targetPackage, err := packageHelper.PackageGenPath(module.ThriftPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate target package path")
	}

	clientStructsPath, err := packageHelper.TargetClientStructPath(
		module.ThriftPath,
	)
	if err != nil {
		return nil, errors.Wrap(
			err, "failed to generate target client structs path",
		)
	}

	moduleSpec := &ModuleSpec{
		ThriftFile:              module.ThriftPath,
		GoPackage:               targetPackage,
		PackageName:             module.GetName(),
		GoClientFilePath:        clientPath,
		GoClientStructsFilePath: clientStructsPath,
	}
	if err := moduleSpec.AddServices(module, packageHelper); err != nil {
		return nil, err
	}
	if err := moduleSpec.AddImports(module, packageHelper); err != nil {
		return nil, err
	}
	return moduleSpec, nil
}

// AddImports adds imported Go packages in ModuleSpec in alphabetical order.
func (ms *ModuleSpec) AddImports(module *compile.Module, packageHelper *PackageHelper) error {
	for _, pkg := range module.Includes {
		if err := ms.addTypeImport(pkg.Module.ThriftPath, packageHelper); err != nil {
			return errors.Wrapf(err, "can't add import %s", pkg.Module.ThriftPath)
		}
	}

	if err := ms.addTypeImport(ms.ThriftFile, packageHelper); err != nil {
		return errors.Wrapf(err, "can't add import %s", ms.ThriftFile)
	}

	// Adds imports for downstream services.
	for _, service := range ms.Services {
		for _, method := range service.Methods {
			if d := method.Downstream; d != nil && !ms.isPackageIncluded(d.GoPackage) {
				ms.IncludedPackages = append(ms.IncludedPackages, method.Downstream.GoPackage)
			}
		}
	}
	sort.Strings(ms.IncludedPackages)
	return nil
}

// AddServices adds services in ModuleSpec in alphabetical order of service names.
func (ms *ModuleSpec) AddServices(module *compile.Module, packageHelper *PackageHelper) error {
	names := make([]string, 0, len(module.Services))
	for name := range module.Services {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		serviceSpec, err := NewServiceSpec(module.Services[name], packageHelper)
		if err != nil {
			return err
		}
		ms.Services = append(ms.Services, serviceSpec)
	}
	return nil
}

// NewServiceSpec creates a service specification from given thrift file path.
func NewServiceSpec(spec *compile.ServiceSpec, packageHelper *PackageHelper) (*ServiceSpec, error) {
	serviceSpec := &ServiceSpec{
		Name:       spec.Name,
		ThriftFile: spec.File,
	}
	funcNames := make([]string, 0, len(spec.Functions))
	for name := range spec.Functions {
		funcNames = append(funcNames, name)
	}
	sort.Strings(funcNames)
	for _, funcName := range funcNames {
		method, err := serviceSpec.NewMethod(spec.Functions[funcName], packageHelper)
		if err != nil {
			return nil, errors.Wrapf(err, "service %s method %s", spec.Name, funcName)
		}
		serviceSpec.Methods = append(serviceSpec.Methods, method)
	}
	return serviceSpec, nil
}

// NewMethod creates new method specification.
func (s *ServiceSpec) NewMethod(funcSpec *compile.FunctionSpec, packageHelper *PackageHelper) (*MethodSpec, error) {
	method := &MethodSpec{}
	var err error
	var ok bool
	method.Name = funcSpec.MethodName()
	if method.HTTPMethod, ok = funcSpec.Annotations[antHTTPMethod]; !ok {
		return nil, errors.Errorf("missing anotation '%s' for HTTP method", antHTTPMethod)
	}
	if method.HTTPPath, ok = funcSpec.Annotations[antHTTPPath]; !ok {
		return nil, errors.Errorf("missing anotation '%s' for HTTP path", antHTTPPath)
	}
	method.EndpointName = funcSpec.Annotations[antHandler]
	method.Headers = headers(funcSpec.Annotations[antHTTPHeaders])
	if err := method.setDownstream(funcSpec.Annotations[antHTTPDownstream], s.ThriftFile, packageHelper); err != nil {
		return nil, err
	}
	if err = method.setExceptionStatusCode(funcSpec.ResultSpec); err != nil {
		return nil, err
	}
	if err = method.setOKStatusCode(funcSpec.Annotations[antHTTPStatus]); err != nil {
		return nil, err
	}
	if err = method.setResponseType(s.ThriftFile, funcSpec.ResultSpec, packageHelper); err != nil {
		return nil, err
	}
	if err = method.setRequestType(s.ThriftFile, funcSpec, packageHelper); err != nil {
		return nil, err
	}
	if method.HTTPMethod == "GET" && method.RequestType != "" {
		return nil, errors.Errorf("invalid annotation: HTTP GET method with body type")
	}
	return method, nil
}

func (ms *ModuleSpec) addTypeImport(thriftPath string, packageHelper *PackageHelper) error {
	newPkg, err := packageHelper.TypeImportPath(thriftPath)
	if err != nil {
		return err
	}
	if !ms.isPackageIncluded(newPkg) {
		ms.IncludedPackages = append(ms.IncludedPackages, newPkg)
	}
	return nil
}

func (ms *ModuleSpec) isPackageIncluded(pkg string) bool {
	for _, includedPkg := range ms.IncludedPackages {
		if pkg == includedPkg {
			return true
		}
	}
	return false
}
