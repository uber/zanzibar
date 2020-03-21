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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandle(t *testing.T) {
	r := &Router{}

	handled := false
	err := r.Handle("GET", "/*",
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handled = true
		}), false)
	assert.NoError(t, err, "unexpected error")

	req, _ := http.NewRequest("GET", "/foo", nil)
	r.ServeHTTP(nil, req, false)
	assert.True(t, handled)
}

func TestParamsFromContext(t *testing.T) {
	r := &Router{}

	handled := false
	err := r.Handle("GET", "/:var",
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			params := ParamsFromContext(req.Context())
			assert.Equal(t, 1, len(params))
			assert.Equal(t, "var", params[0].Key)
			assert.Equal(t, "foo", params[0].Value)
			handled = true
		}), false)
	assert.NoError(t, err, "unexpected error")

	req, _ := http.NewRequest("GET", "/foo", nil)
	r.ServeHTTP(nil, req, false)
	assert.True(t, handled)
}

func TestPanicHandler(t *testing.T) {
	handled := false
	r := &Router{
		PanicHandler: func(writer http.ResponseWriter, req *http.Request, i interface{}) {
			handled = true
		},
	}

	err := r.Handle("GET", "/foo",
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			panic("something went wrong")
		}), false)
	assert.NoError(t, err, "unexpected error")

	req, _ := http.NewRequest("GET", "/foo", nil)
	r.ServeHTTP(nil, req, false)
	assert.True(t, handled)
}

func TestMethodNotAllowedDefault(t *testing.T) {
	r := &Router{HandleMethodNotAllowed: true}

	handled := false
	err := r.Handle("GET", "/foo",
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handled = true
		}), false)
	assert.NoError(t, err, "unexpected error")
	err = r.Handle("PUT", "/bar",
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handled = true
		}), false)
	assert.NoError(t, err, "unexpected error")

	req, _ := http.NewRequest("PUT", "/foo", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req, false)
	assert.False(t, handled)
	assert.Equal(t, http.StatusMethodNotAllowed, res.Result().StatusCode)
	assert.Equal(t, "GET", res.Result().Header.Get("Allow"))
}

func TestMethodNotAllowedCustom(t *testing.T) {
	r := &Router{
		HandleMethodNotAllowed: true,
		MethodNotAllowed: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("life", "42")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}),
	}

	handled := false
	err := r.Handle("GET", "/foo",
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handled = true
		}), false)
	assert.NoError(t, err, "unexpected error")
	err = r.Handle("POST", "/foo",
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handled = true
		}), false)
	assert.NoError(t, err, "unexpected error")

	req, _ := http.NewRequest("PUT", "/foo", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req, false)
	assert.False(t, handled)
	assert.Equal(t, "42", res.Result().Header.Get("life"))
	assert.Equal(t, "GET, POST", res.Result().Header.Get("Allow"))
}

func TestNotFoundDefault(t *testing.T) {
	r := &Router{}

	req, _ := http.NewRequest("GET", "/foo", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req, false)
	assert.Equal(t, http.StatusNotFound, res.Result().StatusCode)
}

func TestNotFoundCustom(t *testing.T) {
	handled := false
	r := &Router{
		NotFound: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			handled = true
		}),
	}

	req, _ := http.NewRequest("GET", "/foo", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req, false)
	assert.True(t, handled)
	assert.Equal(t, http.StatusNotFound, res.Result().StatusCode)
}
