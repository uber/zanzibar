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

package testcircuitbreaker

import (
	"strings"
	"testing"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/stretchr/testify/assert"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"

	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
)

func TestCircuitBreakerSettings(t *testing.T) {
	// create gateway to get circuit breakers configured
	_, err := benchGateway.CreateGateway(
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
		t.Error("got bootstrap err: " + err.Error())
		return
	}
	// get circuit breaker settings
	settings := hystrix.GetCircuitSettings()
	for circuitBreakerName := range settings {
		// checks circuit breaker names have method names and not service method names
		assert.True(t, !strings.Contains(circuitBreakerName, "::"), "Incorrect circuit breaker names")
		// checks circuit breaker names contain '-' which separates client id from method name
		assert.True(t, strings.Contains(circuitBreakerName, "-"), "Circuit breaker name should have '-")
	}
	// circuit breaker names for multi client
	multiCircuitBreakerNames := [2]string{"multi-HelloA", "multi-HelloB"}
	// map circuit breaker name to qps level
	methodToQPSLevel := map[string]string{
		"multi-HelloA": "default",
		"multi-HelloB": "default",
	}
	// circuit breaker parameters from test.yaml
	circuitBreakerConfig := map[string]map[string]int{
		"default": {
			"sleepWindowInMilliseconds": 5000,
			"errorPercentThreshold":     20,
			"requestVolumeThreshold":    20,
			"maxConcurrentRequests":     20,
		},
		"1": {
			"sleepWindowInMilliseconds": 7000,
			"errorPercentThreshold":     10,
			"requestVolumeThreshold":    15,
			"maxConcurrentRequests":     20,
		},
	}
	// client overrides for multi client in test.yaml
	multiClientOverrides := map[string]int{
		"timeout":               10000,
		"maxConcurrentRequests": 1000,
		"errorPercentThreshold": 20,
	}
	// loop through circuit breaker names to create expected settings using
	// circuit breaker configurations
	for _, name := range multiCircuitBreakerNames {

		var sleepWindowInMilliseconds int
		var requestVolumeThreshold int

		if val, ok := methodToQPSLevel[name]; ok {
			// these values have no client overrides
			sleepWindowInMilliseconds = circuitBreakerConfig[val]["sleepWindowInMilliseconds"]
			requestVolumeThreshold = circuitBreakerConfig[val]["requestVolumeThreshold"]
		}
		// client overrides: if client set configurations for circuit breakers
		// override values are used for all client's circuit breakers
		timeout := multiClientOverrides["timeout"]
		maxConcurrentRequests := multiClientOverrides["maxConcurrentRequests"]
		errorPercentThreshold := multiClientOverrides["errorPercentThreshold"]

		expectedSettings := &hystrix.Settings{
			Timeout:                time.Duration(timeout) * time.Millisecond,
			MaxConcurrentRequests:  maxConcurrentRequests,
			RequestVolumeThreshold: uint64(requestVolumeThreshold),
			SleepWindow:            time.Duration(sleepWindowInMilliseconds) * time.Millisecond,
			ErrorPercentThreshold:  errorPercentThreshold,
		}
		// compare expected circuit breaker settings for multi client with actual settings
		assert.Equal(t, settings[name], expectedSettings)
	}
	// circuit breakers with default qps levels
	bazCircuitBreakers := [17]string{"baz-EchoBinary", "baz-EchoBool", "baz-EchoDouble", "baz-EchoEnum", "baz-EchoI16", "baz-EchoI32", "baz-EchoI64", "baz-EchoI8", "baz-EchoString", "baz-EchoStringList", "baz-EchoStringMap", "baz-EchoStringSet", "baz-EchoStructList", "baz-EchoStructSet", "baz-EchoTypedef", "baz-TestUUID", "baz-URLTest"}

	for _, name := range bazCircuitBreakers {
		// client overrides for baz
		timeout := 10000
		maxConcurrentRequests := 1000
		requestVolumeThreshold := 20
		sleepWindowInMilliseconds := 5000
		errorPercentThreshold := 20
		expectedSettings := &hystrix.Settings{
			Timeout:                time.Duration(timeout) * time.Millisecond,
			MaxConcurrentRequests:  maxConcurrentRequests,
			RequestVolumeThreshold: uint64(requestVolumeThreshold),
			SleepWindow:            time.Duration(sleepWindowInMilliseconds) * time.Millisecond,
			ErrorPercentThreshold:  errorPercentThreshold,
		}
		// compare expected circuit breaker settings for baz client with actual settings
		assert.Equal(t, settings[name], expectedSettings)
	}
}
