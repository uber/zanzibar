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
	"time"

	"github.com/uber/zanzibar/examples/example-gateway/build/clients/quux/module"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"
)

// Client is a custom client that does nothing yet
type Client interface{}

type quux struct {
	loggers map[string]*zap.Logger
	metrics map[string]*zanzibar.OutboundCustomMetrics
}

// NewClient creates a new Quux client, with client method loggers and metrics.
func NewClient(g *zanzibar.Gateway, deps *module.Dependencies) Client {
	logger, scope, loggers, metrics := zanzibar.SetupCustomClientLoggersMetrics(
		"quux", []string{"Foo"}, deps.Default,
	)

	logger.Debug("Created new Quux client")
	scope.Counter("created").Inc(1)

	return &quux{
		loggers: loggers,
		metrics: metrics,
	}
}

// Foo implementation, using client method logger and metrics
func (c *quux) Foo() (err error) {
	startTime := time.Now()
	defer c.metrics["Foo"].EmitOutboundCustomMetrics(startTime, err)
	c.loggers["Foo"].Debug("Called Foo")

	// Custom Foo implementation...

	return err
}
