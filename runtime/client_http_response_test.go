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
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	zanzibar "github.com/uber/zanzibar/runtime"
	"github.com/uber/zanzibar/runtime/jsonwrapper"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

func TestReadAndUnmarshalNonStructBody(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		&testGateway.Options{
			LogWhitelist: map[string]bool{
				"Could not read response body": true,
			},
			KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
			KnownTChannelBackends: []string{"baz"},
			ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		},
		exampleGateway.CreateGateway,
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

	client := zanzibar.NewHTTPClientContext(
		bgateway.ActualGateway.ContextLogger,
		bgateway.ActualGateway.ContextMetrics,
		jsonwrapper.NewDefaultJSONWrapper(),
		"bar",
		map[string]string{
			"echo": "bar::echo",
		},
		baseURL,
		map[string]string{},
		time.Second,
		true,
	)
	ctx := context.Background()

	req := zanzibar.NewClientHTTPRequest(ctx, "bar", "echo", "bar::echo", client,
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		})

	err = req.WriteJSON("POST", baseURL+"/bar/echo", nil, myJson{})
	assert.NoError(t, err)
	res, err := req.Do()
	assert.NoError(t, err)

	var resp string
	assert.NoError(t, res.ReadAndUnmarshalBody(&resp))
	assert.Equal(t, "foo", resp)

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Finished an outgoing client HTTP request"], 1)
}

func TestReadAndUnmarshalNonStructBodyUnmarshalError(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		&testGateway.Options{
			LogWhitelist: map[string]bool{
				"Could not read response body": true,
			},
			KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
			KnownTChannelBackends: []string{"baz"},
			ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		},
		exampleGateway.CreateGateway,
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

	client := zanzibar.NewHTTPClientContext(
		bgateway.ActualGateway.ContextLogger,
		bgateway.ActualGateway.ContextMetrics,
		jsonwrapper.NewDefaultJSONWrapper(),
		"bar",
		map[string]string{
			"echo": "bar::echo",
		},
		baseURL,
		map[string]string{},
		time.Second,
		true,
	)
	ctx := context.Background()

	req := zanzibar.NewClientHTTPRequest(ctx, "bar", "echo", "bar::echo", client,
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		})

	err = req.WriteJSON("POST", baseURL+"/bar/echo", nil, myJson{})
	assert.NoError(t, err)
	res, err := req.Do()
	assert.NoError(t, err)

	var resp string
	assert.Error(t, res.ReadAndUnmarshalBody(&resp))
	assert.Equal(t, "", resp)

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Could not parse response json"], 1)
	assert.Len(t, logs["Finished an outgoing client HTTP request"], 1)
}

func TestUnknownStatusCode(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		&testGateway.Options{
			LogWhitelist: map[string]bool{
				"Could not emit statusCode metric": true,
			},
			KnownHTTPBackends: []string{"bar"},
			ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
		},
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	addr := bgateway.HTTPBackends()["bar"].RealAddr
	baseURL := "http://" + addr

	// test UnknownStatusCode along with follow redirect by default
	fakeEcho := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", baseURL+"/bar-path")
		w.WriteHeader(303)
		_, err := w.Write([]byte(`false`))
		assert.NoError(t, err)
	}

	fakeNormal := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(999)
		_, err := w.Write([]byte(`false`))
		assert.NoError(t, err)
	}

	bgateway.HTTPBackends()["bar"].HandleFunc("POST", "/bar/echo", fakeEcho)
	bgateway.HTTPBackends()["bar"].HandleFunc("GET", "/bar-path", fakeNormal)

	client := zanzibar.NewHTTPClientContext(
		bgateway.ActualGateway.ContextLogger,
		bgateway.ActualGateway.ContextMetrics,
		jsonwrapper.NewDefaultJSONWrapper(),
		"bar",
		map[string]string{
			"echo": "bar::echo",
		},
		baseURL,
		map[string]string{},
		time.Second,
		true,
	)

	ctx := context.Background()

	req := zanzibar.NewClientHTTPRequest(ctx, "bar", "echo", "bar::echo", client,
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		})

	err = req.WriteJSON("POST", baseURL+"/bar/echo", nil, myJson{})
	assert.NoError(t, err)

	res, err := req.Do()
	assert.NoError(t, err)

	var resp string
	assert.Error(t, res.ReadAndUnmarshalBody(&resp))
	assert.Equal(t, "", resp)
	assert.Equal(t, 999, res.StatusCode)

	logLines := bgateway.Logs("error", "Could not emit statusCode metric")
	assert.NotNil(t, logLines)
	assert.Equal(t, 1, len(logLines))

	lineStruct := logLines[0]
	code := lineStruct["UnknownStatusCode"].(float64)
	assert.Equal(t, 999.0, code)

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Could not parse response json"], 1)
	assert.Len(t, logs["Could not emit statusCode metric"], 1)
	assert.Len(t, logs["Finished an outgoing client HTTP request"], 1)
}

func TestNotFollowRedirect(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		&testGateway.Options{
			LogWhitelist: map[string]bool{
				"Could not emit statusCode metric": true,
			},
			KnownHTTPBackends: []string{"bar"},
			ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
		},
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	addr := bgateway.HTTPBackends()["bar"].RealAddr
	baseURL := "http://" + addr

	redirectURI := baseURL + "/bar-path"
	fakeEcho := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", redirectURI)
		w.WriteHeader(303)
		_, err := w.Write([]byte(`false`))
		assert.NoError(t, err)
	}

	bgateway.HTTPBackends()["bar"].HandleFunc("POST", "/bar/echo", fakeEcho)

	client := zanzibar.NewHTTPClientContext(
		bgateway.ActualGateway.ContextLogger,
		bgateway.ActualGateway.ContextMetrics,
		jsonwrapper.NewDefaultJSONWrapper(),
		"bar",
		map[string]string{
			"echo": "bar::echo",
		},
		baseURL,
		map[string]string{},
		time.Second,
		false,
	)

	ctx := context.Background()

	req := zanzibar.NewClientHTTPRequest(ctx, "bar", "echo", "bar::echo", client,
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		})

	err = req.WriteJSON("POST", baseURL+"/bar/echo", nil, myJson{})
	assert.NoError(t, err)

	res, err := req.Do()
	assert.NoError(t, err)

	var resp string
	assert.Error(t, res.ReadAndUnmarshalBody(&resp))
	assert.Equal(t, "", resp)
	assert.Equal(t, 303, res.StatusCode)
	assert.Equal(t, redirectURI, res.Header["Location"][0])

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Finished an outgoing client HTTP request"], 1)
}

type myJson struct{}

func (sj myJson) MarshalJSON() ([]byte, error) {
	return []byte(`{"name":"foo"}`), nil
}
