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

package repository

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
	testlib "github.com/uber/zanzibar/test/lib"
)

const endpointUpdateRequestDir = "../../examples/example-gateway/endpoints"

func TestUpdateEndpoint(t *testing.T) {
	tempDir, err := copyExample("")
	t.Logf("Temp dir is created at %s\n", tempDir)
	if !assert.NoError(t, err, "Failed to copy example.") {
		return
	}
	requestFiles := []string{
		"googlenow/add_credentials.json",
		"googlenow/check_credentials.json",
		"bar/arg_not_struct.json",
		"baz/call.json",
		"baz/compare.json",
		"bar/normal.json",
	}
	for _, file := range requestFiles {
		t.Logf("Test request in %q\n", file)
		testUpdateEndpointConfig(t, tempDir, filepath.Join(endpointUpdateRequestDir, file))
	}
}

func testUpdateEndpointConfig(t *testing.T, tempDir string, requestFile string) {
	req := &EndpointConfig{}
	err := readJSONFile(requestFile, req)
	assert.NoError(t, err, "Failed to unmarshal endpoint config.")
	r := &Repository{
		localDir: tempDir,
	}
	endpointCfgDir := "endpoints"
	err = r.WriteEndpointConfig(endpointCfgDir, req, "{{placeholder}}")
	if !assert.NoError(t, err, "Failed to write endpoint config.") {
		return
	}
	jsonFile := codegen.CamelToSnake(req.HandleID) + ".json"
	endpointConfigActFile := filepath.Join(tempDir, endpointCfgDir, req.ID, jsonFile)
	actualendpointCfg, err := ioutil.ReadFile(endpointConfigActFile)
	if !assert.NoError(t, err, "Failed to read endpoint config file.") {
		return
	}
	endpointConfigExpFile := filepath.Join(exampleGateway, endpointCfgDir, req.ID, jsonFile)
	testlib.CompareGoldenFile(t, endpointConfigExpFile, actualendpointCfg)

	endpointGroupCfg, err := ioutil.ReadFile(filepath.Join(tempDir, endpointCfgDir, req.ID, endpointConfigFileName))
	if !assert.NoError(t, err, "Failed to read endpoint module config file.") {
		return
	}
	endpointGroupExpFile := filepath.Join(exampleGateway, endpointCfgDir, req.ID, endpointConfigFileName)
	testlib.CompareGoldenFile(t, endpointGroupExpFile, endpointGroupCfg)

	serviceGroupCfg, err := ioutil.ReadFile(filepath.Join(exampleGateway, endpointCfgDir, "../services", "example-gateway", serviceConfigFileName))
	if !assert.NoError(t, err, "Failed to read endpoint module config file.") {
		return
	}
	serviceGroupExpFile := filepath.Join(exampleGateway, endpointCfgDir, "../services", "example-gateway", serviceConfigFileName)
	testlib.CompareGoldenFile(t, serviceGroupExpFile, serviceGroupCfg)
}

func TestUpdateEndpointBadMiddlewareConfig(t *testing.T) {
	tempDir, err := copyExample("")
	t.Logf("Temp dir is created at %s\n", tempDir)
	if !assert.NoError(t, err, "Failed to copy example.") {
		return
	}
	requestFiles := []string{
		"bar/normal.json",
	}
	for _, file := range requestFiles {
		t.Logf("Test request in %q\n", file)
		testUpdateEndpointConfig(t, tempDir, filepath.Join(endpointUpdateRequestDir, file))
	}
}
