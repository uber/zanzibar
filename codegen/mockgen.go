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
	"bytes"
	"encoding/json"
	"go/token"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/mock/mockgen/model"
	"github.com/pkg/errors"
)

const (
	exampleGatewayPkg = "github.com/uber/zanzibar/examples/example-gateway"
	mockgenPkg        = "github.com/golang/mock/mockgen"
	zanzibarPkg       = "github.com/uber/zanzibar"
)

// MockgenBin is a struct abstracts the mockgen binary built from mockgen package in vendor
type MockgenBin struct {
	// Bin is the absolute path to the mockgen binary built from vendor
	Bin string

	pkgHelper *PackageHelper
	tmpl      *Template
}

// NewMockgenBin builds the mockgen binary from vendor directory
func NewMockgenBin(h *PackageHelper, t *Template) (*MockgenBin, error) {
	pkgRoot := h.PackageRoot()
	// we would not need this if example gateway is a fully standalone application
	if pkgRoot == exampleGatewayPkg {
		pkgRoot = zanzibarPkg
	}
	// we assume that the vendor directory is flattened as Glide does
	mockgenDir := path.Join(os.Getenv("GOPATH"), "src", pkgRoot, "vendor", mockgenPkg)
	if _, err := os.Stat(mockgenDir); err != nil {
		return nil, errors.Wrapf(
			err, "error finding mockgen package in the vendor dir: %q does not exist", mockgenDir,
		)
	}

	var mockgenBin = "mockgen.bin"
	if runtime.GOOS == "windows" {
		// Windows won't execute a program unless it has a ".exe" suffix.
		mockgenBin += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", mockgenBin, ".")
	cmd.Dir = mockgenDir

	var stderr, stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(
			err,
			"error running command %q in %s: %s",
			strings.Join(cmd.Args, " "),
			mockgenDir,
			stderr.String(),
		)
	}

	return &MockgenBin{
		Bin: path.Join(mockgenDir, mockgenBin),

		pkgHelper: h,
		tmpl:      t,
	}, nil
}

// GenMock generates mocks for given module instance, dest is the file path relative to
// the instance's generated package dir, pkg is the package name of the generated mocks,
// and intf is the interface to generate mock for
func (m MockgenBin) GenMock(instance *ModuleInstance, dest, pkg, intf string) error {
	importPath := instance.PackageInfo.GeneratedPackagePath
	if instance.ClassType == "custom" {
		importPath = instance.PackageInfo.PackagePath
	}

	genDir := path.Join(os.Getenv("GOPATH"), "src")
	genDir = path.Join(genDir, instance.PackageInfo.GeneratedPackagePath, path.Dir(dest))
	genDest := path.Join(genDir, path.Base(dest))
	if _, err := os.Stat(genDir); os.IsNotExist(err) {
		if err := os.MkdirAll(genDir, os.ModePerm); err != nil {
			return err
		}
	}

	cmd := exec.Command(m.Bin, "-destination", genDest, "-package", pkg, importPath, intf)
	cmd.Dir = genDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(
			err,
			"error running command %q in %s: %s",
			strings.Join(cmd.Args, " "),
			genDir,
			stderr.String(),
		)
	}

	return nil
}

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
func (f *Fixture) Validate(exposedMethods map[string]interface{}) error {
	if f.ImportPath == "" {
		return errors.New("fixture importPath is empty")
	}
	for method := range f.Scenarios {
		if _, ok := exposedMethods[method]; !ok {
			return errors.Errorf("method %q is not an exposed method", method)
		}
	}
	return nil
}

// AugmentMockWithFixture generates mocks with fixture for given module instance's interface
func (m MockgenBin) AugmentMockWithFixture(instance *ModuleInstance, intf string) error {
	var mc moduleConfig
	if err := json.Unmarshal(instance.JSONFileRaw, &mc); err != nil {
		return errors.Wrapf(
			err,
			"error parsing client-config.json for client %q",
			instance.InstanceName,
		)
	}

	if mc.Config == nil || mc.Config.Fixture == nil || mc.Config.Fixture.Scenarios == nil {
		return nil
	}

	importPath := instance.PackageInfo.GeneratedPackagePath
	if instance.ClassType == "custom" {
		importPath = mc.Config.CustomImportPath
		if importPath == "" {
			return errors.Errorf("custom client %q must have customImportPath", instance.ClassName)
		}
	}

	appRoot := filepath.Join(os.Getenv("GOPATH"), "src", m.pkgHelper.PackageRoot())
	pkg, err := ReflectInterface(appRoot, importPath, []string{intf})
	if err != nil {
		return err
	}

	fixture := mc.Config.Fixture

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
		sortedPaths[j] = pth
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
	types, err := m.tmpl.ExecTemplate("client_fixture_types.tmpl", data, m.pkgHelper)
	if err != nil {
		return err
	}
	mock, err := m.tmpl.ExecTemplate("client_mock.tmpl", data, m.pkgHelper)
	if err != nil {
		return err
	}

	buildPath := filepath.Join(m.pkgHelper.CodeGenTargetPath(), instance.Directory)
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
