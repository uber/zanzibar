package ruleengine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuleEngine(t *testing.T) {

	rule1 := RawRule{
		Patterns: []string{"\\.*", "test\\.*"},
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
