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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
)

func TestGenerateBar(t *testing.T) {
	assert.NoError(t, os.RemoveAll(tmpDir), "failed to clean temporary directory")
	if err := os.MkdirAll(tmpDir, os.ModePerm); !assert.NoError(t, err, "failed to create temporary directory", err) {
		return
	}

	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Logf("error removing temporary directory %q: %s", tmpDir, err.Error())
			t.Fail()
		}
	}()

	relativeGatewayPath := "../examples/example-gateway"
	absGatewayPath, err := filepath.Abs(relativeGatewayPath)
	if !assert.NoError(t, err, "failed to get abs path %s", err) {
		return
	}

	packageRoot := "github.com/uber/zanzibar/examples/example-gateway"
	options := &codegen.PackageHelperOptions{
		RelTargetGenDir:               tmpDir,
		RelMiddlewareConfigDir:        "./middlewares",
		RelDefaultMiddlewareConfigDir: "./middlewares/default",
		CopyrightHeader:               testCopyrightHeader,
		GenCodePackage: map[string]string{
			".thrift": packageRoot + "/build/gen-code",
			".proto":  packageRoot + "/build/gen-code",
		},
		TraceKey: "trace-key",
		ModuleSearchPaths: map[string][]string{
			"client":     {"clients/*"},
			"middleware": {"middlewares/*"},
			"endpoint":   {"endpoints/*"},
		},
		ModuleIdlSubDir: map[string]string{
			"endpoints": "endpoints-idl",
			"default":   "clients-idl",
		},
	}

	packageHelper, err := codegen.NewPackageHelper(
		packageRoot,
		absGatewayPath,
		options,
	)
	if !assert.NoError(t, err, "failed to create package helper", err) {
		return
	}

	moduleSystem, err := codegen.NewDefaultModuleSystem(packageHelper, false)
	if !assert.NoError(t, err, "failed to create module system", err) {
		return
	}

	resolvedModules, buildErr := moduleSystem.GenerateBuild(
		"github.com/uber/zanzibar/examples/example-gateway",
		absGatewayPath,
		packageHelper.CodeGenTargetPath(),
		codegen.Options{
			CommitChange: true,
		},
	)
	t.Logf("resolved moduels: %+v", resolvedModules)
	if !assert.NoError(t, buildErr, "failed to create clients init %s", buildErr) {
		return
	}

	if !assert.NoError(t, err, "failed to create gateway spec %s", err) {
		return
	}

	_, err = ioutil.ReadDir(
		filepath.Join(absGatewayPath, tmpDir, "endpoints", "bar"),
	)
	if !assert.NoError(t, err, "cannot read dir %s", err) {
		return
	}
}
