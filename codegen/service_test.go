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

package codegen_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
	testlib "github.com/uber/zanzibar/test/lib"
)

const parsedBarFile = "test_data/bar.json"

func TestModuleSpec(t *testing.T) {
	barThrift := "../examples/example-gateway/idl/clients/bar/bar.thrift"
	m, err := codegen.NewModuleSpec(barThrift, true, false, newPackageHelper(t))
	assert.NoError(t, err, "unable to parse the thrift file")
	convertThriftPathToRelative(m)
	actual, err := json.MarshalIndent(m, "", "\t")
	assert.NoError(t, err, "Unable to marshall response: err = %s", err)
	testlib.CompareGoldenFile(t, parsedBarFile, actual)
}

func TestExceptionValidation(t *testing.T) {
	var (
		barClientThrift   = "../examples/example-gateway/idl/clients/bar/bar.thrift"
		barEndpointThrift = "../examples/example-gateway/idl/endpoints/bar/bar.thrift"
		pkgHelper         = newPackageHelper(t)
	)
	m, err := codegen.NewModuleSpec(barEndpointThrift, true, false, pkgHelper)
	assert.NoError(t, err)
	cs, err := codegen.NewModuleSpec(barClientThrift, true, false, pkgHelper)
	assert.NoError(t, err)
	service := m.Services[0]
	method := service.Methods[0]
	assert.Equal(t, method.Name, "argNotStruct")
	clientSpec := &codegen.ClientSpec{
		ExposedMethods: map[string]string{"argNotStruct": "Bar::argNotStruct"},
		ModuleSpec:     cs,
	}
	assert.NoError(t, err)
	method.ExceptionsIndex = map[string]codegen.ExceptionSpec{"test": {}}
	e := &codegen.EndpointSpec{
		ThriftServiceName: "Bar",
		ThriftMethodName:  "argNotStruct",
		ClientSpec:        clientSpec,
		ClientMethod:      "argNotStruct",
	}
	err = m.SetDownstream(e, pkgHelper)
	assert.NotNil(t, err)
}

func convertThriftPathToRelative(m *codegen.ModuleSpec) {
	index := strings.LastIndex(m.ThriftFile, "zanzibar")
	m.ThriftFile = m.ThriftFile[index:]

	m.CompiledModule = nil
	for _, service := range m.Services {
		service.CompileSpec = nil
		service.ThriftFile = service.ThriftFile[index:]
		for _, method := range service.Methods {
			if method.Downstream != nil {
				convertThriftPathToRelative(method.Downstream)
			}
			method.CompiledThriftSpec = nil
		}
	}
}
