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
	"strings"
	"sync"
	"sync/atomic"

	yaml "github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/uber/zanzibar/parallelize"
	"gopkg.in/validator.v2"
)

const (
	defaultClientInterface = "Client"
	custom                 = "custom"
)

type mockableClient struct {
	ClassConfigBase `yaml:",inline" json:",inline"`
	Dependencies    Dependencies `yaml:"dependencies" json:"dependencies"`
	Config          *struct {
		Fixture          *Fixture `yaml:"fixture" json:"fixture"`
		CustomImportPath string   `yaml:"customImportPath" json:"customImportPath"`
		CustomInterface  string   `yaml:"customInterface,omitempty" json:"customInterface,omitempty"`
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

// ClientMockGenHook returns a PostGenHook to generate client mocks
func ClientMockGenHook(h *PackageHelper, t *Template, parallelizeFactor int) (PostGenHook, error) {
	return func(instances map[string][]*ModuleInstance) error {
		clientInstances := instances["client"]
		mockCount := len(clientInstances)
		if mockCount == 0 {
			return nil
		}

		fmt.Printf("Generating %d client mocks:\n", mockCount)
		bin, err := NewMockgenBin(h, t)
		if err != nil {
			return errors.Wrap(err, "error building mockgen binary")
		}

		importPathMap := make(map[string]string, mockCount)
		fixtureMap := make(map[string]*Fixture, mockCount)
		clientInterfaceMap := make(map[string]string, mockCount)
		pathSymbolMap := make(map[string]string, mockCount)
		runner := parallelize.NewUnboundedRunner(mockCount)
		var mutex sync.Mutex
		for _, instance := range clientInstances {
			f := func(instanceInf interface{}) (interface{}, error) {
				instance := instanceInf.(*ModuleInstance)
				key := instance.ClassType + instance.InstanceName
				client, errClient := newMockableClient(instance.YAMLFileRaw)
				if errClient != nil {
					return nil, errors.Wrapf(
						err,
						"error parsing client-config for client %q",
						instance.InstanceName,
					)
				}

				importPath := instance.PackageInfo.ImportPackagePath()
				customInterface := client.Config.CustomInterface
				if instance.ClassType == custom {
					importPath = client.Config.CustomImportPath
				}

				clientInterface := defaultClientInterface
				// if an interfaces name is provided use that, else use "Client"
				if customInterface != "" {
					clientInterface = customInterface
				}
				mutex.Lock()
				defer mutex.Unlock()
				importPathMap[key] = importPath
				clientInterfaceMap[key] = clientInterface
				// gather all modules that need to generate fixture types
				f := client.Config.Fixture
				if f != nil && f.Scenarios != nil {
					pathSymbolMap[importPath] = clientInterface
					fixtureMap[key] = f
				}
				return nil, nil
			}
			wrk := &parallelize.SingleParamWork{Data: instance, Func: f}
			runner.SubmitWork(wrk)
		}

		if _, err := runner.GetResult(); err != nil {
			return err
		}

		// only run reflect program once to gather interface info for all clients
		pkgs, err := ReflectInterface("", pathSymbolMap)
		if err != nil {
			return errors.Wrap(err, "error parsing Client interfaces")
		}

		var idx int32 = 1
		var files sync.Map
		runner = parallelize.NewBoundedRunner(mockCount, parallelizeFactor*runtime.NumCPU())
		for _, instance := range clientInstances {
			f := func(instanceInf interface{}) (interface{}, error) {
				instance := instanceInf.(*ModuleInstance)
				key := instance.ClassType + instance.InstanceName
				buildDir := h.CodeGenTargetPath()
				genDir := filepath.Join(buildDir, instance.Directory, "mock-client")

				importPath := importPathMap[key]

				// generate mock client, this starts a sub process.
				mock, err := bin.GenMock(importPath, "clientmock", clientInterfaceMap[key])
				if err != nil {
					return nil, errors.Wrapf(
						err,
						"error generating mocks for client %q",
						instance.InstanceName,
					)
				}
				files.Store(filepath.Join(genDir, "mock_client.go"), mock)

				// generate fixture types and augmented mock client
				if f, ok := fixtureMap[key]; ok {
					types, augMock, err := bin.AugmentMockWithFixture(pkgs[importPath], f, clientInterfaceMap[key])
					if err != nil {
						return nil, errors.Wrapf(
							err,
							"error generating fixture types for client %q",
							instance.InstanceName,
						)
					}

					files.Store(filepath.Join(genDir, "types.go"), types)
					files.Store(filepath.Join(genDir, "mock_client_with_fixture.go"), augMock)
				}

				PrintGenLine(
					"mock",
					instance.ClassName,
					instance.InstanceName,
					path.Join(path.Base(buildDir), instance.Directory, "mock-client"),
					int(atomic.LoadInt32(&idx)), mockCount,
				)
				atomic.AddInt32(&idx, 1)
				return nil, nil
			}
			wrk := &parallelize.SingleParamWork{Data: instance, Func: f}
			runner.SubmitWork(wrk)
		}

		if _, err := runner.GetResult(); err != nil {
			return errors.Wrap(err, "encountered errors when generating mock clients")
		}

		files.Range(func(p, data interface{}) bool {
			if err = WriteAndFormat(p.(string), data.([]byte)); err != nil {
				return false
			}
			return true
		})

		return err
	}, nil
}

// ServiceMockGenHook returns a PostGenHook to generate service mocks
func ServiceMockGenHook(h *PackageHelper, t *Template, configFile string) PostGenHook {
	return func(instances map[string][]*ModuleInstance) error {
		mockCount := len(instances["service"])
		if mockCount == 0 {
			return nil
		}

		fmt.Printf("Generating %d service mocks:\n", mockCount)
		ec := make(chan error, mockCount)
		var idx int32 = 1
		var files sync.Map
		var wg sync.WaitGroup
		wg.Add(mockCount)
		for _, instance := range instances["service"] {
			go func(instance *ModuleInstance) {
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
				files.Store(filepath.Join(genDir, "mock_init.go"), mockInit)

				mockService, err := generateServiceMock(instance, h, t, configFile)
				if err != nil {
					ec <- errors.Wrapf(
						err,
						"Error generating service mock_service.go for %s",
						instance.InstanceName,
					)
					return
				}
				files.Store(filepath.Join(genDir, "mock_service.go"), mockService)

				PrintGenLine(
					"mock",
					instance.ClassName,
					instance.InstanceName,
					path.Join(path.Base(buildDir), instance.Directory, "mock-service"),
					int(atomic.LoadInt32(&idx)), mockCount,
				)
				atomic.AddInt32(&idx, 1)
			}(instance)
		}
		wg.Wait()

		select {
		case err := <-ec:
			close(ec)
			errs := []string{err.Error()}
			for e := range ec {
				errs = append(errs, e.Error())
			}
			return errors.Errorf(
				"encountered %d errors when generating mock services:\n%s",
				len(errs),
				strings.Join(errs, "\n"),
			)
		default:
		}

		var err error
		files.Range(func(p, data interface{}) bool {
			if err = WriteAndFormat(p.(string), data.([]byte)); err != nil {
				return false
			}
			return true
		})

		return err
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
func generateServiceMock(instance *ModuleInstance, h *PackageHelper, t *Template, configFile string) ([]byte, error) {
	configPath := path.Join(strings.Replace(instance.Directory, "services", "config", 1), configFile)
	if _, err := os.Stat(filepath.Join(h.ConfigRoot(), configPath)); err != nil {
		if os.IsNotExist(err) {
			configPath = filepath.Join("config", configFile)
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
	return func(instances map[string][]*ModuleInstance) error {
		errChanSize := 0
		shouldGenMap := map[*ModuleInstance][]*EndpointSpec{}
		for _, instance := range instances["endpoint"] {
			if instance.GeneratedSpec() == nil {
				continue
			}
			endpointSpecs := instance.GeneratedSpec().([]*EndpointSpec)
			for _, endpointSpec := range endpointSpecs {
				if endpointSpec.WorkflowType == "custom" {
					shouldGenMap[instance] = endpointSpecs
					errChanSize += len(endpointSpecs)
					break
				}
			}
		}

		genEndpointCount := len(shouldGenMap)
		if genEndpointCount == 0 {
			return nil
		}
		errChanSize += genEndpointCount

		fmt.Printf("Generating %d endpoint mocks:\n", genEndpointCount)
		ec := make(chan error, errChanSize)
		var idx int32 = 1
		var files sync.Map
		var wg sync.WaitGroup
		for instance, endpointSpecs := range shouldGenMap {
			wg.Add(1)
			go func(instance *ModuleInstance, endpointSpecs []*EndpointSpec) {
				defer wg.Done()

				buildDir := h.CodeGenTargetPath()
				genDir := filepath.Join(buildDir, instance.Directory, "mock-workflow")

				cwf, err := FindClientsWithFixture(instance)
				if err != nil {
					ec <- errors.Wrapf(
						err,
						"Error generating mock endpoint %s",
						instance.InstanceName,
					)
					return
				}

				var subWg sync.WaitGroup

				subWg.Add(1)
				go func() {
					defer subWg.Done()

					mockClientsType, err := generateEndpointMockClientsType(instance, cwf, h, t)
					if err != nil {
						ec <- errors.Wrapf(
							err,
							"Error generating mock clients type.go for endpoint %s",
							instance.InstanceName,
						)
						return
					}
					files.Store(filepath.Join(genDir, "type.go"), mockClientsType)
				}()

				for _, endpointSpec := range endpointSpecs {
					if endpointSpec.WorkflowType != "custom" {
						continue
					}

					subWg.Add(1)

					go func(espec *EndpointSpec) {
						defer subWg.Done()

						mockWorkflow, err := generateMockWorkflow(espec, instance, cwf, h, t)
						if err != nil {
							ec <- errors.Wrapf(
								err,
								"Error generating mock workflow for %s",
								instance.InstanceName,
							)
							return
						}
						filename := strings.ToLower(espec.ThriftServiceName) + "_" +
							strings.ToLower(espec.ThriftMethodName) + "_workflow_mock.go"
						files.Store(filepath.Join(genDir, filename), mockWorkflow)
					}(endpointSpec)
				}

				subWg.Wait()
				PrintGenLine(
					"mock",
					instance.ClassName,
					instance.InstanceName,
					path.Join(path.Base(buildDir), instance.Directory, "mock-workflow"),
					int(atomic.LoadInt32(&idx)), genEndpointCount,
				)
				atomic.AddInt32(&idx, 1)
			}(instance, endpointSpecs)
		}
		wg.Wait()

		select {
		case err := <-ec:
			close(ec)
			errs := []string{err.Error()}
			for e := range ec {
				errs = append(errs, e.Error())
			}
			return errors.Errorf(
				"encountered %d errors when generating mock endpoint workflows:\n%s",
				len(errs),
				strings.Join(errs, "\n"),
			)
		default:
		}

		var err error

		files.Range(func(p, data interface{}) bool {
			if err = WriteAndFormat(p.(string), data.([]byte)); err != nil {
				return false
			}
			return true
		})

		return err
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
