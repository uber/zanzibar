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
	"os"
	"testing"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
)

const clientGenFile = "data/bar_client.gogen"

func TestGenerate(t *testing.T) {
	tmpl, err := gencode.NewTemplate("../../codegen/templates/*.tmpl")
	if !assert.NoError(t, err, "failed to create template %s", err) {
		return
	}
	tmpDir, err := ioutil.TempDir("", "example")
	if !assert.NoError(t, err, "failed to create tmp dir") {
		return
	}
	pkgHelper := &gencode.PackageHelper{
		ThriftRootDir:   "examples/example-gateway/idl/github.com",
		TypeFileRootDir: "examples/example-gateway/gen-code",
		TargetGenDir:    tmpDir,
	}
	file, err := tmpl.GenerateClientFile("../../examples/example-gateway/idl/github.com/uber/zanzibar/clients/bar/bar.thrift", pkgHelper)
	if !assert.NoError(t, err, "failed to generate code %s", err) {
		return
	}
	b, err := ioutil.ReadFile(file)
	if !assert.NoError(t, err, "failed to read genereated file %s", err) {
		return
	}
	CompareGoldenFile(t, clientGenFile, b, *updateGoldenFile)
	assert.NoError(t, os.RemoveAll(tmpDir), "failed to clean temporary directory")
}
