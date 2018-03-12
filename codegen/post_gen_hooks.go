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
	"path"
	"path/filepath"
	"sync"

	"encoding/json"
	"github.com/pkg/errors"
)

const clientInterface = "Client"

// helper struct to pull out the fixture config
type moduleConfig struct {
	Config *config `json:"config"`
}

// config is the struct corresponding to the config field in client-config.json
type config struct {
	CustomImportPath string   `json:"customImportPath"`
	Fixture          *Fixture `json:"fixture"`
}

// Fixture specifies client fixture import path and all scenarios
type Fixture struct {
	// ImportPath is the package where the user-defined Fixture global variable is contained.
	// The Fixture object defines, for a given client, the standardized list of fixture scenarios for that client
	ImportPath string `json:"importPath"`
	// Scenarios is a map from zanzibar's exposed method name to a list of user-defined fixture scenarios for a client
	Scenarios map[string][]string `json:"scenarios"`
}

// Validate the fixture configuration
func (f *Fixture) Validate(methods map[string]interface{}) error {
	if f.ImportPath == "" {
		return errors.New("fixture importPath is empty")
	}
	for m := range f.Scenarios {
		if _, ok := methods[m]; !ok {
			return errors.Errorf("%q is not a valid method", m)
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

	return func(instances map[string][]*ModuleInstance) error {
		fmt.Println("Generating client mocks:")
		mockCount := len(instances["client"])
		files := make(map[string][]byte)
		ec := make(chan error, mockCount)
		var wg sync.WaitGroup
		wg.Add(mockCount)
		for i, instance := range instances["client"] {
			go func(instance *ModuleInstance, i int) {
				defer wg.Done()

				var mc moduleConfig
				if err := json.Unmarshal(instance.JSONFileRaw, &mc); err != nil {
					ec <- errors.Wrapf(
						err,
						"error parsing client-config.json for client %q",
						instance.InstanceName,
					)
					return
				}

				buildDir := h.CodeGenTargetPath()
				genDir := filepath.Join(buildDir, instance.Directory, "mock-client")

				importPath := instance.PackageInfo.GeneratedPackagePath
				if instance.ClassType == "custom" {
					importPath = mc.Config.CustomImportPath
					if importPath == "" {
						ec <- errors.Errorf("custom client %q must have customImportPath", instance.ClassName)
						return
					}
				}

				// generate mock client
				mock, err := bin.GenMock(importPath, "clientmock", clientInterface)
				if err != nil {
					ec <- errors.Wrapf(
						err,
						"error generating mocks for client %q",
						instance.InstanceName,
					)
					return
				}
				files[filepath.Join(genDir, "mock_client.go")] = mock

				// generate fixture types and augmented mock client
				f := mc.Config.Fixture
				if f != nil && f.Scenarios != nil {
					types, augMock, err := bin.AugmentMockWithFixture(importPath, f, clientInterface)
					if err != nil {
						ec <- errors.Wrapf(
							err,
							"error generating mock client with fixtures for client %q",
							instance.InstanceName,
						)
						return
					}

					files[filepath.Join(genDir, "types.go")] = types
					files[filepath.Join(genDir, "mock_client_with_fixture.go")] = augMock
				}

				fmt.Printf(
					genFormattor,
					"mock",
					instance.ClassName,
					instance.InstanceName,
					path.Join(path.Base(buildDir), instance.Directory, "mock-client"),
					i+1, mockCount,
				)
			}(instance, i)
		}
		wg.Wait()

		select {
		case err := <-ec:
			return err
		default:
		}

		for p, data := range files {
			if err := writeAndFormat(p, data); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

// ServiceMockGenHook returns a PostGenHook to generate service mocks
func ServiceMockGenHook(h *PackageHelper, t *Template) PostGenHook {
	return func(instances map[string][]*ModuleInstance) error {
		fmt.Println("Generating service mocks:")
		mockCount := len(instances["service"])
		files := make(map[string][]byte)
		ec := make(chan error, mockCount)
		var wg sync.WaitGroup
		wg.Add(mockCount)
		for i, instance := range instances["service"] {
			go func(instance *ModuleInstance, i int) {
				defer wg.Done()

				buildDir := h.CodeGenTargetPath()
				genDir := filepath.Join(buildDir, instance.Directory, "mock-service")

				mockInit, err := generateMockInitializer(instance, h, t)
				if err != nil {
					ec <- errors.Wrapf(
						err,
						"Error generating service mock_init.go for %s",
						instance.InstanceName,
					)
					return
				}
				files[filepath.Join(genDir, "mock_init.go")] = mockInit

				mockService, err := t.ExecTemplate("service_mock.tmpl", instance, h)
				if err != nil {
					ec <- errors.Wrapf(
						err,
						"Error generating service mock_service.go for %s",
						instance.InstanceName,
					)
					return
				}
				files[filepath.Join(genDir, "mock_service.go")] = mockService

				fmt.Printf(
					genFormattor,
					"mock",
					instance.ClassName,
					instance.InstanceName,
					path.Join(path.Base(buildDir), instance.Directory, "mock-service"),
					i+1, mockCount,
				)
			}(instance, i)
		}
		wg.Wait()

		select {
		case err := <-ec:
			return err
		default:
		}

		for p, data := range files {
			if err := writeAndFormat(p, data); err != nil {
				return err
			}
		}
		return nil
	}
}

// generateMockInitializer generates code to initialize modules with leaf nodes being mocks
func generateMockInitializer(instance *ModuleInstance, h *PackageHelper, t *Template) ([]byte, error) {
	leafWithFixture := map[string]string{}
	for _, leaf := range instance.RecursiveDependencies["client"] {
		var mc moduleConfig
		if err := json.Unmarshal(leaf.JSONFileRaw, &mc); err != nil {
			return nil, errors.Wrapf(
				err,
				"error parsing client-config.json for client %q",
				instance.InstanceName,
			)
		}
		if mc.Config != nil && mc.Config.Fixture != nil {
			leafWithFixture[leaf.InstanceName] = mc.Config.Fixture.ImportPath
		}
	}
	data := map[string]interface{}{
		"Instance":        instance,
		"LeafWithFixture": leafWithFixture,
	}
	return t.ExecTemplate("module_mock_initializer.tmpl", data, h)
}

// writeAndFormat writes the data to given file path, creates path if it does not exist and formats the file
func writeAndFormat(path string, data []byte) error {
	if err := writeFile(path, data); err != nil {
		return errors.Wrapf(
			err,
			"Error writing to file %q",
			path,
		)
	}
	return FormatGoFile(path)
}
