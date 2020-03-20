package router

import (
	"net/http"
	"strings"

	zanzibar "github.com/uber/zanzibar/runtime"
)

func (t *Trie) isWhitelistedPath(path string, config *zanzibar.StaticConfig) bool {
	if config == nil {
		return false
	}
	var whitelistedPaths []string
	config.MustGetStruct("router.whitelistedPaths", &whitelistedPaths)
	if len(whitelistedPaths) > 0 {
		for _, whitelistedPath := range whitelistedPaths {
			if strings.HasPrefix(whitelistedPath, path) {
				return true;
			}
		}
	}
	return false;
}

// set sets the handler for given path, creates new child node if necessary
// lastKeyCharSlash tracks whether the previous key char is a '/', used to decide it is a pattern or not
// when the current key char is ':'. lastPathCharSlash tracks whether the previous path char is a '/',
// used to decide it is a pattern or not when the current path char is ':'.
func (t *tnode) setForWhitelistedPath(path string, value http.Handler, lastKeyCharSlash, lastPathCharSlash bool) error {
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
		if (t.key[keyMatchIdx] == ':' && lastKeyCharSlash) &&
			(path[pathMatchIdx] == ':' && lastPathCharSlash) {
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
			if _, _, err := c.get(path[pathMatchIdx:], lastKeyCharSlash, lastPathCharSlash, false); err == nil {
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
					err := c.set(path[i:], value, lastKeyCharSlash, lastPathCharSlash)
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

func (t *tnode) addChildren(child *tnode, lastPathCharSlash bool) {
	if lastPathCharSlash && child.key[0] != ':' {
		// Prepending if child is not a pattern of :var
		t.children = append([]*tnode{child}, t.children...)
	} else {
		// Appending if the child is of pattern :var
		t.children = append(t.children, child)
	}
}

func (t *tnode) getForWhitelistedPath(path string, lastKeyCharSlash, lastPathCharSlash, colonAsPattern bool) (http.Handler, []Param, error) {
	keyLength, pathLength := len(t.key), len(path)
	var params []Param

	// find the longest matched prefix
	var keyIdx, pathIdx int
	for keyIdx < keyLength && pathIdx < pathLength {
		if t.key[keyIdx] == ':' && lastKeyCharSlash &&
			path[pathIdx] == ':' && lastPathCharSlash {
			keyStartIdx, pathStartIdx := keyIdx+1, pathIdx+1
			for keyIdx < keyLength && t.key[keyIdx] != '/' {
				keyIdx++
			}
			for pathIdx < pathLength && path[pathIdx] != '/' {
				pathIdx++
			}
			params = append(params, Param{t.key[keyStartIdx:keyIdx], path[pathStartIdx:pathIdx]})
		} else if t.key[keyIdx] == ':' && lastKeyCharSlash && colonAsPattern {
			// wildcard starts - match until next slash
			keyStartIdx, pathStartIdx := keyIdx+1, pathIdx
			for keyIdx < keyLength && t.key[keyIdx] != '/' {
				keyIdx++
			}
			for pathIdx < pathLength && path[pathIdx] != '/' {
				pathIdx++
			}
			params = append(params, Param{t.key[keyStartIdx:keyIdx], path[pathStartIdx:pathIdx]})
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

	// longest matched prefix matches up to node key length and path length
	if pathIdx == pathLength {
		if t.value != nil {
			return t.value, params, nil
		}
		return nil, nil, errNotFound
	}

	// longest matched prefix matches up to node key length but not path length
	for _, c := range t.children {
		if v, ps, err := c.get(path[pathIdx:], lastKeyCharSlash, lastPathCharSlash, colonAsPattern); err == nil {
			return v, append(params, ps...), nil
		}
	}

	return nil, nil, errNotFound
}