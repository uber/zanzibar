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
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
)

type handler struct{}

var testHandler = handler{}
var staticHandler = handler{}
var variableHandler = handler{}
var splatHandler = handler{}

type TestClientSpec struct {
	Info string
}

type TestHTTPClientGenerator struct{}

func (*TestHTTPClientGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	return &BuildResult{
		Spec: &TestClientSpec{
			Info: "http",
		},
	}, nil
}

type TestTChannelClientGenerator struct{}

func (*TestTChannelClientGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	return &BuildResult{
		Spec: &TestClientSpec{
			Info: "tchannel",
		},
	}, nil
}

type TestHTTPEndpointGenerator struct{}

func (*TestHTTPEndpointGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	return nil, nil
}

func TestExampleService(t *testing.T) {
	moduleSystem := NewModuleSystem()
	var err error

	err = moduleSystem.RegisterClass("client", ModuleClass{
		ClassType: MultiModule,
		Directory: "clients",
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

	err = moduleSystem.RegisterClass("endpoint", ModuleClass{
		ClassType:         MultiModule,
		ClassDependencies: []string{"client"},
		Directory:         "endpoints",
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

	err = moduleSystem.RegisterClass("client", ModuleClass{
		ClassType: MultiModule,
		Directory: "clients",
	})
	if err == nil {
		t.Errorf("Expected double definition of client class to error")
	}

	err = moduleSystem.RegisterClass("newclient", ModuleClass{
		ClassType: MultiModule,
		Directory: "./clients/",
	})
	if err == nil {
		t.Errorf("Expected registering a module in the same directory to throw")
	}

	err = moduleSystem.RegisterClass("newclient", ModuleClass{
		ClassType: MultiModule,
		Directory: "./clients/../../../foo",
	})
	if err == nil {
		t.Errorf("Expected registering a module in an external directory to throw")
	}

	currentDir := getTestDirName()
	testServiceDir := path.Join(currentDir, "test-service")

	// TODO: this should return a collection of errors if they occur
	instances, err := moduleSystem.GenerateBuild(
		"github.com/uber/zanzibar/codegen/test-service",
		testServiceDir,
		path.Join(testServiceDir, "build"),
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
		JSONFileName:  "client-config.json",
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
		JSONFileName:  "client-config.json",
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

	expectedEndpointInstance := ModuleInstance{
		BaseDirectory: testServiceDir,
		ClassName:     "endpoint",
		ClassType:     "http",
		Directory:     "endpoints/health",
		InstanceName:  "health",
		JSONFileName:  "endpoint-config.json",
		PackageInfo: &PackageInfo{
			ExportName:            "NewEndpoint",
			ExportType:            "Endpoint",
			GeneratedPackageAlias: "healthEndpointGenerated",
			GeneratedPackagePath:  "github.com/uber/zanzibar/codegen/test-service/build/endpoints/health",
			IsExportGenerated:     true,
			PackageAlias:          "healthEndpointStatic",
			PackageName:           "healthEndpoint",
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

	for className, classInstances := range instances {
		if className == "client" {
			if len(classInstances) != 2 {
				t.Errorf(
					"Expected 2 client class instance but found %d",
					len(classInstances),
				)
			}

			compareInstances(t, classInstances[0], &expectedClientInstance)
			compareInstances(t, classInstances[1], &expectedClientDependency)
		} else if className == "endpoint" {
			if len(classInstances) != 1 {
				t.Errorf(
					"Expected 1 endpoint class instance but found %d",
					len(classInstances),
				)
			}

			i := classInstances[0]

			compareInstances(t, i, &expectedEndpointInstance)

			clientDependency := i.ResolvedDependencies["client"][0]
			clientSpec := clientDependency.GeneratedSpec().(*TestClientSpec)

			if clientSpec.Info != i.ClassType {
				t.Errorf(
					"Expected client spec info on generated client spec",
				)
			}
		} else {
			t.Errorf("Unexpected resolved class type %s", className)
		}
	}
}

func TestExampleServiceCycles(t *testing.T) {
	moduleSystem := NewModuleSystem()
	var err error

	err = moduleSystem.RegisterClass("client", ModuleClass{
		ClassType: MultiModule,
		Directory: "clients",
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

	err = moduleSystem.RegisterClass("endpoint", ModuleClass{
		ClassType:         MultiModule,
		ClassDependencies: []string{"client"},
		Directory:         "endpoints",
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
	testServiceDir := path.Join(currentDir, "test-service-cycle")

	_, err = moduleSystem.GenerateBuild(
		"github.com/uber/zanzibar/codegen/test-service",
		testServiceDir,
		path.Join(testServiceDir, "build"),
	)
	if err == nil {
		t.Errorf("Expected cycle error generating build")
	}

	if err.Error() != "Dependency cycle: example-a cannot be initialized before example-b" {
		t.Errorf("Expected error due to dependency cycle, received: %s", err)
	}
}

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
		JSONFileName: "client-config.json",
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
	expected *ModuleInstance,
) {
	if actual.ClassName != expected.ClassName {
		t.Errorf(
			"Expected class name to be %s but found %s",
			expected.ClassName,
			actual.ClassName,
		)
	}

	if actual.BaseDirectory != expected.BaseDirectory {
		t.Errorf(
			"Expected %s base directory to be %s but found %s",
			expected.ClassName,
			expected.BaseDirectory,
			actual.BaseDirectory,
		)
	}

	if actual.ClassType != expected.ClassType {
		t.Errorf(
			"Expected %s class type to be %s but found %s",
			expected.ClassName,
			expected.ClassType,
			actual.ClassType,
		)
	}

	if len(actual.Dependencies) != len(expected.Dependencies) {
		t.Errorf(
			"Expected %s to have %d dependencies but found %d",
			expected.ClassName,
			len(expected.Dependencies),
			len(actual.Dependencies),
		)
	}

	for di, expectedDependency := range expected.Dependencies {
		actualDependency := actual.Dependencies[di]

		if actualDependency.ClassName != expectedDependency.ClassName {
			t.Errorf(
				"Expected %s dependency %d class name to be %s but found %s",
				expected.ClassName,
				di,
				expectedDependency.ClassName,
				actualDependency.ClassName,
			)
		}

		if actualDependency.InstanceName != expectedDependency.InstanceName {
			t.Errorf(
				"Expected %s dependency %d instance name to be %s but found %s",
				expected.InstanceName,
				di,
				expectedDependency.InstanceName,
				actualDependency.InstanceName,
			)
		}
	}

	if actual.Directory != expected.Directory {
		t.Errorf(
			"Expected %s directory to be %s but found %s",
			expected.ClassName,
			expected.Directory,
			actual.Directory,
		)
	}

	if actual.InstanceName != expected.InstanceName {
		t.Errorf(
			"Expected %s instance name to be %s but found %s",
			expected.ClassName,
			expected.InstanceName,
			actual.InstanceName,
		)
	}

	if actual.JSONFileName != expected.JSONFileName {
		t.Errorf(
			"Expected %s json file name to be %s but found %s",
			expected.ClassName,
			expected.JSONFileName,
			actual.JSONFileName,
		)
	}

	if actual.PackageInfo.ExportName != expected.PackageInfo.ExportName {
		t.Errorf(
			"Expected %s package export name to be %s but found %s",
			expected.ClassName,
			expected.PackageInfo.ExportName,
			actual.PackageInfo.ExportName,
		)
	}

	if actual.PackageInfo.ExportType != expected.PackageInfo.ExportType {
		t.Errorf(
			"Expected %s package export type to be %s but found %s",
			expected.ClassName,
			expected.PackageInfo.ExportType,
			actual.PackageInfo.ExportType,
		)
	}

	if actual.PackageInfo.GeneratedPackageAlias != expected.PackageInfo.GeneratedPackageAlias {
		t.Errorf(
			"Expected %s generated package alias to be %s but found %s",
			expected.ClassName,
			expected.PackageInfo.GeneratedPackageAlias,
			actual.PackageInfo.GeneratedPackageAlias,
		)
	}

	if actual.PackageInfo.GeneratedPackagePath != expected.PackageInfo.GeneratedPackagePath {
		t.Errorf(
			"Expected %s generated package path to be %s but found %s",
			expected.ClassName,
			expected.PackageInfo.GeneratedPackagePath,
			actual.PackageInfo.GeneratedPackagePath,
		)
	}

	if actual.PackageInfo.IsExportGenerated != expected.PackageInfo.IsExportGenerated {
		t.Errorf(
			"Expected %s IsExportGenerated to be %t but found %t",
			expected.ClassName,
			expected.PackageInfo.IsExportGenerated,
			actual.PackageInfo.IsExportGenerated,
		)
	}

	if actual.PackageInfo.PackageAlias != expected.PackageInfo.PackageAlias {
		t.Errorf(
			"Expected %s package alias to be %s but found %s",
			expected.ClassName,
			expected.PackageInfo.PackageAlias,
			actual.PackageInfo.PackageAlias,
		)
	}

	if actual.PackageInfo.PackageName != expected.PackageInfo.PackageName {
		t.Errorf(
			"Expected %s package name to be %s but found %s",
			expected.ClassName,
			expected.PackageInfo.PackageName,
			actual.PackageInfo.PackageName,
		)
	}

	if actual.PackageInfo.PackagePath != expected.PackageInfo.PackagePath {
		t.Errorf(
			"Expected %s package path to be %s but found %s",
			expected.ClassName,
			expected.PackageInfo.PackagePath,
			actual.PackageInfo.PackagePath,
		)
	}

	if actual.PackageInfo.ExportName != expected.PackageInfo.ExportName {
		t.Errorf(
			"Expected %s package export name to be %s but found %s",
			expected.ClassName,
			expected.PackageInfo.ExportName,
			actual.PackageInfo.ExportName,
		)
	}

	if len(expected.ResolvedDependencies) != len(actual.ResolvedDependencies) {
		t.Errorf(
			"Expected %s %s to have %d resolved dependency class names but found %d",
			expected.InstanceName,
			expected.ClassName,
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
