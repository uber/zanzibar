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

package codegen_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
)

var fooThrift = filepath.Join(
	os.Getenv("GOPATH"),
	"/src/github.com/uber/zanzibar/",
	"examples/example-gateway/idl/",
	"github.com/uber/zanzibar/clients/foo/foo.thrift")

func newPackageHelper(t *testing.T) *codegen.PackageHelper {
	relativeGatewayPath := "../../examples/example-gateway"
	absGatewayPath, err := filepath.Abs(relativeGatewayPath)
	if !assert.NoError(t, err, "failed to get abs path %s", err) {
		return nil
	}

	h, err := codegen.NewPackageHelper(
		filepath.Join(absGatewayPath, "idl"),
		"github.com/uber/zanzibar/examples/example-gateway/gen-code",
		tmpDir,
		filepath.Join(absGatewayPath, "idl/github.com/uber/zanzibar"),
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
	assert.Equal(t, "github.com/uber/zanzibar/examples/example-gateway/gen-code/github.com/uber/zanzibar/clients/foo/foo", p, "wrong type import path")
	_, err = h.TypeImportPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/idl/github.com/uber/zanzibar/clients/foo/foo.go")
	assert.Error(t, err, "should return error for not a thrift file")
	_, err = h.TypeImportPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/zanzibar/clients/foo/foo.thrift")
	assert.Error(t, err, "should return error for not in IDL dir")
}

func TestTypePackageName(t *testing.T) {
	h := newPackageHelper(t)
	packageName, err := h.TypePackageName(fooThrift)
	assert.Nil(t, err, "should not return error")
	assert.Equal(t, "foo", packageName, "wrong package name")
	_, err = h.TypeImportPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/idl/github.com/uber/zanzibar/clients/foo/foo.txt")
	assert.Error(t, err, "should return error for not a thrift file")
}

func TestPackageGenPath(t *testing.T) {
	h := newPackageHelper(t)
	p, err := h.PackageGenPath(fooThrift)
	assert.Nil(t, err, "should not return error")
	exp := "github.com/uber/zanzibar/.tmp_gen/clients/foo"
	assert.Equal(t, exp, p, "wrong generated Go package path")
	_, err = h.PackageGenPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/idl/github.com/uber/zanzibar/clients/foo/foo.go")
	assert.Error(t, err, "should return error for not a thrift file")
	_, err = h.PackageGenPath("/Users/xxx/go/src/github.com/uber/zanzibar/examples/example-gateway/zanzibar/clients/foo/foo.thrift")
	assert.Error(t, err, "should return error for not in IDL dir")
}
