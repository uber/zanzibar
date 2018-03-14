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

	// projRoot is the absolute path of the project, it is also where the vendor directory is
	projRoot  string
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
	projRoot := filepath.Join(os.Getenv("GOPATH"), "src", pkgRoot)

	// we assume that the vendor directory is flattened as Glide does
	mockgenDir := path.Join(projRoot, "vendor", mockgenPkg)
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

		projRoot:  projRoot,
		pkgHelper: h,
		tmpl:      t,
	}, nil
}

// GenMock generates mocks for given module instance, pkg is the package name of the generated mocks,
// and intf is the interface name to generate mock for
func (m MockgenBin) GenMock(importPath, pkg, intf string) ([]byte, error) {
	cmd := exec.Command(m.Bin, "-package", pkg, importPath, intf)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(
			err,
			"error running command %q: %s",
			strings.Join(cmd.Args, " "),
			stderr.String(),
		)
	}

	return stdout.Bytes(), nil
}

// byMethodName implements sort.Interface for []*modelMethod based on the Name field
type byMethodName []*model.Method

func (b byMethodName) Len() int           { return len(b) }
func (b byMethodName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byMethodName) Less(i, j int) bool { return b[i].Name < b[j].Name }

// AugmentMockWithFixture generates mocks with fixture for the interface in the given importPath
func (m MockgenBin) AugmentMockWithFixture(importPath string, f *Fixture, intf string) (types, mock []byte, err error) {
	pkg, err := ReflectInterface(m.projRoot, importPath, []string{intf})
	if err != nil {
		return
	}

	methodsMap := make(map[string]*model.Method, len(pkg.Interfaces[0].Methods))
	validationMap := make(map[string]interface{}, len(pkg.Interfaces[0].Methods))
	for _, m := range pkg.Interfaces[0].Methods {
		methodsMap[m.Name] = m
		validationMap[m.Name] = struct{}{}
	}

	if err = f.Validate(validationMap); err != nil {
		err = errors.Wrap(err, "invalid fixture config")
		return
	}

	exposedMethods := make([]*model.Method, 0, len(f.Scenarios))
	for name := range f.Scenarios {
		exposedMethods = append(exposedMethods, methodsMap[name])
	}

	// sort methods in given fixture config for predictable fixture type generation
	sort.Sort(byMethodName(exposedMethods))

	imports := pkg.Imports()
	pkgPathToAlias := make(map[string]string, len(imports))
	usedAliases := make(map[string]bool, len(imports))
	for pkgPath := range imports {
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

	methods := make([]*reflectMethod, 0, len(exposedMethods))
	for _, m := range exposedMethods {
		numIn := len(m.In)
		in := make(map[string]string, numIn)
		inString := make([]string, 0, numIn)
		for i, param := range m.In {
			arg := "arg" + strconv.Itoa(i)
			in[arg] = param.Type.String(pkgPathToAlias, "")
			inString = append(inString, arg)
		}

		numOut := len(m.Out)
		out := make(map[string]string, numOut)
		outString := make([]string, 0, numOut)
		for i, param := range m.Out {
			ret := "ret" + strconv.Itoa(i)
			out[ret] = param.Type.String(pkgPathToAlias, "")
			outString = append(outString, ret)
		}

		methods = append(methods, &reflectMethod{
			Name:      m.Name,
			In:        in,
			Out:       out,
			InString:  strings.Join(inString, " ,"),
			OutString: strings.Join(outString, " ,"),
		})
	}

	data := map[string]interface{}{
		"Imports": pkgPathToAlias,
		"Methods": methods,
		"Fixture": f,
	}
	types, err = m.tmpl.ExecTemplate("fixture_types.tmpl", data, m.pkgHelper)
	if err != nil {
		return
	}
	mock, err = m.tmpl.ExecTemplate("augmented_mock.tmpl", data, m.pkgHelper)
	if err != nil {
		types = nil
		return
	}
	return
}

type reflectMethod struct {
	Name                string
	In, Out             map[string]string
	InString, OutString string
}
