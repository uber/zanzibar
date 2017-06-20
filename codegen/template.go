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

package codegen

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	tmpl "text/template"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/codegen/template_bundle"
)

// MainFiles are group of files generated for main entry point.
type MainFiles struct {
	MainFile     string
	MainTestFile string
}

var funcMap = tmpl.FuncMap{
	"lower":         strings.ToLower,
	"title":         strings.Title,
	"Title":         strings.Title,
	"fullTypeName":  fullTypeName,
	"camel":         camelCase,
	"split":         strings.Split,
	"dec":           decrement,
	"basePath":      filepath.Base,
	"pascal":        pascalCase,
	"jsonMarshal":   jsonMarshal,
	"isPointerType": isPointerType,
	"unref":         unref,
}

func fullTypeName(typeName, packageName string) string {
	if typeName == "" || strings.Contains(typeName, ".") {
		return typeName
	}
	return packageName + "." + typeName
}

func decrement(num int) int {
	return num - 1
}

func jsonMarshal(jsonObj map[string]interface{}) (string, error) {
	str, err := json.Marshal(jsonObj)
	return string(str), err
}

func isPointerType(t string) bool {
	return strings.HasPrefix(t, "*")
}

func unref(t string) string {
	if strings.HasPrefix(t, "*") {
		return strings.TrimLeft(t, "*")
	}
	return t
}

// Template generates code for edge gateway clients and edgegateway endpoints.
type Template struct {
	template *tmpl.Template
}

// NewTemplate creates a bundle of templates.
func NewTemplate() (*Template, error) {
	t := tmpl.New("main").Funcs(funcMap)
	for _, file := range templates.AssetNames() {
		fileContent, err := templates.Asset(file)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read bin data for template %s", file)
		}
		if _, err := t.New(file).Parse(string(fileContent)); err != nil {
			return nil, errors.Wrapf(err, "Could not parse template %s", file)
		}
	}
	return &Template{
		template: t,
	}, nil
}

// ClientMeta ...
type ClientMeta struct {
	PackageName      string
	ExportName       string
	ExportType       string
	ClientID         string
	IncludedPackages []GoPackageImport
	Services         []*ServiceSpec
	ExposedMethods   map[string]string
}

func findMethod(
	m *ModuleSpec, serviceName string, methodName string,
) *MethodSpec {
	for _, service := range m.Services {
		if service.Name != serviceName {
			continue
		}

		for _, method := range service.Methods {
			if method.Name == methodName {
				return method
			}
		}
	}
	return nil
}

// ClientInfoMeta ...
type ClientInfoMeta struct {
	IsPointerType   bool
	FieldName       string
	PackagePath     string
	PackageAlias    string
	ExportName      string
	ExportType      string
	DepPackageAlias string
	DepFieldNames   []string
}

// ClientsInitFilesMeta ...
type ClientsInitFilesMeta struct {
	IncludedPackages []GoPackageImport
	ClientInfo       []ClientInfoMeta
}

type sortByClientName []*ClientSpec

func (c sortByClientName) Len() int {
	return len(c)
}
func (c sortByClientName) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c sortByClientName) Less(i, j int) bool {
	return c[i].ClientName < c[j].ClientName
}

// EndpointRegisterInfo ...
type EndpointRegisterInfo struct {
	EndpointType string
	Constructor  string
	Method       *MethodSpec
	EndpointID   string
	HandlerID    string
	PackageName  string
	HandlerType  string
	HandlerName  string
	Middlewares  []MiddlewareSpec
}

// EndpointsRegisterMeta ...
type EndpointsRegisterMeta struct {
	IncludedPackages []GoPackageImport
	Endpoints        []EndpointRegisterInfo
}

type sortByEndpointName []*EndpointSpec

func (c sortByEndpointName) Len() int {
	return len(c)
}

func (c sortByEndpointName) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c sortByEndpointName) Less(i, j int) bool {
	if c[i].EndpointType != c[j].EndpointType {
		return c[i].EndpointType < c[j].EndpointType
	}
	return (c[i].EndpointID + c[i].HandleID) <
		(c[j].EndpointID + c[j].HandleID)
}

func contains(arr []GoPackageImport, value string) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i].PackageName == value {
			return true
		}
	}
	return false
}

// GenerateEndpointRegisterFile will generate the registration file
// that mounts all the endpoints in the router.
func (t *Template) GenerateEndpointRegisterFile(
	endpointsMap map[string]*EndpointSpec, h *PackageHelper,
) (string, error) {
	endpoints := make([]*EndpointSpec, 0, len(endpointsMap))
	for _, v := range endpointsMap {
		endpoints = append(endpoints, v)
	}
	sort.Sort(sortByEndpointName(endpoints))

	includedPkgs := []GoPackageImport{}
	endpointsInfo := make([]EndpointRegisterInfo, 0, len(endpoints))

	for i := 0; i < len(endpoints); i++ {
		espec := endpoints[i]

		method := findMethod(
			espec.ModuleSpec,
			espec.ThriftServiceName,
			espec.ThriftMethodName,
		)
		if method == nil {
			return "", errors.Errorf(
				"Could not find serviceName %q + methodName %q in module",
				espec.ThriftServiceName, espec.ThriftMethodName,
			)
		}

		// TODO: unify constructor naming
		var endpointType, handlerType, constructor string
		switch espec.EndpointType {
		case "http":
			endpointType = "HTTP"
			handlerType = "*" + espec.ModuleSpec.PackageName + "." +
				strings.Title(method.Name) + "Handler"
			constructor = "New" + strings.Title(method.Name) + "Endpoint"
		case "tchannel":
			endpointType = "TChannel"
			handlerType = "zanzibar.TChannelHandler"
			constructor = "New" + strings.Title(method.ThriftService) +
				strings.Title(method.Name) + "Handler"
		default:
			panic("Unsupported endpoint type: " + espec.EndpointType)
		}
		handlerName := strings.Title(espec.EndpointID) + strings.Title(method.Name) +
			endpointType + "Handler"

		aliasName := aliasImport(espec, h)
		includedPkgs = addEndpointPackage(espec, aliasName, includedPkgs)
		includedPkgs = addMiddlewarePackages(espec.Middlewares, includedPkgs)

		info := EndpointRegisterInfo{
			EndpointType: endpointType,
			Constructor:  constructor,
			EndpointID:   espec.EndpointID,
			HandlerID:    espec.HandleID,
			Method:       method,
			PackageName:  aliasName,
			HandlerType:  handlerType,
			HandlerName:  handlerName,
			Middlewares:  espec.Middlewares,
		}
		endpointsInfo = append(endpointsInfo, info)
	}

	meta := &EndpointsRegisterMeta{
		IncludedPackages: includedPkgs,
		Endpoints:        endpointsInfo,
	}

	targetFile := h.TargetEndpointsRegisterPath()
	err := t.execTemplateAndFmt(
		"endpoint_register.tmpl",
		targetFile,
		meta,
		h)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

func aliasImport(spec *EndpointSpec, h *PackageHelper) string {
	// error is ok to ignore as the goPkg is generated based on packageRoot
	relEndpointDir, _ := filepath.Rel(h.PackageRoot(), spec.GoPackageName)
	relGenDir, _ := filepath.Rel(h.configRoot, h.CodeGenTargetPath())
	relEndpointDir = strings.TrimPrefix(relEndpointDir, filepath.Clean(relGenDir)+"/endpoints/")
	parts := strings.Split(relEndpointDir, "/")
	aliasName := parts[0]
	for i, v := range parts {
		if i != 0 {
			aliasName = aliasName + strings.Title(v)
		}
	}
	return aliasName
}

func addEndpointPackage(espec *EndpointSpec, aliasName string, includedPkgs []GoPackageImport) []GoPackageImport {
	var goPkg string
	switch espec.WorkflowType {
	case "httpClient", "tchannelClient", "custom":
		goPkg = espec.GoPackageName
	default:
		panic("Unsupported WorkflowType: " + espec.WorkflowType)
	}

	if aliasName == espec.ModuleSpec.PackageName {
		aliasName = ""
	}
	if !contains(includedPkgs, goPkg) {
		includedPkgs = append(includedPkgs, GoPackageImport{
			PackageName: goPkg,
			AliasName:   aliasName,
		})
	}
	return includedPkgs
}

func addMiddlewarePackages(middlewares []MiddlewareSpec, includedPkgs []GoPackageImport) []GoPackageImport {
	for _, m := range middlewares {
		if !contains(includedPkgs, m.Path) {
			includedPkgs = append(includedPkgs, GoPackageImport{
				PackageName: m.Path,
				AliasName:   "",
			})
		}
	}
	return includedPkgs
}

// MainMeta ...
type MainMeta struct {
	IncludedPackages        []GoPackageImport
	GatewayName             string
	RelativePathToAppConfig string
}

func (t *Template) execTemplateAndFmt(
	templName string,
	filePath string,
	data interface{},
	pkgHelper *PackageHelper,
) error {

	file, err := openFileOrCreate(filePath)
	if err != nil {
		return errors.Wrapf(err, "failed to open file: %s", err)
	}

	_, err = io.WriteString(file, "// Code generated by zanzibar \n"+
		"// @generated \n \n")
	if err != nil {
		return errors.Wrapf(err, "failed to write to file: %s", err)
	}

	_, err = io.WriteString(file, pkgHelper.copyrightHeader)
	if err != nil {
		return errors.Wrapf(err, "failed to write to file: %s", err)
	}
	_, err = io.WriteString(file, "\n\n")
	if err != nil {
		return errors.Wrapf(err, "failed to write to file: %s", err)
	}

	if err := t.template.ExecuteTemplate(file, templName, data); err != nil {
		return errors.Wrapf(err, "failed to execute template files for file %s", file)
	}

	gofmtCmd := exec.Command("gofmt", "-s", "-w", "-e", filePath)
	gofmtCmd.Stdout = os.Stdout
	gofmtCmd.Stderr = os.Stderr

	if err := gofmtCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to gofmt file: %s", filePath)
	}

	goimportsCmd := exec.Command("goimports", "-w", "-e", filePath)
	goimportsCmd.Stdout = os.Stdout
	goimportsCmd.Stderr = os.Stderr

	if err := goimportsCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to goimports file: %s", filePath)
	}

	if err := file.Close(); err != nil {
		return errors.Wrap(err, "failed to close file")
	}

	return nil
}

func (t *Template) execTemplate(
	tplName string,
	tplData interface{},
	pkgHelper *PackageHelper,
) ([]byte, error) {
	tplBuffer := bytes.NewBuffer(nil)

	_, err := io.WriteString(tplBuffer, "// Code generated by zanzibar \n"+
		"// @generated \n \n")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to write to file: %s", err)
	}

	_, err = io.WriteString(tplBuffer, pkgHelper.copyrightHeader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to write to file: %s", err)
	}
	_, err = io.WriteString(tplBuffer, "\n\n")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to write to file: %s", err)
	}

	if err := t.template.ExecuteTemplate(
		tplBuffer,
		tplName,
		tplData,
	); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating template %s",
			tplName,
		)
	}

	return tplBuffer.Bytes(), nil
}

func openFileOrCreate(file string) (*os.File, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
			return nil, errors.Wrapf(
				err, "could not make directory: %s", file,
			)
		}
	}
	return os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
}
