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
	// Whether the ThriftFile should have annotations or not
	WantAnnot bool
	// Go package path of this module.
	GoPackage string
	// Go package name, generated base on module name.
	PackageName string
	// Go client types file path, generated from thrift file.
	GoThriftTypesFilePath string
	// Generated imports
	IncludedPackages []GoPackageImport
	Services         []*ServiceSpec
}

// GoPackageImport ...
type GoPackageImport struct {
	PackageName string
	AliasName   string
}

// ServiceSpec specifies a service.
type ServiceSpec struct {
	// Service name
	Name string
	// Source thrift file to generate the code.
	ThriftFile string
	// Whether the service should have annotations or not
	WantAnnot bool
	// List of methods/endpoints of the service
	Methods []*MethodSpec
	// thriftrw compile spec.
	CompileSpec *compile.ServiceSpec
}

// NewModuleSpec returns a specification for a thrift module
func NewModuleSpec(thrift string, wantAnnot bool, packageHelper *PackageHelper) (*ModuleSpec, error) {
	module, err := compile.Compile(thrift)
	if err != nil {
		return nil, errors.Wrap(err, "failed parse thrift file")
	}
	targetPackage, err := packageHelper.PackageGenPath(module.ThriftPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate target package path")
	}

	moduleSpec := &ModuleSpec{
		WantAnnot:   wantAnnot,
		ThriftFile:  module.ThriftPath,
		GoPackage:   targetPackage,
		PackageName: camelCase(module.GetName()),
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
		serviceSpec, err := NewServiceSpec(module.Services[name], ms.WantAnnot, packageHelper)
		if err != nil {
			return err
		}
		ms.Services = append(ms.Services, serviceSpec)
	}
	return nil
}

// NewServiceSpec creates a service specification from given thrift file path.
func NewServiceSpec(spec *compile.ServiceSpec, wantAnnot bool, packageHelper *PackageHelper) (*ServiceSpec, error) {
	serviceSpec := &ServiceSpec{
		WantAnnot:   wantAnnot,
		Name:        spec.Name,
		ThriftFile:  spec.File,
		CompileSpec: spec,
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

// SetDownstream ...
func (ms *ModuleSpec) SetDownstream(
	serviceName string, methodName string,
	clientSpec *ClientSpec, clientService string, clientMethod string,
	h *PackageHelper,
) error {
	var service *ServiceSpec
	for _, v := range ms.Services {
		if v.Name == serviceName {
			service = v
			break
		}
	}

	if service == nil {
		return errors.Errorf(
			"Module does not have service (%s)\n", serviceName,
		)
	}

	var method *MethodSpec
	for _, v := range service.Methods {
		if v.Name == methodName {
			method = v
			break
		}
	}

	if method == nil {
		return errors.Errorf(
			"Service (%s) does not have method (%s)\n", serviceName, methodName,
		)
	}

	err := method.setDownstream(
		clientSpec.ModuleSpec, serviceName, clientMethod,
	)
	if err != nil {
		return err
	}

	// If this is an endpoint then a downstream will be defined.
	// If if it a client it will not be.
	if method.Downstream != nil {
		var downstreamMethod *MethodSpec

		for _, dsMethod := range method.Downstream.Services[0].Methods {
			if method.Name == dsMethod.Name {
				downstreamMethod = dsMethod
				break
			}
		}
		if downstreamMethod == nil {
			return errors.Errorf("Failed to map %s to one of the downstream methods: %v  ", method.Name, method.Downstream.Services[0].Methods)
		}
		downstreamSpec := downstreamMethod.CompiledThriftSpec
		funcSpec := method.CompiledThriftSpec

		err := method.setTypeConverters(funcSpec, downstreamSpec, h)
		if err != nil {
			return err
		}
	}

	// Adds imports for downstream services.
	if !ms.isPackageIncluded(clientSpec.GoPackageName) {

		ms.IncludedPackages = append(
			ms.IncludedPackages, GoPackageImport{
				PackageName: clientSpec.GoPackageName,
				AliasName:   "",
			},
		)
	}

	// Adds imports for thrift types used by downstream services.
	for _, service := range ms.Services {
		for _, method := range service.Methods {
			d := method.Downstream
			if d != nil && !ms.isPackageIncluded(d.GoPackage) {
				// thrift types file is optional...
				if method.Downstream.GoThriftTypesFilePath == "" {
					continue
				}

				ms.IncludedPackages = append(
					ms.IncludedPackages, GoPackageImport{
						PackageName: method.Downstream.GoThriftTypesFilePath,
						AliasName:   "",
					},
				)
			}
		}
	}

	return nil
}

// NewMethod creates new method specification.
func (s *ServiceSpec) NewMethod(
	funcSpec *compile.FunctionSpec, packageHelper *PackageHelper,
) (*MethodSpec, error) {
	return NewMethod(s.ThriftFile, funcSpec, packageHelper, s.WantAnnot, s.Name)
}

func (ms *ModuleSpec) addTypeImport(thriftPath string, packageHelper *PackageHelper) error {
	newPkg, err := packageHelper.TypeImportPath(thriftPath)
	if err != nil {
		return err
	}
	aliasName, err := packageHelper.TypePackageName(thriftPath)
	if err != nil {
		return err
	}

	if !ms.isPackageIncluded(newPkg) {
		ms.IncludedPackages = append(
			ms.IncludedPackages, GoPackageImport{
				PackageName: newPkg,
				AliasName:   aliasName,
			},
		)
	}
	return nil
}

func (ms *ModuleSpec) isPackageIncluded(pkg string) bool {
	for _, includedPkg := range ms.IncludedPackages {
		if pkg == includedPkg.PackageName {
			return true
		}
	}
	return false
}
