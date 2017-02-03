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

package gencode

import (
	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

// ModuleSpec collects the service specifications from thrift file.
type ModuleSpec struct {
	Services []*ServiceSpec
}

// ServiceSpec specifies a service.
type ServiceSpec struct {
	// Service name
	Name string
	// Go package name, generated base on module name.
	PackageName string
	// Go file path, generated from thrift file and service name.
	FilePath string
	// Source thrift file to generate the code.
	ThriftFile string
	// Generated imports
	IncludedPackages []string
	// List of methods/endpoints of the service
	Methods []*MethodSpec
}

// NewModuleSpec returns a specification for a thrift module
func NewModuleSpec(thrift string, h *PackageHelper) (*ModuleSpec, error) {
	module, err := compile.Compile(thrift)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse thrift file")
	}
	moduleSpec := new(ModuleSpec)
	for _, s := range module.Services {
		serviceSpec, err := NewServiceSpec(module.GetName(), s, h)
		if err != nil {
			return nil, err
		}
		for _, pkg := range module.Includes {
			if e := serviceSpec.addImports(pkg.Module.ThriftPath, h); e != nil {
				return nil, errors.Wrapf(e, "can't add import %s", pkg.Module.ThriftPath)
			}
		}
		moduleSpec.Services = append(moduleSpec.Services, serviceSpec)
	}
	return moduleSpec, nil
}

// NewServiceSpec creates a service specification from given thrift file path.
func NewServiceSpec(packageName string, spec *compile.ServiceSpec, packageHelper *PackageHelper) (*ServiceSpec, error) {
	filePath, err := packageHelper.TargetGenPath(spec.File)
	if err != nil {
		return nil, err
	}
	serviceSpec := &ServiceSpec{
		Name:        spec.Name,
		PackageName: packageName,
		ThriftFile:  spec.File,
		FilePath:    filePath,
	}
	for _, f := range spec.Functions {
		method, err := serviceSpec.NewMethod(f, packageHelper)
		if err != nil {
			return nil, errors.Wrapf(err, "service %s method %s", spec.Name, f.MethodName())
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
	return method, nil
}

func (s *ServiceSpec) addImports(thriftPath string, h *PackageHelper) error {
	if thriftPath == s.ThriftFile {
		return nil
	}
	newPkg, err := h.TypeImportPath(thriftPath)
	if err != nil {
		return err
	}
	for _, pkg := range s.IncludedPackages {
		if newPkg == pkg {
			return nil
		}
	}
	s.IncludedPackages = append(s.IncludedPackages, newPkg)
	return nil
}
