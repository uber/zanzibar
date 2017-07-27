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

package bar_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	barGen "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/bar/bar"
	"github.com/uber/zanzibar/test/lib/bench_gateway"
	"github.com/uber/zanzibar/test/lib/test_gateway"

	"github.com/stretchr/testify/assert"

	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
)

var defaultTestOptions *testGateway.Options = &testGateway.Options{
	KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
	KnownTChannelBackends: []string{"baz"},
}
var defaultTestConfig map[string]interface{} = map[string]interface{}{
	"clients.baz.serviceName": "baz",
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

	result, _, err := bar.EchoI8(
		context.Background(), nil, &barGen.Echo_EchoI8_Args{Arg: arg},
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

	result, _, err := bar.EchoI16(
		context.Background(), nil, &barGen.Echo_EchoI16_Args{Arg: arg},
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

	result, _, err := bar.EchoI32(
		context.Background(), nil, &barGen.Echo_EchoI32_Args{Arg: arg},
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

	result, _, err := bar.EchoI64(
		context.Background(), nil, &barGen.Echo_EchoI64_Args{Arg: arg},
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

	var arg float64 = 42.0
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

	result, _, err := bar.EchoDouble(
		context.Background(), nil, &barGen.Echo_EchoDouble_Args{Arg: arg},
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

	result, _, err := bar.EchoBool(
		context.Background(), nil, &barGen.Echo_EchoBool_Args{Arg: arg},
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

	result, _, err := bar.EchoString(
		context.Background(), nil, &barGen.Echo_EchoString_Args{Arg: arg},
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
	marshaled, err := json.Marshal(arg)
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
			_, err = w.Write(marshaled)
			assert.NoError(t, err)
		},
	)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bar := deps.Client.Bar

	result, _, err := bar.EchoBinary(
		context.Background(), nil, &barGen.Echo_EchoBinary_Args{Arg: arg},
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

	result, _, err := bar.EchoEnum(
		context.Background(), nil, &barGen.Echo_EchoEnum_Args{Arg: arg},
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

	result, _, err := bar.EchoTypedef(
		context.Background(), nil, &barGen.Echo_EchoTypedef_Args{Arg: arg},
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

	result, _, err := bar.EchoStringSet(
		context.Background(), nil, &barGen.Echo_EchoStringSet_Args{Arg: arg},
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
			MapIntWithRange: map[string]int32{
				"0": int32(0),
			},
			MapIntWithoutRange: map[string]int32{
				"0": int32(0),
			},
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

	result, _, err := bar.EchoStructSet(
		context.Background(), nil, &barGen.Echo_EchoStructSet_Args{Arg: arg},
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

	result, _, err := bar.EchoStringList(
		context.Background(), nil, &barGen.Echo_EchoStringList_Args{Arg: arg},
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
			MapIntWithRange: map[string]int32{
				"0": int32(0),
			},
			MapIntWithoutRange: map[string]int32{
				"0": int32(0),
			},
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

	result, _, err := bar.EchoStructList(
		context.Background(), nil, &barGen.Echo_EchoStructList_Args{Arg: arg},
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
			MapIntWithRange: map[string]int32{
				"0": int32(0),
			},
			MapIntWithoutRange: map[string]int32{
				"0": int32(0),
			},
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

	result, _, err := bar.EchoStringMap(
		context.Background(), nil, &barGen.Echo_EchoStringMap_Args{Arg: arg},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}

func TestEchoStructMap(t *testing.T) {
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

	arg := []struct {
		Key   *barGen.BarResponse
		Value string
	}{
		{
			Key: &barGen.BarResponse{
				StringField:     "a",
				IntWithRange:    int32(0),
				IntWithoutRange: int32(0),
				MapIntWithRange: map[string]int32{
					"0": int32(0),
				},
				MapIntWithoutRange: map[string]int32{
					"0": int32(0),
				}},
			Value: "a",
		},
	}
	marshaled, err := json.Marshal(arg)
	assert.NoError(t, err)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"POST", "/echo/struct-map",
		func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			err = r.Body.Close()
			assert.NoError(t, err)

			var req barGen.Echo_EchoStructMap_Args
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

	result, _, err := bar.EchoStructMap(
		context.Background(), nil, &barGen.Echo_EchoStructMap_Args{Arg: arg},
	)
	assert.NoError(t, err)
	assert.Equal(t, arg, result)
}
