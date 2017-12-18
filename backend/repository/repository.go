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
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
)

type meta struct {
	lastUpdate time.Time
	version    string
}

// Repository operates one local repository:
// get/set configurations and sync with its remote repository.
type Repository struct {
	remote          string
	localDir        string
	fetcher         Fetcher
	refreshInterval time.Duration
	// Cached gateway config.
	gatewayCfgVal atomic.Value
	// Cached error when getting gateway config.
	gatewayCfgErr atomic.Value
	meta          atomic.Value
}

// Fetcher fetches remote repository to local repository.
type Fetcher interface {
	// Clone copys remote repository to localRoot and returns the path to the
	// repository or an error.
	Clone(localRoot, remote string) (string, error)
	// Update syncs local repository with remote repository and returns whether
	// there is update or an error.
	Update(localDir, remote string) (bool, error)
	// Version returns current version of local repository or an error.
	Version(localDir string) (string, error)
}

// NewRepository creates a repository from a remote git repository.
func NewRepository(localRoot, remote string, fetcher Fetcher,
	refreshInterval time.Duration) (*Repository, error) {
	localDir, err := fetcher.Clone(localRoot, remote)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to create local repository for %s", remote)
	}
	version, err := fetcher.Version(localDir)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to get the version of local repository for %s", remote)
	}

	meta := &meta{
		version:    version,
		lastUpdate: time.Now(),
	}
	repo := &Repository{
		remote:          remote,
		localDir:        localDir,
		fetcher:         fetcher,
		refreshInterval: refreshInterval,
	}
	repo.meta.Store(meta)
	return repo, nil
}

// Remote returns the remote for the repository.
func (r *Repository) Remote() string {
	return r.remote
}

// LocalDir returns the local directory for the repository.
func (r *Repository) LocalDir() string {
	return r.localDir
}

// Version returns the version of local repository.
func (r *Repository) Version() string {
	meta := r.meta.Load().(*meta)
	return meta.version
}

// Update update the local repository and returns whether there is an update.
func (r *Repository) Update() bool {
	now := time.Now()
	meta := r.meta.Load().(*meta)
	if now.Sub(meta.lastUpdate) < r.refreshInterval {
		return false
	}
	isUpdated, err := r.fetcher.Update(r.localDir, r.remote)
	if err != nil {
		return false
	}
	version, err := r.fetcher.Version(r.localDir)
	if err != nil {
		meta.version = "unknown"
	} else {
		meta.version = version
	}
	meta.lastUpdate = now
	r.meta.Store(meta)
	return isUpdated
}

// ReadFile returns the content of a file in a relativePath in a repository.
func (r *Repository) ReadFile(relativePath string) ([]byte, error) {
	path := filepath.Join(r.localDir, relativePath)
	return ioutil.ReadFile(path)
}
