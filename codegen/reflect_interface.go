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
	PathAliasMap  map[string]string
	PathSymbolMap map[string]string
}

// ReflectInterface uses reflection to obtain interface information for each path symbol pair in the pathSympolMap
// projRoot is the root dir where mockgen is installed as a vendor package
func ReflectInterface(projRoot string, pathSymbolMap map[string]string) (map[string]*model.Package, error) {
	// We use TempDir instead of TempFile so we can control the filename.
	tmpDir, err := ioutil.TempDir("./", "gomock_reflect_")
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
	paths := make(map[string]bool, len(pathSymbolMap))
	for p := range pathSymbolMap {
		paths[p] = true
	}
	data := reflectData{
		PathAliasMap:  uniqueAlias(paths),
		PathSymbolMap: pathSymbolMap,
	}
	var program bytes.Buffer
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

	var pkgs map[string]*model.Package
	if err := gob.NewDecoder(&stdout).Decode(&pkgs); err != nil {
		return nil, err
	}
	return pkgs, nil
}

// This program reflects on an interface value, and prints the
// gob encoding of a model.Package to standard output.
// JSON doesn't work because of the model.Type interface.
var reflectProgram = template.Must(template.New("program").Parse(`
{{$pathAliasMap := .PathAliasMap}}
{{$pathSymbolMap := .PathSymbolMap}}
package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"reflect"

	"github.com/golang/mock/mockgen/model"

	{{range $importPath, $alias := $pathAliasMap}}
	{{$alias}} "{{$importPath}}"
	{{end}}
)

func main() {
	its := []struct{
		path, sym string
		typ 	  reflect.Type
	}{
		{{range $importPath, $symbol := $pathSymbolMap}}
		{"{{$importPath}}", "{{$symbol}}", reflect.TypeOf((*{{index $pathAliasMap $importPath}}.{{$symbol}})(nil)).Elem()},
		{{end}}
	}

	pkgs := make(map[string]*model.Package, {{len $pathSymbolMap}})
	{{range $importPath, $symbol := $pathSymbolMap}}
	pkgs["{{$importPath}}"] = &model.Package{
		Name: "{{index $pathAliasMap $importPath}}",
	}
	{{end}}

	stderr := os.Stderr
	for _, it := range its {
		intf, err := model.InterfaceFromInterfaceType(it.typ)
		if err != nil {
			fmt.Fprintf(stderr, "Reflection: %v\n", err)
			os.Exit(1)
		}
		intf.Name = it.sym
		pkgs[it.path].Interfaces = []*model.Interface{intf}
	}
	if err := gob.NewEncoder(os.Stdout).Encode(pkgs); err != nil {
		fmt.Fprintf(stderr, "gob encode: %v\n", err)
		os.Exit(1)
	}
}
`))
