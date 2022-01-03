// Copyright (c) 2022 Uber Technologies, Inc.
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
	"bytes"
	"path/filepath"
	"runtime"
	"sync"
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
		filepath.Join(
			getDirName(),
			"..", "..", "..", "examples", "example-gateway", "build.yaml",
		),
	}
}

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(file)
}

// Buffer is a goroutine safe bytes.Buffer
type Buffer struct {
	buffer bytes.Buffer
	mutex  sync.Mutex
}

// Write appends the contents of p to the buffer, growing the buffer as needed. It returns
// the number of bytes written.
func (s *Buffer) Write(p []byte) (n int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.buffer.Write(p)
}

// String returns the contents of the unread portion of the buffer
// as a string.  If the Buffer is a nil pointer, it returns "<nil>".
func (s *Buffer) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.buffer.String()
}
