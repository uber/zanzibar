// Copyright (c) 2019 Uber Technologies, Inc.
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

package serverless_endpoint_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

func TestServerlessField(t *testing.T) {
	var counter int = 0

	gateway, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		TestBinary:        util.DefaultMainFile("example-gateway"),
		ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()


	res, err := gateway.MakeRequest(
		"POST",
		"/serverless/post-request",
		nil, nil,
	)

	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, counter)

	respBytes, err := ioutil.ReadAll(res.Body)
	if !assert.NoError(t, err, "got http resp error") {
		return
	}

	assert.Equal(t, `"hello"`, string(respBytes))
}
