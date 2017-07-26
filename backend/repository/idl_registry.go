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
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
)

// IDLRegistry provides the content and version of a thrift file.
type IDLRegistry interface {
	// RootDir returns the path of the root directory for the registry.
	RootDir() string
	// ThriftRootDir returns the relative path of the thrift root directory.
	ThriftRootDir() string
	// Updates the IDLRegistry to sync with remote source.
	Update() error
	// Returns the file version or content of a thrift file.
	ThriftMeta(path string, needFileContent bool) (*ThriftMeta, error)
}

type idlRegistry struct {
	idlDir string
	repo   *Repository
	// Map from thrift file to its meta, which is updated when underlying Repository being updated.
	metaMap map[string]*ThriftMeta
	// Lock for metaMap.
	lock sync.RWMutex
}

const metaJSONFile = "meta.json"

// NewIDLRegistry assumes all thrift files are located under the dir of
// relative path `idlDirRel` of a single repository and there is a
// meta.json file containing all version information of thrift files.
func NewIDLRegistry(idlDirRel string, repo *Repository) (IDLRegistry, error) {
	newMap, err := allFileMeta(filepath.Join(repo.LocalDir(), metaJSONFile))
	if err != nil {
		return nil, err
	}
	return &idlRegistry{
		idlDir:  idlDirRel,
		repo:    repo,
		metaMap: newMap,
	}, nil
}

// RootDir returns the path of the root directory for the registry.
func (reg *idlRegistry) RootDir() string {
	return reg.repo.LocalDir()
}

// ThriftRootDir returns the relative path of the thrift root directory.
func (reg *idlRegistry) ThriftRootDir() string {
	return reg.idlDir
}

// IDLRegistryFile returns the content and meta data of a file in IDL-registry.
func (reg *idlRegistry) ThriftMeta(path string, needFileContent bool) (*ThriftMeta, error) {
	if err := reg.Update(); err != nil {
		return nil, err
	}
	reg.lock.RLock()
	defer reg.lock.RUnlock()
	meta, ok := reg.metaMap[path]
	if !ok {
		return nil, errors.Errorf("failed to find file %q", path, reg.metaMap)
	}
	if !needFileContent {
		return meta, nil
	}
	file := filepath.Join(reg.repo.LocalDir(), reg.idlDir, path)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "faile to read thrift file content %q", path)
	}
	return &ThriftMeta{
		Path:    path,
		Version: meta.Version,
		Content: string(content),
	}, nil
}

// Types used to parse meta.json file under idlRegistry repository.
type idlMetaJSON struct {
	// Remote name -> Remote meta.
	Remotes map[string]remoteMeta `json:"remotes"`
}

type remoteMeta struct {
	// File subpath -> SHASureg.
	SHASums map[string]string `json:"shasums"`
}

func allFileMeta(metaJSONPath string) (map[string]*ThriftMeta, error) {
	var metaJSONContent idlMetaJSON
	if err := readJSONFile(metaJSONPath, &metaJSONContent); err != nil {
		return nil, errors.Wrap(err, "failed to read meta.json of IDL registry")
	}
	meta := make(map[string]*ThriftMeta, len(metaJSONContent.Remotes))
	for remote, rMeta := range metaJSONContent.Remotes {
		for subpath, sha := range rMeta.SHASums {
			path := filepath.Join(remote, subpath)
			meta[path] = &ThriftMeta{
				Path:    path,
				Version: sha,
			}
		}
	}
	return meta, nil
}

// Update updates the repository.
func (reg *idlRegistry) Update() error {
	if !reg.repo.Update() {
		return nil
	}
	newMap, err := allFileMeta(filepath.Join(reg.repo.LocalDir(), metaJSONFile))
	if err != nil {
		return err
	}
	reg.lock.Lock()
	defer reg.lock.Unlock()
	reg.metaMap = newMap
	return nil
}
