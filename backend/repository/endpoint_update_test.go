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

package repository

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

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

func TestDeleteEndpoint(t *testing.T) {
	tempDir, err := copyExample("")
	t.Logf("Temp dir is created at %s\n", tempDir)
	if !assert.NoError(t, err, "Failed to copy example") {
		return
	}
	r := &Repository{
		localDir: tempDir,
	}

	testDeleteEndpoint(t, r, "bar", "argNotStruct")

	testDeleteEndpoint(t, r, "baz", "trans")
	testDeleteEndpoint(t, r, "baz", "call")
	testDeleteEndpoint(t, r, "baz", "ping")
	testDeleteEndpoint(t, r, "baz", "compare")
	testDeleteEndpoint(t, r, "baz", "sillyNoop")
}

func testDeleteEndpoint(t *testing.T, r *Repository, endpointGroup, endpointName string) {
	gatewayCfg, err := r.LatestGatewayConfig()
	if !assert.NoError(t, err, "fetch gateway config error") {
		return
	}

	// Make sure endpoint is there
	if !assert.Equal(t, true, checkEndpointExists(t, r, endpointGroup, endpointName)) {
		return
	}
	configDir := filepath.Join(r.absPath("endpoints"), codegen.CamelToSnake(endpointGroup))
	configFile := filepath.Join(configDir, codegen.CamelToSnake(endpointName)+".json")
	groupConfigFile := filepath.Join(configDir, "endpoint-config.json")
	groupConfig := &codegen.EndpointClassConfig{}
	serviceConfigFile := filepath.Join(r.absPath("endpoints"), "..", "/services/", gatewayCfg.ID, serviceConfigFileName)
	serviceConfig := &codegen.EndpointClassConfig{}
	if !assert.NoError(t, readJSONFile(groupConfigFile, groupConfig), "cannot read endpoint group config") {
		return
	}
	_, err = ioutil.ReadFile(configFile)
	if !assert.NoError(t, err, "endpoint config file missing") {
		return
	}
	if !assert.NoError(t, readJSONFile(serviceConfigFile, serviceConfig), "cannot read service config") {
		return
	}
	if !assert.NotEqual(t, -1, findString(groupConfig.Name, serviceConfig.Dependencies["endpoint"]), "service config missing endpoint") {
		return
	}

	// Try to delete endpoint
	if !assert.NoError(t, r.deleteEndpointConfig("endpoints", endpointGroup+"."+endpointName), "delete endpoint failed") {
		return
	}

	// Make sure endpoint is gone
	if !assert.Equal(t, false, checkEndpointExists(t, r, endpointGroup, endpointName)) {
		return
	}
	_, err = ioutil.ReadFile(configFile)
	if !assert.Error(t, err, "endpoint file still exists") {
		return
	}

	// If that was last endpoint in group, entire directory should be gone, service config should be updated
	if len(groupConfig.Config.Endpoints) == 1 {
		_, err = ioutil.ReadDir(configDir)
		if !assert.Error(t, err, "endpoint directory should be deleted") {
			return
		}
		if !assert.NoError(t, readJSONFile(serviceConfigFile, serviceConfig), "cannot read service config") {
			return
		}
		if !assert.Equal(t, -1, findString(groupConfig.Name, serviceConfig.Dependencies["endpoint"]), "service config not updated") {
			return
		}
	}

}

func testUpdateEndpointConfig(t *testing.T, tempDir string, requestFile string) {
	req := &EndpointConfig{}
	err := readJSONFile(requestFile, req)
	assert.NoError(t, err, "Failed to unmarshal endpoint config.")
	r := &Repository{
		localDir:        tempDir,
		refreshInterval: time.Hour * 65535,
	}
	r.meta.Store(&meta{
		lastUpdate: time.Now(),
		version:    "test",
	})
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

func checkEndpointExists(t *testing.T, r *Repository, endpointGroup, endpointName string) bool {
	gatewayCfg, err := r.LatestGatewayConfig()
	if !assert.NoError(t, err, "fetch gateway config error") {
		return false
	}

	for _, endpoint := range gatewayCfg.Endpoints {
		if endpoint.ID == endpointGroup && endpoint.HandleID == endpointName {
			return true
		}
	}
	return false
}
