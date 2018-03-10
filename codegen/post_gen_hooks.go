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
	"encoding/json"
	"fmt"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/mock/mockgen/model"
	"github.com/pkg/errors"
)

const clientInterface = "Client"

// ClientMockGenHook returns a PostGenHook to generate client mocks
func ClientMockGenHook(h *PackageHelper, t *Template) (PostGenHook, error) {
	bin, err := NewMockgenBin(h.PackageRoot())
	if err != nil {
		return nil, errors.Wrap(err, "error building mockgen binary")
	}

	return func(instances map[string][]*ModuleInstance) error {
		fmt.Println("Generating client mocks:")
		mockCount := len(instances["client"])
		for i, instance := range instances["client"] {
			if err := bin.GenMock(
				instance,
				"mock-client/mock_client.go",
				"clientmock",
				"Client",
			); err != nil {
				return errors.Wrapf(
					err,
					"error generating mocks for client %q",
					instance.InstanceName,
				)
			}

			if err := genMockClientWithFixtureCustom(instance, h, t); err != nil {
				return errors.Wrapf(
					err,
					"error generating mock client with fixtures for client %q",
					instance.InstanceName,
				)
			}

			fmt.Printf(
				genFormattor,
				"mock",
				instance.ClassName,
				instance.InstanceName,
				path.Join(path.Base(h.CodeGenTargetPath()), instance.Directory, "mock-client"),
				i+1, mockCount,
			)
		}
		return nil
	}, nil
}

type clientConfig struct {
	Config *config `json:"config"`
}

type config struct {
	CustomImportPath string   `json:"customImportPath"`
	Fixture          *Fixture `json:"fixture"`
}

func genMockClientWithFixtureCustom(instance *ModuleInstance, h *PackageHelper, t *Template) error {
	var cc clientConfig
	if err := json.Unmarshal(instance.JSONFileRaw, &cc); err != nil {
		return errors.Wrapf(
			err,
			"error parsing client-config.json for client %q",
			instance.InstanceName,
		)
	}

	if cc.Config == nil || cc.Config.Fixture == nil || cc.Config.Fixture.Scenarios == nil {
		return nil
	}

	//importPath := instance.PackageInfo.GeneratedPackagePath
	var importPath string
	if instance.ClassType == "custom" {
		importPath = cc.Config.CustomImportPath
		if importPath == "" {
			return errors.Errorf("custom client %q must have customImportPath", instance.ClassName)
		}
	} else {
		return nil
	}

	appRoot := filepath.Join(os.Getenv("GOPATH"), "src", h.PackageRoot())
	pkg, err := ReflectInterface(appRoot, importPath, []string{clientInterface})
	if err != nil {
		return err
	}

	fixture := cc.Config.Fixture

	methodsMap := make(map[string]*model.Method, len(pkg.Interfaces[0].Methods))
	validationMap := make(map[string]interface{}, len(pkg.Interfaces[0].Methods))
	for _, m := range pkg.Interfaces[0].Methods {
		methodsMap[m.Name] = m
		validationMap[m.Name] = struct{}{}
	}
	if err := fixture.Validate(validationMap); err != nil {
		return errors.Wrapf(
			err,
			"invalid fixture config for client %q",
			instance.InstanceName,
		)
	}

	// sort methods in given fixture config for predictable fixture type generation
	numMethods := len(fixture.Scenarios)
	sortedMethods := make([]string, numMethods, numMethods)
	i := 0
	for name := range fixture.Scenarios {
		sortedMethods[i] = name
		i++
	}
	sort.Strings(sortedMethods)

	exposedMethods := make([]*model.Method, numMethods, numMethods)
	for i, methodName := range sortedMethods {
		exposedMethods[i] = methodsMap[methodName]
	}

	imports := pkg.Imports()

	// Sort keys to make import alias generation predictable
	sortedPaths := make([]string, len(imports), len(imports))
	j := 0
	for pth := range imports {
		sortedPaths[i] = pth
		j++
	}
	sort.Strings(sortedPaths)

	pkgPathToAlias := make(map[string]string, len(imports))
	usedAliases := make(map[string]bool, len(imports))
	for _, pkgPath := range sortedPaths {
		base := camelCase(path.Base(pkgPath))
		pkgAlias := base
		i := 0
		for usedAliases[pkgAlias] || token.Lookup(pkgAlias).IsKeyword() {
			pkgAlias = base + strconv.Itoa(i)
			i++
		}

		pkgPathToAlias[pkgPath] = pkgAlias
		usedAliases[pkgAlias] = true
	}

	methods := make([]*reflectMethod, len(exposedMethods))
	for i, m := range exposedMethods {
		numIn := len(m.In)
		in := make(map[string]string, numIn)
		inString := make([]string, numIn, numIn)
		for i, param := range m.In {
			arg := "arg" + strconv.Itoa(i)
			in[arg] = param.Type.String(pkgPathToAlias, "")
			inString[i] = arg
		}

		numOut := len(m.Out)
		out := make(map[string]string, numOut)
		outString := make([]string, numOut, numOut)
		for i, param := range m.Out {
			ret := "ret" + strconv.Itoa(i)
			out[ret] = param.Type.String(pkgPathToAlias, "")
			outString[i] = ret
		}

		methods[i] = &reflectMethod{
			Name:      m.Name,
			In:        in,
			Out:       out,
			InString:  strings.Join(inString, " ,"),
			OutString: strings.Join(outString, " ,"),
		}
	}

	data := map[string]interface{}{
		"Imports": pkgPathToAlias,
		"Methods": methods,
		"Fixture": fixture,
	}
	types, err := t.ExecTemplate("client_fixture_types_custom.tmpl", data, h)
	if err != nil {
		return err
	}
	mock, err := t.ExecTemplate("client_mock_custom.tmpl", data, h)
	if err != nil {
		return err
	}

	buildPath := filepath.Join(h.CodeGenTargetPath(), instance.Directory)
	typesPath := filepath.Join(buildPath, "mock-client/types.go")
	mockPath := filepath.Join(buildPath, "mock-client/mock_client_with_fixture.go")
	if err := writeAndFormat(typesPath, types); err != nil {
		return err
	}
	if err := writeAndFormat(mockPath, mock); err != nil {
		return err
	}

	return nil
}

type reflectMethod struct {
	Name                string
	In, Out             map[string]string
	InString, OutString string
}

func writeAndFormat(path string, data []byte) error {
	if err := writeFile(path, data); err != nil {
		return errors.Wrapf(
			err,
			"Error writing to file %q",
			path,
		)
	}
	if err := FormatGoFile(path); err != nil {
		return err
	}
	return nil
}

// ServiceMockGenHook returns a PostGenHook to generate server mocks
func ServiceMockGenHook(h *PackageHelper, t *Template) PostGenHook {
	return func(instances map[string][]*ModuleInstance) error {
		fmt.Println("Generating service mocks:")
		mockCount := len(instances["service"])
		for i, instance := range instances["service"] {
			mockInit, err := generateMockInitializer(instance, h, t)
			if err != nil {
				return errors.Wrapf(
					err,
					"Error generating service mock_init.go for %s",
					instance.InstanceName,
				)
			}
			mockService, err := t.ExecTemplate("service_mock.tmpl", instance, h)
			if err != nil {
				return errors.Wrapf(
					err,
					"Error generating service mock_service.go for %s",
					instance.InstanceName,
				)
			}

			buildPath := filepath.Join(h.CodeGenTargetPath(), instance.Directory)
			mockInitPath := filepath.Join(buildPath, "mock-service/mock_init.go")
			mockServicePath := filepath.Join(buildPath, "mock-service/mock_service.go")

			if err := writeAndFormat(mockInitPath, mockInit); err != nil {
				return err
			}
			if err := writeAndFormat(mockServicePath, mockService); err != nil {
				return err
			}

			fmt.Printf(
				genFormattor,
				"mock",
				instance.ClassName,
				instance.InstanceName,
				path.Join(path.Base(h.CodeGenTargetPath()), instance.Directory, "mock-service"),
				i+1, mockCount,
			)
		}
		return nil
	}
}

// generateMockInitializer generates code to initialize modules with leaf nodes being mocks
func generateMockInitializer(instance *ModuleInstance, h *PackageHelper, t *Template) ([]byte, error) {
	leafWithFixture := map[string]string{}
	for _, leaf := range instance.RecursiveDependencies["client"] {
		spec, ok := leaf.genSpec.(*ClientSpec)
		if ok && spec.Fixture != nil {
			leafWithFixture[leaf.InstanceName] = spec.Fixture.ImportPath
		}
	}
	data := map[string]interface{}{
		"Instance":        instance,
		"LeafWithFixture": leafWithFixture,
	}
	return t.ExecTemplate("module_mock_initializer.tmpl", data, h)
}
