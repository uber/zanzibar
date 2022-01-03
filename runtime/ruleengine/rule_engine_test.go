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

package ruleengine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuleEngine(t *testing.T) {

	rule1 := RawRule{
		Patterns: []string{"x-test-env", `test.*`},
		Value:    "test-staging",
	}
	rule2 := RawRule{
		Patterns: []string{"x-container", `^sandbox$`},
		Value:    "dummy-sandbox",
	}
	rw := RuleWrapper{
		Rules: []RawRule{rule1, rule2},
	}
	re := NewRuleEngine(rw)

	var tests = []struct {
		patternValues []string
		matchValue    string
		noMatch       bool
	}{
		{patternValues: []string{"x-container", "sandbox"}, matchValue: "dummy-sandbox"},
		{patternValues: []string{"x-test-env", "test1"}, matchValue: "test-staging"},
		{patternValues: []string{"x-container", "sandbox123"}, noMatch: true},
		{patternValues: []string{"x-container"}, noMatch: true},
	}
	for _, tt := range tests {
		t.Run("tests", func(t *testing.T) {
			val, exists := re.GetValue(tt.patternValues...)
			if !tt.noMatch {
				assert.Equal(t, val, tt.matchValue)
				assert.True(t, exists)
			} else {
				assert.False(t, exists)
			}
		})
	}
}
