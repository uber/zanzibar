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
	"io/ioutil"
	"os"
	"testing"

	"path/filepath"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
)

const (
	clientThrift   = "../examples/example-gateway/idl/github.com/uber/zanzibar/clients/bar/bar.thrift"
	endpointThrift = "../examples/example-gateway/idl/github.com/uber/zanzibar/endpoints/bar/bar.thrift"
	tmpDir         = "../.tmp_gen"
)

func cmpGoldenFile(t *testing.T, actualFile string, goldenFileDir string) {
	b, err := ioutil.ReadFile(actualFile)
	if !assert.NoError(t, err, "failed to read genereated file %s", err) {
		return
	}
	goldenFile := filepath.Join(goldenFileDir, filepath.Base(actualFile)+"gen")
	CompareGoldenFile(t, goldenFile, b, *updateGoldenFile)
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

	gateway, err := codegen.NewGatewaySpec(
		absGatewayPath,
		filepath.Join(absGatewayPath, "idl"),
		"github.com/uber/zanzibar/examples/example-gateway/build/gen-code",
		tmpDir,
		filepath.Join(absGatewayPath, "idl/github.com/uber/zanzibar"),
		"./clients",
		"./endpoints",
		"./middlewares/middleware-config.json",
		"example-gateway",
	)
	if !assert.NoError(t, err, "failed to create gateway spec %s", err) {
		return
	}

	err = gateway.GenerateClients()
	if !assert.NoError(t, err, "failed to create clients %s", err) {
		return
	}

	cmpGoldenFile(
		t,
		filepath.Join(tmpDir, "clients", "bar", "bar.go"),
		"./test_data/clients",
	)
	cmpGoldenFile(
		t,
		filepath.Join(tmpDir, "clients", "bar", "bar_structs.go"),
		"./test_data/clients",
	)

	err = gateway.GenerateEndpoints()
	if !assert.NoError(t, err, "failed to create endpoints %s", err) {
		return
	}

	endpoints, err := ioutil.ReadDir(
		filepath.Join(tmpDir, "endpoints", "bar"),
	)
	if !assert.NoError(t, err, "cannot read dir %s", err) {
		return
	}

	for _, file := range endpoints {
		footer := file.Name()[len(file.Name())-8 : len(file.Name())]
		if footer == "_test.go" {
			continue
		}

		cmpGoldenFile(
			t,
			filepath.Join(tmpDir, "endpoints", "bar", file.Name()),
			"./test_data/endpoints",
		)
	}
	cmpGoldenFile(
		t,
		filepath.Join(tmpDir, "endpoints", "bar", "bar_structs.go"),
		"./test_data/endpoints",
	)

	for _, file := range endpoints {
		footer := file.Name()[len(file.Name())-8 : len(file.Name())]
		if footer != "_test.go" {
			continue
		}

		cmpGoldenFile(
			t,
			filepath.Join(tmpDir, "endpoints", "bar", file.Name()),
			"./test_data/endpoint_tests",
		)
	}
}
