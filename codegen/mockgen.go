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

package codegen

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

const (
	exampleGatewayPkg = "github.com/uber/zanzibar/examples/example-gateway"
	mockgenPkg        = "github.com/golang/mock/mockgen"
	zanzibarPkg       = "github.com/uber/zanzibar"
)

// MockgenBin is a struct abstracts the mockgen binary built from mockgen package in vendor
type MockgenBin struct {
	Bin string
}

// NewMockgenBin builds the mockgen binary from vendor directory
func NewMockgenBin(projRoot string) (*MockgenBin, error) {
	goPath := os.Getenv("GOPATH")
	// we would not need this if example gateway is a fully standalone application
	if projRoot == exampleGatewayPkg {
		projRoot = zanzibarPkg
	}
	// we assume that the vendor directory is flattened as Glide does
	mockgenDir := path.Join(goPath, "src", projRoot, "vendor", mockgenPkg)
	if _, err := os.Stat(mockgenDir); err != nil {
		return nil, errors.Wrapf(
			err, "error finding mockgen package in the vendor dir: %q does not exist", mockgenDir,
		)
	}

	var mockgenBin = "mockgen.bin"
	if runtime.GOOS == "windows" {
		// Windows won't execute a program unless it has a ".exe" suffix.
		mockgenBin += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", mockgenBin, ".")
	cmd.Dir = mockgenDir

	var stderr, stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(
			err,
			"error running command %q in %s: %s",
			strings.Join(cmd.Args, " "),
			mockgenDir,
			stderr.String(),
		)
	}

	return &MockgenBin{
		Bin: path.Join(mockgenDir, mockgenBin),
	}, nil
}

// GenMock generates mocks for given module instance, dest is the file path relative to
// the instance's generated package dir, pkg is the package name of the generated mocks,
// and intf is the interface to generate mock for
func (m MockgenBin) GenMock(instance *ModuleInstance, dest, pkg, intf string) error {
	importPath := instance.PackageInfo.PackagePath
	if _, ok := instance.Config["thriftFile"]; ok {
		importPath = instance.PackageInfo.GeneratedPackagePath
	}

	genDir := path.Join(os.Getenv("GOPATH"), "src")
	genDir = path.Join(genDir, instance.PackageInfo.GeneratedPackagePath, path.Dir(dest))
	genDest := path.Join(genDir, path.Base(dest))
	if _, err := os.Stat(genDir); os.IsNotExist(err) {
		if err := os.MkdirAll(genDir, os.ModePerm); err != nil {
			return err
		}
	}

	cmd := exec.Command(m.Bin, "-destination", genDest, "-package", pkg, importPath, intf)
	cmd.Dir = genDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(
			err,
			"error running command %q in %s: %s",
			strings.Join(cmd.Args, " "),
			genDir,
			stderr.String(),
		)
	}

	return nil
}

// ClientMockGenHook returns a PostGenHook to generate client mocks
func ClientMockGenHook(h *PackageHelper) (PostGenHook, error) {
	bin, err := NewMockgenBin(h.PackageRoot())
	if err != nil {
		return nil, errors.Wrap(err, "error building mockgen binary")
	}

	return func(instances map[string][]*ModuleInstance) error {
		fmt.Println("Generating client mocks:")
		mockCount := len(instances["client"])
		for i, instance := range instances["client"] {
			if err := bin.GenMock(
				instance,
				"mock-client/mock_client.go",
				"clientmock",
				"Client",
			); err != nil {
				return errors.Wrapf(
					err,
					"error generating mocks for client %q",
					instance.InstanceName,
				)
			}
			fmt.Printf(
				"Generating %12s %12s %-20s in %-40s %d/%d\n",
				"mock",
				instance.ClassName,
				instance.InstanceName,
				path.Join(path.Base(h.CodeGenTargetPath()), instance.Directory),
				i+1, mockCount,
			)
		}
		return nil
	}, nil
}
