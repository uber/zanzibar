// Copyright (c) 2020 Uber Technologies, Inc.
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
	"net/textproto"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	validator2 "gopkg.in/validator.v2"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/uber/zanzibar/parallelize"
	"go.uber.org/thriftrw/compile"
)

// EndpointMeta saves meta data used to render an endpoint.
type EndpointMeta struct {
	Instance               *ModuleInstance
	Spec                   *EndpointSpec
	GatewayPackageName     string
	IncludedPackages       []GoPackageImport
	Method                 *MethodSpec
	ClientName             string
	ClientID               string
	ClientMethodName       string
	WorkflowPkg            string
	ReqHeaders             map[string]*TypedHeader
	IsClientlessEndpoint   bool
	ReqHeadersKeys         []string
	ReqRequiredHeadersKeys []string
	ResHeaders             map[string]*TypedHeader
	ResHeadersKeys         []string
	ResRequiredHeadersKeys []string
	TraceKey               string
	DeputyReqHeader        string
	DefaultHeaders         []string
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
	TestFixtures       map[string]*EndpointTestFixture `yaml:"testFixtures" json:"testFixtures"`
	ReqHeaders         map[string]*TypedHeader
	ResHeaders         map[string]*TypedHeader
	ClientName         string
	ClientID           string
	RelativePathToRoot string
	IncludedPackages   []GoPackageImport
}

// FixtureBlob implements default string used for (http | tchannel)
// request/response
type FixtureBlob map[string]interface{}

func toStringMap(i map[string]interface{}) map[string]interface{} {
	m := make(map[string]interface{}, len(i))
	for k, v := range i {
		key := k
		switch val := v.(type) {
		case map[string]interface{}:
			m[key] = toStringMap(val)
		case FixtureBlob:
			m[key] = toStringMap(val)
		default:
			m[key] = v
		}
	}
	return m
}

// String convert FixtureBlob to string
func (fb *FixtureBlob) String() string {
	str, err := json.Marshal(toStringMap(*fb))
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
	BodyType   BodyType     `yaml:"bodyType,omitempty" json:"bodyType,omitempty"`
	BodyString string       `yaml:"bodyString,omitempty" json:"bodyString,omitempty"` // set BodyString if response body is string
	BodyJSON   *FixtureBlob `yaml:"bodyJson,omitempty" json:"bodyJson,omitempty"`     // set Body if response body is object
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
	StatusCode int          `yaml:"statusCode" json:"statusCode"`
	Body       *FixtureBody `yaml:"body,omitempty" json:"body,omitempty"`
}

// FixtureResponse is test fixture for client/endpoint response
type FixtureResponse struct {
	ResponseType     ProtocalType         `yaml:"responseType" json:"responseType"`
	HTTPResponse     *FixtureHTTPResponse `yaml:"httpResponse,omitempty" json:"httpResponse,omitempty"`
	TChannelResponse FixtureBlob          `yaml:"tchannelResponse,omitempty" json:"tchannelResponse,omitempty"`
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
	Method HTTPMethodType `yaml:"method,omitempty" json:"method,omitempty"`
	Body   *FixtureBody   `yaml:"body,omitempty" json:"body,omitempty"`
}

// FixtureRequest is test fixture for client/endpoint request
type FixtureRequest struct {
	RequestType     ProtocalType        `yaml:"requestType" json:"requestType"`
	HTTPRequest     *FixtureHTTPRequest `yaml:"httpRequest,omitempty" json:"httpRequest,omitempty"`
	TChannelRequest FixtureBlob         `yaml:"tchannelRequest,omitempty" json:"tchannelRequest,omitempty"`
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

// FixtureHeaders implements default string used for headers
type FixtureHeaders map[string]interface{}

// EndpointTestFixture saves mocked requests/responses for an endpoint test.
type EndpointTestFixture struct {
	TestName           string                        `yaml:"testName" json:"testName"`
	EndpointID         string                        `yaml:"endpointId" json:"endpointId"`
	HandleID           string                        `yaml:"handleId" json:"handleId"`
	EndpointRequest    FixtureRequest                `yaml:"endpointRequest" json:"endpointRequest"` // there's no difference between http or tchannel request
	EndpointReqHeaders FixtureHeaders                `yaml:"endpointReqHeaders" json:"endpointReqHeaders"`
	EndpointResponse   FixtureResponse               `yaml:"endpointResponse" json:"endpointResponse"`
	EndpointResHeaders FixtureHeaders                `yaml:"endpointResHeaders" json:"endpointResHeaders"`
	ClientTestFixtures map[string]*ClientTestFixture `yaml:"clientTestFixtures" json:"clientTestFixtures"`
	TestServiceName    string                        `yaml:"testServiceName" json:"testServiceName"` // The service module that mounts the endpoint
}

// ClientTestFixture saves mocked client request/response for an endpoint test.
type ClientTestFixture struct {
	ClientID         string          `yaml:"clientId" json:"clientId"`
	ClientMethod     string          `yaml:"clientMethod" json:"clientMethod"`
	ClientRequest    FixtureRequest  `yaml:"clientRequest" json:"clientRequest"` // there's no difference between http or tchannel request
	ClientReqHeaders FixtureHeaders  `yaml:"clientReqHeaders" json:"clientReqHeaders"`
	ClientResponse   FixtureResponse `yaml:"clientResponse" json:"clientResponse"`
	ClientResHeaders FixtureHeaders  `yaml:"clientResHeaders" json:"clientResHeaders"`
}

// NewDefaultModuleSystemWithMockHook creates a fresh instance of the default zanzibar
// module system (clients, endpoints, services) with a post build hook to generate client and service mocks
func NewDefaultModuleSystemWithMockHook(
	h *PackageHelper,
	clientsMock bool,
	workflowMock bool,
	serviceMock bool,
	configFile string,
	parallelizeFactor int,
	selectiveBuilding bool,
	hooks ...PostGenHook,
) (*ModuleSystem, error) {
	t, err := NewDefaultTemplate()
	if err != nil {
		return nil, err
	}

	var clientMockGenHook, workflowMockGenHook, serviceMockGenHook PostGenHook
	if clientsMock {
		clientMockGenHook, err = ClientMockGenHook(h, t, parallelizeFactor)
		if err != nil {
			return nil, errors.Wrap(err, "error creating client mock gen hook")
		}
		hooks = append(hooks, clientMockGenHook)
	}

	if workflowMock {
		workflowMockGenHook = WorkflowMockGenHook(h, t)
		hooks = append(hooks, workflowMockGenHook)
	}

	if serviceMock {
		serviceMockGenHook = ServiceMockGenHook(h, t, configFile)
		hooks = append(hooks, serviceMockGenHook)
	}

	return NewDefaultModuleSystem(h, selectiveBuilding, hooks...)
}

// NewDefaultModuleSystem creates a fresh instance of the default zanzibar
// module system (clients, endpoints, services)
func NewDefaultModuleSystem(
	h *PackageHelper,
	selectiveBuilding bool,
	hooks ...PostGenHook,
) (*ModuleSystem, error) {
	system := NewModuleSystem(h.moduleSearchPaths, h.defaultDependencies, selectiveBuilding, hooks...)

	tmpl, err := NewDefaultTemplate()
	if err != nil {
		return nil, err
	}

	// Register client module class and type generators
	if err := system.RegisterClass(ModuleClass{
		Name:       "client",
		NamePlural: "clients",
		ClassType:  MultiModule,
	}); err != nil {
		return nil, errors.Wrapf(err, "Error registering client class")
	}

	if err := system.RegisterClassType("client", "http", newHTTPClientGenerator(tmpl, h)); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering HTTP client class type",
		)
	}

	if err := system.RegisterClassType("client", "tchannel", newTChannelClientGenerator(tmpl, h)); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering TChannel client class type",
		)
	}

	if err := system.RegisterClassType("client", "custom", newCustomClientGenerator(tmpl, h)); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering custom client class type",
		)
	}

	if err := system.RegisterClassType("client", "grpc", newGRPCClientGenerator(tmpl, h)); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering grpc client class type",
		)
	}

	if err := system.RegisterClass(ModuleClass{
		Name:       "middleware",
		NamePlural: "middlewares",
		ClassType:  MultiModule,
		DependsOn:  []string{"client"},
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering middleware class",
		)
	}

	if err := system.RegisterClassType("middleware", "http", &MiddlewareGenerator{
		templates:     tmpl,
		packageHelper: h,
	}); err != nil {
		return nil, errors.Wrapf(
			err,
			"Error registering Gateway middleware class type",
		)
	}

	if err := system.RegisterClassType("middleware", "tchannel", &MiddlewareGenerator{
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
		Name:       "endpoint",
		NamePlural: "endpoints",
		ClassType:  MultiModule,
		DependsOn:  []string{"client", "middleware"},
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
		Name:       "service",
		NamePlural: "services",
		ClassType:  MultiModule,
		DependsOn:  []string{"endpoint"},
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

func readEndpointConfig(rawConfig []byte) (*EndpointClassConfig, error) {
	var endpointConfig EndpointClassConfig
	if err := yaml.Unmarshal(rawConfig, &endpointConfig); err != nil {
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

// httpClientGenerator generates an instance of a zanzibar http client
type httpClientGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
	validator     *validator2.Validator
}

// newHTTPClientGenerator generates an instance of a zanzibar http client
func newHTTPClientGenerator(templates *Template, packageHelper *PackageHelper) *httpClientGenerator {
	validator := getExposedMethodValidator()
	return &httpClientGenerator{
		templates:     templates,
		packageHelper: packageHelper,
		validator:     validator,
	}
}

// ComputeSpec returns the spec for a HTTP client
func (g *httpClientGenerator) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	// Parse the client config from the endpoint YAML file
	clientConfig, err := newClientConfig(instance.YAMLFileRaw, g.validator)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading HTTP client %q YAML config",
			instance.InstanceName,
		)
	}

	clientSpec, err := clientConfig.NewClientSpec(
		instance,
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error initializing HTTPClientSpec for %q",
			instance.InstanceName,
		)
	}

	return clientSpec, nil
}

// Generate returns the HTTP client build result, which contains the files and
// the generated client spec
func (g *httpClientGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	clientSpecUntyped, err := g.ComputeSpec(instance)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error initializing HTTPClientSpec for %q",
			instance.InstanceName,
		)
	}
	clientSpec := clientSpecUntyped.(*ClientSpec)

	exposedMethods := reverseExposedMethods(clientSpec)

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

// tchannelClientGenerator generates an instance of a zanzibar TChannel client
type tchannelClientGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
	validator     *validator2.Validator
}

// newTChannelClientGenerator generates an instance of a zanzibar TChannel client
func newTChannelClientGenerator(templates *Template, packageHelper *PackageHelper) *tchannelClientGenerator {
	validator := getExposedMethodValidator()
	return &tchannelClientGenerator{
		templates:     templates,
		packageHelper: packageHelper,
		validator:     validator,
	}
}

func getExposedMethodValidator() *validator2.Validator {
	validator := validator2.NewValidator()
	validator.SetValidationFunc("exposedMethods", validateExposedMethods)
	return validator
}

// ComputeSpec computes the TChannel client spec
func (g *tchannelClientGenerator) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	// Parse the client config from the endpoint YAML file
	clientConfig, err := newClientConfig(instance.YAMLFileRaw, g.validator)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading TChannel client %q YAML config",
			instance.InstanceName,
		)
	}

	clientSpec, err := clientConfig.NewClientSpec(
		instance,
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error initializing TChannelClientSpec for %q",
			instance.InstanceName,
		)
	}

	return clientSpec, nil
}

// Generate returns the TChannel client build result, which contains the files
// and the generated client spec
func (g *tchannelClientGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	clientSpecUntyped, err := g.ComputeSpec(instance)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error initializing TChannelClientSpec for %q",
			instance.InstanceName,
		)
	}
	clientSpec := clientSpecUntyped.(*ClientSpec)

	exposedMethods := reverseExposedMethods(clientSpec)

	clientMeta := &ClientMeta{
		Instance:         instance,
		ExportName:       clientSpec.ExportName,
		ExportType:       clientSpec.ExportType,
		Services:         clientSpec.ModuleSpec.Services,
		IncludedPackages: clientSpec.ModuleSpec.IncludedPackages,
		ClientID:         clientSpec.ClientID,
		ExposedMethods:   exposedMethods,
		SidecarRouter:    clientSpec.SidecarRouter,
		DeputyReqHeader:  g.packageHelper.DeputyReqHeader(),
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

	genTestServer, _ := instance.Config["genTestServer"].(bool)
	if genTestServer {
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
		serverFilePath := baseName + "_test_server.go"
		files[serverFilePath] = server
	}

	if dependencies != nil {
		files["module/dependencies.go"] = dependencies
	}

	return &BuildResult{
		Files: files,
		Spec:  clientSpec,
	}, nil
}

// reverse index and filter the exposed methods map as the gen-thrift-spec can be subset
func reverseExposedMethods(clientSpec *ClientSpec) map[string]string {
	reversed := map[string]string{}
	for exposedMethod, idlMethod := range clientSpec.ExposedMethods {
		if hasMethod(clientSpec, idlMethod) {
			reversed[idlMethod] = exposedMethod
		}
	}
	return reversed
}

func hasMethod(cspec *ClientSpec, idlMethod string) bool {
	segments := strings.Split(idlMethod, "::")
	service := segments[0]
	method := segments[1]

	if cspec.ModuleSpec.Services != nil {
		return hasThriftMethod(cspec.ModuleSpec.Services, service, method)
	}

	return hasProtoMethod(cspec.ModuleSpec.ProtoServices, service, method)
}

func hasThriftMethod(thriftSpec []*ServiceSpec, service, method string) bool {
	for _, s := range thriftSpec {
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

func hasProtoMethod(protoSpec []*ProtoService, service, method string) bool {
	for _, s := range protoSpec {
		if s.Name == service {
			for _, m := range s.RPC {
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

// customClientGenerator gathers the custom client spec for future use in ClientsInitGenerator
type customClientGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
	validator     *validator2.Validator
}

// newCustomClientGenerator generates custom client spec for future use in ClientsInitGenerator
func newCustomClientGenerator(templates *Template, packageHelper *PackageHelper) *customClientGenerator {
	validator := validator2.NewValidator()
	return &customClientGenerator{
		templates:     templates,
		packageHelper: packageHelper,
		validator:     validator,
	}
}

// ComputeSpec computes the client spec for a custom client
func (g *customClientGenerator) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	// Parse the client config from the endpoint YAML file
	clientConfig, err := newClientConfig(instance.YAMLFileRaw, g.validator)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading custom client %q YAML config",
			instance.InstanceName,
		)
	}

	clientSpec, err := clientConfig.NewClientSpec(
		instance,
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error initializing CustomClientSpec for %q",
			instance.InstanceName,
		)
	}

	return clientSpec, nil
}

// Generate returns the custom client build result, which contains the
// generated client spec and no files
func (g *customClientGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	clientSpecUntyped, err := g.ComputeSpec(instance)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error initializing CustomClientSpec for %q",
			instance.InstanceName,
		)
	}
	clientSpec := clientSpecUntyped.(*ClientSpec)

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
 * gRPC client generator
 */

// gRPCClientGenerator generates grpc clients.
type gRPCClientGenerator struct {
	templates     *Template
	packageHelper *PackageHelper
	validator     *validator2.Validator
}

// NewgrpcClientGenerator generates an instance of a zanzibar grpc client
func newGRPCClientGenerator(templates *Template, packageHelper *PackageHelper) *gRPCClientGenerator {
	validator := getExposedMethodValidator()
	return &gRPCClientGenerator{
		templates:     templates,
		packageHelper: packageHelper,
		validator:     validator,
	}
}

// ComputeSpec returns the spec for a gRPC client
func (g *gRPCClientGenerator) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	// Parse the client config from the endpoint YAML file
	clientConfig, err := newClientConfig(instance.YAMLFileRaw, g.validator)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error reading gRPC client %q YAML config",
			instance.InstanceName,
		)
	}

	clientSpec, err := clientConfig.NewClientSpec(
		instance,
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error initializing gRPCClientSpec for %q",
			instance.InstanceName,
		)
	}

	return clientSpec, nil
}

// Generate returns the gRPC client build result, which contains the files and
// the generated client spec
func (g *gRPCClientGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	clientSpecUntyped, err := g.ComputeSpec(instance)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error initializing GRPCClientSpec for %q",
			instance.InstanceName,
		)
	}
	clientSpec := clientSpecUntyped.(*ClientSpec)

	reversedMethods := reverseExposedMethods(clientSpec)

	// @rpatali: Update all struct to use more general field IDLFile instead of thriftFile.
	clientMeta := &ClientMeta{
		ProtoServices:    clientSpec.ModuleSpec.ProtoServices,
		Instance:         instance,
		ExportName:       clientSpec.ExportName,
		ExportType:       clientSpec.ExportType,
		Services:         nil,
		IncludedPackages: clientSpec.ModuleSpec.IncludedPackages,
		ClientID:         clientSpec.ClientID,
		ExposedMethods:   reversedMethods,
	}

	client, err := g.templates.ExecTemplate(
		"grpc_client.tmpl",
		clientMeta,
		g.packageHelper,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error executing gRPC client template for %q",
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
		Spec:  (*ClientSpec)(nil),
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

// ComputeSpec computes the endpoint specs for a group of endpoints
func (g *EndpointGenerator) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	endpointYamls := []string{}
	endpointSpecs := []*EndpointSpec{}
	clientSpecs := readClientDependencySpecs(instance)

	endpointConfig, err := readEndpointConfig(instance.YAMLFileRaw)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading HTTP endpoint %q YAML config",
			instance.InstanceName,
		)
	}

	endpointConfigDir := filepath.Join(
		instance.BaseDirectory,
		instance.Directory,
	)
	for _, fileName := range endpointConfig.Config.Endpoints {
		endpointYamls = append(
			endpointYamls, filepath.Join(endpointConfigDir, fileName),
		)
	}
	var wg sync.WaitGroup
	wg.Add(len(endpointYamls))
	ch := make(chan endpointSpecRes, len(endpointYamls))
	for _, yamlFile := range endpointYamls {
		go func(yamlFile string) {
			defer wg.Done()
			espec, err := NewEndpointSpec(yamlFile, g.packageHelper, g.packageHelper.MiddlewareSpecs())
			if err != nil {
				err = errors.Wrapf(
					err, "Error creating endpoint spec for endpoint: %s", yamlFile,
				)
				ch <- endpointSpecRes{err: err}
				return
			}
			err = espec.SetDownstream(clientSpecs, g.packageHelper)
			if err != nil {
				err = errors.Wrapf(
					err, "Error parsing downstream info for endpoint: %s", yamlFile,
				)
				ch <- endpointSpecRes{err: err}
				return
			}
			ch <- endpointSpecRes{espec: espec, err: err}
		}(yamlFile)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	for endpointSpecRes := range ch {
		if endpointSpecRes.err != nil {
			return nil, endpointSpecRes.err
		}
		endpointSpecs = append(endpointSpecs, endpointSpecRes.espec)
	}

	return endpointSpecs, nil
}

type endpointSpecRes struct {
	espec *EndpointSpec
	err   error
}

// Generate returns the endpoint build result, which contains a file per
// endpoint handler and a list of handler specs
func (g *EndpointGenerator) Generate(
	instance *ModuleInstance,
) (*BuildResult, error) {
	endpointMeta := make([]*EndpointMeta, 0)
	endpointSpecsUntyped, err := g.ComputeSpec(instance)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error generating endpoint specs for %q",
			instance.InstanceName,
		)
	}
	endpointSpecs := endpointSpecsUntyped.([]*EndpointSpec)

	for _, espec := range endpointSpecs {
		// allow configured header to pass down to switch downstream service dynmamic
		reqHeaders := espec.ReqHeaders
		if reqHeaders == nil {
			reqHeaders = make(map[string]*TypedHeader)
		}
		shk := textproto.CanonicalMIMEHeaderKey(g.packageHelper.DeputyReqHeader())
		reqHeaders[shk] = &TypedHeader{
			Name:        shk,
			TransformTo: shk,
			Field:       &compile.FieldSpec{Required: false},
		}
		espec.ReqHeaders = reqHeaders
	}
	var fileMap sync.Map
	runner := parallelize.NewUnboundedRunner(len(endpointSpecs) + 1)
	for _, espec := range endpointSpecs {
		f := func(especInf interface{}) (interface{}, error) {
			espec := especInf.(*EndpointSpec)
			meta, err := g.generateEndpointFile(espec, instance, &fileMap)
			if err != nil {
				err = errors.Wrapf(
					err,
					"Error executing endpoint template %q",
					instance.InstanceName,
				)
				return nil, err
			}
			err = g.generateEndpointTestFile(espec, instance, &fileMap)
			if err != nil {
				err = errors.Wrapf(
					err,
					"Error executing endpoint test template %q",
					instance.InstanceName,
				)
				return nil, err
			}
			return meta, err
		}
		wrk := &parallelize.SingleParamWork{Data: espec, Func: f}
		runner.SubmitWork(wrk)
	}

	f := func() (interface{}, error) {
		dependencies, err := GenerateDependencyStruct(
			instance,
			g.packageHelper,
			g.templates,
		)
		if err != nil {
			err = errors.Wrapf(
				err,
				"Error generating service dependencies for %s",
				instance.InstanceName,
			)
			return nil, err
		}
		if dependencies != nil {
			fileMap.Store("module/dependencies.go", dependencies)
		}
		return nil, nil
	}
	runner.SubmitWork(parallelize.StatelessFunc(f))

	results, err := runner.GetResult()
	if err != nil {
		return nil, err
	}
	for _, meta := range results {
		if meta == nil {
			continue
		}
		endpointMeta = append(endpointMeta, meta.(*EndpointMeta))
	}

	// sort for deterministic code gen order in endpoint.go file
	sort.SliceStable(endpointMeta, func(i, j int) bool {
		return endpointMeta[i].Spec.HandleID < endpointMeta[j].Spec.HandleID
	})

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
	fileMap.Store("endpoint.go", endpointCollection)

	files := make(map[string][]byte)
	fileMap.Range(func(key, value interface{}) bool {
		files[key.(string)] = value.([]byte)
		return true
	})

	return &BuildResult{
		Files: files,
		Spec:  endpointSpecs,
	}, nil
}

type endpointMetaResult struct {
	meta *EndpointMeta
	err  error
}

func (g *EndpointGenerator) generateEndpointFile(e *EndpointSpec, instance *ModuleInstance,
	out *sync.Map) (*EndpointMeta, error) {
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
		if _, ok := out.Load(structFilePath); !ok {
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
			out.Store(structFilePath, structs)
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
	includedPackages = append(includedPackages, GoPackageImport{
		PackageName: instance.PackageInfo.GeneratedPackagePath + "/workflow",
		AliasName:   "workflow",
	})
	if e.WorkflowType == customWorkflow {
		includedPackages = append(includedPackages, GoPackageImport{
			PackageName: e.WorkflowImportPath,
			AliasName:   "custom" + strings.Title(m.PackageName),
		})
	}

	workflowPkg := "workflow"
	if method.Downstream == nil && e.WorkflowType == "custom" {
		workflowPkg = "custom" + strings.Title(m.PackageName)
	}

	clientID := e.ClientID
	clientName := ""
	if e.ClientSpec != nil {
		clientName = e.ClientSpec.ClientName
	}

	// TODO: http client needs to support multiple thrift services
	meta := &EndpointMeta{
		Instance:               instance,
		Spec:                   e,
		GatewayPackageName:     g.packageHelper.GoGatewayPackageName(),
		IncludedPackages:       includedPackages,
		Method:                 method,
		ReqHeaders:             e.ReqHeaders,
		IsClientlessEndpoint:   e.IsClientlessEndpoint,
		ReqHeadersKeys:         sortedHeaders(e.ReqHeaders, false),
		ReqRequiredHeadersKeys: sortedHeaders(e.ReqHeaders, true),
		ResHeadersKeys:         sortedHeaders(e.ResHeaders, false),
		ResRequiredHeadersKeys: sortedHeaders(e.ResHeaders, true),
		ResHeaders:             e.ResHeaders,
		ClientID:               clientID,
		ClientName:             clientName,
		ClientMethodName:       e.ClientMethod,
		WorkflowPkg:            workflowPkg,
		TraceKey:               g.packageHelper.traceKey,
		DeputyReqHeader:        g.packageHelper.DeputyReqHeader(),
		DefaultHeaders:         e.DefaultHeaders,
	}

	targetPath := e.TargetEndpointPath(thriftServiceName, method.Name)
	if e.EndpointType == "tchannel" {
		targetPath = strings.TrimSuffix(targetPath, ".go") + "_tchannel.go"
	}
	endpointFilePath, err := filepath.Rel(endpointDirectory, targetPath)
	if err != nil {
		endpointFilePath = targetPath
	}

	workCount := 2
	runner := parallelize.NewUnboundedRunner(workCount)
	f := func() (interface{}, error) {
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
		out.Store(endpointFilePath, endpoint)
		return nil, nil
	}
	runner.SubmitWork(parallelize.StatelessFunc(f))

	f = func() (interface{}, error) {
		var tmpl string
		if e.IsClientlessEndpoint {
			tmpl = "clientless-workflow.tmpl"
		} else {
			tmpl = "workflow.tmpl"
		}
		workflow, err := g.templates.ExecTemplate(tmpl, meta, g.packageHelper)
		if err != nil {
			return nil, errors.Wrap(err, "Error executing workflow template")
		}
		out.Store("workflow/"+endpointFilePath, workflow)
		return nil, nil
	}
	runner.SubmitWork(parallelize.StatelessFunc(f))
	if _, err := runner.GetResult(); err != nil {
		return nil, err
	}

	return meta, nil
}

func (g *EndpointGenerator) generateEndpointTestFile(
	e *EndpointSpec, instance *ModuleInstance, out *sync.Map,
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
		ReqHeaders:   e.ReqHeaders,
		ResHeaders:   e.ResHeaders,
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

	out.Store(endpointTestFilePath, endpointTest)

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

// ComputeSpec computes the spec for a gateway
func (generator *GatewayServiceGenerator) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	return nil, nil
}

// Generate returns the gateway build result, which contains the service and
// service test main files, and no spec
func (generator *GatewayServiceGenerator) Generate(instance *ModuleInstance) (*BuildResult, error) {
	var fileMap sync.Map
	workCount := 5
	runner := parallelize.NewUnboundedRunner(workCount)

	f := func() (interface{}, error) {
		service, err := generator.templates.ExecTemplate(
			"service.tmpl",
			instance,
			generator.packageHelper,
		)
		if err != nil {
			err = errors.Wrapf(
				err,
				"Error generating service service.go for %s",
				instance.InstanceName,
			)
			return nil, err
		}
		fileMap.Store("service.go", service)
		return nil, nil
	}
	runner.SubmitWork(parallelize.StatelessFunc(f))

	f = func() (interface{}, error) {
		// generate main.go
		main, err := generator.templates.ExecTemplate(
			"main.tmpl",
			instance,
			generator.packageHelper,
		)
		if err != nil {
			err = errors.Wrapf(
				err,
				"Error generating service main.go for %s",
				instance.InstanceName,
			)
			return nil, err
		}
		fileMap.Store("main/main.go", main)
		return nil, nil
	}
	runner.SubmitWork(parallelize.StatelessFunc(f))

	f = func() (interface{}, error) {
		// generate main_test.go
		mainTest, err := generator.templates.ExecTemplate(
			"main_test.tmpl",
			instance,
			generator.packageHelper,
		)
		if err != nil {
			err = errors.Wrapf(
				err,
				"Error generating service main_test.go for %s",
				instance.InstanceName,
			)
			return nil, err
		}
		fileMap.Store("main/main_test.go", mainTest)
		return nil, nil
	}
	runner.SubmitWork(parallelize.StatelessFunc(f))

	f = func() (interface{}, error) {
		dependencies, err := GenerateDependencyStruct(
			instance,
			generator.packageHelper,
			generator.templates,
		)
		if err != nil {
			err = errors.Wrapf(
				err,
				"Error generating service dependencies for %s",
				instance.InstanceName,
			)
			return nil, err
		}
		fileMap.Store("module/dependencies.go", dependencies)
		return nil, nil
	}
	runner.SubmitWork(parallelize.StatelessFunc(f))

	f = func() (interface{}, error) {
		initializer, err := GenerateInitializer(
			instance,
			generator.packageHelper,
			generator.templates,
		)
		if err != nil {
			err = errors.Wrapf(
				err,
				"Error generating service initializer for %s",
				instance.InstanceName,
			)
			return nil, err
		}
		fileMap.Store("module/init.go", initializer)
		return nil, nil
	}
	runner.SubmitWork(parallelize.StatelessFunc(f))

	if _, err := runner.GetResult(); err != nil {
		return nil, err
	}

	files := make(map[string][]byte)
	fileMap.Range(func(key, value interface{}) bool {
		files[key.(string)] = value.([]byte)
		return true
	})

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

// ComputeSpec computes the spec for a middleware
func (g *MiddlewareGenerator) ComputeSpec(
	instance *ModuleInstance,
) (interface{}, error) {
	return nil, nil
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
	templateName := "middleware_http.tmpl"
	if instance.ClassType == "tchannel" {
		templateName = "middleware_tchannel.tmpl"
	}

	bytes, err := g.templates.ExecTemplate(templateName, instance, g.packageHelper)
	if err != nil {
		return err
	}

	baseName := filepath.Base(instance.Directory)

	out[baseName+".go"] = bytes

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
	genCustom, _ := instance.Config["customInterface"].(string)
	if genCustom != "" {
		instance.PackageInfo.ExportType = instance.Config["customInterface"].(string)
	}
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
	ProtoServices    []*ProtoService
	ExposedMethods   map[string]string
	SidecarRouter    string
	Fixture          *Fixture
	DeputyReqHeader  string
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
