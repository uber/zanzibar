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

package codegen

import (
	"path/filepath"

	"encoding/json"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/module"
)

// NewDefaultModuleSystem creates a fresh instance of the default zanzibar
// module system (clients, endpoints)
func NewDefaultModuleSystem(h *PackageHelper) (*module.System, error) {
	system := module.NewSystem()
	tmpl, err := NewTemplate(templateDir)

	if err != nil {
		return nil, err
	}

	// Register client module class and type generators
	if err := system.RegisterClass("client", module.Class{
		Directory:         "clients",
		ClassType:         module.MultiModule,
		ClassDependencies: []string{},
	}); err != nil {
		return nil, errors.Wrapf(err, "Error registering client class")
	}

	if err := system.RegisterClassType("client", "http", &HTTPClientGenerator{
		templates:     tmpl,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering http client class type",
		)
	}

	if err := system.RegisterClassType("client", "tchannel", &TCahnnelClientGenerator{
		templates:     tmpl,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering TChannel client class type",
		)
	}

	// TODO: Register endpoint module class and type generators
	return system, nil
}

/*
 * HTTP Client Generator
 */

// HTTPClientGenerator generates an instance of a zanzibar http client
type HTTPClientGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
}

// Generate returns the HTTP client generated files as a map of relative file
// path (relative to the target buid directory) to file bytes.
func (generator *HTTPClientGenerator) Generate(
	instance *module.Instance,
) (map[string][]byte, error) {
	// Parse the client config from the endpoint JSON file
	clientConfig, err := readClientConfig(instance.JSONFileRaw)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading HTTP client %s JSON config",
			instance.InstanceName,
		)
	}

	clientSpec, err := NewHTTPClientSpec(
		filepath.Join(
			instance.BaseDirectory,
			instance.Directory,
			instance.JSONFileName,
		),
		clientConfig,
		generator.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error initializing HTTPClientSpec for %s",
			instance.InstanceName,
		)
	}

	clientMeta := &ClientMeta{
		PackageName:      clientSpec.ModuleSpec.PackageName,
		Services:         clientSpec.ModuleSpec.Services,
		IncludedPackages: clientSpec.ModuleSpec.IncludedPackages,
		ClientID:         clientSpec.ClientID,
	}

	client, err := generator.templates.execTemplate(
		"http_client.tmpl",
		clientMeta,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error executing HTTP client template for %s",
			instance.InstanceName,
		)
	}

	structs, err := generator.templates.execTemplate("structs.tmpl", clientMeta)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error executing HTTP client structs template for %s",
			instance.InstanceName,
		)
	}

	clientDirectory := filepath.Join(
		generator.packageHelper.CodeGenTargetPath(),
		instance.Directory,
	)

	clientFilePath, err := filepath.Rel(clientDirectory, clientSpec.GoFileName)
	if err != nil {
		clientFilePath = clientSpec.GoFileName
	}

	structFilePath, err := filepath.Rel(
		clientDirectory,
		clientSpec.GoStructsFileName,
	)
	if err != nil {
		structFilePath = clientSpec.GoStructsFileName
	}

	// Return the client files
	files := map[string][]byte{}
	files[clientFilePath] = client
	files[structFilePath] = structs
	return files, nil
}

/*
 * TChannel Client Generator
 */

// TCahnnelClientGenerator generates an instance of a zanzibar TChannel client
type TCahnnelClientGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
}

// Generate returns the TChannel client generated files as a map of relative file
// path (relative to the target build directory) to file bytes.
func (generator *TCahnnelClientGenerator) Generate(
	instance *module.Instance,
) (map[string][]byte, error) {
	// Parse the client config from the endpoint JSON file
	clientConfig, err := readClientConfig(instance.JSONFileRaw)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading TChannel client %s JSON config",
			instance.InstanceName,
		)
	}

	clientSpec, err := NewTChannelClientSpec(
		filepath.Join(
			instance.BaseDirectory,
			instance.Directory,
			instance.JSONFileName,
		),
		clientConfig,
		generator.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error initializing TChannelClientSpec for %s",
			instance.InstanceName,
		)
	}

	clientMeta := &ClientMeta{
		PackageName:      clientSpec.ModuleSpec.PackageName,
		Services:         clientSpec.ModuleSpec.Services,
		IncludedPackages: clientSpec.ModuleSpec.IncludedPackages,
		ClientID:         clientSpec.ClientID,
	}

	client, err := generator.templates.execTemplate(
		"tchannel_client.tmpl",
		clientMeta,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error executing TChannel client template for %s",
			instance.InstanceName,
		)
	}

	clientDirectory := filepath.Join(
		generator.packageHelper.CodeGenTargetPath(),
		instance.Directory,
	)

	clientFilePath, err := filepath.Rel(clientDirectory, clientSpec.GoFileName)
	if err != nil {
		clientFilePath = clientSpec.GoFileName
	}

	// Return the client files
	files := map[string][]byte{}
	files[clientFilePath] = client
	return files, nil
}

func readClientConfig(rawConfig []byte) (*clientClassConfig, error) {
	var clientConfig clientClassConfig
	if err := json.Unmarshal(rawConfig, &clientConfig); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading config for HTTP client instance",
		)
	}
	clientConfig.Config["clientId"] = clientConfig.Name
	clientConfig.Config["clientType"] = clientConfig.Type
	return &clientConfig, nil
}
