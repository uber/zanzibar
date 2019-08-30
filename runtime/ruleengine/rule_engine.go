package ruleengine

import (
	"errors"
	"regexp"
)

type RuleWrapper struct {
	Rules []RawRule
}

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

var (
	noPatternMatch = errors.New("pattern mismatch")
)

type RuleEngine interface {
	GetValue(patternValues []string) (interface{}, bool)
}

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

func (r *ruleEngine) GetValue(patternValues []string) (interface{}, bool) {
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
