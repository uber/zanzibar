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
	"bufio"
	"bytes"
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
	clientConfig, err := generator.readConfig(instance.JSONFileRaw)
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

	// TODO: add helper to template to return byte array
	clientBuffer := bytes.NewBuffer(nil)
	clientWriter := bufio.NewWriter(clientBuffer)
	if err := generator.templates.template.ExecuteTemplate(
		clientWriter,
		"http_client.tmpl",
		clientMeta,
	); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error executing HTTP Client template for %s",
			instance.InstanceName,
		)
	}
	if err := clientWriter.Flush(); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating HTTP client from template for %s",
			instance.InstanceName,
		)
	}

	structsBuffer := bytes.NewBuffer(nil)
	structsWriter := bufio.NewWriter(structsBuffer)
	if err := generator.templates.template.ExecuteTemplate(
		structsWriter,
		"structs.tmpl",
		clientMeta,
	); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error executing HTTP client structs template for %s",
			instance.InstanceName,
		)
	}
	if err := structsWriter.Flush(); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating HTTP client structs from template for %s",
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
	files[clientFilePath] = clientBuffer.Bytes()
	files[structFilePath] = structsBuffer.Bytes()
	return files, nil
}

// ReadConfig parses the raw interface config to an HTTPClientConfig struct
func (*HTTPClientGenerator) readConfig(
	rawConfig []byte,
) (*clientClassConfig, error) {
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
