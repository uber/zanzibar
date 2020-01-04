// Copyright (c) 2020 Uber Technologies, Inc.
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

package util

import (
	"path/filepath"
	"runtime"
)

// DefaultMainFile returns the expected main.go file path for an example
// service inside the example-gateway
func DefaultMainFile(serviceName string) string {
	return filepath.Join(
		getDirName(),
		"..", "..", "..", "examples", "example-gateway",
		"build", "services", serviceName, "main", "main.go",
	)
}

// DefaultConfigFiles returns the typically expected default config files that
// a service would use. This includes root production.yaml and the service's
// own production.yaml
func DefaultConfigFiles(serviceName string) []string {
	return []string{
		filepath.Join(
			getDirName(),
			"..", "..", "..", "examples", "example-gateway", "config",
			"test.yaml",
		),
		filepath.Join(
			getDirName(),
			"..", "..", "..", "examples", "example-gateway", "config",
			"example-gateway", "test.yaml",
		),
	}
}

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(file)
}
