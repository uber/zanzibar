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

package zanzibar_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"

	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
)

func TestCreatingTChannel(t *testing.T) {
	_, err := benchGateway.CreateGateway(
		map[string]interface{}{
			"tchannel.serviceName": "",
			"serviceName":          "",
		},
		nil,
		exampleGateway.CreateGateway,
	)
	assert.Error(t, err)

	assert.Contains(t, err.Error(), "no service name provided")
}
