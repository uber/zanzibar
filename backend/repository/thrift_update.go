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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	metaJSONFilePath = "idls.json"
)

// ThriftConfig returns the meta of thrifts for a runtime repository.
func (r *Repository) ThriftConfig(idlRoot string) (map[string]*ThriftMeta, error) {
	r.RLock()
	defer r.RUnlock()
	config := make(map[string]*ThriftMeta)
	err := readJSONFile(r.absPath(metaJSONFilePath), &config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal idls.json file")
	}
	return config, nil
}

// WriteThriftFileAndConfig writes the update the file contents and meta config.
func (r *Repository) WriteThriftFileAndConfig(thriftMeta map[string]*ThriftMeta) error {
	cfg, err := r.LatestGatewayConfig()
	if err != nil {
		return errors.Wrap(err, "invalid configuration before updating thrifts")
	}
	thriftRootDir := cfg.ThriftRootDir
	curMeta, err := r.ThriftConfig(thriftRootDir)
	if err != nil {
		return err
	}
	// Merge updated thriftMeta into curMeta
	for path, meta := range thriftMeta {
		curMeta[path] = meta
	}
	r.Lock()
	defer r.Unlock()
	for _, meta := range thriftMeta {
		if err := r.writeThriftFile(thriftRootDir, meta); err != nil {
			return err
		}
		meta.Content = ""
	}
	return r.writeThriftConfig(curMeta)
}

// writeThriftFile update the content of a thrift file.
func (r *Repository) writeThriftFile(idlRoot string, meta *ThriftMeta) error {
	if meta == nil || meta.Path == "" || meta.Content == "" {
		return errors.Errorf("meta is invalid: %+v", meta)
	}
	path := r.absPath(filepath.Join(idlRoot, meta.Path))
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create dir %s", dir)
	}
	if err := ioutil.WriteFile(path, []byte(meta.Content), os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to write thrift file")
	}
	return nil
}

// writeThriftConfig writes the meta of trifts into a runtime repository.
func (r *Repository) writeThriftConfig(meta map[string]*ThriftMeta) error {
	b, err := json.MarshalIndent(meta, "", "\t")
	if err != nil {
		return errors.Wrap(err, "failed to marshal thrift meta")
	}
	path := r.absPath(metaJSONFilePath)
	err = ioutil.WriteFile(path, b, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "failed to write thrift idls.json")
	}
	return nil
}

// ThriftFileVersion returns the version of a thrft file.
func (r *Repository) ThriftFileVersion(thriftFile string) (string, error) {
	path := r.absPath(metaJSONFilePath)
	r.RLock()
	defer r.RUnlock()
	meta := make(map[string]*ThriftMeta)
	err := readJSONFile(path, &meta)
	if err != nil {
		return "", errors.Wrap(err, "failed to unmarshal the content of thrift meta file")
	}
	thriftMeta, ok := meta[thriftFile]
	if !ok {
		return "", errors.Errorf("can't find thrift file <%s> in thrift meta file", thriftFile)
	}
	return thriftMeta.Version, nil
}
