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

package repository

import (
	"io/ioutil"
	"path/filepath"

	"github.com/uber/zanzibar/codegen"
	"go.uber.org/thriftrw/compile"

	"github.com/pkg/errors"
)

const (
	runtimePrefix = "edge-gateways-runtime"
)

// Manager manages all remote repositories.
type Manager struct {
	// Maps from gateway ID to its repository.
	// This map is created once and read-only afterwards.
	RepoMap map[string]*Repository

	// root directory for local copies of remote repositories.
	localRootDir string

	// IDL-registry repository.
	IDLRegistry IDLRegistry

	annoPrefix string
}

// NewManager creates a manager for remote git repositories.
func NewManager(
	repoMap map[string]*Repository,
	localRoot string,
	idlRegistry IDLRegistry,
	annoPrefix string,
) *Manager {
	manager := &Manager{
		RepoMap:      repoMap,
		localRootDir: localRoot,
		IDLRegistry:  idlRegistry,
		annoPrefix:   annoPrefix,
	}
	return manager
}

// NewRuntimeRepository creates a repository for running Zanzibar with new configurations.
func (m *Manager) NewRuntimeRepository(gatewayID string) (*Repository, error) {
	r, ok := m.RepoMap[gatewayID]
	if !ok {
		return nil, errors.Errorf("gateway %q not found", gatewayID)
	}
	root, err := ioutil.TempDir(m.localRootDir, runtimePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create runtime directory")
	}
	return NewRepository(root, r.Remote(), r.fetcher, r.refreshInterval)
}

// IDLThriftService returns thrift services contained in a thrift file in IDL-registry.
func (m *Manager) IDLThriftService(path string) (map[string]*ThriftService, error) {
	if err := m.IDLRegistry.Update(); err != nil {
		return nil, err
	}

	localRootDir := m.IDLRegistry.RootDir()
	thriftRootDir := m.IDLRegistry.ThriftRootDir()
	packageHelper, err := codegen.NewPackageHelper(
		"idl-registry",          // packgeRoot
		"",                      // managedThriftFolder
		localRootDir,            // configDirName
		"",                      // middlewareConfig
		thriftRootDir,           // thriftRootDir
		"idl-registry/gen-code", // genCodePackage
		"./build",               // targetGenDir
		"",                      // copyrightHeader
		m.annoPrefix,
	)
	if err != nil {
		return nil, err
	}

	// Parse service module as tchannel service.
	thriftAbsPath := filepath.Join(localRootDir, thriftRootDir, path)
	mspec, err := codegen.NewModuleSpec(thriftAbsPath, false, false, packageHelper)
	if err != nil {
		return nil, err
	}
	serviceType := TCHANNEL
	// Parse HTTP annotations.
	if _, err := codegen.NewModuleSpec(thriftAbsPath, true, false, packageHelper); err == nil {
		serviceType = HTTP
	}
	serviceMap := map[string]*ThriftService{}
	for _, service := range mspec.Services {
		tservice := &ThriftService{
			Name: service.Name,
			Path: path,
		}
		tservice.Methods = make([]ThriftMethod, len(service.Methods))
		for i, method := range service.Methods {
			tservice.Methods[i].Name = method.Name
			tservice.Methods[i].Type = serviceType
		}
		serviceMap[service.Name] = tservice
	}
	return serviceMap, nil
}

// ThriftList returns the full list of thrift files in a gateway.
func (m *Manager) ThriftList(gateway string) (map[string]*ThriftMeta, error) {
	repo, ok := m.RepoMap[gateway]
	if !ok {
		return nil, errors.Errorf("gateway %s is not found", gateway)
	}
	cfg, err := repo.GatewayConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get gateway config")
	}
	return repo.ThriftConfig(cfg.ThriftRootDir)
}

// ThriftFile returns the content and meta data of a file in a gateway.
func (m *Manager) ThriftFile(gateway, path string) (*ThriftMeta, error) {
	repo, ok := m.RepoMap[gateway]
	if !ok {
		return nil, errors.Errorf("gateway %s is not found", gateway)
	}
	cfg, err := repo.GatewayConfig()
	if err != nil {
		return nil, err
	}
	thriftRootDir := cfg.ThriftRootDir
	content, err := repo.ReadFile(filepath.Join(thriftRootDir, path))
	if err != nil {
		return nil, errors.Wrapf(err, "faile to read thrift file content: %s", path)
	}
	version, err := repo.ThriftFileVersion(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get thrift file version")
	}
	return &ThriftMeta{
		Path:    path,
		Version: version,
		Content: string(content),
	}, nil
}

// CompileThriftFile returns the content and meta data of a file in a gateway.
func (m *Manager) CompileThriftFile(gateway, path string) (*Module, error) {
	repo, ok := m.RepoMap[gateway]
	if !ok {
		return nil, errors.Errorf("gateway %s is not found", gateway)
	}
	cfg, err := repo.GatewayConfig()
	if err != nil {
		return nil, err
	}
	prefix := filepath.Join(repo.LocalDir(), cfg.ThriftRootDir)
	absPath := filepath.Join(prefix, path)
	return CompileThriftFile(absPath, prefix)
}

// CompileThriftCode parses raw bytes in to compiled thrift.
func (m *Manager) CompileThriftCode(rawCode string) (*Module, error) {
	return CompileThriftCode([]byte(rawCode), "", "")
}

// CodeThriftFile returns the content and meta data of a file in a gateway.
func (m *Manager) CodeThriftFile(rawCode, gateway, path string) (*Module, error) {
	repo, ok := m.RepoMap[gateway]
	if !ok {
		return nil, errors.Errorf("gateway %s is not found", gateway)
	}
	cfg, err := repo.GatewayConfig()
	if err != nil {
		return nil, err
	}
	prefix := filepath.Join(repo.LocalDir(), cfg.ThriftRootDir)
	absPath := filepath.Join(prefix, path)
	return CompileThriftCode([]byte(rawCode), absPath, prefix)
}

// UpdateThriftFiles update thrift files to their master version in the IDL-registry.
func (m *Manager) UpdateThriftFiles(r *Repository, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	newMeta, err := m.thriftMetaInIDLRegistry(paths)
	if err != nil {
		return errors.Wrap(err, "failed to get thrift files from IDL-registry")
	}

	if err := r.WriteThriftFileAndConfig(newMeta); err != nil {
		return errors.Wrapf(err, "failed to update current thrift config")
	}
	return nil
}

// UpdateManagedThriftFiles will write new thrift files to disk.
func (m *Manager) UpdateManagedThriftFiles(
	r *Repository, files []ManagedThriftFile,
) error {
	if len(files) == 0 {
		return nil
	}

	if err := r.WriteManagedThriftFiles(files); err != nil {
		return errors.Wrap(err, "failed to write managed thrift files")
	}
	return nil
}

// UpdateClients updates configurations for a list of clients.
func (m *Manager) UpdateClients(r *Repository, clientCfgDir string, req []ClientConfig) error {
	for i := range req {
		thrift := req[i].ThriftFile
		// Adds non-existing file into the repository.
		version, versionErr := r.ThriftFileVersion(thrift)
		if versionErr != nil {
			if err := m.UpdateThriftFiles(r, []string{thrift}); err != nil {
				return errors.Wrapf(err, "failed to add thrift file %s into temp repository", thrift)
			}
		}
		if err := r.UpdateClientConfigs(&req[i], clientCfgDir, version); err != nil {
			return errors.Wrapf(err, "failed to add thrift file %s into temp repository", thrift)
		}
	}
	return nil
}

// UpdateEndpoints updates configurations for a list of endpoints.
func (m *Manager) UpdateEndpoints(r *Repository, endpointCfgDir string, req []EndpointConfig) error {
	for i := range req {
		thrift := req[i].ThriftFile
		// Adds non-existing file into the repository.
		version, versionErr := r.ThriftFileVersion(thrift)
		if versionErr != nil {
			if err := m.UpdateThriftFiles(r, []string{thrift}); err != nil {
				return errors.Wrapf(err, "failed to add thrift file %s into temp repository", thrift)
			}
		}
		if err := r.WriteEndpointConfig(endpointCfgDir, &req[i], version); err != nil {
			return errors.Wrapf(err, "failed to add thrift file %s into temp repository", thrift)
		}
	}
	return nil
}

// ManagedThriftFile represents a thrift file we want to write
type ManagedThriftFile struct {
	SourceCode string `json:"source_code"`
	Filename   string `json:"filename"`
}

// UpdateRequest is the request to update thrift files, clients and endpoints.
type UpdateRequest struct {
	ThriftFiles        []string            `json:"thrift_files"`
	ManagedThriftFiles []ManagedThriftFile `json:"managed_thrift_files"`
	ClientUpdates      []ClientConfig      `json:"client_updates"`
	EndpointUpdates    []EndpointConfig    `json:"endpoint_updates"`
	DIffID             *string             `json:"diff_id,omitempty"`
}

// Validate validates the request and surface user friendly error
func (m *Manager) Validate(r *Repository, req *UpdateRequest) error {
	cfg, err := r.GatewayConfig()
	if err != nil {
		return err
	}
	packageHelper, err := codegen.NewPackageHelper(
		cfg.PackageRoot,
		cfg.ManagedThriftFolder,
		r.LocalDir(),
		cfg.MiddlewareConfigDir,
		m.IDLRegistry.ThriftRootDir(),
		cfg.GenCodePackage,
		"./build",
		"",
		m.annoPrefix,
	)
	if err != nil {
		return err
	}
	moduleSystem, err := codegen.NewDefaultModuleSystem(packageHelper)
	if err != nil {
		return err
	}
	_, err = moduleSystem.GenerateBuild(
		packageHelper.PackageRoot(),
		r.LocalDir(),
		packageHelper.CodeGenTargetPath(),
		false,
	)
	return err
}

// UpdateAll writes the updates for thrift files, clients and endpoints.
func (m *Manager) UpdateAll(r *Repository, clientCfgDir, endpointCfgDir string, req *UpdateRequest) error {
	if err := m.UpdateManagedThriftFiles(r, req.ManagedThriftFiles); err != nil {
		return errors.Wrap(err, "Failed to update managed thrift files")
	}
	if err := m.UpdateThriftFiles(r, req.ThriftFiles); err != nil {
		return errors.Wrap(err, "failed to update thrift files")
	}
	if err := m.UpdateClients(r, clientCfgDir, req.ClientUpdates); err != nil {
		return errors.Wrap(err, "failed to update clients")
	}
	if err := m.UpdateEndpoints(r, endpointCfgDir, req.EndpointUpdates); err != nil {
		return errors.Wrap(err, "failed to update endpoints")
	}
	return nil
}

// thriftMetaInIDLRegistry returns meta for a set of thrift file in IDL-registry.
func (m *Manager) thriftMetaInIDLRegistry(paths []string) (map[string]*ThriftMeta, error) {
	meta := make(map[string]*ThriftMeta, len(paths))
	idlRootAbsPath := filepath.Join(m.IDLRegistry.RootDir(), m.IDLRegistry.ThriftRootDir())
	for _, path := range paths {
		module, err := compile.Compile(filepath.Join(idlRootAbsPath, path))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse thrift file %s", path)
		}
		if err := addThriftDependencies(idlRootAbsPath, module, meta); err != nil {
			return nil, errors.Wrapf(err, "failed to add dependencies for thrift file %s", path)
		}
	}
	for path := range meta {
		tm, err := m.IDLRegistry.ThriftMeta(path, true)
		if err != nil {
			return nil, err
		}
		meta[path] = tm
	}
	return meta, nil
}

func addThriftDependencies(idlRoot string, module *compile.Module, meta map[string]*ThriftMeta) error {
	relPath, err := filepath.Rel(idlRoot, module.ThriftPath)
	if err != nil {
		return errors.Wrapf(err, "failed to find relative path for thrift file %q under dir %q",
			module.ThriftPath, idlRoot)
	}
	if _, ok := meta[relPath]; ok {
		return nil
	}
	// Add the thrift file to meta.
	meta[relPath] = nil
	for _, includedModule := range module.Includes {
		if err := addThriftDependencies(idlRoot, includedModule.Module, meta); err != nil {
			return err
		}
	}
	return nil
}
