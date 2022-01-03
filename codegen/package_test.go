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

package codegen_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
	"go.uber.org/thriftrw/compile"
)

const tmpDir = "../../.tmp_gen"

var fooThrift = filepath.Join(
	os.Getenv("GOPATH"),
	"/src/github.com/uber/zanzibar/",
	"examples/example-gateway/idl/",
	"clients-idl/clients/foo/foo.thrift")

var testCopyrightHeader = `// Copyright (c) 2018 Uber Technologies, Inc.
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
// THE SOFTWARE.`

func newPackageHelper(t *testing.T) *codegen.PackageHelper {
	relativeGatewayPath := "../examples/example-gateway"
	absGatewayPath, err := filepath.Abs(relativeGatewayPath)
	if !assert.NoError(t, err, "failed to get abs path %s", err) {
		return nil
	}

	packageRoot := "github.com/uber/zanzibar/examples/example-gateway"
	options := &codegen.PackageHelperOptions{
		RelTargetGenDir: tmpDir,
		CopyrightHeader: testCopyrightHeader,
		GenCodePackage: map[string]string{
			".thrift": packageRoot + "/build/gen-code",
			".proto":  packageRoot + "/build/gen-code",
		},
		TraceKey: "trace-key",
	}

	h, err := codegen.NewPackageHelper(
		packageRoot,
		absGatewayPath,
		options,
	)
	if !assert.NoError(t, err, "failed to create package helper") {
		return nil
	}
	return h
}

func TestImportPath(t *testing.T) {
	h := newPackageHelper(t)
	p, err := h.TypeImportPath(fooThrift)
	assert.Nil(t, err, "should not return error")
	assert.Equal(t, "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/foo/foo",
		p, "wrong type import path")
	_, err = h.TypeImportPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/build/idl/github.com/uber/zanzibar/clients/foo/foo.go")
	assert.Error(t, err, "should return error for not a thrift file")
	_, err = h.TypeImportPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/build/zanzibar/clients/foo/foo.thrift")
	assert.Error(t, err, "should return error for not in IDL dir")
}

func TestTypePackageName(t *testing.T) {
	h := newPackageHelper(t)
	packageName, err := h.TypePackageName(fooThrift)
	assert.Nil(t, err, "should not return error")
	assert.Equal(t, "clientsIDlClientsFooFoo", packageName, "wrong package name")
	_, err = h.TypeImportPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/build/idl/github.com/uber/zanzibar/clients/foo/foo.txt")
	assert.Error(t, err, "should return error for not a thrift file")
}

func TestEndpointTestConfigPath(t *testing.T) {
	h := newPackageHelper(t)
	p := h.EndpointTestConfigPath("foo", "bar")
	exp := "foo/bar_test.yaml"
	assert.Equal(t, exp, p, "wrong generated endpoint test config path")
}

func TestUnhashableKeyInMap(t *testing.T) {
	h := newPackageHelper(t)
	spec := &compile.MapSpec{
		KeySpec:   &compile.BinarySpec{},
		ValueSpec: &compile.StringSpec{},
	}
	typ, err := h.TypeFullName(spec)
	assert.NoError(t, err)
	assert.Equal(t, "[]struct{Key []byte; Value string}", typ)
}

func TestUnhashableValueInSet(t *testing.T) {
	h := newPackageHelper(t)
	spec := &compile.SetSpec{
		ValueSpec: &compile.BinarySpec{},
	}
	typ, err := h.TypeFullName(spec)
	assert.NoError(t, err)
	assert.Equal(t, "[][]byte", typ)
}

func TestGoCustomTypeError(t *testing.T) {
	h := newPackageHelper(t)
	spec := &compile.StructSpec{}
	_, err := h.TypeFullName(spec)
	assert.Error(t, err)
	assert.Equal(t, "GoCustomType called with native type (*compile.StructSpec) &{false   0 []  map[]}", err.Error())
}
