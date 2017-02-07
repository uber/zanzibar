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

package gencode

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// BuildModuleSpecForEndpoint builds the module spec for an endpoint.
func BuildModuleSpecForEndpoint(endpointDir string) (*ModuleSpec, error) {
	h := &PackageHelper{
		ThriftRootDir:   "examples/example-gateway/idl/github.com/uber/zanzibar/endpoints",
		TypeFileRootDir: "examples/example-gateway/gen-code/uber/zanzibar/endpoints/bar",
		TargetGenDir:    endpointDir,
	}
	importPath := "github.com/uber/zanzibar/examples/example-gateway/idl/github.com/uber/"
	p, err := build.Default.Import(importPath, "", build.FindOnly)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create build path for endpoint generation.")
	}
	thrift := filepath.Join(p.Dir, "zanzibar/endpoints/bar/bar.thrift")

	m, err := NewModuleSpec(thrift, h)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse thrift file.")
	}
	if len(m.Services) == 0 {
		return nil, errors.Errorf("No services build from %s", thrift)
	}
	return m, nil
}

// GenerateHandler builds the generated code in endpointDir for a handler from a spec and template.
func GenerateHandler(method *MethodSpec, tmpl *template.Template, endpointDir string) {
	// MethodSpec containes the handler name as endpoint.handler
	endpointName := strings.Split(method.EndpointName, ".")[0]
	handlerName := strings.Split(method.Name, ".")[0]
	dest := endpointDir + string(os.PathSeparator) + strings.ToLower(handlerName) + "_handler.go"
	file, err := os.Create(dest)
	if err != nil {
		fmt.Printf("Could not create %s: %s (skip)\n", dest, err)
		return
	}

	// TODO(sindelar): Use an endpoint to client map instead of proxy naming.
	downstreamService := endpointName
	downstreamMethod := handlerName

	vals := map[string]string{
		"MyHandler":         handlerName,
		"Package":           endpointName,
		"DownstreamService": downstreamService,
		"DownstreamMethod":  downstreamMethod,
	}
	tmpl.ExecuteTemplate(file, "endpoint_template.tmpl", vals)

	file.Close()
	return
}

// GenerateTestCase builds the generated test and benchmarking code in endpointDir for a handler from a spec and template.
func GenerateTestCase(method *MethodSpec, tmpl *template.Template, endpointDir string) {
	endpointName := strings.Split(method.EndpointName, ".")[0]
	handlerName := strings.Split(method.Name, ".")[0]
	dest := endpointDir + string(os.PathSeparator) + strings.ToLower(handlerName) + "_handler_test.go"
	file, err := os.Create(dest)
	if err != nil {
		fmt.Printf("Could not create %s: %s (skip)\n", dest, err)
		return
	}

	// TODO(sindelar): Use an endpoint to client map instead of proxy naming.
	downstreamService := endpointName
	downstreamMethod := handlerName

	// TODO(sindelar): Dummy data, read from golden file.
	var clientResponse = "{\\\"statusCode\\\":200}"
	var endpointPath = "/googlenow/add-credentials"
	var endpointHTTPMethod = "POST"
	var clientPath = "/add-credentials"
	var clientHTTPMethod = "POST"
	var endpointRequest = "{\\\"testrequest\\\"}"

	vals := map[string]string{
		"MyHandler":          handlerName,
		"Package":            endpointName,
		"DownstreamService":  downstreamService,
		"DownstreamMethod":   downstreamMethod,
		"EndpointPath":       endpointPath,
		"EndpointHttpMethod": endpointHTTPMethod,
		"ClientPath":         clientPath,
		"ClientHttpMethod":   clientHTTPMethod,
		"ClientResponse":     clientResponse,
		"EndpointRequest":    endpointRequest,
	}
	tmpl.Execute(file, vals)

	file.Close()
	return
}
