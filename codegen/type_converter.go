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

package codegen

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

// TypeConverter can generate a function body that converts two thriftrw
// FieldGroups from one to another. It's assumed that the converted code
// operates on two variables, "in" and "out" and that both are a go struct.
type TypeConverter struct {
	Lines  []string
	Helper PackageNameResolver
}

// PackageNameResolver interface allows for resolving what the
// package name for a thrift file is. This depends on where the
// thrift-based structs are generated.
type PackageNameResolver interface {
	TypePackageName(thriftFile string) (string, error)
}

func (c *TypeConverter) getGoTypeName(
	valueType compile.TypeSpec,
) (string, error) {
	switch s := valueType.(type) {
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
		panic("Not Implemented")
	case *compile.SetSpec:
		panic("Not Implemented")
	case *compile.ListSpec:
		panic("Not Implemented")
	case *compile.EnumSpec, *compile.StructSpec, *compile.TypedefSpec:
		return c.getIdentifierName(s)
	default:
		panic(fmt.Sprintf("Unknown type (%T) %v", valueType, valueType))
	}
}

func (c *TypeConverter) getIdentifierName(
	fieldType compile.TypeSpec,
) (string, error) {
	pkgName, err := c.Helper.TypePackageName(fieldType.ThriftFile())
	if err != nil {
		return "", errors.Wrapf(
			err,
			"could not lookup fieldType when building converter for %s :",
			fieldType.ThriftName(),
		)
	}
	typeName := pkgName + "." + fieldType.ThriftName()
	return typeName, nil
}

func (c *TypeConverter) genStructConverter(
	keyPrefix string,
	indent string,
	fromFields []*compile.FieldSpec,
	toFields []*compile.FieldSpec,
) error {
	for i := 0; i < len(toFields); i++ {
		toField := toFields[i]

		var fromField *compile.FieldSpec
		for j := 0; j < len(fromFields); j++ {
			if fromFields[j].Name == toField.Name {
				fromField = fromFields[j]
				break
			}
		}

		if fromField == nil {
			return errors.Errorf(
				"cannot map by name for the field %s",
				toField.Name,
			)
		}

		toIdentifier := indent + "out." + keyPrefix + strings.Title(toField.Name)
		fromIdentifier := "in." + keyPrefix + strings.Title(fromField.Name)

		// Override thrift type names to avoid naming collisions between endpoint
		// and client types.
		switch toFieldType := toField.Type.(type) {
		case *compile.BoolSpec:
			var line string
			if toField.Required {
				line = toIdentifier + " = bool(" + fromIdentifier + ")"
			} else {
				line = toIdentifier + " = (*bool)(" + fromIdentifier + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.I8Spec:
			var line string
			if toField.Required {
				line = toIdentifier + " = int8(" + fromIdentifier + ")"
			} else {
				line = toIdentifier + " = (*int8)(" + fromIdentifier + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.I16Spec:
			var line string
			if toField.Required {
				line = toIdentifier + " = int16(" + fromIdentifier + ")"
			} else {
				line = toIdentifier + " = (*int16)(" + fromIdentifier + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.I32Spec:
			var line string
			if toField.Required {
				line = toIdentifier + " = int32(" + fromIdentifier + ")"
			} else {
				line = toIdentifier + " = (*int32)(" + fromIdentifier + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.EnumSpec:
			typeName, err := c.getIdentifierName(toField.Type)
			if err != nil {
				return err
			}

			var line string
			if toField.Required {
				line = toIdentifier + " = " + typeName + "(" + fromIdentifier + ")"
			} else {
				line = toIdentifier + " = (*" + typeName + ")(" + fromIdentifier + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.I64Spec:
			var line string
			if toField.Required {
				line = toIdentifier + " = int64(" + fromIdentifier + ")"
			} else {
				line = toIdentifier + " = (*int64)(" + fromIdentifier + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.DoubleSpec:
			var line string
			if toField.Required {
				line = toIdentifier + " = float64(" + fromIdentifier + ")"
			} else {
				line = toIdentifier + " = (*float64)(" + fromIdentifier + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.StringSpec:
			var line string
			if toField.Required {
				line = toIdentifier + " = string(" + fromIdentifier + ")"
			} else {
				line = toIdentifier + " = (*string)(" + fromIdentifier + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.BinarySpec:
			line := toIdentifier + " = []byte(" + fromIdentifier + ")"
			c.Lines = append(c.Lines, line)
		case *compile.TypedefSpec:
			typeName, err := c.getIdentifierName(toField.Type)
			if err != nil {
				return err
			}

			var line string
			if toField.Required {
				line = toIdentifier + " = " + typeName + "(" + fromIdentifier + ")"
			} else {
				line = toIdentifier + " = (*" + typeName + ")(" + fromIdentifier + ")"
			}
			c.Lines = append(c.Lines, line)

		case *compile.StructSpec:
			typeName, err := c.getIdentifierName(toField.Type)
			if err != nil {
				return err
			}
			subToFields := toFieldType.Fields

			fromFieldType, ok := fromField.Type.(*compile.StructSpec)
			if !ok {
				return errors.Errorf(
					"could not convert struct fields, "+
						"incompatible type for %s :",
					toField.Name,
				)
			}

			line := "if " + fromIdentifier + " != nil {"
			c.Lines = append(c.Lines, line)

			line = "	" + toIdentifier + " = &" + typeName + "{}"
			c.Lines = append(c.Lines, line)

			subFromFields := fromFieldType.Fields
			err = c.genStructConverter(
				keyPrefix+strings.Title(toField.Name)+".",
				indent+"	",
				subFromFields,
				subToFields,
			)
			if err != nil {
				return err
			}

			line = "} else {"
			c.Lines = append(c.Lines, line)

			line = "	" + toIdentifier + " = nil"
			c.Lines = append(c.Lines, line)

			line = "}"
			c.Lines = append(c.Lines, line)

		case *compile.ListSpec:
			typeName, err := c.getGoTypeName(toFieldType.ValueSpec)
			if err != nil {
				return err
			}

			_, isStruct := toFieldType.ValueSpec.(*compile.StructSpec)
			if isStruct {
				line := toIdentifier + " = make([]*" +
					typeName + ", len(" + fromIdentifier + "))"
				c.Lines = append(c.Lines, line)
			} else {
				line := toIdentifier + " = make([]" +
					typeName + ", len(" + fromIdentifier + "))"
				c.Lines = append(c.Lines, line)
			}

			line := "for index, value := range " + fromIdentifier + " {"
			c.Lines = append(c.Lines, line)

			if isStruct {
				line = toIdentifier + "[index] = " +
					"(*" + typeName + ")(value)"
				c.Lines = append(c.Lines, line)
			} else {
				line = toIdentifier + "[index] = " +
					typeName + "(value)"
				c.Lines = append(c.Lines, line)
			}

			line = "}"
			c.Lines = append(c.Lines, line)

		default:
			// fmt.Printf("Unknown type %s for field %s \n",
			// 	toField.Type.TypeCode().String(), toField.Name,
			// )

			// pkgName, err := h.TypePackageName(toField.Type.ThriftFile())
			// if err != nil {
			// 	return nil, err
			// }
			// typeName := pkgName + "." + toField.Type.ThriftName()
			// line := prefix + "(*" + typeName + ")" + postfix
			// c.Lines = append(c.Lines, line)
		}
	}

	return nil
}

// GenStructConverter will add lines to the TypeConverter for mapping
// from one go struct to another based on two thriftrw.FieldGroups
func (c *TypeConverter) GenStructConverter(
	fromFields []*compile.FieldSpec,
	toFields []*compile.FieldSpec,
) error {
	err := c.genStructConverter("", "", fromFields, toFields)
	if err != nil {
		return err
	}

	return nil
}
