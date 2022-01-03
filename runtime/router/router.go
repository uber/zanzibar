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
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// Router dispatches http requests to a registered http.Handler.
// It implements a similar interface to the one in github.com/julienschmidt/httprouter,
// the main differences are:
// 1. this router does not treat "/a/:b" and "/a/b/c" as conflicts (https://github.com/julienschmidt/httprouter/issues/175)
// 2. this router does not treat "/a/:b" and "/a/:c" as different routes and therefore does not allow them to be registered at the same time (https://github.com/julienschmidt/httprouter/issues/6)
// 3. this router does not treat "/a" and "/a/" as different routes
// 4. this router treats "/a" and "/:b" as different paths for whitelisted paths
// Also the `*` pattern is greedy, if a handler is register for `/a/*`, then no handler
// can be further registered for any path that starts with `/a/`
type Router struct {
	tries map[string]*Trie

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed http.Handler

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound http.Handler

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})

	// Used for special behavior using which different handlers can configured
	// for paths such as /a and /:b in router.
	WhitelistedPaths []string

	// TODO: (clu) maybe support OPTIONS
}

type paramsKey string

// urlParamsKey is the request context key under which URL params are stored.
const urlParamsKey = paramsKey("urlParamsKey")

// ParamsFromContext pulls the URL parameters from a request context,
// or returns nil if none are present.
func ParamsFromContext(ctx context.Context) []Param {
	p, _ := ctx.Value(urlParamsKey).([]Param)
	return p
}

// Handle registers a http.Handler for given method and path.
func (r *Router) Handle(method, path string, handler http.Handler) error {
	if r.tries == nil {
		r.tries = make(map[string]*Trie)
	}

	trie, ok := r.tries[method]
	if !ok {
		trie = NewTrie()
		r.tries[method] = trie
	}
	err := trie.Set(path, handler, r.isWhitelistedPath(path))
	if err == errExist {
		return &urlFailure{url: path, method: method}
	}
	return err
}

// urlFailure captures errors for conflicting paths
type urlFailure struct {
	method string
	url    string
}

// Error returns the error string
func (e *urlFailure) Error() string {
	return fmt.Sprintf("panic: path: %q method: %q conflicts with an existing path", e.url, e.method)
}

// ServeHTTP dispatches the request to a register handler to handle.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.PanicHandler != nil {
		defer func(w http.ResponseWriter, req *http.Request) {
			if recovered := recover(); recovered != nil {
				r.PanicHandler(w, req, recovered)
			}
		}(w, req)
	}

	reqPath := req.URL.Path
	isWhitelisted := r.isWhitelistedPath(reqPath)
	if trie, ok := r.tries[req.Method]; ok {
		if handler, params, err := trie.Get(reqPath, isWhitelisted); err == nil {
			ctx := context.WithValue(req.Context(), urlParamsKey, params)
			req = req.WithContext(ctx)
			handler.ServeHTTP(w, req)
			return
		}
	}

	if r.HandleMethodNotAllowed {
		if allowed := r.allowed(reqPath, req.Method, isWhitelisted); allowed != "" {
			w.Header().Set("Allow", allowed)
			if r.MethodNotAllowed != nil {
				r.MethodNotAllowed.ServeHTTP(w, req)
			} else {
				http.Error(w,
					http.StatusText(http.StatusMethodNotAllowed),
					http.StatusMethodNotAllowed,
				)
			}
			return
		}
	}

	if r.NotFound != nil {
		r.NotFound.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}

func (r *Router) allowed(path, reqMethod string, isWhitelisted bool) string {
	var allow []string

	for method, trie := range r.tries {
		if method == reqMethod || method == http.MethodOptions {
			continue
		}

		if _, _, err := trie.Get(path, isWhitelisted); err == nil {
			allow = append(allow, method)
		}
	}
	sort.Slice(allow, func(i, j int) bool {
		return allow[i] < allow[j]
	})

	return strings.Join(allow, ", ")
}

func (r *Router) isWhitelistedPath(path string) bool {
	for _, whitelistedPath := range r.WhitelistedPaths {
		whitelistedPathTokens := strings.Split(whitelistedPath, "/")
		pathTokens := strings.Split(path, "/")
		if len(whitelistedPathTokens) != len(pathTokens) {
			continue
		}

		isMatched := true
		for i, token := range whitelistedPathTokens {
			if pathTokens[i] != token && token[0] != ':' {
				isMatched = false
				break
			}
		}
		if isMatched {
			return true
		}
	}
	return false
}
