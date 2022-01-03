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

package router

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	errPath     = errors.New("bad path")
	errExist    = errors.New("path value already set")
	errNotFound = errors.New("not found")
)

type paramMismatch struct {
	expected, actual string
	existingPath     string
}

// Error returns the error string
func (e *paramMismatch) Error() string {
	return fmt.Sprintf("param key mismatch: expected is %s but got %s", e.expected, e.actual)
}

// Param is a url parameter where key is the url segment pattern (without :) and
// value is the actual segment of a matched url.
// e.g. url /foo/123 matches /foo/:id, the url param has key "id" and value "123"
type Param struct {
	Key, Value string
}

// Trie is a radix trie to store string value at given url path,
// a trie node corresponds to an arbitrary path substring.
type Trie struct {
	root *tnode
}

type tnode struct {
	key      string
	value    http.Handler
	children []*tnode
}

// NewTrie creates a new trie.
func NewTrie() *Trie {
	return &Trie{
		root: &tnode{
			key: "",
		},
	}
}

// Set sets the value for given path, returns error if path already set.
// When a http.Handler is registered for a given path, a subsequent Get returns the registered
// handler if the url passed to Get call matches the set path. Match in this context could mean either
// equality (e.g. url is "/foo" and path is "/foo") or url matches path pattern, which has two forms:
// - path ends with "/*", e.g. url "/foo" and "/foo/bar" both matches path "/*"
// - path contains colon wildcard ("/:"), e.g. url "/a/b" and "/a/c" bot matches path "/a/:var"
// isWhitelisted - Used for special behavior using which different handlers can configured for paths such as /a and /:b in router
func (t *Trie) Set(path string, value http.Handler, isWhitelisted bool) error {
	if path == "" || strings.Contains(path, "//") {
		return errPath
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// ignore trailing slash
	path = strings.TrimSuffix(path, "/")

	// validate "*"
	if strings.Contains(path, "*") && !strings.HasSuffix(path, "/*") {
		return errors.New("/* must be the last path segment")
	}
	if strings.Count(path, "*") > 1 {
		return errors.New("path can not contain more than one *")
	}

	colonAsPattern := !isWhitelisted
	err := t.root.set(path, value, false, false, colonAsPattern, isWhitelisted)

	if e, ok := err.(*paramMismatch); ok {
		return fmt.Errorf("path %q has a different param key %q, it should be the same key %q as in existing path %q", path, e.actual, e.expected, e.existingPath)
	}
	return err
}

// Get returns the http.Handler for given path, returns error if not found.
// It also returns the url params if given path contains any, e.g. if a handler is registered for
// "/:foo/bar", then calling Get with path "/xyz/bar" returns a param whose key is "foo" and value is "xyz".
// isWhitelisted - Used for special behavior using which different handlers can configured for paths such as /a and /:b in router
func (t *Trie) Get(path string, isWhitelisted bool) (http.Handler, []Param, error) {
	if path == "" || strings.Contains(path, "//") {
		return nil, nil, errPath
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// ignore trailing slash
	path = strings.TrimSuffix(path, "/")
	colonAsPattern := isWhitelisted
	return t.root.get(path, false, false, colonAsPattern, isWhitelisted)
}

// set sets the handler for given path, creates new child node if necessary
// lastKeyCharSlash tracks whether the previous key char is a '/', used to decide it is a pattern or not
// when the current key char is ':'. lastPathCharSlash tracks whether the previous path char is a '/',
// used to decide it is a pattern or not when the current path char is ':'.
func (t *tnode) set(path string, value http.Handler, lastKeyCharSlash, lastPathCharSlash, colonAsPattern, isWhitelisted bool) error {
	// find the longest common prefix
	var shorterLength, i int
	keyLength, pathLength := len(t.key), len(path)
	if keyLength > pathLength {
		shorterLength = pathLength
	} else {
		shorterLength = keyLength
	}
	for i < shorterLength && t.key[i] == path[i] {
		i++
	}

	// Find the first character that differs between "path" and this node's key, if it exists.
	// If we encounter a colon wildcard, ensure that the wildcard in path matches the wildcard
	// in this node's key for that segment. The segment is a colon wildcard only when the colon
	// is immediately after slash, e.g. "/:foo", "/x/:y". "/a:b" is not a colon wildcard segment.
	var keyMatchIdx, pathMatchIdx int
	for keyMatchIdx < keyLength && pathMatchIdx < pathLength {
		if t.isSetWildCardPattern(path, keyMatchIdx, pathMatchIdx, lastKeyCharSlash, lastPathCharSlash, isWhitelisted) {
			keyStartIdx, pathStartIdx := keyMatchIdx, pathMatchIdx
			same := t.key[keyMatchIdx] == path[pathMatchIdx]
			for keyMatchIdx < keyLength && t.key[keyMatchIdx] != '/' {
				keyMatchIdx++
			}
			for pathMatchIdx < pathLength && path[pathMatchIdx] != '/' {
				pathMatchIdx++
			}
			if same && (keyMatchIdx-keyStartIdx) != (pathMatchIdx-pathStartIdx) {
				return &paramMismatch{
					t.key[keyStartIdx:keyMatchIdx],
					path[pathStartIdx:pathMatchIdx],
					t.key,
				}
			}
		} else if t.key[keyMatchIdx] == path[pathMatchIdx] {
			keyMatchIdx++
			pathMatchIdx++
		} else {
			break
		}
		lastKeyCharSlash = t.key[keyMatchIdx-1] == '/'
		lastPathCharSlash = path[pathMatchIdx-1] == '/'
	}

	// If the node key is fully matched, we match the rest path with children nodes to see if a value
	// already exists for the path.
	if keyMatchIdx == keyLength {
		for _, c := range t.children {
			if _, _, err := c.get(path[pathMatchIdx:], lastKeyCharSlash, lastPathCharSlash, colonAsPattern, isWhitelisted); err == nil {
				return errExist
			}
		}
	}

	// node key is longer than longest common prefix
	if i < keyLength {
		// key/path suffix being "*" means a conflict
		if path[i:] == "*" || t.key[i:] == "*" {
			return errExist
		}

		// split the node key, add new node with node key minus longest common prefix
		split := &tnode{
			key:      t.key[i:],
			value:    t.value,
			children: t.children,
		}
		t.key = t.key[:i]
		t.value = nil
		t.children = []*tnode{split}

		// path is equal to longest common prefix
		// set value on current node after split
		if i == pathLength {
			t.value = value
		} else {
			// path is longer than longest common prefix
			// add new node with path minus longest common prefix
			newNode := &tnode{
				key:   path[i:],
				value: value,
			}
			t.addChildren(newNode, lastPathCharSlash)
		}
	}

	// node key is equal to longest common prefix
	if i == keyLength {
		// path is equal to longest common prefix
		if i == pathLength {
			// node is guaranteed to have zero value,
			// otherwise it would have caused errExist earlier
			t.value = value
		} else {
			// path is longer than node key, try to recurse on node children
			for _, c := range t.children {
				if c.key[0] == path[i] {
					lastKeyCharSlash = i > 0 && t.key[i-1] == '/'
					lastPathCharSlash = i > 0 && path[i-1] == '/'
					err := c.set(path[i:], value, lastKeyCharSlash, lastPathCharSlash, colonAsPattern, isWhitelisted)
					if e, ok := err.(*paramMismatch); ok {
						e.existingPath = t.key + e.existingPath
						return e
					}
					return err
				}
			}
			// no children to recurse, add node with path minus longest common path
			newNode := &tnode{
				key:   path[i:],
				value: value,
			}
			t.addChildren(newNode, lastPathCharSlash)
		}
	}

	return nil
}

func (t *tnode) get(path string, lastKeyCharSlash, lastPathCharSlash, colonAsPattern, isWhitelistedPath bool) (http.Handler, []Param, error) {
	keyLength, pathLength := len(t.key), len(path)
	var params []Param

	// find the longest matched prefix
	var keyIdx, pathIdx int
	for keyIdx < keyLength && pathIdx < pathLength {
		if t.isGetWildCardPattern(path, keyIdx, pathIdx, lastKeyCharSlash, lastPathCharSlash, colonAsPattern, isWhitelistedPath) {
			// wildcard starts - match until next slash
			keyStartIdx, pathStartIdx := keyIdx+1, pathIdx
			for keyIdx < keyLength && t.key[keyIdx] != '/' {
				keyIdx++
			}
			for pathIdx < pathLength && path[pathIdx] != '/' {
				pathIdx++
			}

			if t.key[keyStartIdx-1] == ':' {
				params = append(params, Param{t.key[keyStartIdx:keyIdx], path[pathStartIdx:pathIdx]})
			}
		} else if t.key[keyIdx] == path[pathIdx] {
			keyIdx++
			pathIdx++
		} else {
			break
		}
		lastKeyCharSlash = t.key[keyIdx-1] == '/'
		lastPathCharSlash = path[pathIdx-1] == '/'
	}

	if keyIdx < keyLength {
		// path matches up to node key's second to last character,
		// the last char of node key is "*" and path is no shorter than longest matched prefix
		if t.key[keyIdx:] == "*" && pathIdx < pathLength {
			return t.value, params, nil
		}
		return nil, nil, errNotFound
	}

	// ':' in path matches '*' in node key
	if keyIdx > 0 && t.key[keyIdx-1] == '*' {
		return t.value, params, nil
	}

	// longest matched prefix matches up to node key length and path length
	if pathIdx == pathLength {
		if t.value != nil {
			return t.value, params, nil
		}
		return nil, nil, errNotFound
	}

	// longest matched prefix matches up to node key length but not path length
	for _, c := range t.children {
		if v, ps, err := c.get(path[pathIdx:], lastKeyCharSlash, lastPathCharSlash, colonAsPattern, isWhitelistedPath); err == nil {
			return v, append(params, ps...), nil
		}
	}

	return nil, nil, errNotFound
}

func (t *tnode) addChildren(child *tnode, lastPathCharSlash bool) {
	if lastPathCharSlash && child.key[0] != ':' {
		// Prepending if child is not a pattern of :var
		t.children = append([]*tnode{child}, t.children...)
	} else {
		// Appending if the child is of pattern :var
		t.children = append(t.children, child)
	}
}

func (t *tnode) isSetWildCardPattern(path string, keyIdx, pathIdx int, lastKeyCharSlash, lastPathCharSlash, isWhitelistedPath bool) bool {
	if isWhitelistedPath {
		// For whitelisted paths, it will treat as wild card pattern only if key and path params are :var
		return t.key[keyIdx] == ':' && lastKeyCharSlash && path[pathIdx] == ':' && lastPathCharSlash
	}
	// For normal paths, tt will treat as wild card pattern either if key or path params are :var
	return (t.key[keyIdx] == ':' && lastKeyCharSlash) || (path[pathIdx] == ':' && lastPathCharSlash)
}

func (t *tnode) isGetWildCardPattern(path string, keyIdx, pathIdx int, lastKeyCharSlash, lastPathCharSlash, colonAsPattern, isWhitelistedPath bool) bool {
	if isWhitelistedPath {
		// For whitelisted paths, it will treat as wild card pattern only if
		// 1. Param is the key is of type :var and
		// 2. Param is the path is of type :var or colonAsPattern is true
		return t.key[keyIdx] == ':' && lastKeyCharSlash && ((path[pathIdx] == ':' && lastPathCharSlash) || colonAsPattern)
	}
	// For normal paths, it will treat as wild card pattern only if
	// 1. Param is the key is of type :var or
	// 2. Param is the path is of type :var and colonAsPattern is true
	return (t.key[keyIdx] == ':' && lastKeyCharSlash) || (path[pathIdx] == ':' && lastPathCharSlash && colonAsPattern)
}
