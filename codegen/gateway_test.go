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
	"testing"

	yaml "github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"

	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
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
	middlewareObj := make(map[string]interface{})
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
