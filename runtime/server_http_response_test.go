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

package zanzibar_test

import (
	"context"
	"io/ioutil"
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/assert"
	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
)

func TestInvalidStatusCode(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(nil, nil)
	if !assert.NoError(t, err) {
		return
	}

	bgateway := gateway.(*benchGateway.BenchGateway)
	bgateway.ActualGateway.Router.Register(
		"GET", "/foo", zanzibar.NewEndpoint(
			bgateway.ActualGateway,
			"foo",
			"foo",
			func(
				ctx context.Context,
				req *zanzibar.ServerHTTPRequest,
				res *zanzibar.ServerHTTPResponse,
			) {
				res.WriteJSONBytes(999, []byte("true"))
			},
		),
	)

	resp, err := gateway.MakeRequest("GET", "/foo", nil)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, resp.Status, "999 status code 999")
	assert.Equal(t, resp.StatusCode, 999)

	bytes, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "true", string(bytes))

	errorLogs := bgateway.GetErrorLogs()
	logLines := errorLogs["Could not emit statusCode metric"]

	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))

	line := logLines[0]
	lineStruct := map[string]interface{}{}
	jsonErr := json.Unmarshal([]byte(line), &lineStruct)
	if !assert.NoError(t, jsonErr, "cannot decode json lines") {
		return
	}

	code := lineStruct["UnexpectedStatusCode"].(float64)
	assert.Equal(t, 999.0, code)
}
