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

package baz_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	bazClient "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/baz/base"
	bazGen "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/baz/baz"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	zanzibar "github.com/uber/zanzibar/runtime"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

var defaultTestOptions = &testGateway.Options{
	KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
	KnownTChannelBackends: []string{"baz"},
	ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
}
var defaultTestConfig = map[string]interface{}{
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
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoI8_Args,
	) (int8, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoI8", "SecondService::echoI8",
		bazClient.NewSecondServiceEchoI8Handler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoI8(
		context.Background(), nil, &bazGen.SecondService_EchoI8_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoI16_Args,
	) (int16, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoI16", "SecondService::echoI16",
		bazClient.NewSecondServiceEchoI16Handler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoI16(
		context.Background(), nil, &bazGen.SecondService_EchoI16_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoI32_Args,
	) (int32, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoI32", "SecondService::echoI32",
		bazClient.NewSecondServiceEchoI32Handler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoI32(
		context.Background(), nil, &bazGen.SecondService_EchoI32_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoI64_Args,
	) (int64, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoI64", "SecondService::echoI64",
		bazClient.NewSecondServiceEchoI64Handler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoI64(
		context.Background(), nil, &bazGen.SecondService_EchoI64_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoDouble_Args,
	) (float64, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoDouble", "SecondService::echoDouble",
		bazClient.NewSecondServiceEchoDoubleHandler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoDouble(
		context.Background(), nil, &bazGen.SecondService_EchoDouble_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoBool_Args,
	) (bool, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoBool", "SecondService::echoBool",
		bazClient.NewSecondServiceEchoBoolHandler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoBool(
		context.Background(), nil, &bazGen.SecondService_EchoBool_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoString_Args,
	) (string, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoString", "SecondService::echoString",
		bazClient.NewSecondServiceEchoStringHandler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoString(
		context.Background(), nil, &bazGen.SecondService_EchoString_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoBinary_Args,
	) ([]byte, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoBinary", "SecondService::echoBinary",
		bazClient.NewSecondServiceEchoBinaryHandler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoBinary(
		context.Background(), nil, &bazGen.SecondService_EchoBinary_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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

	apple := bazGen.FruitApple
	arg := &apple
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoEnum_Args,
	) (bazGen.Fruit, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return *arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoEnum", "SecondService::echoEnum",
		bazClient.NewSecondServiceEchoEnumHandler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoEnum(
		context.Background(), nil, &bazGen.SecondService_EchoEnum_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: zanzibar.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, *arg, result)
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

	var arg base.UUID = "uuid"
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoTypedef_Args,
	) (base.UUID, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoTypedef", "SecondService::echoTypedef",
		bazClient.NewSecondServiceEchoTypedefHandler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoTypedef(
		context.Background(), nil, &bazGen.SecondService_EchoTypedef_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoStringSet_Args,
	) (map[string]struct{}, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoStringSet", "SecondService::echoStringSet",
		bazClient.NewSecondServiceEchoStringSetHandler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoStringSet(
		context.Background(), nil, &bazGen.SecondService_EchoStringSet_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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

	arg := []*base.BazResponse{
		{
			Message: "a",
		},
	}
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoStructSet_Args,
	) ([]*base.BazResponse, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoStructSet", "SecondService::echoStructSet",
		bazClient.NewSecondServiceEchoStructSetHandler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoStructSet(
		context.Background(), nil, &bazGen.SecondService_EchoStructSet_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoStringList_Args,
	) ([]string, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoStringList", "SecondService::echoStringList",
		bazClient.NewSecondServiceEchoStringListHandler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoStringList(
		context.Background(), nil, &bazGen.SecondService_EchoStringList_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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

	arg := []*base.BazResponse{
		{
			Message: "a",
		},
	}
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoStructList_Args,
	) ([]*base.BazResponse, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoStructList", "SecondService::echoStructList",
		bazClient.NewSecondServiceEchoStructListHandler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoStructList(
		context.Background(), nil, &bazGen.SecondService_EchoStructList_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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

	arg := map[string]*base.BazResponse{
		"a": {
			Message: "a",
		},
	}
	fake := func(
		ctx context.Context,
		reqHeaders map[string]string,
		args *bazGen.SecondService_EchoStringMap_Args,
	) (map[string]*base.BazResponse, map[string]string, error) {
		assert.Equal(t, arg, args.Arg)
		return arg, nil, nil
	}

	err = bgateway.TChannelBackends()["baz"].Register(
		"baz", "echoStringMap", "SecondService::echoStringMap",
		bazClient.NewSecondServiceEchoStringMapHandler(fake),
	)
	assert.NoError(t, err)
	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	baz := deps.Client.Baz

	_, result, _, err := baz.EchoStringMap(
		context.Background(), nil, &bazGen.SecondService_EchoStringMap_Args{Arg: arg}, &zanzibar.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
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
//		Key   *base.BazResponse
//		Value string
//	}{
//		{
//			Key: &base.BazResponse{
//				Message: "a",
//			},
//			Value: "a",
//		},
//	}
//	fake := func(
//		ctx context.Context,
//		reqHeaders map[string]string,
//		args *bazGen.SecondService_EchoStructMap_Args,
//	) ([]struct {
//		Key   *base.BazResponse
//		Value string
//	}, map[string]string, error) {
//		assert.Equal(t, arg, args.Arg)
//		return arg, nil, nil
//	}
//
//	err = bgateway.TChannelBackends()["baz"].Register(
//		"baz", "echoStructMap", "SecondService::echoStructMap",
//		bazClient.NewSecondServiceEchoStructMapHandler(fake),
//	)
//	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
//	baz := deps.Client.Baz
//
//	result, _, err := baz.EchoStructMap(
//		context.Background(), nil, &bazGen.SecondService_EchoStructMap_Args{Arg: arg},
//	)
//	assert.NoError(t, err)
//	assert.Equal(t, arg, result)
//}
