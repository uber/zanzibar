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

package gencode_test

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/lib/gencode"
)

type MethodSlice []*gencode.MethodSpec

func (s MethodSlice) Len() int {
	return len(s)
}

func (s MethodSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s MethodSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func TestModuleSpec(t *testing.T) {
	barThrift := "../../examples/example-gateway/idl/github.com/uber/zanzibar/clients/bar/bar.thrift"
	m, err := gencode.NewModuleSpec(barThrift, newPackageHelper())
	assert.NoError(t, err, "unable to parse the thrift file")
	b, err := ioutil.ReadFile("bar.json")
	assert.NoError(t, err, "unable to read the json file")
	exp := new(gencode.ModuleSpec)
	err = json.Unmarshal(b, exp)
	assert.NoError(t, err, "unable to unmarshal expected file")
	assert.True(t, strings.HasSuffix(m.Services[0].ThriftFile, exp.Services[0].ThriftFile), "the ThriftFile should have relative path as in exp")
	exp.Services[0].ThriftFile = m.Services[0].ThriftFile

	sort.Sort(MethodSlice(exp.Services[0].Methods))
	sort.Sort(MethodSlice(m.Services[0].Methods))
	assert.True(t, reflect.DeepEqual(*exp, *m), "unexpected generated module spec: exp=\n%+v, actual=\n%+v", exp.Services[0], m.Services[0])
}
