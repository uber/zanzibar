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
	"testing"

	"github.com/stretchr/testify/assert"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
	yaml "gopkg.in/yaml.v2"
)

var defaultTestOptions = &testGateway.Options{
	KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
	KnownTChannelBackends: []string{"baz"},
	ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
}
var defaultTestConfig = map[string]interface{}{
	"clients.baz.serviceName": "baz",
}

func TestHeadersPropagateMultiDest(t *testing.T) {
	cfg := `{
		"name": "headersPropagate",
		"options": {
			"propagate": [
				{
					"from":"x-uuid",
					"to":"request.dest"
				},
				{
					"from":"x-token",
					"to":"request.dest"
				}
			]
		}
	}`
	middlewareObj := make(map[interface{}]interface{})
	err := yaml.Unmarshal([]byte(cfg), &middlewareObj)
	assert.Nil(t, err)
	_, err = setPropagateMiddleware(middlewareObj)
	assert.NotNil(t, err)
}

func TestGracefulShutdown(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	assert.NoError(t, err)
	bg := gateway.(*benchGateway.BenchGateway)
	bg.ActualGateway.Shutdown()
}

func TestGetModuleConfigNameWhenConfigFileIsYAML(t *testing.T) {
	expectedFileName := "yamlFile"
	instance := &ModuleInstance{
		JSONFileName: "unexpectedFileName",
		YAMLFileName: expectedFileName,
	}
	assert.Equal(t, expectedFileName, getModuleConfigFileName(instance))
}

func TestGetModuleConfigNameWhenConfigFileIsJSON(t *testing.T) {
	expectedFileName := "jsonFile"
	instance := &ModuleInstance{
		JSONFileName: expectedFileName,
		YAMLFileName: "",
	}
	assert.Equal(t, expectedFileName, getModuleConfigFileName(instance))
}

func TestExposedMethodKeyTypeError(t *testing.T) {
	clientConfig := &ClientClassConfig{
		Config: map[string]interface{}{
			"exposedMethods": map[interface{}]interface{}{
				1: "justice",
			},
		},
	}
	exposedMethods, err := getExposedMethods(clientConfig)
	assert.Nil(t, exposedMethods)
	assert.Error(t, err)
}

func TestExposedMethodValueTypeError(t *testing.T) {
	clientConfig := &ClientClassConfig{
		Config: map[string]interface{}{
			"exposedMethods": map[interface{}]interface{}{
				"sicence": 2.71,
			},
		},
	}
	exposedMethods, err := getExposedMethods(clientConfig)
	assert.Nil(t, exposedMethods)
	assert.Error(t, err)
}

func TestExposedMethodsTypeError(t *testing.T) {
	clientConfig := &ClientClassConfig{
		Config: map[string]interface{}{
			"exposedMethods": []string{
				"should", "be", "map",
			},
		},
	}
	exposedMethods, err := getExposedMethods(clientConfig)
	assert.Nil(t, exposedMethods)
	assert.Error(t, err)
}

func TestExposedMethodsKeyString(t *testing.T) {
	expectedMethods := map[string]string{
		"method1": "func1",
		"method2": "func2",
	}

	clientConfig := &ClientClassConfig{
		Config: map[string]interface{}{
			"exposedMethods": map[string]interface{}{
				"method1": "func1",
				"method2": "func2",
			},
		},
	}
	exposedMethods, err := getExposedMethods(clientConfig)
	assert.NoError(t, err)
	assert.Equal(t, expectedMethods, exposedMethods)
}

func TestExposedMethodsKeyInterface(t *testing.T) {
	expectedMethods := map[string]string{
		"method1": "func1",
		"method2": "func2",
	}

	clientConfig := &ClientClassConfig{
		Config: map[string]interface{}{
			"exposedMethods": map[interface{}]interface{}{
				"method1": "func1",
				"method2": "func2",
			},
		},
	}
	exposedMethods, err := getExposedMethods(clientConfig)
	assert.NoError(t, err)
	assert.Equal(t, expectedMethods, exposedMethods)
}

func TestExposedMethodsDuplication(t *testing.T) {
	clientConfig := &ClientClassConfig{
		Config: map[string]interface{}{
			"exposedMethods": map[string]interface{}{
				"method1": "func1",
				"method2": "func1",
			},
		},
	}
	exposedMethods, err := getExposedMethods(clientConfig)
	assert.Nil(t, exposedMethods)
	assert.Error(t, err)
}
