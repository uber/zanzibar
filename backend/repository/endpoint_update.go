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
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/codegen"
	reqerr "github.com/uber/zanzibar/codegen/errors"
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
	gatewayCfg, err := r.GatewayConfig()
	if err != nil {
		return errors.Wrap(err, "loading gateway config error")
	}
	endpointDir := codegen.CamelToSnake(strings.TrimSuffix(config.ID, "."+config.HandleID))
	dir := filepath.Join(r.absPath(endpointCfgDir), endpointDir)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "failed to create endpoint config dir")
	}
	fileName := codegen.CamelToSnake(config.HandleID) + ".json"
	config.ThriftFileSha = thriftFileSha
	err = writeToJSONFile(filepath.Join(dir, fileName), config)
	if err != nil {
		return errors.Wrap(err, "failed to write to endpoint config file")
	}
	err = updateEndpointMetaJSON(dir, endpointConfigFileName, fileName, config)
	if err != nil {
		return errors.Wrap(err, "failed to write endpoint group configuration")
	}
	serviceConfigDir := filepath.Join(r.absPath(endpointCfgDir), "..", "/services/", gatewayCfg.ID)
	err = updateServiceMetaJSON(serviceConfigDir, serviceConfigFileName, config)
	if err != nil {
		return errors.Wrap(err, "failed to write service group configuration")
	}
	return nil
}

// deleteEndpointConfig deletes an individual endpoint config, updates endpoint group config and endpoint meta json
// endpointName must have the form endpointId.handleId
func (r *Repository) deleteEndpointConfig(endpointCfgDir string, endpointName string) error {
	gatewayCfg, err := r.LatestGatewayConfig()
	if err != nil {
		return errors.Wrap(err, "loading gateway config error")
	}
	nameParts := strings.Split(endpointName, ".")
	if len(nameParts) != 2 {
		return errors.Errorf("endpointName must have the form endpointId.handleId, got: %s", endpointName)
	}
	endpointDir := filepath.Join(r.absPath(endpointCfgDir), codegen.CamelToSnake(nameParts[0]))

	// Read endpoint group config
	endpointGroupConfigFile := filepath.Join(endpointDir, endpointConfigFileName)
	groupConfig := &codegen.EndpointClassConfig{}
	if err = readJSONFile(endpointGroupConfigFile, groupConfig); err != nil {
		return errors.Wrapf(err, "failed to read %s for %s", endpointConfigFileName, endpointName)
	}

	// Read the existing individual endpoint config json to be deleted
	endpointIndexInGroup := r.findEndpointInGroup(endpointDir, nameParts[1], groupConfig.Config.Endpoints)
	if endpointIndexInGroup == -1 {
		return errors.Errorf("cannot find individual config json for %s", endpointName)
	}
	endpointConfigFile := filepath.Join(endpointDir, groupConfig.Config.Endpoints[endpointIndexInGroup])
	endpointConfig := &EndpointConfig{}
	if err = readJSONFile(endpointConfigFile, endpointConfig); err != nil {
		return errors.Wrap(err, "failed to read existing endpoint config")
	}

	// Need this to update the endpoint group client dependency list
	clientDep, err := r.GetAllClientDependencies()
	if err != nil {
		return err
	}

	// Remove the individual endpoint config json
	if err = os.Remove(endpointConfigFile); err != nil {
		return errors.Wrap(err, "failed to delete individual endpoint config json")
	}

	if len(groupConfig.Config.Endpoints) == 1 {
		// this is the last endpoint remaining in this group, delete the group and update service config
		if err = os.RemoveAll(endpointDir); err != nil {
			return errors.Wrapf(err, "failed to remove endpoint group config directory for %s", endpointConfig.ID)
		}

		serviceConfigFile := filepath.Join(r.absPath(endpointCfgDir), "..", "/services/", gatewayCfg.ID, serviceConfigFileName)
		serviceConfig := &codegen.EndpointClassConfig{}
		if err = readJSONFile(serviceConfigFile, serviceConfig); err != nil {
			return errors.Wrap(err, "failed to read service config json")
		}
		if i := findString(endpointConfig.ID, serviceConfig.Dependencies["endpoint"]); i != -1 {
			serviceConfig.Dependencies["endpoint"] = append(serviceConfig.Dependencies["endpoint"][:i], serviceConfig.Dependencies["endpoint"][i+1:]...)
		}
		if err = writeToJSONFile(serviceConfigFile, serviceConfig); err != nil {
			return errors.Wrap(err, "failed to write updated service config json")
		}

	} else {
		// there are other endpoints in this group, need to update group config

		// Remove this endpoint from the group list
		groupConfig.Config.Endpoints = append(groupConfig.Config.Endpoints[:endpointIndexInGroup], groupConfig.Config.Endpoints[endpointIndexInGroup+1:]...)

		// Remove client dependency from group config if no other endpoints in this group use it
		otherEndpoints, _ := clientDep[endpointConfig.ClientID]
		if len(filterStringsByPrefix(endpointConfig.ID+".", otherEndpoints)) == 1 {
			if i := findString(endpointConfig.ClientID, groupConfig.Dependencies["client"]); i != -1 {
				groupConfig.Dependencies["client"] = append(groupConfig.Dependencies["client"][:i], groupConfig.Dependencies["client"][i+1:]...)
			}
		}
		// Remove middleware dependency from group config if no other endpoints in this group use it
		for _, m := range endpointConfig.Middlewares {
			otherEndpoints = getEndpointsWithMiddleware(gatewayCfg, m.Name, endpointConfig.ID)
			if len(otherEndpoints) == 1 {
				if i := findString(m.Name, groupConfig.Dependencies["middleware"]); i != -1 {
					groupConfig.Dependencies["middleware"] = append(groupConfig.Dependencies["middleware"][:i], groupConfig.Dependencies["middleware"][i+1:]...)
				}
			}
		}

		if err = writeToJSONFile(endpointGroupConfigFile, groupConfig); err != nil {
			return errors.Wrapf(err, "failed to write updated endpoint group config for %s", endpointConfig.ID)
		}
	}

	return nil
}

func getEndpointsWithMiddleware(gatewayCfg *Config, middleware string, endpointGroup string) []string {
	endpoints := make([]string, 0)
	for _, endpoint := range gatewayCfg.Endpoints {
		if endpoint.ID == endpointGroup {
			found := false
			for _, m := range endpoint.Middlewares {
				if m.Name == middleware {
					found = true
					break
				}
			}
			if found {
				endpoints = append(endpoints, endpoint.HandleID)
			}
		}
	}
	return endpoints
}

func (r *Repository) validateEndpointCfg(req *EndpointConfig) error {
	gatewayConfig, err := r.LatestGatewayConfig()
	if err != nil {
		return errors.Wrap(err, "invalid configuration before updating endpoint")
	}
	clientCfg, ok := gatewayConfig.Clients[req.ClientID]
	if !ok {
		return reqerr.NewRequestError(
			reqerr.EndpointsClientID, errors.Errorf("can't find client %q", req.ClientID))
	}
	if clientCfg.Type == HTTP {
		req.WorkflowType = "httpClient"
	} else if clientCfg.Type == TCHANNEL {
		req.WorkflowType = "tchannelClient"
	} else {
		return reqerr.NewRequestError(reqerr.ClientsType,
			errors.Errorf("client type %q is not supported", clientCfg.Type))
	}

	for _, mid := range req.Middlewares {
		for _, midCfg := range gatewayConfig.RawMiddlewares {
			if mid.Name == midCfg.Name {
				absFile := "file://" + r.absPath(midCfg.OptionsSchemaFile)
				err := codegen.SchemaValidateGo(absFile, mid.Options)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *Repository) findEndpointInGroup(endpointDir string, handleID string, jsonList []string) int {
	expectedJSONName := codegen.CamelToSnake(handleID) + ".json"
	expectedPath := filepath.Clean(filepath.Join(endpointDir, expectedJSONName))

	// Try to find right json file in list (should succeed if naming convention is followed)
	for i := range jsonList {
		if filepath.Clean(filepath.Join(endpointDir, jsonList[i])) == expectedPath {
			return i
		}
	}

	// Try to find right json by examining each one
	for i := range jsonList {
		config := &EndpointConfig{}
		if err := readJSONFile(filepath.Join(endpointDir, jsonList[i]), config); err != nil {
			return -1
		}
		if config.HandleID == handleID {
			return i
		}
	}

	return -1
}

// updateServiceMetaJSON adds an endpoint group in the meta json file or updates the config for an existing endpoint.
func updateServiceMetaJSON(configDir, serviceConfigJSONPath string, cfg *EndpointConfig) error {
	metaFilePath := filepath.Join(configDir, serviceConfigJSONPath)
	fileContent := new(codegen.EndpointClassConfig)
	endpoints := []string{}
	if _, err := os.Stat(metaFilePath); !os.IsNotExist(err) {
		err := readJSONFile(metaFilePath, fileContent)
		if err != nil {
			return err
		}
	}
	if fileContent.Dependencies != nil {
		endpoints = fileContent.Dependencies["endpoint"]
	} else {
		fileContent.Dependencies = make(map[string][]string)
	}
	sort.Strings(endpoints)
	i := sort.SearchStrings(endpoints, cfg.ID)
	// not update if client id already exist
	if i < len(endpoints) && endpoints[i] == cfg.ID {
		return nil
	}
	// update endpoint list with the new client id
	fileContent.Dependencies["endpoint"] = append(fileContent.Dependencies["endpoint"], cfg.ID)
	sort.Strings(fileContent.Dependencies["endpoint"])
	return writeToJSONFile(metaFilePath, fileContent)
}

// updateEndpointMetaJSON adds an endpoint in the meta json file or updates the config for an existing endpoint.
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
	if c := fileContent.Dependencies["client"]; findString(cfg.ClientID, c) == -1 {
		fileContent.Dependencies["client"] = append(c, cfg.ClientID)
	}

	for _, m := range cfg.Middlewares {
		if middlewares := fileContent.Dependencies["middleware"]; findString(m.Name, middlewares) == -1 {
			fileContent.Dependencies["middleware"] = append(middlewares, m.Name)
		}
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
	if findString(newFilePath, oldFilePaths) == -1 {
		curEndpoints = append(curEndpoints, newFile)
	}
	return curEndpoints, nil
}

func findString(target string, array []string) int {
	for i, str := range array {
		if str == target {
			return i
		}
	}
	return -1
}

func filterStringsByPrefix(prefix string, array []string) []string {
	filtered := make([]string, 0, len(array))
	for _, str := range array {
		if strings.HasPrefix(str, prefix) {
			filtered = append(filtered, str)
		}
	}
	return filtered
}
