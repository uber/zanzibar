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
	"github.com/emicklei/proto"
)

// ProtoModule is an internal representation of a parsed Proto file.
type ProtoModule struct {
	PackageName string
	FilePath    string
	Imports     []string
	Services    Services
}

// Services is list of ProtoServices
type Services []*ProtoService

func (a Services) Len() int {
	return len(a)
}

func (a Services) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Services) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}

// ProtoService is an internal representation of Proto service and methods in that service.
type ProtoService struct {
	Name string
	RPC  []*ProtoRPC
}

// ProtoRPC is an internal representation of Proto RPC method and its request/response types.
type ProtoRPC struct {
	Name     string
	Request  *ProtoMessage
	Response *ProtoMessage
}

// ProtoMessage is an internal representation of a Proto Message.
type ProtoMessage struct {
	Name string
}

type visitor struct {
	Module *ProtoModule
}

func newVisitor() *visitor {
	return &visitor{
		Module: new(ProtoModule),
	}
}

func (v *visitor) Visit(proto *proto.Proto) *ProtoModule {
	v.Module = &ProtoModule{
		FilePath: proto.Filename,
		Services: make([]*ProtoService, 0),
		Imports:  make([]string, 0),
	}
	for _, e := range proto.Elements {
		e.Accept(v)
	}
	return v.Module
}

func (v *visitor) VisitService(e *proto.Service) {
	v.Module.Services = append(v.Module.Services, &ProtoService{
		Name: e.Name,
		RPC:  make([]*ProtoRPC, 0),
	})
	for _, c := range e.Elements {
		c.Accept(v)
	}
}

func (v *visitor) VisitRPC(r *proto.RPC) {
	s := v.Module.Services[len(v.Module.Services)-1]
	s.RPC = append(s.RPC, &ProtoRPC{
		Name:     r.Name,
		Request:  &ProtoMessage{Name: r.RequestType},
		Response: &ProtoMessage{Name: r.ReturnsType},
	})
}

func (v *visitor) VisitPackage(e *proto.Package) {
	v.Module.PackageName = e.Name
}

func (v *visitor) VisitImport(e *proto.Import) {
	v.Module.Imports = append(v.Module.Imports, e.Filename)
}

// From the current use case, the following visits are no-op
// since we only require the service, rpc methods and the request/response
// types of those methods.

func (v *visitor) VisitMessage(e *proto.Message)         {}
func (v *visitor) VisitSyntax(e *proto.Syntax)           {}
func (v *visitor) VisitOption(e *proto.Option)           {}
func (v *visitor) VisitNormalField(e *proto.NormalField) {}
func (v *visitor) VisitEnumField(e *proto.EnumField)     {}
func (v *visitor) VisitEnum(e *proto.Enum)               {}
func (v *visitor) VisitComment(e *proto.Comment)         {}
func (v *visitor) VisitOneof(o *proto.Oneof)             {}
func (v *visitor) VisitOneofField(o *proto.OneOfField)   {}
func (v *visitor) VisitReserved(r *proto.Reserved)       {}
func (v *visitor) VisitMapField(f *proto.MapField)       {}
func (v *visitor) VisitGroup(g *proto.Group)             {}
func (v *visitor) VisitExtensions(e *proto.Extensions)   {}
