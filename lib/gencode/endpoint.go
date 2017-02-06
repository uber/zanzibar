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
		endpoint := os.Args[i]
		endpointDir := prefix + string(os.PathSeparator) + strings.ToLower(endpoint)
		os.Mkdir(endpointDir, 0755)
		// PLACEHOLDER, REPLACE WITH HANDLERS FROM GENERATED CODE.
		// Read all handlers for an endpoint and generate each.
		handlers := []string{"foo"}
		for j := 0; j < len(handlers); j++ {
			dest := endpointDir + string(os.PathSeparator) + strings.ToLower(handlers[j]) + "_handler.go"
			file, err := os.Create(dest)
			if err != nil {
				fmt.Printf("Could not create %s: %s (skip)\n", dest, err)
				continue
			}
			// TODO(sindelar): Use an endpoint to client map.
			downstreamService := endpoint
			downstreamMethod := os.Args[i]

			vals := map[string]string{
				"MyHandler":         handlers[j],
				"Package":           endpoint,
				"DownstreamService": downstreamService,
				"DownstreamMethod":  downstreamMethod,
			}
			handlerTemplate.ExecuteTemplate(file, "endpoint_template.tmpl", vals)

			file.Close()

			dest = endpointDir + string(os.PathSeparator) + strings.ToLower(handlers[j]) + "_handler_test.go"
			file, err = os.Create(dest)
			if err != nil {
				fmt.Printf("Could not create %s: %s (skip)\n", dest, err)
				continue
			}

			// TODO(sindelar): Read from golden file.
			var clientResponse = "{\"statusCode\":200}"
			var endpointPath = "/googlenow/add-credentials"
			var endpointHttpMethod = "POST"
			var clientPath = "/add-credentials"
			var clientHttpMethod = "POST"

			vals = map[string]string{
				"MyHandler":          handlers[j],
				"Package":            endpoint,
				"DownstreamService":  downstreamService,
				"DownstreamMethod":   downstreamMethod,
				"EndpointPath":       endpointPath,
				"EndpointHttpMethod": endpointHttpMethod,
				"ClientPath":         clientPath,
				"ClientHttpMethod":   clientHttpMethod,
				"ClientResponse":     clientResponse,
			}
			testCaseTemplate.Execute(file, vals)

			file.Close()
		}
	}
}
