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

package gateway_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

func TestBootstrapError(t *testing.T) {
	gateway1, err := testGateway.CreateGateway(t, nil, &testGateway.Options{
		TestBinary:  util.DefaultMainFile("example-gateway"),
		ConfigFiles: util.DefaultConfigFiles("example-gateway"),
	})
	if !assert.NoError(t, err, "must be able to create gateway") {
		return
	}

	defer gateway1.Close()

	assert.NotNil(t, gateway1, "gateway exists")

	config2 := map[string]interface{}{}
	config2["http.port"] = int32(gateway1.HTTPPort())
	gateway2, err := testGateway.CreateGateway(t, config2, &testGateway.Options{
		LogWhitelist: map[string]bool{
			"Error listening on port": true,
		},
		TestBinary:  util.DefaultMainFile("example-gateway"),
		ConfigFiles: util.DefaultConfigFiles("example-gateway"),
	})

	assert.Error(t, err, "expected err creating server")
	assert.Nil(t, gateway2, "expected no gateway")

	switch err := errors.Cause(err).(type) {
	case *testGateway.MalformedStdoutError:
		assert.Contains(t, err.StdoutLine, "Error listening on port",
			"error should be about listening on port")
		assert.Contains(t, err.StdoutLine, "address already in use",
			"error message is about address in use")
	default:
		assert.Fail(t, "got weird error")
	}
}

// TODO(argo): Let us redo this test to actually verify other invalid keys in config
func TestBootstrapWithBadLogLevel(t *testing.T) {
	t.Skip("Skip for now to redo test to be more inclusive of other bad config keys")
	gateway, err := testGateway.CreateGateway(t, map[string]interface{}{
		"logger.level": "invalid",
	}, &testGateway.Options{
		TestBinary:  util.DefaultMainFile("example-gateway"),
		ConfigFiles: util.DefaultConfigFiles("example-gateway"),
	})
	assert.Error(t, err, "got bootstrap err")
	assert.Nil(t, gateway)
}
