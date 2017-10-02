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

import "github.com/xeipuuv/gojsonschema"

// ResultError abstracts a subset of interface in gojsonschema
type ResultError interface {
	Field() string
	Type() string
	Description() string
	Value() interface{}
	String() string
}

// Result interface type subset of gojsonschema
type Result interface {
	Valid() bool
	Errors() []ResultError
}

type resultStruct struct {
	valid  bool
	errors []ResultError
}

func (r *resultStruct) Valid() bool {
	return r.valid
}

func (r *resultStruct) Errors() []ResultError {
	return r.errors
}

// SchemaValidator abstracts undlerlying schema validation library
type SchemaValidator struct{}

// ValidateGo performs JSON schema validation from schema file on to arbitrary Go object
func (s *SchemaValidator) ValidateGo(schemaFile string, arbitraryObj map[string]interface{}) (Result, error) {
	// gojsonschmea caches schema validators
	schemaLoader := gojsonschema.NewReferenceLoader(schemaFile)
	jsonLoader := gojsonschema.NewGoLoader(arbitraryObj)
	res, err := gojsonschema.Validate(schemaLoader, jsonLoader)

	if res == nil {
		return nil, err
	}

	// convert gojsonschema to our result type
	resultErrors := res.Errors()
	myResult := &resultStruct{
		res.Valid(),
		make([]ResultError, len(res.Errors())),
	}
	for i, resultError := range resultErrors {
		myResult.errors[i] = resultError.(ResultError)
	}

	return myResult, err
}
