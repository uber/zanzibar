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

package zanzibar

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

type routerSuite struct {
	suite.Suite

	router *httpRouter
	gw     *Gateway
	scope  tally.TestScope
}

func (s *routerSuite) SetupTest() {
	logger, _ := zap.NewDevelopment()
	s.gw = new(Gateway)
	s.gw.Logger = logger
	s.gw.ContextLogger = NewContextLogger(zap.NewNop())
	s.scope = tally.NewTestScope("", nil)
	s.gw.RootScope = s.scope
	s.gw.Config = NewStaticConfigOrDie(
		[]*ConfigOption{},
		map[string]interface{}{
			"router.whitelistedPaths": []string{"/a/b"},
		},
	)
	s.router = NewHTTPRouter(s.gw).(*httpRouter)
}

func (s *routerSuite) TestRouter() {
	err := s.router.Handle("GET", "/noslash", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("noslash\n"))
	}))
	s.NoError(err)
	err = s.router.Handle("GET", "/withslash/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("withslash\n"))
	}))
	s.NoError(err)
	err = s.router.Handle("POST", "/postonly", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("postonly\n"))
	}))
	s.NoError(err)
	err = s.router.Handle("GET", "/panicerror", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(errors.New("test"))
	}))
	s.NoError(err)
	err = s.router.Handle("GET", "/panicstring", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test")
	}))
	s.NoError(err)

	cases := []struct {
		RequestMethod string
		RequestPath   string
		RequestBody   []byte
		ResponseCode  int
		ResponseBody  []byte
	}{
		{"GET", "/notfound", nil, http.StatusNotFound, []byte("404 page not found\n")},
		{"GET", "/noslash", nil, http.StatusOK, []byte("noslash\n")},
		{"GET", "/noslash/", nil, http.StatusOK, []byte("noslash\n")},
		{"GET", "/withslash", nil, http.StatusOK, []byte("withslash\n")},
		{"GET", "/withslash/", nil, http.StatusOK, []byte("withslash\n")},
		{"GET", "/postonly", nil, http.StatusMethodNotAllowed, []byte("Method Not Allowed\n")},
		{"GET", "/panicerror", nil, http.StatusInternalServerError, []byte("Internal Server Error\n")},
		{"GET", "/panicstring", nil, http.StatusInternalServerError, []byte("Internal Server Error\n")},
	}

	for i, testCase := range cases {
		req := httptest.NewRequest(testCase.RequestMethod, testCase.RequestPath, bytes.NewBuffer(testCase.RequestBody))
		req.Header.Set("Test_Key", "Test_Value")
		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		s.Equal(testCase.ResponseCode, w.Code, "expected response code for %dth test case to be equal", i)
		body, _ := ioutil.ReadAll(w.Result().Body)
		s.Equal(testCase.ResponseBody, body, "expected response body for %dth test case to be equal", i)
	}
}

func (s *routerSuite) TestExtractParameters() {
	err := s.router.Handle("GET", "/foo/:a", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := ParamsFromContext(r.Context())
		s.Equal("x", params.Get("a"))
	}))
	s.NoError(err)
	req := httptest.NewRequest("GET", "/foo/x", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
}

func (s *routerSuite) TestRouteConflict() {
	err := s.router.Handle("GET", "/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("foo1\n"))
	}))
	s.NoError(err)

	err = s.router.Handle("GET", "/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("foo2\n"))
	}))
	s.Error(err)

	// Test that the original route is still registered and working correctly.

	req := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
	body, _ := ioutil.ReadAll(w.Result().Body)
	s.Equal([]byte("foo1\n"), body)
}

func (s *routerSuite) TestRouteConflictVariable() {
	err := s.router.Handle("GET", "/foo/:a", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("foo a\n"))
	}))
	s.NoError(err)

	err = s.router.Handle("GET", "/foo/:b", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("foo b\n"))
	}))
	s.Error(err)

	// Test that the original route is still registered and working correctly.

	req := httptest.NewRequest("GET", "/foo/x", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
	body, _ := ioutil.ReadAll(w.Result().Body)
	s.Equal([]byte("foo a\n"), body)
}

func TestRouterSuite(t *testing.T) {
	s := new(routerSuite)
	suite.Run(t, s)
}
