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
	"fmt"
	"path/filepath"

	"github.com/thriftrw/thriftrw-go/compile"
)

// Module represents a compiled Thrift module. In contrast to thriftrw-go's
// compile.Module, all fields of this Module are struct and hence can be
// serialized and deserialized.
type Module struct {
	Name       string `json:"name"`
	ThriftPath string `json:"thrift_path"`

	// Mapping from the /Thrift name/ to the compiled representation of
	// different definitions.

	Includes  map[string]*IncludedModule `json:"includes,omitempty"`
	Constants map[string]*Constant       `json:"constants,omitempty"`
	Types     map[string]*TypeSpec       `json:"types,omitempty"`
	Services  map[string]*ServiceSpec    `json:"services,omitempty"`
}

// IncludedModule represents an included module.
type IncludedModule struct {
	Name   string  `json:"name"`
	Module *Module `json:"module,omitempty"`
}

// Constant represents a single named constant value from the Thrift file.
type Constant struct {
	Name string    `json:"name"`
	File string    `json:"file"`
	Type *TypeSpec `json:"type"`
	// value in string format, such as "true" and "15"
	Value string `json:"value"`
}

// TypeSpec contains information about thrift types.
type TypeSpec struct {
	Name string `json:"name"`
	// empty file for built-in types
	File        string              `json:"file,omitempty"`
	Annotations compile.Annotations `json:"annotations,omitempty"`
	// The following fields defines specific types other than the
	// built-in types. At most one of the field can not be nil.
	StructType    *StructTypeSpec   `json:"struct_type,omitempty"`
	EnumType      *compile.EnumSpec `json:"enum_type,omitempty"`
	MapType       *MapTypeSpec      `json:"map_type,omitempty"`
	ListValueType *TypeSpec         `json:"list_value_type,omitempty"`
	SetValueType  *TypeSpec         `json:"set_value_type,omitempty"`
	TypeDefTarget *TypeSpec         `json:"type_def_target,omitempty"`
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
}

// FunctionSpec is a single function inside a Service.
type FunctionSpec struct {
	Name     string       `json:"name"`
	ArgsSpec []*FieldSpec `json:"args_spec,omitempty"`
	// nil if OneWay is true
	ResultSpec  *ResultSpec         `json:"result_spec,omitempty"`
	OneWay      bool                `json:"one_way"`
	Annotations compile.Annotations `json:"annotations,omitempty"`
}

// FieldSpec represents a single field of a struct or parameter list.
type FieldSpec struct {
	ID          int16               `json:"id"`
	Name        string              `json:"name"`
	Type        *TypeSpec           `json:"type"`
	Required    bool                `json:"required"`
	Default     string              `json:"default,omitempty"`
	Annotations compile.Annotations `json:"annotations,omitempty"`
}

// ResultSpec contains information about a Function's result type.
type ResultSpec struct {
	ReturnType *TypeSpec    `json:"return_type"`
	Exceptions []*FieldSpec `json:"exceptions,omitempty"`
}

// ConvertModule converts a compile.Module into Module.
func ConvertModule(module *compile.Module, basePath string) *Module {
	incModules := includedModules(module.Includes, basePath)
	constants := make(map[string]*Constant)
	for name, c := range module.Constants {
		constants[name] = constant(c, basePath)
	}
	types := make(map[string]*TypeSpec)
	for name, t := range module.Types {
		types[name] = typeSpec(t, basePath)
	}
	serviceSpecs := convertServiceSpec(module.Services, basePath)
	return &Module{
		Name:       module.Name,
		ThriftPath: relPath(basePath, module.ThriftPath),
		Includes:   incModules,
		Constants:  constants,
		Types:      types,
		Services:   serviceSpecs,
	}
}

func relPath(basePath, targetPath string) string {
	rel, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return targetPath
	}
	return rel
}

func includedModules(includes map[string]*compile.IncludedModule, basePath string) map[string]*IncludedModule {
	includedModule := make(map[string]*IncludedModule)
	for name, incModule := range includes {
		includedModule[name] = &IncludedModule{
			Name:   incModule.Name,
			Module: ConvertModule(incModule.Module, basePath),
		}
	}
	return includedModule
}

func constant(c *compile.Constant, basePath string) *Constant {
	if c == nil {
		return nil
	}
	return &Constant{
		Name:  c.Name,
		File:  relPath(basePath, c.File),
		Type:  typeSpec(c.Type, basePath),
		Value: constantValue(c.Value),
	}
}

func constantValue(value compile.ConstantValue) string {
	if value == nil {
		return ""
	}
	switch t := value.(type) {
	case compile.ConstantBool:
		return fmt.Sprintf("%t", t)
	case compile.ConstantInt:
		return fmt.Sprintf("%d", t)
	case compile.ConstantString:
		return string(t)
	case compile.ConstantDouble:
		return fmt.Sprintf("%f", t)
	case compile.EnumItemReference:
		return fmt.Sprintf("%s.%s", t.Enum.Name, t.Item.Name)
	}
	// TODO(zw): Add complete type converstions and make this message panic.
	return fmt.Sprintf("unknown contantant value type: %v", value)
}

func typeSpec(t compile.TypeSpec, basePath string) *TypeSpec {
	if t == nil {
		return nil
	}
	ts := &TypeSpec{
		Name:        t.ThriftName(),
		File:        relPath(basePath, t.ThriftFile()),
		Annotations: t.ThriftAnnotations(),
	}
	switch t := t.(type) {
	case *compile.StructSpec:
		ts.StructType = &StructTypeSpec{
			Kind:   structureType(int(t.Type)),
			Fields: fieldSpecs(t.Fields, basePath),
		}

		return ts
	case *compile.EnumSpec:
		t.File = relPath(basePath, t.File)
		ts.EnumType = t
		return ts
	case *compile.TypedefSpec:
		ts.TypeDefTarget = typeSpec(t.Target, basePath)
		return ts
	case *compile.MapSpec:
		ts.MapType = &MapTypeSpec{
			KeyType:   typeSpec(t.KeySpec, basePath),
			ValueType: typeSpec(t.ValueSpec, basePath),
		}
		return ts
	case *compile.ListSpec:
		ts.ListValueType = typeSpec(t.ValueSpec, basePath)
		return ts
	case *compile.SetSpec:
		ts.SetValueType = typeSpec(t.ValueSpec, basePath)
		return ts
	}
	return ts
}

func structureType(i int) StructureType {
	switch i {
	case 1:
		return StructureTypeStruct
	case 2:
		return StructureTypeUnion
	case 3:
		return StructureTypeException
	}
	return StructureTypeUnknown
}

func convertServiceSpec(services map[string]*compile.ServiceSpec, basePath string) map[string]*ServiceSpec {
	specs := make(map[string]*ServiceSpec)
	for name, spec := range services {
		functions := make(map[string]*FunctionSpec)
		for fname, funcSpec := range spec.Functions {
			functions[fname] = functionSpec(funcSpec, basePath)
		}
		specs[name] = &ServiceSpec{
			Name:        spec.Name,
			File:        relPath(basePath, spec.File),
			Functions:   functions,
			Annotations: spec.Annotations,
		}
	}
	return specs
}
func functionSpec(fs *compile.FunctionSpec, basePath string) *FunctionSpec {
	if fs == nil {
		return nil
	}
	funcSpec := &FunctionSpec{
		Name:        fs.Name,
		OneWay:      fs.OneWay,
		Annotations: fs.Annotations,
	}
	funcSpec.ArgsSpec = fieldSpecs(compile.FieldGroup(fs.ArgsSpec), basePath)
	if fs.ResultSpec != nil {
		funcSpec.ResultSpec = &ResultSpec{
			ReturnType: typeSpec(fs.ResultSpec.ReturnType, basePath),
			Exceptions: fieldSpecs(fs.ResultSpec.Exceptions, basePath),
		}
	}
	return funcSpec
}

func fieldSpecs(fieldSpecs compile.FieldGroup, basePath string) []*FieldSpec {
	if len(fieldSpecs) == 0 {
		return nil
	}
	specs := make([]*FieldSpec, len(fieldSpecs))
	for i, fs := range fieldSpecs {
		specs[i] = fieldSpec(fs, basePath)
	}
	return specs
}

func fieldSpec(fs *compile.FieldSpec, basePath string) *FieldSpec {
	return &FieldSpec{
		ID:          fs.ID,
		Name:        fs.Name,
		Type:        typeSpec(fs.Type, basePath),
		Required:    fs.Required,
		Default:     constantValue(fs.Default),
		Annotations: fs.Annotations,
	}
}
