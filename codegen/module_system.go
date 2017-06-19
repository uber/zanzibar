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
)

// EndpointMeta saves meta data used to render an endpoint.
type EndpointMeta struct {
	GatewayPackageName string
	PackageName        string
	IncludedPackages   []GoPackageImport
	Method             *MethodSpec
	ClientName         string
	ClientID           string
	ClientMethodName   string
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
	ClientID         string
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
) (*ModuleSystem, error) {
	system := NewModuleSystem()
	tmpl, err := NewTemplate()

	if err != nil {
		return nil, err
	}

	// Register client module class and type generators
	if err := system.RegisterClass("client", ModuleClass{
		Directory:         "clients",
		ClassType:         MultiModule,
		ClassDependencies: []string{"client"},
	}); err != nil {
		return nil, errors.Wrapf(err, "Error registering client class")
	}

	if err := system.RegisterClassType("client", "http", &HTTPClientGenerator{
		templates:     tmpl,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering http client class type",
		)
	}

	if err := system.RegisterClassType("client", "tchannel", &TChannelClientGenerator{
		templates:     tmpl,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering TChannel client class type",
		)
	}

	if err := system.RegisterClassType("client", "custom", &CustomClientGenerator{
		templates:     tmpl,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering custom client class type",
		)
	}

	if err := system.RegisterClass("clients", ModuleClass{
		Directory:         "clients",
		ClassType:         SingleModule,
		ClassDependencies: []string{},
	}); err != nil {
		return nil, errors.Wrapf(err, "Error registering clientInit class")
	}
	if err := system.RegisterClassType("clients", "init", &ClientsInitGenerator{
		templates:     tmpl,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering clientInit class type",
		)
	}

	// Register endpoint module class and type generators
	if err := system.RegisterClass("endpoint", ModuleClass{
		Directory:         "endpoints",
		ClassType:         MultiModule,
		ClassDependencies: []string{"client"},
	}); err != nil {
		return nil, errors.Wrapf(err, "Error registering endpoint class")
	}

	if err := system.RegisterClassType("endpoint", "http", &EndpointGenerator{
		templates:     tmpl,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering HTTP endpoint class type",
		)
	}

	if err := system.RegisterClassType("endpoint", "tchannel", &EndpointGenerator{
		templates:     tmpl,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering HTTP endpoint class type",
		)
	}

	if err := system.RegisterClass("service", ModuleClass{
		Directory:         "services",
		ClassType:         MultiModule,
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
}

// Generate returns the HTTP client build result, which contains the files and
// the generated client spec
func (g *HTTPClientGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
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
		instance,
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

	clientMeta := &ClientMeta{
		ExportName:       clientSpec.ExportName,
		ExportType:       clientSpec.ExportType,
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

	// When it is possible to generate structs for all module types, the
	// module system will do this transparently. For now we are opting in
	// on a per-generator basis.
	dependencies, err := GenerateDependencyStruct(
		instance,
		g.packageHelper,
		g.templates,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating dependencies struct for %q %q",
			instance.ClassName,
			instance.InstanceName,
		)
	}

	baseName := filepath.Base(instance.Directory)
	clientFilePath := baseName + ".go"

	files := map[string][]byte{
		clientFilePath: client,
	}

	if dependencies != nil {
		files["dependencies.go"] = dependencies
	}

	return &BuildResult{
		Files: files,
		Spec:  clientSpec,
	}, nil
}

/*
 * TChannel Client Generator
 */

// TChannelClientGenerator generates an instance of a zanzibar TChannel client
type TChannelClientGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
}

// Generate returns the TChannel client build result, which contains the files
// and the generated client spec
func (g *TChannelClientGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
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
		instance,
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

	// reverse index the exposed methods map
	exposedMethods := map[string]string{}
	for k, v := range clientSpec.ExposedMethods {
		exposedMethods[v] = k
	}

	// TODO: Verify all exposedMethods exist and are valid.
	for exposedMethod := range exposedMethods {
		segments := strings.Split(exposedMethod, "::")
		thriftService := segments[0]
		thriftMethod := segments[1]
		found := false

		for _, candidateService := range clientSpec.ModuleSpec.Services {
			if found {
				break
			}
			if candidateService.Name != thriftService {
				continue
			}

			for _, candidateMethod := range candidateService.Methods {
				if candidateMethod.Name != thriftMethod {
					continue
				}

				found = true
				break
			}
		}

		if !found {
			return nil, errors.Errorf(
				"Invalid exposedMethods for tchannel client (%q). "+
					"The exposedMethod (%q) does not exist",
				instance.InstanceName,
				exposedMethod,
			)
		}
	}

	clientMeta := &ClientMeta{
		ExportName:       clientSpec.ExportName,
		ExportType:       clientSpec.ExportType,
		PackageName:      clientSpec.ModuleSpec.PackageName,
		Services:         clientSpec.ModuleSpec.Services,
		IncludedPackages: clientSpec.ModuleSpec.IncludedPackages,
		ClientID:         clientSpec.ClientID,
		ExposedMethods:   exposedMethods,
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

	// When it is possible to generate structs for all module types, the
	// module system will do this transparently. For now we are opting in
	// on a per-generator basis.
	dependencies, err := GenerateDependencyStruct(
		instance,
		g.packageHelper,
		g.templates,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating dependencies struct for %q %q",
			instance.ClassName,
			instance.InstanceName,
		)
	}

	baseName := filepath.Base(instance.Directory)
	clientFilePath := baseName + ".go"
	serverFilePath := baseName + "_test_server.go"

	files := map[string][]byte{
		clientFilePath: client,
		serverFilePath: server,
	}

	if dependencies != nil {
		files["dependencies.go"] = dependencies
	}

	return &BuildResult{
		Files: files,
		Spec:  clientSpec,
	}, nil
}

/*
 * Custom Client Generator
 */

// CustomClientGenerator gathers the custom client spec for future use in ClientsInitGenerator
type CustomClientGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
}

// Generate returns the custom client build result, which contains the
// generated client spec and no files
func (g *CustomClientGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	// Parse the client config from the endpoint JSON file
	clientConfig, err := readClientConfig(instance.JSONFileRaw)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading custom client %q JSON config",
			instance.InstanceName,
		)
	}

	clientSpec, err := NewCustomClientSpec(
		instance,
		clientConfig,
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error initializing CustomClientSpec for %q",
			instance.InstanceName,
		)
	}

	// When it is possible to generate structs for all module types, the
	// module system will do this transparently. For now we are opting in
	// on a per-generator basis.
	dependencies, err := GenerateDependencyStruct(
		instance,
		g.packageHelper,
		g.templates,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating dependencies struct for %q %q",
			instance.ClassName,
			instance.InstanceName,
		)
	}

	files := map[string][]byte{}

	if dependencies != nil {
		files["dependencies.go"] = dependencies
	}

	return &BuildResult{
		Files: files,
		Spec:  clientSpec,
	}, nil
}

/*
 * Clients Init Generator
 */

// ClientsInitGenerator generates a clients initialization file
type ClientsInitGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
}

// Generate returns the client initializer build result, which contains the
// generated clients initializer file and no spec
func (g *ClientsInitGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	clients := readClientDependencySpecs(instance)

	includedPkgs := []GoPackageImport{}
	for _, client := range clients {
		// TODO: there shouldn't be a special thing for custom
		if client.ClientType == "custom" {
			includedPkgs = append(includedPkgs, GoPackageImport{
				PackageName: client.CustomImportPath,
				AliasName:   client.ImportPackageAlias,
			})
			continue
		}

		if len(client.ModuleSpec.Services) != 0 {
			includedPkgs = append(includedPkgs, GoPackageImport{
				PackageName: client.ImportPackagePath,
				AliasName:   client.ImportPackageAlias,
			})
		}
	}

	clientInfo := []ClientInfoMeta{}
	for _, client := range clients {
		// TODO: there shouldn't be a special thing for custom
		if client.ClientType == "custom" {
			isPointerType := false
			exportType := client.CustomClientType
			if strings.HasPrefix(exportType, "*") {
				isPointerType = true
				exportType = strings.TrimLeft(exportType, "*")
			}

			clientInfo = append(clientInfo, ClientInfoMeta{
				IsPointerType: isPointerType,
				FieldName:     strings.Title(client.ClientName),
				PackagePath:   client.ImportPackagePath,
				PackageAlias:  client.ImportPackageAlias,
				ExportName:    client.ExportName,
				ExportType:    exportType,
			})
			continue
		}

		module := client.ModuleSpec
		if len(module.Services) == 0 {
			continue
		}

		clientInfo = append(clientInfo, ClientInfoMeta{
			IsPointerType: true,
			FieldName:     strings.Title(client.ClientName),
			PackagePath:   client.ImportPackagePath,
			PackageAlias:  client.ImportPackageAlias,
			ExportName:    client.ExportName,
			ExportType:    client.ExportType,
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

	return &BuildResult{
		Files: map[string][]byte{
			"clients.go": clientsInit,
		},
	}, nil
}

/*
 * Endpoint Generator
 */

// EndpointGenerator generates a group of zanzibar http endpoints that proxy corresponding clients
type EndpointGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
}

// Generate returns the endpoint build result, which contains a file per
// endpoint handler and a list of handler specs
func (g *EndpointGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	ret := map[string][]byte{}
	endpointJsons := []string{}
	endpointSpecs := []*EndpointSpec{}
	clientSpecs := readClientDependencySpecs(instance)

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

		endpointSpecs = append(endpointSpecs, espec)

		err = espec.SetDownstream(clientSpecs, g.packageHelper)
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
	return &BuildResult{
		Files: ret,
		Spec:  endpointSpecs,
	}, nil
}

func (g *EndpointGenerator) generateEndpointFile(
	e *EndpointSpec, instance *ModuleInstance, out map[string][]byte,
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

	clientID := e.ClientID
	clientName := ""
	if e.ClientSpec != nil {
		clientName = e.ClientSpec.ClientName
	}

	// TODO: http client needs to support multiple thrift services
	meta := &EndpointMeta{
		GatewayPackageName: g.packageHelper.GoGatewayPackageName(),
		PackageName:        m.PackageName,
		IncludedPackages:   includedPackages,
		Method:             method,
		ReqHeaderMap:       e.ReqHeaderMap,
		ReqHeaderMapKeys:   e.ReqHeaderMapKeys,
		ResHeaderMap:       e.ResHeaderMap,
		ResHeaderMapKeys:   e.ResHeaderMapKeys,
		ClientID:           clientID,
		ClientName:         clientName,
		ClientMethodName:   e.ClientMethod,
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

func (g *EndpointGenerator) generateEndpointTestFile(
	e *EndpointSpec, instance *ModuleInstance, out map[string][]byte,
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
		ClientID:    e.ClientSpec.ClientID,
	}

	tempName := "endpoint_test.tmpl"
	if e.WorkflowType == "tchannelClient" {
		meta.ClientName = e.ClientSpec.ClientName

		meta.IncludedPackages = append(
			method.Downstream.IncludedPackages,
			GoPackageImport{
				AliasName:   method.Downstream.PackageName,
				PackageName: e.ClientSpec.ImportPackagePath,
			},
		)
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

/*
 * Gateway Service Generator
 */

// GatewayServiceGenerator generates an entry point for a single service as
// a main.go that bootstraps the service and its dependencies
type GatewayServiceGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
}

// Generate returns the gateway build result, which contains the service and
// service test main files, and no spec
func (generator *GatewayServiceGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
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

	return &BuildResult{
		Files: map[string][]byte{
			"zanzibar-defaults.json": productionConfig,
			"main.go":                main,
			"main_test.go":           mainTest,
		},
	}, nil
}

func readClientDependencySpecs(instance *ModuleInstance) []*ClientSpec {
	clients := []*ClientSpec{}

	for _, clientDep := range instance.ResolvedDependencies["client"] {
		clients = append(clients, clientDep.GeneratedSpec().(*ClientSpec))
	}

	sort.Sort(sortByClientName(clients))

	return clients
}

// GenerateDependencyStruct generates a module struct with placeholders for the
// instance module based on the defined dependency configuration
func GenerateDependencyStruct(
	instance *ModuleInstance,
	packageHelper *PackageHelper,
	template *Template,
) ([]byte, error) {
	hasDependencies := false

	for _, moduleInstances := range instance.ResolvedDependencies {
		if len(moduleInstances) > 0 {
			hasDependencies = true
			break
		}
	}

	if !hasDependencies {
		return nil, nil
	}

	return template.execTemplate(
		"dependency_struct.tmpl",
		instance,
		packageHelper,
	)
}
