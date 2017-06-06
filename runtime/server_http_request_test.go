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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints"

	zanzibar "github.com/uber/zanzibar/runtime"

	"github.com/uber/zanzibar/test/lib/bench_gateway"
	"github.com/uber/zanzibar/test/lib/test_gateway"
)

func TestInvalidReadAndUnmarshalBody(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		&testGateway.Options{
			LogWhitelist: map[string]bool{
				"Could not ReadAll() body": true,
			},
			KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
			KnownTChannelBackends: []string{"baz"},
		},
		clients.CreateClients,
		endpoints.Register,
	)

	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)
	endpoint := zanzibar.NewRouterEndpoint(
		bgateway.ActualGateway,
		"foo",
		"foo",
		func(
			ctx context.Context,
			req *zanzibar.ServerHTTPRequest,
			res *zanzibar.ServerHTTPResponse,
		) {
			res.WriteJSON(200, nil, nil)
		},
	)

	httpReq, _ := http.NewRequest(
		"GET",
		"/health-check",
		&corruptReader{},
	)

	req := zanzibar.NewServerHTTPRequest(
		httptest.NewRecorder(),
		httpReq,
		nil,
		endpoint,
	)
	dJ := &dummyJson{}
	assert.False(t, req.ReadAndUnmarshalBody(dJ))

	errLogs := gateway.ErrorLogs()
	logLines := errLogs["Could not ReadAll() body"]
	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))
}

type dummyJson struct {
	foo string
}

func (d *dummyJson) UnmarshalJSON(b []byte) (err error) {
	return errors.New("Failed to unmarshal")
}

type corruptReader struct{}

func (c *corruptReader) Read(b []byte) (n int, err error) {
	return 0, errors.New("Failed to read body")
}
