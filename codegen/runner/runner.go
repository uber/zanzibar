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
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/codegen"
	"go.uber.org/config"
	"io/ioutil"
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

type codegenOptions struct {
	PackageRoot string `yaml:"packageRoot"`
	StagingReqHeader string `yaml:"stagingReqHeader"`
	CopyrightHeaderRelPath string `yaml:"copyrightHeader"`
	RelThriftRootDir string `yaml:"thriftRootDir"`
	RelTargetGenDir string `yaml:"targetGenDir"`
	RelMiddlewareConfigDir string `yaml:"middlewareConfig"`
	AnnotationPrefix string `yaml:"annotationPrefix"`
	GenCodePackage string `yaml:"genCodePackage"`
	DeputyReqHeader string `yaml:"deputyReqHeader"`
	TraceKey string `yaml:"traceKey"`
	GenMock bool `yaml:"genMock"`

	// Deprecated
	ClientConfig string `yaml:"clientConfig"`
	EndpointConfig string `yaml:"endpointConfig"`
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
	configProvider, err := config.NewYAML(config.File(*configFile))
	checkError(
		err, fmt.Sprintf("can not read config %q", *configFile),
	)


	codegenConfig := new(codegenOptions)
	err = configProvider.Get("").Populate(&codegenConfig)
	checkError(
		err, fmt.Sprintf("can not parse config %q", *configFile),
	)

	// TODO(jacobg): Default value?
	if codegenConfig.StagingReqHeader == "" {
		codegenConfig.StagingReqHeader = "X-Zanzibar-Use-Staging"
	}
	if codegenConfig.DeputyReqHeader == "" {
		codegenConfig.StagingReqHeader = "x-deputy-forwarded"
	}

	var copyrightHeader string
	if codegenConfig.CopyrightHeaderRelPath != "" {
		configRoot, err = filepath.Abs(configRoot)
		checkError(
			err, fmt.Sprintf("can not get abs path of config dir %s", configRoot),
		)
		copyrightAbsPath := filepath.Join(configRoot, codegenConfig.CopyrightHeaderRelPath)
		bytes, err := ioutil.ReadFile(copyrightAbsPath)
		checkError(
			err, fmt.Sprintf("can not read copyright header file"),
		)
		copyrightHeader = string(bytes)
	}

	options := &codegen.PackageHelperOptions{
		RelThriftRootDir:       codegenConfig.RelThriftRootDir,
		RelTargetGenDir:        codegenConfig.RelTargetGenDir,
		RelMiddlewareConfigDir: codegenConfig.RelMiddlewareConfigDir,
		AnnotationPrefix:       codegenConfig.AnnotationPrefix,
		GenCodePackage:         codegenConfig.GenCodePackage,
		CopyrightHeader:        copyrightHeader,
		StagingReqHeader:       codegenConfig.StagingReqHeader,
		DeputyReqHeader:        codegenConfig.DeputyReqHeader,
		TraceKey:               codegenConfig.TraceKey,
	}

	packageHelper, err := codegen.NewPackageHelper(
		codegenConfig.PackageRoot,
		configRoot,
		options,
	)
	checkError(
		err, fmt.Sprintf("Can't build package helper %s", configRoot),
	)

	var moduleSystem *codegen.ModuleSystem
	if codegenConfig.GenMock {
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
