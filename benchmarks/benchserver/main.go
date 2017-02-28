// Copyright (c) 2017 Uber Technologies, Inc.
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

package main

import (
	"net/http"

	"github.com/uber-go/zap"
	testBackend "github.com/uber/zanzibar/test/lib/test_backend"
)

// PORT that the bench server listens to
var PORT int32 = 8092

func main() {
	logger := zap.New(zap.NewJSONEncoder())
	backend := testBackend.CreateBackend(PORT)
	err := backend.Bootstrap()
	if err != nil {
		panic(err)
	}

	handleContacts := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(202)
	}
	handleGoogleNow := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(202)
	}
	backend.HandleFunc("POST", "/foo/contacts", handleContacts)
	backend.HandleFunc("POST", "/foo/contacts", handleGoogleNow)

	logger.Info("Listening on port & serving")

	backend.Wait()
}
