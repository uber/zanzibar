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
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
)

func getProjectDir() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(filepath.Dir(file))
}

var templateDir = filepath.Join(
	getProjectDir(), "./codegen/templates/*.tmpl",
)

// GatewaySpec collects information for the entire gateway
type GatewaySpec struct {
	// package helper for gateway
	PackageHelper *PackageHelper
	// tempalte instance for gateway
	Template *Template

	ClientModules   map[string]*ModuleSpec
	EndpointModules map[string]*ModuleSpec

	configDirName     string
	clientThriftDir   string
	endpointThriftDir string
}

// NewGatewaySpec sets up gateway spec
func NewGatewaySpec(
	configDirName string,
	thriftRootDir string,
	typeFileRootDir string,
	targetGenDir string,
	gatewayThriftRootDir string,
	clientThriftDir string,
	endpointThriftDir string,
) (*GatewaySpec, error) {
	packageHelper, err := NewPackageHelper(
		thriftRootDir,
		typeFileRootDir,
		targetGenDir,
		gatewayThriftRootDir,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot build package helper")
	}

	tmpl, err := NewTemplate(templateDir)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create template")
	}

	clientThrifts, err := filepath.Glob(filepath.Join(
		configDirName,
		clientThriftDir,
		"*/*.thrift",
	))
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load client thrift files")
	}

	endpointThrifts, err := filepath.Glob(filepath.Join(
		configDirName,
		endpointThriftDir,
		"*/*.thrift",
	))
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load endpoint thrift files")
	}

	spec := &GatewaySpec{
		PackageHelper:   packageHelper,
		Template:        tmpl,
		ClientModules:   map[string]*ModuleSpec{},
		EndpointModules: map[string]*ModuleSpec{},

		configDirName:     configDirName,
		clientThriftDir:   clientThriftDir,
		endpointThriftDir: endpointThriftDir,
	}

	for _, thrift := range clientThrifts {
		module, err := NewModuleSpec(thrift, packageHelper)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot parse client thrift file %s :", thrift,
			)
		}
		spec.ClientModules[thrift] = module
	}
	for _, thrift := range endpointThrifts {
		module, err := NewModuleSpec(thrift, packageHelper)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot parse endpoint thrift file %s :", thrift,
			)
		}
		spec.EndpointModules[thrift] = module
	}

	return spec, nil
}

// GenerateClients will generate all the clients for the gateway
func (gateway *GatewaySpec) GenerateClients() error {
	for _, module := range gateway.ClientModules {
		_, err := gateway.Template.GenerateClientFile(
			module.ThriftFile, gateway.PackageHelper,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// GenerateEndpoints will generate all the endpoints for the gateway
func (gateway *GatewaySpec) GenerateEndpoints() error {
	for _, module := range gateway.EndpointModules {
		_, err := gateway.Template.GenerateEndpointFile(
			module.ThriftFile, gateway.PackageHelper,
		)
		if err != nil {
			return err
		}

		_, err = gateway.Template.GenerateEndpointTestFile(
			module.ThriftFile, gateway.PackageHelper,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
