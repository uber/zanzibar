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
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	httpClientYAML = `
name: test
type: http
dependencies:
  client:
    - a
    - b
config:
  idlFileSha: idlFileSha
  idlFile: clients/bar/bar.thrift
  customImportPath: path
  sidecarRouter: sidecar
  fixture:
    importPath: import
    scenarios:
      scenario:
        - s1
        - s2
  exposedMethods:
    a: method
`

	tchannelClientYAML = `
name: test
type: tchannel
dependencies:
  client:
    - a
    - b
config:
  idlFileSha: idlFileSha
  idlFile: clients/bar/bar.thrift
  customImportPath: path
  sidecarRouter: sidecar
  fixture:
    importPath: import
    scenarios:
      scenario:
        - s1
        - s2
  exposedMethods:
    a: method
`
	grpcClientYAML = `
name: test
type: grpc
dependencies:
  client:
    - a
    - b
config:
  idlFileSha: idlFileSha
  idlFile: clients/echo/echo.proto
  customImportPath: path
  fixture:
    importPath: import
    scenarios:
      scenario:
        - s1
        - s2
  exposedMethods:
    a: method
`

	customClientYAML = `
name: test
type: custom
dependencies:
  client:
    - a
    - b
config:
  customImportPath: path
  fixture:
    importPath: import
    scenarios:
      scenario:
        - s1
        - s2
`
)

func TestNewHTTPClientConfigUnmarshalFilure(t *testing.T) {
	invalidYAML := "{{{"
	validator := getExposedMethodValidator()
	_, err := newHTTPClientConfig([]byte(invalidYAML), validator)
	expectedErr := "Could not parse HTTP client config data: error converting YAML to JSON: yaml: line 1: did not find expected node content"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestNewTChannelClientConfigUnmarshalFilure(t *testing.T) {
	invalidYAML := "{{{"
	validator := getExposedMethodValidator()
	_, err := newTChannelClientConfig([]byte(invalidYAML), validator)
	expectedErr := "Could not parse TChannel client config data: error converting YAML to JSON: yaml: line 1: did not find expected node content"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestNewGRPCClientConfigUnmarshalFilure(t *testing.T) {
	invalidYAML := "{{{"
	validator := getExposedMethodValidator()
	_, err := newGRPCClientConfig([]byte(invalidYAML), validator)
	expectedErr := "could not parse gRPC client config data: error converting YAML to JSON: yaml: line 1: did not find expected node content"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestNewCustomClientConfigUnmarshalFilure(t *testing.T) {
	invalidYAML := "{{{"
	validator := getExposedMethodValidator()
	_, err := newCustomClientConfig([]byte(invalidYAML), validator)
	expectedErr := "Could not parse Custom client config data: error converting YAML to JSON: yaml: line 1: did not find expected node content"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func doSubConfigMissingTest(t *testing.T, clientType string) {
	configYAML := fmt.Sprintf(`
name: test
type: %s
# config is missing
`, clientType)

	validator := getExposedMethodValidator()
	_, err := newClientConfig([]byte(configYAML), validator)
	assert.Error(t, err)
	assert.Equal(
		t,
		fmt.Sprintf("%s client config validation failed: Config: zero value", clientType),
		err.Error(),
	)
}

func TestNewClientConfigSubConfigMissing(t *testing.T) {
	doSubConfigMissingTest(t, "http")
	doSubConfigMissingTest(t, "tchannel")
	doSubConfigMissingTest(t, "grpc")
	doSubConfigMissingTest(t, "custom")
}

func doThriftFileMissingTest(t *testing.T, clientType string) {
	configYAML := fmt.Sprintf(`
name: test
type: %s
config:
  idlFileSha: idlFileSha
  # idlFile is missing
`, clientType)
	validator := getExposedMethodValidator()
	_, err := newClientConfig([]byte(configYAML), validator)
	assert.Error(t, err)
	assert.Equal(
		t,
		fmt.Sprintf("%s client config validation failed: Config.IDLFile: zero value", clientType),
		err.Error())
}

func TestThriftFileMissingValidation(t *testing.T) {
	doThriftFileMissingTest(t, "http")
	doThriftFileMissingTest(t, "tchannel")
	doThriftFileMissingTest(t, "grpc")
}

func doThriftFileShaMissingTest(t *testing.T, clientType string) {
	configYAML := fmt.Sprintf(`
name: test
type: %s
config:
  # idlFileSha is missing
  idlFile: idlFile
`, clientType)

	validator := getExposedMethodValidator()
	_, err := newClientConfig([]byte(configYAML), validator)
	assert.NoError(t, err)
}

func TestThriftFileShaMissingValidation(t *testing.T) {
	doThriftFileShaMissingTest(t, "http")
	doThriftFileShaMissingTest(t, "tchannel")
	doThriftFileShaMissingTest(t, "grpc")
}

func TestCustomClientRequiresCustomImportPath(t *testing.T) {
	configYAML := `
name: test
type: custom
config:
  idlFileSha: idlFileSha
  # CustomImportPath is missing
`
	validator := getExposedMethodValidator()
	_, err := newCustomClientConfig([]byte(configYAML), validator)
	expectedErr := "custom client config validation failed: Config.CustomImportPath: zero value"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func doDuplicatedExposedMethodsTest(t *testing.T, clientType string) {
	configYAML := fmt.Sprintf(`
name: test
type: %s
config:
  idlFileSha: idlFileSha
  idlFile: idlFile
  exposedMethods:
    a: method
    b: method
`, clientType)
	validator := getExposedMethodValidator()
	_, err := newClientConfig([]byte(configYAML), validator)
	assert.Error(t, err)
	assert.Equal(
		t,
		fmt.Sprintf("%s client config validation failed: Config.ExposedMethods: value \"method\" of the exposedMethods is not unique", clientType),
		err.Error(),
	)
}

func TestNewClientConfigDuplicatedMethodsFailure(t *testing.T) {
	doDuplicatedExposedMethodsTest(t, "http")
	doDuplicatedExposedMethodsTest(t, "tchannel")
	doDuplicatedExposedMethodsTest(t, "grpc")
}

func TestGetConfigTypeFailure(t *testing.T) {
	clientType, err := clientType([]byte("{{{"))

	expectedErr := "Could not parse client config data to determine client type: error converting YAML to JSON: yaml: line 1: did not find expected node content"
	assert.Equal(t, "", clientType)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestGetConfigTypeSuccess(t *testing.T) {
	configYAML := `
type: client_type
`
	clientType, err := clientType([]byte(configYAML))

	assert.Equal(t, "client_type", clientType)
	assert.NoError(t, err)
}

func TestNewClientConfigTypeError(t *testing.T) {
	configYAML := `
name: test
type: : # malformated type
config:
  idlFileSha: idlFileSha
`

	validator := getExposedMethodValidator()
	_, err := newClientConfig([]byte(configYAML), validator)
	expectedErr := "Could not determine client type: Could not parse client config data to determine client type: error converting YAML to JSON: yaml: line 3: mapping values are not allowed in this context"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestNewClientConfigUnknownClientType(t *testing.T) {
	configYAML := `
name: test
type: unknown
config:
  idlFileSha: idlFileSha
`
	validator := getExposedMethodValidator()
	_, err := newClientConfig([]byte(configYAML), validator)
	expectedErr := "Unknown client type \"unknown\""
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestNewClientConfigGetHTTPClient(t *testing.T) {
	validator := getExposedMethodValidator()
	client, err := newClientConfig([]byte(httpClientYAML), validator)
	expectedClient := HTTPClientConfig{
		ClassConfigBase: ClassConfigBase{
			Name: "test",
			Type: "http",
		},
		Dependencies: Dependencies{
			Client: []string{"a", "b"},
		},
		Config: &ClientIDLConfig{
			ExposedMethods: map[string]string{
				"a": "method",
			},
			IDLFileSha:    "idlFileSha",
			IDLFile:       "clients/bar/bar.thrift",
			SidecarRouter: "sidecar",
			Fixture: &Fixture{
				ImportPath: "import",
				Scenarios: map[string][]string{
					"scenario": {"s1", "s2"},
				},
			},
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, &expectedClient, client)
}

func TestNewClientConfigGetTChannelClient(t *testing.T) {
	validator := getExposedMethodValidator()
	client, err := newClientConfig([]byte(tchannelClientYAML), validator)
	expectedClient := TChannelClientConfig{
		ClassConfigBase: ClassConfigBase{
			Name: "test",
			Type: "tchannel",
		},
		Dependencies: Dependencies{
			Client: []string{"a", "b"},
		},
		Config: &ClientIDLConfig{
			ExposedMethods: map[string]string{
				"a": "method",
			},
			IDLFileSha:    "idlFileSha",
			IDLFile:       "clients/bar/bar.thrift",
			SidecarRouter: "sidecar",
			Fixture: &Fixture{
				ImportPath: "import",
				Scenarios: map[string][]string{
					"scenario": {"s1", "s2"},
				},
			},
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, &expectedClient, client)
}

func TestNewClientConfigGetGRPCClient(t *testing.T) {
	validator := getExposedMethodValidator()
	client, err := newClientConfig([]byte(grpcClientYAML), validator)
	expectedClient := GRPCClientConfig{
		ClassConfigBase: ClassConfigBase{
			Name: "test",
			Type: "grpc",
		},
		Dependencies: Dependencies{
			Client: []string{"a", "b"},
		},
		Config: &ClientIDLConfig{
			ExposedMethods: map[string]string{
				"a": "method",
			},
			IDLFileSha: "idlFileSha",
			IDLFile:    "clients/echo/echo.proto",
			Fixture: &Fixture{
				ImportPath: "import",
				Scenarios: map[string][]string{
					"scenario": {"s1", "s2"},
				},
			},
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, &expectedClient, client)
}

func TestNewClientConfigGetCustomClient(t *testing.T) {
	validator := getExposedMethodValidator()
	client, err := newClientConfig([]byte(customClientYAML), validator)
	expectedClient := CustomClientConfig{
		ClassConfigBase: ClassConfigBase{
			Name: "test",
			Type: "custom",
		},
		Dependencies: Dependencies{
			Client: []string{"a", "b"},
		},
		Config: &struct {
			Fixture          *Fixture `yaml:"fixture" json:"fixture"`
			CustomImportPath string   `yaml:"customImportPath" json:"customImportPath" validate:"nonzero"`
			CustomInterface  string   `yaml:"customInterface,omitempty" json:"customInterface,omitempty"`
		}{
			CustomImportPath: "path",
			Fixture: &Fixture{
				ImportPath: "import",
				Scenarios: map[string][]string{
					"scenario": {"s1", "s2"},
				},
			},
			CustomInterface: "",
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, &expectedClient, client)
}

func newTestPackageHelper(t *testing.T) *PackageHelper {
	relativeGatewayPath := "../examples/example-gateway"
	absGatewayPath, err := filepath.Abs(relativeGatewayPath)
	assert.NoError(t, err, "failed to get abs path %s", relativeGatewayPath)

	packageRoot := "github.com/uber/zanzibar/examples/example-gateway"
	options := &PackageHelperOptions{
		RelTargetGenDir: "tmpDir",
		CopyrightHeader: "copyright",
		GenCodePackage: map[string]string{
			".thrift": packageRoot + "/build/gen-code",
			".proto":  packageRoot + "/build/gen-code",
		},
		TraceKey:        "trace-key",
		ModuleIdlSubDir: map[string]string{"endpoints": "endpoints-idl", "default": "clients-idl"},
	}

	h, err := NewPackageHelper(
		packageRoot,
		absGatewayPath,
		options,
	)
	assert.NoError(t, err, "failed to create package helper")
	return h

}

func TestClientNewClientSpecFailedWithThriftCompilation(t *testing.T) {
	testNewClientSpecFailedWithCompilation(t, "http")
	testNewClientSpecFailedWithCompilation(t, "grpc")
}

func testNewClientSpecFailedWithCompilation(t *testing.T, clientType string) {
	configYAML := fmt.Sprintf(`
name: test
type: %s
config:
  idlFileSha: idlFileSha
  idlFile: NOT_EXIST
  exposedMethods:
    a: method
`, clientType)
	validator := getExposedMethodValidator()
	client, errClient := newClientConfig([]byte(configYAML), validator)
	assert.NoError(t, errClient)

	h := newTestPackageHelper(t)
	_, errSpec := client.NewClientSpec(nil /* ModuleInstance */, h)
	fmt.Println(errSpec)
	assert.Error(t, errSpec)
}

// only for http and tchannel clients
func doNewClientSpecTest(t *testing.T, rawConfig []byte, clientType string) {
	validator := getExposedMethodValidator()
	client, errClient := newClientConfig(rawConfig, validator)
	assert.NoError(t, errClient)
	instance := &ModuleInstance{
		YAMLFileName: "YAMLFileName",
		JSONFileName: "JSONFileName",
		InstanceName: "InstanceName",
		PackageInfo: &PackageInfo{
			ExportName:            "ExportName",
			ExportType:            "ExportType",
			QualifiedInstanceName: "QualifiedInstanceName",
		},
	}
	h := newTestPackageHelper(t)

	idlFile := filepath.Join(h.IdlPath(), h.GetModuleIdlSubDir(false), "clients/bar/bar.thrift")
	expectedSpec := &ClientSpec{
		ModuleSpec:         nil,
		YAMLFile:           instance.YAMLFileName,
		JSONFile:           instance.JSONFileName,
		ClientType:         clientType,
		ImportPackagePath:  instance.PackageInfo.ImportPackagePath(),
		ImportPackageAlias: instance.PackageInfo.ImportPackageAlias(),
		ExportName:         instance.PackageInfo.ExportName,
		ExportType:         instance.PackageInfo.ExportType,
		ThriftFile:         idlFile,
		ClientID:           instance.InstanceName,
		ClientName:         instance.PackageInfo.QualifiedInstanceName,
		ExposedMethods: map[string]string{
			"a": "method",
		},
		SidecarRouter: "sidecar",
	}

	spec, errSpec := client.NewClientSpec(instance, h)
	spec.ModuleSpec = nil // Not interested in ModuleSpec here
	assert.NoError(t, errSpec)
	assert.Equal(t, expectedSpec, spec)
}

func TestHTTPClientNewClientSpec(t *testing.T) {
	doNewClientSpecTest(t, []byte(httpClientYAML), "http")
}

func TestTChannelClientNewClientSpec(t *testing.T) {
	doNewClientSpecTest(t, []byte(tchannelClientYAML), "tchannel")
}

func TestGRPCClientNewClientSpec(t *testing.T) {
	validator := getExposedMethodValidator()
	client, errClient := newClientConfig([]byte(grpcClientYAML), validator)
	assert.NoError(t, errClient)
	instance := &ModuleInstance{
		YAMLFileName: "YAMLFileName",
		JSONFileName: "JSONFileName",
		InstanceName: "InstanceName",
		PackageInfo: &PackageInfo{
			ExportName:            "ExportName",
			ExportType:            "ExportType",
			QualifiedInstanceName: "QualifiedInstanceName",
		},
	}
	h := newTestPackageHelper(t)

	idlFile := filepath.Join(h.IdlPath(), h.GetModuleIdlSubDir(false), "clients/echo/echo.proto")
	expectedSpec := &ClientSpec{
		ModuleSpec:         nil,
		YAMLFile:           instance.YAMLFileName,
		JSONFile:           instance.JSONFileName,
		ClientType:         "grpc",
		ImportPackagePath:  instance.PackageInfo.ImportPackagePath(),
		ImportPackageAlias: instance.PackageInfo.ImportPackageAlias(),
		ExportName:         instance.PackageInfo.ExportName,
		ExportType:         instance.PackageInfo.ExportType,
		ThriftFile:         idlFile,
		ClientID:           instance.InstanceName,
		ClientName:         instance.PackageInfo.QualifiedInstanceName,
		ExposedMethods: map[string]string{
			"a": "method",
		},
	}

	spec, errSpec := client.NewClientSpec(instance, h)
	spec.ModuleSpec = nil // Not interested in ModuleSpec here
	assert.NoError(t, errSpec)
	assert.Equal(t, expectedSpec, spec)
}

func TestCustomClientNewClientSpec(t *testing.T) {
	validator := getExposedMethodValidator()
	client, errClient := newClientConfig([]byte(customClientYAML), validator)
	assert.NoError(t, errClient)
	instance := &ModuleInstance{
		YAMLFileName: "YAMLFileName",
		JSONFileName: "JSONFileName",
		InstanceName: "InstanceName",
		PackageInfo: &PackageInfo{
			ExportName:            "ExportName",
			ExportType:            "ExportType",
			QualifiedInstanceName: "QualifiedInstanceName",
		},
	}

	expectedSpec := &ClientSpec{
		YAMLFile:           instance.YAMLFileName,
		JSONFile:           instance.JSONFileName,
		ImportPackagePath:  instance.PackageInfo.ImportPackagePath(),
		ImportPackageAlias: instance.PackageInfo.ImportPackageAlias(),
		ExportName:         instance.PackageInfo.ExportName,
		ExportType:         instance.PackageInfo.ExportType,
		ClientType:         "custom",
		ClientID:           "test",
		ClientName:         instance.PackageInfo.QualifiedInstanceName,
		CustomImportPath:   "path",
	}

	spec, errSpec := client.NewClientSpec(instance, nil)
	assert.NoError(t, errSpec)
	assert.Equal(t, expectedSpec, spec)
}

func TestConfigNoFixture(t *testing.T) {
	configYAML := `
name: test
type: testable
config:
  customImportPath: path
`
	client, err := newMockableClient([]byte(configYAML))
	assert.NoError(t, err)
	expectedClient := &mockableClient{
		ClassConfigBase: ClassConfigBase{
			Name: "test",
			Type: "testable",
		},
		Config: &struct {
			Fixture          *Fixture `yaml:"fixture" json:"fixture"`
			CustomImportPath string   `yaml:"customImportPath" json:"customImportPath"`
			CustomInterface  string   `yaml:"customInterface,omitempty" json:"customInterface,omitempty"`
		}{
			CustomImportPath: "path",
		},
	}
	assert.Equal(t, expectedClient, client)
}

func TestConfigCustomInterface(t *testing.T) {
	configYAML := `
name: test
type: testable
config:
  customInterface: name
`
	client, err := newMockableClient([]byte(configYAML))
	assert.NoError(t, err)
	expectedClient := &mockableClient{
		ClassConfigBase: ClassConfigBase{
			Name: "test",
			Type: "testable",
		},
		Config: &struct {
			Fixture          *Fixture `yaml:"fixture" json:"fixture"`
			CustomImportPath string   `yaml:"customImportPath" json:"customImportPath"`
			CustomInterface  string   `yaml:"customInterface,omitempty" json:"customInterface,omitempty"`
		}{
			CustomInterface: "name",
		},
	}
	assert.Equal(t, expectedClient, client)
}

func TestClientFixtureImportPathMissing(t *testing.T) {
	configYAML := `
name: test
type: testable
config:
  customImportPath: path
  fixture:
    # ImportPath is missing
    scenarios:
      scenario:
        - s1
`
	_, err := newMockableClient([]byte(configYAML))
	expectedErr := "testable client config validation failed: Config.Fixture.ImportPath: zero value"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}
