// Copyright (c) 2022 Uber Technologies, Inc.
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

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	validator2 "gopkg.in/validator.v2"
)

type clientConfig interface {
	NewClientSpec(
		instance *ModuleInstance,
		h *PackageHelper) (*ClientSpec, error)
}

// ClientIDLConfig is the "config" field in the client-config.yaml for
// HTTP/TChannel/gRPC clients.
type ClientIDLConfig struct {
	ExposedMethods  map[string]string `yaml:"exposedMethods" json:"exposedMethods" validate:"exposedMethods"`
	IDLFile         string            `yaml:"idlFile" json:"idlFile" validate:"nonzero"`
	IDLFileSha      string            `yaml:"idlFileSha,omitempty" json:"idlFileSha"`
	SidecarRouter   string            `yaml:"sidecarRouter" json:"sidecarRouter"`
	Fixture         *Fixture          `yaml:"fixture,omitempty" json:"fixture"`
	CustomInterface string            `yaml:"customInterface,omitempty" json:"customInterface,omitempty"`
}

// Fixture specifies client fixture import path and all scenarios
type Fixture struct {
	// ImportPath is the package where the user-defined Fixture global variable is contained.
	// The Fixture object defines, for a given client, the standardized list of fixture scenarios for that client
	ImportPath string `yaml:"importPath" json:"importPath" validate:"nonzero"`
	// Scenarios is a map from zanzibar's exposed method name to a list of user-defined fixture scenarios for a client
	Scenarios map[string][]string `yaml:"scenarios" json:"scenarios"`
}

// Validate the fixture configuration
func (f *Fixture) Validate(methods map[string]interface{}) error {
	if f.ImportPath == "" {
		return errors.New("fixture importPath is empty")
	}
	for m := range f.Scenarios {
		if _, ok := methods[m]; !ok {
			return errors.Errorf("%q is not a valid method", m)
		}
	}
	return nil
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

// HTTPClientConfig represents the "config" field for a HTTP client-config.yaml
type HTTPClientConfig struct {
	ClassConfigBase `yaml:",inline" json:",inline"`
	Dependencies    Dependencies     `yaml:"dependencies,omitempty" json:"dependencies"`
	Config          *ClientIDLConfig `yaml:"config" json:"config" validate:"nonzero"`
}

func newHTTPClientConfig(raw []byte, validator *validator2.Validator) (*HTTPClientConfig, error) {
	config := &HTTPClientConfig{}
	if errUnmarshal := yaml.Unmarshal(raw, config); errUnmarshal != nil {
		return nil, errors.Wrap(
			errUnmarshal, "Could not parse HTTP client config data")
	}

	if errValidate := validator.Validate(config); errValidate != nil {
		return nil, errors.Wrap(
			errValidate, "http client config validation failed")
	}

	return config, nil
}

func newClientSpec(
	clientType string,
	config *ClientIDLConfig,
	instance *ModuleInstance,
	h *PackageHelper,
	annotate bool,
) (*ClientSpec, error) {
	thriftFile := filepath.Join(h.IdlPath(), h.GetModuleIdlSubDir(false), config.IDLFile)
	mspec, err := NewModuleSpec(thriftFile, annotate, false, h)

	if err != nil {
		return nil, err
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
		CustomInterface:    config.CustomInterface,
	}

	return cspec, nil
}

// NewClientSpec creates a client spec from a client module instance
func (c *HTTPClientConfig) NewClientSpec(
	instance *ModuleInstance,
	h *PackageHelper) (*ClientSpec, error) {
	return newClientSpec(c.Type, c.Config, instance, h, true)
}

// TChannelClientConfig represents the "config" field for a TChannel client-config.yaml
type TChannelClientConfig struct {
	ClassConfigBase `yaml:",inline" json:",inline"`
	Dependencies    Dependencies     `yaml:"dependencies,omitempty" json:"dependencies"`
	Config          *ClientIDLConfig `yaml:"config" json:"config" validate:"nonzero"`
}

func newTChannelClientConfig(raw []byte, validator *validator2.Validator) (*TChannelClientConfig, error) {
	config := &TChannelClientConfig{}
	if errUnmarshal := yaml.Unmarshal(raw, config); errUnmarshal != nil {
		return nil, errors.Wrap(
			errUnmarshal, "Could not parse TChannel client config data")
	}

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

// CustomClientConfig represents the config for a custom client.
type CustomClientConfig struct {
	ClassConfigBase `yaml:",inline" json:",inline"`
	Dependencies    Dependencies `yaml:"dependencies" json:"dependencies"`
	Config          *struct {
		Fixture          *Fixture `yaml:"fixture" json:"fixture"`
		CustomImportPath string   `yaml:"customImportPath" json:"customImportPath" validate:"nonzero"`
		CustomInterface  string   `yaml:"customInterface,omitempty" json:"customInterface,omitempty"`
	} `yaml:"config" json:"config" validate:"nonzero"`
}

func newCustomClientConfig(raw []byte, validator *validator2.Validator) (*CustomClientConfig, error) {
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
		CustomInterface:    c.Config.CustomInterface,
	}

	return spec, nil
}

// GRPCClientConfig represents the "config" field for a gRPC client-config.yaml.
type GRPCClientConfig struct {
	ClassConfigBase `yaml:",inline" json:",inline"`
	Dependencies    Dependencies     `yaml:"dependencies,omitempty" json:"dependencies"`
	Config          *ClientIDLConfig `yaml:"config" json:"config" validate:"nonzero"`
}

func newGRPCClientConfig(raw []byte, validator *validator2.Validator) (*GRPCClientConfig, error) {
	config := &GRPCClientConfig{}
	if errUnmarshal := yaml.Unmarshal(raw, config); errUnmarshal != nil {
		return nil, errors.Wrap(
			errUnmarshal, "could not parse gRPC client config data")
	}

	if errValidate := validator.Validate(config); errValidate != nil {
		return nil, errors.Wrap(
			errValidate, "grpc client config validation failed")
	}

	return config, nil
}

func newGRPCClientSpec(
	clientType string,
	config *ClientIDLConfig,
	instance *ModuleInstance,
	h *PackageHelper,
) (*ClientSpec, error) {
	protoFile := filepath.Join(h.IdlPath(), h.GetModuleIdlSubDir(false), config.IDLFile)
	protoSpec, err := NewProtoModuleSpec(protoFile, false, h)
	if err != nil {
		return nil, errors.Wrapf(
			err, "could not build proto spec for proto file %s: ", protoFile,
		)
	}

	cspec := &ClientSpec{
		ModuleSpec:         protoSpec,
		YAMLFile:           instance.YAMLFileName,
		JSONFile:           instance.JSONFileName,
		ClientType:         clientType,
		ImportPackagePath:  instance.PackageInfo.ImportPackagePath(),
		ImportPackageAlias: instance.PackageInfo.ImportPackageAlias(),
		ExportName:         instance.PackageInfo.ExportName,
		ExportType:         instance.PackageInfo.ExportType,
		ThriftFile:         protoFile,
		ClientID:           instance.InstanceName,
		ClientName:         instance.PackageInfo.QualifiedInstanceName,
		ExposedMethods:     config.ExposedMethods,
		SidecarRouter:      config.SidecarRouter,
	}

	return cspec, nil
}

// NewClientSpec creates a client spec from a client module instance
func (c *GRPCClientConfig) NewClientSpec(
	instance *ModuleInstance,
	h *PackageHelper,
) (*ClientSpec, error) {
	return newGRPCClientSpec(c.Type, c.Config, instance, h)
}

func clientType(raw []byte) (string, error) {
	clientConfig := ClassConfigBase{}
	if err := yaml.Unmarshal(raw, &clientConfig); err != nil {
		return "", errors.Wrap(
			err, "Could not parse client config data to determine client type")
	}
	return clientConfig.Type, nil
}

func newClientConfig(raw []byte, validator *validator2.Validator) (clientConfig, error) {
	clientType, errType := clientType(raw)
	if errType != nil {
		return nil, errors.Wrap(
			errType, "Could not determine client type")
	}

	switch clientType {
	case "http":
		return newHTTPClientConfig(raw, validator)
	case "tchannel":
		return newTChannelClientConfig(raw, validator)
	case "grpc":
		return newGRPCClientConfig(raw, validator)
	case "custom":
		return newCustomClientConfig(raw, validator)
	default:
		return nil, errors.Errorf(
			"Unknown client type %q", clientType)
	}
}
