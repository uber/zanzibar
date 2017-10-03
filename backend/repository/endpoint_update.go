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
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/codegen"
)

// WriteEndpointConfig writes endpoint configs and its test cases into a runtime repository and
// also updates the meta json file for all endpoints.
func (r *Repository) WriteEndpointConfig(
	endpointCfgDir string,
	config *EndpointConfig,
	thriftFileSha string,
) error {
	if err := r.validateEndpointCfg(config); err != nil {
		return errors.Wrap(err, "invalid endpoint config")
	}
	r.Lock()
	defer r.Unlock()
	endpointDir := codegen.CamelToSnake(strings.TrimSuffix(config.ID, "."+config.HandleID))
	dir := filepath.Join(r.absPath(endpointCfgDir), endpointDir)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "failed to create endpoint config dir")
	}
	baseName := codegen.CamelToSnake(config.HandleID)
	fileName := baseName + ".json"
	testName := baseName + "_test.json"
	config.ThriftFileSha = thriftFileSha
	err = writeToJSONFile(filepath.Join(dir, testName), config.TestFixtures)
	if err != nil {
		return errors.Wrap(err, "failed to write to endpoint test cases file")
	}
	err = writeToJSONFile(filepath.Join(dir, fileName), config)
	if err != nil {
		return errors.Wrap(err, "failed to write to endpoint config file")
	}
	err = updateEndpointMetaJSON(dir, endpointConfigFileName, fileName, config)
	if err != nil {
		return errors.Wrap(err, "failed to write endpoint group configuration")
	}
	return nil
}

func (r *Repository) validateEndpointCfg(req *EndpointConfig) error {
	gatewayConfig, err := r.LatestGatewayConfig()
	if err != nil {
		return errors.Wrap(err, "invalid configuration before updating endpoint")
	}
	clientCfg, ok := gatewayConfig.Clients[req.ClientID]
	if !ok {
		return NewRequestError(
			ClientID, errors.Errorf("can't find client %q", req.ClientID))
	}
	if clientCfg.Type == HTTP {
		req.WorkflowType = "httpClient"
	} else if clientCfg.Type == TCHANNEL {
		req.WorkflowType = "tchannelClient"
	} else {
		return NewRequestError(ClientType,
			errors.Errorf("client type %q is not supported", clientCfg.Type))
	}

	for _, mid := range req.Middlewares {
		for _, midCfg := range r.gatewayConfig.Middlewares {
			if mid.Name == midCfg.Name {
				err := codegen.SchemaValidateGo(midCfg.SchemaFile, mid.Options)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// updateEndpointMetaJSON adds an endpoint in the meta json file or updates the config for an exsiting endpoint.
func updateEndpointMetaJSON(configDir, metaFile, newFile string, cfg *EndpointConfig) error {
	metaFilePath := filepath.Join(configDir, metaFile)
	fileContent := new(codegen.EndpointClassConfig)
	if _, err := os.Stat(metaFilePath); !os.IsNotExist(err) {
		err := readJSONFile(metaFilePath, fileContent)
		if err != nil {
			return err
		}
	}
	var err error
	fileContent.Config.Endpoints, err = addToEndpointList(fileContent.Config.Endpoints, newFile, configDir)
	if err != nil {
		return err
	}
	if fileContent.Dependencies == nil {
		fileContent.Dependencies = make(map[string][]string)
	}
	if c := fileContent.Dependencies["client"]; !findString(cfg.ClientID, c) {
		fileContent.Dependencies["client"] = append(c, cfg.ClientID)
	}
	fileContent.Name = cfg.ID
	if fileContent.Type == "" {
		fileContent.Type = string(cfg.Type)
	}
	return writeToJSONFile(metaFilePath, fileContent)
}

// addToEndpointList adds 'newFile' to the endpoint list if it doesn't exist.
func addToEndpointList(curEndpoints []string, newFile string, configDir string) ([]string, error) {
	newFilePath, err := filepath.Abs(filepath.Join(configDir, newFile))
	if err != nil {
		return nil, err
	}
	oldFilePaths := make([]string, len(curEndpoints))
	for i, path := range curEndpoints {
		file, err := filepath.Abs(filepath.Join(configDir, path))
		if err != nil {
			return nil, err
		}
		oldFilePaths[i] = file
	}
	if !findString(newFilePath, oldFilePaths) {
		curEndpoints = append(curEndpoints, newFile)
	}
	return curEndpoints, nil
}

func findString(target string, array []string) bool {
	for _, str := range array {
		if str == target {
			return true
		}
	}
	return false
}
