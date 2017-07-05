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

package lib

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
)

// UpdateGoldenFile sets the flag if updates golden files with expected response.
var UpdateGoldenFile = flag.Bool("update", false, "Updates the golden files with expected response.")

// CompareGoldenFile compares or updates golden file with some JSON bytes.
func CompareGoldenFile(t *testing.T, goldenFilePath string, actual []byte) {
	if *UpdateGoldenFile {
		err := ioutil.WriteFile(goldenFilePath, actual, os.ModePerm)
		if err != nil {
			t.Fatalf("Fail to write into file : %s\n", err)
		}
		return
	}
	exp, err := ioutil.ReadFile(goldenFilePath)
	assert.NoError(t, err, "Failed to read %s with error %s", goldenFilePath, err)
	if bytes.Equal(exp, actual) {
		return
	}

	diffCtx := difflib.ContextDiff{
		A:        difflib.SplitLines(string(exp)),
		B:        difflib.SplitLines(string(actual)),
		FromFile: "Expected",
		ToFile:   "Actual",
		Context:  5,
		Eol:      "\n",
	}
	d, err := difflib.GetContextDiffString(diffCtx)
	assert.NoError(t, err, "Failed to compute diff.")
	t.Errorf("Result doesn't match golden file %s.\n %s\n", goldenFilePath, d)
}
