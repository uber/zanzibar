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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/codegen"
	"go.uber.org/thriftrw/compile"
	"go.uber.org/thriftrw/wire"
)

type fakeSpec struct{}

func (f *fakeSpec) ThriftName() string {
	return "fake"
}
func (f *fakeSpec) ThriftAnnotations() compile.Annotations {
	return nil
}
func (f *fakeSpec) Link(s compile.Scope) (compile.TypeSpec, error) {
	return nil, nil
}
func (f *fakeSpec) TypeCode() wire.Type {
	return 0
}
func (f *fakeSpec) ThriftFile() string {
	return ""
}
func (f *fakeSpec) ForEachTypeReference(func(compile.TypeSpec) error) error {
	return nil
}

type fakePackageNameResover struct{}

func (f *fakePackageNameResover) TypePackageName(string) (string, error) {
	return "", nil
}

func TestGoTypeUnknownType(t *testing.T) {
	s := &fakeSpec{}
	p := &fakePackageNameResover{}
	assert.PanicsWithValue(t, "Unknown type (*codegen_test.fakeSpec) for fake", func() {
		_, _ = codegen.GoType(p, s)
	})

}

func TestCustomTypeError(t *testing.T) {
	p := &fakePackageNameResover{}
	es := &compile.EnumSpec{
		File: "",
	}
	typ, err := codegen.GoType(p, es)
	assert.Equal(t, "", typ)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GoCustomType called with native type (*compile.EnumSpec)")
}

func TestCustomTypeInListError(t *testing.T) {
	p := &fakePackageNameResover{}
	es := &compile.ListSpec{
		ValueSpec: &compile.StructSpec{
			File: "",
		},
	}
	typ, err := codegen.GoType(p, es)
	assert.Equal(t, "", typ)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GoCustomType called with native type (*compile.StructSpec)")
}

func TestCustomTypeInSetError(t *testing.T) {
	p := &fakePackageNameResover{}
	es := &compile.SetSpec{
		ValueSpec: &compile.TypedefSpec{
			File: "",
		},
	}
	typ, err := codegen.GoType(p, es)
	assert.Equal(t, "", typ)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GoCustomType called with native type (*compile.TypedefSpec)")
}

func TestCustomTypeAsMapKeyError(t *testing.T) {
	p := &fakePackageNameResover{}
	es := &compile.MapSpec{
		KeySpec: &compile.TypedefSpec{
			File: "",
		},
	}
	typ, err := codegen.GoType(p, es)
	assert.Equal(t, "", typ)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GoCustomType called with native type (*compile.TypedefSpec)")
}

func TestCustomTypeAsMapValueError(t *testing.T) {
	p := &fakePackageNameResover{}
	es := &compile.MapSpec{
		KeySpec: &compile.StringSpec{},
		ValueSpec: &compile.TypedefSpec{
			File: "",
		},
	}
	typ, err := codegen.GoType(p, es)
	assert.Equal(t, "", typ)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GoCustomType called with native type (*compile.TypedefSpec)")
}
