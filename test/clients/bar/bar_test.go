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

package bar_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	barGen "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/bar/bar"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"

	"github.com/stretchr/testify/assert"

	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	zanzibar "github.com/uber/zanzibar/runtime"
)

var defaultTestOptions = &testGateway.Options{
	KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
	KnownTChannelBackends: []string{"baz"},
	ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
}
var defaultTestConfig = map[string]interface{}{
	"clients.baz.serviceName": "baz",
}

func TestHelloWorld(t *testing.T) {
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

	// note: we set clients.bar.followRedirect: false in test.yaml
	bgateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/hello",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Location", "http://example.com/")
			w.WriteHeader(303)
			_, err := w.Write([]byte(`hello world`))
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.Hello(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.Equal(t, new(barGen.SeeOthersRedirection), err)
	assert.Equal(t, "", result)
}

func TestEchoI8(t *testing.T) {
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

	var arg int8 = 42
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/i8",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoI8_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoI8(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoI8_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoI16(t *testing.T) {
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

	var arg int16 = 42
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/i16",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoI16_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoI16(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoI16_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoI32(t *testing.T) {
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

	var arg int32 = 42
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/i32",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoI32_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoI32(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoI32_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoI64(t *testing.T) {
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

	var arg int64 = 42
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/i64",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoI64_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoI64(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoI64_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoDouble(t *testing.T) {
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

	var arg = 42.0
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/double",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoDouble_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoDouble(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoDouble_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoBool(t *testing.T) {
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

	arg := true
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/bool",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoBool_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoBool(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoBool_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoString(t *testing.T) {
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

	arg := "hola"
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/string",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoString_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoString(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoString_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoBinary(t *testing.T) {
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

	arg := []byte{97} // "a"
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/binary",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoBinary_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(arg)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoBinary(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoBinary_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoEnum(t *testing.T) {
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

	apple := barGen.FruitApple
	arg := &apple
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/enum",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoEnum_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoEnum(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoEnum_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, barGen.FruitApple, result)
}

func TestEchoTypedef(t *testing.T) {
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

	var arg barGen.UUID = "uuid"
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/typedef",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoTypedef_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoTypedef(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoTypedef_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoStringSet(t *testing.T) {
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

	arg := map[string]struct{}{
		"a": {},
		"b": {},
	}
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/string-set",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoStringSet_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoStringSet(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoStringSet_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoStructSet(t *testing.T) {
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

	arg := []*barGen.BarResponse{
		{
			StringField:     "a",
			IntWithRange:    int32(0),
			IntWithoutRange: int32(0),
			MapIntWithRange: map[barGen.UUID]int32{
				barGen.UUID("fakeUUID"): int32(0),
			},
			MapIntWithoutRange: map[string]int32{
				"0": int32(0),
			},
			BinaryField: []byte("d29ybGQ="),
		},
	}
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/struct-set",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoStructSet_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoStructSet(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoStructSet_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoStringList(t *testing.T) {
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

	arg := []string{"a", "b"}
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/string-list",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoStringList_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoStringList(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoStringList_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoStructList(t *testing.T) {
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

	arg := []*barGen.BarResponse{
		{
			StringField:     "a",
			IntWithRange:    int32(0),
			IntWithoutRange: int32(0),
			MapIntWithRange: map[barGen.UUID]int32{
				barGen.UUID("fakeUUID"): int32(0),
			},
			MapIntWithoutRange: map[string]int32{
				"0": int32(0),
			},
			BinaryField: []byte("d29ybGQ="),
		},
	}
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/struct-list",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoStructList_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg[0], arg[0])

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoStructList(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoStructList_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoI32Map(t *testing.T) {
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

	arg := map[int32]*barGen.BarResponse{
		42: {
			StringField:     "a",
			IntWithRange:    int32(0),
			IntWithoutRange: int32(0),
			MapIntWithRange: map[barGen.UUID]int32{
				barGen.UUID("fakeUUID"): int32(0),
			},
			MapIntWithoutRange: map[string]int32{
				"0": int32(0),
			},
			BinaryField: []byte("d29ybGQ="),
		},
	}
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/i32-map",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoI32Map_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoI32Map(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoI32Map_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoStringMap(t *testing.T) {
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

	arg := map[string]*barGen.BarResponse{
		"a": {
			StringField:     "a",
			IntWithRange:    int32(0),
			IntWithoutRange: int32(0),
			MapIntWithRange: map[barGen.UUID]int32{
				barGen.UUID("fakeUUID"): int32(0),
			},
			MapIntWithoutRange: map[string]int32{
				"0": int32(0),
			},
			BinaryField: []byte("d29ybGQ="),
		},
	}
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/string-map",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoStringMap_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, req.Arg, arg)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.EchoStringMap(
		context.Background(), map[string]string{
			"x-uuid": "a-uuid",
		}, &barGen.Echo_EchoStringMap_Args{Arg: arg},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

//func TestEchoStructMap(t *testing.T) {
//	gateway, err := benchGateway.CreateGateway(
//		defaultTestConfig,
//		defaultTestOptions,
//		exampleGateway.CreateGateway,
//	)
//	if !assert.NoError(t, err) {
//		return
//	}
//	defer gateway.Close()
//
//	bgateway := gateway.(*benchGateway.BenchGateway)
//
//	arg := []struct {
//		Key   *barGen.BarResponse
//		Value string
//	}{
//		{
//			Key: &barGen.BarResponse{
//				StringField:     "a",
//				IntWithRange:    int32(0),
//				IntWithoutRange: int32(0),
//				MapIntWithRange: map[barGen.UUID]int32{
//					barGen.UUID("fakeUUID"): int32(0),
//				},
//				MapIntWithoutRange: map[string]int32{
//					"0": int32(0),
//				},
//				BinaryField: []byte("d29ybGQ=")},
//			Value: "a",
//		},
//	}
//	marshaled, err := json.Marshal(arg)
//	assert.NoError(t, err)
//
//	bgateway.HTTPBackends()["bar"].HandleFunc(
//		"POST", "/echo/struct-map",
//		func(w http.ResponseWriter, r *http.Request) {
//			body, err := ioutil.ReadAll(r.Body)
//			assert.NoError(t, err)
//
//			err = r.Body.Close()
//			assert.NoError(t, err)
//
//			var req barGen.Echo_EchoStructMap_Args
//			err = json.Unmarshal(body, &req)
//			assert.NoError(t, err)
//			assert.Equal(t, req.Arg, arg)
//
//			w.WriteHeader(200)
//			_, err = w.Write(marshaled)
//			assert.NoError(t, err)
//		},
//	)
//	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
//	bar := deps.Client.Bar
//
//	result, _, err := bar.EchoStructMap(
//		context.Background(), map[string]string{
//			"x-uuid": "a-uuid",
//		}, &barGen.Echo_EchoStructMap_Args{Arg: arg},
//	)
//	assert.NoError(t, err)
//	assert.Equal(t, arg, result)
//}

func TestNestedQueryParamCallWithNil(t *testing.T) {
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

	_, result, _, err := bar.ArgWithNestedQueryParams(
		context.Background(), nil, &barGen.Bar_ArgWithNestedQueryParams_Args{
			Request: nil,
		},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NotNil(t, err)
	assert.Equal(t, "The field .Request is required", err.Error())
	assert.Nil(t, result)
}

func TestNormalRecur(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	arg := barGen.BarRequestRecur{
		Name: "parent",
		Recur: &barGen.BarRequestRecur{
			Name: "child",
			Recur: &barGen.BarRequestRecur{
				Name:  "grandchild",
				Recur: nil,
			},
		},
	}

	bgateway := gateway.(*benchGateway.BenchGateway)
	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/bar/recur",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Bar_NormalRecur_Args
			err = json.Unmarshal(body, &req)
			assert.NoError(t, err)
			assert.Equal(t, *req.Request, arg)

			res := barGen.BarResponseRecur{
				Nodes:  make([]string, 0),
				Height: 0,
			}
			for node := req.Request; node != nil; node = node.Recur {
				res.Nodes = append(res.Nodes, node.Name)
				res.Height++
			}
			marshaled, err := json.Marshal(res)
			assert.NoError(t, err)

			w.WriteHeader(200)
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, result, _, err := bar.NormalRecur(
		context.Background(),
		nil,
		&barGen.Bar_NormalRecur_Args{
			Request: &arg,
		},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, int32(3), result.Height)
	assert.Len(t, result.Nodes, 3)
	assert.Equal(t, "parent", result.Nodes[0])
	assert.Equal(t, "child", result.Nodes[1])
	assert.Equal(t, "grandchild", result.Nodes[2])
}

func TestDeleteFoo(t *testing.T) {
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
		"DELETE", "/bar/foo",
		func(w http.ResponseWriter, r *http.Request) {
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	_, _, err = bar.DeleteFoo(
		context.Background(),
		map[string]string{"x-uuid": "a-uuid"},
		&barGen.Bar_DeleteFoo_Args{UserUUID: "a-uuid"},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
}

func TestDeleteWithQueryParams(t *testing.T) {
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
		"DELETE", "/bar/withQueryParams",
		func(w http.ResponseWriter, r *http.Request) {
			qparams := r.URL.Query()
			assert.Equal(t, "foo", qparams["filter"][0])
			assert.Equal(t, "3", qparams["count"][0])
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar
	var count int32 = 3

	_, _, err = bar.DeleteWithQueryParams(
		context.Background(),
		nil,
		&barGen.Bar_DeleteWithQueryParams_Args{Filter: "foo", Count: &count},
		&zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2) * time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
}
