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

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

// GoType returns the Go type string representation for the given thrift type.
func GoType(p PackageNameResolver, spec compile.TypeSpec) (string, error) {
	switch s := spec.(type) {
	case *compile.BoolSpec:
		return "bool", nil
	case *compile.I8Spec:
		return "int8", nil
	case *compile.I16Spec:
		return "int16", nil
	case *compile.I32Spec:
		return "int32", nil
	case *compile.I64Spec:
		return "int64", nil
	case *compile.DoubleSpec:
		return "float64", nil
	case *compile.StringSpec:
		return "string", nil
	case *compile.BinarySpec:
		return "[]byte", nil
	case *compile.MapSpec:
		k, err := GoReferenceType(p, s.KeySpec)
		if err != nil {
			return "", err
		}
		v, err := GoReferenceType(p, s.ValueSpec)
		if err != nil {
			return "", err
		}
		if !isHashable(s.KeySpec) {
			return fmt.Sprintf("[]struct{Key %s; Value %s}", k, v), nil
		}
		return fmt.Sprintf("map[%s]%s", k, v), nil
	case *compile.ListSpec:
		v, err := GoReferenceType(p, s.ValueSpec)
		if err != nil {
			return "", err
		}
		return "[]" + v, nil
	case *compile.SetSpec:
		v, err := GoReferenceType(p, s.ValueSpec)
		if err != nil {
			return "", err
		}
		if !isHashable(s.ValueSpec) || isSliceSetType(s) {
			return fmt.Sprintf("[]%s", v), nil
		}
		return fmt.Sprintf("map[%s]struct{}", v), nil
	case *compile.EnumSpec, *compile.StructSpec, *compile.TypedefSpec:
		return GoCustomType(p, spec)
	default:
		panic(fmt.Sprintf("Unknown type (%T) for %s", spec, spec.ThriftName()))
	}
}

// GoReferenceType returns the Go reference type string representation for the given thrift type.
// for types like slice and map that are already of reference type, it returns the result of GoType;
// for struct type, it returns the pointer of the result of GoType.
func GoReferenceType(p PackageNameResolver, spec compile.TypeSpec) (string, error) {
	t, err := GoType(p, spec)
	if err != nil {
		return "", err
	}

	if IsStructType(spec) {
		t = "*" + t
	}

	return t, nil
}

// GoCustomType returns the user-defined Go type with its importing package.
func GoCustomType(p PackageNameResolver, spec compile.TypeSpec) (string, error) {
	f := spec.ThriftFile()
	if f == "" {
		return "", fmt.Errorf("GoCustomType called with native type (%T) %v", spec, spec)
	}

	pkg, err := p.TypePackageName(f)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get package for custom type (%T) %v", spec, spec)
	}

	return pkg + "." + PascalCase(spec.ThriftName()), nil
}

// IsStructType returns true if the given thrift type is struct, false otherwise.
func IsStructType(spec compile.TypeSpec) bool {
	spec = compile.RootTypeSpec(spec)
	_, isStruct := spec.(*compile.StructSpec)
	return isStruct
}

// isHashable returns true if the given type is considered hashable by thriftrw.
//
// Only primitive types, enums, and typedefs of other hashable types are considered hashable.
// binary is not considered a primitive type because it is represented as []byte in Go.
func isHashable(spec compile.TypeSpec) bool {
	spec = compile.RootTypeSpec(spec)
	switch spec.(type) {
	case *compile.BoolSpec, *compile.I8Spec, *compile.I16Spec, *compile.I32Spec,
		*compile.I64Spec, *compile.DoubleSpec, *compile.StringSpec, *compile.EnumSpec:
		return true
	default:
		return false
	}
}

// IsSliceSetType returns true if the given thrift type is a Set implemented as a slice (as opposed to a map)
func isSliceSetType(spec compile.TypeSpec) bool {
	spec = compile.RootTypeSpec(spec)
	_, isSet := spec.(*compile.SetSpec)
	return isSet && spec.ThriftAnnotations()["go.type"] == "slice"
}

func pointerMethodType(typeSpec compile.TypeSpec) string {
	var pointerMethod string

	switch typeSpec.(type) {
	case *compile.BoolSpec:
		pointerMethod = "Bool"
	case *compile.I8Spec:
		pointerMethod = "Int8"
	case *compile.I16Spec:
		pointerMethod = "Int16"
	case *compile.I32Spec:
		pointerMethod = "Int32"
	case *compile.I64Spec:
		pointerMethod = "Int64"
	case *compile.DoubleSpec:
		pointerMethod = "Float64"
	case *compile.StringSpec:
		pointerMethod = "String"
	case *compile.EnumSpec:
		pointerMethod = "Int32"
	default:
		panic(fmt.Sprintf(
			"Unknown type (%T) for %s for allocating a pointer",
			typeSpec, typeSpec.ThriftName(),
		))
	}

	return pointerMethod
}

type walkFieldVisitor func(
	goPrefix string,
	thriftPrefix string,
	field *compile.FieldSpec,
) bool

func walkFieldGroups(
	fields compile.FieldGroup,
	visitField walkFieldVisitor,
) bool {
	seen := map[*compile.FieldSpec]bool{}

	return walkFieldGroupsInternal("", "", fields, visitField, seen)
}

func walkFieldGroupsInternal(
	goPrefix string,
	thriftPrefix string,
	fields compile.FieldGroup,
	visitField walkFieldVisitor,
	seen map[*compile.FieldSpec]bool,
) bool {
	for i := 0; i < len(fields); i++ {
		field := fields[i]

		if seen[field] {
			// Skip this field if we have already considered it
			continue
		}
		seen[field] = true

		bail := visitField(goPrefix, thriftPrefix, field)
		if bail {
			return true
		}

		realType := compile.RootTypeSpec(field.Type)
		switch t := realType.(type) {
		case *compile.BinarySpec:
		case *compile.StringSpec:
		case *compile.BoolSpec:
		case *compile.DoubleSpec:
		case *compile.I8Spec:
		case *compile.I16Spec:
		case *compile.I32Spec:
		case *compile.I64Spec:
		case *compile.EnumSpec:
		case *compile.ListSpec:
		case *compile.SetSpec:
		case *compile.StructSpec:
			bail := walkFieldGroupsInternal(
				goPrefix+"."+PascalCase(field.Name),
				thriftPrefix+"."+field.Name,
				t.Fields,
				visitField,
				seen,
			)
			if bail {
				return true
			}
		case *compile.MapSpec:
			// TODO: implement
		default:
			panic("unknown Spec")
		}
	}

	return false
}
