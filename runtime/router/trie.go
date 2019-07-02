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

// Set sets the value for given path, returns error if path already set
func (t *Trie) Set(path string, value http.Handler) error {
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

	err := t.root.set(path, value)
	if e, ok := err.(*paramMismatch); ok {
		return fmt.Errorf("path %q has a different param key %q, it should be the same key %q as in existing path %q", path, e.actual, e.expected, e.existingPath)
	}
	return err
}

// Get returns the value for given path, returns error if not found
func (t *Trie) Get(path string) (http.Handler, []Param, error) {
	if path == "" || strings.Contains(path, "//") {
		return nil, nil, errPath
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// ignore trailing slash
	path = strings.TrimSuffix(path, "/")
	return t.root.get(path)
}

func (t *tnode) set(path string, value http.Handler) error {
	// find the longest common prefix
	var l, i int
	m, n := len(t.key), len(path)
	if m > n {
		l = n
	} else {
		l = m
	}
	for i < l && t.key[i] == path[i] {
		i++
	}

	// find index j, k in key and path to which they match,
	// j and k is only useful to check if there is conflict,
	// they are not where splits happen, splits happen at index i.
	var j, k int
	for j < m && k < n {
		if t.key[j] == ':' || path[k] == ':' {
			oj, ok := j, k
			same := t.key[j] == path[k]
			for j < m && t.key[j] != '/' {
				j++
			}
			for k < n && path[k] != '/' {
				k++
			}
			if same && (j-oj) != (k-ok) {
				return &paramMismatch{
					t.key[oj:j],
					path[ok:k],
					t.key,
				}
			}
		} else if t.key[j] == path[k] {
			j++
			k++
		} else {
			break
		}
	}

	// conflicts caused by ":" is only possible when j == m
	if j == m {
		for _, c := range t.children {
			if c.key[0] == path[k] || c.key[0] == ':' || path[k] == ':' {
				if _, _, err := c.get(path[k:]); err == nil {
					return errExist
				}
			}
		}
	}

	// node ley is longer than longest common prefix
	if i < m {
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
		if i == n {
			t.value = value
		} else {
			// path is longer than longest common prefix
			// add new node with path minus longest common prefix
			newNode := &tnode{
				key:   path[i:],
				value: value,
			}
			t.children = append(t.children, newNode)
		}
	}

	// node key is equal to longest common prefix
	if i == m {
		// path is equal to longest common prefix
		if i == n {
			// node is guaranteed to have zero value,
			// otherwise it would have caused errExist earlier
			t.value = value
		} else {
			// path is longer than node key, try to recurse on node children
			for _, c := range t.children {
				if c.key[0] == path[i] {
					err := c.set(path[i:], value)
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
			t.children = append(t.children, newNode)
		}
	}

	return nil
}

func (t *tnode) get(path string) (http.Handler, []Param, error) {
	m, n := len(t.key), len(path)
	var params []Param

	// find the longest matched prefix
	var j, k int
	for j < m && k < n {
		if t.key[j] == ':' {
			oj, ok := j+1, k
			for j < m && t.key[j] != '/' {
				j++
			}
			for k < n && path[k] != '/' {
				k++
			}
			params = append(params, Param{t.key[oj:j], path[ok:k]})
		} else if path[k] == ':' { // necessary for conflict check used in set call
			for j < m && t.key[j] != '/' {
				j++
			}
			for k < n && path[k] != '/' {
				k++
			}
		} else if t.key[j] == path[k] {
			j++
			k++
		} else {
			break
		}
	}

	if j < m {
		// path matches up to node key's second to last character,
		// the last char of node key is "*" and path is no shorter than longest matched prefix
		if t.key[j:] == "*" && k < n {
			return t.value, params, nil
		}
		return nil, nil, errNotFound
	}

	// ':' in path matches '*' in node key
	if j > 0 && t.key[j-1] == '*' {
		return t.value, params, nil
	}

	// longest matched prefix matches up to node key length and path length
	if k == n {
		if t.value != nil {
			return t.value, params, nil
		}
		return nil, nil, errNotFound
	}

	// TODO: recursion to iteration for speed
	// longest matched prefix matches up to node key length but not path length
	for _, c := range t.children {
		if c.key[0] == path[k] || c.key[0] == ':' || path[k] == ':' {
			if v, ps, err := c.get(path[k:]); err == nil {
				return v, append(params, ps...), nil
			}
		}
	}

	return nil, nil, errNotFound
}
