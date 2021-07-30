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

package testcircuitbreaker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
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

// values set when no client override and no qps level
var defaultSettings map[string]int = map[string]int{
	"sleepWindowInMilliseconds": 5000,
	"maxConcurrentRequests":     20,
	"errorPercentThreshold":     20,
	"requestVolumeThreshold":    20,
}

// from test.yaml
var bazClientOverrides map[string]int = map[string]int{
	"timeout":               10000,
	"maxConcurrentRequests": 1000,
}

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
	// get qps config settings
	qpsConfig := GetQPSConfigSettings()
	// test circuit breaker settings for each method match prod values
	settings := hystrix.GetCircuitSettings()
	for circuitBreakerName := range settings {
		// checks circuit breaker names have method names and not service method names
		assert.True(t, !strings.Contains(circuitBreakerName, "::"), "Incorrect circuit breaker names")
	}
	names := [27]string{"baz-EchoBinary", "baz-EchoBool", "baz-EchoDouble", "baz-EchoEnum", "baz-EchoI16", "baz-EchoI32", "baz-EchoI64", "baz-EchoI8", "baz-EchoString", "baz-EchoStringList", "baz-EchoStringMap", "baz-EchoStringSet", "baz-EchoStructList", "baz-EchoStructSet", "baz-EchoTypedef", "baz-Call", "baz-Compare", "baz-GetProfile", "baz-HeaderSchema", "baz-Ping", "baz-DeliberateDiffNoop", "baz-TestUUID", "baz-Trans", "baz-TransHeaders", "baz-TransHeadersNoReq", "baz-TransHeadersType", "baz-URLTest"}
	count := 0
	// get methods to qpsLevels
	methodToQPSLevel := GetExpectedMethodToQPSLevels()
	// client overrides from test.yaml
	timeout := bazClientOverrides["timeout"]
	max := bazClientOverrides["maxConcurrentRequests"]
	for _, name := range names {
		// default values for parameters
		sleepWindow := defaultSettings["sleepWindowInMilliseconds"]
		errorPercentage := defaultSettings["errorPercentThreshold"]
		reqThreshold := defaultSettings["requestVolumeThreshold"]
		// config values set by qps level
		if val, ok := methodToQPSLevel[name]; ok {
			s := strconv.Itoa(val)
			sleepWindow = qpsConfig[s]["sleepWindowInMilliseconds"]
			errorPercentage = qpsConfig[s]["errorPercentThreshold"]
			reqThreshold = qpsConfig[s]["requestVolumeThreshold"]
			count++
		}
		expectedSettings := &hystrix.Settings{
			Timeout:                time.Duration(timeout) * time.Millisecond,
			MaxConcurrentRequests:  max,
			RequestVolumeThreshold: uint64(reqThreshold),
			SleepWindow:            time.Duration(sleepWindow) * time.Millisecond,
			ErrorPercentThreshold:  errorPercentage,
		}
		assert.Equal(t, settings[name], expectedSettings)
	}
	assert.Equal(t, len(methodToQPSLevel), count)
}

func GetQPSConfigSettings() map[string]map[string]int {
	pwd, _ := os.Getwd()
	pathToConfig := filepath.Join(pwd, "/qps-config.json")
	jsonFile, err := os.Open(pathToConfig)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var qpsConfig map[string]map[string]int
	json.Unmarshal(byteValue, &qpsConfig)
	return qpsConfig
}

func GetExpectedMethodToQPSLevels() map[string]int {
	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "/expected_method_qps_levels.json")
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var methodToQPSLevel map[string]int
	json.Unmarshal(byteValue, &methodToQPSLevel)
	return methodToQPSLevel
}
