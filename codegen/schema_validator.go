// Copyright (c) 2018 Uber Technologies, Inc.
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
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// SchemaValidateGo performs JSON schema validation from schema file on to arbitrary Go object
func SchemaValidateGo(schemaFile string, arbitraryObj map[string]interface{}) error {
	// gojsonschmea caches schema validators
	schemaLoader := gojsonschema.NewReferenceLoader(schemaFile)
	jsonLoader := gojsonschema.NewGoLoader(arbitraryObj)
	result, err := gojsonschema.Validate(schemaLoader, jsonLoader)

	if err != nil && result != nil {
		resultErrors, _ := json.Marshal(result.Errors())
		return errors.Wrap(err, "schema validation error\nErrors:\n"+string(resultErrors))
	}
	if err != nil {
		return errors.Wrap(err, "schema validation error unknown")
	}
	if !result.Valid() {
		msg := "schema validation error :"
		for _, resErr := range result.Errors() {

			details, _ := json.Marshal(resErr.Details())
			msg += fmt.Sprintf("\nType: %s\nField: %s\nDescription: %s\nDetails: %s\n",
				resErr.Type(),
				resErr.Field(),
				resErr.Description(),
				string(details),
			)
		}
		return errors.New(msg)
	}
	return nil
}
