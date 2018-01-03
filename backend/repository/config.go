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

package repository

import (
	"os"
	"path/filepath"
	"strings"

	"encoding/json"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/codegen"
	zanzibar "github.com/uber/zanzibar/runtime"
)

// Config stores configuration for a gateway.
type Config struct {
	ID                   string
	Repository           string
	Team                 string
	Tier                 int
	ThriftRootDir        string
	PackageRoot          string
	ManagedThriftFolder  string
	GenCodePackage       string
	TargetGenDir         string
	ClientConfigDir      string
	EndpointConfigDir    string
	MiddlewareConfigFile string
	// Maps endpointID to configuration.
	Endpoints map[string]*EndpointConfig
	// Maps clientID to configuration.
	Clients        map[string]*ClientConfig
	ThriftServices ThriftServiceMap
	Middlewares    map[string]interface{}
	RawMiddlewares map[string]*codegen.MiddlewareConfig
}

// ThriftServiceMap maps thrift file -> service name -> *ThriftService
type ThriftServiceMap map[string]map[string]*ThriftService

// EndpointConfig stores configuration for an endpoint.
type EndpointConfig struct {
	ID               string                                  `json:"endpointId"`
	Type             ProtocolType                            `json:"endpointType"`
	HandleID         string                                  `json:"handleId"`
	ThriftFile       string                                  `json:"thriftFile"`
	ThriftFileSha    string                                  `json:"thriftFileSha,omitempty"`
	ThriftMethodName string                                  `json:"thriftMethodName"`
	WorkflowType     string                                  `json:"workflowType"`
	ClientID         string                                  `json:"clientId"`
	ClientMethod     string                                  `json:"clientMethod"`
	TestFixtures     map[string]*codegen.EndpointTestFixture `json:"testFixtures"`
	Middlewares      []*codegen.MiddlewareSpec               `json:"middlewares"`
	ReqHeaderMap     map[string]string                       `json:"reqHeaderMap"`
	ResHeaderMap     map[string]string                       `json:"resHeaderMap"`
}

// ClientConfig stores configuration for an client.
type ClientConfig struct {
	Name              string            `json:"name"`
	Type              ProtocolType      `json:"type"`
	ThriftFile        string            `json:"thriftFile,omitempty"`
	ServiceName       string            `json:"serviceName,omitempty"`
	ExposedMethods    map[string]string `json:"exposedMethods,omitempty"`
	IP                string            `json:"ip,omitempty"`
	Port              int64             `json:"port,omitempty"`
	Timeout           int64             `json:"clientTimeout,omitempty"`
	TimeoutPerAttempt int64             `json:"clientTimeoutPerAttempt,omitempty"`
	RoutingKey        string            `json:"routingKey"`
}

// ThriftService is a service defined in Thrift file.
type ThriftService struct {
	Name    string
	Path    string
	Methods []ThriftMethod
}

// ThriftMethod is a method defined in a Thrift Service.
type ThriftMethod struct {
	Name string
	Type ProtocolType
}

// ThriftMeta is the meta about a thrift file.
type ThriftMeta struct {
	// relative path under thrift root directory.
	Path string `json:"path"`
	// committed version
	Version string `json:"version"`
	// content of the thrift file
	Content string `json:"content,omitempty"`
}

// ProtocolType represents tranportation protocol type.
type ProtocolType string

const (
	gatewayConfigFile      = "gateway.json"
	productionCfgJSONPath  = "config/production.json"
	clientConfigFileName   = "client-config.json"
	clientModuleFileName   = "clients-config.json"
	endpointConfigFileName = "endpoint-config.json"
	serviceConfigFileName  = "service-config.json"
)

const (
	// HTTP type
	HTTP ProtocolType = "http"
	// TCHANNEL type
	TCHANNEL ProtocolType = "tchannel"
	// CUSTOM type
	CUSTOM ProtocolType = "custom"
	// UNKNOWN type
	UNKNOWN ProtocolType = "unknown"
)

// GatewayConfig returns the cached gateway configuration of the repository if
// the repository has not been updated for a certain time.
func (r *Repository) GatewayConfig() (*Config, error) {
	var (
		cfgVal *Config
		cfgErr error
	)
	if curCfgVal := r.gatewayCfgVal.Load(); curCfgVal != nil {
		cfgVal = curCfgVal.(*Config)
	}
	if curCfgErr := r.gatewayCfgErr.Load(); curCfgErr != nil {
		cfgErr = curCfgErr.(error)
	}

	if (cfgVal != nil || cfgErr != nil) && !r.Update() {
		return cfgVal, cfgErr
	}
	return r.LatestGatewayConfig()
}

// LatestGatewayConfig returns the configuration of current repository.
func (r *Repository) LatestGatewayConfig() (*Config, error) {
	// The newCfg won't be changed once created.
	newCfg, newCfgErr := r.newGatewayConfig()
	if newCfg != nil {
		r.gatewayCfgVal.Store(newCfg)
	}
	if newCfgErr != nil {
		r.gatewayCfgErr.Store(newCfgErr)
	}
	return newCfg, newCfgErr
}

// newGatewayConfig regenerates the configuration for a repository.
func (r *Repository) newGatewayConfig() (configuration *Config, cfgErr error) {
	defer func() {
		if p := recover(); p != nil {
			cfgErr = errors.Errorf(
				"panic when getting configuration for gateway %s: %+v",
				r.LocalDir(), p,
			)
		}
	}()
	configDir := r.absPath(r.LocalDir())
	cfg := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(filepath.Join(configDir, gatewayConfigFile)),
	}, nil)
	config := &Config{
		ID:                   cfg.MustGetString("gatewayName"),
		Repository:           r.remote,
		ThriftRootDir:        cfg.MustGetString("thriftRootDir"),
		PackageRoot:          cfg.MustGetString("packageRoot"),
		ManagedThriftFolder:  cfg.MustGetString("managedThriftFolder"),
		GenCodePackage:       cfg.MustGetString("genCodePackage"),
		TargetGenDir:         cfg.MustGetString("targetGenDir"),
		ClientConfigDir:      cfg.MustGetString("clientConfig"),
		EndpointConfigDir:    cfg.MustGetString("endpointConfig"),
		MiddlewareConfigFile: cfg.MustGetString("middlewareConfig"),
	}
	pkgHelper, err := codegen.NewPackageHelper(
		config.PackageRoot,
		config.ManagedThriftFolder,
		configDir,
		config.MiddlewareConfigFile,
		config.ThriftRootDir,
		config.GenCodePackage,
		config.TargetGenDir,
		"",
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create package helper")
	}

	moduleSystem, err := codegen.NewDefaultModuleSystem(pkgHelper)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create module system")
	}

	moduleInstances, err := moduleSystem.ResolveModules(
		pkgHelper.PackageRoot(),
		configDir,
		pkgHelper.CodeGenTargetPath(),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve module instances")
	}

	gatewaySpec, err := codegen.NewGatewaySpec(
		moduleInstances,
		pkgHelper,
		configDir,
		config.EndpointConfigDir,
		config.MiddlewareConfigFile,
		config.ID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read gateway spec")
	}

	if config.ThriftServices, err = r.thriftservices(config.ThriftRootDir, pkgHelper); err != nil {
		return nil, errors.Wrap(err, "failed to read thrift services")
	}

	if config.Clients, err = r.clientConfigs(config.ThriftRootDir, gatewaySpec); err != nil {
		return nil, errors.Wrap(err, "failed to read client configuration")
	}
	config.Endpoints = r.endpointConfigs(config.ThriftRootDir, gatewaySpec)

	if config.Middlewares, config.RawMiddlewares, err = r.middlewareConfigs(
		config.MiddlewareConfigFile,
	); err != nil {
		return nil, errors.Wrap(err, "fail to read middleware configuration")
	}
	return config, nil
}

func (r *Repository) clientConfigs(
	thriftRootDir string,
	gatewaySpec *codegen.GatewaySpec,
) (map[string]*ClientConfig, error) {
	cfgs := make(map[string]*ClientConfig, len(gatewaySpec.ClientModules))
	productionCfgJSON := map[string]interface{}{}
	path := r.absPath(productionCfgJSONPath)
	err := readJSONFile(path, &productionCfgJSON)
	if err != nil {
		return nil, err
	}
	for _, spec := range gatewaySpec.ClientModules {
		cfgs[spec.ClientID], err = clientConfig(spec, productionCfgJSON)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse config %q for client %q", path, spec.ClientID)
		}
		if cfgs[spec.ClientID].Type != CUSTOM {
			cfgs[spec.ClientID].ThriftFile = r.relativePath(thriftRootDir, spec.ThriftFile)
		}
	}
	return cfgs, nil
}

func clientConfig(spec *codegen.ClientSpec, productionCfgJSON map[string]interface{}) (*ClientConfig, error) {
	clientConfig := &ClientConfig{
		Name:           spec.ClientID,
		Type:           ProtocolTypeFromString(spec.ClientType),
		ExposedMethods: spec.ExposedMethods,
	}
	// TODO(zw): handle configurations of a customized client.
	if clientConfig.Type != HTTP && clientConfig.Type != TCHANNEL {
		return clientConfig, nil
	}

	var prefix string
	if spec.SidecarRouter != "" {
		prefix = "sidecarRouter." + spec.SidecarRouter + "." + string(clientConfig.Type) + "."
	} else {
		prefix = "clients." + spec.ClientID + "."
	}
	var err error
	clientConfig.IP, err = convStrVal(productionCfgJSON, prefix+"ip")
	if err != nil {
		return nil, err
	}
	clientConfig.Port, err = convInt64Val(productionCfgJSON, prefix+"port")
	if err != nil {
		return nil, err
	}
	if clientConfig.Type == HTTP {
		return clientConfig, nil
	}
	// tchannel related fields.
	prefix = "clients." + spec.ClientID + "."
	clientConfig.ServiceName, err = convStrVal(productionCfgJSON, prefix+"serviceName")
	if err != nil {
		return nil, err
	}
	clientConfig.Timeout, err = convInt64Val(productionCfgJSON, prefix+"timeout")
	if err != nil {
		return nil, err
	}
	clientConfig.TimeoutPerAttempt, err = convInt64Val(productionCfgJSON, prefix+"timeoutPerAttempt")
	if err != nil {
		return nil, err
	}
	return clientConfig, nil
}

func convStrVal(m map[string]interface{}, key string) (string, error) {
	if _, ok := m[key]; !ok {
		return "", errors.Errorf("key %q is not found", key)
	}
	val, ok := m[key].(string)
	if !ok {
		return "", errors.Errorf("key %q is not string, but %T", key, m[key])
	}
	return val, nil
}

func convInt64Val(m map[string]interface{}, key string) (int64, error) {
	if _, ok := m[key]; !ok {
		return 0, errors.Errorf("key %q is not found", key)
	}
	val, ok := m[key].(float64)
	if !ok {
		return 0, errors.Errorf("key %q is not int, but %T", key, m[key])
	}
	return int64(val), nil
}

func (r *Repository) endpointConfigs(thriftRootDir string, gatewaySpec *codegen.GatewaySpec) map[string]*EndpointConfig {
	cfgs := make(map[string]*EndpointConfig, len(gatewaySpec.EndpointModules))
	for _, spec := range gatewaySpec.EndpointModules {
		endpointID := spec.EndpointID + "." + spec.HandleID
		cfgs[endpointID] = &EndpointConfig{
			ID:               spec.EndpointID,
			Type:             ProtocolTypeFromString(spec.EndpointType),
			HandleID:         spec.HandleID,
			ThriftFile:       r.relativePath(thriftRootDir, spec.ThriftFile),
			ThriftMethodName: spec.ThriftServiceName + "::" + spec.ThriftMethodName,
			WorkflowType:     spec.WorkflowType,
			ClientID:         spec.ClientID,
			ClientMethod:     spec.ClientMethod,
			// TODO(zw): add test fixtures and middleware config.
		}
	}
	return cfgs
}

func (r *Repository) middlewareConfigs(
	middlewareConfigFile string,
) (map[string]interface{}, map[string]*codegen.MiddlewareConfig, error) {
	middlewares := make(map[string]*codegen.MiddlewareConfig)
	cfgs := make(map[string]interface{})
	var rawCfgs map[string][]*codegen.MiddlewareConfig

	if middlewareConfigFile == "" {
		return cfgs, middlewares, nil
	}
	bytes, err := r.ReadFile(middlewareConfigFile)
	if err != nil {
		return nil, nil, err
	}

	if len(bytes) == 0 {
		return nil, nil, errors.New("middlewares file is an empty file")
	}

	err = json.Unmarshal(bytes, &rawCfgs)
	if err != nil {
		return nil, nil, err
	}

	if _, ok := rawCfgs["middlewares"]; !ok {
		return nil, nil, errors.New(
			"middlewares config is missing root object property \"middlewares\"",
		)
	}

	for _, mc := range rawCfgs["middlewares"] {
		schemaFile := mc.SchemaFile
		bytes, err := r.ReadFile(schemaFile)
		if err != nil {
			return nil, nil, errors.Wrapf(
				err, "could not read schemaFile(%s)", schemaFile,
			)
		}

		rawData := map[string]interface{}{}
		err = json.Unmarshal(bytes, &rawData)
		if err != nil {
			return nil, nil, errors.Wrapf(
				err, "could not parse schemaFile(%s)", schemaFile,
			)
		}

		cfgs[mc.Name] = rawData
		middlewares[mc.Name] = mc
	}

	return cfgs, middlewares, nil
}

func (r *Repository) thriftservices(thriftRootDir string, packageHelper *codegen.PackageHelper) (ThriftServiceMap, error) {
	idlMap := make(ThriftServiceMap)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || filepath.Ext(path) != ".thrift" {
			return nil
		}
		if err != nil {
			return err
		}
		// Parse service module as tchannel service.
		mspec, err := codegen.NewModuleSpec(path, false, false, packageHelper)
		if err != nil {
			return errors.Wrapf(err, "failed to genenerate module spec for thrift %s", path)
		}
		serviceType := TCHANNEL
		// Parse HTTP annotations.
		if _, err := codegen.NewModuleSpec(path, true, false, packageHelper); err == nil {
			serviceType = HTTP
		}
		relativePath := r.relativePath(thriftRootDir, path)
		idlMap[relativePath] = map[string]*ThriftService{}
		for _, service := range mspec.Services {
			tservice := &ThriftService{
				Name: service.Name,
				Path: relativePath,
			}
			tservice.Methods = make([]ThriftMethod, len(service.Methods))
			for i, method := range service.Methods {
				tservice.Methods[i].Name = method.Name
				tservice.Methods[i].Type = serviceType
			}
			idlMap[relativePath][service.Name] = tservice
		}
		return nil
	}
	if err := filepath.Walk(r.absPath(thriftRootDir), walkFn); err != nil {
		return nil, errors.Wrapf(err, "failed to traverse IDL dir")
	}
	return idlMap, nil
}

func (r *Repository) absPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	path, err := filepath.Abs(filepath.Join(r.localDir, path))
	if err != nil {
		return err.Error()
	}
	return path
}

func (r *Repository) relativePath(rootDir string, filePath string) string {
	rootAbsDir := r.absPath(rootDir)
	fileAbsPath := r.absPath(filePath)
	relative := strings.TrimPrefix(fileAbsPath, rootAbsDir)
	return strings.TrimPrefix(relative, "/")
}

// ProtocolTypeFromString converts a string to ProtocolType.
func ProtocolTypeFromString(str string) ProtocolType {
	switch str {
	case "http":
		return HTTP
	case "tchannel":
		return TCHANNEL
	case "custom":
		return CUSTOM
	default:
		return UNKNOWN
	}
}
