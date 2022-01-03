// Copyright (c) 2022 Uber Technologies, Inc.
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
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const packageRoot = "github.com/uber/zanzibar/codegen/test-service"

type TestClientSpec struct {
	Info string
}

type TestEndpointSpec struct {
	Info string
}

var tSpec = TestClientSpec{
	Info: "tchannel",
}
var hSpec = TestClientSpec{
	Info: "http",
}
var gSpec = TestClientSpec{
	Info: "grpc",
}

type TestHTTPClientGenerator struct{}

func (t *TestHTTPClientGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	return &BuildResult{
		Spec: &hSpec,
	}, nil
}

func (*TestHTTPClientGenerator) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	return &hSpec, nil
}

type TestTChannelClientGenerator struct{}

func (t *TestTChannelClientGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	return &BuildResult{
		Spec: &tSpec,
	}, nil
}

func (*TestTChannelClientGenerator) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	return &tSpec, nil
}

type TestGRPCClientGenerator struct{}

func (*TestGRPCClientGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	return &BuildResult{
		Spec: &tSpec,
	}, nil
}

func (*TestGRPCClientGenerator) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	return &tSpec, nil
}

type TestHTTPEndpointGenerator struct{}

func (*TestHTTPEndpointGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	return nil, nil
}

func (*TestHTTPEndpointGenerator) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	return &TestEndpointSpec{
		Info: "http",
	}, nil
}

type TestHTTPEndpointGeneratorSkip struct{}

func (*TestHTTPEndpointGeneratorSkip) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	return nil, nil
}

func (*TestHTTPEndpointGeneratorSkip) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	return nil, errors.Wrapf(&ErrorSkipCodeGen{IDLFile: "dummy"}, "")
}

type TestServiceSpec struct {
	Info string
}

type TestServiceGenerator struct{}

func (*TestServiceGenerator) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	return &TestServiceSpec{
		Info: "gateway",
	}, nil
}

func (*TestServiceGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	return nil, nil
}

func TestExampleService(t *testing.T) {
	moduleSystem := NewModuleSystem(
		map[string][]string{
			"client": {
				"clients/*",
				"endpoints/*/*",
			},
			"endpoint": {
				"endpoints/*",
				"another/*",
				"more-endpoints/*",
			},
			"middleware": {"middlewares/*"},
			"service":    {"services/*"},
		},
		map[string][]string{},
		false,
	)
	var err error

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	})
	if err != nil {
		t.Errorf("Unexpected error registering client class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"http",
		&TestHTTPClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"tchannel",
		&TestTChannelClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering tchannel client class type: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"grpc",
		&TestGRPCClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error regarding grpc client class type :%s", err)
	}

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "endpoint",
		NamePlural: "endpoints",
		ClassType:  MultiModule,
		DependsOn:  []string{"client"},
	})
	if err != nil {
		t.Errorf("Unexpected error registering endpoint class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"endpoint",
		"http",
		&TestHTTPEndpointGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"endpoint",
		"http",
		&TestHTTPEndpointGenerator{},
	)
	if err == nil {
		t.Errorf("Expected double creation of http endpoint to error")
	}

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	})
	if err == nil {
		t.Errorf("Expected double definition of client class to error")
	}

	currentDir := getTestDirName()
	testServiceDir := path.Join(currentDir, "test-service")

	// TODO: this should return a collection of errors if they occur
	instances, err := moduleSystem.GenerateBuild(
		"github.com/uber/zanzibar/codegen/test-service",
		testServiceDir,
		path.Join(testServiceDir, "build"),
		Options{
			CommitChange: true,
		},
	)
	if err != nil {
		t.Errorf("Unexpected error generating build %s", err)
	}

	expectedClientDependency := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "tchannel",
		Directory:     "clients/example-dependency",
		InstanceName:  "example-dependency",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampledependencyClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example-dependency",
			IsExportGenerated:     true,
			PackageAlias:          "exampledependencyClientStatic",
			PackageName:           "exampledependencyClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example-dependency",
		},
		Dependencies:          []ModuleDependency{},
		ResolvedDependencies:  map[string][]*ModuleInstance{},
		RecursiveDependencies: map[string][]*ModuleInstance{},
	}

	expectedClientInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "http",
		Directory:     "clients/example",
		InstanceName:  "example",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampleClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example",
			IsExportGenerated:     true,
			PackageAlias:          "exampleClientStatic",
			PackageName:           "exampleClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example-dependency",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
	}

	expectedGRPCClientInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "grpc",
		Directory:     "clients/example-example",
		InstanceName:  "example-grpc",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampleClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example-grpc",
			IsExportGenerated:     true,
			PackageAlias:          "exampleClientStatic",
			PackageName:           "exampleClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example-grpc",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example-dependency",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
	}

	expectedEmbeddedClient := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "http",
		Directory:     "endpoints/health/embedded-client",
		InstanceName:  "embedded-client",
		JSONFileName:  "",
		YAMLFileName:  "endpoint-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "embeddedClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/endpoints/health/embedded-client",
			IsExportGenerated:     true,
			PackageAlias:          "embeddedClientStatic",
			PackageName:           "embeddedClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/endpoints/health/embedded-client",
		},
		Dependencies:          []ModuleDependency{},
		ResolvedDependencies:  map[string][]*ModuleInstance{},
		RecursiveDependencies: map[string][]*ModuleInstance{},
	}

	expectedHealthEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "endpoints/health",
		InstanceName:  "health",
		JSONFileName:  "endpoint-config.json",
		YAMLFileName:  "",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "healthendpointgenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/endpoints/health",
			IsExportGenerated:     true,
			PackageAlias:          "healthendpointstatic",
			PackageName:           "healthendpoint",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/endpoints/health",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientInstance,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				// Note that the dependencies are ordered
				&expectedClientDependency,
				&expectedClientInstance,
			},
		},
	}

	expectedFooEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "more-endpoints/foo",
		InstanceName:  "more-endpoints/foo",
		JSONFileName:  "",
		YAMLFileName:  "endpoint-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "fooendpointgenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/more-endpoints/foo",
			IsExportGenerated:     true,
			PackageAlias:          "fooendpointstatic",
			PackageName:           "fooendpoint",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/more-endpoints/foo",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientInstance,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				// Note that the dependencies are ordered
				&expectedClientDependency,
				&expectedClientInstance,
			},
		},
	}

	expectedBarEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "another/bar",
		InstanceName:  "another/bar",
		JSONFileName:  "",
		YAMLFileName:  "endpoint-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "barendpointgenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/another/bar",
			IsExportGenerated:     true,
			PackageAlias:          "barendpointstatic",
			PackageName:           "barendpoint",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/another/bar",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientInstance,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				// Note that the dependencies are ordered
				&expectedClientDependency,
				&expectedClientInstance,
			},
		},
	}

	expectedClients := []*ModuleInstance{
		&expectedClientInstance,
		&expectedClientDependency,
		&expectedEmbeddedClient,
		&expectedGRPCClientInstance,
	}
	// Note: Stable ordering is not required by POSIX
	expectedEndpoints := []*ModuleInstance{
		&expectedHealthEndpointInstance,
		&expectedBarEndpointInstance,
		&expectedFooEndpointInstance,
	}

	assertInstances(t, expectedClients, expectedEndpoints, instances)
}

func TestExampleServiceIncremental(t *testing.T) {
	moduleSystem := NewModuleSystem(
		map[string][]string{
			"client": {
				"clients/*",
				"endpoints/*/*",
			},
			"endpoint": {
				"endpoints/*",
				"another/*",
				"more-endpoints/*",
			},
			"middleware": {"middlewares/*"},
			"service":    {"services/*"},
		},
		map[string][]string{},
		false,
	)
	var err error

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	})
	if err != nil {
		t.Errorf("Unexpected error registering client class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"http",
		&TestHTTPClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"tchannel",
		&TestTChannelClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering tchannel client class type: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"grpc",
		&TestGRPCClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error regarding grpc client class type :%s", err)
	}

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "endpoint",
		NamePlural: "endpoints",
		ClassType:  MultiModule,
		DependsOn:  []string{"client"},
	})
	if err != nil {
		t.Errorf("Unexpected error registering endpoint class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"endpoint",
		"http",
		&TestHTTPEndpointGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"endpoint",
		"http",
		&TestHTTPEndpointGenerator{},
	)
	if err == nil {
		t.Errorf("Expected double creation of http endpoint to error")
	}

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	})
	if err == nil {
		t.Errorf("Expected double definition of client class to error")
	}

	currentDir := getTestDirName()
	testServiceDir := path.Join(currentDir, "test-service")
	targetGenDir := path.Join(testServiceDir, "build")

	resolvedModules, err := moduleSystem.ResolveModules(
		packageRoot,
		testServiceDir,
		targetGenDir,
		Options{},
	)
	if err != nil {
		t.Errorf("Unexpected error generating modukes %s", err)
	}

	instances, err := moduleSystem.IncrementalBuild(
		packageRoot,
		testServiceDir,
		targetGenDir,
		[]ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
			{
				ClassName:    "client",
				InstanceName: "example-grpc",
			},
		},
		resolvedModules,
		Options{
			CommitChange:     true,
			QPSLevelsEnabled: true,
		},
	)
	if err != nil {
		t.Errorf("Unexpected error generating build %s", err)
	}
	expectedQPSLevels := map[string]int{
		"embeddedClientTest-TestMethod": 1,
	}
	expectedClientDependency := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "tchannel",
		Directory:     "clients/example-dependency",
		InstanceName:  "example-dependency",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampledependencyClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example-dependency",
			IsExportGenerated:     true,
			PackageAlias:          "exampledependencyClientStatic",
			PackageName:           "exampledependencyClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example-dependency",
		},
		Dependencies:          []ModuleDependency{},
		ResolvedDependencies:  map[string][]*ModuleInstance{},
		RecursiveDependencies: map[string][]*ModuleInstance{},
		QPSLevels:             expectedQPSLevels,
	}

	expectedClientInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "http",
		Directory:     "clients/example",
		InstanceName:  "example",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampleClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example",
			IsExportGenerated:     true,
			PackageAlias:          "exampleClientStatic",
			PackageName:           "exampleClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example-dependency",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedGRPCClientInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "grpc",
		Directory:     "clients/example-example",
		InstanceName:  "example-grpc",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampleClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example-grpc",
			IsExportGenerated:     true,
			PackageAlias:          "exampleClientStatic",
			PackageName:           "exampleClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example-grpc",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example-dependency",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedHealthEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "endpoints/health",
		InstanceName:  "health",
		JSONFileName:  "endpoint-config.json",
		YAMLFileName:  "",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "healthendpointgenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/endpoints/health",
			IsExportGenerated:     true,
			PackageAlias:          "healthendpointstatic",
			PackageName:           "healthendpoint",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/endpoints/health",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientInstance,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				// Note that the dependencies are ordered
				&expectedClientDependency,
				&expectedClientInstance,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedFooEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "more-endpoints/foo",
		InstanceName:  "more-endpoints/foo",
		JSONFileName:  "",
		YAMLFileName:  "endpoint-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "fooendpointgenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/more-endpoints/foo",
			IsExportGenerated:     true,
			PackageAlias:          "fooendpointstatic",
			PackageName:           "fooendpoint",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/more-endpoints/foo",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientInstance,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				// Note that the dependencies are ordered
				&expectedClientDependency,
				&expectedClientInstance,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedBarEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "another/bar",
		InstanceName:  "another/bar",
		JSONFileName:  "",
		YAMLFileName:  "endpoint-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "barendpointgenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/another/bar",
			IsExportGenerated:     true,
			PackageAlias:          "barendpointstatic",
			PackageName:           "barendpoint",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/another/bar",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientInstance,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				// Note that the dependencies are ordered
				&expectedClientDependency,
				&expectedClientInstance,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedClients := []*ModuleInstance{
		&expectedClientInstance,
		&expectedGRPCClientInstance,
	}
	expectedEndpoints := []*ModuleInstance{
		&expectedHealthEndpointInstance,
		&expectedFooEndpointInstance,
		&expectedBarEndpointInstance,
	}
	assertInstances(t, expectedClients, expectedEndpoints, instances)
}

func TestExampleServiceIncrementalSkip(t *testing.T) {
	moduleSystem := NewModuleSystem(
		map[string][]string{
			"client": {
				"clients/*",
				"endpoints/*/*",
			},
			"endpoint": {
				"endpoints/*",
				"another/*",
				"more-endpoints/*",
			},
			"middleware": {"middlewares/*"},
			"service":    {"services/*"},
		},
		map[string][]string{},
		false,
	)
	var err error

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	})
	if err != nil {
		t.Errorf("Unexpected error registering client class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"http",
		&TestHTTPClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"tchannel",
		&TestTChannelClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering tchannel client class type: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"grpc",
		&TestGRPCClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error regarding grpc client class type :%s", err)
	}

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "endpoint",
		NamePlural: "endpoints",
		ClassType:  MultiModule,
		DependsOn:  []string{"client"},
	})
	if err != nil {
		t.Errorf("Unexpected error registering endpoint class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"endpoint",
		"http",
		&TestHTTPEndpointGeneratorSkip{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	})
	if err == nil {
		t.Errorf("Expected double definition of client class to error")
	}

	currentDir := getTestDirName()
	testServiceDir := path.Join(currentDir, "test-service")
	targetGenDir := path.Join(testServiceDir, "build")

	resolvedModules, err := moduleSystem.ResolveModules(
		packageRoot,
		testServiceDir,
		targetGenDir,
		Options{},
	)
	if err != nil {
		t.Errorf("Unexpected error generating modukes %s", err)
	}

	instances, err := moduleSystem.IncrementalBuild(
		packageRoot,
		testServiceDir,
		targetGenDir,
		[]ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
			{
				ClassName:    "client",
				InstanceName: "example-grpc",
			},
		},
		resolvedModules,
		Options{
			CommitChange:     true,
			QPSLevelsEnabled: true,
		},
	)
	if err != nil {
		t.Errorf("Unexpected error generating build %s", err)
	}

	expectedClientDependency := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "tchannel",
		Directory:     "clients/example-dependency",
		InstanceName:  "example-dependency",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampledependencyClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example-dependency",
			IsExportGenerated:     true,
			PackageAlias:          "exampledependencyClientStatic",
			PackageName:           "exampledependencyClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example-dependency",
		},
		Dependencies:          []ModuleDependency{},
		ResolvedDependencies:  map[string][]*ModuleInstance{},
		RecursiveDependencies: map[string][]*ModuleInstance{},
	}

	expectedClientInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "http",
		Directory:     "clients/example",
		InstanceName:  "example",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampleClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example",
			IsExportGenerated:     true,
			PackageAlias:          "exampleClientStatic",
			PackageName:           "exampleClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example-dependency",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
	}

	expectedGRPCClientInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "grpc",
		Directory:     "clients/example-example",
		InstanceName:  "example-grpc",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampleClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example-grpc",
			IsExportGenerated:     true,
			PackageAlias:          "exampleClientStatic",
			PackageName:           "exampleClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example-grpc",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example-dependency",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
	}

	expectedClients := []*ModuleInstance{
		&expectedClientInstance,
		&expectedGRPCClientInstance,
	}
	var expectedEndpoints []*ModuleInstance
	assertInstances(t, expectedClients, expectedEndpoints, instances)
}

func TestExampleServiceIncrementalSelective(t *testing.T) {
	moduleSystem := NewModuleSystem(
		map[string][]string{
			"client": {
				"clients/*",
				"endpoints/*/*",
			},
			"endpoint": {
				"endpoints/*",
				"another/*",
				"more-endpoints/*",
			},
			"middleware": {"middlewares/*"},
			"service":    {"services/*"},
		},
		map[string][]string{},
		true,
	)
	var err error

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	})
	if err != nil {
		t.Errorf("Unexpected error registering client class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"http",
		&TestHTTPClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"tchannel",
		&TestTChannelClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering tchannel client class type: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"grpc",
		&TestGRPCClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error regarding grpc client class type :%s", err)
	}

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "endpoint",
		NamePlural: "endpoints",
		ClassType:  MultiModule,
		DependsOn:  []string{"client"},
	})
	if err != nil {
		t.Errorf("Unexpected error registering endpoint class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"endpoint",
		"http",
		&TestHTTPEndpointGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"endpoint",
		"http",
		&TestHTTPEndpointGenerator{},
	)
	if err == nil {
		t.Errorf("Expected double creation of http endpoint to error")
	}

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	})
	if err == nil {
		t.Errorf("Expected double definition of client class to error")
	}

	if err := moduleSystem.RegisterClass(ModuleClass{
		Name:       "service",
		NamePlural: "services",
		ClassType:  MultiModule,
		DependsOn:  []string{"endpoint"},
	}); err != nil {
		t.Errorf("Unexpected error registering service: %s", err)
	}

	if err := moduleSystem.RegisterClassType("service", "gateway",
		&TestServiceGenerator{}); err != nil {
		t.Errorf("Unexpected error registering service class type: %s", err)
	}

	currentDir := getTestDirName()
	testServiceDir := path.Join(currentDir, "test-service")
	targetGenDir := path.Join(testServiceDir, "build")

	resolvedModules, err := moduleSystem.ResolveModules(
		packageRoot,
		testServiceDir,
		targetGenDir,
		Options{},
	)
	if err != nil {
		t.Errorf("Unexpected error generating modukes %s", err)
	}

	instances, err := moduleSystem.IncrementalBuild(
		packageRoot,
		testServiceDir,
		targetGenDir,
		[]ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
			{
				ClassName:    "client",
				InstanceName: "example-grpc",
			},
		},
		resolvedModules,
		Options{
			CommitChange:     true,
			QPSLevelsEnabled: true,
		},
	)
	if err != nil {
		t.Errorf("Unexpected error generating build %s", err)
	}

	expectedQPSLevels := map[string]int{
		"embeddedClientTest-TestMethod": 1,
	}
	expectedClientDependency := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "tchannel",
		Directory:     "clients/example-dependency",
		InstanceName:  "example-dependency",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampledependencyClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example-dependency",
			IsExportGenerated:     true,
			PackageAlias:          "exampledependencyClientStatic",
			PackageName:           "exampledependencyClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example-dependency",
		},
		Dependencies:          []ModuleDependency{},
		ResolvedDependencies:  map[string][]*ModuleInstance{},
		RecursiveDependencies: map[string][]*ModuleInstance{},
		QPSLevels:             expectedQPSLevels,
	}

	expectedClientInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "http",
		Directory:     "clients/example",
		InstanceName:  "example",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampleClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example",
			IsExportGenerated:     true,
			PackageAlias:          "exampleClientStatic",
			PackageName:           "exampleClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example-dependency",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedGRPCClientInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "grpc",
		Directory:     "clients/example-example",
		InstanceName:  "example-grpc",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampleClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example-grpc",
			IsExportGenerated:     true,
			PackageAlias:          "exampleClientStatic",
			PackageName:           "exampleClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example-grpc",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example-dependency",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedHealthEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "endpoints/health",
		InstanceName:  "health",
		JSONFileName:  "endpoint-config.json",
		YAMLFileName:  "",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "healthendpointgenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/endpoints/health",
			IsExportGenerated:     true,
			PackageAlias:          "healthendpointstatic",
			PackageName:           "healthendpoint",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/endpoints/health",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientInstance,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				// Note that the dependencies are ordered
				&expectedClientDependency,
				&expectedClientInstance,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedFooEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "more-endpoints/foo",
		InstanceName:  "more-endpoints/foo",
		JSONFileName:  "",
		YAMLFileName:  "endpoint-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "fooendpointgenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/more-endpoints/foo",
			IsExportGenerated:     true,
			PackageAlias:          "fooendpointstatic",
			PackageName:           "fooendpoint",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/more-endpoints/foo",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientInstance,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				// Note that the dependencies are ordered
				&expectedClientDependency,
				&expectedClientInstance,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedBarEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "another/bar",
		InstanceName:  "another/bar",
		JSONFileName:  "",
		YAMLFileName:  "endpoint-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "barendpointgenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/another/bar",
			IsExportGenerated:     true,
			PackageAlias:          "barendpointstatic",
			PackageName:           "barendpoint",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/another/bar",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientInstance,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				// Note that the dependencies are ordered
				&expectedClientDependency,
				&expectedClientInstance,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedClients := []*ModuleInstance{
		&expectedClientInstance,
		&expectedGRPCClientInstance,
	}
	expectedEndpoints := []*ModuleInstance{
		&expectedHealthEndpointInstance,
		&expectedFooEndpointInstance,
		&expectedBarEndpointInstance,
	}
	assertInstances(t, expectedClients, expectedEndpoints, instances)
}

func TestExampleServiceIncrementalWithDisabledQPSLevels(t *testing.T) {
	m := NewModuleSystem(
		map[string][]string{
			"client": {
				"clients/*",
				"endpoints/*/*",
			},
			"endpoint": {
				"endpoints/*",
				"another/*",
				"more-endpoints/*",
			},
			"middleware": {"middlewares/*"},
			"service":    {"services/*"},
		},
		map[string][]string{},
		true,
	)
	registerClass(t, m, "client", "clients", MultiModule, []string{})
	registerClassType(t, m, "client", "http", &TestHTTPClientGenerator{})
	registerClassType(t, m, "client", "tchannel", &TestTChannelClientGenerator{})
	registerClassType(t, m, "client", "grpc", &TestGRPCClientGenerator{})
	registerClass(t, m, "endpoint", "endpoints", MultiModule, []string{"client"})
	registerClassType(t, m, "endpoint", "http", &TestHTTPEndpointGenerator{})

	currentDir := getTestDirName()
	testServiceDir := path.Join(currentDir, "test-service")
	targetGenDir := path.Join(testServiceDir, "build")

	resolvedModules, err := m.ResolveModules(
		packageRoot,
		testServiceDir,
		targetGenDir,
		Options{
			CommitChange: true,
		},
	)
	if err != nil {
		t.Errorf("Unexpected error generating modules %s", err)
	}

	instances, err := m.IncrementalBuild(
		packageRoot, testServiceDir, targetGenDir, []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
			{
				ClassName:    "client",
				InstanceName: "example-grpc",
			},
		},
		resolvedModules,
		Options{
			CommitChange:     true,
			QPSLevelsEnabled: false,
		},
	)
	if err != nil {
		t.Errorf("Unexpected error generating build %s", err)
	}

	expectedQPSLevels := map[string]int{}
	expectedClientDependency := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "tchannel",
		Directory:     "clients/example-dependency",
		InstanceName:  "example-dependency",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampledependencyClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example-dependency",
			IsExportGenerated:     true,
			PackageAlias:          "exampledependencyClientStatic",
			PackageName:           "exampledependencyClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example-dependency",
		},
		Dependencies:          []ModuleDependency{},
		ResolvedDependencies:  map[string][]*ModuleInstance{},
		RecursiveDependencies: map[string][]*ModuleInstance{},
		QPSLevels:             expectedQPSLevels,
	}

	expectedClientInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "http",
		Directory:     "clients/example",
		InstanceName:  "example",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampleClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example",
			IsExportGenerated:     true,
			PackageAlias:          "exampleClientStatic",
			PackageName:           "exampleClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example-dependency",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedGRPCClientInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "client",
		ClassType:     "grpc",
		Directory:     "clients/example-example",
		InstanceName:  "example-grpc",
		JSONFileName:  "",
		YAMLFileName:  "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampleClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example-grpc",
			IsExportGenerated:     true,
			PackageAlias:          "exampleClientStatic",
			PackageName:           "exampleClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example-grpc",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example-dependency",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedHealthEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "endpoints/health",
		InstanceName:  "health",
		JSONFileName:  "endpoint-config.json",
		YAMLFileName:  "",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "healthendpointgenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/endpoints/health",
			IsExportGenerated:     true,
			PackageAlias:          "healthendpointstatic",
			PackageName:           "healthendpoint",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/endpoints/health",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientInstance,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientDependency,
				&expectedClientInstance,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedFooEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "more-endpoints/foo",
		InstanceName:  "more-endpoints/foo",
		JSONFileName:  "",
		YAMLFileName:  "endpoint-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "fooendpointgenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/more-endpoints/foo",
			IsExportGenerated:     true,
			PackageAlias:          "fooendpointstatic",
			PackageName:           "fooendpoint",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/more-endpoints/foo",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientInstance,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				// Note that the dependencies are ordered
				&expectedClientDependency,
				&expectedClientInstance,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedBarEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "another/bar",
		InstanceName:  "another/bar",
		JSONFileName:  "",
		YAMLFileName:  "endpoint-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "barendpointgenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/another/bar",
			IsExportGenerated:     true,
			PackageAlias:          "barendpointstatic",
			PackageName:           "barendpoint",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/another/bar",
		},
		Dependencies: []ModuleDependency{
			{
				ClassName:    "client",
				InstanceName: "example",
			},
		},
		ResolvedDependencies: map[string][]*ModuleInstance{
			"client": {
				&expectedClientInstance,
			},
		},
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {
				// Note that the dependencies are ordered
				&expectedClientDependency,
				&expectedClientInstance,
			},
		},
		QPSLevels: expectedQPSLevels,
	}

	expectedClients := []*ModuleInstance{
		&expectedClientInstance,
		&expectedGRPCClientInstance,
	}
	expectedEndpoints := []*ModuleInstance{
		&expectedHealthEndpointInstance,
		&expectedFooEndpointInstance,
		&expectedBarEndpointInstance,
	}
	assertInstances(t, expectedClients, expectedEndpoints, instances)
}

func TestDefaultDependency(t *testing.T) {
	moduleSystem := NewModuleSystem(
		map[string][]string{
			"client":     {"clients/*"},
			"endpoint":   {"endpoints/*"},
			"middleware": {"middlewares/*"},
			"service":    {"services/*"},
		},
		map[string][]string{
			"endpoint": {
				"clients/*",
			},
		},
		false,
	)
	var err error

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	})
	if err != nil {
		t.Errorf("Unexpected error registering client class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"http",
		&TestHTTPClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "endpoint",
		NamePlural: "endpoints",
		ClassType:  MultiModule,
		DependsOn:  []string{"client"},
	})
	if err != nil {
		t.Errorf("Unexpected error registering endpoint class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"endpoint",
		"http",
		&TestHTTPEndpointGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	currentDir := getTestDirName()
	testServiceDir := path.Join(currentDir, "test-service")

	instances, err := moduleSystem.GenerateBuild(
		"github.com/uber/zanzibar/codegen/test-service",
		testServiceDir,
		path.Join(testServiceDir, "build"),
		Options{
			CommitChange: true,
		},
	)
	if err != nil {
		t.Errorf("Unexpected generation failure")
	}

	for className, classInstances := range instances {
		if className == "client" {
			expectedLen := 3
			if len(classInstances) != expectedLen {
				t.Errorf(
					"Expected %d client class instance but found %d",
					expectedLen,
					len(classInstances),
				)
			}
		} else if className == "endpoint" {
			expectedLen := 1
			if len(classInstances) != expectedLen {
				t.Errorf(
					"Expected %d endpoint class instance but found %d",
					expectedLen,
					len(classInstances),
				)
			}

			for _, instance := range classInstances {
				expectedLen = 4
				if len(instance.Dependencies) != expectedLen {
					t.Errorf(
						"Expected %s to have %d dependencies but found %d",
						instance.ClassName,
						expectedLen,
						len(instance.Dependencies),
					)
				}
				expectedLen = 3
				if len(instance.ResolvedDependencies["client"]) != expectedLen {
					t.Errorf(
						"Expected %s to have %d resolved dependencies but found %d",
						instance.ClassName,
						expectedLen,
						len(instance.ResolvedDependencies["client"]),
					)
				}
				expectedLen = 3
				if len(instance.RecursiveDependencies["client"]) != expectedLen {
					t.Errorf(
						"Expected %s to have %d recursive dependencies but found %d",
						instance.ClassName,
						expectedLen,
						len(instance.RecursiveDependencies["client"]),
					)
				}
			}
		} else {
			t.Errorf("Unexpected resolved class type %s", className)
		}
	}
}

func TestSingleDefaultDependency(t *testing.T) {
	moduleSystem := NewModuleSystem(
		map[string][]string{
			"client":     {"clients/*"},
			"endpoint":   {"endpoints/*"},
			"middleware": {"middlewares/*"},
			"service":    {"services/*"},
		},
		map[string][]string{
			"endpoint": {
				"clients/example",
			},
		},
		false,
	)
	var err error

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	})
	if err != nil {
		t.Errorf("Unexpected error registering client class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"http",
		&TestHTTPClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "endpoint",
		NamePlural: "endpoints",
		ClassType:  MultiModule,
		DependsOn:  []string{"client"},
	})
	if err != nil {
		t.Errorf("Unexpected error registering endpoint class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"endpoint",
		"http",
		&TestHTTPEndpointGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	currentDir := getTestDirName()
	testServiceDir := path.Join(currentDir, "test-service")

	instances, err := moduleSystem.GenerateBuild(
		"github.com/uber/zanzibar/codegen/test-service",
		testServiceDir,
		path.Join(testServiceDir, "build"),
		Options{
			CommitChange: true,
		},
	)
	if err != nil {
		t.Errorf("Unexpected generation failure")
	}

	for className, classInstances := range instances {
		if className == "client" {
			expectedLen := 3
			if len(classInstances) != expectedLen {
				t.Errorf(
					"Expected %d client class instance but found %d",
					expectedLen,
					len(classInstances),
				)
			}
		} else if className == "endpoint" {
			expectedLen := 1
			if len(classInstances) != expectedLen {
				t.Errorf(
					"Expected %d endpoint class instance but found %d",
					expectedLen,
					len(classInstances),
				)
			}

			for _, instance := range classInstances {
				expectedLen = 2
				if len(instance.Dependencies) != expectedLen {
					t.Errorf(
						"Expected %s to have %d dependencies but found %d",
						instance.ClassName,
						expectedLen,
						len(instance.Dependencies),
					)
				}
				expectedLen = 1
				if len(instance.ResolvedDependencies["client"]) != expectedLen {
					t.Errorf(
						"Expected %s to have %d resolved dependencies but found %d",
						instance.ClassName,
						expectedLen,
						len(instance.ResolvedDependencies["client"]),
					)
				}
				expectedLen = 2
				if len(instance.RecursiveDependencies["client"]) != expectedLen {
					t.Errorf(
						"Expected %s to have %d recursive dependencies but found %d",
						instance.ClassName,
						expectedLen,
						len(instance.RecursiveDependencies["client"]),
					)
				}
			}
		} else {
			t.Errorf("Unexpected resolved class type %s", className)
		}
	}
}

func TestNoClassDefaultDependency(t *testing.T) {
	moduleSystem := NewModuleSystem(
		map[string][]string{
			"client":   {"clients/*"},
			"endpoint": {"endpoints/*"},
		},
		map[string][]string{
			"endpoint": {
				"clients/example",
			},
		},
		false,
	)
	var err error

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	})
	if err != nil {
		t.Errorf("Unexpected error registering client class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"client",
		"http",
		&TestHTTPClientGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "endpoint",
		NamePlural: "endpoints",
		ClassType:  MultiModule,
	})
	if err != nil {
		t.Errorf("Unexpected error registering endpoint class: %s", err)
	}

	err = moduleSystem.RegisterClassType(
		"endpoint",
		"http",
		&TestHTTPEndpointGenerator{},
	)
	if err != nil {
		t.Errorf("Unexpected error registering http client class type: %s", err)
	}

	currentDir := getTestDirName()
	testServiceDir := path.Join(currentDir, "test-service")

	_, err = moduleSystem.GenerateBuild(
		"github.com/uber/zanzibar/codegen/test-service",
		testServiceDir,
		path.Join(testServiceDir, "build"),
		Options{
			CommitChange: true,
		},
	)
	if err == nil {
		t.Errorf("Expected failure due to default dependency directory which is not a dependency")
	}
}

// func TestExampleServiceCycles(t *testing.T) {
// 	moduleSystem := NewModuleSystem(
// 		map[string][]string{
// 			"client":   {"clients/*"},
// 			"endpoint": {"endpoints/*"},
// 		},
// 		map[string][]string{},
// 	)
// 	var err error
//
// 	err = moduleSystem.RegisterClass(ModuleClass{
// 		Name:       "client",
// 		NamePlural: "clients",
// 		ClassType:  MultiModule,
// 	})
// 	if err != nil {
// 		t.Errorf("Unexpected error registering client class: %s", err)
// 	}
//
// 	err = moduleSystem.RegisterClassType(
// 		"client",
// 		"http",
// 		&TestHTTPClientGenerator{},
// 	)
// 	if err != nil {
// 		t.Errorf("Unexpected error registering http client class type: %s", err)
// 	}
//
// 	err = moduleSystem.RegisterClass(ModuleClass{
// 		Name:       "endpoint",
// 		NamePlural: "endpoints",
// 		ClassType:  MultiModule,
// 		DependsOn:  []string{"client"},
// 	})
// 	if err != nil {
// 		t.Errorf("Unexpected error registering endpoint class: %s", err)
// 	}
//
// 	err = moduleSystem.RegisterClassType(
// 		"endpoint",
// 		"http",
// 		&TestHTTPEndpointGenerator{},
// 	)
// 	if err != nil {
// 		t.Errorf("Unexpected error registering http client class type: %s", err)
// 	}
//
// 	currentDir := getTestDirName()
// 	testServiceDir := path.Join(currentDir, "test-service-cycle")
//
// 	_, err = moduleSystem.GenerateBuild(
// 		"github.com/uber/zanzibar/codegen/test-service",
// 		testServiceDir,
// 		path.Join(testServiceDir, "build"),
// 		true,
// 	)
// 	if err == nil {
// 		t.Errorf("Expected cycle error generating build")
// 	}
//
// 	if err.Error() != "Dependency cycle: example-a cannot be initialized before example-b" {
// 		t.Errorf("Expected error due to dependency cycle, received: %s", err)
// 	}
// }

func TestSortDependencies(t *testing.T) {
	testInstanceA := createTestInstance("example-a")
	testInstanceB := createTestInstance("example-b")
	testInstanceC := createTestInstance("example-c", testInstanceB)
	testInstanceD := createTestInstance("example-d", testInstanceC)
	testInstanceE := createTestInstance("example-e")
	testInstanceF := createTestInstance("example-f")
	testInstanceG := createTestInstance("example-g", testInstanceF)

	if !peerDepends(testInstanceC, testInstanceB) {
		t.Errorf("Expected test instance c to depend on test instance b")
	}
	if !peerDepends(testInstanceD, testInstanceC) {
		t.Errorf("Expected test instance d to depend on test instance c")
	}
	if !peerDepends(testInstanceG, testInstanceF) {
		t.Errorf("Expected test instance g to depend on test instance f")
	}

	permutations := generatePermutations([]*ModuleInstance{
		testInstanceA,
		testInstanceB,
		testInstanceC,
		testInstanceD,
		testInstanceE,
		testInstanceF,
		testInstanceG,
	})

	for _, classList := range permutations {
		sortedList, err := sortDependencyList("client", classList)

		if err != nil {
			t.Errorf("Unexpected error when sorting peer list: %s", err)
		}

		if !isBefore(sortedList, "example-b", "example-c") {
			t.Errorf("Expected example-b to be before example-c in sorted list")
		}

		if !isBefore(sortedList, "example-c", "example-d") {
			t.Errorf("Expected example-c to be before example-d in sorted list")
		}

		if !isBefore(sortedList, "example-f", "example-g") {
			t.Errorf("Expected example-f to be before example-g in sorted list")
		}
	}

}

func TestSortModuleClasses(t *testing.T) {
	ms := NewModuleSystem(map[string][]string{}, map[string][]string{}, false)
	err := ms.RegisterClass(ModuleClass{
		Name:       "a",
		NamePlural: "as",
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "b",
		NamePlural: "bs",
		DependsOn:  []string{"a"},
		DependedBy: []string{"c"},
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "c",
		NamePlural: "cs",
		DependsOn:  []string{"b"},
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "d",
		NamePlural: "ds",
		DependsOn:  []string{"a"},
		DependedBy: []string{"c"},
	})
	assert.NoError(t, err)
	expected := []string{"a", "b", "d", "c"}
	err = ms.resolveClassOrder()
	assert.NoError(t, err)
	assert.Equal(t, expected, ms.classOrder)
}

func TestSortModuleClassesNoDeps(t *testing.T) {
	ms := NewModuleSystem(map[string][]string{}, map[string][]string{}, false)
	err := ms.RegisterClass(ModuleClass{
		Name:       "a",
		NamePlural: "as",
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "b",
		NamePlural: "bs",
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "c",
		NamePlural: "cs",
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "d",
		NamePlural: "ds",
	})
	assert.NoError(t, err)
	expected := []string{"a", "b", "c", "d"}
	err = ms.resolveClassOrder()
	assert.NoError(t, err)
	assert.Equal(t, expected, ms.classOrder)
}

func TestSortModuleClassesUndefined(t *testing.T) {
	ms := NewModuleSystem(map[string][]string{}, map[string][]string{}, false)
	err := ms.RegisterClass(ModuleClass{
		Name:       "a",
		NamePlural: "as",
		DependsOn:  []string{"c"},
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "b",
		NamePlural: "bs",
		DependsOn:  []string{"a"},
	})
	assert.NoError(t, err)
	err = ms.resolveClassOrder()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module class \"a\" depends on \"c\" which is not defined")
}

func TestSortModuleClassesUndefined2(t *testing.T) {
	ms := NewModuleSystem(map[string][]string{}, map[string][]string{}, false)
	err := ms.RegisterClass(ModuleClass{
		Name:       "a",
		NamePlural: "as",
		DependedBy: []string{"c"},
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "b",
		NamePlural: "bs",
		DependedBy: []string{"a"},
	})
	assert.NoError(t, err)
	err = ms.resolveClassOrder()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module class \"a\" is depended by \"c\" which is not defined")
}

func TestSortableModuleClassCycle(t *testing.T) {
	ms := NewModuleSystem(map[string][]string{}, map[string][]string{}, false)
	err := ms.RegisterClass(ModuleClass{
		Name:       "a",
		NamePlural: "as",
		DependsOn:  []string{"b"},
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "b",
		NamePlural: "bs",
		DependsOn:  []string{"a"},
	})
	assert.NoError(t, err)

	err = ms.resolveClassOrder()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dependency cycle detected for module class \"a\": a->b")
}

func TestSortableModuleClassCycle2(t *testing.T) {
	ms := NewModuleSystem(map[string][]string{}, map[string][]string{}, false)
	err := ms.RegisterClass(ModuleClass{
		Name:       "a",
		NamePlural: "as",
		DependedBy: []string{"b"},
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "b",
		NamePlural: "bs",
		DependedBy: []string{"a"},
	})
	assert.NoError(t, err)

	err = ms.resolveClassOrder()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dependency cycle detected for module class \"a\": a->b")
}

func TestSortModuleClassesIndirectCycle(t *testing.T) {
	ms := NewModuleSystem(map[string][]string{}, map[string][]string{}, false)
	err := ms.RegisterClass(ModuleClass{
		Name:       "a",
		NamePlural: "as",
		DependsOn:  []string{"b"},
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "b",
		NamePlural: "bs",
		DependsOn:  []string{"c"},
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "c",
		NamePlural: "cs",
		DependsOn:  []string{"a"},
	})
	assert.NoError(t, err)
	err = ms.resolveClassOrder()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dependency cycle detected for module class \"a\": a->b->c->a")
}

func TestSortModuleClassesIndirectCycle2(t *testing.T) {
	ms := NewModuleSystem(map[string][]string{}, map[string][]string{}, false)
	err := ms.RegisterClass(ModuleClass{
		Name:       "a",
		NamePlural: "as",
		DependedBy: []string{"b"},
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "b",
		NamePlural: "bs",
		DependedBy: []string{"c"},
	})
	assert.NoError(t, err)
	err = ms.RegisterClass(ModuleClass{
		Name:       "c",
		NamePlural: "cs",
		DependedBy: []string{"a"},
	})
	assert.NoError(t, err)

	err = ms.resolveClassOrder()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dependency cycle detected for module class \"a\": a->c->b->a")
}

func createTestInstance(name string, depInstances ...*ModuleInstance) *ModuleInstance {
	dependencies := []ModuleDependency{}
	resolvedDependencies := map[string][]*ModuleInstance{}
	recursiveDependencies := map[string][]*ModuleInstance{}

	for _, dep := range depInstances {
		dependencies = append(dependencies, ModuleDependency{
			ClassName:    dep.ClassName,
			InstanceName: dep.InstanceName,
		})

		resolvedList, ok := resolvedDependencies[dep.ClassName]

		if !ok {
			resolvedList = []*ModuleInstance{}
		}

		resolvedDependencies[dep.ClassName] = appendUniqueModule(
			resolvedList,
			dep,
		)

		recursiveDependencies[dep.ClassName] =
			resolvedDependencies[dep.ClassName]

		for className, recursiveDepList := range dep.RecursiveDependencies {
			for _, recursiveDep := range recursiveDepList {
				recursiveList := recursiveDependencies[className]

				if recursiveList == nil {
					recursiveList = []*ModuleInstance{}
				}

				recursiveDependencies[className] = appendUniqueModule(
					recursiveList,
					recursiveDep,
				)
			}
		}
	}

	return &ModuleInstance{
		ClassName:    "client",
		ClassType:    "http",
		Directory:    "clients/example",
		InstanceName: name,
		YAMLFileName: "client-config.yaml",
		PackageInfo: &PackageInfo{
			ExportName:            "NewClient",
			ExportType:            "Client",
			GeneratedPackageAlias: "exampleClientGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/clients/example",
			IsExportGenerated:     true,
			PackageAlias:          "exampleClientStatic",
			PackageName:           "exampleClient",
			PackagePath:           "github.com/uber/zanzibar/codegen/test-service/clients/example",
		},
		Dependencies:          dependencies,
		ResolvedDependencies:  resolvedDependencies,
		RecursiveDependencies: recursiveDependencies,
	}
}

func generatePermutations(instances []*ModuleInstance) [][]*ModuleInstance {
	results := [][]*ModuleInstance{}

	if len(instances) <= 1 {
		return [][]*ModuleInstance{
			instances,
		}
	}

	for i := 0; i < len(instances); i++ {
		instanceCopy := make([]*ModuleInstance, len(instances))
		copy(instanceCopy, instances)
		instanceCopy[i], instanceCopy[0] = instanceCopy[0], instanceCopy[i]
		permutations := generatePermutations(instanceCopy[1:])

		for _, permutation := range permutations {
			results = append(results, append(permutation, instanceCopy[0]))
		}
	}
	return results
}

// Returns true of module with instance name a comes before instance name b
func isBefore(sortedList []*ModuleInstance, a string, b string) bool {
	hasSeenA := false
	for _, module := range sortedList {
		if module.InstanceName == a {
			hasSeenA = true
		}

		if module.InstanceName == b {
			return hasSeenA
		}
	}
	return false
}

func getTestDirName() string {
	_, file, _, _ := runtime.Caller(0)
	dirname := filepath.Dir(file)
	// Strip _obj dirs generated by test -cover ...
	if filepath.Base(dirname) == "_obj" {
		dirname = filepath.Dir(dirname)
	}
	// if absolute then fini.
	if filepath.IsAbs(dirname) {
		return dirname
	}
	// If dirname is not absolute then its a package name...
	return filepath.Join(os.Getenv("GOPATH"), "src", dirname)
}

func compareInstances(
	t *testing.T,
	actual *ModuleInstance,
	expectedList []*ModuleInstance,
) {
	expectedMap := make(map[string]*ModuleInstance)
	for _, expected := range expectedList {
		expectedMap[expected.InstanceName] = expected
	}
	expected := expectedMap[actual.InstanceName]

	assert.Equal(
		t,
		expected.QPSLevels,
		actual.QPSLevels,
		fmt.Sprintf("Expected qps levels %v but found %v", expected.QPSLevels, actual.QPSLevels),
	)

	if actual.ClassName != expected.ClassName {
		t.Errorf(
			"Expected class name of %q %q to be %q but found %q",
			expected.ClassName,
			expected.InstanceName,
			expected.ClassName,
			actual.ClassName,
		)
	}

	if actual.BaseDirectory != expected.BaseDirectory {
		t.Errorf(
			"Expected %q %q base directory to be %q but found %q",
			expected.ClassName,
			expected.InstanceName,
			expected.BaseDirectory,
			actual.BaseDirectory,
		)
	}

	if actual.ClassType != expected.ClassType {
		t.Errorf(
			"Expected %q %q class type to be %q but found %q",
			expected.ClassName,
			expected.InstanceName,
			expected.ClassName,
			actual.ClassName,
		)
	}

	if len(actual.Dependencies) != len(expected.Dependencies) {
		t.Errorf(
			"Expected %q %q to have %d dependencies but found %d",
			expected.ClassName,
			expected.InstanceName,
			len(expected.Dependencies),
			len(actual.Dependencies),
		)
	}

	for di, expectedDependency := range expected.Dependencies {
		actualDependency := actual.Dependencies[di]

		if actualDependency.ClassName != expectedDependency.ClassName {
			t.Errorf(
				"Expected %q %q dependency %d class name to be %s but found %s",
				expected.ClassName,
				expected.InstanceName,
				di,
				expectedDependency.ClassName,
				actualDependency.ClassName,
			)
		}

		if actualDependency.InstanceName != expectedDependency.InstanceName {
			t.Errorf(
				"Expected %q %q dependency %d instance name to be %s but found %s",
				expected.ClassName,
				expected.InstanceName,
				di,
				expectedDependency.InstanceName,
				actualDependency.InstanceName,
			)
		}
	}

	if actual.Directory != expected.Directory {
		t.Errorf(
			"Expected %q %q directory to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			expected.Directory,
			actual.Directory,
		)
	}

	if actual.InstanceName != expected.InstanceName {
		t.Errorf(
			"Expected %q %q instance name to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			expected.InstanceName,
			actual.InstanceName,
		)
	}

	if actual.YAMLFileName != expected.YAMLFileName {
		t.Errorf(
			"Expected %q %q yaml file name to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			expected.YAMLFileName,
			actual.YAMLFileName,
		)
	}

	if actual.JSONFileName != expected.JSONFileName {
		t.Errorf(
			"Expected %q %q json file name to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			expected.JSONFileName,
			actual.JSONFileName,
		)
	}

	if actual.PackageInfo.ExportName != expected.PackageInfo.ExportName {
		t.Errorf(
			"Expected %q %q package export name to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			expected.PackageInfo.ExportName,
			actual.PackageInfo.ExportName,
		)
	}

	if actual.PackageInfo.ExportType != expected.PackageInfo.ExportType {
		t.Errorf(
			"Expected %q %q package export type to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			expected.PackageInfo.ExportType,
			actual.PackageInfo.ExportType,
		)
	}

	if actual.PackageInfo.GeneratedPackageAlias != expected.PackageInfo.GeneratedPackageAlias {
		t.Errorf(
			"Expected %q %q generated package alias to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			expected.PackageInfo.GeneratedPackageAlias,
			actual.PackageInfo.GeneratedPackageAlias,
		)
	}

	if actual.PackageInfo.GeneratedPackagePath != expected.PackageInfo.GeneratedPackagePath {
		t.Errorf(
			"Expected %q %q generated package path to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			expected.PackageInfo.GeneratedPackagePath,
			actual.PackageInfo.GeneratedPackagePath,
		)
	}

	if actual.PackageInfo.IsExportGenerated != expected.PackageInfo.IsExportGenerated {
		t.Errorf(
			"Expected %q %q IsExportGenerated to be %t but found %t",
			expected.ClassName,
			expected.InstanceName,
			expected.PackageInfo.IsExportGenerated,
			actual.PackageInfo.IsExportGenerated,
		)
	}

	if actual.PackageInfo.PackageAlias != expected.PackageInfo.PackageAlias {
		t.Errorf(
			"Expected %q %q package alias to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			expected.PackageInfo.PackageAlias,
			actual.PackageInfo.PackageAlias,
		)
	}

	if actual.PackageInfo.PackageName != expected.PackageInfo.PackageName {
		t.Errorf(
			"Expected %q %q package name to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			expected.PackageInfo.PackageName,
			actual.PackageInfo.PackageName,
		)
	}

	if actual.PackageInfo.PackagePath != expected.PackageInfo.PackagePath {
		t.Errorf(
			"Expected %q %q package path to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			expected.PackageInfo.PackagePath,
			actual.PackageInfo.PackagePath,
		)
	}

	if actual.PackageInfo.ExportName != expected.PackageInfo.ExportName {
		t.Errorf(
			"Expected %q %q package export name to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			expected.PackageInfo.ExportName,
			actual.PackageInfo.ExportName,
		)
	}

	if len(expected.ResolvedDependencies) != len(actual.ResolvedDependencies) {
		t.Errorf(
			"Expected %q %q to have %d resolved dependency class names but found %d",
			expected.ClassName,
			expected.InstanceName,
			len(expected.ResolvedDependencies),
			len(actual.ResolvedDependencies),
		)
	}

	for className, expectedDeps := range expected.ResolvedDependencies {
		actualDeps := actual.ResolvedDependencies[className]

		if len(actualDeps) != len(expectedDeps) {
			t.Errorf(
				"Expected %s %s to have %d resolved %s dependencies but found %d",
				expected.InstanceName,
				expected.ClassName,
				len(expected.Dependencies),
				className,
				len(actual.Dependencies),
			)
		}

		for i, expectedDependency := range expectedDeps {
			actualDependency := actualDeps[i]
			if actualDependency.ClassName != expectedDependency.ClassName ||
				actualDependency.InstanceName != expectedDependency.InstanceName {
				t.Errorf(
					"Expected %s %s to have %s %s as the resolved %s dependency at index %d, but found %s %s",
					expected.InstanceName,
					expected.ClassName,
					expectedDependency.InstanceName,
					expectedDependency.ClassName,
					className,
					i,
					actualDependency.InstanceName,
					actualDependency.ClassName,
				)
			}
		}
	}

	if len(expected.RecursiveDependencies) != len(actual.RecursiveDependencies) {
		t.Errorf(
			"Expected %s %s to have %d recursive dependency class names but found %d",
			expected.InstanceName,
			expected.ClassName,
			len(expected.ResolvedDependencies),
			len(actual.ResolvedDependencies),
		)
	}

	for className, expectedDeps := range expected.RecursiveDependencies {
		actualDeps := actual.RecursiveDependencies[className]

		if len(actualDeps) != len(expectedDeps) {
			t.Errorf(
				"Expected %s %s to have %d recursive %s dependencies but found %d",
				expected.InstanceName,
				expected.ClassName,
				len(expectedDeps),
				className,
				len(actualDeps),
			)
		}

		for i, expectedDependency := range expectedDeps {
			actualDependency := actualDeps[i]
			if actualDependency.ClassName != expectedDependency.ClassName ||
				actualDependency.InstanceName != expectedDependency.InstanceName {
				t.Errorf(
					"Expected %s %s to have %s %s as the recursive %s dependency at index %d, but found %s %s",
					expected.InstanceName,
					expected.ClassName,
					expectedDependency.InstanceName,
					expectedDependency.ClassName,
					className,
					i,
					actualDependency.InstanceName,
					actualDependency.ClassName,
				)
			}
		}
	}
}

func TestModulePathStripping(t *testing.T) {
	assert.Equal(t, "foo", stripModuleClassName("clients", "clients/foo"))
	assert.Equal(t, "a/b/d", stripModuleClassName("c", "a/b/c/d"))
	assert.Equal(t, "tchannel/foo", stripModuleClassName("endpoints", "endpoints/tchannel/foo"))
}

func TestModuleSearchDuplicateGlobs(t *testing.T) {
	moduleSystem := NewModuleSystem(
		map[string][]string{"client": {"clients/*", "clients/*"}},
		map[string][]string{},
		false,
	)
	var err error

	err = moduleSystem.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	})
	assert.NoError(t, err)

	err = moduleSystem.RegisterClassType(
		"client",
		"http",
		&TestHTTPClientGenerator{},
	)
	assert.NoError(t, err)

	currentDir := getTestDirName()
	testServiceDir := path.Join(currentDir, "test-service")

	instances, err := moduleSystem.GenerateBuild(
		"github.com/uber/zanzibar/codegen/test-service",
		testServiceDir,
		path.Join(testServiceDir, "build"),
		Options{
			CommitChange: false,
		},
	)
	assert.NoError(t, err)

	var instance *ModuleInstance
	for _, v := range instances["client"] {
		if v.InstanceName == "example" {
			instance = v
		}
	}
	assert.Equal(t, []string{"client"}, instance.DependencyOrder)
}

func TestTransitiveSimple(t *testing.T) {
	ls := &ModuleInstance{
		InstanceName: "location-store",
		ClassName:    "client",
	}
	endpoint := &ModuleInstance{
		InstanceName: "getLocation",
		ClassName:    "endpoint",
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {ls},
		},
	}
	service := &ModuleInstance{
		InstanceName: "edge-gateway",
		ClassName:    "service",
		RecursiveDependencies: map[string][]*ModuleInstance{
			"endpoint": {endpoint},
		},
	}

	graph := map[string][]*ModuleInstance{
		"client":   {ls},
		"endpoint": {endpoint},
		"service":  {service},
	}

	ms := &ModuleSystem{
		classOrder: []string{"client", "endpoint", "service"},
	}
	results, err := ms.collectTransitiveDependencies([]ModuleDependency{
		{
			InstanceName: "location-store",
			ClassName:    "client",
		},
	}, graph)
	assert.NoError(t, err)

	t.Logf("%+v", results)

	assert.Len(t, results["client"], 1)
	assert.Equal(t, "location-store", results["client"][0].InstanceName)
	assert.Len(t, results["endpoint"], 1)
	assert.Equal(t, "getLocation", results["endpoint"][0].InstanceName)
	assert.Len(t, results["service"], 1)
	assert.Equal(t, "edge-gateway", results["service"][0].InstanceName)
}

func TestTransitiveMultipleClients(t *testing.T) {
	pp := &ModuleInstance{
		InstanceName: "passport",
		ClassName:    "client",
	}
	ls := &ModuleInstance{
		InstanceName: "location-store",
		ClassName:    "client",
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {pp},
		},
	}
	endpoint := &ModuleInstance{
		InstanceName: "getLocation",
		ClassName:    "endpoint",
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {ls},
		},
	}
	service := &ModuleInstance{
		InstanceName: "edge-gateway",
		ClassName:    "service",
		RecursiveDependencies: map[string][]*ModuleInstance{
			"endpoint": {endpoint},
		},
	}

	graph := map[string][]*ModuleInstance{
		"client":   {ls, pp},
		"endpoint": {endpoint},
		"service":  {service},
	}

	ms := &ModuleSystem{
		classOrder: []string{"client", "endpoint", "service"},
	}
	results, err := ms.collectTransitiveDependencies([]ModuleDependency{
		{
			InstanceName: "passport",
			ClassName:    "client",
		},
	}, graph)
	assert.NoError(t, err)

	t.Logf("%+v", results)

	assert.Len(t, results["client"], 2)
	// map iteration order is undefined
	clients := make(map[string]*ModuleInstance)
	for _, instance := range results["client"] {
		clients[instance.InstanceName] = instance
	}
	assert.NotNil(t, clients["location-store"])
	assert.NotNil(t, clients["passport"])
	assert.Len(t, results["endpoint"], 1)
	assert.Equal(t, "getLocation", results["endpoint"][0].InstanceName)
	assert.Len(t, results["service"], 1)
	assert.Equal(t, "edge-gateway", results["service"][0].InstanceName)
}

func TestTransitiveDoesntBuildUnrelated(t *testing.T) {
	ls := &ModuleInstance{
		InstanceName: "location-store",
		ClassName:    "client",
	}
	endpoint := &ModuleInstance{
		InstanceName: "getLocation",
		ClassName:    "endpoint",
		RecursiveDependencies: map[string][]*ModuleInstance{
			"client": {ls},
		},
	}

	unused := &ModuleInstance{
		InstanceName: "getLocation",
		ClassName:    "endpoint",
	}
	service := &ModuleInstance{
		InstanceName: "edge-gateway",
		ClassName:    "service",
		RecursiveDependencies: map[string][]*ModuleInstance{
			"endpoint": {endpoint},
		},
	}

	graph := map[string][]*ModuleInstance{
		"client":   {ls},
		"endpoint": {endpoint, unused},
		"service":  {service},
	}

	ms := &ModuleSystem{
		classOrder: []string{"client", "endpoint", "service"},
	}
	results, err := ms.collectTransitiveDependencies([]ModuleDependency{
		{
			InstanceName: "location-store",
			ClassName:    "client",
		},
	}, graph)
	assert.NoError(t, err)

	t.Logf("%+v", results)

	assert.Len(t, results["endpoint"], 1)
	assert.Equal(t, "getLocation", results["endpoint"][0].InstanceName)
}

func TestDisabledPopulateQPSLevel(t *testing.T) {
	qpsLevels, err := PopulateQPSLevels("endpoint-path", false)
	assert.Nil(t, err)
	assert.Equal(t, map[string]int{}, qpsLevels)
}

// TODO: refactor other tests to call this to register classes
func registerClass(t *testing.T, m *ModuleSystem, name, pluralName string, moduleClassType moduleClassType, dependencies []string) {
	if err := m.RegisterClass(ModuleClass{
		Name:       name,
		NamePlural: pluralName,
		ClassType:  moduleClassType,
		DependsOn:  dependencies,
	}); err != nil {
		t.Errorf("Unexpected error registering %s class: %s", name, err)
	}
}

// TODO: refactor other tests to call this to register class types
func registerClassType(t *testing.T, m *ModuleSystem, className, classType string, generator BuildGenerator) {
	if err := m.RegisterClassType(className, classType, generator); err != nil {
		t.Errorf("Unexpected error registering %s %s class type %s", className, classType, err)
	}
}

func assertInstances(t *testing.T, expectedClients, expectedEndpoints []*ModuleInstance, instances map[string][]*ModuleInstance) {
	for className, classInstances := range instances {
		sort.Slice(classInstances, func(i, j int) bool {
			return classInstances[i].InstanceName < classInstances[j].InstanceName
		})

		sort.Slice(expectedEndpoints, func(i, j int) bool {
			return expectedEndpoints[i].InstanceName < expectedEndpoints[j].InstanceName
		})

		sort.Slice(expectedClients, func(i, j int) bool {
			return expectedClients[i].InstanceName < expectedClients[j].InstanceName
		})

		if className == "client" {
			if len(classInstances) != len(expectedClients) {
				t.Errorf(
					"Expected %d client class instance but found %d",
					len(expectedClients),
					len(classInstances),
				)
			}

			for _, instance := range expectedClients {
				compareInstances(t, instance, expectedClients)
			}
		} else if className == "endpoint" {
			if len(classInstances) != len(expectedEndpoints) {
				t.Errorf(
					"Expected %d endpoint class instance but found %d",
					len(expectedEndpoints),
					len(classInstances),
				)
			}

			for _, instance := range classInstances {
				compareInstances(t, instance, expectedEndpoints)

				clientDependency := instance.ResolvedDependencies["client"][0]
				clientSpec, ok := clientDependency.GeneratedSpec().(*TestClientSpec)
				if !ok {
					t.Errorf("type casting failed %s\n", clientDependency)
				}

				if clientSpec.Info != instance.ClassType {
					t.Errorf(
						"Expected client spec info on generated client spec",
					)
				}
			}
		} else if className == "service" {
			if len(classInstances) != 1 {
				t.Errorf("Expected 1 service class instance but found %d", len(classInstances))
			}
		} else {
			t.Errorf("Unexpected resolved class type %s", className)
		}
	}
}
