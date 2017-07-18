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

// LineBuilder struct just appends/builds lines.
type LineBuilder struct {
	lines []string
}

// append helper will add a line to TypeConverter
func (l *LineBuilder) append(parts ...string) {
	line := strings.Join(parts, "")
	l.lines = append(l.lines, line)
}

// appendf helper will add a formatted line to TypeConverter
func (l *LineBuilder) appendf(format string, parts ...interface{}) {
	line := fmt.Sprintf(format, parts...)
	l.lines = append(l.lines, line)
}

// GetLines returns the lines in the line builder
func (l *LineBuilder) GetLines() []string {
	return l.lines
}

// PackageNameResolver interface allows for resolving what the
// package name for a thrift file is. This depends on where the
// thrift-based structs are generated.
type PackageNameResolver interface {
	TypePackageName(thriftFile string) (string, error)
}

// TypeConverter can generate a function body that converts two thriftrw
// FieldGroups from one to another. It's assumed that the converted code
// operates on two variables, "in" and "out" and that both are a go struct.
type TypeConverter struct {
	LineBuilder
	Helper PackageNameResolver
}

func (c *TypeConverter) getGoTypeName(valueType compile.TypeSpec) (string, error) {
	return GoType(c.Helper, valueType)
}

func (c *TypeConverter) assignWithOverride(
	indent string,
	toIdentifier string,
	typeName string,
	fromIdentifier string,
	overriddenIdentifier string,
) {
	if overriddenIdentifier == "" {
		c.append(indent, toIdentifier, " = ", typeName, "(", fromIdentifier, ")")
		return
	}

	// TODO: Verify all intermediate objects are not nil
	c.append(indent, toIdentifier, " = ", typeName, "(", overriddenIdentifier, ")")
	c.append(indent, "if ", fromIdentifier, " != nil {")
	c.append(indent, "\t", toIdentifier, " = ", typeName, "(", fromIdentifier, ")")
	c.append(indent, "}")
}

func (c *TypeConverter) getIdentifierName(fieldType compile.TypeSpec) (string, error) {
	t, err := goCustomType(c.Helper, fieldType)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"could not lookup fieldType when building converter for %s",
			fieldType.ThriftName(),
		)
	}
	return t, nil
}

func (c *TypeConverter) genConverterForStruct(
	toFieldName string,
	toFieldType *compile.StructSpec,
	fromFieldType compile.TypeSpec,
	fromIdentifier string,
	keyPrefix string,
	fromPrefix string,
	indent string,
	fieldMap map[string]FieldMapperEntry,
) error {
	toIdentifier := "out." + keyPrefix

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

	c.append(indent, "if ", fromIdentifier, " != nil {")

	c.append(indent, "\t", toIdentifier, " = &", typeName, "{}")

	subFromFields := fromFieldStruct.Fields
	subFromFieldsMap := make(map[string]FieldMapperEntry)
	// Build subfield mapping
	for k, v := range fieldMap {
		if strings.HasPrefix(v.QualifiedName, keyPrefix) {
			// string keyPrefix and append
			subFromFieldsMap[k] = v
		}
	}

	err = c.genStructConverter(
		keyPrefix+".",
		fromPrefix+".",
		indent+"\t",
		subFromFields,
		subToFields,
		subFromFieldsMap,
	)
	if err != nil {
		return err
	}

	c.append(indent, "} else {")

	c.append(indent, "\t", toIdentifier, " = nil")

	c.append(indent, "}")

	return nil
}

func (c *TypeConverter) genConverterForPrimitive(
	toField *compile.FieldSpec,
	toIdentifier string,
	fromIdentifier string,
	overriddenIdentifier string,
	indent string,
) error {
	typeName, err := c.getGoTypeName(toField.Type)
	if err != nil {
		return err
	}
	if toField.Required {
		c.assignWithOverride(
			indent,
			toIdentifier,
			typeName,
			fromIdentifier,
			overriddenIdentifier,
		)
	} else {
		c.assignWithOverride(
			indent,
			toIdentifier,
			fmt.Sprintf("(*%s)", typeName),
			fromIdentifier,
			overriddenIdentifier,
		)
	}
	return nil
}

func (c *TypeConverter) genConverterForList(
	toFieldType *compile.ListSpec,
	toField *compile.FieldSpec,
	fromField *compile.FieldSpec,
	overriddenField *compile.FieldSpec,
	toIdentifier string,
	fromIdentifier string,
	overriddenIdentifier string,
	keyPrefix string,
	indent string,
) error {
	typeName, err := c.getGoTypeName(toFieldType.ValueSpec)
	if err != nil {
		return err
	}

	valueStruct, isStruct := toFieldType.ValueSpec.(*compile.StructSpec)
	sourceIdentifier := fromIdentifier
	checkOverride := false
	if overriddenIdentifier != "" {
		// Determine which map (from or overrride) to use
		c.appendf("sourceList := %s", overriddenIdentifier)
		if isStruct {
			c.appendf("isOverridden := false")
		}
		// TODO(sindelar): Verify how optional thrift lists are defined.
		c.appendf("if %s != nil {", fromIdentifier)

		c.appendf("\tsourceList = %s", fromIdentifier)
		if isStruct {
			c.append("\tisOverridden = true")
		}
		c.append("}")

		sourceIdentifier = "sourceList"
		checkOverride = true
	}

	if isStruct {
		c.appendf(
			"%s = make([]*%s, len(%s))",
			toIdentifier, typeName, sourceIdentifier,
		)
	} else {
		c.appendf(
			"%s = make([]%s, len(%s))",
			toIdentifier, typeName, sourceIdentifier,
		)
	}

	c.append("for index, value := range ", sourceIdentifier, " {")

	if isStruct {
		nestedIndent := "\t" + indent

		if checkOverride {
			nestedIndent = "\t" + nestedIndent
			c.append("\t", "if isOverridden {")
		}
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
			keyPrefix+pascalCase(toField.Name)+"[index]",
			strings.TrimPrefix(fromIdentifier, "in.")+"[index]",
			nestedIndent,
			nil,
		)
		if err != nil {
			return err
		}
		if checkOverride {
			c.append("\t", "} else {")

			overriddenFieldType, ok := overriddenField.Type.(*compile.ListSpec)
			if !ok {
				return errors.Errorf(
					"Could not convert field (%s): type is not list",
					overriddenField.Name,
				)
			}

			err = c.genConverterForStruct(
				toField.Name,
				valueStruct,
				overriddenFieldType.ValueSpec,
				"value",
				keyPrefix+pascalCase(toField.Name)+"[index]",
				strings.TrimPrefix(overriddenIdentifier, "in.")+"[index]",
				nestedIndent,
				nil,
			)
			if err != nil {
				return err
			}
			c.append("\t", "}")
		}
	} else {
		c.append("\t", toIdentifier, "[index] = ", typeName, "(value)")
	}

	c.append("}")
	return nil
}

func (c *TypeConverter) genConverterForMap(
	toFieldType *compile.MapSpec,
	toField *compile.FieldSpec,
	fromField *compile.FieldSpec,
	overriddenField *compile.FieldSpec,
	toIdentifier string,
	fromIdentifier string,
	overriddenIdentifier string,
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
	sourceIdentifier := fromIdentifier
	checkOverride := false
	if overriddenIdentifier != "" {
		// Determine which map (from or overrride) to use
		c.appendf("sourceList := %s", overriddenIdentifier)

		if isStruct {
			c.appendf("isOverridden := false")
		}
		// TODO(sindelar): Verify how optional thrift map are defined.
		c.appendf("if %s != nil {", fromIdentifier)

		c.appendf("\tsourceList = %s", fromIdentifier)
		if isStruct {
			c.append("\tisOverridden = true")
		}
		c.append("}")

		sourceIdentifier = "sourceList"
		checkOverride = true
	}

	if isStruct {
		c.appendf(
			"%s = make(map[string]*%s, len(%s))",
			toIdentifier, typeName, sourceIdentifier,
		)
	} else {
		c.appendf(
			"%s = make(map[string]%s, len(%s))",
			toIdentifier, typeName, sourceIdentifier,
		)
	}

	c.appendf("for key, value := range %s {", sourceIdentifier)

	if isStruct {
		nestedIndent := "\t" + indent

		if checkOverride {
			nestedIndent = "\t" + nestedIndent
			c.append("\t", "if isOverridden {")
		}

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
			keyPrefix+pascalCase(toField.Name)+"[key]",
			strings.TrimPrefix(fromIdentifier, "in.")+"[key]",
			nestedIndent,
			nil,
		)
		if err != nil {
			return err
		}

		if checkOverride {
			c.append("\t", "} else {")

			overriddenFieldType, ok := overriddenField.Type.(*compile.MapSpec)
			if !ok {
				return errors.Errorf(
					"Could not convert field (%s): type is not map",
					overriddenField.Name,
				)
			}

			err = c.genConverterForStruct(
				toField.Name,
				valueStruct,
				overriddenFieldType.ValueSpec,
				"value",
				keyPrefix+pascalCase(toField.Name)+"[key]",
				strings.TrimPrefix(overriddenIdentifier, "in.")+"[key]",
				nestedIndent,
				nil,
			)
			if err != nil {
				return err
			}
			c.append("\t", "}")
		}
	} else {
		c.append("\t", toIdentifier, "[key] = ", typeName, "(value)")
	}

	c.append("}")
	return nil
}

func (c *TypeConverter) genStructConverter(
	keyPrefix string,
	fromPrefix string,
	indent string,
	fromFields []*compile.FieldSpec,
	toFields []*compile.FieldSpec,
	fieldMap map[string]FieldMapperEntry,
) error {
	for i := 0; i < len(toFields); i++ {
		toField := toFields[i]

		// Check for same named field
		var fromField *compile.FieldSpec
		for j := 0; j < len(fromFields); j++ {
			if fromFields[j].Name == toField.Name {
				fromField = fromFields[j]
				break
			}
		}

		toIdentifier := "out." + keyPrefix + pascalCase(toField.Name)
		overriddenIdentifier := ""
		fromIdentifier := ""

		// Check for mapped field
		var overriddenField *compile.FieldSpec

		v, ok := fieldMap[keyPrefix+pascalCase(toField.Name)]
		if ok {
			if fromField == nil {
				fromField = v.Field
				fromIdentifier = "in." + v.QualifiedName
			} else {
				if v.Override {
					// check for required/optional setting
					if !v.Field.Required {
						overriddenField = fromField
						overriddenIdentifier = "in." + fromPrefix +
							pascalCase(overriddenField.Name)
					}
					// If override is true and the new field is required,
					// there's a default instantiation value and will always
					// overwrite.
					fromField = v.Field
					fromIdentifier = "in." + v.QualifiedName
				} else {
					// If override is false and the from field is required,
					// From is always populated and will never be overwritten.
					if !fromField.Required {
						overriddenField = v.Field
						overriddenIdentifier = "in." + v.QualifiedName
					}
				}
			}
		}

		if fromField == nil {
			return errors.Errorf(
				"cannot map by name for the field %s",
				toField.Name,
			)
		}

		if fromIdentifier == "" {
			fromIdentifier = "in." + fromPrefix + pascalCase(fromField.Name)
		}

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
				toField, toIdentifier, fromIdentifier, overriddenIdentifier, indent,
			)
			if err != nil {
				return err
			}
		case *compile.BinarySpec:
			// TODO: handle override. Check if binarySpec can be optional.
			c.append(toIdentifier, " = []byte(", fromIdentifier, ")")
		case *compile.TypedefSpec:
			typeName, err := c.getIdentifierName(toField.Type)
			if err != nil {
				return err
			}

			// TODO: typedef for struct is invalid here ...
			if toField.Required {
				c.assignWithOverride(
					indent,
					toIdentifier,
					typeName,
					fromIdentifier,
					overriddenIdentifier,
				)
			} else {
				c.assignWithOverride(
					indent,
					toIdentifier,
					fmt.Sprintf("(*%s)", typeName),
					fromIdentifier,
					overriddenIdentifier,
				)
			}

		case *compile.StructSpec:
			err := c.genConverterForStruct(
				toField.Name,
				toFieldType,
				fromField.Type,
				fromIdentifier,
				keyPrefix+pascalCase(toField.Name),
				keyPrefix+pascalCase(fromField.Name),
				indent,
				fieldMap,
			)
			if err != nil {
				return err
			}
		case *compile.ListSpec:
			err := c.genConverterForList(
				toFieldType,
				toField,
				fromField,
				overriddenField,
				toIdentifier,
				fromIdentifier,
				overriddenIdentifier,
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
				overriddenField,
				toIdentifier,
				fromIdentifier,
				overriddenIdentifier,
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
// from one go struct to another based on two thriftrw.FieldGroups.
// fieldMap is a may from keys that are the qualified field path names for
// destination fields (from the incoming request) and the entries are source
// fields (sent to the downstream client)
func (c *TypeConverter) GenStructConverter(
	fromFields []*compile.FieldSpec,
	toFields []*compile.FieldSpec,
	fieldMap map[string]FieldMapperEntry,
) error {
	err := c.genStructConverter("", "", "", fromFields, toFields, fieldMap)
	if err != nil {
		return err
	}

	return nil
}

// FieldMapperEntry defines a source field and optional arguments
// converting and overriding fields.
type FieldMapperEntry struct {
	QualifiedName string
	Field         *compile.FieldSpec

	Override      bool
	typeConverter string // TODO: implement. i.e string(int) etc
	transform     string // TODO: implement. i.e. camelCasing, Title, etc
}
