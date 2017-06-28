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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/mock/mock_repository"
)

func TestNewRepository(t *testing.T) {
	root := os.TempDir()
	localDir := filepath.Join(root, "repo")
	remote := "gitolite@domain/repo"
	fetcher := mock_repository.NewMockFetcher(gomock.NewController(t))
	fetcher.EXPECT().Clone(root, remote).Return(localDir, nil)
	fetcher.EXPECT().Version(localDir).Return("version1", nil)
	repository, err := NewRepository(root, remote, fetcher, 1*time.Minute)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, remote, repository.Remote(), "Remote mismatch.")
	assert.Equal(t, localDir, repository.LocalDir(), "LocalDir Mismatch.")
	assert.Equal(t, "version1", repository.Version(), "Version Mismatch.")
}

func TestNewRepositoryWithCloneError(t *testing.T) {
	root := os.TempDir()
	remote := "gitolite@domain/repo"
	fetcher := mock_repository.NewMockFetcher(gomock.NewController(t))
	errClone := errors.New("failed to clone")
	fetcher.EXPECT().Clone(root, remote).Return("", errClone)
	_, err := NewRepository(root, remote, fetcher, 1*time.Minute)
	assert.Error(t, err, "Should return an error.")
}

func TestNewRepositoryWithVersionError(t *testing.T) {
	root := os.TempDir()
	remote := "gitolite@domain/repo"
	localDir := filepath.Join(root, "repo")
	fetcher := mock_repository.NewMockFetcher(gomock.NewController(t))
	errVersion := errors.New("failed to get version")
	fetcher.EXPECT().Clone(root, remote).Return(localDir, nil)
	fetcher.EXPECT().Version(localDir).Return("", errVersion)
	_, err := NewRepository(root, remote, fetcher, 1*time.Minute)
	assert.Error(t, err, "Should return an error.")
}

func TestRepositoryUpdateSuccessfully(t *testing.T) {
	fetcher := mock_repository.NewMockFetcher(gomock.NewController(t))
	r := &Repository{
		remote:          "remote",
		localDir:        "local_dir",
		version:         "init_version",
		fetcher:         fetcher,
		lastUpdateTime:  time.Now().Add(-1 * time.Minute),
		refreshInterval: 30 * time.Second,
	}
	fetcher.EXPECT().Update(r.localDir, r.remote).Return(true, nil)
	fetcher.EXPECT().Version(r.localDir).Return("new_version", nil)
	assert.Equal(t, true, r.Update(), "Failed to update repository.")
	assert.Equal(t, "new_version", r.Version(), "Failed to update version.")
}

func TestRepositoryUpdateTooSoon(t *testing.T) {
	fetcher := mock_repository.NewMockFetcher(gomock.NewController(t))
	r := &Repository{
		remote:          "remote",
		localDir:        "local_dir",
		version:         "init_version",
		fetcher:         fetcher,
		lastUpdateTime:  time.Now().Add(-1 * time.Minute),
		refreshInterval: 3 * time.Minute,
	}
	assert.Equal(t, false, r.Update(), "Should not update repository.")
}

func TestRepositoryUpdateWithError(t *testing.T) {
	fetcher := mock_repository.NewMockFetcher(gomock.NewController(t))
	r := &Repository{
		remote:          "remote",
		localDir:        "local_dir",
		version:         "init_version",
		fetcher:         fetcher,
		lastUpdateTime:  time.Now().Add(-1 * time.Minute),
		refreshInterval: 30 * time.Second,
	}
	errUpdate := errors.New("failed to upate")
	fetcher.EXPECT().Update(r.localDir, r.remote).Return(false, errUpdate)
	assert.Equal(t, false, r.Update(), "Should not update repository.")
	assert.Equal(t, "init_version", r.Version(), "Should not update version.")
}

func TestRepositoryUpdateWithUnknownVersion(t *testing.T) {
	fetcher := mock_repository.NewMockFetcher(gomock.NewController(t))
	r := &Repository{
		remote:          "remote",
		localDir:        "local_dir",
		version:         "init_version",
		fetcher:         fetcher,
		lastUpdateTime:  time.Now().Add(-1 * time.Minute),
		refreshInterval: 30 * time.Second,
	}
	fetcher.EXPECT().Update(r.localDir, r.remote).Return(true, nil)
	errVersion := errors.New("failed to fetch version")
	fetcher.EXPECT().Version(r.localDir).Return("", errVersion)
	assert.Equal(t, true, r.Update(), "Failed to update repository.")
	assert.Equal(t, "unknown", r.Version(), "Failed to update version.")
}
