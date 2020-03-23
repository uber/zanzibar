// Copyright (c) 2020 Uber Technologies, Inc.
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
	"testing"
)

func TestTrieLiteralPathForWhitelisted(t *testing.T) {
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

	runTrieTestsWithWhitelist(t, tree, tests, true)
}

func TestTriePathsWithPattenForWhitelisted(t *testing.T) {
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
		// test "/:" does not collide with "/a"
		{op: set, path: "/:x", value: "1"},
		// test "/:/b" does not collide with "/a/*"
		{op: set, path: "/:x/b", value: "2"},

		{op: get, path: "/x/", expectedValue: "1", expectedParams: []Param{{"x", "x"}}},
		{op: get, path: "/a/b/", expectedValue: "bar"},
		{op: get, path: "/x/b", expectedValue: "2", expectedParams: []Param{{"x", "x"}}},
	}
	runTrieTestsWithWhitelist(t, trie, tests, true)

	trie = NewTrie()
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
	runTrieTestsWithWhitelist(t, trie, tests, true)

	trie = NewTrie()
	tests = []ts{
		{op: set, path: "/:a", value: "foo"},
		{op: get, path: "/:a", expectedValue: "foo", expectedParams: []Param{{"a", ":a"}}},
	}
	runTrieTestsWithWhitelist(t, trie, tests, true)

	trie = NewTrie()
	tests = []ts{
		// test "/a" does not collide with "/:a/b"
		{op: set, path: "/:a/b", value: "foo"},
		{op: set, path: "/a", value: "bar"},
		{op: get, path: "/a", expectedValue: "bar"},
		{op: get, path: "/x/b/", expectedValue: "foo", expectedParams: []Param{{"a", "x"}}},
		{op: get, path: "/a/", expectedValue: "bar"},
	}
	runTrieTestsWithWhitelist(t, trie, tests, true)

	trie = NewTrie()
	tests = []ts{
		// test "/:a/b" does not collide with "/a"
		{op: set, path: "/a", value: "bar"},
		{op: set, path: "/:a/b", value: "foo"},
		{op: get, path: "/a", expectedValue: "bar"},
		{op: get, path: "/x/b/", expectedValue: "foo", expectedParams: []Param{{"a", "x"}}},
		{op: get, path: "/a/", expectedValue: "bar"},
	}
	runTrieTestsWithWhitelist(t, trie, tests, true)

	trie = NewTrie()
	tests = []ts{
		// more ":" tests
		{op: set, path: "/:a", value: "1"},
		{op: set, path: "/b", value: "2"},
		{op: get, path: "/a/", expectedValue: "1", expectedParams: []Param{{"a", "a"}}},
		{op: get, path: "/b/", expectedValue: "2"},
	}
	runTrieTestsWithWhitelist(t, trie, tests, true)

	tests = []ts{
		// more ":" tests
		{op: set, path: "/a", value: "1"},
		{op: set, path: "/:b", value: "2"},
		{op: get, path: "/a/", expectedValue: "1"},
		{op: get, path: "/b/", expectedValue: "2", expectedParams: []Param{{"b", "b"}}},
	}

	trie = NewTrie()
	tests = []ts{
		// more ":" tests
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
	runTrieTestsWithWhitelist(t, trie, tests, true)

	trie = NewTrie()
	tests = []ts{
		// more ":" tests
		{op: set, path: "/a/b", value: "1"},
		{op: set, path: "/a/b/ccc/x", value: "2"},
		{op: set, path: "/a/b/c/dope/f", value: "3"},
		{op: set, path: "/a/b/ccc/:", value: "4"},
		{op: set, path: "/a/b/c/:/:/", value: "5"},
		{op: get, path: "/a/b/ccc", errMsg: errNotFound.Error()},
		{op: get, path: "/a/b/:", errMsg: errNotFound.Error()},
	}
	runTrieTestsWithWhitelist(t, trie, tests, true)

	trie = NewTrie()
	tests = []ts{
		// more ":" tests
		{op: set, path: "/a/:b/c", value: "1"},
		{op: set, path: "/a/:b/d", value: "2"},
		{op: set, path: "/a/", value: "3"},
		{op: get, path: "/a/b/c", expectedValue: "1", expectedParams: []Param{{"b", "b"}}},
		{op: get, path: "/a/b/d", expectedValue: "2", expectedParams: []Param{{"b", "b"}}},
	}
	runTrieTestsWithWhitelist(t, trie, tests, true)
}
