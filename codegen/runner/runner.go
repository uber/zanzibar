// Copyright (c) 2019 Uber Technologies, Inc.
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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/codegen"
	"github.com/uber/zanzibar/runtime"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func checkError(err error, message string) {
	if err != nil {
		fmt.Printf("%s:\n%s\n", message, err)

		causeErr := errors.Cause(err)
		if causeErr, ok := causeErr.(stackTracer); ok {
			fmt.Printf("%+v \n", causeErr.StackTrace())
		}

		os.Exit(1)
	}
}

func main() {
	configFile := flag.String("config", "", "the config file path")
	moduleName := flag.String("instance", "", "")
	moduleClass := flag.String("type", "", "")
	flag.Parse()

	if *configFile == "" {
		flag.Usage()
		os.Exit(1)
		return
	}

	configRoot := filepath.Dir(*configFile)
	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(*configFile),
	}, nil)

	configRoot, err := filepath.Abs(configRoot)
	checkError(
		err, fmt.Sprintf("can not get abs path of config dir %s", configRoot),
	)

	copyright := []byte("")
	if config.ContainsKey("copyrightHeader") {
		bytes, err := ioutil.ReadFile(filepath.Join(
			configRoot,
			config.MustGetString("copyrightHeader"),
		))
		if err == nil {
			copyright = bytes
		}
	}

	stagingReqHeader := "X-Zanzibar-Use-Staging"
	if config.ContainsKey("stagingReqHeader") {
		stagingReqHeader = config.MustGetString("stagingReqHeader")
	}

	deputyReqHeader := "x-deputy-forwarded"
	if config.ContainsKey("deputyReqHeader") {
		deputyReqHeader = config.MustGetString("deputyReqHeader")
	}

	searchPaths := make([]string, 0)
	config.MustGetStruct("moduleSearchPaths", &searchPaths)
	options := &codegen.PackageHelperOptions{
		RelThriftRootDir:       config.MustGetString("thriftRootDir"),
		RelTargetGenDir:        config.MustGetString("targetGenDir"),
		RelMiddlewareConfigDir: config.MustGetString("middlewareConfig"),
		AnnotationPrefix:       config.MustGetString("annotationPrefix"),
		GenCodePackage:         config.MustGetString("genCodePackage"),
		CopyrightHeader:        string(copyright),
		StagingReqHeader:       stagingReqHeader,
		DeputyReqHeader:        deputyReqHeader,
		TraceKey:               config.MustGetString("traceKey"),
		ModuleSearchPaths:      searchPaths,
	}

	packageHelper, err := codegen.NewPackageHelper(
		config.MustGetString("packageRoot"),
		configRoot,
		options,
	)
	checkError(
		err, fmt.Sprintf("Can't build package helper %s", configRoot),
	)

	genMock := config.ContainsKey("genMock")
	if genMock {
		genMock = config.MustGetBoolean("genMock")
	}
	var moduleSystem *codegen.ModuleSystem
	if genMock {
		moduleSystem, err = codegen.NewDefaultModuleSystemWithMockHook(packageHelper)
	} else {
		moduleSystem, err = codegen.NewDefaultModuleSystem(packageHelper)
	}
	checkError(
		err, fmt.Sprintf("Error creating module system %s", configRoot),
	)

	fmt.Printf("Generating module system components:\n")

	if *moduleClass != "" && *moduleName != "" {
		_, err = moduleSystem.GenerateIncrementalBuild(
			packageHelper.PackageRoot(),
			configRoot,
			packageHelper.CodeGenTargetPath(),
			[]codegen.ModuleDependency{
				{
					ClassName:    *moduleClass,
					InstanceName: *moduleName,
				},
			},
			true,
		)
	} else {
		//lint:ignore SA1019 Migration to incremental builds is ongoing
		_, err = moduleSystem.GenerateBuild(
			packageHelper.PackageRoot(),
			configRoot,
			packageHelper.CodeGenTargetPath(),
			true,
		)
	}
	checkError(err, "Failed to generate module system components")
}
