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

package repository

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const exampleRegistry = "../../examples/idl-registry"

func TestIDLRegsitry(t *testing.T) {
	reg, err := NewTempIDLRegistry()
	if !assert.NoError(t, err, "Failed to create temp IDL registry.") {
		return
	}
	assert.NotEmpty(t, reg.RootDir(), "RootDir should not be empty.")
	assert.Equal(t, "idl", reg.ThriftRootDir(), "Thrift root is mismatched.")
	assert.NoError(t, reg.Update(), "Update should not return error.")

	thriftPath := "clients/baz/baz.thrift"
	meta, err := reg.ThriftMeta(thriftPath, true)
	if !assert.NoError(t, err, "Failed to get thrift meta.") {
		return
	}
	assert.Equal(t, "e0729d2985ec0aeb95891cba969f1fadecedb648", meta.Version, "Thrift version mismatched.")
	assert.Equal(t, thriftPath, meta.Path)
	assert.NotEmpty(t, meta.Content, "Thrift content should not be empty.")

	goldenFile := filepath.Join(exampleGateway, "idl", thriftPath)
	b, err := ioutil.ReadFile(goldenFile)
	if !assert.NoError(t, err, "Failed to read goldenFile.") {
		return
	}
	assert.Equal(t, string(b), meta.Content, "Thrift file content mismatched.")

	thriftPath = "clients/baz/baz.thrift"
	meta, err = reg.ThriftMeta(thriftPath, false)
	if !assert.NoError(t, err, "Failed to get thrift meta.") {
		return
	}
	assert.Equal(t, "e0729d2985ec0aeb95891cba969f1fadecedb648", meta.Version, "Thrift version mismatched.")
	assert.Equal(t, thriftPath, meta.Path)
	assert.Empty(t, meta.Content, "Thrift content should be empty.")
}

func NewTempIDLRegistry() (IDLRegistry, error) {
	rf := &registryFetcher{
		remote: "local-idl-registry",
	}
	r, err := NewRepository("", rf.remote, rf, -1*time.Second)
	if err != nil {
		return nil, err
	}
	return NewIDLRegistry("idl", r)
}

type registryFetcher struct {
	remote string
}

func (rf *registryFetcher) Clone(localRoot, remote string) (string, error) {
	if remote != rf.remote {
		return "", errors.Errorf("remote is expected to be %q but got %q", rf.remote, remote)
	}

	tempDir, err := ioutil.TempDir(localRoot, "zanzibar")
	if err != nil {
		return "", err
	}
	tempRegistry := filepath.Join(tempDir, "examples", "idl-registry")
	if err := os.MkdirAll(tempRegistry, os.ModePerm); err != nil {
		return "", err
	}

	err = copyDir(exampleRegistry, tempRegistry, []string{})
	if err != nil {
		return "", err
	}
	return tempRegistry, nil
}

func (rf *registryFetcher) Update(localDir, remote string) (bool, error) {
	return true, nil
}

func (rf *registryFetcher) Version(localDir string) (string, error) {
	return time.Now().String(), nil
}
