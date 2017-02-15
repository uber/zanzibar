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
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	tmpl "text/template"

	"github.com/pkg/errors"
)

// EndpointFiles are group of files generated for an endpoint.
type EndpointFiles struct {
	HandlerFiles []string
	StructFile   string
}

// ClientFiles are group of files generated for a client.
type ClientFiles struct {
	ClientFile string
	StructFile string
}

// EndpointMeta saves meta data used to render an endpoint.
type EndpointMeta struct {
	PackageName      string
	IncludedPackages []string
	Method           *MethodSpec
}

// EndpointTestMeta saves meta data used to render an endpoint test.
type EndpointTestMeta struct {
	PackageName string
	Method      *MethodSpec
}

var camelingRegex = regexp.MustCompile("[0-9A-Za-z]+")
var funcMap = tmpl.FuncMap{
	"title":        strings.Title,
	"Title":        strings.Title,
	"fullTypeName": fullTypeName,
	"statusCodes":  statusCodes,
	"camel":        camelCase,
}

func fullTypeName(typeName, packageName string) string {
	if typeName == "" || strings.Contains(typeName, ".") {
		return typeName
	}
	return packageName + "." + typeName
}

func statusCodes(codes []StatusCode) string {
	if len(codes) == 0 {
		return "[]int{}"
	}
	buf := bytes.NewBufferString("[]int{")
	for i := 0; i < len(codes)-1; i++ {
		if _, err := buf.WriteString(strconv.Itoa(codes[i].Code) + ","); err != nil {
			return err.Error()
		}
	}
	if _, err := buf.WriteString(strconv.Itoa(codes[len(codes)-1].Code) + "}"); err != nil {
		return err.Error()
	}
	return string(buf.Bytes())
}

func camelCase(src string) string {
	byteSrc := []byte(src)
	chunks := camelingRegex.FindAll(byteSrc, -1)
	for idx, val := range chunks {
		if idx > 0 {
			chunks[idx] = bytes.Title(val)
		} else {
			chunks[idx][0] = bytes.ToLower(val[0:1])[0]
		}
	}
	return string(bytes.Join(chunks, nil))
}

// Template generates code for edge gateway clients and edgegateway endpoints.
type Template struct {
	template *tmpl.Template
}

// NewTemplate creates a bundle of templates.
func NewTemplate(templatePattern string) (*Template, error) {
	t, err := tmpl.New("main").Funcs(funcMap).ParseGlob(templatePattern)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse template files")
	}
	return &Template{
		template: t,
	}, nil
}

// GenerateClientFile generates Go http code for services defined in thrift file.
// It returns the path of generated client file and struct file or an error.
func (t *Template) GenerateClientFile(thrift string, h *PackageHelper) (*ClientFiles, error) {
	m, err := NewModuleSpec(thrift, h)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse thrift file:")
	}
	if len(m.Services) == 0 {
		return nil, nil
	}

	m.PackageName = m.PackageName + "Client"
	err = t.execTemplateAndFmt("http_client.tmpl", m.GoClientFilePath, m)
	if err != nil {
		return nil, err
	}

	err = t.execTemplateAndFmt("structs.tmpl", m.GoStructsFilePath, m)
	if err != nil {
		return nil, err
	}

	return &ClientFiles{
		ClientFile: m.GoClientFilePath,
		StructFile: m.GoStructsFilePath,
	}, nil
}

// GenerateEndpointFile generates Go code for an zanzibar endpoint defined in
// thrift file. It returns the path of generated method files, struct file or
// an error.
func (t *Template) GenerateEndpointFile(thrift string, h *PackageHelper) (*EndpointFiles, error) {
	m, err := NewModuleSpec(thrift, h)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse thrift file:")
	}
	if len(m.Services) == 0 {
		return nil, nil
	}

	err = t.execTemplateAndFmt(
		"structs.tmpl", m.GoStructsFilePath, m,
	)
	if err != nil {
		return nil, err
	}

	endpointFiles := &EndpointFiles{
		HandlerFiles: make([]string, 0, len(m.Services[0].Methods)),
		StructFile:   m.GoStructsFilePath,
	}

	for _, service := range m.Services {
		for _, method := range service.Methods {
			dest, err := h.TargetEndpointPath(thrift, service.Name, method.Name)
			if err != nil {
				return nil, errors.Wrapf(err,
					"Could not generate endpoint path, service %s, method %s",
					service, method)
			}
			meta := &EndpointMeta{
				PackageName:      m.PackageName,
				IncludedPackages: m.IncludedPackages,
				Method:           method,
			}
			err = t.execTemplateAndFmt("endpoint.tmpl", dest, meta)
			if err != nil {
				return nil, err
			}
			endpointFiles.HandlerFiles = append(endpointFiles.HandlerFiles, dest)
		}
	}

	return endpointFiles, nil
}

// GenerateEndpointTestFile generates Go code for testing an zanzibar endpoint
// defined in a thrift file. It returns the path of generated test files,
// or an error.
func (t *Template) GenerateEndpointTestFile(thrift string, h *PackageHelper) ([]string, error) {
	m, err := NewModuleSpec(thrift, h)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse thrift file:")
	}
	if len(m.Services) == 0 {
		return nil, nil
	}

	testFiles := make([]string, 0, len(m.Services[0].Methods))
	for _, service := range m.Services {
		for _, method := range service.Methods {
			dest, err := h.TargetEndpointTestPath(thrift, service.Name, method.Name)
			if err != nil {
				return nil, errors.Wrapf(err,
					"Could not generate endpoint test path, service %s, method %s",
					service, method)
			}
			meta := &EndpointTestMeta{
				PackageName: m.PackageName,
				Method:      method,
			}
			err = t.execTemplateAndFmt("endpoint_test.tmpl", dest, meta)
			if err != nil {
				return nil, err
			}
			testFiles = append(testFiles, dest)
		}
	}

	return testFiles, nil
}

func (t *Template) execTemplateAndFmt(templName string, filePath string, data interface{}) error {
	file, err := openFileOrCreate(filePath)
	if err != nil {
		return errors.Wrapf(err, "failed to open file: ", err)
	}
	if err := t.template.ExecuteTemplate(file, templName, data); err != nil {
		return errors.Wrapf(err, "failed to execute template files for file %s", file)
	}

	gofmtCmd := exec.Command("gofmt", "-s", "-w", "-e", filePath)
	gofmtCmd.Stdout = os.Stdout
	gofmtCmd.Stderr = os.Stderr

	if err := gofmtCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to gofmt file: %s", filePath)
	}

	goimportsCmd := exec.Command("goimports", "-w", "-e", filePath)
	goimportsCmd.Stdout = os.Stdout
	goimportsCmd.Stderr = os.Stderr

	if err := goimportsCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to goimports file: %s", filePath)
	}

	if err := file.Close(); err != nil {
		return errors.Wrap(err, "failed to close file")
	}

	return nil
}

func openFileOrCreate(file string) (*os.File, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
}
