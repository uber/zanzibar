// Copyright (c) 2018 Uber Technologies, Inc.
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
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// EndpointMeta saves meta data used to render an endpoint.
type EndpointMeta struct {
	Instance           *ModuleInstance
	Spec               *EndpointSpec
	GatewayPackageName string
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

// EndpointCollectionMeta saves information used to generate an initializer
// for a collection of endpoints
type EndpointCollectionMeta struct {
	Instance     *ModuleInstance
	EndpointMeta []*EndpointMeta
}

// StructMeta saves information to generate an endpoint's thrift structs file
type StructMeta struct {
	Instance *ModuleInstance
	Spec     *ModuleSpec
}

// EndpointTestMeta saves meta data used to render an endpoint test.
type EndpointTestMeta struct {
	Instance           *ModuleInstance
	Method             *MethodSpec
	TestFixtures       map[string]*EndpointTestFixture `json:"testFixtures"`
	ClientName         string
	ClientID           string
	RelativePathToRoot string
	IncludedPackages   []GoPackageImport
}

// FixtureBlob is map[string]interface{} that implements default string
// used for headers and (http | tchannel) request/response
type FixtureBlob map[string]interface{}

// String convert FixtureBlob to string
func (fb *FixtureBlob) String() string {
	str, err := json.Marshal(fb)
	if err != nil {
		panic(err)
	}
	return string(str)
}

// BodyType can either be `json` or `string`
type BodyType string

// ProtocalType can either be `http` or `tchannel`
type ProtocalType string

// HTTPMethodType can only be valid http method
type HTTPMethodType string

// FixtureBody is used to create http body in test fixtures
type FixtureBody struct {
	BodyType   BodyType     `json:"bodyType,omitempty"`
	BodyString string       `json:"bodyString,omitempty"` // set BodyString if response body is string
	BodyJSON   *FixtureBlob `json:"bodyJson,omitempty"`   // set BodyJSON if response body is object
}

// String convert FixtureBody to string
// This String() panics inside if type and value mismatch during unmarshal
// because template cannot handle errors
func (fb *FixtureBody) String() string {
	switch fb.BodyType {
	case "string":
		return fb.BodyString
	case "json":
		if fb.BodyJSON == nil {
			panic(errors.New("invalid http body type"))
		}
		return fb.BodyJSON.String()
	default:
		panic(errors.New("invalid http body type"))
	}
}

// FixtureHTTPResponse is test fixture for http response
type FixtureHTTPResponse struct {
	StatusCode int          `json:"statusCode"`
	Body       *FixtureBody `json:"body,omitempty"`
}

// FixtureResponse is test fixture for client/endpoint response
type FixtureResponse struct {
	ResponseType     ProtocalType         `json:"responseType"`
	HTTPResponse     *FixtureHTTPResponse `json:"httpResponse,omitempty"`
	TChannelResponse FixtureBlob          `json:"tchannelResponse,omitempty"`
}

// Body returns the string representation of FixtureResponse
func (fr *FixtureResponse) Body() string {
	switch fr.ResponseType {
	case "tchannel":
		return fr.TChannelResponse.String()
	case "http":
		res := ""
		if fr.HTTPResponse.Body != nil {
			res = fr.HTTPResponse.Body.String()
		}
		return res
	default:
		panic(errors.New("invalid response type"))
	}
}

// FixtureHTTPRequest is test fixture for client/endpoint request
type FixtureHTTPRequest struct {
	Method HTTPMethodType `json:"method,omitempty"`
	Body   *FixtureBody   `json:"body,omitempty"`
}

// FixtureRequest is test fixture for client/endpoint request
type FixtureRequest struct {
	RequestType     ProtocalType        `json:"requestType"`
	HTTPRequest     *FixtureHTTPRequest `json:"httpRequest,omitempty"`
	TChannelRequest FixtureBlob         `json:"tchannelRequest,omitempty"`
}

// Body returns the string representation of FixtureRequest
func (fr *FixtureRequest) Body() string {
	switch fr.RequestType {
	case "tchannel":
		return fr.TChannelRequest.String()
	case "http":
		res := ""
		if fr.HTTPRequest.Body != nil {
			res = fr.HTTPRequest.Body.String()
		}
		return res
	default:
		panic(errors.New("invalid request type"))
	}
}

// EndpointTestFixture saves mocked requests/responses for an endpoint test.
type EndpointTestFixture struct {
	TestName           string                        `json:"testName"`
	EndpointID         string                        `json:"endpointId"`
	HandleID           string                        `json:"handleId"`
	EndpointRequest    FixtureRequest                `json:"endpointRequest"` // there's no difference between http or tchannel request
	EndpointReqHeaders FixtureBlob                   `json:"endpointReqHeaders"`
	EndpointResponse   FixtureResponse               `json:"endpointResponse"`
	EndpointResHeaders FixtureBlob                   `json:"endpointResHeaders"`
	ClientTestFixtures map[string]*ClientTestFixture `json:"clientTestFixtures"`
	TestServiceName    string                        `json:"testServiceName"` // The service module that mounts the endpoint
}

// ClientTestFixture saves mocked client request/response for an endpoint test.
type ClientTestFixture struct {
	ClientID         string          `json:"clientId"`
	ClientMethod     string          `json:"clientMethod"`
	ClientRequest    FixtureRequest  `json:"clientRequest"` // there's no difference between http or tchannel request
	ClientReqHeaders FixtureBlob     `json:"clientReqHeaders"`
	ClientResponse   FixtureResponse `json:"clientResponse"`
	ClientResHeaders FixtureBlob     `json:"clientResHeaders"`
}

// NewDefaultModuleSystemWithMockHook creates a fresh instance of the default zanzibar
// module system (clients, endpoints, services) with a post build hook to generate client and service mocks
func NewDefaultModuleSystemWithMockHook(
	h *PackageHelper,
	hooks ...PostGenHook,
) (*ModuleSystem, error) {
	t, err := NewDefaultTemplate()
	if err != nil {
		return nil, err
	}

	clientMockGenHook, err := ClientMockGenHook(h, t)
	if err != nil {
		return nil, errors.Wrap(err, "error creating client mock gen hook")
	}

	serviceMockGenHook := ServiceMockGenHook(h, t)

	allHooks := append([]PostGenHook{clientMockGenHook, serviceMockGenHook}, hooks...)
	return NewDefaultModuleSystem(h, allHooks...)
}

// NewDefaultModuleSystem creates a fresh instance of the default zanzibar
// module system (clients, endpoints, services)
func NewDefaultModuleSystem(
	h *PackageHelper,
	hooks ...PostGenHook,
) (*ModuleSystem, error) {
	system := NewModuleSystem(hooks...)

	tmpl, err := NewDefaultTemplate()
	if err != nil {
		return nil, err
	}

	// Register client module class and type generators
	if err := system.RegisterClass(ModuleClass{
		Name:        "client",
		Directories: []string{"clients"},
		ClassType:   MultiModule,
	}); err != nil {
		return nil, errors.Wrapf(err, "Error registering client class")
	}

	if err := system.RegisterClassType("client", "http", &HTTPClientGenerator{
		templates:     tmpl,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering HTTP client class type",
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

	if err := system.RegisterClass(ModuleClass{
		Name:        "middleware",
		Directories: []string{"middlewares"},
		ClassType:   MultiModule,
		DependsOn:   []string{"client"},
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering middleware class",
		)
	}

	if err := system.RegisterClassType("middleware", "default", &MiddlewareGenerator{
		templates:     tmpl,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering Gateway middleware class type",
		)
	}

	// Register endpoint module class and type generators
	if err := system.RegisterClass(ModuleClass{
		Name:        "endpoint",
		Directories: []string{"endpoints"},
		ClassType:   MultiModule,
		DependsOn:   []string{"client", "middleware"},
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
			"Error registering TChannel endpoint class type",
		)
	}

	if err := system.RegisterClass(ModuleClass{
		Name:        "service",
		Directories: []string{"services"},
		ClassType:   MultiModule,
		DependsOn:   []string{"endpoint"},
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering service class",
		)
	}

	if err := system.RegisterClassType("service", "gateway",
		NewGatewayServiceGenerator(tmpl, h)); err != nil {
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

	exposedMethods, err := reverseExposedMethods(clientSpec, instance)
	if err != nil {
		return nil, err
	}

	clientMeta := &ClientMeta{
		Instance:         instance,
		ExportName:       clientSpec.ExportName,
		ExportType:       clientSpec.ExportType,
		Services:         clientSpec.ModuleSpec.Services,
		IncludedPackages: clientSpec.ModuleSpec.IncludedPackages,
		ClientID:         clientSpec.ClientID,
		ExposedMethods:   exposedMethods,
		// TODO: http client integration with sidecar router
		SidecarRouter: clientSpec.SidecarRouter,
	}

	client, err := g.templates.ExecTemplate(
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
		files["module/dependencies.go"] = dependencies
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

	exposedMethods, err := reverseExposedMethods(clientSpec, instance)
	if err != nil {
		return nil, err
	}

	clientMeta := &ClientMeta{
		Instance:         instance,
		ExportName:       clientSpec.ExportName,
		ExportType:       clientSpec.ExportType,
		Services:         clientSpec.ModuleSpec.Services,
		IncludedPackages: clientSpec.ModuleSpec.IncludedPackages,
		ClientID:         clientSpec.ClientID,
		ExposedMethods:   exposedMethods,
		SidecarRouter:    clientSpec.SidecarRouter,
	}

	client, err := g.templates.ExecTemplate(
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

	server, err := g.templates.ExecTemplate(
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
		files["module/dependencies.go"] = dependencies
	}

	return &BuildResult{
		Files: files,
		Spec:  clientSpec,
	}, nil
}

// reverse index and validate the exposed methods map
func reverseExposedMethods(clientSpec *ClientSpec, instance *ModuleInstance) (map[string]string, error) {
	reversed := map[string]string{}
	for exposedMethod, thriftMethod := range clientSpec.ExposedMethods {
		reversed[thriftMethod] = exposedMethod
		if !hasMethod(clientSpec, thriftMethod) {
			return nil, errors.Errorf(
				"Invalid exposedMethods for client %q, method %q not found",
				instance.InstanceName,
				thriftMethod,
			)
		}
	}

	return reversed, nil
}

func hasMethod(cspec *ClientSpec, thriftMethod string) bool {
	segments := strings.Split(thriftMethod, "::")
	service := segments[0]
	method := segments[1]

	for _, s := range cspec.ModuleSpec.Services {
		if s.Name == service {
			for _, m := range s.Methods {
				if m.Name == method {
					return true
				}
			}
		}

	}
	return false
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
		files["module/dependencies.go"] = dependencies
	}

	return &BuildResult{
		Files: files,
		Spec:  clientSpec,
	}, nil
}

/*
 * Endpoint Generator
 */

// EndpointGenerator generates a group of zanzibar endpoints that proxy corresponding clients
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
	endpointMeta := []*EndpointMeta{}
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

		meta, err := g.generateEndpointFile(espec, instance, ret)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"Error executing endpoint template %q",
				instance.InstanceName,
			)
		}
		endpointMeta = append(endpointMeta, meta)

		err = g.generateEndpointTestFile(espec, instance, ret)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"Error executing endpoint test template %q",
				instance.InstanceName,
			)
		}
	}

	dependencies, err := GenerateDependencyStruct(
		instance,
		g.packageHelper,
		g.templates,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating service dependencies for %s",
			instance.InstanceName,
		)
	}
	if dependencies != nil {
		ret["module/dependencies.go"] = dependencies
	}

	endpointCollection, err := g.templates.ExecTemplate(
		"endpoint_collection.tmpl",
		&EndpointCollectionMeta{
			Instance:     instance,
			EndpointMeta: endpointMeta,
		},
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating service dependencies for %s",
			instance.InstanceName,
		)
	}
	ret["endpoint.go"] = endpointCollection

	return &BuildResult{
		Files: ret,
		Spec:  endpointSpecs,
	}, nil
}

func (g *EndpointGenerator) generateEndpointFile(
	e *EndpointSpec, instance *ModuleInstance, out map[string][]byte,
) (*EndpointMeta, error) {
	m := e.ModuleSpec
	methodName := e.ThriftMethodName
	thriftServiceName := e.ThriftServiceName

	if len(m.Services) == 0 {
		return nil, nil
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
			meta := &StructMeta{
				Instance: instance,
				Spec:     m,
			}
			structs, err := g.templates.ExecTemplate(
				"structs.tmpl",
				meta,
				g.packageHelper,
			)
			if err != nil {
				return nil, err
			}
			out[structFilePath] = structs
		}
	}

	method := findMethod(m, thriftServiceName, methodName)
	if method == nil {
		return nil, errors.Errorf(
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
		workflowName = strings.Title(thriftServiceName) + strings.Title(method.Name) + "Endpoint"
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
		Instance:           instance,
		Spec:               e,
		GatewayPackageName: g.packageHelper.GoGatewayPackageName(),
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
		endpoint, err = g.templates.ExecTemplate("endpoint.tmpl", meta, g.packageHelper)
	} else if e.EndpointType == "tchannel" {
		endpoint, err = g.templates.ExecTemplate("tchannel_endpoint.tmpl", meta, g.packageHelper)
	} else {
		err = errors.Errorf("Endpoint type '%s' is not supported", e.EndpointType)
	}

	if err != nil {
		return nil, errors.Wrap(err, "Error executing endpoint template")
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

	return meta, nil
}

func (g *EndpointGenerator) generateEndpointTestFile(
	e *EndpointSpec, instance *ModuleInstance, out map[string][]byte,
) error {
	if len(e.TestFixtures) < 1 { // skip tests if testFixtures is missing
		return nil
	}
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

	endpointDirectory := filepath.Join(
		g.packageHelper.CodeGenTargetPath(),
		instance.Directory,
	)
	targetPath := e.TargetEndpointTestPath(serviceName, methodName)
	endpointTestFilePath, err := filepath.Rel(endpointDirectory, targetPath)
	if err != nil {
		endpointTestFilePath = targetPath
	}

	meta := &EndpointTestMeta{
		Instance:     instance,
		Method:       method,
		TestFixtures: e.TestFixtures,
		ClientID:     e.ClientSpec.ClientID,
	}

	relativePath, err := filepath.Rel(
		targetPath, g.packageHelper.CodeGenTargetPath(),
	)
	if err != nil {
		return errors.Wrap(err,
			"Error computing relative path for endpoint test template",
		)
	}

	meta.RelativePathToRoot = relativePath

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

	endpointTest, err := g.templates.ExecTemplate(tempName, meta, g.packageHelper)
	if err != nil {
		return errors.Wrap(err, "Error executing endpoint test template")
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

// NewGatewayServiceGenerator creates a new gateway service generator.
func NewGatewayServiceGenerator(t *Template, h *PackageHelper) *GatewayServiceGenerator {
	return &GatewayServiceGenerator{
		templates:     t,
		packageHelper: h,
	}
}

// Generate returns the gateway build result, which contains the service and
// service test main files, and no spec
func (generator *GatewayServiceGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	service, err := generator.templates.ExecTemplate(
		"service.tmpl",
		instance,
		generator.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating service service.go for %s",
			instance.InstanceName,
		)
	}

	// generate main.go
	main, err := generator.templates.ExecTemplate(
		"main.tmpl",
		instance,
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
	mainTest, err := generator.templates.ExecTemplate(
		"main_test.tmpl",
		instance,
		generator.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating service main_test.go for %s",
			instance.InstanceName,
		)
	}

	dependencies, err := GenerateDependencyStruct(
		instance,
		generator.packageHelper,
		generator.templates,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating service dependencies for %s",
			instance.InstanceName,
		)
	}

	initializer, err := GenerateInitializer(
		instance,
		generator.packageHelper,
		generator.templates,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating service initializer for %s",
			instance.InstanceName,
		)
	}

	files := map[string][]byte{
		"service.go":        service,
		"main/main.go":      main,
		"main/main_test.go": mainTest,
		"module/init.go":    initializer,
	}

	if dependencies != nil {
		files["module/dependencies.go"] = dependencies
	}

	return &BuildResult{
		Files: files,
	}, nil
}

/*
 * Middleware Generator
 */

// MiddlewareGenerator generates a group of zanzibar endpoints that proxy corresponding clients
type MiddlewareGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
}

// Generate returns the endpoint build result, which contains a file per
// endpoint handler and a list of handler specs
func (g *MiddlewareGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	ret := map[string][]byte{}

	dependencies, err := GenerateDependencyStruct(
		instance,
		g.packageHelper,
		g.templates,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating service dependencies for %s %s",
			instance.InstanceName,
			instance.ClassName,
		)
	}
	if dependencies != nil {
		ret["module/dependencies.go"] = dependencies
	}

	err = g.generateMiddlewareFile(instance, ret)

	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating middleware file for %s %s",
			instance.InstanceName,
			instance.ClassName,
		)
	}
	return &BuildResult{
		Files: ret,
		Spec:  nil,
	}, nil
}

func (g *MiddlewareGenerator) generateMiddlewareFile(instance *ModuleInstance, out map[string][]byte) error {
	bytes, err := g.templates.ExecTemplate("middleware.tmpl", instance, g.packageHelper)
	if err != nil {
		return err
	}
	out[instance.InstanceName+".go"] = bytes
	return nil
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
	return template.ExecTemplate(
		"dependency_struct.tmpl",
		instance,
		packageHelper,
	)
}

// GenerateInitializer generates a file that initializes a module fully
// recursively, i.e. by initializing all of its dependencies in the correct
// order
func GenerateInitializer(
	instance *ModuleInstance,
	packageHelper *PackageHelper,
	template *Template,
) ([]byte, error) {
	return template.ExecTemplate(
		"module_initializer.tmpl",
		instance,
		packageHelper,
	)
}

/*
 * General client meta
 */

// ClientMeta ...
type ClientMeta struct {
	Instance         *ModuleInstance
	ExportName       string
	ExportType       string
	ClientID         string
	IncludedPackages []GoPackageImport
	Services         []*ServiceSpec
	ExposedMethods   map[string]string
	SidecarRouter    string
	Fixture          *Fixture
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
