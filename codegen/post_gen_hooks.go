// Copyright (c) 2019 Uber Technologies, Inc.
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
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
)

const (
	clientInterface = "Client"
	custom          = "custom"
)

type mockableClient struct {
	ClassConfigBase `yaml:",inline" json:",inline"`
	Dependencies    Dependencies `yaml:"dependencies" json:"dependencies"`
	Config          *struct {
		Fixture          *Fixture `yaml:"fixture" json:"fixture"`
		CustomImportPath string   `yaml:"customImportPath" json:"customImportPath"`
	} `yaml:"config" json:"config" validate:"nonzero"`
}

func newMockableClient(raw []byte) (*mockableClient, error) {
	config := &mockableClient{}
	if errUnmarshal := yaml.Unmarshal(raw, config); errUnmarshal != nil {
		return nil, errors.Wrap(
			errUnmarshal, "Could not parse testable client config data")
	}

	if errValidate := validator.Validate(config); errValidate != nil {
		return nil, errors.Wrap(
			errValidate, "testable client config validation failed")
	}

	return config, nil
}

// simpleHookFunc is a wrapper around a generate function
type simpleHookFunc struct {
	h *PackageHelper

	name            string
	directorySuffix string
	ShouldRunHook   func(instance *ModuleInstance) bool
	Generator       func(instance *ModuleInstance) (map[string][]byte, error)
}

var _ PostGenHook = (*simpleHookFunc)(nil)

func (h *simpleHookFunc) Name() string {
	return h.name
}

func (h *simpleHookFunc) Build(instance *ModuleInstance) error {
	if !h.ShouldRunHook(instance) {
		return nil
	}

	files, err := h.Generator(instance)
	if err != nil {
		return err
	}

	outputDir := filepath.Join(h.h.CodeGenTargetPath(), instance.Directory, h.directorySuffix)
	prettyDir, _ := filepath.Rel(h.h.configRoot, outputDir)

	PrintGenLine(
		"mock",
		instance.ClassName,
		instance.InstanceName,
		prettyDir,
		1,
		1,
	)

	for path, bytes := range files {
		err := WriteAndFormat(filepath.Join(outputDir, path), bytes)
		if err != nil {
			return err
		}
	}

	return nil
}

// ClientMockGenHook returns a PostGenHook to generate client mocks
func ClientMockGenHook(h *PackageHelper, t *Template) (PostGenHook, error) {
	bin, err := NewMockgenBin(h, t)
	if err != nil {
		return nil, errors.Wrap(err, "error building mockgen binary")
	}

	return &simpleHookFunc{
		h:               h,
		name:            "mock-client",
		directorySuffix: "mock-client",
		ShouldRunHook: func(instance *ModuleInstance) bool {
			return instance.ClassName == "client"
		},
		Generator: func(instance *ModuleInstance) (map[string][]byte, error) {
			fixtureMap := make(map[string]*Fixture)
			pathSymbolMap := make(map[string]string)
			key := instance.ClassType + instance.InstanceName
			client, errClient := newMockableClient(instance.YAMLFileRaw)
			if errClient != nil {
				return nil, errors.Wrapf(
					err,
					"error parsing client-config for client %q",
					instance.InstanceName,
				)
			}

			importPath := instance.PackageInfo.GeneratedPackagePath
			if instance.ClassType == custom {
				importPath = client.Config.CustomImportPath
			}

			// gather all modules that need to generate fixture types
			f := client.Config.Fixture
			if f != nil && f.Scenarios != nil {
				pathSymbolMap[importPath] = clientInterface
				fixtureMap[key] = f
			}

			// only run reflect program once to gather interface info for all clients
			pkgs, err := ReflectInterface("", pathSymbolMap)
			if err != nil {
				return nil, errors.Wrap(err, "error parsing Client interfaces")
			}

			files := make(map[string][]byte)

			// buildDir := h.CodeGenTargetPath()
			// genDir := filepath.Join(buildDir, instance.Directory, "mock-client")

			// generate mock client, this starts a sub process
			mock, err := bin.GenMock(importPath, "clientmock", clientInterface)
			if err != nil {
				return nil, errors.Wrapf(
					err,
					"error generating mocks for client %q",
					instance.InstanceName,
				)
			}
			files["mock_client.go"] = mock

			// generate fixture types and augmented mock client
			if f, ok := fixtureMap[key]; ok {
				types, augMock, err := bin.AugmentMockWithFixture(pkgs[importPath], f, clientInterface)
				if err != nil {
					return nil, errors.Wrapf(
						err,
						"error generating fixture types for client %q",
						instance.InstanceName,
					)
				}

				files["types.go"] = types
				files["mock_client_with_fixture.go"] = augMock
			}

			return files, nil
		},
	}, nil
}

// ServiceMockGenHook returns a PostGenHook to generate service mocks
func ServiceMockGenHook(h *PackageHelper, t *Template) PostGenHook {
	return &simpleHookFunc{
		h:               h,
		directorySuffix: "mock-service",

		ShouldRunHook: func(instance *ModuleInstance) bool {
			return instance.ClassName == "service"
		},
		Generator: func(instance *ModuleInstance) (map[string][]byte, error) {
			files := make(map[string][]byte)

			mockInit, err := generateMockInitializer(instance, h, t)
			if err != nil {
				return nil, errors.Wrapf(
					err,
					"Error generating service mock_init.go for %s",
					instance.InstanceName,
				)
			}
			files["mock_init.go"] = mockInit

			mockService, err := generateServiceMock(instance, h, t)
			if err != nil {
				return nil, errors.Wrapf(
					err,
					"Error generating service mock_service.go for %s",
					instance.InstanceName,
				)
			}
			files["mock_service.go"] = mockService

			return files, nil
		},
	}
}

// generateMockInitializer generates code to initialize modules with leaf nodes being mocks
func generateMockInitializer(instance *ModuleInstance, h *PackageHelper, t *Template) ([]byte, error) {
	leafWithFixture, err := FindClientsWithFixture(instance)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"Instance":        instance,
		"LeafWithFixture": leafWithFixture,
	}
	return t.ExecTemplate("module_mock_initializer.tmpl", data, h)
}

// generateServiceMock generates mock service
func generateServiceMock(instance *ModuleInstance, h *PackageHelper, t *Template) ([]byte, error) {
	configPath := path.Join(strings.Replace(instance.Directory, "services", "config", 1), "test.yaml")
	if _, err := os.Stat(filepath.Join(h.ConfigRoot(), configPath)); err != nil {
		if os.IsNotExist(err) {
			configPath = "config/test.yaml"
		}
	}
	data := map[string]interface{}{
		"Instance":       instance,
		"TestConfigPath": filepath.Join(h.packageRoot, configPath),
	}
	return t.ExecTemplate("service_mock.tmpl", data, h)
}

// FindClientsWithFixture finds the given module's dependent clients that have fixture config
func FindClientsWithFixture(instance *ModuleInstance) (map[string]string, error) {
	clientsWithFixture := map[string]string{}
	for _, leaf := range instance.RecursiveDependencies["client"] {
		client, err := newMockableClient(leaf.YAMLFileRaw)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"error parsing client-config for client %q",
				instance.InstanceName,
			)
		}

		if client.Config.Fixture != nil {
			clientsWithFixture[leaf.InstanceName] = client.Config.Fixture.ImportPath
		}
	}
	return clientsWithFixture, nil
}

// WriteAndFormat writes the data (Go code) to given file path, creates path if it does not exist and formats the file using gofmt
func WriteAndFormat(path string, data []byte) error {
	if err := writeFile(path, data); err != nil {
		return errors.Wrapf(
			err,
			"Error writing to file %q",
			path,
		)
	}
	return FormatGoFile(path)
}

// WorkflowMockGenHook returns a PostGenHook to generate endpoint workflow mocks
func WorkflowMockGenHook(h *PackageHelper, t *Template) PostGenHook {
	return &simpleHookFunc{
		h:               h,
		directorySuffix: "mock-workflow",

		ShouldRunHook: func(instance *ModuleInstance) bool {
			if instance.ClassName != "endpoint" {
				return false
			}

			specs := instance.genSpec.([]*EndpointSpec)
			for _, spec := range specs {
				if spec.WorkflowType == "custom" {
					return true
				}
			}

			return false
		},

		Generator: func(instance *ModuleInstance) (map[string][]byte, error) {
			files := make(map[string][]byte)

			cwf, err := FindClientsWithFixture(instance)
			if err != nil {
				return nil, errors.Wrapf(
					err,
					"Error generating mock endpoint %s",
					instance.InstanceName,
				)
			}

			mockClientsType, err := generateEndpointMockClientsType(instance, cwf, h, t)
			if err != nil {
				return nil, errors.Wrapf(
					err,
					"Error generating mock clients type.go for endpoint %s",
					instance.InstanceName,
				)
			}

			files["type.go"] = mockClientsType

			specs := instance.genSpec.([]*EndpointSpec)
			var neededSpecs []*EndpointSpec

			for _, spec := range specs {
				if spec.WorkflowType == "custom" {
					neededSpecs = append(neededSpecs, spec)
				}
			}

			for _, endpointSpec := range neededSpecs {
				mockWorkflow, err := generateMockWorkflow(endpointSpec, instance, cwf, h, t)
				if err != nil {
					return nil, errors.Wrapf(
						err,
						"Error generating mock workflow for %s",
						instance.InstanceName,
					)
				}
				filename := strings.ToLower(endpointSpec.ThriftServiceName) + "_" +
					strings.ToLower(endpointSpec.ThriftMethodName) + "_workflow_mock.go"
				files[filename] = mockWorkflow
			}

			return files, nil
		},
	}
}

// generateMockWorkflow generates an initializer that creates an endpoint workflow with mock clients
func generateMockWorkflow(
	espec *EndpointSpec,
	instance *ModuleInstance,
	clientsWithFixture map[string]string,
	h *PackageHelper,
	t *Template,
) ([]byte, error) {
	data := map[string]interface{}{
		"Instance":           instance,
		"EndpointSpec":       espec,
		"ClientsWithFixture": clientsWithFixture,
	}
	return t.ExecTemplate("workflow_mock.tmpl", data, h)
}

// generateEndpointMockClientsType generates the type that contains mock clients for the endpoint
func generateEndpointMockClientsType(
	instance *ModuleInstance,
	clientsWithFixture map[string]string,
	h *PackageHelper,
	t *Template,
) ([]byte, error) {
	data := map[string]interface{}{
		"Instance":           instance,
		"ClientsWithFixture": clientsWithFixture,
	}
	return t.ExecTemplate("workflow_mock_clients_type.tmpl", data, h)
}
