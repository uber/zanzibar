package repository

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

const gatewayConfigFile = "gateway.json"

// Repository operates one local repository:
// get/set configurations and sync with its remote repository.
type Repository struct {
	remote          string
	localDir        string
	fetcher         Fetcher
	refreshInterval time.Duration

	// The following fields and the content of local repository are protected
	// by RWMutex.
	version        string
	lastUpdateTime time.Time
	sync.RWMutex
}

// Fetcher fetches remote repository to local repository.
type Fetcher interface {
	// Clone copys remote repository to localRoot and returns the path to the
	// repository or an error.
	Clone(localRoot, remote string) (string, error)
	// Update syncs local repository with remote repository and returns whether
	// there is update or an error.
	Update(localDir, remote string) (bool, error)
	// Version returns current version of local respository or an error.
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
			"failed to get the version of local respository for %s", remote)
	}
	return &Repository{
		remote:          remote,
		localDir:        localDir,
		version:         version,
		fetcher:         fetcher,
		lastUpdateTime:  time.Now(),
		refreshInterval: refreshInterval,
	}, nil
}

// Remote returns the remote for the respository.
func (r *Repository) Remote() string {
	return r.remote
}

// LocalDir returns the local directory for the respository.
func (r *Repository) LocalDir() string {
	return r.localDir
}

// Commit returns the version of local repository.
func (r *Repository) Version() string {
	r.RLock()
	defer r.RUnlock()
	return r.version
}
