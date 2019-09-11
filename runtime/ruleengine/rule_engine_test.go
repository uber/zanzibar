// Copyright (c) 2019 Uber Technologies, Inc.
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
		Patterns: []string{"RTAPI-Container", "test\\.*"},
		Value:    "presentation-staging",
	}
	rule2 := RawRule{
		Patterns: []string{"x-api-environment", "sandbox"},
		Value:    "external-api-sandbox",
	}
	rw := RuleWrapper{
		Rules: []RawRule{rule1, rule2},
	}
	re := NewRuleEngine(rw)

	var tests = []struct {
		patternValues []string
	}{
		{patternValues: []string{"x-api-environment", "sandbox"}},
		{patternValues: []string{"RTAPI-Container", "test1"}},
	}
	for _, tt := range tests {
		t.Run("tests", func(t *testing.T) {
			val, exists := re.GetValue(tt.patternValues)
			assert.True(t, exists)
			assert.NotNil(t, val)
		})
	}
}
