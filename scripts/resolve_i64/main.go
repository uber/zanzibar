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

package main

import (
	"fmt"
	"os"

	//"github.com/uber/zanzibar/codegen"
	"go.uber.org/thriftrw/compile"
	"html/template"
	"bytes"
	"io/ioutil"
	"github.com/uber/zanzibar/runtime"
	"runtime"
	"path/filepath"
	"strings"
	"github.com/uber/zanzibar/codegen"
	"github.com/pkg/errors"
)

const (
	templateFileName = "i64.tmpl"
	outFileName = "/types_i64.go"
)

type naivePackageNameResolver struct {
}

func (r *naivePackageNameResolver) TypePackageName(
	thriftFile string,
) (string, error) {
	if thriftFile[0] == '.' {
		return "", errors.Errorf("Naive does not support relative imports")
	}

	return "", nil
}

// Meta is the struct container for i64 related meta data and package name
type Meta struct {
	PackageName string
	Types []I64Struct
}

// I64Struct is the struct container for i64 related meta data
type I64Struct struct {
	IsLong bool
	IsTimestamp bool
	TypedefType string
}

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)
	return zanzibar.GetDirnameFromRuntimeCaller(file)
}

func main() {
	thriftFile := os.Args[1]
	module, err := compile.Compile(thriftFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse thrift file: %s", thriftFile))
	}
	meta := &Meta{}

	s := strings.TrimSuffix(thriftFile, ".thrift")
	s = strings.Replace(s, "/idl/", "/build/gen-code/", 1)
	meta.PackageName = filepath.Base(s)

	for _, typeDef := range module.Types {
		t, ok := typeDef.(*compile.TypedefSpec)
		if !ok {
			continue
		}
		i64Struct := I64Struct{}
		if t.Target != nil {
			typThriftAnnotation := t.Target.ThriftAnnotations()
			if typThriftAnnotation != nil {
				p := naivePackageNameResolver{}
				refType, err := codegen.GoReferenceType(&p, t)
				if err != nil {
					fmt.Fprintln(os.Stderr, fmt.Errorf("error parsing reference type: %s", err.Error()))
					os.Exit(1)
					return
				}

				i64Struct.TypedefType = refType[1:]
				if typThriftAnnotation["json.type"] == "Long" {
					i64Struct.IsLong = true
				}
				if typThriftAnnotation["json.type"] == "Date" {
					i64Struct.IsTimestamp = true
				}
			}
		}
		if i64Struct.IsTimestamp || i64Struct.IsLong {
			meta.Types = append(meta.Types, i64Struct)
		}
	}
	tmpl, err := template.ParseFiles(filepath.Join(getDirName(), templateFileName))
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf(
			"error parsing template %s: %s", templateFileName, err.Error()))
		os.Exit(1)
		return
	}
	tplBuffer := bytes.NewBuffer(nil)
	err = tmpl.Execute(tplBuffer, meta)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf(
			"error executing template %s: %s", templateFileName, err.Error()))
		os.Exit(1)
		return
	}
	outName := s + outFileName
	err = ioutil.WriteFile(outName, tplBuffer.Bytes(), 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf(
			"error writing file to %s, err: %s", templateFileName, err.Error()))
		os.Exit(1)
		return
	}

	err = codegen.FormatGoFile(outName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
