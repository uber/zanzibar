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
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

// TypeConverter can generate a function body that converts two thriftrw
// FieldGroups from one to another. It's assumed that the converted code
// operates on two variables, "in" and "out" and that both are a go struct.
type TypeConverter struct {
	Lines  []string
	Helper *PackageHelper
}

// ConvertFields will add lines to the TypeConverter for mapping
// from one go struct to another based on two thriftrw.FieldGroups
func (c *TypeConverter) ConvertFields(
	keyPrefix string,
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

		prefix := "out." + keyPrefix + strings.Title(toField.Name)
		postfix := "in." + keyPrefix + strings.Title(fromField.Name)

		// Override thrift type names to avoid naming collisions between endpoint
		// and client types.
		switch toFieldType := toField.Type.(type) {
		case *compile.BoolSpec:
			var line string
			if toField.Required {
				line = prefix + " = bool(" + postfix + ")"
			} else {
				line = prefix + " = (*bool)(" + postfix + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.I8Spec:
			var line string
			if toField.Required {
				line = prefix + " = int8(" + postfix + ")"
			} else {
				line = prefix + " = (*int8)(" + postfix + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.I16Spec:
			var line string
			if toField.Required {
				line = prefix + " = int16(" + postfix + ")"
			} else {
				line = prefix + " = (*int16)(" + postfix + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.I32Spec:
			var line string
			if toField.Required {
				line = prefix + " = int32(" + postfix + ")"
			} else {
				line = prefix + " = (*int32)(" + postfix + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.I64Spec:
			var line string
			if toField.Required {
				line = prefix + " = int64(" + postfix + ")"
			} else {
				line = prefix + " = (*int64)(" + postfix + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.DoubleSpec:
			var line string
			if toField.Required {
				line = prefix + " = float64(" + postfix + ")"
			} else {
				line = prefix + " = (*float64)(" + postfix + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.StringSpec:
			var line string
			if toField.Required {
				line = prefix + " = string(" + postfix + ")"
			} else {
				line = prefix + " = (*string)(" + postfix + ")"
			}
			c.Lines = append(c.Lines, line)
		case *compile.BinarySpec:
			line := prefix + " = []byte(" + postfix + ")"
			c.Lines = append(c.Lines, line)
		case *compile.StructSpec:
			pkgName, err := c.Helper.TypePackageName(toField.Type.ThriftFile())
			if err != nil {
				return errors.Wrapf(
					err,
					"could not lookup struct when building converter for %s :",
					toField.Name,
				)
			}
			typeName := pkgName + "." + toField.Type.ThriftName()
			subToFields := toFieldType.Fields

			fromFieldType, ok := fromField.Type.(*compile.StructSpec)
			if !ok {
				return errors.Wrapf(
					err,
					"could not convert struct fields, "+
						"incompatible type for %s :",
					toField.Name,
				)
			}

			// TODO: ADD NIL CHECK !!!

			line := "if " + postfix + " != nil {"
			c.Lines = append(c.Lines, line)

			line = "	" + prefix + " = &" + typeName + "{}"
			c.Lines = append(c.Lines, line)

			subFromFields := fromFieldType.Fields
			err = c.ConvertFields(
				"	"+strings.Title(toField.Name)+".",
				subFromFields,
				subToFields,
			)
			if err != nil {
				return err
			}

			line = "} else {"
			c.Lines = append(c.Lines, line)

			line = "	" + prefix + " = nil"
			c.Lines = append(c.Lines, line)

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
