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
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
	testlib "github.com/uber/zanzibar/test/lib"
	"go.uber.org/thriftrw/compile"
)

const (
	gatewayID                    = "example-gateway"
	thriftListGoldenFile         = "../../examples/example-gateway/idls.json"
	thriftServiceGoldenFile      = "data/thrift_service_expected.json"
	clientConfigUpdatedExpFile   = "data/client_config_updated_expected.json"
	endpointConfigUpdatedExpFile = "data/endpoint_config_updated_expected.json"
)

var manager = NewTestManager()

func TestThriftFile(t *testing.T) {
	thriftFile := "clients/baz/baz.thrift"
	meta, err := manager.ThriftFile(gatewayID, thriftFile)
	if !assert.NoErrorf(t, err, "Failed to get meta for %s.", thriftFile) {
		return
	}
	assert.Equal(t, thriftFile, meta.Path)
	assert.Equal(t, "v1", meta.Version)

	_, err = manager.ThriftFile("no-such-gateway", thriftFile)
	assert.Error(t, err, "Should return an error for gateway not found.")
	_, err = manager.ThriftFile(gatewayID, "no-such-thrift")
	assert.Error(t, err, "Should return an error for thrift not found.")
}

func TestThriftList(t *testing.T) {
	meta, err := manager.ThriftList(gatewayID)
	if !assert.NoError(t, err, "Failed to get thrift list.") {
		return
	}
	b, err := json.MarshalIndent(meta, "", "\t")
	if !assert.NoError(t, err, "Failed to marshal meta.") {
		return
	}
	b = append(b, []byte("\n")...)
	testlib.CompareGoldenFile(t, thriftListGoldenFile, b)

	_, err = manager.ThriftList("no-such-gateway")
	assert.Error(t, err, "Should return an error for gateway not found.")
}

func TestIDLThriftService(t *testing.T) {
	thriftFile := "clients/googlenow/googlenow.thrift"
	serviceMap, err := manager.IDLThriftService(thriftFile)
	if !assert.NoError(t, err, "Failed to get thrift services.") {
		return
	}
	b, err := json.MarshalIndent(serviceMap, "", "\t")
	if !assert.NoError(t, err, "Failed to marshal serviceMap.") {
		return
	}
	b = append(b, []byte("\n")...)
	testlib.CompareGoldenFile(t, thriftServiceGoldenFile, b)

	_, err = manager.IDLThriftService("no-such-thrift")
	assert.Error(t, err, "Should return an error for thrift not found.")
}

func TestValidate(t *testing.T) {
	r, err := manager.NewRuntimeRepository(gatewayID)
	assert.Nil(t, err)
	b, err := ioutil.ReadFile(tchannelClientUpdateRequestFile)
	assert.Nil(t, err)
	clientReq := &ClientConfig{}
	err = json.Unmarshal(b, clientReq)
	assert.Nil(t, err)
	endpointCfgUpdateRequestFile := filepath.Join(endpointUpdateRequestDir, "baz/compare.json")
	b, err = ioutil.ReadFile(endpointCfgUpdateRequestFile)
	assert.Nil(t, err)
	endpointReq := &EndpointConfig{}
	err = json.Unmarshal(b, endpointReq)

	assert.Nil(t, err)
	req := &UpdateRequest{
		ThriftFiles:     []string{"clients/baz/baz.thrift"},
		ClientUpdates:   []ClientConfig{*clientReq},
		EndpointUpdates: []EndpointConfig{*endpointReq},
	}
	err = manager.Validate(r, req)
	assert.Nil(t, err)
}

func TestUpdateAll(t *testing.T) {
	r, err := manager.NewRuntimeRepository(gatewayID)
	if !assert.NoError(t, err, "Failed to create runtime repository.") {
		return
	}
	b, err := ioutil.ReadFile(tchannelClientUpdateRequestFile)
	assert.NoError(t, err, "Failed to read client update request file.")
	clientReq := &ClientConfig{}
	err = json.Unmarshal(b, clientReq)
	if !assert.NoError(t, err, "Failed to unmarshal client config.") {
		return
	}
	endpointCfgUpdateRequestFile := filepath.Join(endpointUpdateRequestDir, "baz/compare.json")
	b, err = ioutil.ReadFile(endpointCfgUpdateRequestFile)
	assert.NoError(t, err, "Failed to read endpoint update request file.")
	endpointReq := &EndpointConfig{}
	err = json.Unmarshal(b, endpointReq)
	if !assert.NoError(t, err, "Failed to unmarshal endpoint config.") {
		return
	}

	req := &UpdateRequest{
		ThriftFiles:     []string{"clients/baz/baz.thrift"},
		ClientUpdates:   []ClientConfig{*clientReq},
		EndpointUpdates: []EndpointConfig{*endpointReq},
	}
	clientCfgDir := "clients"
	endpointCfgDir := "endpoints"
	err = manager.UpdateAll(r, clientCfgDir, endpointCfgDir, req)
	if !assert.NoError(t, err, "Failed to update all.") {
		return
	}

	// Verify clients.
	clientName := "baz"
	clientConfigActFile := filepath.Join(r.LocalDir(), clientCfgDir, clientName, clientConfigFileName)
	actualClientCfg, err := ioutil.ReadFile(clientConfigActFile)
	if !assert.NoError(t, err, "Failed to read client config file.") {
		return
	}
	testlib.CompareGoldenFile(t, clientConfigUpdatedExpFile, actualClientCfg)

	productionJSON, err := ioutil.ReadFile(filepath.Join(r.LocalDir(), productionCfgJSONPath))
	if !assert.NoError(t, err, "Failed to read client production JSON config file.") {
		return
	}
	productionJSONExpFile := filepath.Join(exampleGateway, productionCfgJSONPath)
	testlib.CompareGoldenFile(t, productionJSONExpFile, productionJSON)

	// Verify endpoints.
	jsonFile := codegen.CamelToSnake(endpointReq.HandleID) + ".json"
	endpointConfigActFile := filepath.Join(r.LocalDir(), endpointCfgDir, endpointReq.ID, jsonFile)
	actualendpointCfg, err := ioutil.ReadFile(endpointConfigActFile)
	if !assert.NoError(t, err, "Failed to read endpoint config file.") {
		return
	}
	testlib.CompareGoldenFile(t, endpointConfigUpdatedExpFile, actualendpointCfg)

	endpointGroupCfg, err := ioutil.ReadFile(filepath.Join(r.LocalDir(), endpointCfgDir, endpointReq.ID, endpointConfigFileName))
	if !assert.NoError(t, err, "Failed to read endpoint module config file.") {
		return
	}
	endpointGroupExpFile := filepath.Join(exampleGateway, endpointCfgDir, endpointReq.ID, endpointConfigFileName)
	testlib.CompareGoldenFile(t, endpointGroupExpFile, endpointGroupCfg)

	// Try to delete a client before removing its endpoints
	clientReq = &ClientConfig{}
	err = readJSONFile(deleteBazClientRequestFile, clientReq)
	if !assert.NoError(t, err, "Failed to read delete client request json") {
		return
	}
	req = &UpdateRequest{
		ThriftFiles:     []string{},
		ClientUpdates:   []ClientConfig{*clientReq},
		EndpointUpdates: []EndpointConfig{},
	}
	err = manager.UpdateAll(r, clientCfgDir, endpointCfgDir, req)
	if !assert.Error(t, err, "Deleting a client should fail when dependent endpoints exist") {
		return
	}
}

func TestAddThriftDepedencies(t *testing.T) {
	root := "some_path/idl/"
	m1 := compile.Module{
		Name:       "m1",
		ThriftPath: filepath.Join(root, "dir1/m1.thrift"),
	}
	m2 := compile.Module{
		Name:       "m2",
		ThriftPath: filepath.Join(root, "dir1/m2.thrift"),
	}
	m3 := compile.Module{
		Name:       "m3",
		ThriftPath: filepath.Join(root, "dir3/m3.thrift"),
	}
	m4 := compile.Module{
		Name:       "m4",
		ThriftPath: filepath.Join(root, "dir3/m4.thrift"),
	}
	m1.Includes = map[string]*compile.IncludedModule{
		"m2.thrift": {
			Name:   "m2",
			Module: &m2,
		},
		"m3.thrift": {
			Name:   "m3",
			Module: &m3,
		},
	}
	m2.Includes = map[string]*compile.IncludedModule{
		"m4.thrift": {
			Name:   "m4",
			Module: &m4,
		},
	}
	m3.Includes = map[string]*compile.IncludedModule{
		"m4.thrift": {
			Name:   "m4",
			Module: &m4,
		},
	}
	// Constructs cyclic dependencies.
	m4.Includes = map[string]*compile.IncludedModule{
		"m1.thrift": {
			Name:   "m1",
			Module: &m1,
		},
	}
	meta := make(map[string]*ThriftMeta)
	err := addThriftDependencies(root, &m1, meta)
	if !assert.NoError(t, err, "Failed to add dependencies.") {
		return
	}
	assert.Equal(t, 4, len(meta), "Should have 4 thrift files")
	for _, f := range []string{"dir1/m1.thrift", "dir1/m2.thrift", "dir3/m3.thrift", "dir3/m4.thrift"} {
		_, ok := meta[f]
		assert.True(t, ok, "Should have %q in the list.", f)
	}
}

func NewTestManager() *Manager {
	idlRegistor, err := NewTempIDLRegistry()
	if err != nil {
		log.Fatalf("Failed to create IDL registry: %v\n", err)
	}
	lf := &localFetcher{
		remote: "example-gateway",
	}
	r, err := NewRepository("", lf.remote, lf, 30*time.Second)
	if err != nil {
		log.Fatalf("Failed to create gateway repository: %v\n", err)
	}

	repoMap := map[string]*Repository{
		gatewayID: r,
	}
	manager, err := NewManager(repoMap, "", idlRegistor), nil
	if err != nil {
		log.Fatalf("Failed to create test manager: %v\n", err)
	}
	return manager
}

type localFetcher struct {
	remote string
}

func (lf *localFetcher) Clone(localRoot, remote string) (string, error) {
	if remote != lf.remote {
		return "", errors.Errorf("remote is expected to be %q but got %q", lf.remote, remote)
	}

	return copyExample(localRoot)
}

func (lf *localFetcher) Update(localDir, remote string) (bool, error) {
	if remote != lf.remote {
		return false, errors.Errorf("remote is expected to be %q but got %q", lf.remote, remote)
	}
	return false, nil
}

func (lf *localFetcher) Version(localDir string) (string, error) {
	return "v1", nil
}
