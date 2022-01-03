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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	get = "get"
	set = "set"
)

type ts struct {
	op             string
	path           string
	value          string
	errMsg         string
	expectedValue  string
	expectedParams []Param
}

type namedHandler struct {
	id string
}

func (n namedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func runTrieTests(t *testing.T, trie *Trie, tests []ts) {
	runTrieTestsWithWhitelist(t, trie, tests, false)
}

func runTrieTestsWithWhitelist(t *testing.T, trie *Trie, tests []ts, isWhitelisted bool) {
	for _, test := range tests {
		if test.op == set {
			err := trie.Set(test.path, namedHandler{id: test.value}, isWhitelisted)
			if test.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.errMsg)
			}
		}
		if test.op == get {
			v, ps, err := trie.Get(test.path, isWhitelisted)
			if test.errMsg == "" {
				assert.NoError(t, err, test.path)
				assert.Equal(t, test.expectedValue, v.(namedHandler).id)
				assert.Equal(t, test.expectedParams, ps)
			} else {
				assert.EqualError(t, err, test.errMsg)
			}
		}
	}
	//printTrie(trie)
}

func TestTrieLiteralPath(t *testing.T) {
	tree := NewTrie()
	tests := []ts{
		// test blank path
		{op: set, path: "", value: "foo", errMsg: errPath.Error()},
		{op: get, path: "", errMsg: errPath.Error()},
		// test root path
		{op: set, path: "/", value: "foo"},
		{op: get, path: "/", expectedValue: "foo"},
		// test set
		{op: set, path: "/a/b", value: "bar"},
		{op: set, path: "/a/b/c", value: "bar"},
		// test set conflict
		{op: set, path: "/a/b/c", value: "baz", errMsg: errExist.Error()},
		// test trailing slash when set
		{op: set, path: "/a/b/c/", value: "baz", errMsg: errExist.Error()},
		// test not found
		{op: get, path: "/a", errMsg: errNotFound.Error()},
		{op: get, path: "/a/b/d", errMsg: errNotFound.Error()},
		// test get
		{op: get, path: "/a/b", expectedValue: "bar"},
		{op: get, path: "/a/b/c", expectedValue: "bar"},
		// test trailing slash when get
		{op: get, path: "/a/b/c/", expectedValue: "bar"},
		// test missing starting slash
		{op: set, path: "a", value: "foo"},
		{op: get, path: "a", expectedValue: "foo"},
		// test branching
		{op: set, path: "/a/e/f", value: "quxx"},
		{op: get, path: "/a/e/f", expectedValue: "quxx"},
		// test segment overlap
		{op: set, path: "/a/good", value: "good"},
		{op: set, path: "/a/goto", value: "goto"},
		{op: get, path: "/a/good", expectedValue: "good"},
		{op: get, path: "/a/goto", expectedValue: "goto"},
	}

	runTrieTests(t, tree, tests)

	tree = NewTrie()
	runTrieTestsWithWhitelist(t, tree, tests, true)
}

func TestTriePathsWithPatten(t *testing.T) {
	trie := NewTrie()
	tests := []ts{
		// test setting "/*/a" is not allowed
		{op: set, path: "/*/a", value: "foo", errMsg: "/* must be the last path segment"},
		// test setting path with multiple "*" is not allowed
		{op: set, path: "/*/*", value: "foo", errMsg: "path can not contain more than one *"},
		// test "/a" does not collide with "/a/*"
		{op: set, path: "/a", value: "foo"},
		// test "/a/*" match all paths starts with "/a/"
		{op: set, path: "/a/*", value: "bar"},
		{op: get, path: "a", expectedValue: "foo"},
		{op: get, path: "/a/b", expectedValue: "bar"},
		{op: get, path: "/a/b/c/d", expectedValue: "bar"},
		{op: get, path: "/a/b/c/d", expectedValue: "bar"},
		// test paths starts with "/a/" collides with "/a/*"
		{op: set, path: "/a/b/", value: "baz", errMsg: errExist.Error()},
		{op: set, path: "/a/b/c", value: "baz", errMsg: errExist.Error()},
		{op: set, path: "/a/:b", value: "baz", errMsg: errExist.Error()},
		// test "/*" collides with "/a"
		{op: set, path: "/*", value: "baz", errMsg: errExist.Error()},
		// test "/:" collides with "/a"
		{op: set, path: "/:x", value: "baz", errMsg: errExist.Error()},
		// test "/:/b" collides with "/a/*"
		{op: set, path: "/:x/b", value: "baz", errMsg: errExist.Error()},
	}
	runTrieTests(t, trie, tests)

	tests = []ts{
		// test ":a" is not treated as a pattern when queried as a url
		{op: set, path: "/a", value: "foo"},
		{op: get, path: "/:a", errMsg: errNotFound.Error()},

		{op: set, path: "/a:b", value: "bar"},
		{op: set, path: "/a:c", value: "baz"},
		{op: get, path: "/a:b", expectedValue: "bar"},
		{op: get, path: "/ac", errMsg: errNotFound.Error()},
		{op: get, path: "/a:", errMsg: errNotFound.Error()},
	}
	runTrieTests(t, NewTrie(), tests)
	runTrieTestsWithWhitelist(t, NewTrie(), tests, true)

	tests = []ts{
		{op: set, path: "/:a", value: "foo"},
		{op: get, path: "/:a", expectedValue: "foo", expectedParams: []Param{{"a", ":a"}}},
	}
	runTrieTests(t, NewTrie(), tests)
	runTrieTestsWithWhitelist(t, NewTrie(), tests, true)

	tests = []ts{
		// test "/a" does not collide with "/:a/b"
		{op: set, path: "/:a/b", value: "foo"},
		{op: set, path: "/a", value: "bar"},
		{op: get, path: "/a", expectedValue: "bar"},
		{op: get, path: "/x/b/", expectedValue: "foo", expectedParams: []Param{{"a", "x"}}},
		{op: get, path: "/a/", expectedValue: "bar"},
	}
	runTrieTests(t, NewTrie(), tests)
	runTrieTestsWithWhitelist(t, NewTrie(), tests, true)

	tests = []ts{
		// test "/:a/b" does not collide with "/a"
		{op: set, path: "/a", value: "bar"},
		{op: set, path: "/:a/b", value: "foo"},
		{op: get, path: "/a", expectedValue: "bar"},
		{op: get, path: "/x/b/", expectedValue: "foo", expectedParams: []Param{{"a", "x"}}},
		{op: get, path: "/a/", expectedValue: "bar"},
	}
	runTrieTests(t, NewTrie(), tests)
	runTrieTestsWithWhitelist(t, NewTrie(), tests, true)

	tests = []ts{
		// test "/b" collides with "/:"
		{op: set, path: "/:a", value: "foo"},
		{op: set, path: "/b", errMsg: errExist.Error()},
		{op: get, path: "/a/", expectedValue: "foo", expectedParams: []Param{{"a", "a"}}},
	}
	runTrieTests(t, NewTrie(), tests)

	tests = []ts{
		// test "/b" does not collide with "/:" for whitelisted paths
		{op: set, path: "/:a", value: "1"},
		{op: set, path: "/b", value: "2"},
		{op: get, path: "/a/", expectedValue: "1", expectedParams: []Param{{"a", "a"}}},
		{op: get, path: "/b/", expectedValue: "2"},
	}
	runTrieTestsWithWhitelist(t, NewTrie(), tests, true)

	tests = []ts{
		// test "/:" collides with "/b"
		{op: set, path: "/b", value: "foo"},
		{op: set, path: "/:a", errMsg: errExist.Error()},
		{op: get, path: "/b/", expectedValue: "foo"},
	}
	runTrieTests(t, NewTrie(), tests)

	tests = []ts{
		// test "/:" does not collide with "/b" for whitelisted paths
		{op: set, path: "/a", value: "1"},
		{op: set, path: "/:b", value: "2"},
		{op: get, path: "/a/", expectedValue: "1"},
		{op: get, path: "/b/", expectedValue: "2", expectedParams: []Param{{"b", "b"}}},
	}
	runTrieTestsWithWhitelist(t, NewTrie(), tests, true)

	tests = []ts{
		// more ":" tests
		{op: set, path: "/a/b", value: "1"},
		{op: set, path: "/a/b/:cc/d", value: "2"},
		{op: set, path: "/a/b/:x/e", errMsg: "path \"/a/b/:x/e\" has a different param key \":x\", it should be the same key \":cc\" as in existing path \"/a/b/:cc/d\""},
		{op: set, path: "/a/b/c/x", value: "2.1"},
		{op: set, path: "/a/b/:cc/:d/e", value: "3"},
		{op: set, path: "/a/b/c/d/f", value: "4"},
		{op: set, path: "/a/:b/c/d", errMsg: errExist.Error()},
		{op: get, path: "/a/b/some/d", expectedValue: "2", expectedParams: []Param{{"cc", "some"}}},
		{op: get, path: "/a/b/c/x", expectedValue: "2.1"},
		{op: get, path: "/a/b/other/data/e", expectedValue: "3",
			expectedParams: []Param{
				{"cc", "other"},
				{"d", "data"},
			}},
		{op: get, path: "/a/b/c/d/f", expectedValue: "4"},
	}
	runTrieTests(t, NewTrie(), tests)

	tests = []ts{
		// more ":" tests for whitelisted paths
		{op: set, path: "/a/b", value: "1"},
		{op: set, path: "/a/b/:cc/d", value: "2"},
		{op: set, path: "/a/b/:x/e", errMsg: "path \"/a/b/:x/e\" has a different param key \":x\", it should be the same key \":cc\" as in existing path \"/a/b/:cc/d\""},
		{op: set, path: "/a/b/c/x", value: "2.1"},
		{op: set, path: "/a/b/:cc/:d/e", value: "3"},
		{op: set, path: "/a/b/c/d/f", value: "4"},
		{op: set, path: "/a/:b/c/d", value: "5"},
		{op: get, path: "/a/b/some/d", expectedValue: "2", expectedParams: []Param{{"cc", "some"}}},
		{op: get, path: "/a/b/c/x", expectedValue: "2.1"},
		{op: get, path: "/a/b/other/data/e", expectedValue: "3",
			expectedParams: []Param{
				{"cc", "other"},
				{"d", "data"},
			}},
		{op: get, path: "/a/b/c/d/f", expectedValue: "4"},
		{op: get, path: "/a/x/c/d/", expectedValue: "5", expectedParams: []Param{{"b", "x"}}},
	}
	runTrieTestsWithWhitelist(t, NewTrie(), tests, true)

	tests = []ts{
		// more ":" tests
		{op: set, path: "/a/b", value: "1"},
		{op: set, path: "/a/b/ccc/x", value: "2"},
		{op: set, path: "/a/b/c/dope/f", value: "3"},
		{op: set, path: "/a/b/ccc/:", errMsg: errExist.Error()},
		{op: set, path: "/a/b/c/:/:/", errMsg: errExist.Error()},
		{op: get, path: "/a/b/ccc", errMsg: errNotFound.Error()},
		{op: get, path: "/a/b/:", errMsg: errNotFound.Error()},
	}
	runTrieTests(t, NewTrie(), tests)

	tests = []ts{
		// more ":" tests for whitelisted paths
		{op: set, path: "/a/b", value: "1"},
		{op: set, path: "/a/b/ccc/x", value: "2"},
		{op: set, path: "/a/b/c/dope/f", value: "3"},
		{op: set, path: "/a/b/ccc/:", value: "4"},
		{op: set, path: "/a/b/c/:/:/", value: "5"},
		{op: get, path: "/a/b/ccc", errMsg: errNotFound.Error()},
		{op: get, path: "/a/b/:", errMsg: errNotFound.Error()},
	}
	runTrieTestsWithWhitelist(t, NewTrie(), tests, true)

	tests = []ts{
		// more ":" tests
		{op: set, path: "/a/:b/c", value: "1"},
		{op: set, path: "/a/:b/d", value: "2"},
		{op: set, path: "/a/", value: "3"},
		{op: get, path: "/a/b/c", expectedValue: "1", expectedParams: []Param{{"b", "b"}}},
		{op: get, path: "/a/b/d", expectedValue: "2", expectedParams: []Param{{"b", "b"}}},
	}
	runTrieTests(t, NewTrie(), tests)
	runTrieTestsWithWhitelist(t, NewTrie(), tests, true)
}

// simple test for coverage
func TestParamMismatch(t *testing.T) {
	pm := paramMismatch{
		expected: "foo",
		actual:   "bar",
	}
	assert.Equal(t, "param key mismatch: expected is foo but got bar", pm.Error())
}

// utilities for debugging
func printTrie(t *Trie) {
	buf := new(bytes.Buffer)
	var levelsEnded []int
	printNodes(buf, t.root.children, 0, levelsEnded)
	fmt.Println(string(buf.Bytes()))
}

func printNodes(w io.Writer, nodes []*tnode, level int, levelsEnded []int) {
	for i, node := range nodes {
		edge := "├──"
		if i == len(nodes)-1 {
			levelsEnded = append(levelsEnded, level)
			edge = "└──"
		}
		printNode(w, node, level, levelsEnded, edge)
		if len(node.children) > 0 {
			printNodes(w, node.children, level+1, levelsEnded)
		}
	}
}

func printNode(w io.Writer, node *tnode, level int, levelsEnded []int, edge string) {
	for i := 0; i < level; i++ {
		isEnded := false
		for _, l := range levelsEnded {
			if l == i {
				isEnded = true
				break
			}
		}
		if isEnded {
			_, _ = fmt.Fprint(w, "    ")
		} else {
			_, _ = fmt.Fprint(w, "│   ")
		}
	}
	if node.value != nil {
		_, _ = fmt.Fprintf(w, "%s %v (%v)\n", edge, node.key, node.value)
	} else {
		_, _ = fmt.Fprintf(w, "%s %v\n", edge, node.key)
	}
}
