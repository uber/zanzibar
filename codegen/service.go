// Copyright (c) 2018 Uber Technologies, Inc.
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
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

// ModuleSpec collects the service specifications from thrift file.
type ModuleSpec struct {
	// Source thrift file to generate the code.
	ThriftFile string
	// Whether the ThriftFile should have annotations or not
	WantAnnot bool
	// Whether the module is for an endpoint vs downstream client
	IsEndpoint bool
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
	// Whether the service is for an endpoint vs downstream client
	IsEndpoint bool
	// List of methods/endpoints of the service
	Methods []*MethodSpec
	// thriftrw compile spec.
	CompileSpec *compile.ServiceSpec
}

// NewModuleSpec returns a specification for a thrift module
func NewModuleSpec(
	thrift string,
	wantAnnot bool,
	isEndpoint bool,
	packageHelper *PackageHelper,
) (*ModuleSpec, error) {
	module, err := compile.Compile(thrift)
	if err != nil {
		return nil, errors.Wrap(err, "failed parse thrift file")
	}

	moduleSpec := &ModuleSpec{
		WantAnnot:   wantAnnot,
		IsEndpoint:  isEndpoint,
		ThriftFile:  module.ThriftPath,
		PackageName: packageName(module.GetName()),
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
	err := module.Walk(func(dep *compile.Module) error {
		if err := ms.addTypeImport(dep.ThriftPath, packageHelper); err != nil {
			return errors.Wrapf(err, "can't add import %s", dep.ThriftPath)
		}
		return nil
	})
	if err != nil {
		return err
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
		serviceSpec, err := NewServiceSpec(
			module.Services[name],
			ms.WantAnnot,
			ms.IsEndpoint,
			packageHelper,
		)
		if err != nil {
			return err
		}
		ms.Services = append(ms.Services, serviceSpec)
	}
	return nil
}

// NewServiceSpec creates a service specification from given thrift file path.
func NewServiceSpec(
	spec *compile.ServiceSpec,
	wantAnnot bool,
	isEndpoint bool,
	packageHelper *PackageHelper,
) (*ServiceSpec, error) {
	serviceSpec := &ServiceSpec{
		WantAnnot:   wantAnnot,
		IsEndpoint:  isEndpoint,
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
	serviceName string,
	methodName string,
	clientSpec *ClientSpec,
	clientMethod string,
	headersPropagate map[string]FieldMapperEntry,
	reqTransforms map[string]FieldMapperEntry,
	respTransforms map[string]FieldMapperEntry,
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
			"Module does not have service %q\n", serviceName,
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
			"Service %q does not have method %q\n", serviceName, methodName,
		)
	}

	serviceMethod, ok := clientSpec.ExposedMethods[clientMethod]
	if !ok {
		return errors.Errorf("Client %q does not expose method %q", clientSpec.ClientName, clientMethod)
	}
	sm := strings.Split(serviceMethod, "::")

	err := method.setDownstream(clientSpec.ModuleSpec, sm[0], sm[1])

	if err != nil {
		return err
	}

	// Exception validation
	for en := range method.DownstreamMethod.ExceptionsIndex {
		if _, ok := method.ExceptionsIndex[en]; !ok {
			return fmt.Errorf("Missing exception %s in Endpoint schema", en)
		}
	}

	// If this is an endpoint then a downstream will be defined.
	// If if it a client it will not be.
	if method.Downstream != nil {
		// TODO: once all client configs have "exposedMethods" field, we can find the exact
		// service, instead of loop over for service looking for the first matching method,
		// which could totally be wrong method
		downstreamMethod, err := findMethodByName(method.Name, method.Downstream.Services)
		if err != nil {
			return err
		}
		downstreamSpec := downstreamMethod.CompiledThriftSpec
		funcSpec := method.CompiledThriftSpec

		err = method.setTypeConverters(funcSpec, downstreamSpec, reqTransforms, headersPropagate, respTransforms, h, downstreamMethod)
		if err != nil {
			return err
		}

	}

	if method.Downstream != nil && len(headersPropagate) > 0 {

		downstreamMethod, err := findMethodByName(method.Name, method.Downstream.Services)
		if err != nil {
			return err
		}
		downstreamSpec := downstreamMethod.CompiledThriftSpec

		err = method.setHeaderPropagator(method.ReqHeaders, downstreamSpec, headersPropagate, h, downstreamMethod)
		if err != nil {
			return err
		}
	}

	// Adds imports for downstream services.
	if !ms.isPackageIncluded(clientSpec.ImportPackagePath) {

		ms.IncludedPackages = append(
			ms.IncludedPackages, GoPackageImport{
				PackageName: clientSpec.ImportPackagePath,
				AliasName:   clientSpec.ImportPackageAlias,
			},
		)
	}

	// Adds imports for thrift types used by downstream services.
	for _, service := range ms.Services {
		for _, method := range service.Methods {
			d := method.Downstream
			if d != nil && !ms.isPackageIncluded(d.GoThriftTypesFilePath) {
				// thrift types file is optional...
				if d.GoThriftTypesFilePath == "" {
					continue
				}

				ms.IncludedPackages = append(
					ms.IncludedPackages, GoPackageImport{
						PackageName: d.GoThriftTypesFilePath,
						AliasName:   "",
					},
				)
			}
		}
	}

	return nil
}

func findMethodByName(name string, serviceSpecs []*ServiceSpec) (*MethodSpec, error) {
	var allMethods []string
	for _, s := range serviceSpecs {
		for _, dsMethod := range s.Methods {
			allMethods = append(allMethods, s.Name+"::"+dsMethod.Name)
			if name == dsMethod.Name {
				return dsMethod, nil
			}
		}
	}
	return nil, errors.Errorf("failed to map downstream method %q to methods %q defined in thrift file", name, allMethods)
}

// NewMethod creates new method specification.
func (s *ServiceSpec) NewMethod(
	funcSpec *compile.FunctionSpec, packageHelper *PackageHelper,
) (*MethodSpec, error) {
	return NewMethod(s.ThriftFile, funcSpec, packageHelper, s.WantAnnot, s.IsEndpoint, s.Name)
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
