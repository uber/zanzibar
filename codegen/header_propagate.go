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

package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

// HeaderPropagator generates function propagates endpoint request
// headers to client request body
type HeaderPropagator struct {
	LineBuilder
	Helper PackageNameResolver
}

// NewHeaderPropagator returns an instance of HeaderPropagator
func NewHeaderPropagator(h PackageNameResolver) *HeaderPropagator {
	return &HeaderPropagator{
		LineBuilder: LineBuilder{},
		Helper:      h,
	}
}

// Propagate assigns header value to downstream client request fields
// based on fieldMap
func (hp *HeaderPropagator) Propagate(
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
		gotype, err := GoType(hp.Helper, field.Type)
		if err != nil {
			return errors.Errorf("invalid: trying to assign header %s to non-string field in %s",
				val.QualifiedName, field.Name)
		}

		hp.appendf(`if key, ok := headers.Get("%s"); ok {`, val.QualifiedName)
		// patch optional params along the path
		if err := hp.initNilOpt(key, toFields); err != nil {
			return err
		}

		arrs := typeSwitch(key, gotype, field)
		hp.append(arrs...)

		hp.append("}")
	}
	return nil
}

// typeSwitch supports primary type parsing for headers
func typeSwitch(key, gotype string, field *compile.FieldSpec) []string {
	var (
		ret       = []string{}
		typeParse string
		typeCast  string
		assignVal = "key"
	)
	switch gotype {
	case "int8":
		panic(fmt.Sprintf("type byte is note supported for field %q", field.Name))
	case "bool":
		typeParse = "strconv.ParseBool(key)"
		assignVal = "v"
	case "int16":
		typeParse = "strconv.ParseInt(key, 10, 16)"
		assignVal = "val"
		typeCast = "val := int16(v)\n"
	case "int32":
		typeParse = "strconv.ParseInt(key, 10, 32)"
		assignVal = "val"
		typeCast = "val := int32(v)\n"
	case "int64":
		typeParse = "strconv.ParseInt(key, 10, 64)"
		assignVal = "v"
	case "float64":
		typeParse = "strconv.ParseFloat(key, 64)"
		assignVal = "v"
	case "string":
	default:
		typeCast = "val := " + gotype + "(key)\n"
		assignVal = "val"
	}
	if len(typeParse) > 0 {
		ret = append(ret, fmt.Sprintf("if v, err := %s; err == nil {\n", typeParse))
	}
	if len(typeCast) > 0 {
		ret = append(ret, typeCast)
	}
	if !field.Required {
		ret = append(ret, fmt.Sprintf("in.%s = &%s\n", key, assignVal))
	} else {
		ret = append(ret, fmt.Sprintf("in.%s = %s\n", key, assignVal))
	}
	if len(typeParse) > 0 {
		ret = append(ret, "}\n")
	}
	return ret
}

// init optional field that could be nil on field assign path
func (hp *HeaderPropagator) initNilOpt(path string, toFields []*compile.FieldSpec) error {
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
		ftype := f.Type
		t, err := GoCustomType(hp.Helper, ftype)
		if err != nil {
			return errors.Wrapf(
				err,
				"could not lookup fieldType when building converter for %s",
				ftype.ThriftName(),
			)
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
