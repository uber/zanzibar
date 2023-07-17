// Code generated by zanzibar
// @generated

// Copyright (c) 2018 Uber Technologies, Inc.
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

package googlenowendpoint

import (
	"bytes"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	testGateway "github.com/uber/zanzibar/v2/test/lib/test_gateway"
)

func getDirNameCheckCredentialsSuccessfulRequest() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(file)
}

func TestCheckCredentialsSuccessfulRequestOKResponse(t *testing.T) {
	var counter int

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		KnownHTTPBackends: []string{"google-now"},
		TestBinary: filepath.Join(
			getDirNameCheckCredentialsSuccessfulRequest(),
			"../../..",
			"build", "services", "example-gateway",
			"main", "main.go",
		),
		ConfigFiles: []string{
			filepath.Join(
				getDirNameCheckCredentialsSuccessfulRequest(),
				"../../..",
				"config", "test.yaml",
			),
			filepath.Join(
				getDirNameCheckCredentialsSuccessfulRequest(),
				"../../..",
				"config", "example-gateway", "test.yaml",
			),
		},
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	fakeCheckCredentials := func(w http.ResponseWriter, r *http.Request) {

		assert.Equal(
			t,
			"test-uuid",
			r.Header.Get("X-Uuid"))

		w.Header().Set("X-Uuid", "test-uuid")

		w.WriteHeader(202)

		var payload []byte

		// TODO(zw): generate client response.
		if _, err := w.Write(payload); err != nil {
			t.Fatal("can't write fake response")
		}
		counter++
	}

	gateway.HTTPBackends()["google-now"].HandleFunc(
		"POST", "/check-credentials", fakeCheckCredentials,
	)

	headers := map[string]string{}
	headers["X-Token"] = "test-token"
	headers["X-Uuid"] = "test-uuid"

	endpointRequest := []byte(`{"authcode":"test"}`)

	res, err := gateway.MakeRequest(
		"POST",
		"/googlenow/check-credentials",
		headers,
		bytes.NewReader(endpointRequest),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, 202, res.StatusCode)
	assert.Equal(
		t,
		"test-uuid",
		res.Header.Get("X-Uuid"))

	assert.Equal(t, 1, counter)
}
