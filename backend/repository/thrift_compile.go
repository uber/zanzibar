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

package repository

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/ast"
	"go.uber.org/thriftrw/compile"
	"go.uber.org/thriftrw/idl"
)

// Module represents a compiled Thrift module. In contrast to thriftrw-go's
// compile.Module, all fields of this Module are struct and hence can be
// serialized and deserialized.
type Module struct {
	Name       string `json:"name"`
	ThriftPath string `json:"thrift_path"`

	// Mapping from the /Thrift name/ to the compiled representation of
	// different definitions.
	Namespace []*Namespace               `json:"namespace,omitempty"`
	Includes  map[string]*IncludedModule `json:"includes,omitempty"`
	Constants map[string]*Constant       `json:"constants,omitempty"`
	Types     map[string]*TypeSpec       `json:"types,omitempty"`
	Services  map[string]*ServiceSpec    `json:"services,omitempty"`
}

// Namespace statements allow users to choose the package name used by the
// generated code in certain languages.
//
// namespace py foo.bar
type Namespace struct {
	Scope string `json:"scope"`
	Name  string `json:"name"`
	Line  int    `json:"line"`
}

// IncludedModule represents an included module.
type IncludedModule struct {
	Name   string  `json:"name"`
	Line   int     `json:"line"`
	Module *Module `json:"module,omitempty"`
}

// Constant represents a single named constant value from the Thrift file.
type Constant struct {
	Name string    `json:"name"`
	File string    `json:"file"`
	Type *TypeSpec `json:"type"`
	// value in string format, such as "true" and "15"
	Value string `json:"value"`
	Line  int    `json:"line"`
}

// TypeSpec contains information about thrift types.
type TypeSpec struct {
	Name string `json:"name"`
	// empty file for built-in types
	File        string              `json:"file,omitempty"`
	Annotations compile.Annotations `json:"annotations,omitempty"`
	// The following fields defines specific types other than the
	// built-in types. At most one of the field can not be nil.
	StructType    *StructTypeSpec `json:"struct_type,omitempty"`
	EnumType      *EnumSpec       `json:"enum_type,omitempty"`
	MapType       *MapTypeSpec    `json:"map_type,omitempty"`
	ListValueType *TypeSpec       `json:"list_value_type,omitempty"`
	SetValueType  *TypeSpec       `json:"set_value_type,omitempty"`
	TypeDefTarget *TypeSpec       `json:"type_def_target,omitempty"`
	Line          int             `json:"line,omitempty"`
}

// EnumSpec represents an enum defined in the Thrift file.
type EnumSpec struct {
	Name        string              `json:"name"`
	File        string              `json:"file"`
	Items       []EnumItem          `json:"items,omitempty"`
	Annotations compile.Annotations `json:"annotations,omitempty"`
}

// EnumItem is a single item inside an enum.
type EnumItem struct {
	Name        string              `json:"name"`
	Value       int32               `json:"value,omitempty"`
	Annotations compile.Annotations `json:"annotations,omitempty"`
}

// StructTypeSpec defines the type information of a struct.
type StructTypeSpec struct {
	Kind   StructureType `json:"kind"`
	Fields []*FieldSpec  `json:"fields"`
}

// StructureType indicates what type the struct is.
type StructureType string

const (
	// StructureTypeStruct indicates a struct type.
	StructureTypeStruct StructureType = "struct"
	// StructureTypeUnion indicates an union type.
	StructureTypeUnion StructureType = "union"
	// StructureTypeException indicates an exception type.
	StructureTypeException StructureType = "exception"
	// StructureTypeUnknown indicates an unknown type.
	StructureTypeUnknown StructureType = "unknown"
)

// MapTypeSpec defines TypeSpec for a map.
type MapTypeSpec struct {
	KeyType   *TypeSpec `json:"key_type"`
	ValueType *TypeSpec `json:"value_type"`
}

// ServiceSpec is a collection of named functions.
type ServiceSpec struct {
	Name        string                   `json:"name"`
	File        string                   `json:"file"`
	Functions   map[string]*FunctionSpec `json:"functions,omitempty"`
	Annotations compile.Annotations      `json:"annotations,omitempty"`
	Line        int                      `json:"line,omitempty"`
}

// FunctionSpec is a single function inside a Service.
type FunctionSpec struct {
	Name     string       `json:"name"`
	ArgsSpec []*FieldSpec `json:"args_spec,omitempty"`
	// nil if OneWay is true
	ResultSpec  *ResultSpec         `json:"result_spec,omitempty"`
	OneWay      bool                `json:"one_way"`
	Annotations compile.Annotations `json:"annotations,omitempty"`
	Line        int                 `json:"line,omitempty"`
}

// FieldSpec represents a single field of a struct or parameter list.
type FieldSpec struct {
	ID          int16               `json:"id"`
	Name        string              `json:"name"`
	Type        *TypeSpec           `json:"type"`
	Required    bool                `json:"required"`
	Default     string              `json:"default,omitempty"`
	Annotations compile.Annotations `json:"annotations,omitempty"`
	Line        int                 `json:"line,omitempty"`
}

// ResultSpec contains information about a Function's result type.
type ResultSpec struct {
	ReturnType *TypeSpec    `json:"return_type"`
	Exceptions []*FieldSpec `json:"exceptions,omitempty"`
}

// CompileThriftFile compiles a thrift file in a structured module.
func CompileThriftFile(thriftAbsPath string, thriftRootPath string) (*Module, error) {
	b, err := ioutil.ReadFile(thriftAbsPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file %q", thriftAbsPath)
	}
	return CompileThriftCode(b, thriftAbsPath, thriftRootPath)
}

// CompileThriftCode compiles thrift code into a structured module.
func CompileThriftCode(rawCode []byte, thriftAbsPath string, thriftRootPath string) (*Module, error) {
	thriftFile := relPath(thriftRootPath, thriftAbsPath)
	thriftFileName := strings.Split(filepath.Base(thriftFile), ".")[0]

	program, err := idl.Parse(rawCode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse thrift program")
	}

	var namespaces []*Namespace
	includedModules := make(map[string]*IncludedModule)

	for _, header := range program.Headers {
		switch header := header.(type) {
		case *ast.Include:
			targetPath := filepath.Join(filepath.Dir(thriftAbsPath), header.Path)
			incModuleDetails, err := CompileThriftFile(targetPath, thriftRootPath)
			if err == nil {
				includedModules[incModuleDetails.Name] = &IncludedModule{
					Name:   incModuleDetails.Name,
					Module: incModuleDetails,
					Line:   header.Line,
				}
			}
		case *ast.Namespace:
			namespaces = append(namespaces,
				&Namespace{
					Name:  header.Name,
					Scope: header.Scope,
					Line:  header.Line,
				})
		}
	}

	constants := make(map[string]*Constant)
	types := make(map[string]*TypeSpec)
	services := make(map[string]*ServiceSpec)

	for _, definition := range program.Definitions {
		switch definition := definition.(type) {
		case *ast.Constant:
			constants[definition.Name] = &Constant{
				Name:  definition.Name,
				File:  thriftFile,
				Line:  definition.Line,
				Type:  convertAstSubType(definition.Type),
				Value: constantAstValue(definition.Value),
			}
		case *ast.Typedef:
			types[definition.Name] = &TypeSpec{
				Name:          definition.Name,
				File:          thriftFile,
				Line:          definition.Line,
				TypeDefTarget: convertAstSubType(definition.Type),
				Annotations:   convertAstAnnotations(definition.Annotations),
			}
		case *ast.Enum:
			types[definition.Name] = &TypeSpec{
				Name: definition.Name,
				File: thriftFile,
				Line: definition.Line,
				EnumType: &EnumSpec{
					Name:  definition.Name,
					File:  thriftFile,
					Items: convertAstEnumItems(definition.Items),
				},
			}
		case *ast.Struct:
			types[definition.Name] = &TypeSpec{
				Name: definition.Name,
				File: thriftFile,
				Line: definition.Line,
				StructType: &StructTypeSpec{
					Kind:   convertStructType(definition.Type),
					Fields: convertAstFields(definition.Fields),
				},
				Annotations: convertAstAnnotations(definition.Annotations),
			}
		case *ast.Service:
			services[definition.Name] = &ServiceSpec{
				Name:        definition.Name,
				File:        thriftFile,
				Line:        definition.Line,
				Functions:   convertAstFunctions(definition.Functions),
				Annotations: convertAstAnnotations(definition.Annotations),
			}
		default:
			types[definition.Info().Name] = &TypeSpec{
				Name: definition.Info().Name,
				File: thriftFile,
				Line: definition.Info().Line,
			}
		}
	}

	m := &Module{
		Name:       thriftFileName,
		ThriftPath: thriftFile,
		Namespace:  namespaces,
		Includes:   includedModules,
		Types:      types,
		Constants:  constants,
		Services:   services,
	}

	return m, nil
}

func convertAstSubType(t ast.Type) *TypeSpec {
	if t == nil {
		return nil
	}
	ts := &TypeSpec{
		Name: t.String(),
	}
	switch t := t.(type) {
	case ast.MapType:
		ts.MapType = &MapTypeSpec{
			KeyType:   convertAstSubType(t.KeyType),
			ValueType: convertAstSubType(t.ValueType),
		}
	case ast.ListType:
		ts.ListValueType = convertAstSubType(t.ValueType)
	case ast.SetType:
		ts.SetValueType = convertAstSubType(t.ValueType)
	}
	return ts
}

func convertAstFunctions(functions []*ast.Function) map[string]*FunctionSpec {
	if functions == nil {
		return nil
	}

	funcMap := make(map[string]*FunctionSpec)
	for _, function := range functions {
		newFunc := &FunctionSpec{
			Name:        function.Name,
			ArgsSpec:    convertAstFields(function.Parameters),
			OneWay:      function.OneWay,
			Annotations: convertAstAnnotations(function.Annotations),
			Line:        function.Line,
		}
		results := createResultSpec(function)
		if results != nil {
			newFunc.ResultSpec = results
		}
		funcMap[function.Name] = newFunc
	}

	return funcMap
}

func createResultSpec(function *ast.Function) *ResultSpec {
	if function.Exceptions == nil && function.ReturnType == nil {
		return nil
	}
	results := &ResultSpec{}
	if function.Exceptions != nil {
		results.Exceptions = convertAstFields(function.Exceptions)
	}

	if function.ReturnType != nil {
		results.ReturnType = convertAstSubType(function.ReturnType)
	}

	return results
}

func convertStructType(kind ast.StructureType) StructureType {
	if kind == ast.StructType {
		return "struct"
	}

	return "exception"
}

func convertAstFields(fields []*ast.Field) []*FieldSpec {
	if fields == nil {
		return nil
	}

	var outputFields []*FieldSpec
	for _, field := range fields {
		outputFields = append(outputFields, &FieldSpec{
			ID:          int16(field.ID),
			Name:        field.Name,
			Default:     constantAstValue(field.Default),
			Required:    convertRequiredness(field.Requiredness),
			Type:        convertAstSubType(field.Type),
			Annotations: convertAstAnnotations(field.Annotations),
			Line:        field.Line,
		})
	}

	return outputFields
}

func convertRequiredness(isRequired ast.Requiredness) bool {
	if isRequired == ast.Required {
		return true
	}

	return false
}

func convertAstEnumItems(items []*ast.EnumItem) []EnumItem {
	if items == nil {
		return nil
	}

	var enumItems []EnumItem
	for _, item := range items {
		newItem := EnumItem{
			Name:        item.Name,
			Annotations: convertAstAnnotations(item.Annotations),
		}
		if item.Value != nil {
			newItem.Value = int32(*item.Value)
		}
		enumItems = append(enumItems, newItem)
	}

	return enumItems
}

func convertAstAnnotations(annotations []*ast.Annotation) compile.Annotations {
	if annotations == nil {
		return nil
	}

	am := make(compile.Annotations)
	for _, annotation := range annotations {
		am[annotation.Name] = annotation.Value
	}

	return am
}

func constantAstValue(value ast.ConstantValue) string {
	if value == nil {
		return ""
	}
	switch t := value.(type) {
	case ast.ConstantBoolean:
		return fmt.Sprintf("%t", t)
	case ast.ConstantInteger:
		return fmt.Sprintf("%d", t)
	case ast.ConstantString:
		return fmt.Sprintf("%q", string(t))
	case ast.ConstantDouble:
		return fmt.Sprintf("%f", t)
	case ast.ConstantReference:
		return fmt.Sprintf("%s", t.Name)
	}
	// TODO: Add complete type converstions and make this message panic.
	return fmt.Sprintf("unknown constant value type: %v", value)
}

func relPath(basePath, targetPath string) string {
	rel, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return targetPath
	}
	return rel
}

func includedName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

// Code converts the module as normal thrift code.
func (m *Module) Code() string {
	codeBlocks := make([]*CodeBlock, 0, 100)
	for _, ns := range m.Namespace {
		codeBlocks = append(codeBlocks, ns.CodeBlock())
	}
	for _, im := range m.Includes {
		codeBlocks = append(codeBlocks, im.CodeBlock(m.ThriftPath))
	}
	for _, c := range m.Constants {
		codeBlocks = append(codeBlocks, c.CodeBlock(m.ThriftPath))
	}
	for _, t := range m.Types {
		codeBlocks = append(codeBlocks, t.CodeBlock(m.ThriftPath))
	}
	for _, service := range m.Services {
		codeBlocks = append(codeBlocks, service.CodeBlock(m.ThriftPath))
	}
	sort.Sort(codeBlockSlice(codeBlocks))
	result := bytes.NewBuffer(nil)
	for i := range codeBlocks {
		result.WriteString(codeBlocks[i].Code)
		result.WriteString("\n\n")
	}
	return result.String()
}

// CodeBlock defines a code block of a thrift file.
type CodeBlock struct {
	Code  string
	Order int
}

// CodeBlock converts a namespace to a CodeBlock.
func (ns *Namespace) CodeBlock() *CodeBlock {
	return &CodeBlock{
		Code:  fmt.Sprintf("namespace %s %s", ns.Scope, ns.Name),
		Order: ns.Line,
	}
}

// CodeBlock converts a included module to a CodeBlock.
func (im *IncludedModule) CodeBlock(curFilePath string) *CodeBlock {
	relPath, err := filepath.Rel(curFilePath, im.Module.ThriftPath)
	if err != nil {
		relPath = curFilePath
	}
	relPath = filepath.Clean(strings.TrimPrefix(relPath, "../"))
	return &CodeBlock{
		Code:  fmt.Sprintf("include \"%s\"", relPath),
		Order: im.Line,
	}
}

// CodeBlock converts a constant to a CodeBlock.
func (c *Constant) CodeBlock(curFilePath string) *CodeBlock {
	return &CodeBlock{
		Code: fmt.Sprintf("const %s %s = %s%s", c.Type.FullTypeName(curFilePath), c.Name, c.Value,
			annotationsCode(c.Type.Annotations, "\t", "")),
		Order: c.Line,
	}
}

// CodeBlock converts a typeSpec to a CodeBlock.
func (ts *TypeSpec) CodeBlock(curFilePath string) *CodeBlock {
	// Typedef definition
	if ts.TypeDefTarget != nil {
		return &CodeBlock{
			Code: fmt.Sprintf("typedef %s %s%s",
				ts.TypeDefTarget.FullTypeName(curFilePath), ts.Name,
				annotationsCode(ts.Annotations, "/t", "")),
			Order: ts.Line,
		}
	}
	// Enum definition
	if ts.EnumType != nil {
		return enumTypeCodeBlock(ts.EnumType, ts.Line, ts.Annotations)
	}
	// Struct definition
	if ts.StructType != nil {
		return structTypeCodeBlock(ts.StructType, ts.Name, ts.Line, curFilePath, ts.Annotations)
	}
	panic(fmt.Sprintf(
		"No typedef, enum, nor struct definition found: typespec %+v", ts))
}

func enumTypeCodeBlock(e *EnumSpec, line int, annotations compile.Annotations) *CodeBlock {
	result := bytes.NewBuffer(nil)
	result.WriteString(fmt.Sprintf("enum %s {\n", e.Name))
	items := e.Items
	for i := 0; i < len(items)-1; i++ {
		result.WriteString(fmt.Sprintf("\t%s = %d%s,\n", items[i].Name, items[i].Value,
			annotationsCode(items[i].Annotations, "\t\t", "\t")))
	}
	if l := len(items); l > 0 {
		result.WriteString(fmt.Sprintf("\t%s = %d\n%s", items[l-1].Name, items[l-1].Value,
			annotationsCode(items[l-1].Annotations, "\t\t", "\t")))
	}
	result.WriteString("}" + annotationsCode(annotations, "\t", ""))
	return &CodeBlock{
		Code:  result.String(),
		Order: line,
	}
}

func structTypeCodeBlock(st *StructTypeSpec, name string, line int, curFilePath string, annotations compile.Annotations) *CodeBlock {
	result := bytes.NewBuffer(nil)
	result.WriteString(fmt.Sprintf("%s %s {\n", string(st.Kind), name))
	result.WriteString(fieldsCode(st.Fields, curFilePath, "\t", true))
	result.WriteString("}" + annotationsCode(annotations, "\t", ""))
	return &CodeBlock{
		Code:  result.String(),
		Order: line,
	}
}

func fieldsCode(fields []*FieldSpec, curFilePath, lineIndent string, showOptional bool) string {
	sort.Sort(fieldSlice(fields))
	result := bytes.NewBuffer(nil)
	for _, field := range fields {
		var required string
		if field.Required {
			required = " required"
		} else if showOptional {
			required = " optional"
		}
		line := fmt.Sprintf("%s%d:%s %s %s",
			lineIndent, field.ID, required, field.Type.FullTypeName(curFilePath), field.Name)
		if field.Default != "" {
			line += " = " + field.Default
		}
		line += annotationsCode(field.Annotations, lineIndent+"\t", lineIndent) + "\n"
		result.WriteString(line)
	}
	return result.String()
}

// FullTypeName defines the full name referred in current thrift file,
// such as "string", "int", "base.abc".
func (ts *TypeSpec) FullTypeName(curFilePath string) string {
	name := ts.Name
	if ts.ListValueType != nil {
		name = fmt.Sprintf("list<%s>", ts.ListValueType.FullTypeName(curFilePath))
	} else if ts.MapType != nil {
		name = fmt.Sprintf("map<%s, %s>",
			ts.MapType.KeyType.FullTypeName(curFilePath), ts.MapType.ValueType.FullTypeName(curFilePath))
	} else if ts.SetValueType != nil {
		name = fmt.Sprintf("set<%s>", ts.SetValueType.FullTypeName(curFilePath))
	}
	// Built-in type or type defined in current file.
	if ts.File == "" || ts.File == curFilePath {
		return name
	}
	return includedName(ts.File) + "." + name
}

// CodeBlock converts a serviceSpec to a CodeBlock.
func (ss *ServiceSpec) CodeBlock(curFilePath string) *CodeBlock {
	result := bytes.NewBuffer(nil)
	result.WriteString(fmt.Sprintf("service %s {\n", ss.Name))
	functions := make([]*FunctionSpec, 0, len(ss.Functions))
	for _, f := range ss.Functions {
		functions = append(functions, f)
	}
	result.WriteString(functionsCode(functions, curFilePath))
	result.WriteString("}")
	return &CodeBlock{
		Code:  result.String(),
		Order: ss.Line,
	}
}

func functionsCode(functions []*FunctionSpec, curFilePath string) string {
	sort.Sort(functionSlice(functions))
	result := bytes.NewBuffer(nil)
	for i, f := range functions {
		var returnType string
		if f.ResultSpec.ReturnType == nil {
			returnType = "void"
		} else {
			returnType = f.ResultSpec.ReturnType.FullTypeName(curFilePath)
		}
		result.WriteString(fmt.Sprintf("\t%s %s (", returnType, f.Name))
		// Input parameters
		if len(f.ArgsSpec) != 0 {
			result.WriteString(fmt.Sprintf("\n%s\t)",
				fieldsCode(f.ArgsSpec, curFilePath, "\t\t", true)))
		} else {
			result.WriteString(")")
		}
		// Exceptions
		if exceptions := f.ResultSpec.Exceptions; len(exceptions) != 0 {
			result.WriteString(" throws (\n")
			result.WriteString(fieldsCode(exceptions, curFilePath, "\t\t", false))
			result.WriteString("\t)")
		}
		// Annotations
		result.WriteString(annotationsCode(f.Annotations, "\t\t", "\t"))

		if i == len(functions)-1 {
			result.WriteString("\n")
		} else {
			result.WriteString("\n\n")
		}
	}
	return result.String()
}

func annotationsCode(annotations compile.Annotations, lineIndent, rightParenthesesIndent string) string {
	if len(annotations) == 0 {
		return ""
	}
	if len(annotations) == 1 {
		for k, v := range annotations {
			return fmt.Sprintf(" ( %s = \"%s\" )", k, v)
		}
	}
	keys := make([]string, 0, len(annotations))
	for key := range annotations {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := bytes.NewBuffer(nil)
	result.WriteString(" (\n")
	for _, key := range keys {
		result.WriteString(fmt.Sprintf("%s%s = \"%s\"\n", lineIndent, key, annotations[key]))
	}
	result.WriteString(rightParenthesesIndent + ")")
	return result.String()
}

type codeBlockSlice []*CodeBlock

// Len returns length.
func (c codeBlockSlice) Len() int {
	return len(c)
}

// Swap swaps two elements.
func (c codeBlockSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Less determines the order.
func (c codeBlockSlice) Less(i, j int) bool {
	return c[i].Order < c[j].Order
}

type fieldSlice []*FieldSpec

// Len returns length.
func (f fieldSlice) Len() int {
	return len(f)
}

// Swap swaps two elements.
func (f fieldSlice) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// Less determines the order.
func (f fieldSlice) Less(i, j int) bool {
	return f[i].Line < f[j].Line || f[i].Line == f[j].Line && f[i].ID < f[j].ID
}

type functionSlice []*FunctionSpec

// Len returns length.
func (f functionSlice) Len() int {
	return len(f)
}

// Swap swaps two elements.
func (f functionSlice) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// Less determines the order.
func (f functionSlice) Less(i, j int) bool {
	return f[i].Line < f[j].Line || f[i].Line == f[j].Line && f[i].Name < f[j].Name
}
