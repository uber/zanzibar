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
	"time"

	"github.com/stretchr/testify/assert"
	clientsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/bar/bar"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	"github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
	"github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

var defaultTestOptions *testGateway.Options = &testGateway.Options{
	KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
	KnownTChannelBackends: []string{"baz"},
	ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
}
var defaultTestConfig map[string]interface{} = map[string]interface{}{
	"clients.baz.serviceName": "baz",
}

func TestMakingClientWriteJSONWithBadJSON(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)
	client := zanzibar.NewHTTPClient(
		bgateway.ActualGateway.Logger,
		bgateway.ActualGateway.AllHostScope,
		"clientID",
		[]string{"DoStuff"},
		"/",
		time.Second,
	)
	req := zanzibar.NewClientHTTPRequest("clientId", "DoStuff", client)

	err = req.WriteJSON("GET", "/foo", nil, &failingJsonObj{})
	assert.NotNil(t, err)
	assert.Equal(t,
		"Could not serialize clientId.DoStuff request json: cannot serialize",
		err.Error(),
	)

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Could not serialize request json"], 1)
}

func TestMakingClientWriteJSONWithBadHTTPMethod(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)
	client := zanzibar.NewHTTPClient(
		bgateway.ActualGateway.Logger,
		bgateway.ActualGateway.AllHostScope,
		"clientID",
		[]string{"DoStuff"},
		"/",
		time.Second,
	)
	req := zanzibar.NewClientHTTPRequest("clientId", "DoStuff", client)

	err = req.WriteJSON("@INVALIDMETHOD", "/foo", nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t,
		"Could not create outbound clientId.DoStuff request: net/http: invalid method \"@INVALIDMETHOD\"",
		err.Error(),
	)

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Could not create outbound request"], 1)
}

func TestMakingClientCalLWithHeaders(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
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

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	barClient := deps.Client.Bar
	client := barClient.HTTPClient()

	req := zanzibar.NewClientHTTPRequest("bar", "Normal", client)

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

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Finished an outgoing client HTTP request"], 1)
}

func TestBarClientWithoutHeaders(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, _, err = bar.EchoI8(
		context.Background(), nil, &clientsBarBar.Echo_EchoI8_Args{Arg: 42},
	)

	assert.NotNil(t, err)
	assert.Equal(t, "Missing mandatory header: x-uuid", err.Error())

	logs := gateway.AllLogs()

	assert.Equal(t, 1, len(logs))

	lines := logs["Got outbound request without mandatory header"]
	assert.Equal(t, 1, len(lines))

	logLine := lines[0]
	assert.Equal(t, "bar", logLine["clientId"])
	assert.Equal(t, "EchoI8", logLine["methodName"])
	assert.Equal(t, "x-uuid", logLine["headerName"])
}

func TestMakingClientCallWithRespHeaders(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
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
			_, _ = w.Write([]byte(`{
				"stringField":"foo",
				"intWithRange": 0,
				"intWithoutRange": 1,
				"mapIntWithRange": {},
				"mapIntWithoutRange": {}
			}`))
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bClient := deps.Client.Bar

	body, headers, err := bClient.Normal(
		context.Background(), nil, &clientsBarBar.Bar_Normal_Args{},
	)
	assert.NoError(t, err)
	assert.NotNil(t, body)
	assert.Equal(t, "Example-Value", headers["Example-Header"])

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Finished an outgoing client HTTP request"], 1)
}

func TestMakingClientCallWithThriftException(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
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

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bClient := deps.Client.Bar

	body, _, err := bClient.Normal(
		context.Background(), nil, &clientsBarBar.Bar_Normal_Args{},
	)
	assert.Error(t, err)
	assert.Nil(t, body)

	realError := err.(*clientsBarBar.BarException)
	assert.Equal(t, realError.StringField, "test")

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Finished an outgoing client HTTP request"], 1)
}

func TestMakingClientCallWithBadStatusCode(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
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

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bClient := deps.Client.Bar

	body, _, err := bClient.Normal(
		context.Background(), nil, &clientsBarBar.Bar_Normal_Args{},
	)
	assert.Error(t, err)
	assert.Nil(t, body)
	assert.Equal(t, "Unexpected http client response (402)", err.Error())

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Unknown response status code"], 1)
	assert.Len(t, logs["Finished an outgoing client HTTP request"], 1)
}

func TestMakingCallWithThriftException(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
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
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bClient := deps.Client.Bar

	_, err = bClient.ArgNotStruct(
		context.Background(), nil,
		&clientsBarBar.Bar_ArgNotStruct_Args{
			Request: "request",
		},
	)
	assert.Error(t, err)

	realError := err.(*clientsBarBar.BarException)
	assert.Equal(t, realError.StringField, "test")

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Finished an outgoing client HTTP request"], 1)
}

func TestMakingClientCallWithServerError(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar-path",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(`{}`))
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bClient := deps.Client.Bar

	body, _, err := bClient.Normal(
		context.Background(), nil, &clientsBarBar.Bar_Normal_Args{},
	)
	assert.Error(t, err)
	assert.Nil(t, body)
	assert.Equal(t, "Unexpected http client response (500)", err.Error())

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Unknown response status code"], 1)
	assert.Len(t, logs["Finished an outgoing client HTTP request"], 1)
}
