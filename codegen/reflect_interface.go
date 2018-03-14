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
	"encoding/gob"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/golang/mock/mockgen/model"
	"github.com/pkg/errors"
)

type reflectData struct {
	ImportPath string
	Symbols    []string
}

// ReflectInterface uses reflection to obtain interface information identified by symbols in given import path
// projRoot is the root dir where mockgen is installed as a vendor package
func ReflectInterface(projRoot, importPath string, symbols []string) (*model.Package, error) {
	// We use TempDir instead of TempFile so we can control the filename.
	tmpDir, err := ioutil.TempDir(projRoot, "gomock_reflect_")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	const progSource = "prog.go"
	var progBinary = "prog.bin"
	if runtime.GOOS == "windows" {
		// Windows won't execute a program unless it has a ".exe" suffix.
		progBinary += ".exe"
	}

	// Generate program
	var program bytes.Buffer
	data := reflectData{
		ImportPath: importPath,
		Symbols:    symbols,
	}
	if err := reflectProgram.Execute(&program, &data); err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(filepath.Join(tmpDir, progSource), program.Bytes(), 0600); err != nil {
		return nil, err
	}

	// Build it
	var buildStdout, buildStderr bytes.Buffer
	build := exec.Command("go", "build", "-o", progBinary, progSource)
	build.Dir = tmpDir
	build.Stdout = &buildStdout
	build.Stderr = &buildStderr
	if err := build.Run(); err != nil {
		return nil, errors.Wrap(err, buildStderr.String())
	}
	progPath := filepath.Join(tmpDir, progBinary)

	// Run it
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(progPath)
	cmd.Dir = projRoot
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, stderr.String())
	}

	var pkg model.Package
	if err := gob.NewDecoder(&stdout).Decode(&pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

// This program reflects on an interface value, and prints the
// gob encoding of a model.Package to standard output.
// JSON doesn't work because of the model.Type interface.
var reflectProgram = template.Must(template.New("program").Parse(`
package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"reflect"

	"github.com/golang/mock/mockgen/model"

	pkg_ {{printf "%q" .ImportPath}}
)

func main() {
	its := []struct{
		sym string
		typ reflect.Type
	}{
		{{range .Symbols}}
		{ {{printf "%q" .}}, reflect.TypeOf((*pkg_.{{.}})(nil)).Elem()},
		{{end}}
	}
	pkg := &model.Package{
		// NOTE: This behaves contrary to documented behaviour if the
		// package name is not the final component of the import path.
		// The reflect package doesn't expose the package name, though.
		Name: path.Base({{printf "%q" .ImportPath}}),
	}

	stderr := os.Stderr
	for _, it := range its {
		intf, err := model.InterfaceFromInterfaceType(it.typ)
		if err != nil {
			fmt.Fprintf(stderr, "Reflection: %v\n", err)
			os.Exit(1)
		}
		intf.Name = it.sym
		pkg.Interfaces = append(pkg.Interfaces, intf)
	}
	if err := gob.NewEncoder(os.Stdout).Encode(pkg); err != nil {
		fmt.Fprintf(stderr, "gob encode: %v\n", err)
		os.Exit(1)
	}
}
`))
