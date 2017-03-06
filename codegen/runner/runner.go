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

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/uber/zanzibar/codegen"
	"github.com/uber/zanzibar/runtime"
)

var configFile = flag.String("config", "", "the config file path")

const templateDir = "./codegen/templates/*.tmpl"

func checkError(err error, message string) {
	if err != nil {
		fmt.Printf("%s: %s \n", message, err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	if *configFile == "" {
		flag.Usage()
		os.Exit(1)
		return
	}

	configDirName := filepath.Dir(*configFile)
	config := zanzibar.NewStaticConfigOrDie([]string{
		*configFile,
	}, nil)

	gatewaySpec, err := codegen.NewGatewaySpec(
		configDirName,
		filepath.Join(configDirName, config.MustGetString("thriftRootDir")),
		filepath.Join(configDirName, config.MustGetString("typeFileRootDir")),
		filepath.Join(configDirName, config.MustGetString("targetGenDir")),
		filepath.Join(
			configDirName, config.MustGetString("gatewayThriftRootDir"),
		),
		config.MustGetString("clientConfig"),
		config.MustGetString("endpointThriftDir"),
		config.MustGetString("endpointConfig"),
		config.MustGetString("gatewayName"),
	)
	checkError(
		err, fmt.Sprintf("can't create gateway spec %#v", gatewaySpec),
	)

	fmt.Printf("Generating clients for gateway \n")
	err = gatewaySpec.GenerateClients()
	checkError(err, "Failed to generate client files.")

	fmt.Printf("Generating endpoint code for gateway \n")
	err = gatewaySpec.GenerateEndpoints()
	checkError(err, "Failed to generate endpoint files.")

	fmt.Printf("Generating main.go for gateway \n")
	err = gatewaySpec.GenerateMain()
	checkError(err, "Failed to generate main.go.")
}
