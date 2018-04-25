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

package codegen_test

import (
	"io/ioutil"
	"os"
	"testing"

	"path/filepath"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
	testlib "github.com/uber/zanzibar/test/lib"
)

const (
	clientThrift   = "../examples/example-gateway/idl/github.com/uber/zanzibar/clients/bar/bar.thrift"
	endpointThrift = "../examples/example-gateway/idl/github.com/uber/zanzibar/endpoints/bar/bar.thrift"
	tmpDir         = "../../.tmp_gen"
)

func cmpGoldenFile(t *testing.T, actualFile string, goldenFileDir string) {
	b, err := ioutil.ReadFile(actualFile)
	if !assert.NoError(t, err, "failed to read genereated file %s", err) {
		return
	}
	goldenFile := filepath.Join(goldenFileDir, filepath.Base(actualFile)+"gen")
	testlib.CompareGoldenFile(t, goldenFile, b)
}

func TestGenerateBar(t *testing.T) {
	assert.NoError(t, os.RemoveAll(tmpDir), "failed to clean temporary directory")
	if err := os.MkdirAll(tmpDir, os.ModePerm); !assert.NoError(t, err, "failed to create temporary directory", err) {
		return
	}

	relativeGatewayPath := "../examples/example-gateway"
	absGatewayPath, err := filepath.Abs(relativeGatewayPath)
	if !assert.NoError(t, err, "failed to get abs path %s", err) {
		return
	}

	packageHelper, err := codegen.NewPackageHelper(
		"github.com/uber/zanzibar/examples/example-gateway",
		"github.com/uber/zanzibar",
		absGatewayPath,
		"middlewares",
		"./idl",
		"github.com/uber/zanzibar/examples/example-gateway/build/gen-code",
		tmpDir,
		testCopyrightHeader,
		"zanzibar",
		"X-Zanzibar-Use-Staging",
		"trace-key",
		nil,
		nil,
	)
	if !assert.NoError(t, err, "failed to create package helper", err) {
		return
	}

	moduleSystem, err := codegen.NewDefaultModuleSystem(packageHelper)
	if !assert.NoError(t, err, "failed to create module system", err) {
		return
	}

	_, buildErr := moduleSystem.GenerateBuild(
		"github.com/uber/zanzibar/examples/example-gateway",
		absGatewayPath,
		packageHelper.CodeGenTargetPath(),
		true,
	)
	if !assert.NoError(t, buildErr, "failed to create clients init %s", buildErr) {
		return
	}

	if !assert.NoError(t, err, "failed to create gateway spec %s", err) {
		return
	}

	cmpGoldenFile(
		t,
		filepath.Join(absGatewayPath, tmpDir, "clients", "bar", "bar.go"),
		"./test_data/clients",
	)

	/*
		err = gateway.GenerateEndpointRegisterFile()
		if !assert.NoError(t, err, "failed to create endpoint index %s", err) {
			return
		}
	*/

	endpoints, err := ioutil.ReadDir(
		filepath.Join(absGatewayPath, tmpDir, "endpoints", "bar"),
	)
	if !assert.NoError(t, err, "cannot read dir %s", err) {
		return
	}

	for _, file := range endpoints {
		// TODO: if test fixtures not present do not generate golden file
		if file.IsDir() || isTestFile(file.Name()) {
			continue
		}

		cmpGoldenFile(
			t,
			filepath.Join(absGatewayPath, tmpDir, "endpoints", "bar", file.Name()),
			"./test_data/endpoints",
		)
	}
	cmpGoldenFile(
		t,
		filepath.Join(absGatewayPath, tmpDir, "endpoints", "bar", "bar_structs.go"),
		"./test_data/endpoints",
	)

	for _, file := range endpoints {
		if file.IsDir() || !isTestFile(file.Name()) {
			continue
		}

		cmpGoldenFile(
			t,
			filepath.Join(absGatewayPath, tmpDir, "endpoints", "bar", file.Name()),
			"./test_data/endpoint_tests",
		)
	}
}

func isTestFile(filename string) bool {
	if len(filename) < 8 {
		return false
	}
	footer := filename[len(filename)-8:]
	return footer == "_test.go"
}
