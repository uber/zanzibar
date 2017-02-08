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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	tmpl "text/template"

	"github.com/pkg/errors"
)

var funcMap = tmpl.FuncMap{
	"title": strings.Title,
}

// Template generates code for edge gateway clients and edgegateway endpoints.
type Template struct {
	template *tmpl.Template
}

// NewTemplate creates a bundle of templates.
func NewTemplate(templatePattern string) (*Template, error) {
	t, err := tmpl.New("main").Funcs(funcMap).ParseGlob(templatePattern)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse tempalte files")
	}
	return &Template{
		template: t,
	}, nil
}

// GenerateClientFile generates Go http code for services defined in thrift file.
// It returns the path of generated file or error.
func (t *Template) GenerateClientFile(thrift string, h *PackageHelper) (string, error) {
	m, err := NewModuleSpec(thrift, h)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse thrift file:")
	}
	if len(m.Services) == 0 {
		return "", errors.Errorf("no service is found in thrift file %s", thrift)
	}
	file, err := openFileOrCreate(m.GoFilePath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open file: ", err)
	}
	if err := t.template.ExecuteTemplate(file, "http_client.tmpl", m); err != nil {
		return "", errors.Wrapf(err, "failed to execute tempalte files for thrift %s", thrift)
	}
	if err := exec.Command("gofmt", "-s", "-w", "-e", m.GoFilePath).Run(); err != nil {
		return "", errors.Wrapf(err, "failed to gofmt file: %s", m.GoFilePath)
	}
	if err := file.Close(); err != nil {
		return "", errors.Wrap(err, "failed to close file")
	}
	return m.GoFilePath, nil
}

func openFileOrCreate(file string) (*os.File, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
}
