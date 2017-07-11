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
	"thriftFile",
	"thriftFileSha",
}
var mandatoryCustomClientFields = []string{
	"customImportPath",
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
	// The path to the client package import
	ImportPackagePath string
	// The globally unique pacakge alias for the import
	ImportPackageAlias string
	// ExportName is the name that should be used when initializing the module
	// on a dependency struct.
	ExportName string
	// ExportType refers to the type returned by the module initializer
	ExportType string
	// ThriftFile, absolute path to thrift file
	ThriftFile string
	// ClientID, used for logging and metrics, must be lowercase
	// and use dashes.
	ClientID string
	// ClientName, PascalCase name of the client, the generated
	// `Clients` struct will contain a field of this name
	ClientName string
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
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Config       map[string]interface{} `json:"config"`
	Dependencies ClientDependencies     `json:"dependencies"`
}

// ClientDependencies lists all depedencies of a client.
type ClientDependencies struct {
	Client []string `json:"client"`
}

// NewClientSpec creates a client spec from a json file.
func NewClientSpec(
	instance *ModuleInstance,
	h *PackageHelper,
) (*ClientSpec, error) {
	clientConfig := &ClientClassConfig{}

	if err := json.Unmarshal(instance.JSONFileRaw, &clientConfig); err != nil {
		return nil, errors.Wrapf(
			err,
			"Could not parse class config json file: %s",
			instance.JSONFileName,
		)
	}

	switch clientConfig.Type {
	case "http":
		return NewHTTPClientSpec(instance, clientConfig, h)
	case "tchannel":
		return NewTChannelClientSpec(instance, clientConfig, h)
	case "custom":
		return NewCustomClientSpec(instance, clientConfig, h)
	default:
		return nil, errors.Errorf(
			"Cannot support unknown clientType for client %q",
			instance.JSONFileName,
		)
	}
}

// NewTChannelClientSpec creates a client spec from a json file whose type is tchannel
func NewTChannelClientSpec(
	instance *ModuleInstance,
	clientConfig *ClientClassConfig,
	h *PackageHelper,
) (*ClientSpec, error) {
	exposedMethods, ok := clientConfig.Config["exposedMethods"].(map[string]interface{})
	if !ok || len(exposedMethods) == 0 {
		return nil, errors.Errorf(
			"No methods are exposed in client config: %s",
			instance.JSONFileName,
		)
	}

	cspec, err := newClientSpec(instance, clientConfig, false, h)
	if err != nil {
		return nil, err
	}

	cspec.ExposedMethods = make(map[string]string, len(exposedMethods))
	reversed := make(map[string]string, len(exposedMethods))
	for key, val := range exposedMethods {
		cspec.ExposedMethods[key] = val.(string)
		if _, ok := reversed[val.(string)]; ok {
			return nil, errors.Errorf(
				"value %q of the exposedMethods is not unique: %s",
				val.(string),
				instance.JSONFileName,
			)
		}
		reversed[val.(string)] = key
	}

	return cspec, nil
}

// NewCustomClientSpec creates a client spec from a json file whose type is custom
func NewCustomClientSpec(
	instance *ModuleInstance,
	clientConfig *ClientClassConfig,
	h *PackageHelper,
) (*ClientSpec, error) {
	for _, f := range mandatoryCustomClientFields {
		if _, ok := clientConfig.Config[f]; !ok {
			return nil, errors.Errorf(
				"client config %q must have %q field for type custom",
				instance.JSONFileName,
				f,
			)
		}
	}

	clientSpec := &ClientSpec{
		JSONFile:           instance.JSONFileName,
		ImportPackagePath:  instance.PackageInfo.ImportPackagePath(),
		ImportPackageAlias: instance.PackageInfo.ImportPackageAlias(),
		ExportName:         instance.PackageInfo.ExportName,
		ExportType:         instance.PackageInfo.ExportType,
		ClientType:         clientConfig.Type,
		ClientID:           clientConfig.Name,
		ClientName:         instance.PackageInfo.QualifiedInstanceName,
		CustomImportPath:   clientConfig.Config["customImportPath"].(string),
	}

	return clientSpec, nil
}

// NewHTTPClientSpec creates a client spec from a http client module instance
func NewHTTPClientSpec(
	instance *ModuleInstance,
	clientConfig *ClientClassConfig,
	h *PackageHelper,
) (*ClientSpec, error) {
	exposedMethods, ok := clientConfig.Config["exposedMethods"].(map[string]interface{})
	if !ok || len(exposedMethods) == 0 {
		return nil, errors.Errorf(
			"No methods are exposed in client config: %s",
			instance.JSONFileName,
		)
	}

	cspec, err := newClientSpec(instance, clientConfig, true, h)
	if err != nil {
		return nil, err
	}

	cspec.ExposedMethods = make(map[string]string, len(exposedMethods))
	reversed := make(map[string]string, len(exposedMethods))
	for key, val := range exposedMethods {
		cspec.ExposedMethods[key] = val.(string)
		if _, ok := reversed[val.(string)]; ok {
			return nil, errors.Errorf(
				"value %q of the exposedMethods is not unique: %s",
				val.(string),
				instance.JSONFileName,
			)
		}
		reversed[val.(string)] = key
	}

	return cspec, nil
}

func newClientSpec(
	instance *ModuleInstance,
	clientConfig *ClientClassConfig,
	wantAnnot bool, h *PackageHelper,
) (*ClientSpec, error) {
	config := clientConfig.Config

	for i := 0; i < len(mandatoryClientFields); i++ {
		fieldName := mandatoryClientFields[i]
		if _, ok := config[fieldName]; !ok {
			return nil, errors.Errorf(
				"client config %q must have %q field", instance.JSONFileName, fieldName,
			)
		}
	}

	thriftFile := filepath.Join(
		h.ThriftIDLPath(), config["thriftFile"].(string),
	)

	mspec, err := NewModuleSpec(thriftFile, wantAnnot, false, h)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not build module spec for thrift %s: ", thriftFile,
		)
	}
	mspec.PackageName = mspec.PackageName + "Client"

	return &ClientSpec{
		ModuleSpec:         mspec,
		JSONFile:           instance.JSONFileName,
		ClientType:         clientConfig.Type,
		ImportPackagePath:  instance.PackageInfo.ImportPackagePath(),
		ImportPackageAlias: instance.PackageInfo.ImportPackageAlias(),
		ExportName:         instance.PackageInfo.ExportName,
		ExportType:         instance.PackageInfo.ExportType,
		ThriftFile:         thriftFile,
		ClientID:           instance.InstanceName,
		ClientName:         instance.PackageInfo.QualifiedInstanceName,
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
	// Keeps keys in a sorted array so that golden files have
	// deterministic orderings
	ReqHeaderMap     map[string]string
	ReqHeaderMapKeys []string
	// ResHeaderMap, maps headers from client to server.
	// Keeps keys in a sorted array so that golden files have
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
	ClientID string
	// if "httpClient", which client method to call.
	ClientMethod string
	// The client for this endpoint if httpClient or tchannelClient
	ClientSpec *ClientSpec
}

func ensureFields(config map[string]interface{}, mandatoryFields []string, jsonFile string) error {
	for i := 0; i < len(mandatoryFields); i++ {
		fieldName := mandatoryFields[i]
		if _, ok := config[fieldName]; !ok {
			return errors.Errorf(
				"config %q must have %q field", jsonFile, fieldName,
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
			err, "Could not parse json file: %s", jsonFile,
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
			"Cannot support unknown endpointType for endpoint: %s", jsonFile,
		)
	}

	thriftFile := filepath.Join(
		h.ThriftIDLPath(), endpointConfigObj["thriftFile"].(string),
	)

	mspec, err := NewModuleSpec(thriftFile, endpointType == "http", true, h)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not build module spec for thrift: %s", thriftFile,
		)
	}

	var workflowImportPath string
	var clientID string
	var clientMethod string

	workflowType := endpointConfigObj["workflowType"].(string)
	if workflowType == "httpClient" || workflowType == "tchannelClient" {
		iclientID, ok := endpointConfigObj["clientID"]
		if !ok {
			return nil, errors.Errorf(
				"endpoint config %q must have clientName field", jsonFile,
			)
		}
		clientID = iclientID.(string)

		iclientMethod, ok := endpointConfigObj["clientMethod"]
		if !ok {
			return nil, errors.Errorf(
				"endpoint config %q must have clientMethod field", jsonFile,
			)
		}
		clientMethod = iclientMethod.(string)
	} else if workflowType == "custom" {
		iworkflowImportPath, ok := endpointConfigObj["workflowImportPath"]
		if !ok {
			return nil, errors.Errorf(
				"endpoint config %q must have workflowImportPath field",
				jsonFile,
			)
		}
		workflowImportPath = iworkflowImportPath.(string)
	} else {
		return nil, errors.Errorf(
			"Invalid workflowType %q for endpoint %q",
			workflowType, jsonFile,
		)
	}

	dirName, err := filepath.Rel(h.ConfigRoot(), filepath.Dir(jsonFile))
	if err != nil {
		return nil, errors.Errorf("Config file is out of config root: %s", jsonFile)
	}

	goFolderName := filepath.Join(h.CodeGenTargetPath(), dirName)

	goStructsFileName := filepath.Join(
		h.CodeGenTargetPath(),
		dirName,
		filepath.Base(dirName)+"_structs.go",
	)

	goPackageName := filepath.Join(h.GoGatewayPackageName(), dirName)

	thriftInfo := endpointConfigObj["thriftMethodName"].(string)
	parts := strings.Split(thriftInfo, "::")
	if len(parts) != 2 {
		return nil, errors.Errorf(
			"Cannot read thriftMethodName %q for endpoint json file: %s",
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
		ClientID:           clientID,
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
				"middlewares config %q not found.", name,
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
		if v.ClientID == e.ClientID {
			clientSpec = v
			break
		}
	}

	if clientSpec == nil {
		return errors.Errorf(
			"When parsing endpoint json %q, "+
				"could not find client %q in gateway",
			e.JSONFile, e.ClientID,
		)
	}

	e.ClientSpec = clientSpec

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
	middlewareConfig string,
	configDirName string,
) (map[string]*MiddlewareSpec, error) {
	specMap := map[string]*MiddlewareSpec{}
	if middlewareConfig == "" {
		return specMap, nil
	}
	config := filepath.Join(configDirName, middlewareConfig)
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

	for _, mid := range midList {
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

		specMap[name], err = NewMiddlewareSpec(
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
	return specMap, nil
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
	endpointConfigDir   string
	middlewareConfig    string
	copyrightHeaderFile string
}

// NewGatewaySpec sets up gateway spec
func NewGatewaySpec(
	moduleInstances map[string][]*ModuleInstance,
	packageHelper *PackageHelper,
	configDirName string,
	endpointConfig string,
	middlewareConfig string,
	gatewayName string,
) (*GatewaySpec, error) {
	tmpl, err := NewTemplate()
	if err != nil {
		return nil, errors.Wrap(err, "cannot create template")
	}

	endpointGroupJsons := []string{}
	err = filepath.Walk(
		filepath.Join(configDirName, endpointConfig),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == "endpoint-config.json" {
				endpointGroupJsons = append(endpointGroupJsons, path)
			}
			return nil
		},
	)

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
		endpointConfigDir: endpointConfig,
		gatewayName:       gatewayName,
	}

	spec.MiddlewareModules, err = parseMiddlewareConfig(middlewareConfig, configDirName)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Cannot load middlewares:")
	}

	clientModules := moduleInstances["client"]
	clientSpecs := make([]*ClientSpec, len(clientModules))
	for i, clientInstance := range clientModules {
		cspec, err := NewClientSpec(clientInstance, packageHelper)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"Cannot create spec for client module: %s",
				clientInstance.InstanceName,
			)
		}

		clientSpecs[i] = cspec
		spec.ClientModules[cspec.ClientID] = cspec
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
		spec.EndpointModules[espec.EndpointID+"::"+espec.HandleID] = espec
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
