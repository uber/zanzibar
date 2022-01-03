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

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/uber/zanzibar/codegen"
	"go.uber.org/thriftrw/compile"
)

// Given a thrift file, prints the thrift files that need to have json (un)marshallers
// The thrift-generated Go structs that need json (un)marshallers are:
// 1. the wrapper struct that wraps method arguments, generated from given thrift;
// 2. the argument struct for a method with unwrap annotation, generated from given thrift or its includes;
// 3. the return struct, generated from given thrift or its includes;
// 4. the exception struct, generated from given thrift or its includes;
func main() {
	thriftFile := os.Args[1]
	if !strings.HasSuffix(thriftFile, ".thrift") {
		fmt.Printf("Skipping compilation: %s is not a thrift file", thriftFile)
		return
	}

	module, err := compile.Compile(thriftFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse thrift file: %s", thriftFile))
	}
	thisThriftPath := module.ThriftPath
	fmt.Println(thisThriftPath)

	var unwrapAnnot = fmt.Sprintf(codegen.AntHTTPReqDefBoxed, os.Args[2])

	for _, service := range module.Services {
		for _, method := range service.Functions {
			if method.Annotations[unwrapAnnot] == "true" {
				if len(method.ArgsSpec) < 1 {
					panic(fmt.Sprintf("Annotation %q found on method %q, but no argument found: %s",
						unwrapAnnot, method.Name, service.ThriftFile(),
					))
				}

				argThriftPath := method.ArgsSpec[0].Type.ThriftFile()
				if argThriftPath != "" && argThriftPath != thisThriftPath {
					fmt.Println(argThriftPath)
				}

			}

			if method.ResultSpec == nil {
				continue
			}

			if method.ResultSpec.ReturnType != nil {
				returnThriftPath := method.ResultSpec.ReturnType.ThriftFile()
				if returnThriftPath != "" && returnThriftPath != thisThriftPath {
					fmt.Println(returnThriftPath)
				}
			}

			for _, exp := range method.ResultSpec.Exceptions {
				exceptionThriftPath := exp.Type.ThriftFile()
				if exceptionThriftPath != thisThriftPath {
					fmt.Println(exceptionThriftPath)
				}
			}
		}
	}
}
