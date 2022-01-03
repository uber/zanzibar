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

package baz

import (
	"bytes"
	"context"
	"testing"

	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"

	bazServer "github.com/uber/zanzibar/examples/example-gateway/build/clients/baz"

	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
)

var testSillyNoopCounter int

func sillyNoop(ctx context.Context, reqHeaders map[string]string) (map[string]string, error) {
	testSillyNoopCounter++
	return nil, nil

}

func BenchmarkSillyNoop(b *testing.B) {
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
		"baz", "sillyNoop", "SimpleService::sillyNoop",
		bazServer.NewSimpleServiceSillyNoopHandler(sillyNoop),
	)
	if err != nil {
		b.Error("got register err: " + err.Error())
		return
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			res, err := gateway.MakeRequest(
				"GET", "/baz/silly-noop", nil,
				bytes.NewReader([]byte(`{}`)),
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

	b.StopTimer()
	gateway.Close()
}
