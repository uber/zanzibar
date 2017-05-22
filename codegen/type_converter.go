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

func (c *TypeConverter) genConverterForStruct(
	toFieldName string,
	toFieldType *compile.StructSpec,
	fromFieldType compile.TypeSpec,
	fromIdentifier string,
	keyPrefix string,
	indent string,
) error {
	toIdentifier := indent + "out." + keyPrefix

	typeName, err := c.getIdentifierName(toFieldType)
	if err != nil {
		return err
	}
	subToFields := toFieldType.Fields

	fromFieldStruct, ok := fromFieldType.(*compile.StructSpec)
	if !ok {
		return errors.Errorf(
			"could not convert struct fields, "+
				"incompatible type for %s :",
			toFieldName,
		)
	}

	line := "if " + fromIdentifier + " != nil {"
	c.Lines = append(c.Lines, line)

	line = "	" + toIdentifier + " = &" + typeName + "{}"
	c.Lines = append(c.Lines, line)

	subFromFields := fromFieldStruct.Fields
	err = c.genStructConverter(
		keyPrefix+".",
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

	return nil
}

func (c *TypeConverter) genConverterForPrimitive(
	toField *compile.FieldSpec,
	toIdentifier string,
	fromIdentifier string,
) error {
	var line string
	typeName, err := c.getGoTypeName(toField.Type)
	if err != nil {
		return err
	}

	if toField.Required {
		line = toIdentifier + " = " + typeName + "(" + fromIdentifier + ")"
	} else {
		line = toIdentifier + " = (*" + typeName + ")(" + fromIdentifier + ")"
	}
	c.Lines = append(c.Lines, line)
	return nil
}

func (c *TypeConverter) genConverterForList(
	toFieldType *compile.ListSpec,
	toField *compile.FieldSpec,
	fromField *compile.FieldSpec,
	toIdentifier string,
	fromIdentifier string,
	keyPrefix string,
	indent string,
) error {
	typeName, err := c.getGoTypeName(toFieldType.ValueSpec)
	if err != nil {
		return err
	}

	valueStruct, isStruct := toFieldType.ValueSpec.(*compile.StructSpec)
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
		fromFieldType, ok := fromField.Type.(*compile.ListSpec)
		if !ok {
			return errors.Errorf(
				"Could not convert field (%s): type is not list",
				fromField.Name,
			)
		}

		err = c.genConverterForStruct(
			toField.Name,
			valueStruct,
			fromFieldType.ValueSpec,
			"value",
			keyPrefix+strings.Title(toField.Name)+"[index]",
			indent,
		)
		if err != nil {
			return err
		}
	} else {
		line = toIdentifier + "[index] = " + typeName + "(value)"
		c.Lines = append(c.Lines, line)
	}

	line = "}"
	c.Lines = append(c.Lines, line)
	return nil
}

func (c *TypeConverter) genConverterForMap(
	toFieldType *compile.MapSpec,
	toField *compile.FieldSpec,
	fromField *compile.FieldSpec,
	toIdentifier string,
	fromIdentifier string,
	keyPrefix string,
	indent string,
) error {
	typeName, err := c.getGoTypeName(toFieldType.ValueSpec)
	if err != nil {
		return err
	}

	_, isStringKey := toFieldType.KeySpec.(*compile.StringSpec)
	if !isStringKey {
		return errors.Errorf(
			"could not convert key (%s), map is not string-keyed.",
			toField.Name,
		)
	}

	valueStruct, isStruct := toFieldType.ValueSpec.(*compile.StructSpec)
	if isStruct {
		line := toIdentifier + " = make(map[string]*" +
			typeName + ", len(" + fromIdentifier + "))"
		c.Lines = append(c.Lines, line)
	} else {
		line := toIdentifier + " = make(map[string]" +
			typeName + ", len(" + fromIdentifier + "))"
		c.Lines = append(c.Lines, line)
	}

	line := "for key, value := range " + fromIdentifier + " {"
	c.Lines = append(c.Lines, line)

	if isStruct {
		fromFieldType, ok := fromField.Type.(*compile.MapSpec)
		if !ok {
			return errors.Errorf(
				"Could not convert field (%s): type is not map",
				fromField.Name,
			)
		}

		err = c.genConverterForStruct(
			toField.Name,
			valueStruct,
			fromFieldType.ValueSpec,
			"value",
			keyPrefix+strings.Title(toField.Name)+"[key]",
			indent,
		)
		if err != nil {
			return err
		}
	} else {
		line = toIdentifier + "[key] = " + typeName + "(value)"
		c.Lines = append(c.Lines, line)
	}

	line = "}"
	c.Lines = append(c.Lines, line)
	return nil
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
		case
			*compile.BoolSpec,
			*compile.I8Spec,
			*compile.I16Spec,
			*compile.I32Spec,
			*compile.EnumSpec,
			*compile.I64Spec,
			*compile.DoubleSpec,
			*compile.StringSpec:

			err := c.genConverterForPrimitive(
				toField, toIdentifier, fromIdentifier,
			)
			if err != nil {
				return err
			}
		case *compile.BinarySpec:
			line := toIdentifier + " = []byte(" + fromIdentifier + ")"
			c.Lines = append(c.Lines, line)
		case *compile.TypedefSpec:
			typeName, err := c.getIdentifierName(toField.Type)
			if err != nil {
				return err
			}

			var line string
			// TODO: typedef for struct is invalid here ...
			if toField.Required {
				line = toIdentifier + " = " + typeName + "(" + fromIdentifier + ")"
			} else {
				line = toIdentifier + " = (*" + typeName + ")(" + fromIdentifier + ")"
			}
			c.Lines = append(c.Lines, line)

		case *compile.StructSpec:
			err := c.genConverterForStruct(
				toField.Name,
				toFieldType,
				fromField.Type,
				fromIdentifier,
				keyPrefix+strings.Title(toField.Name),
				indent,
			)
			if err != nil {
				return err
			}
		case *compile.ListSpec:
			err := c.genConverterForList(
				toFieldType,
				toField,
				fromField,
				toIdentifier,
				fromIdentifier,
				keyPrefix,
				indent,
			)
			if err != nil {
				return err
			}
		case *compile.MapSpec:
			err := c.genConverterForMap(
				toFieldType,
				toField,
				fromField,
				toIdentifier,
				fromIdentifier,
				keyPrefix,
				indent,
			)
			if err != nil {
				return err
			}
		default:
			// fmt.Printf("Unknown type %s for field %s \n",
			// 	toField.Type.TypeCode().String(), toField.Name,
			// )

			// pkgName, err := h.TypePackageName(toField.Type.ThriftFile())
			// if err != nil {
			// 	return nil, err
			// }
			// typeName := pkgName + "." + toField.Type.ThriftName()
			// line := toIdentifier + "(*" + typeName + ")" + postfix
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
