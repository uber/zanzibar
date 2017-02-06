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

package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/lib/gencode"
)

func main() {
	funcMap := template.FuncMap{
		"Title": strings.Title,
	}

	// Hacky way to find the template directory. Switch to use zw's template helpers
	// once they are landed.
	importPath := "github.com/uber/zanzibar/lib/gencode"
	p, err := build.Default.Import(importPath, "", build.FindOnly)
	if err != nil {
		panic(fmt.Sprintf("Could not create build path for endpoint generation: %s", err))
	}
	templatePath := filepath.Join(p.Dir, "templates/endpoint_template.tmpl")
	handlerTemplate, err := template.New("endpoint_template.tmpl").Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		fmt.Printf("Could not create template %s: %s (skip)\n", templatePath, err)
	}

	// Build a test file with benchmarking to validate the endpoint.
	// In the future, refactor this into a test runner instead of generating test
	// files for each endpoint.
	testCaseTemplatePath := filepath.Join(p.Dir, "templates/endpoint_test_template.tmpl")
	testCaseTemplate, err := template.New("endpoint_test_template.tmpl").Funcs(funcMap).ParseFiles(testCaseTemplatePath)
	if err != nil {
		fmt.Printf("Could not create template %s: %s (skip)\n", templatePath, err)
	}

	prefix := os.Args[1]
	// Iterate over all passed in endpoints.
	for i := 2; i < len(os.Args); i++ {
		//endpoint := os.Args[i]
		endpoint := "bar"
		endpointDir := prefix + string(os.PathSeparator) + strings.ToLower(endpoint)
		os.Mkdir(endpointDir, 0755)

		// Generate handlers and test files for each method
		m, err := BuildModuleSpecForEndpoint(endpointDir)
		if err != nil {
			fmt.Printf("Could not create endpoint specs for %s: %s \n", os.Args[i], err)
			continue
		}
		handlers := m.Services[0].Methods

		for j := 0; j < len(handlers); j++ {
			// TODO: Why is ModuleSpec creating methods for error response
			// like "missingArg"
			// fmt.Printf("Found %s %s \n", handlers[j].EndpointName, handlers[j].Name)
			if (handlers[j].Name == "bar") {
				GenerateHandler(handlers[j], handlerTemplate, endpointDir)
				GenerateTestCase(handlers[j], testCaseTemplate, endpointDir)
			}
		}
	}
}

func BuildModuleSpecForEndpoint(endpointDir string) (*gencode.ModuleSpec, error) {
	h := &gencode.PackageHelper{
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

	m, err := gencode.NewModuleSpec(thrift, h)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse thrift file.")
	}
	if len(m.Services) == 0 {
		return nil, errors.Errorf("No services build from %s", thrift)
	}
	return m, nil
}

func GenerateHandler(method *gencode.MethodSpec, tmpl *template.Template, endpointDir string) {
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

func GenerateTestCase(method *gencode.MethodSpec, tmpl *template.Template, endpointDir string) {
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

	// TODO(sindelar): Read from golden file.
	var clientResponse = "{\\\"statusCode\\\":200}"
	var endpointPath = "/googlenow/add-credentials"
	var endpointHttpMethod = "POST"
	var clientPath = "/add-credentials"
	var clientHttpMethod = "POST"
	var endpointRequest = "{\\\"testrequest\\\"}"

	vals := map[string]string{
		"MyHandler":          handlerName,
		"Package":            endpointName,
		"DownstreamService":  downstreamService,
		"DownstreamMethod":   downstreamMethod,
		"EndpointPath":       endpointPath,
		"EndpointHttpMethod": endpointHttpMethod,
		"ClientPath":         clientPath,
		"ClientHttpMethod":   clientHttpMethod,
		"ClientResponse":     clientResponse,
		"EndpointRequest":    endpointRequest,
	}
	tmpl.Execute(file, vals)

	file.Close()
	return
}
