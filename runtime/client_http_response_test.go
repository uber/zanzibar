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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
	"github.com/uber/zanzibar/test/lib/test_gateway"

	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints"
)

func TestReadAndUnmarshalNonStructBody(t *testing.T) {
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

	fakeEcho := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, err := w.Write([]byte(`"foo"`))
		assert.NoError(t, err)
	}

	bgateway.HTTPBackends()["bar"].HandleFunc("POST", "/bar/echo", fakeEcho)

	addr := bgateway.HTTPBackends()["bar"].RealAddr
	baseURL := "http://" + addr

	client := zanzibar.NewHTTPClient(bgateway.ActualGateway, baseURL)
	req := zanzibar.NewClientHTTPRequest("bar", "echo", client)

	err = req.WriteJSON("POST", baseURL+"/bar/echo", nil, myJson{})
	assert.NoError(t, err)
	res, err := req.Do(context.Background())
	assert.NoError(t, err)

	var resp string
	assert.NoError(t, res.ReadAndUnmarshalNonStructBody(&resp))
	assert.Equal(t, "foo", resp)
}

func TestReadAndUnmarshalNonStructBodyUnmarshalError(t *testing.T) {
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

	fakeEcho := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, err := w.Write([]byte(`false`))
		assert.NoError(t, err)
	}

	bgateway.HTTPBackends()["bar"].HandleFunc("POST", "/bar/echo", fakeEcho)

	addr := bgateway.HTTPBackends()["bar"].RealAddr
	baseURL := "http://" + addr

	client := zanzibar.NewHTTPClient(bgateway.ActualGateway, baseURL)
	req := zanzibar.NewClientHTTPRequest("bar", "echo", client)

	err = req.WriteJSON("POST", baseURL+"/bar/echo", nil, myJson{})
	assert.NoError(t, err)
	res, err := req.Do(context.Background())
	assert.NoError(t, err)

	var resp string
	assert.Error(t, res.ReadAndUnmarshalNonStructBody(&resp))
	assert.Equal(t, "", resp)
}

type myJson struct{}

func (sj myJson) MarshalJSON() ([]byte, error) {
	return []byte(`{"name":"foo"}`), nil
}
