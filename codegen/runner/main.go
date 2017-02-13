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

	"github.com/uber/zanzibar/codegen"
)

func main() {
	// Hacky way to find the template directory. Switch to use zw's template helpers
	// once they are landed.
	importPath := "github.com/uber/zanzibar/codegen"
	p, err := build.Default.Import(importPath, "", build.FindOnly)
	if err != nil {
		panic(fmt.Sprintf("Could not create build path for endpoint generation: %s", err))
	}

	clientTemplatePath := filepath.Join(p.Dir, "templates/*.tmpl")
	clientTemplate, err := codegen.NewTemplate(clientTemplatePath)
	if err != nil {
		fmt.Printf("Could not create template %s: %s (skip)\n", clientTemplatePath, err)
	}

	h, err := codegen.NewPackageHelper(
		"examples/example-gateway/idl",
		"examples/example-gateway/gen-code",
		"examples/example-gateway",
		"examples/example-gateway/idl/github.com/uber/zanzibar",
	)
	if err != nil {
		panic(err)
	}

	fail := false
	prefix := os.Args[1]
	// Iterate over all passed in endpoints.
	for i := 2; i < len(os.Args); i++ {
		fileParts := strings.Split(os.Args[i], string(os.PathSeparator))
		endpoint := fileParts[len(fileParts)-2]
		endpointDir := prefix + string(os.PathSeparator) + strings.ToLower(endpoint)
		err = os.Mkdir(endpointDir, 0755)
		if err != nil {
			panic(err)
		}

		err = os.Mkdir("examples/example-gateway/gen-code/clients", 0755)
		if err != nil {
			panic(err)
		}

		// Hack: only do bar...
		_, err = clientTemplate.GenerateClientFile(
			filepath.Join(
				p.Dir, "..", "examples", "example-gateway",
				"idl", "github.com", "uber", "zanzibar", "clients",
				"bar", "bar.thrift",
			),
			h,
		)
		if err != nil {
			fmt.Printf(
				"Could not create client specs for %s: %s \n",
				os.Args[i], err,
			)
			fail = true
			continue
		}

		_, err = clientTemplate.GenerateHandlerFile(
			filepath.Join(
				p.Dir, "..", "examples", "example-gateway",
				"idl", "github.com", "uber", "zanzibar", "endpoints",
				"bar", "bar.thrift",
			),
			h,
			"bar",
		)
		if err != nil {
			fmt.Printf(
				"Could not create handler specs for %s: %s \n",
				os.Args[i], err,
			)
			fail = true
			continue
		}

		_, err = clientTemplate.GenerateHandlerTestFile(
			filepath.Join(
				p.Dir, "..", "examples", "example-gateway",
				"idl", "github.com", "uber", "zanzibar", "endpoints",
				"bar", "bar.thrift",
			),
			h,
			"bar",
		)
		if err != nil {
			fmt.Printf(
				"Could not create tests specs for %s: %s \n",
				os.Args[i], err,
			)
			fail = true
			continue
		}
	}

	if fail {
		os.Exit(1)
	}
}
