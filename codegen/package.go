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

package codegen

import (
	"path"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

// PackageHelper manages the mapping from thrift file to generated type code and service code.
type PackageHelper struct {
	// The root directory containing thrift files.
	ThriftRootDir string
	// The root directory where all files of go types are generated.
	TypeFileRootDir string
	// The directory to put the generated service code.
	TargetGenDir string
}

// TypeImportPath returns the Go import path for types defined in a thrift file.
func (p PackageHelper) TypeImportPath(thrift string) (string, error) {
	if !strings.HasSuffix(thrift, ".thrift") {
		return "", errors.Errorf("file %s is not .thrift", thrift)
	}
	root := path.Clean(p.ThriftRootDir)
	idx := strings.Index(thrift, root)
	if idx == -1 {
		return "", errors.Errorf("file %s is not in thrift dir", thrift)
	}
	return path.Join("github.com/uber/zanzibar", p.TypeFileRootDir, thrift[idx+len(root):len(thrift)-7]), nil
}

// TypePackageName returns the package name that defines the type.
func (p PackageHelper) TypePackageName(thrift string) (string, error) {
	if !strings.HasSuffix(thrift, ".thrift") {
		return "", errors.Errorf("file %s is not .thrift", thrift)

	}
	file := path.Base(thrift)
	return file[:len(file)-7], nil
}

// TargetGenPath returns the path for generated file for services in a thrift file.
func (p PackageHelper) TargetGenPath(thrift string) (string, error) {
	if !strings.HasSuffix(thrift, ".thrift") {
		return "", errors.Errorf("file %s is not .thrift", thrift)
	}
	root := path.Clean(p.ThriftRootDir)
	idx := strings.Index(thrift, root)
	if idx == -1 {
		return "", errors.Errorf("file %s is not in thrift dir %s", thrift, root)
	}
	goFile := strings.Replace(thrift[idx+len(root):], ".thrift", ".go", -1)
	return path.Join(p.TargetGenDir, goFile), nil
}

// TypeFullName returns the referred Go type name in generated code from curThriftFile.
func (p PackageHelper) TypeFullName(curThriftFile string, typeSpec compile.TypeSpec) (string, error) {
	if typeSpec == nil {
		return "", nil
	}
	tfile := typeSpec.ThriftFile()

	if tfile == "" {
		return typeSpec.ThriftName(), nil
	}

	// if tfile == curThriftFile || tfile == "" {
	// 	return typeSpec.ThriftName(), nil
	// }

	pkg, err := p.TypePackageName(tfile)
	if err != nil {
		return "", errors.Wrap(err, "failed to get the full type")
	}
	return pkg + "." + typeSpec.ThriftName(), nil
}
