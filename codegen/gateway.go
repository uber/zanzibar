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
	"fmt"
	"io/ioutil"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	yaml "github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

const (
	reqHeaders         = "reqHeaderMap"
	resHeaders         = "resHeaderMap"
	customWorkflow     = "custom"
	clientlessWorkflow = "clientless"
)

var mandatoryEndpointFields = []string{
	"endpointType",
	"endpointId",
	"handleId",
	"thriftFile",
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
	JSONFile string // Deprecated
	// YAMLFile for this spec
	YAMLFile string
	// ClientType, currently "http", "tchannel" and "custom" are supported
	ClientType string
	// If "custom" then where to import custom code from
	CustomImportPath string
	// If "custom" then this is where the interface to mock comes from
	CustomInterface string
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

// ModuleClassConfig represents the generic YAML config for
// all modules. This will be provided by the module package.
type ModuleClassConfig struct {
	ClassConfigBase `yaml:",inline" json:",inline"`
	Config          interface{} `yaml:"config" json:"config"`
}

// Dependencies lists all dependencies of a module
type Dependencies struct {
	Client []string `yaml:"client" json:"client"`
	//	Service []string `yaml:"service"`  // example extension
}

// MiddlewareConfigConfig is the inner config object as prescribed by module_system yaml conventions
type MiddlewareConfigConfig struct {
	OptionsSchemaFile string `yaml:"schema" json:"schema"`
	ImportPath        string `yaml:"path" json:"path"`
}

// MiddlewareConfig represents configuration for a middleware as is written in the yaml file
type MiddlewareConfig struct {
	ClassConfigBase `yaml:",inline" json:",inline"`
	Dependencies    *Dependencies           `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Config          *MiddlewareConfigConfig `yaml:"config" json:"config"`
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
			err, "Cannot read middleware yaml schema: %s",
			schPath,
		)
	}

	var midOptSchema map[string]interface{}
	err = yaml.Unmarshal(bytes, &midOptSchema)
	if err != nil {
		return errors.Wrapf(
			err, "Cannot parse yaml schema for middleware options: %s",
			schPath,
		)
	}
	return nil
}

func getModuleConfigFileName(instance *ModuleInstance) string {
	if instance.YAMLFileName != "" {
		return instance.YAMLFileName
	}
	return instance.JSONFileName
}

// MiddlewareSpec holds information about each middleware at the endpoint
type MiddlewareSpec struct {
	// The middleware package name.
	Name string `yaml:"name"`
	// Middleware specific configuration options.
	Options map[string]interface{} `yaml:"options"`
	// Options pretty printed for template initialization
	PrettyOptions map[string]string
	// Module Dependencies,  clients etc.
	Dependencies *Dependencies
	// Go Import Path for MiddlewareHandle implementation
	ImportPath string
	// Location of yaml Schema file for the configured endpoint options
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
	ModuleSpec *ModuleSpec `yaml:"-"`
	// YAMLFile for this endpoint spec
	YAMLFile string `yaml:"-"`
	// GoStructsFileName is where structs are generated
	GoStructsFileName string `yaml:"-"`
	// GoFolderName is the folder where all the endpoints
	// are generated.
	GoFolderName string `yaml:"-"`
	// GoPackageName is the package import path.
	GoPackageName string `yaml:"-"`

	// EndpointType, currently only "http"
	EndpointType string `yaml:"endpointType" validate:"nonzero"`
	// EndpointID, used in metrics and logging, lower case.
	EndpointID string `yaml:"endpointId" validate:"nonzero"`
	// HandleID, used in metrics and logging, lowercase.
	HandleID string `yaml:"handleId" validate:"nonzero"`
	// ThriftFile, the thrift file for this endpoint
	ThriftFile string `yaml:"thriftFile" validate:"nonzero"`
	// ThriftFileSha, the SHA of the thrift file for this endpoint
	ThriftFileSha string `yaml:"thriftFileSha,omitempty"`
	// ThriftMethodName, which thrift method to use.
	ThriftMethodName string `yaml:"thriftMethodName" validate:"nonzero"`
	// ThriftServiceName, which thrift service to use.
	ThriftServiceName string `yaml:"-"`
	// TestFixtures, meta data to generate tests,
	TestFixtures map[string]*EndpointTestFixture `yaml:"testFixtures,omitempty"`
	// Middlewares, meta data to add middlewares,
	Middlewares []MiddlewareSpec `yaml:"middlewares,omitempty"`
	// HeadersPropagate, a map from endpoint request headers to
	// client request fields.
	HeadersPropagate map[string]FieldMapperEntry `yaml:"-"`
	// ReqTransforms, a map from client request fields to endpoint
	// request fields that should override their values.
	ReqTransforms map[string]FieldMapperEntry `yaml:"-"`
	// RespTransforms, a map from endpoint response fields to client
	// response fields that should override their values.
	RespTransforms map[string]FieldMapperEntry `yaml:"-"`
	// DummyReqTransforms is used to transform a clientless request to response mapping
	DummyReqTransforms map[string]FieldMapperEntry `yaml:"-"`
	// ErrTransforms is a map from endpoint exception fields to client exception fields
	// that should override their values
	// Note that this feature is not yet fully implemented in the stand-alone Zanzibar codebase
	ErrTransforms map[string]FieldMapperEntry `yaml:"-"`
	// ReqHeaders maps headers from server to client
	ReqHeaders map[string]*TypedHeader `yaml:"reqHeaderMap,omitempty"`
	// ResHeaders maps headers from client to server
	ResHeaders map[string]*TypedHeader `yaml:"resHeaderMap,omitempty"`
	// DefaultHeaders a slice of headers that are forwarded to downstream when available
	DefaultHeaders []string `yaml:"-"`
	// WorkflowType, either "httpClient" or "custom".
	// A httpClient workflow generates a http client Caller
	// A custom workflow just imports the custom code
	WorkflowType string `yaml:"workflowType" validate:"nonzero"`
	// If "custom" then where to import custom code from
	WorkflowImportPath string `yaml:"workflowImportPath"`
	// if "httpClient", which client to call.
	ClientID string `yaml:"clientId,omitempty"`
	// if "httpClient", which client method to call.
	ClientMethod string `yaml:"clientMethod,omitempty"`
	// The client for this endpoint if httpClient or tchannelClient
	ClientSpec *ClientSpec `yaml:"-"`
	// IsClientlessEndpoint checks if the endpoint is clientless
	IsClientlessEndpoint bool `yaml:"-"`
}

func ensureFields(config map[string]interface{}, mandatoryFields []string, yamlFile string) error {
	for i := 0; i < len(mandatoryFields); i++ {
		fieldName := mandatoryFields[i]
		if _, ok := config[fieldName]; !ok {
			return errors.Errorf(
				"config %q must have %q field", yamlFile, fieldName,
			)
		}
	}
	return nil
}

// NewEndpointSpec creates an endpoint spec from a yaml file.
func NewEndpointSpec(
	yamlFile string,
	h *PackageHelper,
	midSpecs map[string]*MiddlewareSpec,
) (*EndpointSpec, error) {
	_, err := os.Stat(yamlFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not find file %s: ", yamlFile)
	}

	bytes, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not read yaml file %s: ", yamlFile,
		)
	}

	endpointConfigObj := map[string]interface{}{}
	err = yaml.Unmarshal(bytes, &endpointConfigObj)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not parse yaml file: %s", yamlFile,
		)
	}

	if err := ensureFields(endpointConfigObj, mandatoryEndpointFields, yamlFile); err != nil {
		return nil, err
	}

	endpointType := endpointConfigObj["endpointType"]
	if endpointType == "http" {
		if err := ensureFields(endpointConfigObj, mandatoryHTTPEndpointFields, yamlFile); err != nil {
			return nil, err
		}

	}
	if endpointType != "http" && endpointType != "tchannel" {
		return nil, errors.Errorf(
			"Cannot support unknown endpointType for endpoint: %s", yamlFile,
		)
	}

	thriftFile := filepath.Join(
		h.IdlPath(), h.GetModuleIdlSubDir(true), endpointConfigObj["thriftFile"].(string),
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
	var isClientlessEndpoint bool

	workflowType := endpointConfigObj["workflowType"].(string)
	if workflowType == "httpClient" || workflowType == "tchannelClient" {
		iclientID, ok := endpointConfigObj["clientId"]
		if !ok {
			return nil, errors.Errorf(
				"endpoint config %q must have clientName field", yamlFile,
			)
		}
		if iclientID != nil {
			clientID = iclientID.(string)
		}
		iclientMethod, ok := endpointConfigObj["clientMethod"]
		if !ok {
			return nil, errors.Errorf(
				"endpoint config %q must have clientMethod field", yamlFile,
			)
		}
		if iclientMethod != nil {
			clientMethod = iclientMethod.(string)
		}
	} else if workflowType == customWorkflow {
		iworkflowImportPath, ok := endpointConfigObj["workflowImportPath"]
		if !ok {
			return nil, errors.Errorf(
				"endpoint config %q must have workflowImportPath field",
				yamlFile,
			)
		}
		workflowImportPath = iworkflowImportPath.(string)
	} else if workflowType == clientlessWorkflow {
		isClientlessEndpoint = true
	} else {
		return nil, errors.Errorf(
			"Invalid workflowType %q for endpoint %q",
			workflowType, yamlFile,
		)
	}

	dirName, err := filepath.Rel(h.ConfigRoot(), filepath.Dir(yamlFile))
	if err != nil {
		return nil, errors.Errorf("Config file is out of config root: %s", yamlFile)
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
			"Cannot read thriftMethodName %q for endpoint yaml file: %s",
			thriftInfo, yamlFile,
		)
	}

	espec := &EndpointSpec{
		ModuleSpec:           mspec,
		YAMLFile:             yamlFile,
		GoStructsFileName:    goStructsFileName,
		GoFolderName:         goFolderName,
		GoPackageName:        goPackageName,
		EndpointType:         endpointConfigObj["endpointType"].(string),
		EndpointID:           endpointConfigObj["endpointId"].(string),
		HandleID:             endpointConfigObj["handleId"].(string),
		ThriftFile:           thriftFile,
		ThriftServiceName:    parts[0],
		ThriftMethodName:     parts[1],
		WorkflowType:         workflowType,
		WorkflowImportPath:   workflowImportPath,
		IsClientlessEndpoint: isClientlessEndpoint,
		ClientID:             clientID,
		ClientMethod:         clientMethod,
		DefaultHeaders:       h.defaultHeaders,
	}

	defaultMidSpecs, err := getOrderedDefaultMiddlewareSpecs(
		h.ConfigRoot(),
		h.DefaultMiddlewareSpecs(),
		endpointType.(string))
	if err != nil {
		return nil, errors.Wrap(
			err, "error getting ordered default middleware specs",
		)
	}

	return augmentEndpointSpec(espec, endpointConfigObj, midSpecs, defaultMidSpecs)
}

func getOrderedDefaultMiddlewareSpecs(
	cfgDir string,
	middlewareSpecs map[string]*MiddlewareSpec,
	classType string,
) ([]MiddlewareSpec, error) {
	middlewareObj := map[string][]string{}

	middlewareOrderingFile := filepath.Join(cfgDir, "middlewares/default.yaml")
	if _, err := os.Stat(middlewareOrderingFile); os.IsNotExist(err) {
		// Cannot find yaml file, use json file instead
		middlewareOrderingFile = filepath.Join(cfgDir, "middlewares/default.json")
		if _, err := os.Stat(middlewareOrderingFile); os.IsNotExist(err) {
			// This file is not required so it is okay to skip
			return nil, nil
		}
	}

	bytes, err := ioutil.ReadFile(middlewareOrderingFile)
	if err != nil {
		return nil, errors.Wrapf(
			err, "could not read default middleware ordering file: %s", middlewareOrderingFile,
		)
	}
	err = yaml.Unmarshal(bytes, &middlewareObj)
	if err != nil {
		return nil, errors.Wrapf(
			err, "could not parse default middleware ordering file: %s", middlewareOrderingFile,
		)
	}
	middlewareOrderingObj := middlewareObj[classType]

	return sortByMiddlewareOrdering(middlewareOrderingObj, middlewareSpecs)
}

// sortByMiddlewareOrdering sorts middlewareSpecs using the ordering from middlewareOrderingObj
func sortByMiddlewareOrdering(
	middlewareOrderingObj []string,
	middlewareSpecs map[string]*MiddlewareSpec,
) ([]MiddlewareSpec, error) {
	middlewares := make([]MiddlewareSpec, 0)

	for _, middlewareName := range middlewareOrderingObj {
		middlewareSpec, ok := middlewareSpecs[middlewareName]
		if !ok {
			return nil, errors.Errorf("could not find middleware %s", middlewareName)
		}

		middlewares = append(middlewares, *middlewareSpec)
	}

	return middlewares, nil
}

func testFixtures(endpointConfigObj map[string]interface{}) (map[string]*EndpointTestFixture, error) {
	field, ok := endpointConfigObj["testFixtures"]
	if !ok {
		return nil, errors.Errorf("missing testFixtures field")
	}
	testFixturesRaw, err := yaml.Marshal(field)
	if err != nil {
		return nil, err
	}
	var ret map[string]*EndpointTestFixture
	err = yaml.Unmarshal(testFixturesRaw, &ret)
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
			headersMap[textproto.CanonicalMIMEHeaderKey(key)] = value
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

func augmentEndpointSpec(
	espec *EndpointSpec,
	endpointConfigObj map[string]interface{},
	midSpecs map[string]*MiddlewareSpec,
	defaultMidSpecs []MiddlewareSpec,
) (*EndpointSpec, error) {
	middlewares := defaultMidSpecs

	if _, ok := endpointConfigObj["middlewares"]; ok {
		endpointMids, ok := endpointConfigObj["middlewares"].([]interface{})
		if !ok {
			return nil, errors.Errorf(
				"Unable to parse middlewares field",
			)
		}

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
			if name == "transformClientlessReq" {
				dummyResTransforms, err := setTransformMiddleware(middlewareObj)
				if err != nil {
					return nil, err
				}
				espec.DummyReqTransforms = dummyResTransforms
				continue
			}
			if name == "transformError" {
				errTransforms, err := setTransformMiddleware(middlewareObj)
				if err != nil {
					return nil, err
				}
				espec.ErrTransforms = errTransforms
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
			for k, value := range opts {
				key := k
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
	}

	espec.Middlewares = middlewares

	if "http" == endpointConfigObj["endpointType"] {
		testFixtures, err := testFixtures(endpointConfigObj)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse test cases")
		}
		espec.TestFixtures = testFixtures

		// augment request headers
		if err := resolveHeaders(espec, endpointConfigObj, reqHeaders); err != nil {
			return nil, err
		}

		// augment response headers
		if err := resolveHeaders(espec, endpointConfigObj, resHeaders); err != nil {
			return nil, err
		}
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
	return strings.TrimSuffix(e.YAMLFile, filepath.Ext(e.YAMLFile)) + "_test.yaml"
}

// SetDownstream configures the downstream client for this endpoint spec
func (e *EndpointSpec) SetDownstream(
	clientModules []*ClientSpec,
	h *PackageHelper,
) error {
	if e.WorkflowType == customWorkflow {
		return nil
	}

	if e.WorkflowType == clientlessWorkflow {
		return e.ModuleSpec.SetDownstream(e, h)
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
			"When parsing endpoint yaml %q, "+
				"could not find client %q in gateway",
			e.YAMLFile, e.ClientID,
		)
	}

	e.ClientSpec = clientSpec

	return e.ModuleSpec.SetDownstream(e, h)
}

// EndpointConfig represent the "config" field of endpoint-config.yaml
type EndpointConfig struct {
	Ratelimit int32    `yaml:"rateLimit,omitempty" json:"rateLimit"`
	Endpoints []string `yaml:"endpoints" json:"endpoints"`
}

// EndpointClassConfig represents the specific config for
// an endpoint group. This is a downcast of the moduleClassConfig.
type EndpointClassConfig struct {
	ClassConfigBase `yaml:",inline" json:",inline"`
	Dependencies    map[string][]string `yaml:"dependencies" json:"dependencies"`
	Config          *EndpointConfig     `yaml:"config" json:"config" validate:"nonzero"`
}

func parseEndpointYamls(
	endpointGroupYamls []string,
) ([]string, error) {
	endpointYamls := []string{}

	for _, endpointGroupYAML := range endpointGroupYamls {
		bytes, err := ioutil.ReadFile(endpointGroupYAML)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot read endpoint group yaml: %s",
				endpointGroupYAML,
			)
		}

		var endpointConfig EndpointClassConfig
		err = yaml.Unmarshal(bytes, &endpointConfig)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot parse yaml for endpoint group config: %s",
				endpointGroupYAML,
			)
		}

		endpointConfigDir := filepath.Dir(endpointGroupYAML)
		for _, fileName := range endpointConfig.Config.Endpoints {
			endpointYamls = append(
				endpointYamls, filepath.Join(endpointConfigDir, fileName),
			)
		}
	}

	return endpointYamls, nil
}

func parseDefaultMiddlewareConfig(
	defaultMiddlewareConfigDir string,
	configDirName string,
) (map[string]*MiddlewareSpec, error) {
	fullMiddlewareDir := filepath.Join(configDirName, defaultMiddlewareConfigDir)
	_, err := ioutil.ReadDir(fullMiddlewareDir)
	if err != nil {
		return nil, nil
	}

	return parseMiddlewareConfig(defaultMiddlewareConfigDir, configDirName)
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

		instanceConfig := filepath.Join(
			fullMiddlewareDir, file.Name(), "middleware-config.yaml")
		if _, err := os.Stat(instanceConfig); os.IsNotExist(err) {
			// Cannot find yaml file, use json file instead
			instanceConfig = filepath.Join(
				fullMiddlewareDir, file.Name(), "middleware-config.json")
		}

		bytes, err := ioutil.ReadFile(instanceConfig)
		if os.IsNotExist(err) {
			fmt.Printf("Could not read config file for middleware directory \"%s\" skipping...\n", file.Name())
			continue
		} else if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot read middleware config yaml: %s",
				instanceConfig,
			)
		}
		var mid MiddlewareConfig
		err = yaml.Unmarshal(bytes, &mid)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot parse yaml for middleware config yaml: %s",
				instanceConfig,
			)
		}
		err = mid.Validate(configDirName)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot validate middleware: %v",
				mid,
			)
		}
		specMap[mid.Name] = newMiddlewareSpec(&mid)
	}
	return specMap, nil
}
