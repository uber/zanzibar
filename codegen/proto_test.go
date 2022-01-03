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

package codegen

import (
	"strings"
	"testing"

	"github.com/emicklei/proto"
	"github.com/stretchr/testify/assert"
)

const (
	singleServiceSpec = `
		syntax = "proto3";
		package echo;
	
		message Request { string message = 1; }
		message Response { string message = 1; }
	
		service EchoService {
	        rpc EchoMethod(Request) returns (Response);
		}
	`
	multiServiceSpec = `
		syntax = "proto3";
		package echo;
	
		message Request1 { string message = 1; }
		message Response1 { string message = 1; }
		message Request2 { string message = 1; }
		message Response2 { string message = 1; }
		
		service EchoService {
	        rpc EchoMethod1(Request1) returns (Response1);
			rpc EchoMethod2(Request2) returns (Response2);
		}
	`
	mixedServiceSpec = `
		syntax = "proto3";
		package echo;
	
		message Request1 { string message = 1; }
		message Response1 { string message = 1; }
		message Response2 { string message = 1; }
		
		service EchoService {
	        rpc EchoMethod1(Request1) returns (Response1);
			rpc EchoMethod2(Request1) returns (Response2);
		}
	`
	noServiceSpec = `
		syntax = "proto3";
		package echo;
	
		message Request { string message = 1; }
		message Response { string message = 1; }
	`
	emptyServiceSpec = `
		syntax = "proto3";
		package echo;
	
		message Request { string message = 1; }
		message Response { string message = 1; }
	
		service EchoService {}
	`
)

var (
	singleServiceSpecList = &ProtoModule{
		PackageName: "echo",
		Services: []*ProtoService{{
			Name: "EchoService",
			RPC: []*ProtoRPC{
				{
					Name: "EchoMethod",
					Request: &ProtoMessage{
						Name: "Request",
					},
					Response: &ProtoMessage{
						Name: "Response",
					},
				},
			},
		}},
	}
	multiServiceSpecList = &ProtoModule{
		PackageName: "echo",
		Services: []*ProtoService{{
			Name: "EchoService",
			RPC: []*ProtoRPC{
				{
					Name: "EchoMethod1",
					Request: &ProtoMessage{
						Name: "Request1",
					},
					Response: &ProtoMessage{
						Name: "Response1",
					},
				},
				{
					Name: "EchoMethod2",
					Request: &ProtoMessage{
						Name: "Request2",
					},
					Response: &ProtoMessage{
						Name: "Response2",
					},
				},
			},
		}},
	}
	mixedServiceSpecList = &ProtoModule{
		PackageName: "echo",
		Services: []*ProtoService{{
			Name: "EchoService",
			RPC: []*ProtoRPC{
				{
					Name: "EchoMethod1",
					Request: &ProtoMessage{
						Name: "Request1",
					},
					Response: &ProtoMessage{
						Name: "Response1",
					},
				},
				{
					Name: "EchoMethod2",
					Request: &ProtoMessage{
						Name: "Request1",
					},
					Response: &ProtoMessage{
						Name: "Response2",
					},
				},
			},
		}},
	}
	noServiceSpecList = &ProtoModule{
		PackageName: "echo",
	}
	emptyServiceSpecList = &ProtoModule{
		PackageName: "echo",
		Services: []*ProtoService{{
			Name: "EchoService",
			RPC:  make([]*ProtoRPC, 0),
		}},
	}
)

func TestRunner(t *testing.T) {
	assertElementMatch(t, singleServiceSpec, singleServiceSpecList)
	assertElementMatch(t, multiServiceSpec, multiServiceSpecList)
	assertElementMatch(t, mixedServiceSpec, mixedServiceSpecList)
	assertElementMatch(t, noServiceSpec, noServiceSpecList)
	assertElementMatch(t, emptyServiceSpec, emptyServiceSpecList)
}

func assertElementMatch(t *testing.T, specRaw string, specParsed *ProtoModule) {
	r := strings.NewReader(specRaw)
	parser := proto.NewParser(r)
	protoSpec, err := parser.Parse()
	assert.NoErrorf(t, err, "proto spec parsing failed")

	v := newVisitor().Visit(protoSpec)
	assert.Equal(t, specParsed.PackageName, v.PackageName)
	assert.ElementsMatch(t, specParsed.Services, v.Services)
}
