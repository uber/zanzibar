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

package codegen

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

var pascalCaseMap *sync.Map

// CommonInitialisms is taken from https://github.com/golang/lint/blob/206c0f020eba0f7fbcfbc467a5eb808037df2ed6/lint.go#L731
var CommonInitialisms = map[string]bool{
	"ACL":   true,
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"OS":    true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XMPP":  true,
	"XSRF":  true,
	"XSS":   true,
}
var camelingRegex = regexp.MustCompile("[0-9A-Za-z]+")

// LintAcronym correct the naming for initialisms
func LintAcronym(key string) string {
	key = PascalCase(key)
	for k := range CommonInitialisms {
		initial := string(k[0]) + strings.ToLower(k[1:])
		if strings.Contains(key, initial) {
			key = strings.Replace(key, initial, k, -1)
		}
	}
	return key
}

// startsWithInitialism returns the initialism if the given string begins with it
func startsWithInitialism(s string) string {
	var initialism string
	// the longest initialism is 5 char, the shortest 2
	for i := 1; i <= 5; i++ {
		if len(s) > i-1 && CommonInitialisms[s[:i]] {
			initialism = s[:i]
		}
	}
	return initialism
}

// CamelCase converts the given string to camel case
func CamelCase(src string) string {
	byteSrc := []byte(src)
	chunks := camelingRegex.FindAll(byteSrc, -1)
	for idx, val := range chunks {
		if idx > 0 {
			chunks[idx] = ensureGolangAncronymCasing(bytes.Title(val))
		} else {
			chunks[idx][0] = bytes.ToLower(val[0:1])[0]
		}
	}
	return string(bytes.Join(chunks, nil))
}

func packageName(src string) string {
	byteSrc := []byte(src)
	chunks := camelingRegex.FindAll(byteSrc, -1)
	for idx, val := range chunks {
		chunks[idx] = bytes.ToLower(val)
	}
	return string(bytes.Join(chunks, nil))
}

func ensureGolangAncronymCasing(segment []byte) []byte {
	upper := bytes.ToUpper(segment)
	if CommonInitialisms[string(upper)] {
		return upper
	}

	if initial := startsWithInitialism(string(upper)); initial != "" {
		return append([]byte(initial), segment[len(initial):]...)
	}

	return segment
}

// PascalCase converts the given string to pascal case
func PascalCase(src string) string {
	if pascalCaseMap == nil {
		pascalCaseMap = &sync.Map{}
	}
	if res, ok := pascalCaseMap.Load(src); ok {
		return res.(string)
	}
	// borrow pascal casing logic from thriftrw-go since the two implementations
	// must match otherwise the thriftrw-go generated field name does not match
	// the RequestType/ResponseType we use in the endpoint/client templates.
	// https://github.com/thriftrw/thriftrw-go/blob/1c52f516bdc5ca90dc090ba2a8ee0bd11bf04f96/gen/string.go#L48
	words := strings.Split(src, "_")
	res := pascalCase(len(words) == 1 /* all caps */, words...)
	pascalCaseMap.Store(src, res)
	return res
}

// pascalCase combines the given words using PascalCase.
//
// If allowAllCaps is true, when an all-caps word that is not a known
// abbreviation is encountered, it is left unchanged. Otherwise, it is
// Titlecased.
func pascalCase(allowAllCaps bool, words ...string) string {
	for i, chunk := range words {
		if len(chunk) == 0 {
			// foo__bar
			continue
		}

		// known initalism
		init := strings.ToUpper(chunk)
		if _, ok := CommonInitialisms[init]; ok {
			words[i] = init
			continue
		}

		// Was SCREAMING_SNAKE_CASE and not a known initialism so Titlecase it.
		if isAllCaps(chunk) && !allowAllCaps {
			// A single ALLCAPS word does not count as SCREAMING_SNAKE_CASE.
			// There must be at least one underscore.
			words[i] = strings.Title(strings.ToLower(chunk))
			continue
		}

		// Just another word, but could already be camelCased somehow, so just
		// change the first letter.
		head, headIndex := utf8.DecodeRuneInString(chunk)
		words[i] = string(unicode.ToUpper(head)) + string(chunk[headIndex:])
	}

	return strings.Join(words, "")
}

// isAllCaps checks if a string contains all capital letters only. Non-letters
// are not considered.
func isAllCaps(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

// CamelToSnake converts a given string to snake case, based on
// https://github.com/serenize/snaker/blob/master/snaker.go
func CamelToSnake(s string) string {
	var words []string
	var lastPos int
	rs := []rune(s)

	for i := 0; i < len(rs); i++ {
		if i > 0 && unicode.IsUpper(rs[i]) {
			if initialism := startsWithInitialism(s[lastPos:]); initialism != "" {
				words = append(words, initialism)
				i += len(initialism) - 1
				lastPos = i
				continue
			}

			words = append(words, s[lastPos:i])
			lastPos = i
		}
	}
	// append the last word
	if s[lastPos:] != "" {
		words = append(words, s[lastPos:])
	}
	return strings.ToLower(strings.Join(words, "_"))
}

// LowerPascal is Pascal with first char lower
func LowerPascal(str string) string {
	str = PascalCase(str)
	return strings.ToLower(string(str[0])) + string(str[1:])
}
