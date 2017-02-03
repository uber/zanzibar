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

package gencode_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/lib/gencode"
)

var fooThrift = "/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/idl/github.com/uber/zanzibar/clients/foo/foo.thrift"

func newPackageHelper() *gencode.PackageHelper {
	return &gencode.PackageHelper{
		ThriftRootDir:   "examples/example-gateway/idl/github.com",
		TypeFileRootDir: "examples/example-gateway/gen-code",
		TargetGenDir:    "examples/example-gateway/target/",
	}
}

func TestImportPath(t *testing.T) {
	h := newPackageHelper()
	p, err := h.TypeImportPath(fooThrift)
	assert.Nil(t, err, "should not return error")
	assert.Equal(t, "github.com/uber/zanzibar/examples/example-gateway/gen-code/uber/zanzibar/clients/foo/foo", p, "wrong type import path")
	_, err = h.TypeImportPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/idl/github.com/uber/zanzibar/clients/foo/foo.go")
	assert.Error(t, err, "should return error for not a thrift file")
	_, err = h.TypeImportPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/zanzibar/clients/foo/foo.thrift")
	assert.Error(t, err, "should return error for not in IDL dir")
}

func TestTypePackageName(t *testing.T) {
	h := newPackageHelper()
	packageName, err := h.TypePackageName(fooThrift)
	assert.Nil(t, err, "should not return error")
	assert.Equal(t, "foo", packageName, "wrong package name")
	_, err = h.TypeImportPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/idl/github.com/uber/zanzibar/clients/foo/foo.txt")
	assert.Error(t, err, "should return error for not a thrift file")
}

func TestGenPath(t *testing.T) {
	h := newPackageHelper()
	p, err := h.TargetGenPath(fooThrift)
	assert.Nil(t, err, "should not return error")
	assert.Equal(t, "examples/example-gateway/target/uber/zanzibar/clients/foo/foo.go", p, "wrong generated code path")
	_, err = h.TargetGenPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/idl/github.com/uber/zanzibar/clients/foo/foo.go")
	assert.Error(t, err, "should return error for not a thrift file")
	_, err = h.TargetGenPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/zanzibar/clients/foo/foo.thrift")
	assert.Error(t, err, "should return error for not in IDL dir")
}
