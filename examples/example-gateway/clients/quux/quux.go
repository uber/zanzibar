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

package quux

import (
	barClient "github.com/uber/zanzibar/examples/example-gateway/build/clients/bar"
	module "github.com/uber/zanzibar/examples/example-gateway/build/clients/quux/module"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Client is a custom client that does nothing yet
type Client interface{}

type quux struct {
	client *zanzibar.TChannelClient
	bar    barClient.Client
}

// NewClient creates a new Quux client
func NewClient(g *zanzibar.Gateway, deps *module.Dependencies) Client {
	return &quux{
		client: zanzibar.NewTChannelClient(
			deps.Default.Channel,
			deps.Default.Logger,
			deps.Default.Scope,
			&zanzibar.TChannelClientOption{
				ServiceName: "quux",
				ClientID:    "quux",
				MethodNames: map[string]string{
					"Quux::Foo": "Foo",
				},
			},
		),
		bar: deps.Client.Bar,
	}
}

func (c *quux) Foo() {
	logger := c.client.Loggers["Quux::Foo"]
	_, _, err := c.bar.Hello(context.Background(), nil)
	if err != nil {
		logger.Error("hello error", zap.Error(err))
	}
	_, _, err = c.client.Call(context.Background(), "quux", "foo", nil, nil, nil)
	if err != nil {
		logger.Error("client call error", zap.Error(err))
	}
	c.client.Scopes["Quux::Foo"].Counter("foo.called").Inc(1)
}
