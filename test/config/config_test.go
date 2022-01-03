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

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/config"
)

const (
	uberPortHTTPEnv     = "UBER_PORT_HTTP"
	uberPortTChannelEnv = "UBER_PORT_TCHANNEL"
	httpPortKey         = "http.port"
	tchannelPortKey     = "tchannel.port"
)

func TestNewRuntimeConfigOrDie(t *testing.T) {
	httpPortValue := os.Getenv(uberPortHTTPEnv)
	defer func(key, value string) {
		_ = os.Setenv(key, value)
	}(uberPortHTTPEnv, httpPortValue)

	tchannelPortValue := os.Getenv(uberPortTChannelEnv)
	defer func(key, value string) {
		_ = os.Setenv(key, value)
	}(uberPortTChannelEnv, tchannelPortValue)

	_ = os.Setenv(uberPortHTTPEnv, "1111")
	_ = os.Setenv(uberPortTChannelEnv, "2222")
	cfg := config.NewRuntimeConfigOrDie([]string{"test.yaml"}, nil)

	assert.Equal(t, "my-gateway", cfg.MustGetString("serviceName")) // existing config
	assert.Equal(t, int64(1111), cfg.MustGetInt(httpPortKey))       // replaced config
	assert.Equal(t, int64(2222), cfg.MustGetInt(tchannelPortKey))   // new config
}
