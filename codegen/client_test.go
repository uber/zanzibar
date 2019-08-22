// Copyright (c) 2019 Uber Technologies, Inc.
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
  thriftFileSha: thriftFileSha
  thriftFile: clients/bar/bar.thrift
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
  thriftFileSha: thriftFileSha
  thriftFile: clients/bar/bar.thrift
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
	_, err := newHTTPClientConfig([]byte(invalidYAML))
	expectedErr := "Could not parse HTTP client config data: yaml: line 1: did not find expected node content"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestNewTChannelClientConfigUnmarshalFilure(t *testing.T) {
	invalidYAML := "{{{"
	_, err := newTChannelClientConfig([]byte(invalidYAML))
	expectedErr := "Could not parse TChannel client config data: yaml: line 1: did not find expected node content"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestNewCustomClientConfigUnmarshalFilure(t *testing.T) {
	invalidYAML := "{{{"
	_, err := newCustomClientConfig([]byte(invalidYAML))
	expectedErr := "Could not parse Custom client config data: yaml: line 1: did not find expected node content"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func doSubConfigMissingTest(t *testing.T, clientType string) {
	configYAML := fmt.Sprintf(`
name: test
type: %s
# config is missing
`, clientType)

	_, err := newClientConfig([]byte(configYAML))
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
	doSubConfigMissingTest(t, "custom")
}

func doThriftFileMissingTest(t *testing.T, clientType string) {
	configYAML := fmt.Sprintf(`
name: test
type: %s
config:
  thriftFileSha: thriftFileSha
  # thriftFile is missing
`, clientType)

	_, err := newClientConfig([]byte(configYAML))
	assert.Error(t, err)
	assert.Equal(
		t,
		fmt.Sprintf("%s client config validation failed: Config.IDLFile: zero value", clientType),
		err.Error())
}

func TestThriftFileMissingValidation(t *testing.T) {
	doThriftFileMissingTest(t, "http")
	doThriftFileMissingTest(t, "tchannel")
}

func doThriftFileShaMissingTest(t *testing.T, clientType string) {
	configYAML := fmt.Sprintf(`
name: test
type: %s
config:
  # thriftFileSha is missing
  thriftFile: thriftFile
`, clientType)

	_, err := newClientConfig([]byte(configYAML))
	assert.NoError(t, err)
}

func TestThriftFileShaMissingValidation(t *testing.T) {
	doThriftFileShaMissingTest(t, "http")
	doThriftFileShaMissingTest(t, "tchannel")
}

func TestCustomClientRequiresCustomImportPath(t *testing.T) {
	configYAML := `
name: test
type: custom
config:
  thriftFileSha: thriftFileSha
  # CustomImportPath is missing
`
	_, err := newCustomClientConfig([]byte(configYAML))
	expectedErr := "custom client config validation failed: Config.CustomImportPath: zero value"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func doDuplicatedExposedMethodsTest(t *testing.T, clientType string) {
	configYAML := fmt.Sprintf(`
name: test
type: %s
config:
  thriftFileSha: thriftFileSha
  thriftFile: thriftFile
  exposedMethods:
    a: method
    b: method
`, clientType)

	_, err := newClientConfig([]byte(configYAML))
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
}

func TestGetConfigTypeFailure(t *testing.T) {
	clientType, err := clientType([]byte("{{{"))

	expectedErr := "Could not parse client config data to determine client type: yaml: line 1: did not find expected node content"
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
  thriftFileSha: thriftFileSha
`

	_, err := newClientConfig([]byte(configYAML))
	expectedErr := "Could not determine client type: Could not parse client config data to determine client type: yaml: line 3: mapping values are not allowed in this context"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestNewClientConfigUnknownClientType(t *testing.T) {
	configYAML := `
name: test
type: unknown
config:
  thriftFileSha: thriftFileSha
`
	_, err := newClientConfig([]byte(configYAML))
	expectedErr := "Unknown client type \"unknown\""
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestNewClientConfigGetHTTPClient(t *testing.T) {
	client, err := newClientConfig([]byte(httpClientYAML))
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
			IDLFileSha:    "thriftFileSha",
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
	client, err := newClientConfig([]byte(tchannelClientYAML))
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
			IDLFileSha:    "thriftFileSha",
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

func TestNewClientConfigGetCustomClient(t *testing.T) {
	client, err := newClientConfig([]byte(customClientYAML))
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
		}{
			CustomImportPath: "path",
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

func newTestPackageHelper(t *testing.T) *PackageHelper {
	relativeGatewayPath := "../examples/example-gateway"
	absGatewayPath, err := filepath.Abs(relativeGatewayPath)
	assert.NoError(t, err, "failed to get abs path %s", relativeGatewayPath)

	packageRoot := "github.com/uber/zanzibar/examples/example-gateway"
	options := &PackageHelperOptions{
		RelTargetGenDir: "tmpDir",
		CopyrightHeader: "copyright",
		GenCodePackage:  packageRoot + "/build/gen-code",
		TraceKey:        "trace-key",
	}

	h, err := NewPackageHelper(
		packageRoot,
		absGatewayPath,
		options,
	)
	assert.NoError(t, err, "failed to create package helper")
	return h

}

func TestHTTPClientNewClientSpecFailedWithThriftCompilation(t *testing.T) {
	configYAML := `
name: test
type: http
config:
  thriftFileSha: thriftFileSha
  thriftFile: NOT_EXIST
  exposedMethods:
    a: method
`
	client, errClient := newClientConfig([]byte(configYAML))
	assert.NoError(t, errClient)

	h := newTestPackageHelper(t)
	_, errSpec := client.NewClientSpec(nil /* ModuleInstance */, h)
	assert.Error(t, errSpec)
}

// only for http and tchannel clients
func doNewClientSpecTest(t *testing.T, rawConfig []byte, clientType string) {
	client, errClient := newClientConfig(rawConfig)
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

	thriftFile := filepath.Join(h.ThriftIDLPath(), "clients/bar/bar.thrift")
	expectedSpec := &ClientSpec{
		ModuleSpec:         nil,
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

func TestCustomClientNewClientSpec(t *testing.T) {
	client, errClient := newClientConfig([]byte(customClientYAML))
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
		}{
			CustomImportPath: "path",
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
