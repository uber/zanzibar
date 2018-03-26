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
	"fmt"
	"io/ioutil"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/runtime"
	"go.uber.org/thriftrw/compile"
)

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)
	return zanzibar.GetDirnameFromRuntimeCaller(file)
}

const (
	reqHeaders = "reqHeaderMap"
	resHeaders = "resHeaderMap"
)

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
	// The globally unique package alias for the import
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
	// SidecarRouter indicates the client uses the given sidecar router to
	// to communicate with downstream service, it's not relevant to custom clients.
	SidecarRouter string
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
	Dependencies Dependencies           `json:"dependencies"`
}

// Dependencies lists all dependencies of a module
type Dependencies struct {
	Client []string `json:"client"`
	//	Service []string `json:"service"`  // example extension
}

// MiddlewareConfigConfig is the inner config object as prescribed by module_system json conventions
type MiddlewareConfigConfig struct {
	OptionsSchemaFile string `json:"schema"`
	ImportPath        string `json:"path"`
}

// MiddlewareConfig represents configuration for a middleware as is written in the json file
type MiddlewareConfig struct {
	Name         string                  `json:"name"`
	Dependencies *Dependencies           `json:"dependencies,omitempty"`
	Config       *MiddlewareConfigConfig `json:"config"`
}

// Validate the config spec attributes
func (mid *MiddlewareConfig) Validate(configDirName string) error {
	if mid.Name == "" {
		return errors.New("middleware config had empty name")
	}

	if mid.Config.ImportPath == "" {
		return errors.New("middleware config had empty import path")
	}

	if mid.Config.OptionsSchemaFile == "" {
		return errors.New("middleware config had empty schema")
	}

	schPath := filepath.Join(
		configDirName,
		mid.Config.OptionsSchemaFile,
	)

	bytes, err := ioutil.ReadFile(schPath)
	if err != nil {
		return errors.Wrapf(
			err, "Cannot read middleware json schema: %s",
			schPath,
		)
	}

	var midOptSchema map[string]interface{}
	err = json.Unmarshal(bytes, &midOptSchema)
	if err != nil {
		return errors.Wrapf(
			err, "Cannot parse json schema for middleware options: %s",
			schPath,
		)
	}
	return nil
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

// NewHTTPClientSpec creates a client spec from a http client module instance
func NewHTTPClientSpec(
	instance *ModuleInstance,
	clientConfig *ClientClassConfig,
	h *PackageHelper,
) (*ClientSpec, error) {
	return newClientSpec(instance, clientConfig, true, h)
}

// NewTChannelClientSpec creates a client spec from a json file whose type is tchannel
func NewTChannelClientSpec(
	instance *ModuleInstance,
	clientConfig *ClientClassConfig,
	h *PackageHelper,
) (*ClientSpec, error) {
	return newClientSpec(instance, clientConfig, false, h)
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
	mspec.PackageName = mspec.PackageName + "client"

	cspec := &ClientSpec{
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
	}

	sidecarRouter, ok := config["sidecarRouter"].(string)
	if ok {
		cspec.SidecarRouter = sidecarRouter
	}

	exposedMethods, ok := clientConfig.Config["exposedMethods"].(map[string]interface{})
	if !ok || len(exposedMethods) == 0 {
		return nil, errors.Errorf(
			"No methods are exposed in client config: %s",
			instance.JSONFileName,
		)
	}
	cspec.ExposedMethods = make(map[string]string, len(exposedMethods))
	reversed := make(map[string]string, len(exposedMethods))
	for key, val := range exposedMethods {
		v := val.(string)
		cspec.ExposedMethods[key] = v
		if _, ok := reversed[v]; ok {
			return nil, errors.Errorf(
				"value %q of the exposedMethods is not unique: %s",
				v,
				instance.JSONFileName,
			)
		}
		reversed[v] = key
	}

	return cspec, nil
}

// MiddlewareSpec holds information about each middleware at the endpoint
type MiddlewareSpec struct {
	// The middleware package name.
	Name string
	// Middleware specific configuration options.
	Options map[string]interface{}
	// Options pretty printed for template initialization
	PrettyOptions map[string]string
	// Module Dependencies,  clients etc.
	Dependencies *Dependencies
	// Go Import Path for MiddlewareHandle implementation
	ImportPath string
	// Location of JSON Schema file for the configured endpoint options
	OptionsSchemaFile string
}

func newMiddlewareSpec(cfg *MiddlewareConfig) *MiddlewareSpec {
	return &MiddlewareSpec{
		Name:              cfg.Name,
		Dependencies:      cfg.Dependencies,
		ImportPath:        cfg.Config.ImportPath,
		OptionsSchemaFile: cfg.Config.OptionsSchemaFile,
	}
}

// TypedHeader is typed header for headers resolved
// from header schema
type TypedHeader struct {
	Name        string
	TransformTo string
	Field       *compile.FieldSpec
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
	TestFixtures map[string]*EndpointTestFixture
	// Middlewares, meta data to add middlewares,
	Middlewares []MiddlewareSpec
	// HeadersPropagate, a map from endpoint request headers to
	// client request fields.
	HeadersPropagate map[string]FieldMapperEntry
	// ReqTransforms, a map from client request fields to endpoint
	// request fields that should override their values.
	ReqTransforms map[string]FieldMapperEntry
	// RespTransforms, a map from endpoint response fields to client
	// response fields that should override their values.
	RespTransforms map[string]FieldMapperEntry
	// ReqHeaders maps headers from server to client
	ReqHeaders map[string]*TypedHeader
	// ResHeaders maps headers from client to server
	ResHeaders map[string]*TypedHeader
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
		iclientID, ok := endpointConfigObj["clientId"]
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

func testFixtures(endpointConfigObj map[string]interface{}) (map[string]*EndpointTestFixture, error) {
	field, ok := endpointConfigObj["testFixtures"]
	if !ok {
		return nil, errors.Errorf("missing testFixtures field")
	}
	testFixturesRaw, err := json.Marshal(field)
	if err != nil {
		return nil, err
	}
	var ret map[string]*EndpointTestFixture
	err = json.Unmarshal(testFixturesRaw, &ret)
	return ret, err
}

func loadHeadersFromConfig(endpointCfgObj map[string]interface{}, key string) (map[string]string, error) {
	// TODO define endpointConfigObj to avoid type assertion

	headers, ok := endpointCfgObj[key]
	if !ok {
		return nil, errors.Errorf("unable to parse %q", key)
	}
	headersMap := make(map[string]string)
	for key, val := range headers.(map[string]interface{}) {
		switch value := val.(type) {
		case string:
			headersMap[textproto.CanonicalMIMEHeaderKey(key)] = textproto.CanonicalMIMEHeaderKey(value)
		default:
			return nil, errors.Errorf(
				"unable to parse string %q in headers %q", value, headers)
		}
	}
	return headersMap, nil
}

func sortedHeaders(headerMap map[string]*TypedHeader, filterRequired bool) []string {
	var sortedArr = []string{}
	for k, v := range headerMap {
		if !filterRequired {
			sortedArr = append(sortedArr, k)
		} else if v.Field.Required {
			sortedArr = append(sortedArr, k)
		}
	}
	sort.Strings(sortedArr)
	return sortedArr
}

func resolveHeaders(
	espec *EndpointSpec,
	endpointConfigObj map[string]interface{},
	key string,
) error {
	var (
		keyMap = map[string]string{
			reqHeaders: "http.req.metadata",
			resHeaders: "http.res.metadata",
		}
		headersMap    = make(map[string]*TypedHeader)
		headerModels  []string
		annotationKey = keyMap[key]
	)
	defer func() {
		if key == reqHeaders {
			espec.ReqHeaders = headersMap
		} else {
			espec.ResHeaders = headersMap
		}
	}()
	transformMap, err := loadHeadersFromConfig(endpointConfigObj, key)
	if err != nil {
		return err
	}
	method, err := findMethodByName(espec.ThriftMethodName, espec.ModuleSpec.Services)
	if err != nil {
		return err
	}
	for ak, av := range method.CompiledThriftSpec.Annotations {
		if strings.HasSuffix(ak, annotationKey) {
			headerModels = strings.Split(av, ",")
			break
		}
	}
	if len(headerModels) < 1 && len(transformMap) > 0 {
		return errors.Errorf("header models %q unconfigured for transform", key)
	}
	if len(headerModels) < 1 {
		return nil
	}
	for _, m := range headerModels {
		typedHeaders, err := resolveHeaderModels(espec.ModuleSpec, m)
		if err != nil {
			return err
		}
		for hk, hv := range typedHeaders {
			headersMap[textproto.CanonicalMIMEHeaderKey(hk)] = hv
		}
	}
	// apply header transform
	for k, v := range transformMap {
		typedHeader, ok := headersMap[k]
		if !ok {
			return errors.Errorf("unable to find header %q to transform", k)
		}
		typedHeader.TransformTo = v
	}
	return nil
}

func resolveHeaderModels(ms *ModuleSpec, modelPath string) (map[string]*TypedHeader, error) {
	const (
		headerPreix   = "headers"
		httpRefSuffix = "http.ref"
	)
	loadModuleFromInclude := func(moduleName string) *compile.Module {
		for pkgKey, pkg := range ms.CompiledModule.Includes {
			if pkgKey == moduleName {
				return pkg.Module
			}
		}
		return nil
	}
	loadHeaderKeyFromField := func(field *compile.FieldSpec) *string {
		for ak, av := range field.Annotations {
			if avs := strings.Split(av, "."); len(avs) == 2 &&
				avs[0] == headerPreix &&
				strings.HasSuffix(ak, httpRefSuffix) {
				headerKey := avs[1]
				return &headerKey
			}
		}
		return nil
	}
	loadHeadersFromCompiledModule := func(module *compile.Module, structName string) (map[string]*TypedHeader, error) {
		var typeStruct *compile.StructSpec
		var typedHeaders = make(map[string]*TypedHeader)
		for tk, tv := range module.Types {
			if ts, ok := tv.(*compile.StructSpec); tk == structName && ok {
				typeStruct = ts
				break
			}
		}
		if typeStruct == nil {
			return nil, errors.Errorf("unable to find typedHeaders %q", structName)
		}
		for _, field := range typeStruct.Fields {
			hk := loadHeaderKeyFromField(field)
			if hk == nil {
				return nil, errors.Errorf("unable to find header key %q", field.Name)
			}
			headerKey := textproto.CanonicalMIMEHeaderKey(*hk)
			typedHeaders[headerKey] = &TypedHeader{
				Name:        headerKey,
				TransformTo: headerKey,
				Field:       field,
			}
		}
		return typedHeaders, nil
	}
	switch paths := strings.Split(modelPath, "."); len(paths) {
	case 2:
		moduleName := paths[0]
		structName := paths[1]
		module := loadModuleFromInclude(moduleName)
		if module == nil {
			return nil, errors.Errorf("missing module spec %q", moduleName)
		}
		return loadHeadersFromCompiledModule(module, structName)
	default:
		// TODO case 1:
		// default header schema path to .
		return nil, errors.Errorf(
			"malformed header model %q, expecting <module>.<struct>", modelPath)
	}
}

func augmentHTTPEndpointSpec(
	espec *EndpointSpec,
	endpointConfigObj map[string]interface{},
	midSpecs map[string]*MiddlewareSpec,
) (*EndpointSpec, error) {
	testFixtures, err := testFixtures(endpointConfigObj)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse test cases")
	}
	espec.TestFixtures = testFixtures

	endpointMids, ok := endpointConfigObj["middlewares"].([]interface{})
	if !ok {
		return nil, errors.Errorf(
			"Unable to parse middlewares field",
		)
	}
	var middlewares []MiddlewareSpec
	for _, middleware := range endpointMids {
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
		// req/res transform middleware set type converter
		if name == "transformRequest" {
			reqTransforms, err := setTransformMiddleware(middlewareObj)
			if err != nil {
				return nil, err
			}
			espec.ReqTransforms = reqTransforms
			continue
		}
		if name == "transformResponse" {
			resTransforms, err := setTransformMiddleware(middlewareObj)
			if err != nil {
				return nil, err
			}
			espec.RespTransforms = resTransforms
			continue
		}
		// req header propagate middleware set headersPropagator
		if name == "headersPropagate" {
			headersPropagate, err := setPropagateMiddleware(middlewareObj)
			if err != nil {
				return nil, err
			}
			espec.HeadersPropagate = headersPropagate
			continue
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

		prettyOpts := map[string]string{}
		for key, value := range opts {
			rValue := reflect.ValueOf(value)
			kind := rValue.Kind()

			if kind == reflect.Slice && rValue.Len() > 0 {
				rType := rValue.Type()
				rElemType := rType.Elem()
				elemTypeString := rElemType.String()
				if rElemType.Kind() == reflect.Interface {
					rFirstValue := rValue.Index(0)
					rRawFirstValue := rFirstValue.Interface()

					elemTypeString = reflect.TypeOf(rRawFirstValue).String()
				}

				str := fmt.Sprintf("[]%s{", elemTypeString)
				for i := 0; i < rValue.Len(); i++ {
					str += fmt.Sprintf("%#v", rValue.Index(i))
					if i != rValue.Len()-1 {
						str += ","
					}
				}
				str += "}"
				prettyOpts[key] = str
			} else {
				prettyOpts[key] = fmt.Sprintf("%#v", rValue)
			}
		}

		middlewares = append(middlewares, MiddlewareSpec{
			Name:          name,
			ImportPath:    midSpecs[name].ImportPath,
			Options:       opts,
			PrettyOptions: prettyOpts,
		})
	}

	espec.Middlewares = middlewares

	// augment request headers
	if err := resolveHeaders(espec, endpointConfigObj, reqHeaders); err != nil {
		return nil, err
	}

	// augment response headers
	if err := resolveHeaders(espec, endpointConfigObj, resHeaders); err != nil {
		return nil, err
	}

	return espec, nil
}

func setPropagateMiddleware(middlewareObj map[string]interface{}) (map[string]FieldMapperEntry, error) {
	fieldMap := make(map[string]FieldMapperEntry)
	opts, ok := middlewareObj["options"].(map[string]interface{})
	if !ok {
		return nil, errors.New(
			"missing or invalid options for propagate middleware",
		)
	}
	propagates := opts["propagate"].([]interface{})
	dest := make(map[string]string)
	for _, propagate := range propagates {
		propagateMap := propagate.(map[string]interface{})
		fromField, ok := propagateMap["from"].(string)
		if !ok {
			return nil, errors.New(
				"propagate middleware found with no source field",
			)
		}
		toField, ok := propagateMap["to"].(string)
		if !ok {
			return nil, errors.New(
				"propagate middleware found with no destination field",
			)
		}
		if _, ok := dest[toField]; ok {
			return nil, errors.Errorf(
				"propagate multiple source field to destination field %s",
				toField,
			)
		}
		dest[toField] = toField
		fieldMap[toField] = FieldMapperEntry{
			QualifiedName: fromField,
		}
	}
	return fieldMap, nil
}

func setTransformMiddleware(middlewareObj map[string]interface{}) (map[string]FieldMapperEntry, error) {
	fieldMap := make(map[string]FieldMapperEntry)
	opts, ok := middlewareObj["options"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf(
			"transform middleware found with no options.",
		)
	}
	transforms := opts["transforms"].([]interface{})
	for _, transform := range transforms {
		transformMap := transform.(map[string]interface{})
		fromField, ok := transformMap["from"].(string)
		if !ok {
			return nil, errors.New(
				"transform middleware found with no source field",
			)
		}
		toField, ok := transformMap["to"].(string)
		if !ok {
			return nil, errors.New(
				"transform middleware found with no destination field",
			)
		}
		overrideOpt, ok := transformMap["override"].(bool)
		if ok {
			fieldMap[toField] = FieldMapperEntry{
				QualifiedName: fromField,
				Override:      overrideOpt,
			}
		} else {
			fieldMap[toField] = FieldMapperEntry{
				QualifiedName: fromField,
			}
		}
	}
	return fieldMap, nil
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

	return e.ModuleSpec.SetDownstream(e, h)
}

// EndpointClassConfig represents the specific config for
// an endpoint group. This is a downcast of the moduleClassConfig.
type EndpointClassConfig struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Config struct {
		Ratelimit int32    `json:"rateLimit"`
		Endpoints []string `json:"endpoints"`
	} `json:"config"`
	Dependencies map[string][]string `json:"dependencies"`
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
	middlewareConfigDir string,
	configDirName string,
) (map[string]*MiddlewareSpec, error) {
	specMap := map[string]*MiddlewareSpec{}
	if middlewareConfigDir == "" {
		return specMap, nil
	}
	fullMiddlewareDir := filepath.Join(configDirName, middlewareConfigDir)

	files, err := ioutil.ReadDir(fullMiddlewareDir)

	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error reading middleware config directory %q",
			fullMiddlewareDir,
		)
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		instanceConfig := filepath.Join(fullMiddlewareDir, file.Name(), "middleware-config.json")
		bytes, err := ioutil.ReadFile(instanceConfig)
		if os.IsNotExist(err) {
			fmt.Printf("Could not read config file for middleware directory \"%s\" skipping...\n", file.Name())
			continue
		} else if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot read middleware config json: %s",
				instanceConfig,
			)
		}
		var mid MiddlewareConfig
		err = json.Unmarshal(bytes, &mid)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot parse json for middleware config json: %s",
				instanceConfig,
			)
		}
		err = mid.Validate(configDirName)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot validate middleware: %s",
				mid,
			)
		}
		specMap[mid.Name] = newMiddlewareSpec(&mid)
	}
	return specMap, nil
}

// GatewaySpec collects information for the entire gateway
type GatewaySpec struct {
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
	endpointGroupJsons := []string{}
	err := filepath.Walk(
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

	// TODO verify that middlewares specified in Endpoint Specs are listed in the root EndpointConfig module dependencies
	//  this is an interactions between specs and dependencies that the module system does not enforce

	return spec, nil
}
