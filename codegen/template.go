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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	tmpl "text/template"

	"github.com/pkg/errors"
)

const zanzibarPath = "github.com/uber/zanzibar"

// EndpointFiles are group of files generated for an endpoint.
type EndpointFiles struct {
	HandlerFiles []string
	StructFile   string
}

// ClientFiles are group of files generated for a client.
type ClientFiles struct {
	ClientFile string
	StructFile string
}

// MainFiles are group of files generated for main entry point.
type MainFiles struct {
	MainFile     string
	MainTestFile string
}

// EndpointMeta saves meta data used to render an endpoint.
type EndpointMeta struct {
	GatewayPackageName string
	PackageName        string
	IncludedPackages   []GoPackageImport
	Method             *MethodSpec
	ClientName         string
	WorkflowName       string
	ReqHeaderMap       map[string]string
	ReqHeaderMapKeys   []string
	ResHeaderMap       map[string]string
	ResHeaderMapKeys   []string
}

// EndpointTestMeta saves meta data used to render an endpoint test.
type EndpointTestMeta struct {
	PackageName string
	Method      *MethodSpec
	TestStubs   []TestStub
}

// TestStub saves stubbed requests/responses for an endpoint test.
type TestStub struct {
	TestName               string
	EndpointID             string
	HandlerID              string
	EndpointRequest        map[string]interface{} // Json blob
	EndpointRequestString  string
	EndpointReqHeaders     map[string]string      // Json blob
	EndpointReqHeaderKeys  []string               // To keep in canonical order
	EndpointResponse       map[string]interface{} // Json blob
	EndpointResponseString string
	EndpointResHeaders     map[string]string // Json blob
	EndpointResHeaderKeys  []string          // To keep in canonical order

	ClientStubs []ClientStub
}

// ClientStub saves stubbed client request/response for an endpoint test.
type ClientStub struct {
	ClientID             string
	ClientMethod         string
	ClientRequest        map[string]interface{} // Json blob
	ClientRequestString  string
	ClientReqHeaders     map[string]string      // Json blob
	ClientReqHeaderKeys  []string               // To keep in canonical order
	ClientResponse       map[string]interface{} // Json blob
	ClientResponseString string
	ClientResHeaders     map[string]string // Json blob
	ClientResHeaderKeys  []string          // To keep in canonical order
}

var camelingRegex = regexp.MustCompile("[0-9A-Za-z]+")
var funcMap = tmpl.FuncMap{
	"title":        strings.Title,
	"Title":        strings.Title,
	"fullTypeName": fullTypeName,
	"camel":        camelCase,
	"split":        strings.Split,
	"dec":          decrement,
	"basePath":     filepath.Base,
	"pascal":       pascalCase,
	"jsonMarshal":  jsonMarshal,
}

func fullTypeName(typeName, packageName string) string {
	if typeName == "" || strings.Contains(typeName, ".") {
		return typeName
	}
	return packageName + "." + typeName
}

func camelCase(src string) string {
	byteSrc := []byte(src)
	chunks := camelingRegex.FindAll(byteSrc, -1)
	for idx, val := range chunks {
		if idx > 0 {
			chunks[idx] = bytes.Title(val)
		} else {
			chunks[idx][0] = bytes.ToLower(val[0:1])[0]
		}
	}
	return string(bytes.Join(chunks, nil))
}

func decrement(num int) int {
	return num - 1
}

func jsonMarshal(jsonObj map[string]interface{}) (string, error) {
	str, err := json.Marshal(jsonObj)
	return string(str), err
}

func pascalCase(src string) string {
	byteSrc := []byte(src)
	chunks := camelingRegex.FindAll(byteSrc, -1)
	for idx, val := range chunks {
		chunks[idx] = bytes.Title(val)
	}
	return string(bytes.Join(chunks, nil))
}

// Template generates code for edge gateway clients and edgegateway endpoints.
type Template struct {
	template *tmpl.Template
}

// NewTemplate creates a bundle of templates.
func NewTemplate(templatePattern string) (*Template, error) {
	t, err := tmpl.New("main").Funcs(funcMap).ParseGlob(templatePattern)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse template files")
	}
	return &Template{
		template: t,
	}, nil
}

// ClientMeta ...
type ClientMeta struct {
	PackageName      string
	ClientID         string
	IncludedPackages []GoPackageImport
	Services         []*ServiceSpec
}

// GenerateClientFile generates Go http code for services defined in thrift file.
// It returns the path of generated client file and struct file or an error.
func (t *Template) GenerateClientFile(
	c *ClientSpec, h *PackageHelper,
) (*ClientFiles, error) {
	m := c.ModuleSpec

	if len(m.Services) == 0 {
		return nil, nil
	}

	clientMeta := &ClientMeta{
		PackageName:      m.PackageName,
		Services:         m.Services,
		IncludedPackages: m.IncludedPackages,
		ClientID:         c.ClientID,
	}
	err := t.execTemplateAndFmt(
		"http_client.tmpl",
		c.GoFileName,
		clientMeta,
		h.copyrightHeader,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not run http_client template")
	}

	err = t.execTemplateAndFmt(
		"structs.tmpl",
		c.GoStructsFileName,
		clientMeta,
		h.copyrightHeader)
	if err != nil {
		return nil, errors.Wrap(err, "could not run structs template")
	}

	return &ClientFiles{
		ClientFile: c.GoFileName,
		StructFile: c.GoStructsFileName,
	}, nil
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

// GenerateEndpointFile generates Go code for an zanzibar endpoint defined in
// thrift file. It returns the path of generated method files, struct file or
// an error.
func (t *Template) GenerateEndpointFile(
	e *EndpointSpec, h *PackageHelper, thriftServiceName string, methodName string,
) (*EndpointFiles, error) {
	m := e.ModuleSpec

	if len(m.Services) == 0 {
		return nil, nil
	}

	err := t.execTemplateAndFmt(
		"structs.tmpl", e.GoStructsFileName, m, h.copyrightHeader,
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not run structs template")
	}

	endpointFiles := &EndpointFiles{
		HandlerFiles: make([]string, 0, len(m.Services[0].Methods)),
		StructFile:   e.GoStructsFileName,
	}
	method := findMethod(m, thriftServiceName, methodName)
	if method == nil {
		return nil, errors.Errorf(
			"Could not find thriftServiceName (%s) + methodName (%s) in module",
			thriftServiceName, methodName,
		)
	}

	includedPackages := m.IncludedPackages
	if e.WorkflowImportPath != "" {
		includedPackages = append(includedPackages, GoPackageImport{
			PackageName: e.WorkflowImportPath,
			AliasName:   "custom" + strings.Title(m.PackageName),
		})
	}

	var workflowName string
	if method.Downstream != nil {
		workflowName = strings.Title(method.Name) + "Endpoint"
	} else {
		workflowName = "custom" + strings.Title(m.PackageName) + "." +
			strings.Title(method.Name) + "Endpoint"
	}

	dest := e.TargetEndpointPath(thriftServiceName, method.Name)
	meta := &EndpointMeta{
		GatewayPackageName: h.GoGatewayPackageName(),
		PackageName:        m.PackageName,
		IncludedPackages:   includedPackages,
		Method:             method,
		ReqHeaderMap:       e.ReqHeaderMap,
		ReqHeaderMapKeys:   e.ReqHeaderMapKeys,
		ResHeaderMap:       e.ResHeaderMap,
		ResHeaderMapKeys:   e.ResHeaderMapKeys,
		ClientName:         e.ClientName,
		WorkflowName:       workflowName,
	}

	err = t.execTemplateAndFmt("endpoint.tmpl", dest, meta, h.copyrightHeader)
	if err != nil {
		return nil, errors.Wrap(err, "could not run endpoint template")
	}
	endpointFiles.HandlerFiles = append(endpointFiles.HandlerFiles, dest)

	return endpointFiles, nil
}

// GenerateEndpointTestFile generates Go code for testing an zanzibar endpoint
// defined in a thrift file. It returns the path of generated test files,
// or an error.
func (t *Template) GenerateEndpointTestFile(
	e *EndpointSpec, h *PackageHelper, serviceName string, methodName string,
) ([]string, error) {
	m := e.ModuleSpec

	if len(m.Services) == 0 {
		return nil, nil
	}

	testFiles := make([]string, 0, len(m.Services[0].Methods))

	method := findMethod(m, serviceName, methodName)
	if method == nil {
		return nil, errors.Errorf(
			"Could not find serviceName (%s) + methodName (%s) in module",
			serviceName, methodName,
		)
	}

	dest := e.TargetEndpointTestPath(serviceName, methodName)

	// Read test configurations
	testConfigPath := e.EndpointTestConfigPath()

	var testStubs []TestStub
	file, err := ioutil.ReadFile(testConfigPath)
	if err != nil {
		// If the test file does not exist then skip test generation.
		if os.IsNotExist(err) {
			return []string{}, nil
		}

		return nil, errors.Wrapf(err,
			"Could not read endpoint test config for service %s, method %s",
			serviceName, method.Name)
	}
	err = json.Unmarshal(file, &testStubs)
	if err != nil {
		return nil, errors.Wrapf(err,
			"Error parsing test config file.")
	}

	for i := 0; i < len(testStubs); i++ {
		testStub := &testStubs[i]
		testStub.EndpointRequestString, err = jsonMarshal(
			testStub.EndpointRequest)
		if err != nil {
			return nil, errors.Wrapf(err,
				"Error parsing JSON in test config.")
		}
		testStub.EndpointResponseString, err = jsonMarshal(
			testStub.EndpointResponse)
		if err != nil {
			return nil, errors.Wrapf(err,
				"Error parsing JSON in test config.")
		}
		for j := 0; j < len(testStub.ClientStubs); j++ {
			clientStub := &testStub.ClientStubs[j]
			clientStub.ClientRequestString, err = jsonMarshal(
				clientStub.ClientRequest)
			if err != nil {
				return nil, errors.Wrapf(err,
					"Error parsing JSON in test config.")
			}
			clientStub.ClientResponseString, err = jsonMarshal(
				clientStub.ClientResponse)
			if err != nil {
				return nil, errors.Wrapf(err,
					"Error parsing JSON in test config.")
			}
			// Build canonicalized key list to keep templates in order
			// when comparing to golden files.
			clientStub.ClientReqHeaderKeys = make(
				[]string,
				len(clientStub.ClientReqHeaders))
			i := 0
			for k := range clientStub.ClientReqHeaders {
				clientStub.ClientReqHeaderKeys[i] = k
				i++
			}
			sort.Strings(clientStub.ClientReqHeaderKeys)
			clientStub.ClientResHeaderKeys = make(
				[]string,
				len(clientStub.ClientResHeaders))
			i = 0
			for k := range clientStub.ClientResHeaders {
				clientStub.ClientResHeaderKeys[i] = k
				i++
			}
			sort.Strings(clientStub.ClientResHeaderKeys)

		}
		// Build canonicalized key list to keep templates in order
		// when comparing to golden files.
		testStub.EndpointReqHeaderKeys = make(
			[]string,
			len(testStub.EndpointReqHeaders))
		i := 0
		for k := range testStub.EndpointReqHeaders {
			testStub.EndpointReqHeaderKeys[i] = k
			i++
		}
		sort.Strings(testStub.EndpointReqHeaderKeys)
		testStub.EndpointResHeaderKeys = make(
			[]string,
			len(testStub.EndpointResHeaders))
		i = 0
		for k := range testStub.EndpointResHeaders {
			testStub.EndpointResHeaderKeys[i] = k
			i++
		}
		sort.Strings(testStub.EndpointResHeaderKeys)
	}

	meta := &EndpointTestMeta{
		PackageName: m.PackageName,
		Method:      method,
		TestStubs:   testStubs,
	}

	err = t.execTemplateAndFmt(
		"endpoint_test.tmpl",
		dest,
		meta,
		h.copyrightHeader)
	if err != nil {
		return nil, errors.Wrap(err, "could not run endpoint_test template")
	}
	testFiles = append(testFiles, dest)

	return testFiles, nil
}

// ClientInfoMeta ...
type ClientInfoMeta struct {
	FieldName   string
	PackageName string
	TypeName    string
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
	return c[i].GoFileName < c[j].GoFileName
}

// GenerateClientsInitFile generates go code to allocate and initialize
// all of the generated clients
func (t *Template) GenerateClientsInitFile(
	clientsMap map[string]*ClientSpec, h *PackageHelper,
) (string, error) {
	clients := []*ClientSpec{}
	for _, v := range clientsMap {
		clients = append(clients, v)
	}
	sort.Sort(sortByClientName(clients))

	includedPkgs := []GoPackageImport{}
	for i := 0; i < len(clients); i++ {
		if len(clients[i].ModuleSpec.Services) == 0 {
			continue
		}

		pkgName := clients[i].GoPackageName
		if clients[i].ClientType == "custom" {
			pkgName = clients[i].CustomImportPath
		}

		includedPkgs = append(
			includedPkgs, GoPackageImport{
				PackageName: pkgName,
				AliasName:   "",
			},
		)
	}

	clientInfo := []ClientInfoMeta{}
	for i := 0; i < len(clients); i++ {
		module := clients[i].ModuleSpec
		if len(module.Services) == 0 {
			continue
		}

		if len(module.Services) != 1 {
			return "", errors.Errorf(
				"Cannot import client with multiple services: %s",
				module.PackageName,
			)
		}

		//service := module.Services[0]
		clientInfo = append(clientInfo, ClientInfoMeta{
			FieldName:   strings.Title(clients[i].ClientName),
			PackageName: module.PackageName,
			TypeName:    strings.Title(clients[i].ClientName) + "Client",
		})
	}

	meta := &ClientsInitFilesMeta{
		IncludedPackages: includedPkgs,
		ClientInfo:       clientInfo,
	}

	targetFile := h.TargetClientsInitPath()
	err := t.execTemplateAndFmt(
		"init_clients.tmpl",
		targetFile,
		meta,
		h.copyrightHeader)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

// EndpointRegisterInfo ...
type EndpointRegisterInfo struct {
	Method      string
	HTTPPath    string
	EndpointID  string
	HandlerID   string
	PackageName string
	HandlerType string
	MethodName  string
	HandlerName string
	Middlewares []MiddlewareSpec
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

	includedPkgs := []GoPackageImport{
		{
			PackageName: h.GoGatewayPackageName() + "/clients",
			AliasName:   "",
		},
	}
	endpointsInfo := make([]EndpointRegisterInfo, 0, len(endpoints))

	for i := 0; i < len(endpoints); i++ {
		espec := endpoints[i]

		var goPkg string
		if espec.WorkflowType == "httpClient" || espec.WorkflowType == "tchannelClient" {
			goPkg = espec.ModuleSpec.GoPackage
		} else if espec.WorkflowType == "custom" {
			goPkg = espec.ModuleSpec.GoPackage
		} else {
			panic("Unsupported WorkflowType: " + espec.WorkflowType)
		}

		if !contains(includedPkgs, goPkg) {
			includedPkgs = append(includedPkgs, GoPackageImport{
				PackageName: goPkg,
				AliasName:   "",
			})
		}

		method := findMethod(
			espec.ModuleSpec,
			espec.ThriftServiceName,
			espec.ThriftMethodName,
		)
		if method == nil {
			return "", errors.Errorf(
				"Could not find serviceName (%s) + methodName (%s) in module",
				espec.ThriftServiceName, espec.ThriftMethodName,
			)
		}

		handlerType := espec.ModuleSpec.PackageName +
			"." + strings.Title(method.Name) + "Handler"

		info := EndpointRegisterInfo{
			EndpointID:  espec.EndpointID,
			HandlerID:   espec.HandleID,
			Method:      method.HTTPMethod,
			HTTPPath:    method.HTTPPath,
			PackageName: espec.ModuleSpec.PackageName,
			HandlerType: handlerType,
			MethodName:  strings.Title(method.Name),
			HandlerName: strings.Title(method.Name) + "Handler",
			Middlewares: espec.Middlewares,
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
		h.copyrightHeader)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

// MainMeta ...
type MainMeta struct {
	IncludedPackages        []GoPackageImport
	GatewayName             string
	RelativePathToAppConfig string
}

// GenerateMainFile will use main.tmpl to write out the main.go file
// for a gateway.
func (t *Template) GenerateMainFile(
	g *GatewaySpec, h *PackageHelper,
) (*MainFiles, error) {
	rootConfigDirName := g.configDirName
	configDestFileName := h.TargetProductionConfigFilePath()
	gatewayDirName := filepath.Dir(configDestFileName)
	deltaPath, err := filepath.Rel(gatewayDirName, rootConfigDirName)
	if err != nil {
		return nil, errors.Wrap(
			err, "Could not build relative path when generating main file",
		)
	}

	configSrcFileName := path.Join(
		getDirName(), "..", "config", "production.json",
	)
	bytes, err := ioutil.ReadFile(configSrcFileName)
	if err != nil {
		return nil, errors.Wrap(
			err,
			"Could not read config/production.json while generating main file",
		)
	}

	err = ioutil.WriteFile(configDestFileName, bytes, 0644)
	if err != nil {
		return nil, errors.Wrap(
			err,
			"Could not write config file while generating main file",
		)
	}

	meta := &MainMeta{
		IncludedPackages: []GoPackageImport{
			{
				PackageName: h.GoGatewayPackageName() + "/clients",
				AliasName:   "",
			},
			{
				PackageName: h.GoGatewayPackageName() + "/endpoints",
				AliasName:   "",
			},
		},
		GatewayName:             g.gatewayName,
		RelativePathToAppConfig: deltaPath,
	}

	mainFile := h.TargetMainPath()
	err = t.execTemplateAndFmt("main.tmpl", mainFile, meta, h.copyrightHeader)
	if err != nil {
		return nil, err
	}

	mainTestFile := h.TargetMainTestPath()
	err = t.execTemplateAndFmt(
		"main_test.tmpl",
		mainTestFile,
		meta,
		h.copyrightHeader)
	if err != nil {
		return nil, err
	}

	return &MainFiles{
		MainFile:     mainFile,
		MainTestFile: mainTestFile,
	}, nil
}

func (t *Template) execTemplateAndFmt(
	templName string,
	filePath string,
	data interface{},
	copyrightHeader string,
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

	_, err = io.WriteString(file, copyrightHeader)
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
) ([]byte, error) {
	tplBuffer := bytes.NewBuffer(nil)

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
