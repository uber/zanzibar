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
	"path/filepath"
	"runtime"
	"strings"

	"sort"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/runtime"
)

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)
	return zanzibar.GetDirnameFromRuntimeCaller(file)
}

var mandatoryClientFields = []string{
	"clientType",
	"clientId",
	"thriftFile",
	"thriftFileSha",
	"clientName",
	"serviceName",
}
var mandatoryCustomClientFields = []string{
	"customImportPath",
	"customClientType",
	"customPackageName",
	"clientName",
}
var mandatoryEndpointFields = []string{
	"endpointType",
	"endpointId",
	"handleId",
	"thriftFile",
	"thriftFileSha",
	"thriftMethodName",
	"workflowType",
}
var mandatoryHTTPEndpointFields = []string{
	"testFixtures",
	"middlewares",
}

// ClientSpec holds information about each client in the
// gateway included its thriftFile and other meta info
type ClientSpec struct {
	// ModuleSpec holds the thrift module information
	ModuleSpec *ModuleSpec
	// JSONFile for this spec
	JSONFile string
	// ClientType, currently "http", "tchannel" and "custom" are supported
	ClientType string
	// If "custom" then where to import custom code from
	CustomImportPath string
	// If "custom" then what the client type (in Go) in the custom package
	CustomClientType string
	// If "custom" then what the package name (in Go) the client package has
	CustomPackageName string
	// GoFileName, the absolute path where the generate client is
	GoFileName string
	// GoPackageName is the golang package name for the client
	GoPackageName string
	// GoStructsFileName, absolute path where any helper structs
	// are generated for this generated client
	GoStructsFileName string
	// ThriftFile, absolute path to thrift file
	ThriftFile string
	// ClientID, used for logging and metrics, must be lowercase
	// and use dashes.
	ClientID string
	// ClientName, PascalCase name of the client, the generated
	// `Clients` struct will contain a field of this name
	ClientName string
	// ThriftServiceName, if the thrift file has multiple
	// services then this is the service that describes the client
	// TODO: this field needs to be deprecated for multi-service support
	ThriftServiceName string
	// ExposedMethods is a map of exposed method name to thrift "$service::$method"
	// only the method values in this map are generated for the client
	ExposedMethods map[string]string
}

// ModuleClassConfig represents the generic JSON config for
// all modules. This will be provided by the module package.
type ModuleClassConfig struct {
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Config interface{} `json:"config"`
}

// ClientClassConfig represents the specific config for
// a client. This is a downcast of the moduleClassConfig.
type ClientClassConfig struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// NewClientSpec creates a client spec from a json file.
func NewClientSpec(jsonFile string, h *PackageHelper) (*ClientSpec, error) {
	_, err := os.Stat(jsonFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not find file %s: ", jsonFile)
	}

	bytes, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not read json file %s: ", jsonFile,
		)
	}

	clientConfig := &ClientClassConfig{}

	if err := json.Unmarshal(bytes, &clientConfig); err != nil {
		return nil, errors.Wrapf(
			err, "Could not parse class config json file %s: ", jsonFile,
		)
	}

	// Restore the properties in the old config structure
	clientConfig.Config["clientId"] = clientConfig.Name
	clientConfig.Config["clientType"] = clientConfig.Type

	switch clientConfig.Type {
	case "http":
		return NewHTTPClientSpec(jsonFile, clientConfig, h)
	case "tchannel":
		return NewTChannelClientSpec(jsonFile, clientConfig, h)
	case "custom":
		return NewCustomClientSpec(jsonFile, clientConfig, h)
	default:
		return nil, errors.Errorf(
			"Cannot support unknown clientType for client %s", jsonFile,
		)
	}
}

// NewTChannelClientSpec creates a client spec from a json file whose type is tchannel
func NewTChannelClientSpec(jsonFile string, clientConfig *ClientClassConfig, h *PackageHelper) (*ClientSpec, error) {
	exposedMethods := clientConfig.Config["exposedMethods"].(map[string]interface{})
	if len(exposedMethods) == 0 {
		return nil, errors.Errorf("No methods are exposed in client config: %s", jsonFile)
	}

	cspec, err := newClientSpec(jsonFile, clientConfig, false, h)
	if err != nil {
		return nil, err
	}

	cspec.ExposedMethods = map[string]string{}
	reversed := map[string]string{}
	for key, val := range exposedMethods {
		cspec.ExposedMethods[key] = val.(string)
		reversed[val.(string)] = key
	}

	if len(cspec.ExposedMethods) != len(reversed) {
		return nil, errors.Errorf("Keys or values of the exposedMethods of are not unique: %s", jsonFile)
	}

	return cspec, nil
}

// NewCustomClientSpec creates a client spec from a json file whose type is custom
func NewCustomClientSpec(jsonFile string, clientConfig *ClientClassConfig, h *PackageHelper) (*ClientSpec, error) {
	for _, f := range mandatoryCustomClientFields {
		if _, ok := clientConfig.Config[f]; !ok {
			return nil, errors.Errorf(
				"client config %q must have %q field for type custom", jsonFile, f,
			)
		}
	}

	clientSpec := &ClientSpec{
		JSONFile:          jsonFile,
		ClientType:        clientConfig.Type,
		ClientID:          clientConfig.Config["clientId"].(string),
		ClientName:        clientConfig.Config["clientName"].(string),
		CustomImportPath:  clientConfig.Config["customImportPath"].(string),
		CustomClientType:  clientConfig.Config["customClientType"].(string),
		CustomPackageName: clientConfig.Config["customPackageName"].(string),
	}

	return clientSpec, nil
}

// NewHTTPClientSpec creates a client spec from a json file whose type is http
func NewHTTPClientSpec(jsonFile string, clientConfig *ClientClassConfig, h *PackageHelper) (*ClientSpec, error) {
	return newClientSpec(jsonFile, clientConfig, true, h)

}

func newClientSpec(jsonFile string, clientConfig *ClientClassConfig, wantAnnot bool, h *PackageHelper) (*ClientSpec, error) {
	config := clientConfig.Config

	for i := 0; i < len(mandatoryClientFields); i++ {
		fieldName := mandatoryClientFields[i]
		if _, ok := config[fieldName]; !ok {
			return nil, errors.Errorf(
				"client config (%s) must have %s field", jsonFile, fieldName,
			)
		}
	}

	thriftFile := filepath.Join(
		h.ThriftIDLPath(), config["thriftFile"].(string),
	)

	mspec, err := NewModuleSpec(thriftFile, wantAnnot, h)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not build module spec for thrift %s: ", thriftFile,
		)
	}
	mspec.PackageName = mspec.PackageName + "Client"

	baseName := filepath.Base(filepath.Dir(jsonFile))

	goFileName := filepath.Join(
		h.CodeGenTargetPath(),
		"clients",
		baseName,
		baseName+".go",
	)

	goPackageName := filepath.Join(
		h.GoGatewayPackageName(),
		"clients",
		baseName,
	)

	goStructsFileName := filepath.Join(
		h.CodeGenTargetPath(),
		"clients",
		baseName,
		baseName+"_structs.go",
	)

	return &ClientSpec{
		ModuleSpec:        mspec,
		JSONFile:          jsonFile,
		ClientType:        clientConfig.Type,
		GoFileName:        goFileName,
		GoPackageName:     goPackageName,
		GoStructsFileName: goStructsFileName,
		ThriftFile:        thriftFile,
		ClientID:          config["clientId"].(string),
		ClientName:        config["clientName"].(string),
		ThriftServiceName: config["serviceName"].(string),
	}, nil
}

// MiddlewareSpec holds information about each middleware at the endpoint
// level. The same mid
type MiddlewareSpec struct {
	// The middleware package name.
	Name string
	// Go import path for the middleware.
	Path string
	// Middleware specific configuration options.
	Options map[string]interface{}
}

// NewMiddlewareSpec creates a middleware spec from a go file.
func NewMiddlewareSpec(
	name string,
	goFile string,
	jsonFile string,
	configDirName string,
) (*MiddlewareSpec, error) {
	schPath := filepath.Join(
		configDirName,
		jsonFile,
	)

	bytes, err := ioutil.ReadFile(schPath)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Cannot read middleware json schema: %s",
			schPath,
		)
	}

	var midOptSchema map[string]interface{}
	err = json.Unmarshal(bytes, &midOptSchema)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Cannot parse json schema for middleware options: %s",
			schPath,
		)
	}

	// TODO(sindelar): Add middleware validation here. Validate name
	// and package name match. Validate the options json schema matches the options
	// struct
	return &MiddlewareSpec{
		Name: name,
		Path: goFile,
	}, nil
}

// EndpointSpec holds information about each endpoint in the
// gateway including its thriftFile and meta data
type EndpointSpec struct {
	// ModuleSpec holds the thrift module info
	ModuleSpec *ModuleSpec
	// JSONFile for this endpoint spec
	JSONFile string
	// GoStructsFileName is where structs are generated
	GoStructsFileName string
	// GoFolderName is the folder where all the endpoints
	// are generated.
	GoFolderName string
	// GoPackageName is the package import path.
	GoPackageName string

	// EndpointType, currently only "http"
	EndpointType string
	// EndpointID, used in metrics and logging, lower case.
	EndpointID string
	// HandleID, used in metrics and logging, lowercase.
	HandleID string
	// ThriftFile, the thrift file for this endpoint
	ThriftFile string
	// ThriftMethodName, which thrift method to use.
	ThriftMethodName string
	// ThriftServiceName, which thrift service to use.
	ThriftServiceName string
	// TestFixtures, meta data to generate tests,
	// TODO figure out struct type
	TestFixtures []interface{}
	// Middlewares, meta data to add middlewares,
	Middlewares []MiddlewareSpec

	// ReqHeaderMap, maps headers from server to client.
	// Keeps keys in a sorted array so that goldenfiles have
	// deterministic orderings
	ReqHeaderMap     map[string]string
	ReqHeaderMapKeys []string
	// ResHeaderMap, maps headers from client to server.
	// Keeps keys in a sorted array so that goldenfiles have
	// deterministic orderings
	ResHeaderMap     map[string]string
	ResHeaderMapKeys []string

	// WorkflowType, either "httpClient" or "custom".
	// A httpClient workflow generates a http client Caller
	// A custom workflow just imports the custom code
	WorkflowType string
	// If "custom" then where to import custom code from
	WorkflowImportPath string
	// if "httpClient", which client to call.
	ClientName string
	// if "httpClient", which client method to call.
	ClientMethod string
}

func ensureFields(config map[string]interface{}, mandatoryFields []string, jsonFile string) error {
	for i := 0; i < len(mandatoryFields); i++ {
		fieldName := mandatoryFields[i]
		if _, ok := config[fieldName]; !ok {
			return errors.Errorf(
				"config (%s) must have %s field", jsonFile, fieldName,
			)
		}
	}
	return nil
}

// NewEndpointSpec creates an endpoint spec from a json file.
func NewEndpointSpec(
	jsonFile string,
	h *PackageHelper,
	midSpecs map[string]*MiddlewareSpec,
) (*EndpointSpec, error) {
	_, err := os.Stat(jsonFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not find file %s: ", jsonFile)
	}

	bytes, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not read json file %s: ", jsonFile,
		)
	}

	endpointConfigObj := map[string]interface{}{}
	err = json.Unmarshal(bytes, &endpointConfigObj)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not parse json file %s: ", jsonFile,
		)
	}

	if err := ensureFields(endpointConfigObj, mandatoryEndpointFields, jsonFile); err != nil {
		return nil, err
	}

	endpointType := endpointConfigObj["endpointType"]
	if endpointType == "http" {
		if err := ensureFields(endpointConfigObj, mandatoryHTTPEndpointFields, jsonFile); err != nil {
			return nil, err
		}

	}
	if endpointType != "http" && endpointType != "tchannel" {
		return nil, errors.Errorf(
			"Cannot support unknown endpointType for endpoint %s", jsonFile,
		)
	}

	thriftFile := filepath.Join(
		h.ThriftIDLPath(), endpointConfigObj["thriftFile"].(string),
	)

	mspec, err := NewModuleSpec(thriftFile, endpointType == "http", h)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not build module spec for thrift %s: ", thriftFile,
		)
	}

	var workflowImportPath string
	var clientName string
	var clientMethod string

	workflowType := endpointConfigObj["workflowType"].(string)
	if workflowType == "httpClient" || workflowType == "tchannelClient" {
		iclientName, ok := endpointConfigObj["clientName"]
		if !ok {
			return nil, errors.Errorf(
				"endpoint config (%s) must have clientName field", jsonFile,
			)
		}
		clientName = iclientName.(string)

		iclientMethod, ok := endpointConfigObj["clientMethod"]
		if !ok {
			return nil, errors.Errorf(
				"endpoint config (%s) must have clientMethod field", jsonFile,
			)
		}
		clientMethod = iclientMethod.(string)
	} else if workflowType == "custom" {
		iworkflowImportPath, ok := endpointConfigObj["workflowImportPath"]
		if !ok {
			return nil, errors.Errorf(
				"endpoint config (%s) must have workflowImportPath field",
				jsonFile,
			)
		}
		workflowImportPath = iworkflowImportPath.(string)
	} else {
		return nil, errors.Errorf(
			"Invalid workflowType (%s) for endpoint (%s)",
			workflowType, jsonFile,
		)
	}

	dirName := filepath.Base(filepath.Dir(jsonFile))

	goFolderName := filepath.Join(
		h.CodeGenTargetPath(),
		"endpoints",
		dirName,
	)

	goStructsFileName := filepath.Join(
		h.CodeGenTargetPath(),
		"endpoints",
		dirName,
		dirName+"_structs.go",
	)

	goPackageName := filepath.Join(
		h.GoGatewayPackageName(),
		"endpoints",
		dirName,
	)

	thriftInfo := endpointConfigObj["thriftMethodName"].(string)
	parts := strings.Split(thriftInfo, "::")
	if len(parts) != 2 {
		return nil, errors.Errorf(
			"Cannot read thriftMethodName (%s) for endpoint json file %s : ",
			thriftInfo, jsonFile,
		)
	}

	espec := &EndpointSpec{
		ModuleSpec:         mspec,
		JSONFile:           jsonFile,
		GoStructsFileName:  goStructsFileName,
		GoFolderName:       goFolderName,
		GoPackageName:      goPackageName,
		EndpointType:       endpointConfigObj["endpointType"].(string),
		EndpointID:         endpointConfigObj["endpointId"].(string),
		HandleID:           endpointConfigObj["handleId"].(string),
		ThriftFile:         thriftFile,
		ThriftServiceName:  parts[0],
		ThriftMethodName:   parts[1],
		WorkflowType:       workflowType,
		WorkflowImportPath: workflowImportPath,
		ClientName:         clientName,
		ClientMethod:       clientMethod,
	}

	if endpointType == "tchannel" {
		return espec, nil
	}
	return augmentHTTPEndpointSpec(espec, endpointConfigObj, midSpecs)
}

func augmentHTTPEndpointSpec(
	espec *EndpointSpec,
	endpointConfigObj map[string]interface{},
	midSpecs map[string]*MiddlewareSpec,
) (*EndpointSpec, error) {
	espec.TestFixtures = endpointConfigObj["testFixtures"].([]interface{})

	endpointMids, ok := endpointConfigObj["middlewares"].([]interface{})
	if !ok {
		return nil, errors.Errorf(
			"Unable to parse middlewares field",
		)
	}
	middlewares := make([]MiddlewareSpec, len(endpointMids))
	for idx, middleware := range endpointMids {
		middlewareObj, ok := middleware.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf(
				"Unable to parse middleware %s",
				middlewareObj,
			)
		}
		name, ok := middlewareObj["name"].(string)
		if !ok {
			return nil, errors.Errorf(
				"Unable to parse \"name\" field in middleware %s",
				middlewareObj,
			)
		}
		// Verify the middleware name is defined.
		if midSpecs[name] == nil {
			return nil, errors.Errorf(
				"middlewares config (%s) not found.", name,
			)
		}
		// TODO(sindelar): Validate Options against middleware spec and support
		// nested typed objects.
		opts, ok := middlewareObj["options"].(map[string]interface{})
		if !ok {
			opts = make(map[string]interface{})
		}

		middlewares[idx] = MiddlewareSpec{
			Name:    name,
			Path:    midSpecs[name].Path,
			Options: opts,
		}
	}
	espec.Middlewares = middlewares

	reqHeaderMap := make(map[string]string)
	m, ok := endpointConfigObj["reqHeaderMap"]
	if !ok {
		return nil, errors.Errorf(
			"Unable to parse reqHeaderMap %s",
			reqHeaderMap,
		)
	}
	// Do a deep cast to enforce a string -> string map
	castMap := m.(map[string]interface{})
	for key, value := range castMap {
		switch value := value.(type) {
		case string:
			reqHeaderMap[key] = value
		default:
			return nil, errors.Errorf(
				"Unable to parse string %s in reqHeaderMap %s",
				value,
				reqHeaderMap,
			)
		}
	}
	reqHeaderMapKeys := make([]string, len(reqHeaderMap))
	i := 0
	for k := range reqHeaderMap {
		reqHeaderMapKeys[i] = k
		i++
	}
	sort.Strings(reqHeaderMapKeys)
	espec.ReqHeaderMap = reqHeaderMap
	espec.ReqHeaderMapKeys = reqHeaderMapKeys

	resHeaderMap := make(map[string]string)
	m2, ok := endpointConfigObj["resHeaderMap"]
	if !ok {
		return nil, errors.Errorf(
			"Unable to parse resHeaderMap %s",
			resHeaderMap,
		)
	}
	// Do a deep cast to enforce a string -> string map
	castMap = m2.(map[string]interface{})
	for key, value := range castMap {
		switch value := value.(type) {
		case string:
			resHeaderMap[key] = value
		default:
			return nil, errors.Errorf(
				"Unable to parse string %s in resHeaderMap %s",
				value,
				resHeaderMap,
			)
		}
	}
	resHeaderMapKeys := make([]string, len(resHeaderMap))
	i = 0
	for k := range resHeaderMap {
		resHeaderMapKeys[i] = k
		i++
	}
	sort.Strings(resHeaderMapKeys)
	espec.ResHeaderMap = resHeaderMap
	espec.ResHeaderMapKeys = resHeaderMapKeys

	return espec, nil
}

// TargetEndpointPath generates a filepath for each endpoint method
func (e *EndpointSpec) TargetEndpointPath(
	serviceName string, methodName string,
) string {
	baseName := filepath.Base(e.GoFolderName)

	fileName := baseName + "_" + strings.ToLower(serviceName) +
		"_method_" + strings.ToLower(methodName) + ".go"
	return filepath.Join(e.GoFolderName, fileName)
}

// TargetEndpointTestPath generates a filepath for each endpoint test
func (e *EndpointSpec) TargetEndpointTestPath(
	serviceName string, methodName string,
) string {
	baseName := filepath.Base(e.GoFolderName)

	fileName := baseName + "_" + strings.ToLower(serviceName) +
		"_method_" + strings.ToLower(methodName) + "_test.go"
	return filepath.Join(e.GoFolderName, fileName)
}

// EndpointTestConfigPath generates a filepath for each endpoint test config
func (e *EndpointSpec) EndpointTestConfigPath() string {
	return strings.TrimSuffix(e.JSONFile, filepath.Ext(e.JSONFile)) + "_test.json"
}

// SetDownstream configures the downstream client for this endpoint spec
func (e *EndpointSpec) SetDownstream(
	clientModules []*ClientSpec,
	h *PackageHelper,
) error {
	if e.WorkflowType == "custom" {
		return nil
	}

	var clientSpec *ClientSpec
	for _, v := range clientModules {
		if v.ClientName == e.ClientName {
			clientSpec = v
			break
		}
	}

	if clientSpec == nil {
		return errors.Errorf(
			"When parsing endpoint json (%s), "+
				"could not find client (%s) in gateway",
			e.JSONFile, e.ClientName,
		)
	}

	return e.ModuleSpec.SetDownstream(
		e.ThriftServiceName, e.ThriftMethodName,
		clientSpec, e.ClientMethod, h,
	)
}

// EndpointClassConfig represents the specific config for
// an endpoint group. This is a downcast of the moduleClassConfig.
type EndpointClassConfig struct {
	Name         string              `json:"name"`
	Type         string              `json:"type"`
	Dependencies map[string][]string `json:"dependencies"`
	Config       struct {
		Ratelimit int32    `json:"rateLimit"`
		Endpoints []string `json:"endpoints"`
	} `json:"config"`
}

func parseEndpointJsons(
	endpointGroupJsons []string,
) ([]string, error) {
	endpointJsons := []string{}

	for _, endpointGroupJSON := range endpointGroupJsons {
		bytes, err := ioutil.ReadFile(endpointGroupJSON)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot read endpoint group json: %s",
				endpointGroupJSON,
			)
		}

		var endpointConfig EndpointClassConfig
		err = json.Unmarshal(bytes, &endpointConfig)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot parse json for endpoint group config: %s",
				endpointGroupJSON,
			)
		}

		endpointConfigDir := filepath.Dir(endpointGroupJSON)
		for _, fileName := range endpointConfig.Config.Endpoints {
			endpointJsons = append(
				endpointJsons, filepath.Join(endpointConfigDir, fileName),
			)
		}
	}

	return endpointJsons, nil
}

func parseMiddlewareConfig(
	config string,
	configDirName string,
) ([]*MiddlewareSpec, error) {
	bytes, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Cannot read middleware config json: %s",
			config,
		)
	}

	// TODO(sindelar): Use a struct
	var configJSON map[string]interface{}

	err = json.Unmarshal(bytes, &configJSON)

	if err != nil {
		return nil, errors.Wrapf(
			err, "Cannot parse json for middleware config json: %s",
			config,
		)
	}

	midList, ok := configJSON["middlewares"].([]interface{})
	if !ok {
		return nil, errors.Wrapf(
			err, "Cannot parse json for middleware config json: %s",
			config,
		)
	}

	specs := make([]*MiddlewareSpec, len(midList))
	for idx, mid := range midList {
		mid, ok := mid.(map[string]interface{})
		if !ok {
			return nil, errors.Wrapf(
				err, "Cannot parse json for middleware config json: %s",
				config,
			)
		}
		name, okOne := mid["name"].(string)
		schema, okTwo := mid["schema"].(string)
		importPath, okThree := mid["importPath"].(string)
		if !okOne || !okTwo || !okThree {
			return nil, errors.Wrapf(
				err, "Cannot parse json for middleware config json: %s",
				config,
			)
		}

		specs[idx], err = NewMiddlewareSpec(
			name,
			importPath,
			schema,
			configDirName,
		)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot validate middleware: %s",
				mid,
			)
		}
	}
	return specs, nil
}

// GatewaySpec collects information for the entire gateway
type GatewaySpec struct {
	// package helper for gateway
	PackageHelper *PackageHelper
	// tempalte instance for gateway
	Template *Template

	ClientModules     map[string]*ClientSpec
	EndpointModules   map[string]*EndpointSpec
	MiddlewareModules map[string]*MiddlewareSpec

	gatewayName         string
	configDirName       string
	clientConfigDir     string
	endpointConfigDir   string
	middlewareConfig    string
	copyrightHeaderFile string
}

// NewGatewaySpec sets up gateway spec
func NewGatewaySpec(
	packageHelper *PackageHelper,
	configDirName string,
	clientConfig string,
	endpointConfig string,
	middlewareConfig string,
	gatewayName string,
) (*GatewaySpec, error) {
	tmpl, err := NewTemplate()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create template")
	}

	middleConfig := filepath.Join(
		configDirName,
		middlewareConfig,
	)

	clientJsons, err := filepath.Glob(filepath.Join(
		configDirName,
		clientConfig,
		"*",
		"client-config.json",
	))
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load client json files")
	}

	endpointGroupJsons, err := filepath.Glob(filepath.Join(
		configDirName,
		endpointConfig,
		"*",
		"endpoint-config.json",
	))
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load endpoint json files")
	}

	endpointJsons, err := parseEndpointJsons(endpointGroupJsons)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot parse endpoint config")
	}

	spec := &GatewaySpec{
		PackageHelper:     packageHelper,
		Template:          tmpl,
		ClientModules:     map[string]*ClientSpec{},
		EndpointModules:   map[string]*EndpointSpec{},
		MiddlewareModules: map[string]*MiddlewareSpec{},

		configDirName:     configDirName,
		clientConfigDir:   clientConfig,
		endpointConfigDir: endpointConfig,
		gatewayName:       gatewayName,
	}

	middlewares, err := parseMiddlewareConfig(middleConfig, configDirName)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Cannot load middlewares:")
	}
	for _, mspec := range middlewares {
		spec.MiddlewareModules[mspec.Name] = mspec
	}

	clientSpecs := make([]*ClientSpec, len(clientJsons))
	for i, json := range clientJsons {
		cspec, err := NewClientSpec(json, packageHelper)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot parse client json file %s :", json,
			)
		}

		clientSpecs[i] = cspec
		spec.ClientModules[cspec.JSONFile] = cspec
	}
	for _, json := range endpointJsons {
		espec, err := NewEndpointSpec(json, packageHelper, spec.MiddlewareModules)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot parse endpoint json file %s :", json,
			)
		}

		err = espec.SetDownstream(clientSpecs, packageHelper)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot parse downstream info for endpoint: %s", json,
			)
		}
		spec.EndpointModules[espec.JSONFile] = espec
	}

	return spec, nil
}

// GenerateEndpointRegisterFile will generate endpoints registration for the gateway
func (gateway *GatewaySpec) GenerateEndpointRegisterFile() error {
	_, err := gateway.Template.GenerateEndpointRegisterFile(
		gateway.EndpointModules, gateway.PackageHelper,
	)
	return err
}
