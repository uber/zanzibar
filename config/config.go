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

package config

import (
	"os"

	"github.com/uber/zanzibar/runtime"
)

// EnvConfig map from environment variable to config key and data type
type EnvConfig map[string]struct {
	Key      string `json:"key"`
	DataType string `json:"dataType"`
}

// NewRuntimeConfigOrDie returns a static config struct
// that is pre-set with the service configuration defaults
// and overridden by service.env.config
func NewRuntimeConfigOrDie(
	files []string,
	seedConfig map[string]interface{},
) *zanzibar.StaticConfig {
	defaultConfig, err := defaultConfig()
	if err != nil {
		panic("error getting default config")
	}

	var serviceConfig []*zanzibar.ConfigOption
	serviceConfig = append(serviceConfig, defaultConfig)
	for _, configFilePath := range files {
		serviceConfig = append(serviceConfig, zanzibar.ConfigFilePath(configFilePath))
	}

	staticConfig := zanzibar.NewStaticConfigOrDie(serviceConfig, seedConfig)
	getEnvConfig(staticConfig)

	return staticConfig
}

func defaultConfig() (*zanzibar.ConfigOption, error) {
	bytes, err := Asset("production.json")
	if err != nil {
		return nil, err
	}
	return zanzibar.ConfigFileContents(bytes), nil
}

func getEnvConfig(cfg *zanzibar.StaticConfig) {
	var envConfig EnvConfig
	cfg.MustGetStruct("service.env.config", &envConfig)
	for envVar, configKey := range envConfig {
		if value, ok := os.LookupEnv(envVar); ok {
			cfg.SetConfigValue(configKey.Key, []byte(value), configKey.DataType)
		}
	}
}
