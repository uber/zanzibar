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
	"regexp"
)

// RuleWrapper is a container for list of rules
type RuleWrapper struct {
	Rules []RawRule
}

// RawRule defines a rule by defining a pattern and a value
type RawRule struct {
	Patterns []string
	Value    interface{}
}

type ruleEngine struct {
	rules []rule
}

type rule struct {
	patterns []*regexp.Regexp
	value    interface{}
}

// RuleEngine provides a way to get a value if rule matched
type RuleEngine interface {
	// GetValue returns a value, true if rule matched, else returns a nil, false if no rule matched
	GetValue(patternValues ...string) (interface{}, bool)
}

// NewRuleEngine initializes a rule engine
func NewRuleEngine(ruleWrapper RuleWrapper) RuleEngine {
	re := &ruleEngine{}
	for _, rule := range ruleWrapper.Rules {
		cRule := convertRawRules(rule)
		re.rules = append(re.rules, cRule)
	}
	return re
}

func convertRawRules(rawRule RawRule) rule {
	r := rule{value: rawRule.Value}
	for _, pat := range rawRule.Patterns {
		r.patterns = append(r.patterns, regexp.MustCompile(pat))
	}
	return r
}

func (r *ruleEngine) GetValue(patternValues ...string) (interface{}, bool) {
	for _, rule := range r.rules {
		matched := true
		if len(rule.patterns) != len(patternValues) {
			matched = false
			continue
		}

		for i, pat := range rule.patterns {
			if !r.match(pat, patternValues[i]) {
				matched = false
				break
			}
		}

		if matched {
			return rule.value, matched
		}
	}

	return nil, false
}

func (r *ruleEngine) match(pattern *regexp.Regexp, patternValue string) bool {
	return pattern.MatchString(patternValue)
}
