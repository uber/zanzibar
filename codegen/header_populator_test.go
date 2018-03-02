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

package codegen_test

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
	"go.uber.org/thriftrw/compile"
)

type errPackageNameResolver struct{}

func (r *errPackageNameResolver) TypePackageName(string) (string, error) {
	return "", errors.Errorf("Naive does not support relative imports")
}

func newPopulator(h codegen.PackageNameResolver) *codegen.HeaderPopulator {
	return codegen.NewHeaderPopulator(h)
}

func strip(text string) string {
	lines := strings.Split(text, "\n")
	newLines := []string{}
	r := strings.NewReplacer("\t", "", " ", "")
	for _, line := range lines {
		if line == "" {
			continue
		}
		newLines = append(newLines, r.Replace(line))
	}
	return strings.Join(newLines, "\n")
}

func populateHeaders(
	headers []string,
	toStruct string,
	content string,
	populateMap map[string]codegen.FieldMapperEntry,
	pkgHelper codegen.PackageNameResolver,
) (string, error) {

	populator := newPopulator(pkgHelper)
	program, err := compileProgram(content, map[string][]byte{})
	if err != nil {
		return "", err
	}
	err = populator.Populate(
		headers,
		program.Types[toStruct].(*compile.StructSpec).Fields,
		populateMap,
	)
	if err != nil {
		return "", err
	}
	return trim(strings.Join(populator.GetLines(), "\n")), nil
}
func TestErrGetIDName(t *testing.T) {
	populateMap := make(map[string]codegen.FieldMapperEntry)
	populateMap["One.N1.Content"] = codegen.FieldMapperEntry{
		QualifiedName: "content-type",
		Override:      true,
	}
	populateMap["Two.N2.Auth"] = codegen.FieldMapperEntry{
		QualifiedName: "auth",
		Override:      false,
	}
	_, err := populateHeaders(
		[]string{"content-type", "auth"},
		"Bar",
		`
		struct Nested {
			1: required string auth
			2: optional string content
		}
		struct Wrap {
			1: required Nested n1
			2: optional Nested n2
		}
		struct Bar {
			1: required Wrap one
			2: optional Wrap two
		}`,
		populateMap,
		&errPackageNameResolver{},
	)
	assert.Error(t, err)
}

func TestDefault(t *testing.T) {
	populateMap := make(map[string]codegen.FieldMapperEntry)
	populateMap["One"] = codegen.FieldMapperEntry{
		QualifiedName: "content-type",
		Override:      true,
	}
	populateMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "auth",
		Override:      true,
	}
	lines, err := populateHeaders(
		[]string{"content-type", "auth"},
		"Bar",
		`struct Bar {
			1: required string one
			2: optional string two
		}`,
		populateMap,
		&naivePackageNameResolver{},
	)
	assert.NoError(t, err)
	s := `
		if key, ok := headers.Get("content-type"); ok {
			in.One = key
		}
		if key, ok := headers.Get("auth"); ok {
			in.Two = &key
		}`
	assert.Equal(t, strip(s), strip(lines))
}

func TestMissingField(t *testing.T) {
	populateMap := make(map[string]codegen.FieldMapperEntry)
	populateMap["Four"] = codegen.FieldMapperEntry{
		QualifiedName: "content-type",
		Override:      true,
	}
	populateMap["Two.N2"] = codegen.FieldMapperEntry{
		QualifiedName: "auth",
		Override:      false,
	}
	_, err := populateHeaders(
		[]string{"content-type", "auth"},
		"Bar",
		`
		struct Nested {
			1: required string auth
			2: optional string content
		}
		struct Wrap {
			1: required Nested n1
			2: optional Nested n2
		}
		struct Bar {
			1: required Wrap one
			2: optional Wrap two
		}`,
		populateMap,
		&naivePackageNameResolver{},
	)
	assert.Error(t, err)
}

func TestNested(t *testing.T) {
	populateMap := make(map[string]codegen.FieldMapperEntry)
	populateMap["One.N1.Content"] = codegen.FieldMapperEntry{
		QualifiedName: "content-type",
		Override:      true,
	}
	populateMap["Two.N2.Auth"] = codegen.FieldMapperEntry{
		QualifiedName: "auth",
		Override:      false,
	}
	lines, err := populateHeaders(
		[]string{"content-type", "auth"},
		"Bar",
		`
		struct Nested {
			1: required string auth
			2: optional string content
		}
		struct Wrap {
			1: required Nested n1
			2: optional Nested n2
		}
		struct Bar {
			1: required Wrap one
			2: optional Wrap two
		}`,
		populateMap,
		&naivePackageNameResolver{},
	)
	assert.NoError(t, err)
	s := `
	if key, ok := headers.Get("content-type"); ok {
		if in.One == nil {
			in.One = &structs.Wrap{}
		}
		if in.One.N1 == nil {
			in.One.N1 = &structs.Nested{}
		}
		in.One.N1.Content = &key
	}
	if key, ok := headers.Get("auth"); ok {
		if in.Two == nil {
			in.Two = &structs.Wrap{}
		}
		if in.Two.N2 == nil {
			in.Two.N2 = &structs.Nested{}
		}
		if in.Two.N2.Auth!="" {
			in.Two.N2.Auth = key
		}
	}`
	assert.Equal(t, strip(s), strip(lines))
}
