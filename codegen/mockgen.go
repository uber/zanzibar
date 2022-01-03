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
	"bytes"
	"go/token"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/mock/mockgen/model"
	"github.com/pkg/errors"
)

const (
	mockgenPkg = "github.com/golang/mock/mockgen"
)

// MockgenBin is a struct abstracts the mockgen binary built from mockgen package in vendor
type MockgenBin struct {
	pkgHelper *PackageHelper
	tmpl      *Template
}

// NewMockgenBin builds the mockgen binary from vendor directory
func NewMockgenBin(h *PackageHelper, t *Template) (*MockgenBin, error) {
	return &MockgenBin{
		pkgHelper: h,
		tmpl:      t,
	}, nil
}

// GenMock generates mocks for given module instance, pkg is the package name of the generated mocks,
// and intf is the interface name to generate mock for
func (m MockgenBin) GenMock(importPath, pkg, intf string) ([]byte, error) {
	cmd := exec.Command("mockgen", "-package", pkg, importPath, intf)
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

// AugmentMockWithFixture generates mocks with fixture for the interface in the given package
func (m MockgenBin) AugmentMockWithFixture(pkg *model.Package, f *Fixture, intf string) ([]byte, []byte, error) {
	methodsMap := make(map[string]*model.Method, len(pkg.Interfaces[0].Methods))
	validationMap := make(map[string]interface{}, len(pkg.Interfaces[0].Methods))
	for _, m := range pkg.Interfaces[0].Methods {
		methodsMap[m.Name] = m
		validationMap[m.Name] = struct{}{}
	}

	if err := f.Validate(validationMap); err != nil {
		return nil, nil, errors.Wrap(err, "invalid fixture config")
	}

	exposedMethods := make([]*model.Method, 0, len(f.Scenarios))
	for name := range f.Scenarios {
		exposedMethods = append(exposedMethods, methodsMap[name])
	}

	// sort methods in given fixture config for predictable fixture type generation
	sort.Sort(byMethodName(exposedMethods))

	pkgPathToAlias := uniqueAlias(pkg.Imports())
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

		method := &reflectMethod{
			Name:      m.Name,
			In:        in,
			Out:       out,
			InString:  strings.Join(inString, " ,"),
			OutString: strings.Join(outString, " ,"),
		}

		if m.Variadic != nil {
			method.Variadic = "arg" + strconv.Itoa(len(m.In))
			method.VariadicType = m.Variadic.Type.String(pkgPathToAlias, "")
		}

		methods = append(methods, method)
	}

	data := map[string]interface{}{
		"Imports":         pkgPathToAlias,
		"Methods":         methods,
		"Fixture":         f,
		"ClientInterface": intf,
	}
	types, err := m.tmpl.ExecTemplate("fixture_types.tmpl", data, m.pkgHelper)
	if err != nil {
		return nil, nil, err
	}
	mock, err := m.tmpl.ExecTemplate("augmented_mock.tmpl", data, m.pkgHelper)
	if err != nil {
		return nil, nil, err
	}
	return types, mock, nil
}

type reflectMethod struct {
	Name                string
	In, Out             map[string]string
	Variadic            string
	VariadicType        string
	InString, OutString string
}

// uniqueAlias returns a map of import path to alias where the aliases are unique
func uniqueAlias(importPaths map[string]bool) map[string]string {
	pkgPathToAlias := make(map[string]string, len(importPaths))
	usedAliases := make(map[string]bool, len(importPaths))
	for pkgPath := range importPaths {
		base := CamelCase(path.Base(pkgPath))
		pkgAlias := base
		i := 0
		for usedAliases[pkgAlias] || token.Lookup(pkgAlias).IsKeyword() {
			pkgAlias = base + strconv.Itoa(i)
			i++
		}

		pkgPathToAlias[pkgPath] = pkgAlias
		usedAliases[pkgAlias] = true
	}
	return pkgPathToAlias
}
