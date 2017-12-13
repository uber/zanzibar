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
	Helper        PackageNameResolver
	uninitialized map[string]*fieldStruct
	fieldCounter  int

	requestTypeHelper RequestHelper
}

// RequestHelper is the helper struct for method generation
type RequestHelper struct {
	RequestSuffix     string
	RequestInputType  string
	RequestOutputType string
	ResponseType      string
	OutputMethodName  string
}

// NewTypeConverter returns *TypeConverter
func NewTypeConverter(h PackageNameResolver, requestType RequestHelper) *TypeConverter {
	return &TypeConverter{
		LineBuilder:       LineBuilder{},
		Helper:            h,
		uninitialized:     make(map[string]*fieldStruct),
		requestTypeHelper: requestType,
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
	toRequired bool,
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

	// if no fromFieldType assume we're constructing from transform fieldMap  TODO: make this less subtle
	if fromFieldType == nil {
		//  in the direct assignment we do a nil check on fromField here. its hard for unknown number of transforms
		//  initialize the toField with an empty struct only if it's required, otherwise send to uninitialized map
		if toRequired {
			c.append(indent, toIdentifier, " = &", typeName, "{}")
		} else {
			c.uninitialized[toIdentifier] = &fieldStruct{
				Identifier: toIdentifier,
				TypeName:   typeName,
			}
		}

		// recursive call
		err = c.genStructConverter(
			keyPrefix+".",
			fromPrefix+".",
			indent,
			nil,
			subToFields,
			fieldMap,
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

	subFromFields := fromFieldStruct.Fields
	err = c.genStructConverter(
		keyPrefix+".",
		fromPrefix+".",
		indent+"\t",
		subFromFields,
		subToFields,
		fieldMap,
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

	sourceListID := ""
	isOverriddenID := ""

	if overriddenIdentifier != "" {
		sourceListID = c.makeUniqIdentifier("sourceList")
		isOverriddenID = c.makeUniqIdentifier("isOverridden")

		// Determine which map (from or overrride) to use
		c.appendf("%s := %s", sourceListID, overriddenIdentifier)
		if isStruct {
			c.appendf("%s := false", isOverriddenID)
		}
		// TODO(sindelar): Verify how optional thrift lists are defined.
		c.appendf("if %s != nil {", fromIdentifier)

		c.appendf("\t%s = %s", sourceListID, fromIdentifier)
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
			toIdentifier, typeName, sourceIdentifier,
		)
	} else {
		c.appendf(
			"%s = make([]%s, len(%s))",
			toIdentifier, typeName, sourceIdentifier,
		)
	}

	indexID := c.makeUniqIdentifier("index")
	valID := c.makeUniqIdentifier("value")
	c.appendf(
		"for %s, %s := range %s {",
		indexID, valID, sourceIdentifier,
	)

	if isStruct {
		nestedIndent := "\t" + indent

		if checkOverride {
			nestedIndent = "\t" + nestedIndent
			c.appendf("\tif %s {", isOverriddenID)
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
			toField.Required,
			fromFieldType.ValueSpec,
			valID,
			keyPrefix+pascalCase(toField.Name)+"["+indexID+"]",
			strings.TrimPrefix(fromIdentifier, "in.")+"["+indexID+"]",
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
				toField.Required,
				overriddenFieldType.ValueSpec,
				valID,
				keyPrefix+pascalCase(toField.Name)+"["+indexID+"]",
				strings.TrimPrefix(overriddenIdentifier, "in.")+"["+indexID+"]",
				nestedIndent,
				nil,
			)
			if err != nil {
				return err
			}
			c.append("\t", "}")
		}
	} else {
		c.appendf(
			"\t%s[%s] = %s(%s)",
			toIdentifier, indexID, typeName, valID,
		)
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

	valueStruct, isStruct := toFieldType.ValueSpec.(*compile.StructSpec)
	sourceIdentifier := fromIdentifier
	_, isStringKey := toFieldType.KeySpec.(*compile.StringSpec)

	if !isStringKey {
		realType := compile.RootTypeSpec(toFieldType.KeySpec)
		switch realType.(type) {
		case *compile.StringSpec:
			keyType, _ := c.getGoTypeName(toFieldType.KeySpec)
			c.appendf(
				"%s = make(map[%s]%s, len(%s))",
				toIdentifier, keyType, typeName, sourceIdentifier,
			)

			keyID := c.makeUniqIdentifier("key")
			valID := c.makeUniqIdentifier("value")
			c.appendf(
				"for %s, %s := range %s {",
				keyID, valID, sourceIdentifier,
			)
			c.appendf(
				"\t %s[%s(%s)] = %s(%s)",
				toIdentifier, keyType, keyID, typeName, valID,
			)
			c.append("}")
			return nil
		default:
			return errors.Errorf(
				"could not convert key (%s), map is not string-keyed.", toField.Name)
		}
	}

	checkOverride := false
	sourceListID := ""
	isOverriddenID := ""
	if overriddenIdentifier != "" {
		sourceListID = c.makeUniqIdentifier("sourceList")
		isOverriddenID = c.makeUniqIdentifier("isOverridden")

		// Determine which map (from or overrride) to use
		c.appendf("%s := %s", sourceListID, overriddenIdentifier)

		if isStruct {
			c.appendf("%s := false", isOverriddenID)
		}
		// TODO(sindelar): Verify how optional thrift map are defined.
		c.appendf("if %s != nil {", fromIdentifier)

		c.appendf("\t%s = %s", sourceListID, fromIdentifier)
		if isStruct {
			c.appendf("\t%s = true", isOverriddenID)
		}
		c.append("}")

		sourceIdentifier = sourceListID
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

	keyID := c.makeUniqIdentifier("key")
	valID := c.makeUniqIdentifier("value")
	c.appendf(
		"for %s, %s := range %s {",
		keyID, valID, sourceIdentifier,
	)

	if isStruct {
		nestedIndent := "\t" + indent

		if checkOverride {
			nestedIndent = "\t" + nestedIndent
			c.appendf("\tif %s {", isOverriddenID)
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
			toField.Required,
			fromFieldType.ValueSpec,
			valID,
			keyPrefix+pascalCase(toField.Name)+"["+keyID+"]",
			strings.TrimPrefix(fromIdentifier, "in.")+"["+keyID+"]",
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
				toField.Required,
				overriddenFieldType.ValueSpec,
				valID,
				keyPrefix+pascalCase(toField.Name)+"["+keyID+"]",
				strings.TrimPrefix(overriddenIdentifier, "in.")+"["+keyID+"]",
				nestedIndent,
				nil,
			)
			if err != nil {
				return err
			}
			c.append("\t", "}")
		}
	} else {
		c.appendf(
			"\t%s[%s] = %s(%s)",
			toIdentifier, keyID, typeName, valID,
		)
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

		toSubIdentifier := keyPrefix + pascalCase(toField.Name)
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
				fromIdentifier = "in." + transformFrom.QualifiedName
				// else there is a conflicting direct fromField
			} else {
				//  depending on Override flag either the direct fromField or transformFrom is the OverrideField
				if transformFrom.Override {
					// check for required/optional setting
					if !transformFrom.Field.Required {
						overriddenField = fromField
						overriddenIdentifier = "in." + fromPrefix +
							pascalCase(overriddenField.Name)
					}
					// If override is true and the new field is required,
					// there's a default instantiation value and will always
					// overwrite.
					fromField = transformFrom.Field
					fromIdentifier = "in." + transformFrom.QualifiedName
				} else {
					// If override is false and the from field is required,
					// From is always populated and will never be overwritten.
					if !fromField.Required {
						overriddenField = transformFrom.Field
						overriddenIdentifier = "in." + transformFrom.QualifiedName
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
				if toField.Required {
					// unrecoverable error
					return errors.Errorf(
						"required toField %s does not have a valid fromField mapping",
						toField.Name,
					)
				}
				// the toField is optional and there's nothing that maps to it or its sub-fields so we should skip it
				continue
			}
		}

		if fromIdentifier == "" && fromField != nil {
			// should we set this if no fromField ??
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
			*compile.StringSpec,
			*compile.TypedefSpec:

			err := c.genConverterForPrimitiveOrTypedef(
				toField,
				toIdentifier,
				fromField,
				fromIdentifier,
				overriddenField,
				overriddenIdentifier,
				indent,
			)
			if err != nil {
				return err
			}
		case *compile.BinarySpec:
			// TODO: handle override. Check if binarySpec can be optional.
			c.append(toIdentifier, " = []byte(", fromIdentifier, ")")

		case *compile.StructSpec:
			var (
				stFromPrefix = keyPrefix
				stFromType   compile.TypeSpec
			)
			if fromField != nil {
				stFromType = fromField.Type
				stFromPrefix = keyPrefix + pascalCase(fromField.Name)
			}
			err := c.genConverterForStruct(
				toField.Name,
				toFieldType,
				toField.Required,
				stFromType,
				fromIdentifier,
				keyPrefix+pascalCase(toField.Name),
				stFromPrefix,
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
// destination fields (sent to the downstream client) and the entries are source
// fields (from the incoming request)
func (c *TypeConverter) GenStructConverter(
	fromFields []*compile.FieldSpec,
	toFields []*compile.FieldSpec,
	fieldMap map[string]FieldMapperEntry,
	isPrimitive bool,
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

	helper := c.requestTypeHelper
	var requestInput, requestOutput string
	requestInput = "(in " + helper.RequestInputType + ") " + helper.RequestOutputType
	requestOutput = "out := &" + helper.ResponseType + "{}\n"

	c.append("func convertTo", pascalCase(helper.OutputMethodName), c.requestTypeHelper.RequestSuffix,
		requestInput, "{")

	if isPrimitive {
		c.append("out", " := in\t\n")
		c.append("\nreturn out \t\n}")
		return nil
	}

	c.append(requestOutput)
	err := c.genStructConverter("", "", "", fromFields, toFields, fieldMap)
	if err != nil {
		return err
	}
	c.append("\nreturn out")
	c.append("}")
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
			fieldQualName := prefix + pascalCase(spec.Name)
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
	var (
		in  = len("in.")
		out = len("out.")
	)
	if f.FromIdentifier[:in] != "in." || f.ToIdentifier[:out] != "out." {
		panic("isTransform check called on unexpected input  fromIdentifer " + f.FromIdentifier + " toIdentifer " + f.ToIdentifier)
	}
	return f.FromIdentifier[in:] != f.ToIdentifier[out:]
}

// checkOptionalNil nil-checks fields that may not be initialized along
// the assign path
func checkOptionalNil(
	indent string,
	uninitialized map[string]*fieldStruct,
	toIdentifier string,
) []string {
	var ret = make([]string, 0)
	if len(uninitialized) < 1 {
		return ret
	}
	keys := make([]string, 0, len(uninitialized))
	for k := range uninitialized {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, id := range keys {
		if strings.HasPrefix(toIdentifier, id) {
			v := uninitialized[id]
			ret = append(ret, indent+"if "+v.Identifier+" == nil {")
			ret = append(ret, indent+"\t"+v.Identifier+" = &"+v.TypeName+"{}")
			ret = append(ret, indent+"}")
		}
	}
	return ret
}

// Generate generates assignment
func (f *FieldAssignment) Generate(indent string, uninitialized map[string]*fieldStruct) string {
	var lines = make([]string, 0)
	if !f.IsTransform() {
		//  do an extra nil check for Overrides
		if (!f.From.Required && f.To.Required) || f.ExtraNilCheck {
			// need to nil check otherwise we could be cast or deref nil
			lines = append(lines, indent+"if "+f.FromIdentifier+" != nil {")
			lines = append(lines, checkOptionalNil(indent+"\t", uninitialized, f.ToIdentifier)...)
			lines = append(lines, indent+"\t"+f.ToIdentifier+" = "+f.TypeCast+"("+f.FromIdentifier+")")
			lines = append(lines, indent+"}")
		} else {
			lines = append(lines, checkOptionalNil(indent, uninitialized, f.ToIdentifier)...)
			lines = append(lines, indent+f.ToIdentifier+" = "+f.TypeCast+"("+f.FromIdentifier+")")
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
			lines = append(lines, checkOptionalNil(indent, uninitialized, f.ToIdentifier)...)
			lines = append(lines, indent+f.ToIdentifier+" = "+f.TypeCast+"("+f.FromIdentifier+")")
			return strings.Join(lines, "\n")
		}
	}
	lines = append(lines, indent+"if "+strings.Join(checks, " && ")+" {")
	lines = append(lines, checkOptionalNil(indent+"\t", uninitialized, f.ToIdentifier)...)
	lines = append(lines, indent+"\t"+f.ToIdentifier+" = "+f.TypeCast+"("+f.FromIdentifier+")")
	lines = append(lines, indent+"}")
	// TODO else?  log? should this silently eat intermediate nils as none-assignment,  should set nil?
	return strings.Join(lines, "\n")
}

func (c *TypeConverter) assignWithOverride(
	indent string,
	defaultAssign *FieldAssignment,
	overrideAssign *FieldAssignment,
) {
	if overrideAssign != nil && overrideAssign.FromIdentifier != "" {
		c.append(overrideAssign.Generate(indent, c.uninitialized))
		defaultAssign.ExtraNilCheck = true
		c.append(defaultAssign.Generate(indent, c.uninitialized))
		return
	}
	c.append(defaultAssign.Generate(indent, c.uninitialized))
}

func (c *TypeConverter) genConverterForPrimitiveOrTypedef(
	toField *compile.FieldSpec,
	toIdentifier string,
	fromField *compile.FieldSpec,
	fromIdentifier string,
	overriddenField *compile.FieldSpec,
	overriddenIdentifier string,
	indent string,
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
	// generate assignement
	c.assignWithOverride(indent, defaultAssignment, overrideAssignment)
	return nil
}
