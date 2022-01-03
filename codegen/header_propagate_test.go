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

func newPropagator(h codegen.PackageNameResolver) *codegen.HeaderPropagator {
	return codegen.NewHeaderPropagator(h)
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

func propagateHeaders(
	headers []string,
	toStruct string,
	content string,
	propagateMap map[string]codegen.FieldMapperEntry,
	pkgHelper codegen.PackageNameResolver,
) (string, error) {

	propagator := newPropagator(pkgHelper)
	program, err := compileProgram(content, map[string][]byte{})
	if err != nil {
		return "", err
	}
	err = propagator.Propagate(
		headers,
		program.Types[toStruct].(*compile.StructSpec).Fields,
		propagateMap,
	)
	if err != nil {
		return "", err
	}
	return trim(strings.Join(propagator.GetLines(), "\n")), nil
}
func TestErrGetIDName(t *testing.T) {
	propagateMap := make(map[string]codegen.FieldMapperEntry)
	propagateMap["One.N1.Content"] = codegen.FieldMapperEntry{
		QualifiedName: "content-type",
		Override:      true,
	}
	propagateMap["Two.N2.Auth"] = codegen.FieldMapperEntry{
		QualifiedName: "auth",
		Override:      false,
	}
	_, err := propagateHeaders(
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
		propagateMap,
		&errPackageNameResolver{},
	)
	assert.Error(t, err)
}

func TestDefault(t *testing.T) {
	propagateMap := make(map[string]codegen.FieldMapperEntry)
	propagateMap["One"] = codegen.FieldMapperEntry{
		QualifiedName: "content-type",
		Override:      true,
	}
	propagateMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "auth",
		Override:      true,
	}
	lines, err := propagateHeaders(
		[]string{"content-type", "auth"},
		"Bar",
		`struct Bar {
			1: required string one
			2: optional string two
		}`,
		propagateMap,
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

func TestTypeDefString(t *testing.T) {
	propagateMap := make(map[string]codegen.FieldMapperEntry)
	propagateMap["One"] = codegen.FieldMapperEntry{
		QualifiedName: "content-type",
		Override:      true,
	}
	propagateMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "auth",
		Override:      true,
	}
	propagateMap["Three"] = codegen.FieldMapperEntry{
		QualifiedName: "auth",
		Override:      true,
	}
	lines, err := propagateHeaders(
		[]string{"content-type", "auth"},
		"Bar",
		`
		typedef string UUID
		struct Bar {
			1: required UUID one
			2: optional string two
			3: optional UUID three
		}`,
		propagateMap,
		&naivePackageNameResolver{},
	)
	assert.NoError(t, err)
	s := `
		if key,ok := headers.Get("content-type"); ok{
			val := structs.UUID(key)
			in.One=val
		}
		if key,ok := headers.Get("auth"); ok{
			val := structs.UUID(key)
			in.Three=&val
		}
		if key,ok := headers.Get("auth"); ok{
			in.Two=&key
		}`
	assert.Equal(t, strip(s), strip(lines))
}

func TestMissingField(t *testing.T) {
	propagateMap := make(map[string]codegen.FieldMapperEntry)
	propagateMap["Four"] = codegen.FieldMapperEntry{
		QualifiedName: "content-type",
		Override:      true,
	}
	propagateMap["Two.N2"] = codegen.FieldMapperEntry{
		QualifiedName: "auth",
		Override:      false,
	}
	_, err := propagateHeaders(
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
		propagateMap,
		&naivePackageNameResolver{},
	)
	assert.Error(t, err)
}

func TestNested(t *testing.T) {
	propagateMap := make(map[string]codegen.FieldMapperEntry)
	propagateMap["One.N1.Content"] = codegen.FieldMapperEntry{
		QualifiedName: "content-type",
		Override:      true,
	}
	propagateMap["Two.N2.Auth"] = codegen.FieldMapperEntry{
		QualifiedName: "auth",
		Override:      false,
	}
	lines, err := propagateHeaders(
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
		propagateMap,
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
		in.Two.N2.Auth = key
	}`
	assert.Equal(t, strip(s), strip(lines))
}

func TestBytePanic(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()
	propagateMap := make(map[string]codegen.FieldMapperEntry)
	propagateMap["One"] = codegen.FieldMapperEntry{
		QualifiedName: "content-type",
		Override:      true,
	}
	_, _ = propagateHeaders(
		[]string{"content-type", "auth"},
		"Bar",
		`struct Bar {
			1: required byte one
		}`,
		propagateMap,
		&naivePackageNameResolver{},
	)
}

func TestPrimaryType(t *testing.T) {
	propagateMap := make(map[string]codegen.FieldMapperEntry)
	propagateMap["U1"] = codegen.FieldMapperEntry{
		QualifiedName: "x-string",
	}
	propagateMap["U2"] = codegen.FieldMapperEntry{
		QualifiedName: "x-string",
	}
	propagateMap["S1"] = codegen.FieldMapperEntry{
		QualifiedName: "x-string",
	}
	propagateMap["S2"] = codegen.FieldMapperEntry{
		QualifiedName: "x-string",
	}
	propagateMap["I1"] = codegen.FieldMapperEntry{
		QualifiedName: "x-int",
	}
	propagateMap["I2"] = codegen.FieldMapperEntry{
		QualifiedName: "x-int",
	}
	propagateMap["I3"] = codegen.FieldMapperEntry{
		QualifiedName: "x-int",
	}
	propagateMap["I4"] = codegen.FieldMapperEntry{
		QualifiedName: "x-int",
	}
	propagateMap["I5"] = codegen.FieldMapperEntry{
		QualifiedName: "x-int",
	}
	propagateMap["I6"] = codegen.FieldMapperEntry{
		QualifiedName: "x-int",
	}
	propagateMap["B1"] = codegen.FieldMapperEntry{
		QualifiedName: "x-bool",
	}
	propagateMap["B2"] = codegen.FieldMapperEntry{
		QualifiedName: "x-bool",
	}
	propagateMap["F1"] = codegen.FieldMapperEntry{
		QualifiedName: "x-float",
	}
	propagateMap["F2"] = codegen.FieldMapperEntry{
		QualifiedName: "x-float",
	}
	lines, err := propagateHeaders(
		[]string{"x-string", "x-int", "x-float", "x-bool"},
		"Bar",
		`
		typedef string UUID

		struct Bar {
			1: required UUID u1
			2: optional UUID u2
			3: required string s1
			4: optional string s2
			5: required i32	i1
			6: optional i32	i2
			7: required i64 i3
			8: optional i64 i4
			9: required bool b1
			10: optional bool b2
			11: required double f1
			12: optional double f2
			13: required i16 i5
			15: optional i16 i6
		}`,
		propagateMap,
		&naivePackageNameResolver{},
	)
	assert.NoError(t, err)
	s := `
		if key, ok := headers.Get("x-bool"); ok {
			if v, err := strconv.ParseBool(key); err == nil {
				in.B1=v
			}
		}
		if key, ok := headers.Get("x-bool"); ok {
			if v, err := strconv.ParseBool(key); err == nil {
				in.B2=&v
			}
		}
		if key, ok := headers.Get("x-float"); ok {
			if v, err := strconv.ParseFloat(key,64); err == nil {
				in.F1=v
			}
		}
		if key, ok := headers.Get("x-float"); ok {
			if v, err := strconv.ParseFloat(key,64); err == nil {
				in.F2=&v
			}
		}
		if key, ok := headers.Get("x-int"); ok {
			if v, err := strconv.ParseInt(key,10,32); err == nil {
				val:=int32(v)
				in.I1=val
			}
		}
		if key, ok := headers.Get("x-int"); ok {
			if v, err := strconv.ParseInt(key,10,32); err == nil {
				val:=int32(v)
				in.I2=&val
			}
		}
		if key, ok := headers.Get("x-int"); ok {
			if v, err := strconv.ParseInt(key,10,64); err == nil {
				in.I3=v
			}
		}
		if key, ok := headers.Get("x-int"); ok {
			if v, err := strconv.ParseInt(key,10,64); err == nil {
				in.I4=&v
			}
		}
		if key, ok := headers.Get("x-int"); ok {
			if v, err := strconv.ParseInt(key,10,16); err == nil {
				val:=int16(v)
				in.I5=val
			}
		}
		if key, ok := headers.Get("x-int"); ok {
			if v, err := strconv.ParseInt(key,10,16); err == nil {
				val:=int16(v)
				in.I6=&val
			}
		}
		if key, ok := headers.Get("x-string"); ok {
			in.S1=key
		}
		if key, ok := headers.Get("x-string"); ok {
			in.S2=&key
		}
		if key, ok := headers.Get("x-string"); ok {
			val:=structs.UUID(key)
			in.U1=val
		}
		if key, ok := headers.Get("x-string"); ok {
			val:=structs.UUID(key)
			in.U2=&val
		}`
	assert.Equal(t, strip(s), strip(lines))
}
