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

package codegen_test

import (
	"encoding/json"
	"flag"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
)

var updateGoldenFile = flag.Bool("update", false, "Updates the golden files with expected response.")

const parsedBarFile = "data/bar.json"

func TestModuleSpec(t *testing.T) {
	barThrift := "../../examples/example-gateway/idl/github.com/uber/zanzibar/endpoints/bar/bar.thrift"
	m, err := codegen.NewModuleSpec(barThrift, newPackageHelper(t))
	assert.NoError(t, err, "unable to parse the thrift file")
	convertThriftPathToRelative(m)
	actual, err := json.MarshalIndent(m, "", "\t")
	assert.NoError(t, err, "Unable to marshall response: err = %s", err)
	CompareGoldenFile(t, parsedBarFile, actual, *updateGoldenFile)
}

func convertThriftPathToRelative(m *codegen.ModuleSpec) {
	index := strings.Index(m.ThriftFile, "zanzibar")
	m.ThriftFile = m.ThriftFile[index:]

	for _, service := range m.Services {
		service.ThriftFile = service.ThriftFile[index:]
		for _, method := range service.Methods {
			if method.Downstream != nil {
				convertThriftPathToRelative(method.Downstream)
			}
			method.CompiledThriftSpec = nil
		}
	}
}
