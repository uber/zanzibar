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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/codegen"
)

// ClientJSONConfig defines the JSON content of the configuration file.
type ClientJSONConfig struct {
	Name   string       `json:"name"`
	Type   string       `json:"type"`
	Config *configField `json:"config"`
}

type configField struct {
	ThriftFile     string            `json:"thriftFile"`
	ThriftFileSha  string            `json:"thriftFileSha"`
	ExposedMethods map[string]string `json:"exposedMethods,omitempty"`
}

// NewClientConfigJSON converts ClientConfig to ClientJSONConfig.
func NewClientConfigJSON(cfg *ClientConfig) *ClientJSONConfig {
	cfgJSON := &ClientJSONConfig{
		Name: cfg.Name,
		Type: string(cfg.Type),
		Config: &configField{
			ThriftFile:     cfg.ThriftFile,
			ExposedMethods: cfg.ExposedMethods,
		},
	}
	return cfgJSON
}

// UpdateClientConfigs updates JSON configuration files for a client update request.
func (r *Repository) UpdateClientConfigs(req *ClientConfig, clientCfgDir, thriftFileSha string) error {
	if err := validateClientUpdateRequest(req); err != nil {
		return err
	}
	cfgJSON := NewClientConfigJSON(req)
	// Expose all methods in the thrift file by default.
	if len(cfgJSON.Config.ExposedMethods) == 0 {
		cfg, err := r.LatestGatewayConfig()
		if err != nil {
			return errors.Wrap(err, "invalid configuration before updating client")
		}
		exposedMethods, err := allExposedMethods(cfg.ThriftServices, req.ThriftFile)
		if err != nil {
			return errors.Wrapf(err, "failed to generate all methods in thrift file %s", req.ThriftFile)
		}
		cfgJSON.Config.ExposedMethods = exposedMethods
	}
	cfgJSON.Config.ThriftFileSha = thriftFileSha
	clientPath := filepath.Join(r.absPath(clientCfgDir), cfgJSON.Name)
	r.Lock()
	defer r.Unlock()
	if err := os.MkdirAll(clientPath, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create client config dir")
	}
	if err := writeToJSONFile(filepath.Join(clientPath, clientConfigFileName), cfgJSON); err != nil {
		return errors.Wrapf(err, "failed to write config for the client %q", cfgJSON.Name)
	}
	if err := WriteClientModuleJSON(r.absPath(clientCfgDir)); err != nil {
		return errors.Wrap(err, "failed to write module config for all clients")
	}
	if err := UpdateProductionConfigJSON(req, r.absPath(productionCfgJSONPath)); err != nil {
		return errors.Wrap(err, "failed to update gateway production config")
	}
	return nil
}

func allExposedMethods(thriftServices map[string]map[string]*ThriftService, thriftFile string) (map[string]string, error) {
	serviceMap, ok := thriftServices[thriftFile]
	if !ok {
		return nil, errors.Errorf("thrift file %q not found", thriftFile)
	}
	exposedMethods := map[string]string{}
	for _, tservice := range serviceMap {
		for _, method := range tservice.Methods {
			exposedName := tservice.Name + "::" + method.Name
			if pre, ok := exposedMethods[method.Name]; ok {
				return nil, errors.Errorf("duplicated method name for %q and %q", pre, exposedName)
			}
			exposedMethods[method.Name] = exposedName
		}
	}
	return exposedMethods, nil
}

func validateClientUpdateRequest(req *ClientConfig) error {
	if req.Type == "tchannel" && req.ServiceName == "" {
		return errors.New("invalid request: muttley name is required for tchannel client")
	}
	if req.IP == "" || req.Port == 0 {
		return errors.New("invalid request: ip and port are required")
	}
	if req.Type == "tchannel" && req.Timeout == 0 {
		req.Timeout = 10000
	}
	if req.Type == "tchannel" && req.TimeoutPerAttempt == 0 {
		req.TimeoutPerAttempt = 10000
	}
	return nil
}

// WriteClientModuleJSON writes the JSON file for the module to contain all clients.
func WriteClientModuleJSON(clientCfgDir string) error {
	files, err := ioutil.ReadDir(clientCfgDir)
	if err != nil {
		return err
	}
	subDirs := []string{}
	for _, file := range files {
		if file.IsDir() {
			subDirs = append(subDirs, file.Name())
		}
	}
	sort.Strings(subDirs)
	content := &codegen.ClientClassConfig{
		Name:   "clients",
		Type:   "init",
		Config: map[string]interface{}{},
		Dependencies: codegen.ClientDependencies{
			Client: subDirs,
		},
	}
	return writeToJSONFile(filepath.Join(clientCfgDir, clientModuleFileName), content)
}

// UpdateProductionConfigJSON updates the production JSON config with client updates.
func UpdateProductionConfigJSON(req *ClientConfig, productionCfgJSONPath string) error {
	content := map[string]interface{}{}
	if err := readJSONFile(productionCfgJSONPath, &content); err != nil {
		return err
	}
	// Update fields related to a client.
	prefix := "clients." + req.Name + "."
	if req.Type == "tchannel" {
		content[prefix+"serviceName"] = req.ServiceName
		content[prefix+"timeout"] = req.Timeout
		content[prefix+"timeoutPerAttempt"] = req.TimeoutPerAttempt
	}
	content[prefix+"port"] = req.Port
	content[prefix+"ip"] = req.IP
	return writeToJSONFile(productionCfgJSONPath, content)
}

// writeToJSONFile writes content into a json file.
func writeToJSONFile(file string, content interface{}) error {
	b, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		return errors.Wrapf(err, "failed to marshal the content for file %q", file)
	}
	b = append(b, []byte("\n")...)
	if err = ioutil.WriteFile(file, b, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to write to file %q", file)
	}
	return nil
}

func readJSONFile(file string, content interface{}) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "failed to read file %q", file)
	}
	err = json.Unmarshal(b, &content)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal file %q", file)
	}
	return nil
}
