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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	testlib "github.com/uber/zanzibar/test/lib"
)

func TestThriftConfig(t *testing.T) {
	tempDir, err := copyExample(t)
	if !assert.NoError(t, err, "Failed to copy example.") {
		return
	}
	r := &Repository{
		localDir: tempDir,
	}
	meta, err := r.ThriftConfig(r.LocalDir())
	if !assert.NoError(t, err, "Failed to read thrift configs.") {
		return
	}
	b, err := json.MarshalIndent(meta, "", "\t")
	b = append(b, []byte("\n")...)
	if !assert.NoError(t, err, "Failed to marshal meta to bytes.") {
		return
	}
	testlib.CompareGoldenFile(t, "../../examples/example-gateway/idls.json", b)
}

func TestThriftVersion(t *testing.T) {

	tempDir, err := copyExample(t)
	if !assert.NoError(t, err, "Failed to copy example.") {
		return
	}
	r := &Repository{
		localDir: tempDir,
	}
	v, err := r.ThriftFileVersion("endpoints/bar/bar.thrift")
	if !assert.NoError(t, err, "Failed to read thrift version.") {
		return
	}
	assert.Equal(t, "v1", v)

	_, err = r.ThriftFileVersion("no-such-file.thrift")
	assert.Error(t, err, "should failed to fetch a non exist file.")
}

func TestWriteThriftFileAndConfig(t *testing.T) {
	path := "a/new/file/name.thrift"
	version := "new version"
	newMeta := map[string]*ThriftMeta{
		path: &ThriftMeta{
			Path:    path,
			Version: version,
			Content: "filecontent",
		},
	}

	tempDir, err := copyExample(t)
	if !assert.NoError(t, err, "Failed to copy example.") {
		return
	}
	r := &Repository{
		localDir: tempDir,
	}

	err = r.WriteThriftFileAndConfig(newMeta)
	if !assert.NoError(t, err, "Failed to write new thrift configuration.") {
		return
	}

	v, err := r.ThriftFileVersion(path)
	if !assert.NoError(t, err, "Failed to read thrift version.") {
		return
	}
	assert.Equal(t, version, v)
}
