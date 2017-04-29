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
	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	"github.com/uber/zanzibar/examples/example-gateway/build/clients/bar"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints"
	clientsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/bar/bar"
	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
	"github.com/uber/zanzibar/test/lib/test_gateway"
)

func TestMakingClientWriteJSONWithBadJSON(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		nil,
		nil,
		clients.CreateClients,
		endpoints.Register,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)
	client := zanzibar.NewHTTPClient(bgateway.ActualGateway, "/")
	req := zanzibar.NewClientHTTPRequest("clientID", "DoStuff", client)

	err = req.WriteJSON("GET", "/foo", nil, &failingJsonObj{})
	assert.NotNil(t, err)

	assert.Equal(t,
		"Could not serialize json for client: clientID: cannot serialize",
		err.Error(),
	)
}

func TestMakingClientWriteJSONWithBadHTTPMethod(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		nil,
		nil,
		clients.CreateClients,
		endpoints.Register,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)
	client := zanzibar.NewHTTPClient(bgateway.ActualGateway, "/")
	req := zanzibar.NewClientHTTPRequest("clientID", "DoStuff", client)

	err = req.WriteJSON("@INVALIDMETHOD", "/foo", nil, nil)
	assert.NotNil(t, err)

	assert.Equal(t,
		"Could not make outbound request for client: "+
			"clientID: net/http: invalid method \"@INVALIDMETHOD\"",
		err.Error(),
	)
}

func TestMakingClientCalLWithHeaders(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
	}, clients.CreateClients, endpoints.Register)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(r.Header.Get("Example-Header")))
		},
	)

	clients := bgateway.ActualGateway.Clients.(*clients.Clients)
	client := clients.Bar.HTTPClient

	req := zanzibar.NewClientHTTPRequest("bar", "bar-path", client)

	err = req.WriteJSON(
		"POST",
		client.BaseURL+"/bar-path",
		map[string]string{
			"Example-Header": "Example-Value",
		},
		nil,
	)
	assert.NoError(t, err)

	res, err := req.Do(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)

	bytes, err := res.ReadAll()
	assert.NoError(t, err)

	assert.Equal(t, []byte("Example-Value"), bytes)
}

func TestMakingClientCalLWithRespHeaders(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
	}, clients.CreateClients, endpoints.Register)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Example-Header", "Example-Value")
			w.WriteHeader(200)
			_, _ = w.Write([]byte("{}"))
		},
	)
	clients := bgateway.ActualGateway.Clients.(*clients.Clients)
	bClient := clients.Bar

	body, headers, err := bClient.Normal(
		context.Background(), nil, &barClient.NormalHTTPRequest{},
	)
	assert.NoError(t, err)

	assert.NotNil(t, body)
	assert.Equal(t, "Example-Value", headers["Example-Header"])
}

func TestMakingClientCallWithThriftException(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
	}, clients.CreateClients, endpoints.Register)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
			_, _ = w.Write([]byte(`{"stringField":"test"}`))
		},
	)
	clients := bgateway.ActualGateway.Clients.(*clients.Clients)
	bClient := clients.Bar

	body, _, err := bClient.Normal(
		context.Background(), nil, &barClient.NormalHTTPRequest{},
	)
	assert.Error(t, err)
	assert.Nil(t, body)

	realError := err.(*clientsBarBar.BarException)
	assert.Equal(t, realError.StringField, "test")
}

func TestMakingClientCallWithBadStatusCode(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
	}, clients.CreateClients, endpoints.Register)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(402)
			_, _ = w.Write([]byte(`{"stringField":"test"}`))
		},
	)
	clients := bgateway.ActualGateway.Clients.(*clients.Clients)
	bClient := clients.Bar

	body, _, err := bClient.Normal(
		context.Background(), nil, &barClient.NormalHTTPRequest{},
	)
	assert.Error(t, err)
	assert.Nil(t, body)

	assert.Equal(t, "Unexpected http client response (402)", err.Error())
}

func TestMakingCallWithThriftException(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(nil, &testGateway.Options{
		KnownHTTPBackends: []string{"bar"},
	}, clients.CreateClients, endpoints.Register)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	gateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/arg-not-struct-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(403)
			_, _ = w.Write([]byte(`{"stringField":"test"}`))
		},
	)

	bgateway := gateway.(*benchGateway.BenchGateway)
	clients := bgateway.ActualGateway.Clients.(*clients.Clients)

	_, err = clients.Bar.ArgNotStruct(
		context.Background(), nil,
		&barClient.ArgNotStructHTTPRequest{
			Request: "request",
		},
	)
	assert.Error(t, err)

	realError := err.(*clientsBarBar.BarException)
	assert.Equal(t, realError.StringField, "test")
}
