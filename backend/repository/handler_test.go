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

package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	testlib "github.com/uber/zanzibar/test/lib"
	"go.uber.org/zap/zapcore"
)

type TestCase struct {
	Name     string   `json:"name"`
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Request struct {
	Method string                 `json:"method"`
	URL    string                 `json:"url"`
	Header map[string]string      `json:"header,omitempty"`
	Body   map[string]interface{} `json:"body,omitempty"`
}

type Response struct {
	Code int                    `json:"code"`
	Body map[string]interface{} `json:"body,omitempty"`
}

const testCaseFile = "data/handler/test_cases.json"

func TestHandlers(t *testing.T) {
	var testCases []*TestCase
	err := readJSONFile(testCaseFile, &testCases)
	if !assert.NoError(t, err, "Failed to read test cases.") {
		return
	}
	server := NewTestServer()
	t.Log("Server started\n")
	// The following tests will be run in parallel with 'go test -parallel <#threads>'.
	t.Run("group", func(t *testing.T) {
		for _, tc := range testCases {
			tc := tc // Capture range variable
			t.Run(tc.Name, func(t *testing.T) {
				//t.Parallel()
				resp, err := getResponse(t, server, tc.Request)
				if !assert.NoErrorf(t, err, "Failed to get response. Request: %v", tc.Request) {
					return
				}
				if *testlib.UpdateGoldenFile {
					tc.Response = *resp
					return
				}
				checkResponse(t, &tc.Response, resp)
			})
		}
	})
	if *testlib.UpdateGoldenFile {
		t.Log("Updating golden file.")
		err := writeToJSONFile(testCaseFile, testCases)
		assert.NoErrorf(t, err, "Failed to write expected responses into %q", testCaseFile)
	}
}

type logger struct{}

type diffCreator struct{}

func (l *logger) Info(msg string, fields ...zapcore.Field) {
	fmt.Printf("Info message: %s, fields: %v", msg, fields)
}

func (l *logger) Error(msg string, fields ...zapcore.Field) {
	fmt.Printf("Error message: %s, fields: %v", msg, fields)
}

func (d *diffCreator) NewDiff(r *Repository, request *DiffRequest) (string, error) {
	return fmt.Sprintf("http://diff-for-%s", r.Remote()), nil
}

func (d *diffCreator) LandDiff(r *Repository, diffURI string) error {
	return nil
}

func NewTestServer() *httptest.Server {
	manager := NewTestManager()
	handler := NewHandler(manager, &diffCreator{}, "gateway-id", &logger{})
	router := handler.NewHTTPRouter()
	return httptest.NewServer(router)
}

func getResponse(t *testing.T, server *httptest.Server, request Request) (*Response, error) {
	url := server.URL + request.URL
	b, err := json.Marshal(request.Body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(request.Method, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	for k, v := range request.Header {
		req.Header.Add(k, v)
	}
	t.Logf("Sending request: %+v\n", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	var jsonBody map[string]interface{}
	err = json.Unmarshal(body, &jsonBody)
	if err != nil {
		return nil, errors.Wrapf(err, "body is %s", body)
	}
	return &Response{
		Code: resp.StatusCode,
		Body: jsonBody,
	}, nil
}

func checkResponse(t *testing.T, exp *Response, actual *Response) {
	assert.Equal(t, exp.Code, actual.Code, "Status code mismatch.")
	expBytes, err := json.MarshalIndent(exp.Body, "", "\t")
	if !assert.NoError(t, err, "Failed to marshal expected body.") {
		return
	}
	actualBytes, err := json.MarshalIndent(actual.Body, "", "\t")
	if !assert.NoError(t, err, "Failed to marshal actual body.") {
		return
	}
	expBody := string(expBytes)
	actualBody := string(actualBytes)
	if expBody == actualBody {
		return
	}
	testlib.DiffStrings(t, expBody, actualBody)
}
