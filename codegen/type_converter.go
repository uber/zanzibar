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
	"sort"
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

type fieldStruct struct {
	Identifier string
	TypeName   string
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
	Helper          PackageNameResolver
	uninitialized   map[string]*fieldStruct
	fieldCounter    int
	convStructMap   map[string]string
	useRecurGen     bool
	optionalEntries map[string]FieldMapperEntry
}

type toFieldParam struct {
	Type       compile.TypeSpec
	Name       string
	Required   bool
	Identifier string
}

type fromFieldParam struct {
	Type            compile.TypeSpec
	Name            string
	Identifier      string
	ValueIdentifier string
}

type overriddenFieldParam struct {
	Type       compile.TypeSpec
	Name       string
	Identifier string
}

// NewTypeConverter returns *TypeConverter (tc)
// @optionalEntries contains Entries that already set for tc fromFields
// entry set in @optionalEntries, tc will not raise error when this entry
// is required for toFields but missing from fromFields
func NewTypeConverter(h PackageNameResolver, optionalEntries map[string]FieldMapperEntry) *TypeConverter {
	return &TypeConverter{
		LineBuilder:     LineBuilder{},
		Helper:          h,
		uninitialized:   make(map[string]*fieldStruct),
		convStructMap:   make(map[string]string),
		optionalEntries: optionalEntries,
	}
}

func (c *TypeConverter) makeUniqIdentifier(prefix string) string {
	c.fieldCounter++

	return fmt.Sprintf("%s%d", prefix, c.fieldCounter)
}

func (c *TypeConverter) getGoTypeName(valueType compile.TypeSpec) (string, error) {
	return GoType(c.Helper, valueType)
}

//   input of "A.B.C.D" returns ["A","A.B", "A.B.C", "A.B.C.D"]
func getMiddleIdentifiers(identifier string) []string {
	subIds := strings.Split(identifier, ".")

	middleIds := make([]string, 0, len(subIds))

	middleIds = append(middleIds, subIds[0])
	for i := 1; i < len(subIds); i++ {
		middleIds = append(middleIds, fmt.Sprintf("%s.%s", middleIds[i-1], subIds[i]))
	}

	return middleIds
}

//  converts a list of identifier paths into boolean nil check expressions on those paths
func convertIdentifiersToNilChecks(identifiers []string) []string {
	checks := make([]string, 0, len(identifiers)-1)

	for i, ident := range identifiers {
		if i == 0 {
			continue // discard the check on "in" object
		}

		checks = append(checks, ident+" != nil")
	}
	return checks
}

func (c *TypeConverter) getIdentifierName(fieldType compile.TypeSpec) (string, error) {
	t, err := GoCustomType(c.Helper, fieldType)
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
	toRequired bool,
	fromFieldType compile.TypeSpec,
	fromIdentifier string,
	keyPrefix string,
	fromPrefix string,
	indent string,
	fieldMap map[string]FieldMapperEntry,
	prevKeyPrefixes []string,
) error {
	toIdentifier := "out"
	if keyPrefix != "" {
		toIdentifier += "." + keyPrefix
	}

	typeName, _ := c.getIdentifierName(toFieldType)

	subToFields := toFieldType.Fields

	// if no fromFieldType assume we're constructing from transform fieldMap  TODO: make this less subtle
	if fromFieldType == nil {
		//  in the direct assignment we do a nil check on fromField here. its hard for unknown number of transforms
		//  initialize the toField with an empty struct only if it's required, otherwise send to uninitialized map
		if toRequired {
			c.append(indent, toIdentifier, " = &", typeName, "{}")
		} else {
			id := "out"
			if c.useRecurGen {
				id = "outOriginal"
			}
			if keyPrefix != "" {
				id += "." + keyPrefix
			}
			c.uninitialized[id] = &fieldStruct{
				Identifier: id,
				TypeName:   typeName,
			}
		}

		if keyPrefix != "" {
			keyPrefix += "."
		}
		// if fromPrefix != "" {
		// fromPrefix += "."
		// }

		// recursive call
		err := c.genStructConverter(
			keyPrefix,
			fromPrefix,
			indent,
			nil,
			subToFields,
			fieldMap,
			prevKeyPrefixes,
		)
		if err != nil {
			return err
		}
		return nil
	}

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

	if keyPrefix != "" {
		keyPrefix += "."
	}
	if fromPrefix != "" {
		fromPrefix += "."
	}

	subFromFields := fromFieldStruct.Fields
	err := c.genStructConverter(
		keyPrefix,
		fromPrefix,
		indent+"\t",
		subFromFields,
		subToFields,
		fieldMap,
		prevKeyPrefixes,
	)
	if err != nil {
		return err
	}

	c.append(indent, "} else {")
	c.append(indent, "\t", toIdentifier, " = nil")
	c.append(indent, "}")

	return nil
}

func (c *TypeConverter) genConverterForList(
	toField toFieldParam,
	fromField fromFieldParam,
	overriddenField overriddenFieldParam,
	indent string,
) error {
	toFieldType := toField.Type.(*compile.ListSpec)

	typeName, err := c.getGoTypeName(toFieldType.ValueSpec)
	if err != nil {
		return err
	}

	valueStruct, isStruct := compile.RootTypeSpec(toFieldType.ValueSpec).(*compile.StructSpec)
	valueList, isList := compile.RootTypeSpec(toFieldType.ValueSpec).(*compile.ListSpec)
	valueMap, isMap := compile.RootTypeSpec(toFieldType.ValueSpec).(*compile.MapSpec)
	sourceIdentifier := fromField.ValueIdentifier
	checkOverride := false

	sourceListID := ""
	isOverriddenID := ""

	if overriddenField.Identifier != "" {
		sourceListID = c.makeUniqIdentifier("sourceList")
		isOverriddenID = c.makeUniqIdentifier("isOverridden")

		// Determine which map (from or overrride) to use
		c.appendf("%s := %s", sourceListID, overriddenField.Identifier)
		if isStruct {
			c.appendf("%s := false", isOverriddenID)
		}
		// TODO(sindelar): Verify how optional thrift lists are defined.
		c.appendf("if %s != nil {", fromField.Identifier)

		c.appendf("\t%s = %s", sourceListID, fromField.Identifier)
		if isStruct {
			c.appendf("\t%s = true", isOverriddenID)
		}
		c.append("}")

		sourceIdentifier = sourceListID
		checkOverride = true
	}

	if isStruct {
		c.appendf(
			"%s = make([]*%s, len(%s))",
			toField.Identifier, typeName, sourceIdentifier,
		)
	} else {
		c.appendf(
			"%s = make([]%s, len(%s))",
			toField.Identifier, typeName, sourceIdentifier,
		)
	}

	indexID := c.makeUniqIdentifier("index")
	valID := c.makeUniqIdentifier("value")
	c.appendf(
		"for %s, %s := range %s {",
		indexID, valID, sourceIdentifier,
	)

	if isStruct || isList || isMap {
		nestedIndent := "\t" + indent

		fromFieldListType, ok := compile.RootTypeSpec(fromField.Type).(*compile.ListSpec)
		if !ok {
			return errors.Errorf(
				"Could not convert field (%s): type is not list",
				fromField.Name,
			)
		}

		if isStruct {
			if checkOverride {
				nestedIndent = "\t" + nestedIndent
				c.appendf("\tif %s {", isOverriddenID)
			}

			err = c.genConverterForStruct(
				toField.Name,
				valueStruct,
				toField.Required,
				fromFieldListType.ValueSpec,
				valID,
				trimAnyPrefix(toField.Identifier, "out.", "outOriginal.")+"["+indexID+"]",
				trimAnyPrefix(fromField.Identifier, "in.", "inOriginal.")+"["+indexID+"]",
				nestedIndent,
				nil,
				nil,
			)
			if err != nil {
				return err
			}
			if checkOverride {
				c.append("\t", "} else {")

				overriddenFieldListType, ok := overriddenField.Type.(*compile.ListSpec)
				if !ok {
					return errors.Errorf(
						"Could not convert field (%s): type is not list",
						overriddenField.Name,
					)
				}

				err = c.genConverterForStruct(
					toField.Name,
					valueStruct,
					toField.Required,
					overriddenFieldListType.ValueSpec,
					valID,
					trimAnyPrefix(toField.Identifier, "out.", "outOriginal.")+"["+indexID+"]",
					trimAnyPrefix(overriddenField.Identifier, "in.", "inOriginal.")+"["+indexID+"]",
					nestedIndent,
					nil,
					nil,
				)
				if err != nil {
					return err
				}
				c.append("\t", "}")
			}
		} else if isList {
			err = c.genConverterForList(
				toFieldParam{
					valueList,
					toField.Name,
					toField.Required,
					toField.Identifier + "[" + indexID + "]",
				},
				fromFieldParam{
					fromFieldListType.ValueSpec,
					fromField.Name,
					fromField.Identifier + "[" + indexID + "]",
					valID,
				},
				overriddenField,
				nestedIndent)
			if err != nil {
				return err
			}
		} else if isMap {
			err = c.genConverterForMap(
				toFieldParam{
					valueMap,
					toField.Name,
					toField.Required,
					toField.Identifier + "[" + indexID + "]",
				},
				fromFieldParam{
					fromFieldListType.ValueSpec,
					fromField.Name,
					fromField.Identifier + "[" + indexID + "]",
					valID,
				},
				overriddenField,
				nestedIndent)
			if err != nil {
				return err
			}
		}
	} else {
		c.appendf(
			"\t%s[%s] = %s(%s)",
			toField.Identifier, indexID, typeName, valID,
		)
	}

	c.append("}")
	return nil
}

func (c *TypeConverter) genConverterForMap(
	toField toFieldParam,
	fromField fromFieldParam,
	overriddenField overriddenFieldParam,
	indent string,
) error {
	toFieldType := toField.Type.(*compile.MapSpec)

	typeName, err := c.getGoTypeName(toFieldType.ValueSpec)
	if err != nil {
		return err
	}

	valueStruct, isStruct := compile.RootTypeSpec(toFieldType.ValueSpec).(*compile.StructSpec)
	valueList, isList := compile.RootTypeSpec(toFieldType.ValueSpec).(*compile.ListSpec)
	valueMap, isMap := compile.RootTypeSpec(toFieldType.ValueSpec).(*compile.MapSpec)
	sourceIdentifier := fromField.ValueIdentifier
	_, isStringKey := toFieldType.KeySpec.(*compile.StringSpec)
	keyType := "string"

	if !isStringKey {
		realType := compile.RootTypeSpec(toFieldType.KeySpec)
		switch realType.(type) {
		case *compile.StringSpec:
			keyType, _ = c.getGoTypeName(toFieldType.KeySpec)
		default:
			return errors.Errorf(
				"could not convert key (%s), map is not string-keyed.", toField.Name)
		}
	}

	checkOverride := false
	sourceListID := ""
	isOverriddenID := ""
	if overriddenField.Identifier != "" {
		sourceListID = c.makeUniqIdentifier("sourceList")
		isOverriddenID = c.makeUniqIdentifier("isOverridden")

		// Determine which map (from or overrride) to use
		c.appendf("%s := %s", sourceListID, overriddenField.Identifier)

		if isStruct {
			c.appendf("%s := false", isOverriddenID)
		}
		// TODO(sindelar): Verify how optional thrift map are defined.
		c.appendf("if %s != nil {", fromField.Identifier)

		c.appendf("\t%s = %s", sourceListID, fromField.Identifier)
		if isStruct {
			c.appendf("\t%s = true", isOverriddenID)
		}
		c.append("}")

		sourceIdentifier = sourceListID
		checkOverride = true
	}
	if isStruct {
		c.appendf(
			"%s = make(map[%s]*%s, len(%s))",
			toField.Identifier, keyType, typeName, sourceIdentifier,
		)
	} else {
		c.appendf(
			"%s = make(map[%s]%s, len(%s))",
			toField.Identifier, keyType, typeName, sourceIdentifier,
		)
	}

	keyID := c.makeUniqIdentifier("key")
	valID := c.makeUniqIdentifier("value")
	c.appendf(
		"for %s, %s := range %s {",
		keyID, valID, sourceIdentifier,
	)

	if isStruct || isList || isMap {
		nestedIndent := "\t" + indent

		fromFieldMapType, ok := compile.RootTypeSpec(fromField.Type).(*compile.MapSpec)
		if !ok {
			return errors.Errorf(
				"Could not convert field (%s): type is not map",
				fromField.Name,
			)
		}

		toFieldKeyID := keyID

		toTypeKeyName := keyType
		fromTypeKeyName, _ := c.getGoTypeName(fromFieldMapType.KeySpec)

		if fromTypeKeyName != toTypeKeyName {
			toFieldKeyID = toTypeKeyName + "(" + keyID + ")"
		}

		if isStruct {
			if checkOverride {
				nestedIndent = "\t" + nestedIndent
				c.appendf("\tif %s {", isOverriddenID)
			}

			err = c.genConverterForStruct(
				toField.Name,
				valueStruct,
				toField.Required,
				fromFieldMapType.ValueSpec,
				valID,
				trimAnyPrefix(toField.Identifier, "out.", "outOriginal.")+"["+toFieldKeyID+"]",
				trimAnyPrefix(fromField.Identifier, "in.", "inOriginal.")+"["+keyID+"]",
				nestedIndent,
				nil,
				nil,
			)
			if err != nil {
				return err
			}

			if checkOverride {
				c.append("\t", "} else {")

				overriddenFieldMapType, ok := overriddenField.Type.(*compile.MapSpec)
				if !ok {
					return errors.Errorf(
						"Could not convert field (%s): type is not map",
						overriddenField.Name,
					)
				}

				err = c.genConverterForStruct(
					toField.Name,
					valueStruct,
					toField.Required,
					overriddenFieldMapType.ValueSpec,
					valID,
					trimAnyPrefix(toField.Identifier, "out.", "outOriginal.")+"["+toFieldKeyID+"]",
					trimAnyPrefix(overriddenField.Identifier, "in.", "inOriginal.")+"["+keyID+"]",
					nestedIndent,
					nil,
					nil,
				)
				if err != nil {
					return err
				}
				c.append("\t", "}")
			}
		} else if isList {
			err = c.genConverterForList(
				toFieldParam{
					valueList,
					toField.Name,
					toField.Required,
					toField.Identifier + "[" + toFieldKeyID + "]",
				},
				fromFieldParam{
					fromFieldMapType.ValueSpec,
					fromField.Name,
					fromField.Identifier + "[" + keyID + "]",
					valID,
				},
				overriddenField,
				nestedIndent)
			if err != nil {
				return err
			}
		} else if isMap {
			err = c.genConverterForMap(
				toFieldParam{
					valueMap,
					toField.Name,
					toField.Required,
					toField.Identifier + "[" + toFieldKeyID + "]",
				},
				fromFieldParam{
					fromFieldMapType.ValueSpec,
					fromField.Name,
					fromField.Identifier + "[" + keyID + "]",
					valID,
				},
				overriddenField,
				nestedIndent)
			if err != nil {
				return err
			}
		}
	} else {
		if keyType == "string" {
			c.appendf(
				"\t%s[%s] = %s(%s)",
				toField.Identifier, keyID, typeName, valID,
			)

		} else {
			c.appendf(
				"\t %s[%s(%s)] = %s(%s)",
				toField.Identifier, keyType, keyID, typeName, valID,
			)
		}
	}

	c.append("}")
	return nil
}

// recursive function to walk a DFS on toFields and try to assign fromFields or fieldMap tranforms
//  generated code is appended as we traverse the toFields thrift type structure
//  keyPrefix - the identifier (path) of the current position in the "to" struct
//  fromPrefix - the identifier (path) of the corresponding position in the "from" struct
//  indent - a string of tabs for current block scope
//  fromFields - fields in the current from struct,  can be nil  if only fieldMap transforms are applicable in the path
//  toFields - fields in the current to struct
//  fieldMap - a data structure specifying configured transforms   Map[toIdentifier ] ->  fromField FieldMapperEntry
func (c *TypeConverter) genStructConverter(
	keyPrefix string,
	fromPrefix string,
	indent string,
	fromFields []*compile.FieldSpec,
	toFields []*compile.FieldSpec,
	fieldMap map[string]FieldMapperEntry,
	prevKeyPrefixes []string,
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

		toSubIdentifier := keyPrefix + PascalCase(toField.Name)
		toIdentifier := "out." + toSubIdentifier
		overriddenIdentifier := ""
		fromIdentifier := ""

		// Check for mapped field
		var overriddenField *compile.FieldSpec

		// check if this toField satisfies a fieldMap transform
		transformFrom, ok := fieldMap[toSubIdentifier]
		if ok {
			// no existing direct fromField,  just assign the transform
			if fromField == nil {
				fromField = transformFrom.Field
				if c.useRecurGen {
					fromIdentifier = "inOriginal." + transformFrom.QualifiedName
				} else {
					fromIdentifier = "in." + transformFrom.QualifiedName
				}
				// else there is a conflicting direct fromField
			} else {
				//  depending on Override flag either the direct fromField or transformFrom is the OverrideField
				if transformFrom.Override {
					// check for required/optional setting
					if !transformFrom.Field.Required {
						overriddenField = fromField
						overriddenIdentifier = "in." + fromPrefix +
							PascalCase(overriddenField.Name)
					}
					// If override is true and the new field is required,
					// there's a default instantiation value and will always
					// overwrite.
					fromField = transformFrom.Field
					if c.useRecurGen {
						fromIdentifier = "inOriginal." + transformFrom.QualifiedName
					} else {
						fromIdentifier = "in." + transformFrom.QualifiedName
					}
				} else {
					// If override is false and the from field is required,
					// From is always populated and will never be overwritten.
					if !fromField.Required {
						overriddenField = transformFrom.Field
						if c.useRecurGen {
							fromIdentifier = "inOriginal." + transformFrom.QualifiedName
						} else {
							overriddenIdentifier = "in." + transformFrom.QualifiedName
						}
					}
				}
			}
		}

		// neither direct or transform fromField was found
		if fromField == nil {
			// search the fieldMap toField identifiers for matching identifier prefix
			// e.g.  the current toField is a struct and something within it has a transform
			//   a full match identifiers for transform non-struct types would have been caught above
			hasStructFieldMapping := false
			for toID := range fieldMap {
				if strings.HasPrefix(toID, toSubIdentifier) {
					hasStructFieldMapping = true
				}
			}

			//  if there's no fromField and no fieldMap transform that could be applied
			if !hasStructFieldMapping {
				var bypass bool
				// check if required field is filled from other resources
				// it can be used to set system default (customized tracing /auth required for clients),
				// or header propagating
				if c.optionalEntries != nil {
					for toID := range c.optionalEntries {
						if strings.HasPrefix(toID, toSubIdentifier) {
							bypass = true
							break
						}
					}
				}

				// the toField is either covered by optionalEntries, or optional and
				// there's nothing that maps to it or its sub-fields so we should skip it
				if bypass || !toField.Required {
					continue
				}

				// unrecoverable error
				return errors.Errorf(
					"required toField %s does not have a valid fromField mapping",
					toField.Name,
				)
			}
		}

		if fromIdentifier == "" && fromField != nil {
			// should we set this if no fromField ??
			fromIdentifier = "in." + fromPrefix + PascalCase(fromField.Name)
		}

		if prevKeyPrefixes == nil {
			prevKeyPrefixes = []string{}
		}

		var overriddenFieldName string
		var overriddenFieldType compile.TypeSpec
		if overriddenField != nil {
			overriddenFieldName = overriddenField.Name
			overriddenFieldType = overriddenField.Type
		}

		// Override thrift type names to avoid naming collisions between endpoint
		// and client types.
		switch toFieldType := compile.RootTypeSpec(toField.Type).(type) {
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
				toField,
				toIdentifier,
				fromField,
				fromIdentifier,
				overriddenField,
				overriddenIdentifier,
				indent,
				prevKeyPrefixes,
			)
			if err != nil {
				return err
			}
		case *compile.BinarySpec:
			for _, line := range checkOptionalNil(indent, c.uninitialized, toIdentifier, prevKeyPrefixes, c.useRecurGen) {
				c.append(line)
			}
			c.append(toIdentifier, " = []byte(", fromIdentifier, ")")
		case *compile.StructSpec:
			var (
				stFromPrefix = fromPrefix
				stFromType   compile.TypeSpec
				fromTypeName string
			)
			if fromField != nil {
				stFromType = fromField.Type
				stFromPrefix = fromPrefix + PascalCase(fromField.Name)

				fromTypeName, _ = c.getIdentifierName(stFromType)
			}

			toTypeName, err := c.getIdentifierName(toFieldType)
			if err != nil {
				return err
			}

			if converterMethodName, ok := c.convStructMap[toFieldType.Name]; ok {
				// the converter for this struct has already been generated, so just use it
				c.append(indent, "out.", keyPrefix+PascalCase(toField.Name), " = ", converterMethodName, "(", fromIdentifier, ")")
			} else if c.useRecurGen && fromTypeName != "" {
				// generate a callable converter inside function literal
				err = c.genConverterForStructWrapped(
					toField,
					toFieldType,
					toTypeName,
					toSubIdentifier,
					fromTypeName,
					fromIdentifier,
					stFromType,
					fieldMap,
					prevKeyPrefixes,
					indent,
				)
			} else {
				err = c.genConverterForStruct(
					toField.Name,
					toFieldType,
					toField.Required,
					stFromType,
					fromIdentifier,
					keyPrefix+PascalCase(toField.Name),
					stFromPrefix,
					indent,
					fieldMap,
					prevKeyPrefixes,
				)
			}
			if err != nil {
				return err
			}
		case *compile.ListSpec:
			err := c.genConverterForList(
				toFieldParam{
					toFieldType,
					toField.Name,
					toField.Required,
					toIdentifier,
				},
				fromFieldParam{
					fromField.Type,
					fromField.Name,
					fromIdentifier,
					fromIdentifier,
				},
				overriddenFieldParam{
					overriddenFieldType,
					overriddenFieldName,
					overriddenIdentifier,
				},
				indent,
			)
			if err != nil {
				return err
			}
		case *compile.MapSpec:
			err := c.genConverterForMap(
				toFieldParam{
					toFieldType,
					toField.Name,
					toField.Required,
					toIdentifier,
				},
				fromFieldParam{
					fromField.Type,
					fromField.Name,
					fromIdentifier,
					fromIdentifier,
				},
				overriddenFieldParam{
					overriddenFieldType,
					overriddenFieldName,
					overriddenIdentifier,
				},
				indent,
			)
			if err != nil {
				return err
			}
		default:
			// fmt.Printf("Unknown type %s for field %s \n",
			// 	toField.Type.TypeCode().String(), toField.Name,
			// )

			// pkgName, err := h.TypePackageName(toField.Type.IDLFile())
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
// destination fields (sent to the downstream client) and the entries are source
// fields (from the incoming request)
func (c *TypeConverter) GenStructConverter(
	fromFields []*compile.FieldSpec,
	toFields []*compile.FieldSpec,
	fieldMap map[string]FieldMapperEntry,
) error {
	// Add compiled FieldSpecs to the FieldMapperEntry
	fieldMap = addSpecToMap(fieldMap, fromFields, "")
	// Check for vlaues not populated recursively by addSpecToMap
	for k, v := range fieldMap {
		if fieldMap[k].Field == nil {
			return errors.Errorf(
				"Failed to find field ( %s ) for transform.",
				v.QualifiedName,
			)
		}
	}

	c.useRecurGen = c.isRecursiveStruct(toFields) || c.isRecursiveStruct(fromFields)

	if c.useRecurGen && len(fieldMap) != 0 {
		c.append("inOriginal := in; _ = inOriginal")
		c.append("outOriginal := out; _ = outOriginal")
	}

	err := c.genStructConverter("", "", "", fromFields, toFields, fieldMap, nil)
	if err != nil {
		return err
	}

	return nil
}

// FieldMapperEntry defines a source field and optional arguments
// converting and overriding fields.
type FieldMapperEntry struct {
	QualifiedName string //
	Field         *compile.FieldSpec

	Override      bool
	typeConverter string // TODO: implement. i.e string(int) etc
	transform     string // TODO: implement. i.e. camelCasing, Title, etc
}

func addSpecToMap(
	overrideMap map[string]FieldMapperEntry,
	fields compile.FieldGroup,
	prefix string,
) map[string]FieldMapperEntry {
	for k, v := range overrideMap {
		for _, spec := range fields {
			fieldQualName := prefix + PascalCase(spec.Name)
			if v.QualifiedName == fieldQualName {
				v.Field = spec
				overrideMap[k] = v
				break
			} else if strings.HasPrefix(v.QualifiedName, fieldQualName) {
				structSpec := spec.Type.(*compile.StructSpec)
				overrideMap = addSpecToMap(overrideMap, structSpec.Fields, fieldQualName+".")
			}
		}
	}
	return overrideMap
}

// FieldAssignment is responsible to generate an assignment
type FieldAssignment struct {
	From           *compile.FieldSpec
	FromIdentifier string
	To             *compile.FieldSpec
	ToIdentifier   string
	TypeCast       string
	ExtraNilCheck  bool
}

// IsTransform returns true if field assignment is not proxy
func (f *FieldAssignment) IsTransform() bool {
	fromI := strings.Index(f.FromIdentifier, ".") + 1
	toI := strings.Index(f.ToIdentifier, ".") + 1

	// make sure FromIdentifier and ToIdentifier have valid prefixes
	if (f.FromIdentifier[:fromI] != "in." && f.FromIdentifier[:fromI] != "inOriginal.") || f.ToIdentifier[:toI] != "out." {
		panic("isTransform check called on unexpected input  fromIdentifer " + f.FromIdentifier + " toIdentifer " + f.ToIdentifier)
	}

	// it's a transform when FromIdentifier != ToIdentifier (after removing prefix)
	return f.FromIdentifier[fromI:] != f.ToIdentifier[toI:]
}

// checkOptionalNil nil-checks fields that may not be initialized along
// the assign path
func checkOptionalNil(
	indent string,
	uninitialized map[string]*fieldStruct,
	toIdentifier string,
	prevKeyPrefixes []string,
	useRecurGen bool,
) []string {
	var ret = make([]string, 0)
	if len(uninitialized) < 1 {
		return ret
	}
	keys := make([]string, 0, len(uninitialized))
	for k := range uninitialized {
		keys = append(keys, k)
	}

	toIdentifier = trimAnyPrefix(toIdentifier, "out.", "outOriginal.")

	completeID := "out."
	if useRecurGen {
		completeID = "outOriginal."
	}
	completeID += strings.Join(append(prevKeyPrefixes, toIdentifier), ".")

	sort.Strings(keys)
	for _, id := range keys {
		if strings.HasPrefix(completeID, id) {
			v := uninitialized[id]
			ret = append(ret, indent+"if "+v.Identifier+" == nil {")
			ret = append(ret, indent+"\t"+v.Identifier+" = &"+v.TypeName+"{}")
			ret = append(ret, indent+"}")
		}
	}
	return ret
}

// dealing with primary type pointer formatting: (*float64)&Param -> (*float64)(&(Param))
func formatAssign(indent, to, typeCast, from string) string {
	if string(typeCast[len(typeCast)-1]) == "&" {
		typeCast = typeCast[:len(typeCast)-1]
		return fmt.Sprintf("%s%s = %s(&(%s))", indent, to, typeCast, from)
	}
	return fmt.Sprintf("%s%s = %s(%s)", indent, to, typeCast, from)
}

// Generate generates assignment
func (f *FieldAssignment) Generate(indent string, uninitialized map[string]*fieldStruct, prevKeyPrefixes []string, useRecurGen bool) string {
	var lines = make([]string, 0)
	if !f.IsTransform() {
		//  do an extra nil check for Overrides
		if (!f.From.Required && f.To.Required) || f.ExtraNilCheck {
			// need to nil check otherwise we could be cast or deref nil
			lines = append(lines, indent+"if "+f.FromIdentifier+" != nil {")
			lines = append(lines, checkOptionalNil(indent+"\t", uninitialized, f.ToIdentifier, prevKeyPrefixes, useRecurGen)...)
			lines = append(lines, formatAssign(indent+"\t", f.ToIdentifier, f.TypeCast, f.FromIdentifier))
			lines = append(lines, indent+"}")
		} else {
			lines = append(lines, checkOptionalNil(indent, uninitialized, f.ToIdentifier, prevKeyPrefixes, useRecurGen)...)
			lines = append(lines, formatAssign(indent, f.ToIdentifier, f.TypeCast, f.FromIdentifier))
		}
		return strings.Join(lines, "\n")
	}
	//  generates nil checks for intermediate objects on the f.FromIdentifier path
	checks := convertIdentifiersToNilChecks(getMiddleIdentifiers(f.FromIdentifier))
	if !f.ExtraNilCheck {
		// if from is required or from/to are both optional   then we don't need the final outer check
		if f.From.Required || !f.From.Required && !f.To.Required {
			checks = checks[:len(checks)-1]
		}
		if len(checks) == 0 {
			lines = append(lines, checkOptionalNil(indent, uninitialized, f.ToIdentifier, prevKeyPrefixes, useRecurGen)...)
			lines = append(lines, formatAssign(indent, f.ToIdentifier, f.TypeCast, f.FromIdentifier))
			return strings.Join(lines, "\n")
		}
	}
	lines = append(lines, indent+"if "+strings.Join(checks, " && ")+" {")
	lines = append(lines, checkOptionalNil(indent+"\t", uninitialized, f.ToIdentifier, prevKeyPrefixes, useRecurGen)...)
	lines = append(lines, formatAssign(indent+"\t", f.ToIdentifier, f.TypeCast, f.FromIdentifier))
	lines = append(lines, indent+"}")
	// TODO else?  log? should this silently eat intermediate nils as none-assignment,  should set nil?
	return strings.Join(lines, "\n")
}

func (c *TypeConverter) assignWithOverride(
	indent string,
	defaultAssign *FieldAssignment,
	overrideAssign *FieldAssignment,
	prevKeyPrefixes []string,
) {
	if overrideAssign != nil && overrideAssign.FromIdentifier != "" {
		c.append(overrideAssign.Generate(indent, c.uninitialized, prevKeyPrefixes, c.useRecurGen))
		defaultAssign.ExtraNilCheck = true
		c.append(defaultAssign.Generate(indent, c.uninitialized, prevKeyPrefixes, c.useRecurGen))
		return
	}
	c.append(defaultAssign.Generate(indent, c.uninitialized, prevKeyPrefixes, c.useRecurGen))
}

func (c *TypeConverter) genConverterForPrimitive(
	toField *compile.FieldSpec,
	toIdentifier string,
	fromField *compile.FieldSpec,
	fromIdentifier string,
	overriddenField *compile.FieldSpec,
	overriddenIdentifier string,
	indent string,
	prevKeyPrefixes []string,
) error {
	var (
		typeName           string
		toTypeCast         string
		overrideTypeCast   string
		defaultAssignment  *FieldAssignment
		overrideAssignment *FieldAssignment
		err                error
	)
	// resolve type
	if _, ok := toField.Type.(*compile.TypedefSpec); ok {
		typeName, err = c.getIdentifierName(toField.Type)
	} else {
		typeName, err = c.getGoTypeName(toField.Type)
	}
	if err != nil {
		return err
	}
	// set assignment
	if toField.Required {
		toTypeCast = typeName
		overrideTypeCast = typeName
		if !fromField.Required {
			toTypeCast = "*"
		}
		if overriddenField != nil {
			if !overriddenField.Required {
				overrideTypeCast = "*"
			} else {
				overriddenField.Required = true
			}
		}
	} else {
		toTypeCast = fmt.Sprintf("(*%s)", typeName)
		overrideTypeCast = fmt.Sprintf("(*%s)", typeName)
		if fromField.Required {
			toTypeCast = fmt.Sprintf("(*%s)&", typeName)
		}
		if overriddenField != nil && overriddenField.Required {
			overrideTypeCast = fmt.Sprintf("(*%s)&", typeName)
			overriddenField.Required = true
		}
	}
	defaultAssignment = &FieldAssignment{
		From:           fromField,
		FromIdentifier: fromIdentifier,
		To:             toField,
		ToIdentifier:   toIdentifier,
		TypeCast:       toTypeCast,
	}
	if overriddenField != nil {
		overrideAssignment = &FieldAssignment{
			From:           overriddenField,
			FromIdentifier: overriddenIdentifier,
			To:             toField,
			ToIdentifier:   toIdentifier,
			TypeCast:       overrideTypeCast,
		}
	}
	// generate assignment
	c.assignWithOverride(indent, defaultAssignment, overrideAssignment, prevKeyPrefixes)
	return nil
}

func (c *TypeConverter) updateFieldMap(currentMap map[string]FieldMapperEntry, toPrefix string) map[string]FieldMapperEntry {
	newMap := make(map[string]FieldMapperEntry)
	for key, value := range currentMap {
		if strings.HasPrefix(key, toPrefix) && len(key) > len(toPrefix) {
			newMap[key[len(toPrefix)+1:]] = value
		}
	}
	return newMap
}

// Generate a wrapper for genConverterForStruct so its generated code can be re-used later
func (c *TypeConverter) genConverterForStructWrapped(
	toField *compile.FieldSpec,
	toFieldType *compile.StructSpec,
	toTypeName string,
	toSubIdentifier string,
	fromTypeName string,
	fromIdentifier string,
	stFromType compile.TypeSpec,
	fieldMap map[string]FieldMapperEntry,
	prevKeyPrefixes []string,
	indent string,
) error {
	converterMethodName := c.makeUniqIdentifier("convert" + toFieldType.Name + "Helper")
	c.convStructMap[toFieldType.Name] = converterMethodName

	funcSignature := "func(in *" + fromTypeName + ") (out *" + toTypeName + ")"

	c.append(indent, "var ", converterMethodName, " ", funcSignature) // function declaration
	c.append(indent, converterMethodName, " = ", funcSignature, " {") // begin function definition

	err := c.genConverterForStruct(
		toField.Name,
		toFieldType,
		toField.Required,
		stFromType,
		"in",
		"",
		"",
		indent+"\t",
		c.updateFieldMap(fieldMap, toSubIdentifier),
		append(prevKeyPrefixes, toSubIdentifier),
	)
	if err != nil {
		return err
	}

	c.append(indent+"\t", "return")
	c.append(indent, "}") // end of function definition

	c.append(indent, "out.", toSubIdentifier, " = ", converterMethodName, "(", fromIdentifier, ")")

	delete(c.convStructMap, toFieldType.Name)

	return nil
}

// Helper function to detect a recursive struct, looking at its fields, fields of its fields, etc. to see if
// the same type is encountered more than once down a path
func isRecursiveStruct(spec compile.TypeSpec, seenSoFar map[string]bool) bool {
	switch t := spec.(type) {
	case *compile.StructSpec:
		// detected cycle; second time seeing this type
		if _, found := seenSoFar[t.Name]; found {
			return true
		}

		// mark this type as seen
		seenSoFar[t.Name] = true

		// search all fields of this struct
		for _, field := range t.Fields {
			if isRecursiveStruct(field.Type, seenSoFar) {
				return true
			}
		}

		// unmark
		delete(seenSoFar, t.Name)

	// for lists and maps, check element/key types the same way
	case *compile.MapSpec:
		if isRecursiveStruct(t.KeySpec, seenSoFar) || isRecursiveStruct(t.ValueSpec, seenSoFar) {
			return true
		}
	case *compile.ListSpec:
		if isRecursiveStruct(t.ValueSpec, seenSoFar) {
			return true
		}
	}

	return false
}

// Returns true if any of the fields of a struct form a cycle anywhere down the line
// e.g. struct A has optional field of type A -> cycle of length 0
//		struct A has optional field of type B; struct B has optional field of type A -> cycle of length 2
// 		struct A has optional field of type B; struct B has optional field of type B -> cycle of length 0 downstream
func (c *TypeConverter) isRecursiveStruct(fields []*compile.FieldSpec) bool {
	for _, field := range fields {
		if isRecursiveStruct(field.Type, make(map[string]bool)) {
			return true
		}
	}
	return false
}

func trimAnyPrefix(str string, prefixes ...string) string {
	for _, p := range prefixes {
		if strings.HasPrefix(str, p) {
			str = strings.TrimPrefix(str, p)
			break
		}
	}
	return str
}
