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

	"github.com/pkg/errors"
	"github.com/thriftrw/thriftrw-go/compile"
)

// Module represents a compiled Thrift module. In contrast to thriftrw-go's
// compile.Module, all fields of this Module are struct and hence can be
// serialized and deserailized.
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
func ConvertModule(module *compile.Module, basePath string) (*Module, error) {
	incModules, err := includedModules(module.Includes, basePath)
	if err != nil {
		return nil, err
	}
	constants := make(map[string]*Constant)
	for name, c := range module.Constants {
		constants[name], err = constant(c, basePath)
		if err != nil {
			return nil, err
		}
	}
	types := make(map[string]*TypeSpec)
	for name, t := range module.Types {
		types[name] = typeSpec(t, basePath)
	}
	serviceSpecs, err := convertServiceSpec(module.Services, basePath)
	if err != nil {
		return nil, err
	}
	return &Module{
		Name:       module.Name,
		ThriftPath: relPath(basePath, module.ThriftPath),
		Includes:   incModules,
		Constants:  constants,
		Types:      types,
		Services:   serviceSpecs,
	}, nil
}

func relPath(basePath, targetPath string) string {
	rel, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return targetPath
	}
	return rel
}

func includedModules(includes map[string]*compile.IncludedModule, basePath string) (map[string]*IncludedModule, error) {
	includedModule := make(map[string]*IncludedModule)
	for name, incModule := range includes {
		m, err := ConvertModule(incModule.Module, basePath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert included module %q", name)
		}
		includedModule[name] = &IncludedModule{
			Name:   incModule.Name,
			Module: m,
		}
	}
	return includedModule, nil
}

func constant(c *compile.Constant, basePath string) (*Constant, error) {
	if c == nil {
		return nil, nil
	}
	value, err := constantValue(c.Value)
	if err != nil {
		return nil, err
	}
	return &Constant{
		Name:  c.Name,
		File:  relPath(basePath, c.File),
		Type:  typeSpec(c.Type, basePath),
		Value: value,
	}, nil
}

func constantValue(value compile.ConstantValue) (string, error) {
	if value == nil {
		return "", nil
	}
	// TODO(zw): Add more type converstions.
	switch t := value.(type) {
	case compile.ConstantBool:
		return fmt.Sprintf("%t", t), nil
	case compile.ConstantInt:
		return fmt.Sprintf("%d", t), nil
	case compile.ConstantString:
		return string(t), nil
	case compile.ConstantDouble:
		return fmt.Sprintf("%f", t), nil
	case compile.EnumItemReference:
		return fmt.Sprintf("%s.%s", t.Enum.Name, t.Item.Name), nil
	}
	return "", errors.Errorf("unknown constant value %v", value)
}

func typeSpec(t compile.TypeSpec, basePath string) *TypeSpec {
	if t == nil {
		return nil
	}
	return &TypeSpec{
		Name:        t.ThriftName(),
		File:        relPath(basePath, t.ThriftFile()),
		Annotations: t.ThriftAnnotations(),
	}
}

func convertServiceSpec(services map[string]*compile.ServiceSpec, basePath string) (map[string]*ServiceSpec, error) {
	specs := make(map[string]*ServiceSpec)
	for name, spec := range services {
		functions := make(map[string]*FunctionSpec)
		for fname, funcSpec := range spec.Functions {
			f, err := functionSpec(funcSpec, basePath)
			if err != nil {
				return nil, errors.Wrapf(err,
					"failed to convert function spec %q in service %q", fname, name)
			}
			functions[fname] = f
		}
		specs[name] = &ServiceSpec{
			Name:        spec.Name,
			File:        relPath(basePath, spec.File),
			Functions:   functions,
			Annotations: spec.Annotations,
		}
	}
	return specs, nil
}
func functionSpec(fs *compile.FunctionSpec, basePath string) (*FunctionSpec, error) {
	if fs == nil {
		return nil, nil
	}
	funcSpec := &FunctionSpec{
		Name:        fs.Name,
		OneWay:      fs.OneWay,
		Annotations: fs.Annotations,
	}
	argsSpec, err := fieldSpecs(compile.FieldGroup(fs.ArgsSpec), basePath)
	if err != nil {
		return nil, err
	}
	funcSpec.ArgsSpec = argsSpec
	if fs.ResultSpec != nil {
		exceptions, err := fieldSpecs(fs.ResultSpec.Exceptions, basePath)
		if err != nil {
			return nil, err
		}
		funcSpec.ResultSpec = &ResultSpec{
			ReturnType: typeSpec(fs.ResultSpec.ReturnType, basePath),
			Exceptions: exceptions,
		}
	}
	return funcSpec, nil
}

func fieldSpecs(fieldSpecs compile.FieldGroup, basePath string) ([]*FieldSpec, error) {
	if len(fieldSpecs) == 0 {
		return nil, nil
	}
	specs := make([]*FieldSpec, len(fieldSpecs))
	var err error
	for i, fs := range fieldSpecs {
		specs[i], err = fieldSpec(fs, basePath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert fieldSpec %v", fs)
		}
	}
	return specs, nil
}

func fieldSpec(fs *compile.FieldSpec, basePath string) (*FieldSpec, error) {
	defaultValue, err := constantValue(fs.Default)
	if err != nil {
		return nil, err
	}
	return &FieldSpec{
		ID:          fs.ID,
		Name:        fs.Name,
		Type:        typeSpec(fs.Type, basePath),
		Required:    fs.Required,
		Default:     defaultValue,
		Annotations: fs.Annotations,
	}, nil
}
