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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/sergi/go-diff/diffmatchpatch"
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
	return codegen.NewTypeConverter(&naivePackageNameResolver{}, nil)
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

func assertPrettyEqual(t *testing.T, expected string, actual string) {
	success := assert.Equal(t, expected, actual)

	if !success {
		dmp := diffmatchpatch.New()

		diffs := dmp.DiffMain(expected, actual, true)
		fmt.Printf("\n ===== Detailed diff: %s ====== \n\n",
			t.Name(),
		)
		fmt.Printf("%s\n", dmp.DiffPrettyText(diffs))
		fmt.Printf("\n")
	}
}

func TestConvertStrings(t *testing.T) {
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
		out.One = (*bool)(&(in.One))
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
		out.One = (*structs.UUID)(in.One)
		out.Two = structs.UUID(in.Two)
	`), lines)
}

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
	assertPrettyEqual(t, trim(`
		out.One = (*structs.UUID)(&(in.One))
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
		out.One = make([]string, len(in.One))
		for index1, value2 := range in.One {
			out.One[index1] = string(value2)
		}
		out.Two = make([]string, len(in.Two))
		for index3, value4 := range in.Two {
			out.Two[index3] = string(value4)
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
	assertPrettyEqual(t, trim(`
		out.One = make([][]byte, len(in.One))
		for index1, value2 := range in.One {
			out.One[index1] = []byte(value2)
		}
		out.Two = make([][]byte, len(in.Two))
		for index3, value4 := range in.Two {
			out.Two[index3] = []byte(value4)
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
	assertPrettyEqual(t, trim(`
		out.One = make([]*structs.Inner, len(in.One))
		for index1, value2 := range in.One {
			if value2 != nil {
				out.One[index1] = &structs.Inner{}
				out.One[index1].Field = (*string)(in.One[index1].Field)
			} else {
				out.One[index1] = nil
			}
		}
		out.Two = make([]*structs.Inner, len(in.Two))
		for index3, value4 := range in.Two {
			if value4 != nil {
				out.Two[index3] = &structs.Inner{}
				out.Two[index3].Field = (*string)(in.Two[index3].Field)
			} else {
				out.Two[index3] = nil
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
	assertPrettyEqual(t, trim(`
		out.One = make(map[string]string, len(in.One))
		for key1, value2 := range in.One {
			out.One[key1] = string(value2)
		}
		out.Two = make(map[string]string, len(in.Two))
		for key3, value4 := range in.Two {
			out.Two[key3] = string(value4)
		}
	`), lines)
}

func TestConvertMapStringToStruct(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		struct MapValue {
			1: required string one
			2: optional string two
		}
		
		struct Foo {
			1: required map<string, MapValue> uuidMap
		}

		struct Bar {
			1: required map<string, MapValue> uuidMap
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
out.UuidMap = make(map[string]*structs.MapValue, len(in.UuidMap))
for key1, value2 := range in.UuidMap {
	if value2 != nil {
		out.UuidMap[key1] = &structs.MapValue{}
		out.UuidMap[key1].One = string(in.UuidMap[key1].One)
		out.UuidMap[key1].Two = (*string)(in.UuidMap[key1].Two)
	} else {
		out.UuidMap[key1] = nil
	}
}
	`), lines)
}

// Todo Test Override cases

func TestConvertMapTypeDefToStruct(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		typedef string UUID

		struct MapValue {
			1: required string one
			2: optional string two
		}
		
		struct Foo {
			1: required map<UUID, MapValue> uuidMap
		}

		struct Bar {
			1: required map<UUID, MapValue> uuidMap
		}`,
		nil,
		nil,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
out.UuidMap = make(map[structs.UUID]*structs.MapValue, len(in.UuidMap))
for key1, value2 := range in.UuidMap {
	if value2 != nil {
		out.UuidMap[key1] = &structs.MapValue{}
		out.UuidMap[key1].One = string(in.UuidMap[key1].One)
		out.UuidMap[key1].Two = (*string)(in.UuidMap[key1].Two)
	} else {
		out.UuidMap[key1] = nil
	}
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
	assertPrettyEqual(t, trim(`
		out.One = make(map[string]*structs.Inner, len(in.One))
		for key1, value2 := range in.One {
			if value2 != nil {
				out.One[key1] = &structs.Inner{}
				out.One[key1].Field = (*string)(in.One[key1].Field)
			} else {
				out.One[key1] = nil
			}
		}
		out.Two = make(map[string]*structs.Inner, len(in.Two))
		for key3, value4 := range in.Two {
			if value4 != nil {
				out.Two[key3] = &structs.Inner{}
				out.Two[key3].Field = (*string)(in.Two[key3].Field)
			} else {
				out.Two[key3] = nil
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
	assertPrettyEqual(t, trim(`
		out.One = make(map[structs.UUID]string, len(in.One))
		for key1, value2 := range in.One {
			 out.One[structs.UUID(key1)] = string(value2)
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
		out.One = (*structs.UUID)(&(in.Three))
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
		out.Three = (*bool)(&(in.Four))
		out.Four = (*bool)(&(in.Four))
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
	assertPrettyEqual(t, trim(`
		out.One = (*bool)(&(in.One))
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
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

func TestConverterMapStructWithFieldMapToOptField(t *testing.T) {
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
 			2: optional NestedFoo four
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
		if in.Three != nil {
			out.Three = &structs.NestedFoo{}
			out.Three.One = string(in.Three.One)
			out.Three.Two = (*string)(in.Three.Two)
		} else {
			out.Three = nil
		}
		if in.Three != nil && in.Three.Two != nil {
			if out.Five == nil {
				out.Five = &structs.NestedFoo{}
			}
			out.Five.One = *(in.Three.Two)
		}
		if in.Four != nil {
			if out.Five == nil {
				out.Five = &structs.NestedFoo{}
			}
			out.Five.Two = (*string)(&(in.Four.One))
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
	assertPrettyEqual(t, trim(`
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
	assertPrettyEqual(t, trim(`
		out.Six = &structs.NestedC{}
		if in.Five != nil && in.Five.Three != nil {
			if out.Six.Four == nil {
				out.Six.Four = &structs.NestedA{}
			}
			out.Six.Four.One = (*string)(in.Five.Three.One)
		}`),
		lines)
}

func TestConverterMapStructWithFieldMapDeeperOpt(t *testing.T) {
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
 			1: optional NestedC six
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
		if in.Five != nil && in.Five.Three != nil {
			if out.Six == nil {
				out.Six = &structs.NestedC{}
			}
			if out.Six.Four == nil {
				out.Six.Four = &structs.NestedA{}
			}
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
	assertPrettyEqual(t, trim(`
		if in.Four != nil && in.Four.Two != nil {
			if out.Five == nil {
				out.Five = &structs.NestedBar{}
			}
			out.Five.One = *(in.Four.Two)
		}
		if in.Four != nil {
			out.Four = &structs.NestedBar{}
			out.Four.One = string(in.Four.One)
			if in.Three != nil {
				out.Four.Two = (*string)(&(in.Three.One))
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
	assertPrettyEqual(t, trim(`
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
				out.Four.Two = (*string)(&(in.Four.One))
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
	assertPrettyEqual(t, trim(`
		out.One = make([]string, len(in.Two))
		for index1, value2 := range in.Two {
			out.One[index1] = string(value2)
		}
		sourceList3 := in.Two
		if in.One != nil {
			sourceList3 = in.One
		}
		out.Two = make([]string, len(sourceList3))
		for index5, value6 := range sourceList3 {
			out.Two[index5] = string(value6)
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
	assertPrettyEqual(t, trim(`
		out.One = make([]*structs.Inner, len(in.Two))
		for index1, value2 := range in.Two {
			if value2 != nil {
				out.One[index1] = &structs.Inner{}
				if in.Two[index1] != nil {
					out.One[index1].Field = (*string)(in.Two[index1].Field)
				}
			} else {
				out.One[index1] = nil
			}
		}
		sourceList3 := in.Two
		isOverridden4 := false
		if in.One != nil {
			sourceList3 = in.One
			isOverridden4 = true
		}
		out.Two = make([]*structs.Inner, len(sourceList3))
		for index5, value6 := range sourceList3 {
			if isOverridden4 {
				if value6 != nil {
					out.Two[index5] = &structs.Inner{}
					if in.One[index5] != nil {
						out.Two[index5].Field = (*string)(in.One[index5].Field)
					}
				} else {
					out.Two[index5] = nil
				}
			} else {
				if value6 != nil {
					out.Two[index5] = &structs.Inner{}
					out.Two[index5].Field = (*string)(in.Two[index5].Field)
				} else {
					out.Two[index5] = nil
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
	assertPrettyEqual(t, trim(`
		sourceList1 := in.One
		isOverridden2 := false
		if in.Two != nil {
			sourceList1 = in.Two
			isOverridden2 = true
		}
		out.One = make(map[string]*structs.Inner, len(sourceList1))
		for key3, value4 := range sourceList1 {
			if isOverridden2 {
				if value4 != nil {
					out.One[key3] = &structs.Inner{}
					if in.Two[key3] != nil {
						out.One[key3].Field = (*string)(in.Two[key3].Field)
					}
				} else {
					out.One[key3] = nil
				}
			} else {
				if value4 != nil {
					out.One[key3] = &structs.Inner{}
					out.One[key3].Field = (*string)(in.One[key3].Field)
				} else {
					out.One[key3] = nil
				}
			}
		}
		sourceList5 := in.Two
		isOverridden6 := false
		if in.One != nil {
			sourceList5 = in.One
			isOverridden6 = true
		}
		out.Two = make(map[string]*structs.Inner, len(sourceList5))
		for key7, value8 := range sourceList5 {
			if isOverridden6 {
				if value8 != nil {
					out.Two[key7] = &structs.Inner{}
					if in.One[key7] != nil {
						out.Two[key7].Field = (*string)(in.One[key7].Field)
					}
				} else {
					out.Two[key7] = nil
				}
			} else {
				if value8 != nil {
					out.Two[key7] = &structs.Inner{}
					out.Two[key7].Field = (*string)(in.Two[key7].Field)
				} else {
					out.Two[key7] = nil
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

func TestCoverIsTransform(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()
	f := &codegen.FieldAssignment{
		FromIdentifier: "gg.field",
		ToIdentifier:   "lol.field",
	}
	f.IsTransform()
}

func TestConverterRecursiveSimple(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
			1: required string three
			2: optional Foo recur
		}

		struct Bar {
			1: required string three
			2: optional Bar recur
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)

	assertPrettyEqual(t, trim(`
		out.Three = string(in.Three)
		var convertBarHelper1 func(in *structs.Foo) (out *structs.Bar)
		convertBarHelper1 = func(in *structs.Foo) (out *structs.Bar) {
			if in != nil {
				out = &structs.Bar{}
				out.Three = string(in.Three)
				out.Recur = convertBarHelper1(in.Recur)
			} else {
				out = nil
			}
			return
		}
		out.Recur = convertBarHelper1(in.Recur)
	`), lines)
}

func TestConverterRecursiveNested(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	lines, err := convertTypes(
		"Foo", "Bar",
		`
		struct StringFoo {
			1: required string one
			2: optional StringFoo recur 
		}

		struct Foo {
			1: required string three
			2: required StringFoo recur
		}

		struct Bar {
			1: required string three
			2: required StringFoo recur
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)

	assertPrettyEqual(t, trim(`
		out.Three = string(in.Three)
		var convertStringFooHelper1 func(in *structs.StringFoo) (out *structs.StringFoo)
		convertStringFooHelper1 = func(in *structs.StringFoo) (out *structs.StringFoo) {
			if in != nil {
				out = &structs.StringFoo{}
				out.One = string(in.One)
				out.Recur = convertStringFooHelper1(in.Recur)
			} else {
				out = nil
			}
			return
		}
		out.Recur = convertStringFooHelper1(in.Recur)
	`), lines)
}

func TestConverterRecursiveMixedDeep(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	lines, err := convertTypes(
		"Foo", "Bar",
		`
		struct DeepFoo {
			1: required string ten
			2: optional DeepFoo recur
		}

		struct StringFooA {
			1: required string one
			2: optional StringFooA recur 
			3: required DeepFoo deep
		}

		struct StringFooB {
			1: required string one
			2: optional StringFooB recur 
			3: required DeepFoo deep
		}

		struct Foo {
			1: required string three
			2: required StringFooA recur
			3: optional Foo recur2
		}

		struct Bar {
			1: required string three
			2: required StringFooB recur
			3: optional Bar recur2
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)

	assertPrettyEqual(t, trim(`
		out.Three = string(in.Three)
		var convertStringFooBHelper1 func(in *structs.StringFooA) (out *structs.StringFooB)
		convertStringFooBHelper1 = func(in *structs.StringFooA) (out *structs.StringFooB) {
			if in != nil {
				out = &structs.StringFooB{}
				out.One = string(in.One)
				out.Recur = convertStringFooBHelper1(in.Recur)
				var convertDeepFooHelper2 func(in *structs.DeepFoo) (out *structs.DeepFoo)
				convertDeepFooHelper2 = func(in *structs.DeepFoo) (out *structs.DeepFoo) {
					if in != nil {
						out = &structs.DeepFoo{}
						out.Ten = string(in.Ten)
						out.Recur = convertDeepFooHelper2(in.Recur)
					} else {
						out = nil
					}
					return
				}
				out.Deep = convertDeepFooHelper2(in.Deep)
			} else {
				out = nil
			}
			return
		}
		out.Recur = convertStringFooBHelper1(in.Recur)
		var convertBarHelper3 func(in *structs.Foo) (out *structs.Bar)
		convertBarHelper3 = func(in *structs.Foo) (out *structs.Bar) {
			if in != nil {
				out = &structs.Bar{}
				out.Three = string(in.Three)
				var convertStringFooBHelper4 func(in *structs.StringFooA) (out *structs.StringFooB)
				convertStringFooBHelper4 = func(in *structs.StringFooA) (out *structs.StringFooB) {
					if in != nil {
						out = &structs.StringFooB{}
						out.One = string(in.One)
						out.Recur = convertStringFooBHelper4(in.Recur)
						var convertDeepFooHelper5 func(in *structs.DeepFoo) (out *structs.DeepFoo)
						convertDeepFooHelper5 = func(in *structs.DeepFoo) (out *structs.DeepFoo) {
							if in != nil {
								out = &structs.DeepFoo{}
								out.Ten = string(in.Ten)
								out.Recur = convertDeepFooHelper5(in.Recur)
							} else {
								out = nil
							}
							return
						}
						out.Deep = convertDeepFooHelper5(in.Deep)
					} else {
						out = nil
					}
					return
				}
				out.Recur = convertStringFooBHelper4(in.Recur)
				out.Recur2 = convertBarHelper3(in.Recur2)
			} else {
				out = nil
			}
			return
		}
		out.Recur2 = convertBarHelper3(in.Recur2)
	`), lines)
}

func TestConverterRecursiveCyclic(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	lines, err := convertTypes(
		"Foo", "Bar",
		`
		struct StringFooA {
			1: required string one
			2: optional StringFooB fooB
		}
		struct StringFooB {
			1: required string two
			2: optional StringFooA fooA
		}

		struct Foo {
			1: required string three
			2: required StringFooA recur
		}

		struct Bar {
			1: required string three
			2: required StringFooA recur
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)

	assertPrettyEqual(t, trim(`
		out.Three = string(in.Three)
		var convertStringFooAHelper1 func(in *structs.StringFooA) (out *structs.StringFooA)
		convertStringFooAHelper1 = func(in *structs.StringFooA) (out *structs.StringFooA) {
			if in != nil {
				out = &structs.StringFooA{}
				out.One = string(in.One)
				var convertStringFooBHelper2 func(in *structs.StringFooB) (out *structs.StringFooB)
				convertStringFooBHelper2 = func(in *structs.StringFooB) (out *structs.StringFooB) {
					if in != nil {
						out = &structs.StringFooB{}
						out.Two = string(in.Two)
						out.FooA = convertStringFooAHelper1(in.FooA)
					} else {
						out = nil
					}
					return
				}
				out.FooB = convertStringFooBHelper2(in.FooB)
			} else {
				out = nil
			}
			return
		}
		out.Recur = convertStringFooAHelper1(in.Recur)
	`), lines)
}

func TestConverterRecursiveStructWithToReqMissingError(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	_, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional NestedFoo two
 		}

 		struct Foo {
 			1: required NestedFoo three
			2: optional Foo recur
 		}

 		struct Bar {
 			1: required NestedFoo three
			2: optional Bar recur
 			3: required NestedFoo eight
 		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)
	assert.Equal(t, "required toField eight does not have a valid fromField mapping", err.Error())
}

func TestConverterMapTypeDef(t *testing.T) {
	lines, err := convertTypes(
		"Foo", "Bar",
		`
		typedef string UUIDFoo
		typedef string UUIDBar

		struct RecurFoo2 {
			1: required UUIDFoo field1
		}
		
		struct RecurBar2 {
			1: required UUIDBar field1
		}

		struct RecurFoo1 {
			1: required map<UUIDFoo, RecurFoo2> field1
		}
		
		struct RecurBar1 {
			1: required map<UUIDBar, RecurBar2> field1
		}

		struct ProfileFoo {
			1: required RecurFoo1 field1
		}
		struct ProfileBar {
			2: required RecurBar1 field1
		}

 		struct Foo {
 			1: required list<ProfileFoo> one
 		}

 		struct Bar {
			1: required list<ProfileBar> one
 		}`,
		nil,
		nil,
	)
	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
	out.One = make([]*structs.ProfileBar, len(in.One))
	for index1, value2 := range in.One {
		if value2 != nil {
			out.One[index1] = &structs.ProfileBar{}
			if in.One[index1].Field1 != nil {
				out.One[index1].Field1 = &structs.RecurBar1{}
	out.One[index1].Field1.Field1 = make(map[structs.UUIDBar]*structs.RecurBar2, len(in.One[index1].Field1.Field1))
	for key3, value4 := range in.One[index1].Field1.Field1 {
					if value4 != nil {
						out.One[index1].Field1.Field1[structs.UUIDBar(key3)] = &structs.RecurBar2{}
						if in.One[index1] != nil && in.One[index1].Field1 != nil && in.One[index1].Field1.Field1[key3] != nil {
							out.One[index1].Field1.Field1[structs.UUIDBar(key3)].Field1 = structs.UUIDBar(in.One[index1].Field1.Field1[key3].Field1)
						}
					} else {
						out.One[index1].Field1.Field1[structs.UUIDBar(key3)] = nil
					}
	}
			} else {
				out.One[index1].Field1 = nil
			}
		} else {
			out.One[index1] = nil
		}
	}
	`), lines)
}

// can't really use field map to reassign recursive struct values
func TestConverterRecursiveStructWithFieldMap(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Five.One"] = codegen.FieldMapperEntry{
		QualifiedName: "Five.Two",
		Override:      true,
	}
	fieldMap["Five.Two"] = codegen.FieldMapperEntry{
		QualifiedName: "Five.One",
		Override:      true,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: required string two
			3: optional NestedFoo three
 		}

 		struct Foo {
 			1: required NestedFoo five
 		}

 		struct Bar {
 			2: required NestedFoo five
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
		inOriginal := in; _ = inOriginal
		outOriginal := out; _ = outOriginal
		var convertNestedFooHelper1 func(in *structs.NestedFoo) (out *structs.NestedFoo)
		convertNestedFooHelper1 = func(in *structs.NestedFoo) (out *structs.NestedFoo) {
			if in != nil {
				out = &structs.NestedFoo{}
				if inOriginal.Five != nil {
					out.One = string(inOriginal.Five.Two)
				}
				if inOriginal.Five != nil {
					out.Two = string(inOriginal.Five.One)
				}
				out.Three = convertNestedFooHelper1(in.Three)
			} else {
				out = nil
			}
			return
		}
		out.Five = convertNestedFooHelper1(in.Five)
`),
		lines)
}

func TestConverterRecursiveMapOptional(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["Seven.Two"] = codegen.FieldMapperEntry{
		QualifiedName: "Five.One",
		Override:      false,
	}
	fieldMap["Five.One"] = codegen.FieldMapperEntry{
		QualifiedName: "Five.Two",
		Override:      false,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: optional string one
 			2: optional string two
 		}

		struct RecurFoo {
			1: required string seven
			2: optional RecurFoo eight
		}

 		struct Foo {
 			1: optional NestedFoo five
			3: optional RecurFoo six
 		}

 		struct Bar {
 			1: optional NestedFoo five
			2: optional NestedFoo seven
			3: optional RecurFoo six
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
		inOriginal := in; _ = inOriginal
		outOriginal := out; _ = outOriginal
		var convertNestedFooHelper1 func(in *structs.NestedFoo) (out *structs.NestedFoo)
		convertNestedFooHelper1 = func(in *structs.NestedFoo) (out *structs.NestedFoo) {
			if in != nil {
				out = &structs.NestedFoo{}
				if inOriginal.Five != nil {
					out.One = (*string)(inOriginal.Five.Two)
				}
				out.Two = (*string)(in.Two)
			} else {
				out = nil
			}
			return
		}
		out.Five = convertNestedFooHelper1(in.Five)
		if inOriginal.Five != nil {
			if outOriginal.Seven == nil {
				outOriginal.Seven = &structs.NestedFoo{}
			}
			out.Seven.Two = (*string)(inOriginal.Five.One)
		}
		var convertRecurFooHelper2 func(in *structs.RecurFoo) (out *structs.RecurFoo)
		convertRecurFooHelper2 = func(in *structs.RecurFoo) (out *structs.RecurFoo) {
			if in != nil {
				out = &structs.RecurFoo{}
				out.Seven = string(in.Seven)
				out.Eight = convertRecurFooHelper2(in.Eight)
			} else {
				out = nil
			}
			return
		}
		out.Six = convertRecurFooHelper2(in.Six)`),
		lines)
}

func TestConverterRecursiveMapStruct(t *testing.T) {
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

		struct RecurFoo {
			1: required string seven
			2: optional RecurFoo eight
		}

 		struct Foo {
 			1: required NestedFoo three
 			2: required NestedFoo four
			3: optional RecurFoo six
 		}

 		struct Bar {
 			1: required NestedFoo three
 			2: optional NestedFoo five
			3: optional RecurFoo six
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
		inOriginal := in; _ = inOriginal
		outOriginal := out; _ = outOriginal
		var convertNestedFooHelper1 func(in *structs.NestedFoo) (out *structs.NestedFoo)
		convertNestedFooHelper1 = func(in *structs.NestedFoo) (out *structs.NestedFoo) {
			if in != nil {
				out = &structs.NestedFoo{}
				out.One = string(in.One)
				out.Two = (*string)(in.Two)
			} else {
				out = nil
			}
			return
		}
		out.Three = convertNestedFooHelper1(in.Three)
		if inOriginal.Three != nil && inOriginal.Three.Two != nil {
			if outOriginal.Five == nil {
				outOriginal.Five = &structs.NestedFoo{}
			}
			out.Five.One = *(inOriginal.Three.Two)
		}
		if inOriginal.Four != nil {
			if outOriginal.Five == nil {
				outOriginal.Five = &structs.NestedFoo{}
			}
			out.Five.Two = (*string)(&(inOriginal.Four.One))
		}
		var convertRecurFooHelper2 func(in *structs.RecurFoo) (out *structs.RecurFoo)
		convertRecurFooHelper2 = func(in *structs.RecurFoo) (out *structs.RecurFoo) {
			if in != nil {
				out = &structs.RecurFoo{}
				out.Seven = string(in.Seven)
				out.Eight = convertRecurFooHelper2(in.Eight)
			} else {
				out = nil
			}
			return
		}
		out.Six = convertRecurFooHelper2(in.Six)`),
		lines)
}

func TestConverterRecursiveMixedAll(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	lines, err := convertTypes(
		"Foo", "Bar",
		`
		struct NestedFooA {
 			1: optional map<string, NestedFooAA> one
 		}
		struct NestedFooAA {
			1: required list<NestedFooAAA> two
 		}
		struct NestedFooAAA {
			1: required string three
			2: optional NestedFooAAA four
			3: optional NestedFooAA five
 		}

		struct NestedFooB {
 			1: optional map<string, NestedFooBB> one
 		}
		struct NestedFooBB {
			1: required list<NestedFooBBB> two
 		}
		struct NestedFooBBB {
			1: required string three
			2: optional NestedFooBBB four
			3: optional NestedFooBB five
 		}


 		struct Foo {
 			1: optional NestedFooA six
 		}

 		struct Bar {
 			2: optional NestedFooB six
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
		var convertNestedFooBHelper1 func(in *structs.NestedFooA) (out *structs.NestedFooB)
		convertNestedFooBHelper1 = func(in *structs.NestedFooA) (out *structs.NestedFooB) {
			if in != nil {
				out = &structs.NestedFooB{}
		out.One = make(map[string]*structs.NestedFooBB, len(in.One))
		for key2, value3 := range in.One {
					if value3 != nil {
						out.One[key2] = &structs.NestedFooBB{}
		out.One[key2].Two = make([]*structs.NestedFooBBB, len(in.One[key2].Two))
		for index4, value5 := range in.One[key2].Two {
							if value5 != nil {
								out.One[key2].Two[index4] = &structs.NestedFooBBB{}
								out.One[key2].Two[index4].Three = string(in.One[key2].Two[index4].Three)
								var convertNestedFooBBBHelper6 func(in *structs.NestedFooAAA) (out *structs.NestedFooBBB)
								convertNestedFooBBBHelper6 = func(in *structs.NestedFooAAA) (out *structs.NestedFooBBB) {
									if in != nil {
										out = &structs.NestedFooBBB{}
										out.Three = string(in.Three)
										out.Four = convertNestedFooBBBHelper6(in.Four)
										var convertNestedFooBBHelper7 func(in *structs.NestedFooAA) (out *structs.NestedFooBB)
										convertNestedFooBBHelper7 = func(in *structs.NestedFooAA) (out *structs.NestedFooBB) {
											if in != nil {
												out = &structs.NestedFooBB{}
		out.Two = make([]*structs.NestedFooBBB, len(in.Two))
		for index8, value9 := range in.Two {
													if value9 != nil {
														out.Two[index8] = &structs.NestedFooBBB{}
														out.Two[index8].Three = string(in.Two[index8].Three)
														out.Two[index8].Four = convertNestedFooBBBHelper6(in.Two[index8].Four)
														out.Two[index8].Five = convertNestedFooBBHelper7(in.Two[index8].Five)
													} else {
														out.Two[index8] = nil
													}
		}
											} else {
												out = nil
											}
											return
										}
										out.Five = convertNestedFooBBHelper7(in.Five)
									} else {
										out = nil
									}
									return
								}
								out.One[key2].Two[index4].Four = convertNestedFooBBBHelper6(in.One[key2].Two[index4].Four)
								var convertNestedFooBBHelper10 func(in *structs.NestedFooAA) (out *structs.NestedFooBB)
								convertNestedFooBBHelper10 = func(in *structs.NestedFooAA) (out *structs.NestedFooBB) {
									if in != nil {
										out = &structs.NestedFooBB{}
		out.Two = make([]*structs.NestedFooBBB, len(in.Two))
		for index11, value12 := range in.Two {
											if value12 != nil {
												out.Two[index11] = &structs.NestedFooBBB{}
												out.Two[index11].Three = string(in.Two[index11].Three)
												var convertNestedFooBBBHelper13 func(in *structs.NestedFooAAA) (out *structs.NestedFooBBB)
												convertNestedFooBBBHelper13 = func(in *structs.NestedFooAAA) (out *structs.NestedFooBBB) {
													if in != nil {
														out = &structs.NestedFooBBB{}
														out.Three = string(in.Three)
														out.Four = convertNestedFooBBBHelper13(in.Four)
														out.Five = convertNestedFooBBHelper10(in.Five)
													} else {
														out = nil
													}
													return
												}
												out.Two[index11].Four = convertNestedFooBBBHelper13(in.Two[index11].Four)
												out.Two[index11].Five = convertNestedFooBBHelper10(in.Two[index11].Five)
											} else {
												out.Two[index11] = nil
											}
		}
									} else {
										out = nil
									}
									return
								}
								out.One[key2].Two[index4].Five = convertNestedFooBBHelper10(in.One[key2].Two[index4].Five)
							} else {
								out.One[key2].Two[index4] = nil
							}
		}
					} else {
						out.One[key2] = nil
					}
		}
			} else {
				out = nil
			}
			return
		}
		out.Six = convertNestedFooBHelper1(in.Six)
		`),
		lines)
}

func TestNestedMapOfList(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required map<string,list<NestedFoo>> one
 		}

 		struct Bar {
 			1: required map<string,list<NestedFoo>> one
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
	out.One = make(map[string][]*structs.NestedFoo, len(in.One))
	for key1, value2 := range in.One {
	out.One[key1] = make([]*structs.NestedFoo, len(value2))
	for index3, value4 := range value2 {
			if value4 != nil {
				out.One[key1][index3] = &structs.NestedFoo{}
				out.One[key1][index3].One = string(in.One[key1][index3].One)
				out.One[key1][index3].Two = (*string)(in.One[key1][index3].Two)
			} else {
				out.One[key1][index3] = nil
			}
	}
	}
	`), lines)
}

func TestNestedListOfMap(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required list<map<string,NestedFoo>> one
 		}

 		struct Bar {
 			1: required list<map<string,NestedFoo>> one
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
	out.One = make([]map[string]*structs.NestedFoo, len(in.One))
	for index1, value2 := range in.One {
	out.One[index1] = make(map[string]*structs.NestedFoo, len(value2))
	for key3, value4 := range value2 {
			if value4 != nil {
				out.One[index1][key3] = &structs.NestedFoo{}
				out.One[index1][key3].One = string(in.One[index1][key3].One)
				out.One[index1][key3].Two = (*string)(in.One[index1][key3].Two)
			} else {
				out.One[index1][key3] = nil
			}
	}
	}
	`), lines)
}

func TestNestedMapOfMap(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required map<string,map<string,NestedFoo>> one
 		}

 		struct Bar {
 			1: required map<string,map<string,NestedFoo>> one
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
	out.One = make(map[string]map[string]*structs.NestedFoo, len(in.One))
	for key1, value2 := range in.One {
	out.One[key1] = make(map[string]*structs.NestedFoo, len(value2))
	for key3, value4 := range value2 {
			if value4 != nil {
				out.One[key1][key3] = &structs.NestedFoo{}
				out.One[key1][key3].One = string(in.One[key1][key3].One)
				out.One[key1][key3].Two = (*string)(in.One[key1][key3].Two)
			} else {
				out.One[key1][key3] = nil
			}
	}
	}
	`), lines)
}

func TestNestedListOfList(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required list<list<NestedFoo>> one
 		}

 		struct Bar {
 			1: required list<list<NestedFoo>> one
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
	out.One = make([][]*structs.NestedFoo, len(in.One))
	for index1, value2 := range in.One {
	out.One[index1] = make([]*structs.NestedFoo, len(value2))
	for index3, value4 := range value2 {
			if value4 != nil {
				out.One[index1][index3] = &structs.NestedFoo{}
				out.One[index1][index3].One = string(in.One[index1][index3].One)
				out.One[index1][index3].Two = (*string)(in.One[index1][index3].Two)
			} else {
				out.One[index1][index3] = nil
			}
	}
	}
	`), lines)
}

func TestNestedListOfListofList(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required list<list<list<NestedFoo>>> one
 		}

 		struct Bar {
 			1: required list<list<list<NestedFoo>>> one
 		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
	out.One = make([][][]*structs.NestedFoo, len(in.One))
	for index1, value2 := range in.One {
	out.One[index1] = make([][]*structs.NestedFoo, len(value2))
	for index3, value4 := range value2 {
	out.One[index1][index3] = make([]*structs.NestedFoo, len(value4))
	for index5, value6 := range value4 {
				if value6 != nil {
					out.One[index1][index3][index5] = &structs.NestedFoo{}
					out.One[index1][index3][index5].One = string(in.One[index1][index3][index5].One)
					out.One[index1][index3][index5].Two = (*string)(in.One[index1][index3][index5].Two)
				} else {
					out.One[index1][index3][index5] = nil
				}
	}
	}
	}
	`), lines)
}

func TestNestedMapOfMapInvalid(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	_, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required map<string,map<NestedFoo,NestedFoo>> one
 		}

 		struct Bar {
 			1: required map<string,map<NestedFoo,NestedFoo>> one
 		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)
}

func TestNestedListOfMapInvalid(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	_, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required list<map<NestedFoo,NestedFoo>> one
 		}

 		struct Bar {
 			1: required list<map<NestedFoo,NestedFoo>> one
 		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)
}

func TestNestedListOfListOfMapInvalid(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	_, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required list<list<map<NestedFoo,NestedFoo>>> one
 		}

 		struct Bar {
 			1: required list<list<map<NestedFoo,NestedFoo>>> one
 		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)
}

func TestNestedMapOfListOfMapInvalid(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)

	_, err := convertTypes(
		"Foo", "Bar",
		`struct NestedFoo {
 			1: required string one
 			2: optional string two
 		}

 		struct Foo {
 			1: required map<string,list<map<NestedFoo,NestedFoo>>> one
 		}

 		struct Bar {
 			1: required map<string,list<map<NestedFoo,NestedFoo>>> one
 		}`,
		nil,
		fieldMap,
	)

	assert.Error(t, err)
}

func TestConverterMapNestedBinary(t *testing.T) {
	fieldMap := make(map[string]codegen.FieldMapperEntry)
	fieldMap["NestedTwo.NestedNestedTwo.Two"] = codegen.FieldMapperEntry{
		QualifiedName: "Two",
		Override:      false,
	}

	lines, err := convertTypes(
		"Foo", "Bar",
		`struct Foo {
            1: optional string one
			2: optional binary two
		}

        struct NestedBar {
            1: optional NestedNestedBar nestedNestedTwo
        }

        struct NestedNestedBar {
            1: optional binary two
        }

		struct Bar {
            1: optional string one
			2: optional NestedBar nestedTwo
		}`,
		nil,
		fieldMap,
	)

	assert.NoError(t, err)
	assertPrettyEqual(t, trim(`
		out.One = (*string)(in.One)
		if out.NestedTwo == nil {
			out.NestedTwo = &structs.NestedBar{}
		}
		if out.NestedTwo.NestedNestedTwo == nil {
			out.NestedTwo.NestedNestedTwo = &structs.NestedNestedBar{}
		}
		out.NestedTwo.NestedNestedTwo.Two = []byte(in.Two)
	`), lines)
}
