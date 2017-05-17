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
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/module"
)

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
	PackageName      string
	Method           *MethodSpec
	TestStubs        []TestStub
	ClientName       string
	IncludedPackages []GoPackageImport
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

	TestServiceName string // The service module that mounts the endpoint
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

// NewDefaultModuleSystem creates a fresh instance of the default zanzibar
// module system (clients, endpoints)
func NewDefaultModuleSystem(
	h *PackageHelper,
) (*module.System, error) {
	system := module.NewSystem()
	tmpl, err := NewTemplate()

	if err != nil {
		return nil, err
	}

	// Register client module class and type generators
	if err := system.RegisterClass("client", module.Class{
		Directory:         "clients",
		ClassType:         module.MultiModule,
		ClassDependencies: []string{},
	}); err != nil {
		return nil, errors.Wrapf(err, "Error registering client class")
	}

	clientSpecs := map[string]*ClientSpec{}

	if err := system.RegisterClassType("client", "http", &HTTPClientGenerator{
		templates:     tmpl,
		genSpecs:      clientSpecs,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering http client class type",
		)
	}

	if err := system.RegisterClassType("client", "tchannel", &TChannelClientGenerator{
		templates:     tmpl,
		genSpecs:      clientSpecs,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering TChannel client class type",
		)
	}

	if err := system.RegisterClass("init", module.Class{
		Directory:         "clients",
		ClassType:         module.SingleModule,
		ClassDependencies: []string{},
	}); err != nil {
		return nil, errors.Wrapf(err, "Error registering clientInit class")
	}
	if err := system.RegisterClassType("init", "clients", &ClientsInitGenerator{
		templates:     tmpl,
		clientSpecs:   clientSpecs,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering clientInit class type",
		)
	}

	if err := system.RegisterClass("service", module.Class{
		Directory:         "services",
		ClassType:         module.MultiModule,
		ClassDependencies: []string{"client"},
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering service class",
		)
	}

	if err := system.RegisterClassType("service", "gateway", &GatewayServiceGenerator{
		templates:     tmpl,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering Gateway service class type",
		)
	}

	// Register endpoint module class and type generators
	if err := system.RegisterClass("endpoint", module.Class{
		Directory:         "endpoints",
		ClassType:         module.MultiModule,
		ClassDependencies: []string{"client"},
	}); err != nil {
		return nil, errors.Wrapf(err, "Error registering endpoint class")
	}

	if err := system.RegisterClassType("endpoint", "http", &EndpointGenerator{
		templates:     tmpl,
		clientSpecs:   clientSpecs,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering HTTP endpoint class type",
		)
	}

	if err := system.RegisterClassType("endpoint", "tchannel", &EndpointGenerator{
		templates:     tmpl,
		clientSpecs:   clientSpecs,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering HTTP endpoint class type",
		)
	}
	return system, nil
}

func readClientConfig(rawConfig []byte) (*ClientClassConfig, error) {
	var clientConfig ClientClassConfig
	if err := json.Unmarshal(rawConfig, &clientConfig); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading config for client instance",
		)
	}
	clientConfig.Config["clientId"] = clientConfig.Name
	clientConfig.Config["clientType"] = clientConfig.Type
	return &clientConfig, nil
}

func readEndpointConfig(rawConfig []byte) (*EndpointClassConfig, error) {
	var endpointConfig EndpointClassConfig
	if err := json.Unmarshal(rawConfig, &endpointConfig); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading config for endpoint instance",
		)
	}
	return &endpointConfig, nil
}

/*
 * HTTP Client Generator
 */

// HTTPClientGenerator generates an instance of a zanzibar http client
type HTTPClientGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
	genSpecs      map[string]*ClientSpec
}

// Generate returns the HTTP client generated files as a map of relative file
// path (relative to the target build directory) to file bytes.
func (g *HTTPClientGenerator) Generate(
	instance *module.Instance,
) (map[string][]byte, error) {
	// Parse the client config from the endpoint JSON file
	clientConfig, err := readClientConfig(instance.JSONFileRaw)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading HTTP client %q JSON config",
			instance.InstanceName,
		)
	}

	clientSpec, err := NewHTTPClientSpec(
		filepath.Join(
			instance.BaseDirectory,
			instance.Directory,
			instance.JSONFileName,
		),
		clientConfig,
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error initializing HTTPClientSpec for %q",
			instance.InstanceName,
		)
	}

	g.genSpecs[clientSpec.JSONFile] = clientSpec

	clientMeta := &ClientMeta{
		PackageName:      clientSpec.ModuleSpec.PackageName,
		Services:         clientSpec.ModuleSpec.Services,
		IncludedPackages: clientSpec.ModuleSpec.IncludedPackages,
		ClientID:         clientSpec.ClientID,
	}

	client, err := g.templates.execTemplate(
		"http_client.tmpl",
		clientMeta,
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error executing HTTP client template for %q",
			instance.InstanceName,
		)
	}

	clientDirectory := filepath.Join(
		g.packageHelper.CodeGenTargetPath(),
		instance.Directory,
	)

	clientFilePath, err := filepath.Rel(clientDirectory, clientSpec.GoFileName)
	if err != nil {
		clientFilePath = clientSpec.GoFileName
	}

	// Return the client files
	return map[string][]byte{
		clientFilePath: client,
	}, nil
}

/*
 * TChannel Client Generator
 */

// TChannelClientGenerator generates an instance of a zanzibar TChannel client
type TChannelClientGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
	genSpecs      map[string]*ClientSpec
}

// Generate returns the TChannel client generated files as a map of relative file
// path (relative to the target build directory) to file bytes.
func (g *TChannelClientGenerator) Generate(
	instance *module.Instance,
) (map[string][]byte, error) {
	// Parse the client config from the endpoint JSON file
	clientConfig, err := readClientConfig(instance.JSONFileRaw)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading TChannel client %q JSON config",
			instance.InstanceName,
		)
	}

	clientSpec, err := NewTChannelClientSpec(
		filepath.Join(
			instance.BaseDirectory,
			instance.Directory,
			instance.JSONFileName,
		),
		clientConfig,
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error initializing TChannelClientSpec for %q",
			instance.InstanceName,
		)
	}

	g.genSpecs[clientSpec.JSONFile] = clientSpec

	clientMeta := &ClientMeta{
		PackageName:      clientSpec.ModuleSpec.PackageName,
		Services:         clientSpec.ModuleSpec.Services,
		IncludedPackages: clientSpec.ModuleSpec.IncludedPackages,
		ClientID:         clientSpec.ClientID,
	}

	client, err := g.templates.execTemplate(
		"tchannel_client.tmpl",
		clientMeta,
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error executing TChannel client template for %q",
			instance.InstanceName,
		)
	}

	server, err := g.templates.execTemplate(
		"tchannel_client_test_server.tmpl",
		clientMeta,
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error executing TChannel server template for %q",
			instance.InstanceName,
		)
	}

	clientDirectory := filepath.Join(
		g.packageHelper.CodeGenTargetPath(),
		instance.Directory,
	)

	clientFilePath, err := filepath.Rel(clientDirectory, clientSpec.GoFileName)
	if err != nil {
		clientFilePath = clientSpec.GoFileName
	}

	serverFilePath := strings.TrimRight(clientFilePath, ".go") + "_test_server.go"

	// Return the client files
	return map[string][]byte{
		clientFilePath: client,
		serverFilePath: server,
	}, nil
}

/*
 * Clients Init Generator
 */

// ClientsInitGenerator generates a clients initialization file
type ClientsInitGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
	clientSpecs   map[string]*ClientSpec
}

// Generate returns the client init file as a map of relative file
// path (relative to the target build directory) to file bytes.
func (g *ClientsInitGenerator) Generate(
	instance *module.Instance,
) (map[string][]byte, error) {
	clients := []*ClientSpec{}
	for _, v := range g.clientSpecs {
		clients = append(clients, v)
	}
	sort.Sort(sortByClientName(clients))

	includedPkgs := []GoPackageImport{}
	for i := 0; i < len(clients); i++ {
		if clients[i].ClientType == "custom" {
			includedPkgs = append(includedPkgs, GoPackageImport{
				PackageName: clients[i].CustomImportPath,
				AliasName:   clients[i].CustomPackageName,
			})
			continue
		}

		if len(clients[i].ModuleSpec.Services) != 0 {
			includedPkgs = append(includedPkgs, GoPackageImport{
				PackageName: clients[i].GoPackageName,
				AliasName:   "",
			},
			)
		}
	}

	clientInfo := []ClientInfoMeta{}
	for i := 0; i < len(clients); i++ {
		if clients[i].ClientType == "custom" {
			clientInfo = append(clientInfo, ClientInfoMeta{
				FieldName:   strings.Title(clients[i].ClientName),
				PackageName: clients[i].CustomPackageName,
				TypeName:    clients[i].CustomClientType,
			})
			continue
		}

		module := clients[i].ModuleSpec
		if len(module.Services) == 0 {
			continue
		}

		if len(module.Services) != 1 {
			return nil, errors.Errorf(
				"Cannot import client with multiple services: %s",
				module.PackageName,
			)
		}

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

	clientsInit, err := g.templates.execTemplate(
		"init_clients.tmpl",
		meta,
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "Error executing client init template")
	}

	return map[string][]byte{
		"clients.go": clientsInit,
	}, nil
}

/*
 * Endpoint Generator
 */

// EndpointGenerator generates a group of zanzibar http endpoints that proxy corresponding clients
type EndpointGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
	clientSpecs   map[string]*ClientSpec
}

// Generate returns the endpoint generated files as a map of relative file
// path (relative to the target build directory) to file bytes.
func (g *EndpointGenerator) Generate(
	instance *module.Instance,
) (map[string][]byte, error) {
	ret := map[string][]byte{}
	endpointJsons := []string{}

	endpointConfig, err := readEndpointConfig(instance.JSONFileRaw)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading HTTP endpoint %q JSON config",
			instance.InstanceName,
		)
	}

	endpointConfigDir := filepath.Join(
		instance.BaseDirectory,
		instance.Directory,
	)
	for _, fileName := range endpointConfig.Config.Endpoints {
		endpointJsons = append(
			endpointJsons, filepath.Join(endpointConfigDir, fileName),
		)
	}
	for _, jsonFile := range endpointJsons {
		espec, err := NewEndpointSpec(jsonFile, g.packageHelper, g.packageHelper.MiddlewareSpecs())
		if err != nil {
			return nil, errors.Wrapf(
				err, "Error parsing endpoint json file: %s", jsonFile,
			)
		}

		err = espec.SetDownstream(g.clientSpecs, g.packageHelper)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Error parsing downstream info for endpoint: %s", jsonFile,
			)
		}

		err = g.generateEndpointFile(espec, instance, ret)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"Error executing endpoint template %q",
				instance.InstanceName,
			)
		}

		err = g.generateEndpointTestFile(espec, instance, ret)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"Error executing endpoint test template %q",
				instance.InstanceName,
			)
		}
	}
	return ret, nil
}

func (g *EndpointGenerator) generateEndpointFile(
	e *EndpointSpec, instance *module.Instance, out map[string][]byte,
) error {
	m := e.ModuleSpec
	methodName := e.ThriftMethodName
	thriftServiceName := e.ThriftServiceName

	if len(m.Services) == 0 {
		return nil
	}

	endpointDirectory := filepath.Join(
		g.packageHelper.CodeGenTargetPath(),
		instance.Directory,
	)

	var err error
	if e.EndpointType == "http" {
		structFilePath, err := filepath.Rel(endpointDirectory, e.GoStructsFileName)
		if err != nil {
			structFilePath = e.GoStructsFileName
		}
		if _, ok := out[structFilePath]; !ok {
			structs, err := g.templates.execTemplate("structs.tmpl", m, g.packageHelper)
			if err != nil {
				return err
			}
			out[structFilePath] = structs
		}
	}

	method := findMethod(m, thriftServiceName, methodName)
	if method == nil {
		return errors.Errorf(
			"Could not find thriftServiceName %q + methodName %q in module",
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

	meta := &EndpointMeta{
		GatewayPackageName: g.packageHelper.GoGatewayPackageName(),
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

	var endpoint []byte
	if e.EndpointType == "http" {
		endpoint, err = g.templates.execTemplate("endpoint.tmpl", meta, g.packageHelper)
	} else if e.EndpointType == "tchannel" {
		endpoint, err = g.templates.execTemplate("tchannel_endpoint.tmpl", meta, g.packageHelper)
	} else {
		err = errors.Errorf("Endpoint type '%s' is not supported", e.EndpointType)
	}

	if err != nil {
		return errors.Wrap(err, "Error executing endpoint template")
	}

	targetPath := e.TargetEndpointPath(thriftServiceName, method.Name)
	if e.EndpointType == "tchannel" {
		targetPath = strings.TrimRight(targetPath, ".go") + "_tchannel.go"
	}
	endpointFilePath, err := filepath.Rel(endpointDirectory, targetPath)
	if err != nil {
		endpointFilePath = targetPath
	}

	out[endpointFilePath] = endpoint

	return nil
}

/*
 * Gateway Service Generator
 */

// GatewayServiceGenerator generates an entry point for a single service as
// a main.go that bootstraps the service and its dependencies
type GatewayServiceGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
}

// Generate returns the gateway service generated files as a map of relative
// file path (relative to the target buid directory) to file bytes.
func (generator *GatewayServiceGenerator) Generate(
	instance *module.Instance,
) (map[string][]byte, error) {
	// zanzibar-defaults.json is copied from ../config/production.json
	configSrcFileName := path.Join(
		getDirName(), "..", "config", "production.json",
	)
	productionConfig, err := ioutil.ReadFile(configSrcFileName)
	if err != nil {
		return nil, errors.Wrap(
			err,
			"Could not read config/production.json while generating main file",
		)
	}

	// main.go and main_test.go shared meta
	meta := &MainMeta{
		IncludedPackages: []GoPackageImport{
			{
				PackageName: generator.packageHelper.GoGatewayPackageName() +
					"/clients",
				AliasName: "",
			},
			{
				PackageName: generator.packageHelper.GoGatewayPackageName() +
					"/endpoints",
				AliasName: "",
			},
		},
		GatewayName:             instance.InstanceName,
		RelativePathToAppConfig: filepath.Join("..", "..", ".."),
	}

	// generate main.go
	main, err := generator.templates.execTemplate(
		"main.tmpl",
		meta,
		generator.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating service main.go for %s",
			instance.InstanceName,
		)
	}

	// generate main_test.go
	mainTest, err := generator.templates.execTemplate(
		"main_test.tmpl",
		meta,
		generator.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating service main_test.go for %s",
			instance.InstanceName,
		)
	}

	return map[string][]byte{
		"zanzibar-defaults.json": productionConfig,
		"main.go":                main,
		"main_test.go":           mainTest,
	}, nil
}

func (g *EndpointGenerator) generateEndpointTestFile(
	e *EndpointSpec, instance *module.Instance, out map[string][]byte,
) error {
	m := e.ModuleSpec
	methodName := e.ThriftMethodName
	serviceName := e.ThriftServiceName

	if len(m.Services) == 0 {
		return nil
	}

	method := findMethod(m, serviceName, methodName)
	if method == nil {
		return errors.Errorf(
			"Could not find thriftServiceName %q + methodName %q in module",
			serviceName, methodName,
		)
	}

	// Read test configurations
	testConfigPath := e.EndpointTestConfigPath()

	var testStubs []TestStub
	file, err := ioutil.ReadFile(testConfigPath)
	if err != nil {
		// If the test file does not exist then skip test generation.
		if os.IsNotExist(err) {
			return nil
		}

		return errors.Wrapf(err,
			"Error reading endpoint test config for service %q, method %q",
			serviceName, method.Name)
	}
	err = json.Unmarshal(file, &testStubs)
	if err != nil {
		return errors.Wrapf(err,
			"Error parsing test config file.")
	}

	for i := 0; i < len(testStubs); i++ {
		testStub := &testStubs[i]
		testStub.EndpointRequestString, err = jsonMarshal(
			testStub.EndpointRequest)
		if err != nil {
			return errors.Wrapf(err,
				"Error parsing JSON in test config.")
		}
		testStub.EndpointResponseString, err = jsonMarshal(
			testStub.EndpointResponse)
		if err != nil {
			return errors.Wrapf(err,
				"Error parsing JSON in test config.")
		}
		for j := 0; j < len(testStub.ClientStubs); j++ {
			clientStub := &testStub.ClientStubs[j]
			clientStub.ClientRequestString, err = jsonMarshal(
				clientStub.ClientRequest)
			if err != nil {
				return errors.Wrapf(err,
					"Error parsing JSON in test config.")
			}
			clientStub.ClientResponseString, err = jsonMarshal(
				clientStub.ClientResponse)
			if err != nil {
				return errors.Wrapf(err,
					"Error parsing JSON in test config.")
			}
			// Build canonical key list to keep templates in order
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
		// Build canonical key list to keep templates in order
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

	tempName := "endpoint_test.tmpl"
	if e.WorkflowType == "tchannelClient" {
		meta.ClientName = e.ClientName

		genCodeClientPkgName, err := g.packageHelper.TypeImportPath(method.Downstream.ThriftFile)
		if err != nil {
			return errors.Wrap(err, "could not run endpoint_test template")
		}
		genCodeClientAliasName, err := g.packageHelper.TypePackageName(method.Downstream.ThriftFile)
		if err != nil {
			return errors.Wrap(err, "could not run endpoint_test template")
		}
		meta.IncludedPackages = []GoPackageImport{
			{
				AliasName:   method.Downstream.PackageName,
				PackageName: method.Downstream.GoPackage,
			},
			{
				AliasName:   genCodeClientAliasName,
				PackageName: genCodeClientPkgName,
			},
		}
		tempName = "endpoint_test_tchannel_client.tmpl"
	}

	endpointTest, err := g.templates.execTemplate(tempName, meta, g.packageHelper)
	if err != nil {
		return errors.Wrap(err, "Error executing endpoint test template")
	}
	endpointDirectory := filepath.Join(
		g.packageHelper.CodeGenTargetPath(),
		instance.Directory,
	)
	targetPath := e.TargetEndpointTestPath(serviceName, methodName)
	endpointTestFilePath, err := filepath.Rel(endpointDirectory, targetPath)
	if err != nil {
		endpointTestFilePath = targetPath
	}

	out[endpointTestFilePath] = endpointTest

	return nil
}
