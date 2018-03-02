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
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
	"sort"
)

// HeaderPopulator generates function populates endpoint request
// headers to client request body
type HeaderPopulator struct {
	LineBuilder
	Helper PackageNameResolver
}

// NewHeaderPopulator returns an instance of HeaderPopulator
func NewHeaderPopulator(h PackageNameResolver) *HeaderPopulator {
	return &HeaderPopulator{
		LineBuilder: LineBuilder{},
		Helper:      h,
	}
}

// Populate assigns header value to downstream client request fields
// based on fieldMap
func (hp *HeaderPopulator) Populate(
	headers []string,
	toFields []*compile.FieldSpec,
	fieldMap map[string]FieldMapperEntry,
) error {
	sortedKeys := make([]string, len(fieldMap))
	i := 0
	for key := range fieldMap {
		sortedKeys[i] = key
		i++
	}
	sort.Strings(sortedKeys)
	for _, key := range sortedKeys {
		val := fieldMap[key]
		field, err := findField(key, toFields)
		if err != nil {
			return err
		}
		if t, err := GoType(nil, field.Type); err != nil || t != "string" {
			return errors.Errorf("invalid: trying to assign header %s to non-string field in %s",
				val.QualifiedName, field.Name)
		}
		hp.appendf(`if key, ok := headers.Get("%s"); ok {`, val.QualifiedName)
		// patch optional params along the path
		if err := hp.initNilOpt(key, toFields); err != nil {
			return err
		}

		// TODO support primitive types
		// e.g: -H "version: 3.0" -> int64(3)
		if !field.Required {
			if !val.Override {
				hp.appendf("if in.%s != nil {", key)
			}
			hp.appendf(`in.%s = &key`, key)
			if !val.Override {
				hp.appendf("}")
			}
		} else {
			if !val.Override {
				hp.appendf(`if in.%s != "" {`, key)
			}
			hp.appendf(`in.%s = key`, key)
			if !val.Override {
				hp.appendf("}")
			}
		}
		hp.append("}")
	}
	return nil
}

// init optional field that could be nil on field assign path
func (hp *HeaderPopulator) initNilOpt(path string, toFields []*compile.FieldSpec) error {
	initChecks := getMiddleIdentifiers(path)
	if len(initChecks) < 2 {
		return nil
	}
	initChecks = initChecks[:len(initChecks)-1]
	for _, p := range initChecks {
		f, err := findField(p, toFields)
		if err != nil {
			return err
		}
		t, err := hp.getIdentifierName(f.Type)
		if err != nil {
			return err
		}
		hp.appendf("if in.%s == nil {", p)
		hp.appendf("in.%s = &%s{}", p, t)
		hp.append("}")
	}
	return nil
}

func findField(fieldPath string, toFields []*compile.FieldSpec) (*compile.FieldSpec, error) {
	currPath := strings.Split(fieldPath, ".")
	currFields := toFields
	missErr := errors.Errorf("could not find field path in client request %s", fieldPath)

	for len(currFields) > 0 && len(currPath) > 0 {
		prevPath := []string(currPath)
		currPos := currPath[0]
		for _, v := range currFields {
			if strings.ToLower(v.Name) == strings.ToLower(currPos) {
				if len(currPath) == 1 {
					return v, nil
				}
				currPath = currPath[1:]
				t := v.Type.(*compile.StructSpec)
				currFields = t.Fields
				break
			}
		}
		if len(prevPath) == len(currPath) {
			return nil, missErr
		}
	}
	return nil, missErr
}

func (hp *HeaderPopulator) getIdentifierName(fieldType compile.TypeSpec) (string, error) {
	t, err := goCustomType(hp.Helper, fieldType)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"could not lookup fieldType when building converter for %s",
			fieldType.ThriftName(),
		)
	}
	return t, nil
}
