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

package codegen

import (
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/validator.v2"
	yaml "gopkg.in/yaml.v2"
)

// ClientConfigBase is the struct for client-config.yaml. It is the base for
// specific type of client.
type ClientConfigBase struct {
	Name         string       `yaml:"name" json:"name"`
	Type         string       `yaml:"type" json:"type"`
	Dependencies Dependencies `yaml:"dependencies" json:"dependencies"`
}

type clientConfig interface {
	NewClientSpec(
		instance *ModuleInstance,
		h *PackageHelper) (*ClientSpec, error)
	customImportPath() string
	fixture() *Fixture
}

// ClientSubConfig is the "config" field in the client-config.yaml for http
// client and tchannel client.
type ClientSubConfig struct {
	ExposedMethods   map[string]string `yaml:"exposedMethods" json:"exposedMethods" validate:"exposedMethods"`
	CustomImportPath string            `yaml:"customImportPath" json:"customImportPath"`
	ThriftFile       string            `yaml:"thriftFile" json:"thriftFile" validate:"nonzero"`
	ThriftFileSha    string            `yaml:"thriftFileSha" json:"thriftFileSha" validate:"nonzero"`
	SidecarRouter    string            `yaml:"sidecarRouter" json:"sidecarRouter"`
	Fixture          *Fixture          `yaml:"fixture" json:"fixture"`
}

func validateExposedMethods(v interface{}, param string) error {
	methods := v.(map[string]string)

	// Check duplication
	visited := make(map[string]string, len(methods))
	for key, val := range methods {
		if _, ok := visited[val]; ok {
			return errors.Errorf(
				"value %q of the exposedMethods is not unique", val,
			)
		}
		visited[val] = key
	}
	return nil
}

// HTTPClientConfig represents the config for a HTTP client.
type HTTPClientConfig struct {
	ClientConfigBase `yaml:",inline" json:",inline"`
	Config           *ClientSubConfig `yaml:"config" json:"config" validate:"nonzero"`
}

func newHTTPClientConfig(raw []byte) (*HTTPClientConfig, error) {
	config := &HTTPClientConfig{}
	if errUnmarshal := yaml.Unmarshal(raw, config); errUnmarshal != nil {
		return nil, errors.Wrap(
			errUnmarshal, "Could not parse HTTP client config data")
	}

	validator.SetValidationFunc("exposedMethods", validateExposedMethods)
	if errValidate := validator.Validate(config); errValidate != nil {
		return nil, errors.Wrap(
			errValidate, "http client config validation failed")
	}

	return config, nil
}

func newClientSpec(
	clientType string,
	config *ClientSubConfig,
	instance *ModuleInstance,
	h *PackageHelper,
	annotate bool,
) (*ClientSpec, error) {
	thriftFile := filepath.Join(h.ThriftIDLPath(), config.ThriftFile)
	mspec, err := NewModuleSpec(thriftFile, annotate, false, h)

	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not build module spec for thrift %s: ", thriftFile,
		)
	}
	mspec.PackageName = mspec.PackageName + "client"

	cspec := &ClientSpec{
		ModuleSpec:         mspec,
		YAMLFile:           instance.YAMLFileName,
		JSONFile:           instance.JSONFileName,
		ClientType:         clientType,
		ImportPackagePath:  instance.PackageInfo.ImportPackagePath(),
		ImportPackageAlias: instance.PackageInfo.ImportPackageAlias(),
		ExportName:         instance.PackageInfo.ExportName,
		ExportType:         instance.PackageInfo.ExportType,
		ThriftFile:         thriftFile,
		ClientID:           instance.InstanceName,
		ClientName:         instance.PackageInfo.QualifiedInstanceName,
		ExposedMethods:     config.ExposedMethods,
		SidecarRouter:      config.SidecarRouter,
	}

	return cspec, nil
}

// NewClientSpec creates a client spec from a client module instance
func (c *HTTPClientConfig) NewClientSpec(
	instance *ModuleInstance,
	h *PackageHelper) (*ClientSpec, error) {
	return newClientSpec(c.Type, c.Config, instance, h, true)
}

func (c *HTTPClientConfig) customImportPath() string {
	return c.Config.CustomImportPath
}

func (c *HTTPClientConfig) fixture() *Fixture {
	return c.Config.Fixture
}

// TChannelClientConfig represents the config for a TChannel client.
type TChannelClientConfig struct {
	ClientConfigBase `yaml:",inline" json:",inline"`
	Config           *ClientSubConfig `yaml:"config" json:"config" validate:"nonzero"`
}

func newTChannelClientConfig(raw []byte) (*TChannelClientConfig, error) {
	config := &TChannelClientConfig{}
	if errUnmarshal := yaml.Unmarshal(raw, config); errUnmarshal != nil {
		return nil, errors.Wrap(
			errUnmarshal, "Could not parse TChannel client config data")
	}

	validator.SetValidationFunc("exposedMethods", validateExposedMethods)
	if errValidate := validator.Validate(config); errValidate != nil {
		return nil, errors.Wrap(
			errValidate, "tchannel client config validation failed")
	}

	return config, nil
}

// NewClientSpec creates a client spec from a client module instance
func (c *TChannelClientConfig) NewClientSpec(
	instance *ModuleInstance,
	h *PackageHelper) (*ClientSpec, error) {
	return newClientSpec(c.Type, c.Config, instance, h, false)
}

func (c *TChannelClientConfig) customImportPath() string {
	return c.Config.CustomImportPath
}

func (c *TChannelClientConfig) fixture() *Fixture {
	return c.Config.Fixture
}

// CustomClientSubConfig is for custom client
type CustomClientSubConfig struct {
	CustomImportPath string   `yaml:"customImportPath" json:"customImportPath" validate:"nonzero"`
	Fixture          *Fixture `yaml:"fixture" json:"fixture"`
}

// CustomClientConfig represents the config for a custom client.
type CustomClientConfig struct {
	ClientConfigBase `yaml:",inline" json:",inline"`
	Config           *CustomClientSubConfig `yaml:"config" json:"config" validate:"nonzero"`
}

func newCustomClientConfig(raw []byte) (*CustomClientConfig, error) {
	config := &CustomClientConfig{}
	if errUnmarshal := yaml.Unmarshal(raw, config); errUnmarshal != nil {
		return nil, errors.Wrap(
			errUnmarshal, "Could not parse Custom client config data")
	}

	if errValidate := validator.Validate(config); errValidate != nil {
		return nil, errors.Wrap(
			errValidate, "custom client config validation failed")
	}

	return config, nil
}

// NewClientSpec creates a client spec from a http client module instance
func (c *CustomClientConfig) NewClientSpec(
	instance *ModuleInstance,
	h *PackageHelper) (*ClientSpec, error) {

	spec := &ClientSpec{
		YAMLFile:           instance.YAMLFileName,
		JSONFile:           instance.JSONFileName,
		ImportPackagePath:  instance.PackageInfo.ImportPackagePath(),
		ImportPackageAlias: instance.PackageInfo.ImportPackageAlias(),
		ExportName:         instance.PackageInfo.ExportName,
		ExportType:         instance.PackageInfo.ExportType,
		ClientType:         c.Type,
		ClientID:           c.Name,
		ClientName:         instance.PackageInfo.QualifiedInstanceName,
		CustomImportPath:   c.Config.CustomImportPath,
	}

	return spec, nil
}

func (c *CustomClientConfig) customImportPath() string {
	return c.Config.CustomImportPath
}

func (c *CustomClientConfig) fixture() *Fixture {
	return c.Config.Fixture
}

func clientType(raw []byte) (string, error) {
	clientConfig := ClientConfigBase{}
	if err := yaml.Unmarshal(raw, &clientConfig); err != nil {
		return "", errors.Wrap(
			err, "Could not parse client config data to determine client type")
	}
	return clientConfig.Type, nil
}

func newClientConfig(raw []byte) (clientConfig, error) {
	clientType, errType := clientType(raw)
	if errType != nil {
		return nil, errors.Wrap(
			errType, "Could not determine client type")
	}

	switch clientType {
	case "http":
		return newHTTPClientConfig(raw)
	case "tchannel":
		return newTChannelClientConfig(raw)
	case "custom":
		return newCustomClientConfig(raw)
	default:
		return nil, errors.Errorf(
			"Unknown client type %q", clientType)
	}
}
