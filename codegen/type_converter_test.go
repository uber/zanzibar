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

func TestConvertBoolsOptionalToRequired(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: required bool one
		}

		struct Bar {
			1: optional bool one
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = ptr.Bool(in.One)
	`), lines)
}

func TestConvertBoolsRequiredToOptional(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional bool one
		}

		struct Bar {
			1: required bool one
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		if in.One != nil {
			out.One = *(in.One)
		}
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

func TestConvertStructRequiredToOptional(t *testing.T) {
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
		}

		struct Bar {
			3: required NestedBar three
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
	`), lines)
}

func TestConvertStructOptionalToRequired(t *testing.T) {
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
			3: required NestedFoo three
		}

		struct Bar {
			3: optional NestedBar three
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
	assert.NoError(t, err)
	assert.Equal(t, trim(`
		if in.Three != nil {
			out.Three = &structs.NestedBar{}
			out.Three.One = string(in.Three.One)
		} else {
			out.Three = nil
		}
		if in.Four != nil {
			out.Four = &structs.NestedBar{}
			out.Four.One = string(in.Four.One)
		} else {
			out.Four = nil
		}`),
		lines)
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

// TODO this is badly broken,  ptr.StructsUUID is not a thing
func TestConvertTypeDefReqToOpt(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`typedef string UUID

		struct Foo {
			1: required UUID one
			2: optional UUID two
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
		out.One = ptr.StructsUUID(in.One)
		if in.Two != nil {
			out.Two = *(in.Two)
		}
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

func TestConvertWithCorrectKeyMapNotStringKey(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		typedef string UUID

		struct Foo {
			1: optional map<UUID, string> one
		}

		struct Bar {
			1: optional map<UUID, string> one
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = make(map[structs.UUID]string, len(in.One))
		for key, value := range in.One {
			out.One[ structs.UUID(key)] = string(value)
		}
	`), lines)
}

func TestConvertWithBadKeyMapMapKey(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		typedef string UUID

		struct Foo {
			1: optional map<map<string, string>, string> one
			2: required map<i32, string> two
		}

		struct Bar {
			1: optional map<map<string, string>, string> one
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

// Enduse that common acronyms are handled consistently with the
// Thrift compiled acronym strings.
func TestConvertStructWithAcoronymTypes(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
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
		fieldMap,
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
}

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

func TestConvertTypeDefReqToReqWithOptOverride(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "One",
		Override:      true,
	}
	lines, err := convertTypes(
		"Foo", "Bar",
		`typedef string UUID

		struct Foo {
			1: optional UUID one
			2: required UUID two
		}

		struct Bar {
			1: required UUID two
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.Two = structs.UUID(in.Two)
		if in.One != nil {
			out.Two = *(in.One)
		}
	`), lines)
}

func TestConvertTypeDefOptToReqWithOverride(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "One",
		Override:      false,
	}
	fieldMap["One"] = codegen.FieldMapperEntry{
		QualifiedName: "Three",
		Override:      false,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`typedef string UUID

		struct Foo {
			1: optional UUID one
			2: optional UUID two
			3: required UUID three
		}

		struct Bar {
			1: optional UUID one
			2: required UUID two
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = ptr.StructsUUID(in.Three)
		if in.One != nil {
			out.One = (*structs.UUID)(in.One)
		}
		if in.One != nil {
			out.Two = *(in.One)
		}
		if in.Two != nil {
			out.Two = *(in.Two)
		}
	`), lines)
}

func TestConverterMapNewField(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "One",
		Override:      true,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: required bool one
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
		out.One = bool(in.One)
		out.Two = bool(in.One)
	`), lines)
}

func TestConverterMapOptionalNoOverride(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["One"] = codegen.FieldMapperEntry{
		QualifiedName: "Two",
		Override:      false,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: optional bool one
			2: optional bool two
		}

		struct Bar {
			1: optional bool one
			2: optional bool two
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = (*bool)(in.Two)
		if in.One != nil {
			out.One = (*bool)(in.One)
		}
		out.Two = (*bool)(in.Two)
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
		out.One = (*bool)(in.One)
		if in.Two != nil {
			out.One = (*bool)(in.Two)
		}
		if in.Two != nil {
			out.Two = *(in.Two)
		}
		if in.One != nil {
			out.Two = *(in.One)
		}
		out.Three = ptr.Bool(in.Four)
		out.Four = ptr.Bool(in.Four)
		`), lines)
}

func TestConverterMapOverrideReqToOpt(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["One"] = codegen.FieldMapperEntry{
		QualifiedName: "Two",
		Override:      true,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: required bool one
			2: optional bool two
		}

		struct Bar {
			1: optional bool one
			2: optional bool two
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.One = ptr.Bool(in.One)
		if in.Two != nil {
			out.One = (*bool)(in.Two)
		}
		out.Two = (*bool)(in.Two)	
	`), lines)
}

func TestConverterMapStructWithSubFields2(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Three.One"] = codegen.FieldMapperEntry{
		QualifiedName: "Four.Two",
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
			if in.Four != nil && in.Four.Two != nil {
				out.Three.One = *(in.Four.Two)
			}
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

func TestConverterMapStructWithFromReqDropped(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required NestedFoo three
 			2: required NestedFoo four
 		}

 		struct Bar {
 			1: optional NestedFoo three
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		if in.Three != nil {
			out.Three = &structs.NestedFoo{}
			out.Three.One = string(in.Three.One)
			out.Three.Two = (*string)(in.Three.Two)
		} else {
			out.Three = nil
		}`),
		lines)
}

func TestConverterMapStructWithFromOptDropped(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required NestedFoo three
 			2: optional NestedFoo four
 		}

 		struct Bar {
 			1: optional NestedFoo three
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		if in.Three != nil {
			out.Three = &structs.NestedFoo{}
			out.Three.One = string(in.Three.One)
			out.Three.Two = (*string)(in.Three.Two)
		} else {
			out.Three = nil
		}`),
		lines)
}

func TestConverterMapStructWithToOptMissingOk(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	//fieldMap["Five.One"] = codegen.FieldMapperEntry{
	//	QualifiedName: "Four.Two",
	//	Override:      true,
	//}
	//fieldMap["Four.Two"] = codegen.FieldMapperEntry{
	//	QualifiedName: "Three.One",
	//	Override:      true,
	//}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required NestedFoo three
 		}

 		struct Bar {
 			1: required NestedFoo three
 			2: optional NestedFoo four
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		if in.Three != nil {
			out.Three = &structs.NestedFoo{}
			out.Three.One = string(in.Three.One)
			out.Three.Two = (*string)(in.Three.Two)
		} else {
			out.Three = nil
		}`),
		lines)
}

func TestConverterMapStructWithToReqMissingError(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	_, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required NestedFoo three
 		}

 		struct Bar {
 			1: required NestedFoo three
 			2: required NestedFoo four
 		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "required toField four does not have a valid fromField mapping")
}

func TestConverterMapStructWithFieldMapToReqField(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Four"] = codegen.FieldMapperEntry{
		QualifiedName: "Three",
		Override:      true,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required NestedFoo three
 		}

 		struct Bar {
 			1: required NestedFoo three
 			2: required NestedFoo four
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		if in.Three != nil {
			out.Three = &structs.NestedFoo{}
			out.Three.One = string(in.Three.One)
			out.Three.Two = (*string)(in.Three.Two)
		} else {
			out.Three = nil
		}
		if in.Three != nil {
			out.Four = &structs.NestedFoo{}
			if in.Three != nil {
				out.Four.One = string(in.Three.One)
			}
			if in.Three != nil {
				out.Four.Two = (*string)(in.Three.Two)
			}
		} else {
			out.Four = nil
		}`),
		lines)
}

func TestConverterMapStructWithFieldMap(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Five.One"] = codegen.FieldMapperEntry{
		QualifiedName: "Three.Two",
		Override:      true,
	}
	fieldMap["Five.Two"] = codegen.FieldMapperEntry{
		QualifiedName: "Four.One",
		Override:      true,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required NestedFoo three
 			2: required NestedFoo four
 		}

 		struct Bar {
 			1: required NestedFoo three
 			2: optional NestedFoo five
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		if in.Three != nil {
			out.Three = &structs.NestedFoo{}
			out.Three.One = string(in.Three.One)
			out.Three.Two = (*string)(in.Three.Two)
		} else {
			out.Three = nil
		}
		out.Five = &structs.NestedFoo{}
		if in.Three != nil && in.Three.Two != nil {
			out.Five.One = *(in.Three.Two)
		}
		if in.Four != nil {
			out.Five.Two = ptr.String(in.Four.One)
		}`),
		lines)
}

func TestConverterMapStructWithFieldMapErrorCase(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	//  Missing  Five.One transform
	fieldMap["Five.Two"] = codegen.FieldMapperEntry{
		QualifiedName: "Four.One",
		Override:      true,
	}

	_, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
			1: required string one
			2: optional string two
		}

		struct Foo {
			1: required NestedFoo three
			2: required NestedFoo four
		}

		struct Bar {
			1: required NestedFoo three
			2: optional NestedFoo five
		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)

	assert.Contains(t,
		err.Error(),
		"required toField one does not have a valid fromField mapping",
	)
}

func TestConverterMapStructWithFieldMapDeeper1(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Six.Four.One"] = codegen.FieldMapperEntry{
		QualifiedName: "Five.Three.One",
		Override:      true,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedA {
 			1: required string one
 		}

 		struct NestedB {
 			1: required NestedA three
		}

		struct NestedC {
			1: required NestedA four
		}

 		struct Foo {
 			1: required NestedB five
 		}

 		struct Bar {
 			1: required NestedC six
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.Six = &structs.NestedC{}
		out.Six.Four = &structs.NestedA{}
		if in.Five != nil && in.Five.Three != nil {
			out.Six.Four.One = string(in.Five.Three.One)
		}`),
		lines)
}

func TestConverterMapStructWithFieldMapDeeper2(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Six.Four.One"] = codegen.FieldMapperEntry{
		QualifiedName: "Five.Three.One",
		Override:      true,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedA {
 			1: optional string one
 		}

 		struct NestedB {
 			1: required NestedA three
		}

		struct NestedC {
			1: optional NestedA four
		}

 		struct Foo {
 			1: required NestedB five
 		}

 		struct Bar {
 			1: required NestedC six
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.Six = &structs.NestedC{}
		out.Six.Four = &structs.NestedA{}
		if in.Five != nil && in.Five.Three != nil {
			out.Six.Four.One = (*string)(in.Five.Three.One)
		}`),
		lines)
}

func TestConverterMapStructWithSubFieldsSwap(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Five.One"] = codegen.FieldMapperEntry{
		QualifiedName: "Four.Two",
		Override:      true,
	}
	fieldMap["Four.Two"] = codegen.FieldMapperEntry{
		QualifiedName: "Three.One",
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
 			3: optional NestedBar five
 			4: required NestedBar four
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		out.Five = &structs.NestedBar{}
		if in.Four != nil && in.Four.Two != nil {
			out.Five.One = *(in.Four.Two)
		}
		if in.Four != nil {
			out.Four = &structs.NestedBar{}
			out.Four.One = string(in.Four.One)
			if in.Three != nil {
				out.Four.Two = ptr.String(in.Three.One)
			}
		} else {
			out.Four = nil
		}
	`), lines)
}

func TestConverterMapStructWithSubFieldsReqToOpt(t *testing.T) {
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
			3: required NestedBar three
			4: optional NestedBar four
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assert.Equal(t, trim(`
		if in.Three != nil {
			out.Three = &structs.NestedBar{}
			out.Three.One = string(in.Three.One)
			if in.Three != nil && in.Three.Two != nil {
				out.Three.One = *(in.Three.Two)
			}
			out.Three.Two = (*string)(in.Three.Two)
		} else {
			out.Three = nil
		}
		if in.Four != nil {
			out.Four = &structs.NestedBar{}
			out.Four.One = string(in.Four.One)
			if in.Four != nil {
				out.Four.Two = ptr.String(in.Four.One)
			}
		} else {
			out.Four = nil
		}
	`), lines)
}

func TestConverterMapOverride(t *testing.T) {
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
				if in.Two[index] != nil {
					out.One[index].Field = (*string)(in.Two[index].Field)
				}
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
					if in.One[index] != nil {
						out.Two[index].Field = (*string)(in.One[index].Field)
					}
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
					if in.Two[key] != nil {
						out.One[key].Field = (*string)(in.Two[key].Field)
					}
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
					if in.One[key] != nil {
						out.Two[key].Field = (*string)(in.One[key].Field)
					}
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

func TestConvertWithMisMatchListTypesForOverride(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
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
			1: optional list<Inner> one
			2: optional string two
		}

		struct Bar {
			1: optional list<Inner> one
			2: optional list<Inner> two
		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Equal(t,
		"Could not convert field (two): type is not list",
		err.Error(),
	)
}

func TestConverterMapListTypeIncompatabile(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
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
			1: optional list<Inner> one
			2: optional list<string> two
		}

		struct Bar {
			1: optional list<Inner> one
			2: optional list<Inner> two
		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Equal(t,
		"could not convert struct fields, incompatible type for two :",
		err.Error(),
	)
}

func TestConvertWithMisMatchMapTypesForOverride(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
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
			2: optional string two
		}

		struct Bar {
			1: optional map<string, Inner> one
			2: optional map<string, Inner> two
		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Equal(t,
		"Could not convert field (two): type is not map",
		err.Error(),
	)
}

func TestConverterMapMapTypeIncompatabile(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
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
			2: optional map<string, string> two
		}

		struct Bar {
			1: optional map<string, Inner> one
			2: optional map<string, Inner> two
		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Equal(t,
		"could not convert struct fields, incompatible type for two :",
		err.Error(),
	)
}

func TestConverterInvalidMapping(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "Garbage",
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
			2: optional map<string, string> two
		}

		struct Bar {
			1: optional map<string, Inner> one
			2: optional map<string, Inner> two
		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)
	assert.Equal(t, "", lines)
	assert.Equal(t,
		"Failed to find field ( Garbage ) for transform.",
		err.Error(),
	)
}

// coverages error handling
func TestConvertNestedStructError(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Two"] = codegen.FieldMapperEntry{
		QualifiedName: "One",
		Override:      true,
	}
	_, err := convertTypes(
		"Foo", "Bar", `
		struct NestedFoo {
			1: required string three
		}

		struct NestedBar {
			1: required string four
		}


		struct Foo {
			1: required NestedFoo one
		}

		struct Bar {
			1: required NestedBar two
		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)
	assert.Contains(t,
		err.Error(),
		"required toField four does not have a valid fromField mapping",
	)
}
