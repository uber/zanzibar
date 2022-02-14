// Code generated by zanzibar
// @generated

// Copyright (c) 2018 Uber Technologies, Inc.
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

package main

import (
	"bytes"
	"context"
	"flag"
	"strings"

	"github.com/uber/zanzibar/config"

	"github.com/pkg/errors"

	_ "go.uber.org/automaxprocs"
	"go.uber.org/fx"
	"go.uber.org/zap"

	zanzibar "github.com/uber/zanzibar/runtime"

	app "github.com/uber/zanzibar/examples/example-gateway"
	service "github.com/uber/zanzibar/examples/example-gateway/build/services/echo-gateway"
	uberconfig "go.uber.org/config"
)

var configFiles *string

// Module defines the Zanzibar application module for Echo-gateway
var Module = fx.Options(
	fx.Provide(New),
	fx.Invoke(run),
)

func opts() fx.Option {
	return fx.Options(
		append(
			[]fx.Option{Module},
			app.GetOverrideFxOptions()...,
		)...,
	)
}

// Params defines the dependencies of the New module.
type Params struct {
	fx.In
	Lifecycle fx.Lifecycle
}

// Result defines the objects that the New module provides
type Result struct {
	fx.Out
	// Gateway corresponds to the fully built server gateway
	Gateway *zanzibar.Gateway
	// Provider is an abstraction over the Zanzibar config store
	Provider uberconfig.Provider `name:"zanzibarConfig"`
}

func main() {
	fx.New(opts()).Run()
}

// run is the main entry point for Echo-gateway
func run(gateway *zanzibar.Gateway) {
	gateway.Logger.Info("Started Echo-gateway",
		zap.String("realHTTPAddr", gateway.RealHTTPAddr),
		zap.String("realTChannelAddr", gateway.RealTChannelAddr),
		zap.Any("config", gateway.InspectOrDie()),
	)
}

// New exports functionality similar to Module, but allows the caller to wrap
// or modify Result. Most users should use Module instead.
func New(p Params) (Result, error) {
	readFlags()
	gateway, err := createGateway()
	if err != nil {
		return Result{}, errors.Wrap(err, "failed to create gateway server")
	}

	// Represent the zanzibar config in YAML that will be used to expose a config provider
	yamlCfg, err := gateway.Config.AsYaml()
	if err != nil {
		return Result{}, errors.Wrap(err, "unable to marshal Zanzibar config to YAML")
	}
	provider, err := uberconfig.NewYAML(
		[]uberconfig.YAMLOption{
			uberconfig.Source(bytes.NewReader(yamlCfg)),
		}...,
	)
	if err != nil {
		return Result{}, errors.Wrap(err, "unable to provide a YAML view from Zanzibar config")
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err = gateway.Bootstrap()
			if err != nil {
				panic(errors.Wrap(err, "failed to bootstrap gateway server"))
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			gateway.Logger.Info("fx OnStop() hook activated")
			gateway.WaitGroup.Add(1)
			gateway.Shutdown()
			gateway.WaitGroup.Done()
			return nil
		},
	})

	return Result{
		Gateway:  gateway,
		Provider: provider,
	}, nil
}

func createGateway() (*zanzibar.Gateway, error) {
	cfg := getConfig()

	if gateway, _, err := service.CreateGateway(cfg, app.AppOptions); err != nil {
		return nil, err
	} else {
		return gateway, nil
	}
}

func getConfig() *zanzibar.StaticConfig {
	var files []string

	if configFiles == nil {
		files = []string{}
	} else {
		files = strings.Split(*configFiles, ";")
	}

	return config.NewRuntimeConfigOrDie(files, nil)
}

func readFlags() {
	configFiles = flag.String(
		"config",
		"",
		"an ordered, semi-colon separated list of configuration files to use",
	)
	flag.Parse()
}
