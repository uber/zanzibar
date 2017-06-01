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

package codegen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
	"go.uber.org/thriftrw/compile"
)

var thriftExtensionLength int = len(".thrift")

type dummyFS map[string][]byte

// Read returns the contents for the specified file.
func (fs dummyFS) Read(filename string) ([]byte, error) {
	if contents, ok := fs[filename]; ok {
		return contents, nil
	}
	return nil, os.ErrNotExist
}

// Abs returns the absolute path for the specified file.
// The dummy implementation always returns the original path.
func (dummyFS) Abs(filename string) (string, error) {
	return filename, nil
}

type naivePackageNameResolver struct {
}

func (r *naivePackageNameResolver) TypePackageName(
	thriftFile string,
) (string, error) {
	if thriftFile[0] == '.' {
		return "", errors.Errorf("Naive does not support relative imports")
	}

	_, fileName := filepath.Split(thriftFile)

	return fileName[0 : len(fileName)-thriftExtensionLength], nil
}

func newTypeConverter() *codegen.TypeConverter {
	return &codegen.TypeConverter{
		Helper: &naivePackageNameResolver{},
	}
}

func compileProgram(
	content string,
	otherFiles map[string][]byte,
) (*compile.Module, error) {
	if otherFiles == nil {
		otherFiles = map[string][]byte{}
	}
	otherFiles["structs.thrift"] = []byte(content)

	program, err := compile.Compile(
		"structs.thrift",
		compile.Filesystem(dummyFS(otherFiles)),
	)
	if err != nil {
		return nil, err
	}
	return program, nil
}

func convertTypes(
	fromStruct string,
	toStruct string,
	content string,
	otherFiles map[string][]byte,
	overrideMap map[string]codegen.FieldMapperEntry,
) (string, error) {
	converter := newTypeConverter()
	program, err := compileProgram(content, otherFiles)
	if err != nil {
		return "", err
	}
	overrideMap, err = addSpecToMap(overrideMap,
		program.Types[toStruct].(*compile.StructSpec).Fields,
		"")
	if err != nil {
		return "", err
	}

	err = converter.GenStructConverter(
		program.Types[fromStruct].(*compile.StructSpec).Fields,
		program.Types[toStruct].(*compile.StructSpec).Fields,
		overrideMap,
	)
	if err != nil {
		return "", err
	}

	return trim(strings.Join(converter.GetLines(), "\n")), nil
}

func addSpecToMap(
	overrideMap map[string]codegen.FieldMapperEntry,
	fields compile.FieldGroup,
	prefix string,
) (map[string]codegen.FieldMapperEntry, error) {
	for k, v := range overrideMap {
		for _, spec := range fields {
			fieldQualName := prefix + strings.Title(spec.Name)
			if v.QualifiedName == fieldQualName {
				v.Field = spec
				overrideMap[k] = v
			} else if strings.HasPrefix(v.QualifiedName, fieldQualName) {
				overrideMap, _ = addSpecToMap(
					overrideMap,
					spec.Type.(*compile.StructSpec).Fields,
					fieldQualName+".",
				)
			}
		}
	}
	return overrideMap, nil
}

func countTabs(line string) int {
	count := 0
	for i := 0; i < len(line); i++ {
		if line[i] == '\t' {
			count++
		} else {
			break
		}
	}
	return count
}

func trim(text string) string {
	lines := strings.Split(text, "\n")
	tabs := -1
	newLines := []string{}

	for _, line := range lines {
		if line == "" {
			continue
		}

		if tabs == -1 {
			tabs = countTabs(line)
		}

		if tabs >= len(line) {
			newLines = append(newLines, "")
		} else {
			newLines = append(newLines, line[tabs:])
		}
	}

	newText := strings.Join(newLines, "\n")
	return strings.TrimSpace(newText)
}

func TestConverStrings(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional string one
			2: required string two
		}

		struct Bar {
			1: optional string one
			2: required string two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = (*string)(in.One)
		out.Two = string(in.Two)
	`), lines)
}

func TestConvertBools(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional bool one
			2: required bool two
		}

		struct Bar {
			1: optional bool one
			2: required bool two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = (*bool)(in.One)
		out.Two = bool(in.Two)
	`), lines)
}

func TestConvertInt8(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional i8 one
			2: required i8 two
		}

		struct Bar {
			1: optional i8 one
			2: required i8 two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = (*int8)(in.One)
		out.Two = int8(in.Two)
	`), lines)
}

func TestConvertInt16(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional i16 one
			2: required i16 two
		}

		struct Bar {
			1: optional i16 one
			2: required i16 two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = (*int16)(in.One)
		out.Two = int16(in.Two)
	`), lines)
}

func TestConvertInt32(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional i32 one
			2: required i32 two
		}

		struct Bar {
			1: optional i32 one
			2: required i32 two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = (*int32)(in.One)
		out.Two = int32(in.Two)
	`), lines)
}

func TestConvertInt64(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional i64 one
			2: required i64 two
		}

		struct Bar {
			1: optional i64 one
			2: required i64 two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = (*int64)(in.One)
		out.Two = int64(in.Two)
	`), lines)
}

func TestConvertDouble(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional double one
			2: required double two
		}

		struct Bar {
			1: optional double one
			2: required double two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = (*float64)(in.One)
		out.Two = float64(in.Two)
	`), lines)
}

func TestConvertBinary(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional binary one
			2: required binary two
		}

		struct Bar {
			1: optional binary one
			2: required binary two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = []byte(in.One)
		out.Two = []byte(in.Two)
	`), lines)
}

func TestConvertStruct(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
			1: required string one
			2: optional string two
		}

		struct NestedBar {
			1: required string one
			2: optional string two
		}

		struct Foo {
			3: optional NestedFoo three
			4: required NestedFoo four
		}

		struct Bar {
			3: optional NestedBar three
			4: required NestedBar four
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		if in.Three != nil {
			out.Three = &structs.NestedBar{}
			out.Three.One = string(in.Three.One)
			out.Three.Two = (*string)(in.Three.Two)
		} else {
			out.Three = nil
		}
		if in.Four != nil {
			out.Four = &structs.NestedBar{}
			out.Four.One = string(in.Four.One)
			out.Four.Two = (*string)(in.Four.Two)
		} else {
			out.Four = nil
		}
	`), lines)
}

func TestHandlesMissingFields(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
			1: required string one
		}

		struct NestedBar {
			1: required string one
			2: optional string two
		}

		struct Foo {
			3: optional NestedFoo three
			4: required NestedFoo four
		}

		struct Bar {
			3: optional NestedBar three
			4: required NestedBar four
		}`,
		nil,
		nil,
	)

	assert.Equal(t, "cannot map by name for the field two", err.Error())
	assert.Equal(t, "", lines)
}

func TestStructTypeMisMatch(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
			1: required string one
			2: optional string two
		}

		struct NestedBar {
			1: required string one
			2: optional string two
		}

		struct Foo {
			3: optional NestedFoo three
			4: required string four
		}

		struct Bar {
			3: optional NestedBar three
			4: required NestedBar four
		}`,
		nil,
		nil,
	)

	assert.Equal(t, "", lines)
	assert.Equal(t, "could not convert struct fields, "+
		"incompatible type for four :", err.Error())
}

func TestConvertTypeDef(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`typedef string UUID

		struct Foo {
			1: optional UUID one
			2: required UUID two
		}

		struct Bar {
			1: optional UUID one
			2: required UUID two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = (*structs.UUID)(in.One)
		out.Two = structs.UUID(in.Two)
	`), lines)
}

func TestConvertEnum(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`enum ItemState {
			REQUIRED,
			OPTIONAL
		}

		struct Foo {
			1: optional ItemState one
			2: required ItemState two
		}

		struct Bar {
			1: optional ItemState one
			2: required ItemState two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = (*structs.ItemState)(in.One)
		out.Two = structs.ItemState(in.Two)
	`), lines)
}

func TestConvertWithBadImportTypedef(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		include "../../bar.thrift"
		
		struct Foo {
			1: optional bar.MyString one
			2: required string two
		}

		struct Bar {
			1: optional bar.MyString one
			2: required string two
		}`,
		map[string][]byte{
			"../../bar.thrift": []byte(`
			typedef string MyString
			`),
		},
		nil,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Contains(t,
		err.Error(),
		"could not lookup fieldType when building converter for MyString",
	)
}

func TestConvertWithBadImportEnum(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		include "../../bar.thrift"
		
		struct Foo {
			1: optional bar.MyEnum one
			2: required string two
		}

		struct Bar {
			1: optional bar.MyEnum one
			2: required string two
		}`,
		map[string][]byte{
			"../../bar.thrift": []byte(`
			enum MyEnum {
				REQUIRED,
				OPTIONAL
			}
			`),
		},
		nil,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Contains(t,
		err.Error(),
		"failed to get package for custom type (*compile.EnumSpec)",
	)
}

func TestConvertWithBadImportStruct(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		include "../../bar.thrift"
		
		struct Foo {
			1: optional bar.MyStruct one
			2: required string two
		}

		struct Bar {
			1: optional bar.MyStruct one
			2: required string two
		}`,
		map[string][]byte{
			"../../bar.thrift": []byte(`
			struct MyStruct {
				1: optional string one
			}
			`),
		},
		nil,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Contains(t,
		err.Error(),
		"could not lookup fieldType when building converter for MyStruct",
	)
}

func TestConvertListOfString(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional list<string> one
			2: required list<string> two
		}

		struct Bar {
			1: optional list<string> one
			2: required list<string> two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = make([]string, len(in.One))
		for index, value := range in.One {
			out.One[index] = string(value)
		}
		out.Two = make([]string, len(in.Two))
		for index, value := range in.Two {
			out.Two[index] = string(value)
		}
	`), lines)
}

func TestConvertListOfBinary(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional list<binary> one
			2: required list<binary> two
		}

		struct Bar {
			1: optional list<binary> one
			2: required list<binary> two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = make([][]byte, len(in.One))
		for index, value := range in.One {
			out.One[index] = []byte(value)
		}
		out.Two = make([][]byte, len(in.Two))
		for index, value := range in.Two {
			out.Two[index] = []byte(value)
		}
	`), lines)
}

func TestConvertListOfStruct(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		struct Inner {
			1: optional string field
		}
		
		struct Foo {
			1: optional list<Inner> one
			2: required list<Inner> two
		}

		struct Bar {
			1: optional list<Inner> one
			2: required list<Inner> two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = make([]*structs.Inner, len(in.One))
		for index, value := range in.One {
			if value != nil {
				out.One[index] = &structs.Inner{}
				out.One[index].Field = (*string)(in.One[index].Field)
			} else {
				out.One[index] = nil
			}
		}
		out.Two = make([]*structs.Inner, len(in.Two))
		for index, value := range in.Two {
			if value != nil {
				out.Two[index] = &structs.Inner{}
				out.Two[index].Field = (*string)(in.Two[index].Field)
			} else {
				out.Two[index] = nil
			}
		}
	`), lines)
}

func TestConvertWithBadImportListOfStruct(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		include "../../bar.thrift"
		
		struct Foo {
			1: optional list<bar.MyStruct> one
			2: required string two
		}

		struct Bar {
			1: optional list<bar.MyStruct> one
			2: required string two
		}`,
		map[string][]byte{
			"../../bar.thrift": []byte(`
			struct MyStruct {
				1: optional string one
			}
			`),
		},
		nil,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Contains(t,
		err.Error(),
		"failed to get package for custom type (*compile.StructSpec)",
	)
}

func TestConvertWithMisMatchListTypes(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		struct Inner {
			1: optional string field
		}
		
		struct Foo {
			1: optional list<Inner> one
			2: required string two
		}

		struct Bar {
			1: optional list<Inner> one
			2: required list<Inner> two
		}`,
		nil,
		nil,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Equal(t,
		"Could not convert field (two): type is not list",
		err.Error(),
	)
}

func TestConvertWithBadImportListOfBadStruct(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		include "../../bar.thrift"

		struct Inner {
			1: optional bar.MyStruct field
		}
		
		struct Foo {
			1: optional list<Inner> one
			2: required string two
		}

		struct Bar {
			1: optional list<Inner> one
			2: required string two
		}`,
		map[string][]byte{
			"../../bar.thrift": []byte(`
			struct MyStruct {
				1: optional string one
			}
			`),
		},
		nil,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Contains(t,
		err.Error(),
		"could not lookup fieldType when building converter for MyStruct",
	)
}

func TestConvertMapOfString(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional map<string, string> one
			2: required map<string, string> two
		}

		struct Bar {
			1: optional map<string, string> one
			2: required map<string, string> two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = make(map[string]string, len(in.One))
		for key, value := range in.One {
			out.One[key] = string(value)
		}
		out.Two = make(map[string]string, len(in.Two))
		for key, value := range in.Two {
			out.Two[key] = string(value)
		}
	`), lines)
}

func TestConvertMapOfStruct(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		struct Inner {
			1: optional string field
		}
		
		struct Foo {
			1: optional map<string, Inner> one
			2: required map<string, Inner> two
		}

		struct Bar {
			1: optional map<string, Inner> one
			2: required map<string, Inner> two
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = make(map[string]*structs.Inner, len(in.One))
		for key, value := range in.One {
			if value != nil {
				out.One[key] = &structs.Inner{}
				out.One[key].Field = (*string)(in.One[key].Field)
			} else {
				out.One[key] = nil
			}
		}
		out.Two = make(map[string]*structs.Inner, len(in.Two))
		for key, value := range in.Two {
			if value != nil {
				out.Two[key] = &structs.Inner{}
				out.Two[key].Field = (*string)(in.Two[key].Field)
			} else {
				out.Two[key] = nil
			}
		}
	`), lines)
}

func TestConvertWithBadImportMapOfStruct(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		include "../../bar.thrift"
		
		struct Foo {
			1: optional map<string, bar.MyStruct> one
			2: required string two
		}

		struct Bar {
			1: optional map<string, bar.MyStruct> one
			2: required string two
		}`,
		map[string][]byte{
			"../../bar.thrift": []byte(`
			struct MyStruct {
				1: optional string one
			}
			`),
		},
		nil,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Contains(t,
		err.Error(),
		"failed to get package for custom type (*compile.StructSpec)",
	)
}

func TestConvertWithMisMatchMapTypes(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		struct Inner {
			1: optional string field
		}
		
		struct Foo {
			1: optional map<string, Inner> one
			2: required string two
		}

		struct Bar {
			1: optional map<string, Inner> one
			2: required map<string, Inner> two
		}`,
		nil,
		nil,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Equal(t,
		"Could not convert field (two): type is not map",
		err.Error(),
	)
}

func TestConvertWithBadImportMapOfBadStruct(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		include "../../bar.thrift"

		struct Inner {
			1: optional bar.MyStruct field
		}
		
		struct Foo {
			1: optional map<string, Inner> one
			2: required string two
		}

		struct Bar {
			1: optional map<string, Inner> one
			2: required string two
		}`,
		map[string][]byte{
			"../../bar.thrift": []byte(`
			struct MyStruct {
				1: optional string one
			}
			`),
		},
		nil,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Contains(t,
		err.Error(),
		"could not lookup fieldType when building converter for MyStruct",
	)
}

func TestConvertWithBadKeyMapOfString(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional map<i32, string> one
			2: required map<i32, string> two
		}

		struct Bar {
			1: optional map<i32, string> one
			2: required map<i32, string> two
		}`,
		nil,
		nil,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Equal(t,
		"could not convert key (one), map is not string-keyed.",
		err.Error(),
	)
}
<<<<<<< HEAD
<<<<<<< HEAD

// Enduse that common acronyms are handled consistently with the
// Thrift compiled acronym strings.
func TestConvertStructWithAcoronymTypes(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
			1: required string uuid
			2: optional string two
		}

		struct NestedBar {
			1: required string uuid
			2: optional string two
		}

		struct Foo {
			3: optional NestedFoo three
			4: required NestedFoo four
		}

		struct Bar {
			3: optional NestedBar three
			4: required NestedBar four
		}`,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		if in.Three != nil {
			out.Three = &structs.NestedBar{}
			out.Three.UUID = string(in.Three.UUID)
			out.Three.Two = (*string)(in.Three.Two)
		} else {
			out.Three = nil
		}
		if in.Four != nil {
			out.Four = &structs.NestedBar{}
			out.Four.UUID = string(in.Four.UUID)
			out.Four.Two = (*string)(in.Four.Two)
		} else {
			out.Four = nil
		}
	`), lines)
=======
=======

>>>>>>> 323165ba... add helper method for tertiary assignment
func TestConverterMap(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["One"] = codegen.FieldMapperEntry{
		QualifiedName: "Two",
		Override:      true,
	}
	fieldMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "One",
		Override:      true,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: required bool one
			2: required bool two
		}

		struct Bar {
			1: required bool one
			2: required bool two
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = bool(in.Two)
		out.Two = bool(in.One)
	`), lines)
}

func TestConverterMapOverrideOptional(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["One"] = codegen.FieldMapperEntry{
		QualifiedName: "Two",
		Override:      true,
	} // Map from required filed, unconditional assignment
	fieldMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "One",
		Override:      true,
	}
	fieldMap["Three"] = codegen.FieldMapperEntry{
		QualifiedName: "Four",
		Override:      true,
	}
	fieldMap["Four"] = codegen.FieldMapperEntry{
		QualifiedName: "Three",
		Override:      false,
	} // Map to required filed, unconditional assignment

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional bool one
			2: optional bool two
			3: optional bool three
			4: required bool four
		}

		struct Bar {
			1: optional bool one
			2: required bool two
			3: optional bool three
			4: optional bool four
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = (*bool)(in.Two)
		out.Two = bool(in.Two)
		if in.One != nil {
			out.Two = bool(in.One)
		}
		out.Three = (*bool)(in.Three)
		if in.Four != nil {
			out.Three = (*bool)(in.Four)
		}
		out.Four = (*bool)(in.Four)
		`), lines)
}

func TestConverterMapStructWithSubFields(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Three.One"] = codegen.FieldMapperEntry{
		QualifiedName: "Three.Two",
		Override:      true,
	}
	fieldMap["Four.Two"] = codegen.FieldMapperEntry{
		QualifiedName: "Four.One",
		Override:      true,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
			1: required string one
			2: optional string two
		}

		struct NestedBar {
			1: required string one
			2: optional string two
		}

		struct Foo {
			3: optional NestedFoo three
			4: required NestedFoo four
		}

		struct Bar {
			3: optional NestedBar three
			4: required NestedBar four
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		if in.Three != nil {
			out.Three = &structs.NestedBar{}
			out.Three.One = string(in.Three.One)
			if in.Three.Two != nil {
				out.Three.One = string(in.Three.Two)
			}
			out.Three.Two = (*string)(in.Three.Two)
		} else {
			out.Three = nil
		}
		if in.Four != nil {
			out.Four = &structs.NestedBar{}
			out.Four.One = string(in.Four.One)
			out.Four.Two = (*string)(in.Four.One)
		} else {
			out.Four = nil
		}
	`), lines)
}

<<<<<<< HEAD
func TestConverterMapOverride(t *testing.T) {
>>>>>>> 3c1579fb... Add transforms with overrride
=======
func TestConverterMapListType(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["One"] = codegen.FieldMapperEntry{
		QualifiedName: "Two",
		Override:      true,
	}
	fieldMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "One",
		Override:      true,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional list<string> one
			2: required list<string> two
		}

		struct Bar {
			1: optional list<string> one
			2: required list<string> two
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = make([]string, len(in.Two))
		for index, value := range in.Two {
			out.One[index] = string(value)
		}
		sourceList := in.Two
		if in.One != nil {
			sourceList = in.One
		}
		out.Two = make([]string, len(sourceList))
		for index, value := range sourceList {
			out.Two[index] = string(value)
		}
	`), lines)
}

func TestConverterMapListOfStructType(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["One"] = codegen.FieldMapperEntry{
		QualifiedName: "Two",
		Override:      true,
	}
	fieldMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "One",
		Override:      true,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Inner {
			1: optional string field
		}

		struct Foo {
			1: optional list<Inner> one
			2: required list<Inner> two
		}

		struct Bar {
			1: optional list<Inner> one
			2: required list<Inner> two
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = make([]*structs.Inner, len(in.Two))
		for index, value := range in.Two {
			if value != nil {
				out.One[index] = &structs.Inner{}
				out.One[index].Field = (*string)(in.Two[index].Field)
			} else {
				out.One[index] = nil
			}
		}
		sourceList := in.Two
		isOverridden := false
		if in.One != nil {
			sourceList = in.One
			isOverridden = true
		}
		out.Two = make([]*structs.Inner, len(sourceList))
		for index, value := range sourceList {
			if isOverridden {
				if value != nil {
					out.Two[index] = &structs.Inner{}
					out.Two[index].Field = (*string)(in.One[index].Field)
				} else {
					out.Two[index] = nil
				}
			} else {
				if value != nil {
					out.Two[index] = &structs.Inner{}
					out.Two[index].Field = (*string)(in.Two[index].Field)
				} else {
					out.Two[index] = nil
				}
			}
		}
	`), lines)
>>>>>>> 567902ee... Add tests, fix converter to conform to tests
}

func TestConverterMapMapType(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["One"] = codegen.FieldMapperEntry{
		QualifiedName: "Two",
		Override:      true,
	}
	fieldMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "One",
		Override:      true,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`
		struct Inner {
			1: optional string field
		}
		
		struct Foo {
			1: optional map<string, Inner> one
			2: optional map<string, Inner> two
		}

		struct Bar {
			1: optional map<string, Inner> one
			2: optional map<string, Inner> two
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		sourceList := in.One
		isOverridden := false
		if in.Two != nil {
			sourceList = in.Two
			isOverridden = true
		}
		out.One = make(map[string]*structs.Inner, len(sourceList))
		for key, value := range sourceList {
			if isOverridden {
				if value != nil {
					out.One[key] = &structs.Inner{}
					out.One[key].Field = (*string)(in.Two[key].Field)
				} else {
					out.One[key] = nil
				}
			} else {
				if value != nil {
					out.One[key] = &structs.Inner{}
					out.One[key].Field = (*string)(in.One[key].Field)
				} else {
					out.One[key] = nil
				}
			}
		}
		sourceList := in.Two
		isOverridden := false
		if in.One != nil {
			sourceList = in.One
			isOverridden = true
		}
		out.Two = make(map[string]*structs.Inner, len(sourceList))
		for key, value := range sourceList {
			if isOverridden {
				if value != nil {
					out.Two[key] = &structs.Inner{}
					out.Two[key].Field = (*string)(in.One[key].Field)
				} else {
					out.Two[key] = nil
				}
			} else {
				if value != nil {
					out.Two[key] = &structs.Inner{}
					out.Two[key].Field = (*string)(in.Two[key].Field)
				} else {
					out.Two[key] = nil
				}
			}
		}
	`), lines)

}
