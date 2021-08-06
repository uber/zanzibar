// Copyright (c) 2021 Uber Technologies, Inc.
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

package baz

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/isopropylcyanide/hystrix-go/hystrix"
	"github.com/stretchr/testify/assert"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"

	bazServer "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/baz/baz"

	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
)

var testCallCounter int

func call(
	ctx context.Context, reqHeaders map[string]string, args *baz.SimpleService_Call_Args,
) (map[string]string, error) {
	testCallCounter++
	r := args.Arg
	if r.B1 && r.S2 == "hello" && r.I3 == 42 {
		return nil, nil
	}
	return nil, errors.New("Wrong Args")
}

func BenchmarkCall(b *testing.B) {
	gateway, err := benchGateway.CreateGateway(
		map[string]interface{}{
			"clients.baz.serviceName": "baz",
		},
		&testGateway.Options{
			KnownHTTPBackends:     []string{"bar", "contacts", "google-now"},
			KnownTChannelBackends: []string{"baz"},
			ConfigFiles:           util.DefaultConfigFiles("example-gateway"),
		},
		exampleGateway.CreateGateway,
	)
	if err != nil {
		b.Error("got bootstrap err: " + err.Error())
		return
	}

	err = gateway.TChannelBackends()["baz"].Register(
		"baz", "call", "SimpleService::call",
		bazServer.NewSimpleServiceCallHandler(call),
	)
	if err != nil {
		b.Error("got register err: " + err.Error())
		return
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := gateway.MakeRequest(
				"POST", "/baz/call", map[string]string{
					"x-token": "token",
					"x-uuid":  "uuid",
				},
				bytes.NewReader([]byte(`{"arg":{"b1":true,"s2":"hello","i3":42}}`)),
			)
			if err != nil {
				b.Error("got http error: " + err.Error())
				break
			}
			if res.Status != "204 No Content" {
				b.Error("got bad status error: " + res.Status)
				break
			}
		}
	})

	// test circuit breaker settings for each method match prod values
	settings := hystrix.GetCircuitSettings()
	names := [27]string{"baz-EchoBinary", "baz-EchoBool", "baz-EchoDouble", "baz-EchoEnum", "baz-EchoI16", "baz-EchoI32", "baz-EchoI64", "baz-EchoI8", "baz-EchoString", "baz-EchoStringList", "baz-EchoStringMap", "baz-EchoStringSet", "baz-EchoStructList", "baz-EchoStructSet", "baz-EchoTypedef", "baz-Call", "baz-Compare", "baz-GetProfile", "baz-HeaderSchema", "baz-Ping", "baz-DeliberateDiffNoop", "baz-TestUUID", "baz-Trans", "baz-TransHeaders", "baz-TransHeadersNoReq", "baz-TransHeadersType", "baz-URLTest"}
	// baz config values from production.json
	timeout := 10000
	max := 1000
	sleepWindow := 5000
	errorPercentage := 20
	reqThreshold := 20
	expectedSettings := &hystrix.Settings{
		Timeout:                time.Duration(timeout) * time.Millisecond,
		MaxConcurrentRequests:  max,
		RequestVolumeThreshold: uint64(reqThreshold),
		SleepWindow:            time.Duration(sleepWindow) * time.Millisecond,
		ErrorPercentThreshold:  errorPercentage,
	}
	for _, name := range names {
		assert.Equal(b, settings[name], expectedSettings)
	}

	b.StopTimer()
	gateway.Close()
}
