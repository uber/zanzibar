package repository

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
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
func NewClientConfigJSON(cfg ClientConfig) *ClientJSONConfig {
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
	cfgJSON := NewClientConfigJSON(*req)
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

func validateClientUpdateRequest(req *ClientConfig) error {
	if len(req.ExposedMethods) == 0 {
		return errors.New("invalid request: no method is exposed for the client")
	}
	if req.Type == "tchannel" && req.MuttleyName == "" {
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

type clientModuleJSONConfig struct {
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Config       interface{}  `json:"config"`
	Dependencies dependencies `json:"dependencies"`
}

type dependencies struct {
	Client []string `json:"client"`
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
	content := &clientModuleJSONConfig{
		Name:   "clients",
		Type:   "init",
		Config: struct{}{},
		Dependencies: dependencies{
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
		content[prefix+"serviceName"] = req.MuttleyName
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
