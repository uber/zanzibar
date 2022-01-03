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
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/thriftrw/compile"
)

const (
	thriftExtension       = ".thrift"
	protoExtension        = ".proto"
	endpointModuleName    = "endpoints"
	defaultModuleFallback = "default"
)

// PackageHelper manages the mapping from idl file to generated type code and service code.
type PackageHelper struct {
	// The project package name
	packageRoot string
	// The absolute root dir path for all configs, i.e., clients, endpoints, idl, etc
	configRoot string
	// The absolute root directory containing idl files
	idlRootDir string
	// moduleIdlSubDir defines subdir for idl per module
	moduleIdlSubDir map[string]string
	// The map of idl type to go package name of where the generated structs are
	// map is keyed with idl file extension, e.g., .thrift, .proto
	genCodePackage map[string]string
	// The absolute directory to put the generated service code
	targetGenDir string
	// The go package name where all the generated code is
	goGatewayNamespace string
	// The root directory for the gateway test config files
	testConfigsRootDir string
	// String containing copyright header to add to generated code.
	copyrightHeader string
	// annotation prefix to parse for thrift schema
	annotationPrefix string
	// The middlewares available for the endpoints
	middlewareSpecs map[string]*MiddlewareSpec
	// The default middlewares for all endpoints
	defaultMiddlewareSpecs map[string]*MiddlewareSpec
	// The default headers to forward with all requests when present
	defaultHeaders []string
	// Use deputy client when this header is set
	deputyReqHeader string
	// traceKey is the key for unique trace id that identifies request / response pair
	traceKey string
	// moduleSearchPaths is a dictionary of glob patterns for folders that contain *-config.yaml files
	moduleSearchPaths map[string][]string
	// defaultDependencies is a dictionary of glob patterns for default dependencies
	defaultDependencies map[string][]string
}

// NewDefaultPackageHelperOptions returns a new default PackageHelperOptions, all optional fields are set as default.
func NewDefaultPackageHelperOptions() *PackageHelperOptions {
	return &PackageHelperOptions{}
}

// PackageHelperOptions set the optional configurations for the project package.
// Each field is optional, if not set, it is set to the corresponding default value.
type PackageHelperOptions struct {
	// relative path to the idl dir, defaults to "./idl"
	RelIdlRootDir string
	// subdir for idl per module
	ModuleIdlSubDir map[string]string
	// relative path to the target dir that will contain generated code, defaults to "./build"
	RelTargetGenDir string
	// relative path to the middleware config dir, defaults to ""
	RelMiddlewareConfigDir string
	// relative path to the default middleware config dir, defaults to ""
	RelDefaultMiddlewareConfigDir string

	// map of idl type to package path of the generated code
	// map should be keyed with idl file extension, e.g., .thrift, .proto
	GenCodePackage map[string]string

	// thrift http annotation prefix, defaults to "zanzibar"
	AnnotationPrefix string

	// copyright header in generated code, defaults to ""
	CopyrightHeader string
	// header key to redirect client requests to staging environment, defaults to "X-Zanzibar-Use-Staging"
	StagingReqHeader string
	// The default headers to forward with all requests when present
	DefaultHeaders []string
	// header key to redirect client requests to local environment, defaults to "x-deputy-forwarded"
	DeputyReqHeader string
	// header key to uniquely identifies request/response pair, defaults to "x-trace-id"
	TraceKey string
	// ModuleSearchPaths is a dictionary of glob patterns for folders that contain *-config.yaml files
	ModuleSearchPaths map[string][]string
	// DefaultDependencies is a dictionary of glob patterns for folders that contain default dependencies
	DefaultDependencies map[string][]string

	// key to read qps levels or not in the endpoint levels
	QPSLevelsEnabled bool

	// key to enable custom initialisation
	CustomInitialisationEnabled bool
}

func (p *PackageHelperOptions) relTargetGenDir() string {
	if p.RelTargetGenDir != "" {
		return p.RelTargetGenDir
	}
	return "./build"
}

func (p *PackageHelperOptions) relIdlRootDir() string {
	if p.RelIdlRootDir != "" {
		return p.RelIdlRootDir
	}
	return "./idl"
}

func (p *PackageHelperOptions) moduleIdlSubDir() map[string]string {
	return p.ModuleIdlSubDir
}

func (p *PackageHelperOptions) relMiddlewareConfigDir() string {
	return p.RelMiddlewareConfigDir
}

func (p *PackageHelperOptions) relDefaultMiddlewareConfigDir() string {
	return p.RelDefaultMiddlewareConfigDir
}

func (p *PackageHelperOptions) genCodePackage(packageRoot string) map[string]string {
	if p.GenCodePackage != nil {
		return p.GenCodePackage
	}

	defaultGenCodePath := path.Join(packageRoot, p.relTargetGenDir(), "gen-code")
	return map[string]string{
		"thrift": defaultGenCodePath,
		"proto":  defaultGenCodePath,
	}
}

func (p *PackageHelperOptions) annotationPrefix() string {
	if p.AnnotationPrefix != "" {
		return p.AnnotationPrefix
	}
	return "zanzibar"
}

func (p *PackageHelperOptions) copyrightHeader() string {
	if p.CopyrightHeader != "" {
		return p.CopyrightHeader
	}
	return ""
}

func (p *PackageHelperOptions) deputyReqHeader() string {
	if p.DeputyReqHeader != "" {
		return p.DeputyReqHeader
	}
	return "x-deputy-forwarded"
}

func (p *PackageHelperOptions) traceKey() string {
	if p.TraceKey != "" {
		return p.TraceKey
	}
	return "x-uber-id"
}

// NewPackageHelper creates a package helper.
// packageRoot is the project root package path, configRoot is the project config dir path.
// options can be nil, in which case all optional settings for the package are set to default values.
func NewPackageHelper(
	packageRoot string,
	configRoot string,
	options *PackageHelperOptions,
) (*PackageHelper, error) {
	absConfigRoot, err := filepath.Abs(configRoot)
	if err != nil {
		return nil, errors.Errorf(
			"%s is not valid path: %s", configRoot, err,
		)
	}

	if options == nil {
		options = NewDefaultPackageHelperOptions()
	}

	goGatewayNamespace := path.Join(packageRoot, options.relTargetGenDir())

	middlewareSpecs, err := parseMiddlewareConfig(options.relMiddlewareConfigDir(), absConfigRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load middlewares")
	}

	defaultMiddlewareSpecs, err := parseDefaultMiddlewareConfig(
		options.relDefaultMiddlewareConfigDir(),
		absConfigRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load default middlewares")
	}

	moduleIdlSubDir := options.moduleIdlSubDir()
	if len(moduleIdlSubDir) == 0 {
		moduleIdlSubDir = map[string]string{endpointModuleName: ".", defaultModuleFallback: "."}
	}

	p := &PackageHelper{
		packageRoot:            packageRoot,
		configRoot:             absConfigRoot,
		idlRootDir:             filepath.Join(absConfigRoot, options.relIdlRootDir()),
		genCodePackage:         options.genCodePackage(packageRoot),
		goGatewayNamespace:     goGatewayNamespace,
		targetGenDir:           filepath.Join(absConfigRoot, options.relTargetGenDir()),
		copyrightHeader:        options.copyrightHeader(),
		middlewareSpecs:        middlewareSpecs,
		defaultMiddlewareSpecs: defaultMiddlewareSpecs,
		annotationPrefix:       options.annotationPrefix(),
		deputyReqHeader:        options.deputyReqHeader(),
		traceKey:               options.traceKey(),
		moduleSearchPaths:      options.ModuleSearchPaths,
		defaultDependencies:    options.DefaultDependencies,
		defaultHeaders:         options.DefaultHeaders,
		moduleIdlSubDir:        moduleIdlSubDir,
	}
	return p, nil
}

// PackageRoot returns the service's root package name
func (p PackageHelper) PackageRoot() string {
	return p.packageRoot
}

// ConfigRoot returns the service's absolute config root path
func (p PackageHelper) ConfigRoot() string {
	return p.configRoot
}

// MiddlewareSpecs returns a map of middlewares available
func (p PackageHelper) MiddlewareSpecs() map[string]*MiddlewareSpec {
	return p.middlewareSpecs
}

// DefaultMiddlewareSpecs returns a map of default middlewares
func (p PackageHelper) DefaultMiddlewareSpecs() map[string]*MiddlewareSpec {
	return p.defaultMiddlewareSpecs
}

// GenCodePackage returns map of idl type to the file path of the idl generated code folder
func (p PackageHelper) GenCodePackage() map[string]string {
	return p.genCodePackage
}

// TypeImportPath returns the Go import path for types defined in a idlFile file.
func (p PackageHelper) TypeImportPath(idlFile string) (string, error) {
	if !strings.HasSuffix(idlFile, thriftExtension) && !strings.HasSuffix(idlFile, protoExtension) {
		return "", errors.Errorf("idl file %s is not %s or %s", idlFile, thriftExtension, protoExtension)
	}
	var suffix string
	// for a filepath: a/b/c.(thrift|proto)
	// - thrift generates code in path: a/b/c/c.go
	// - proto generates code in path: a/b/c.go
	if strings.HasSuffix(idlFile, thriftExtension) {
		suffix = strings.TrimSuffix(idlFile, thriftExtension)
	} else {
		suffix = filepath.Dir(idlFile)
	}

	idx := strings.Index(idlFile, p.idlRootDir)
	if idx == -1 {
		return "", errors.Errorf(
			"file %s is not in IDL dir (%s)",
			idlFile, p.idlRootDir,
		)
	}

	ext := filepath.Ext(idlFile)
	genCodePkg, ok := p.genCodePackage[ext]
	if !ok {
		return "", errors.Errorf("genCodePackage for %q idl file is not configured in build.yaml", ext)
	}
	return path.Join(
		genCodePkg,
		idlFile[idx+len(p.idlRootDir):len(suffix)],
	), nil
}

// GoGatewayPackageName returns the name of the gateway package
func (p PackageHelper) GoGatewayPackageName() string {
	return p.goGatewayNamespace
}

// IdlPath returns the file path to the thrift idl folder
func (p PackageHelper) IdlPath() string {
	return p.idlRootDir
}

// CodeGenTargetPath returns the file path where the code should
// be generated.
func (p PackageHelper) CodeGenTargetPath() string {
	return p.targetGenDir
}

// TypePackageName returns the package name that defines the type.
func (p PackageHelper) TypePackageName(idlFile string) (string, error) {
	if !strings.HasSuffix(idlFile, thriftExtension) && !strings.HasSuffix(idlFile, protoExtension) {
		return "", errors.Errorf("file %s is not %s or %s", idlFile, thriftExtension, protoExtension)
	}
	idx := strings.Index(idlFile, p.idlRootDir)
	if idx == -1 {
		return "", errors.Errorf(
			"file %s is not in IDL dir (%s)",
			idlFile, p.idlRootDir,
		)
	}
	var prefix string
	if strings.HasSuffix(idlFile, thriftExtension) {
		prefix = strings.TrimSuffix(idlFile, thriftExtension)
	} else {
		prefix = strings.TrimSuffix(idlFile, protoExtension)
	}

	// Strip the leading / and the trailing .thrift/.proto suffix.
	segment := idlFile[idx+len(p.idlRootDir)+1 : len(prefix)]

	packageName := strings.Replace(segment, "/", "_", -1)
	return CamelCase(packageName), nil
}

func (p PackageHelper) getRelativeFileName(idlFile string) (string, error) {
	if !strings.HasSuffix(idlFile, thriftExtension) && !strings.HasSuffix(idlFile, protoExtension) {
		return "", errors.Errorf("file %s is not %s or %s", idlFile, thriftExtension, protoExtension)
	}
	idx := strings.Index(idlFile, p.idlRootDir)
	if idx == -1 {
		return "", errors.Errorf(
			"file %s is not in IDL dir (%s)",
			idlFile, p.idlRootDir,
		)
	}
	return idlFile[idx+len(p.idlRootDir):], nil
}

// GetModuleIdlSubDir returns subdir for idl per module
func (p PackageHelper) GetModuleIdlSubDir(isEndpoint bool) string {
	className := p.getModuleClass(isEndpoint)
	if subDir, ok := p.moduleIdlSubDir[className]; ok {
		return subDir
	}
	panic(fmt.Sprintf("unrecognized module %s", className))
}

func (p PackageHelper) getModuleClass(isEndpoint bool) string {
	if isEndpoint {
		return endpointModuleName
	}
	return defaultModuleFallback
}

// TargetClientsInitPath returns where the clients init should go
func (p PackageHelper) TargetClientsInitPath() string {
	return path.Join(p.targetGenDir, "clients", "clients.go")
}

// TargetEndpointsRegisterPath returns where the endpoints register file
// should be written to
func (p PackageHelper) TargetEndpointsRegisterPath() string {
	return path.Join(p.targetGenDir, "endpoints", "register.go")
}

// EndpointTestConfigPath returns the path for the endpoint test configs
func (p PackageHelper) EndpointTestConfigPath(
	serviceName, methodName string,
) string {
	fileName := strings.ToLower(methodName) + "_test.yaml"
	return path.Join(p.testConfigsRootDir, strings.ToLower(serviceName), fileName)
}

// TypeFullName returns the referred Go type name in generated code.
func (p PackageHelper) TypeFullName(typeSpec compile.TypeSpec) (string, error) {
	if typeSpec == nil {
		return "", nil
	}
	return GoType(p, typeSpec)
}

// DeputyReqHeader returns the header name that will be checked to determine
// if a request should go to the deputy downstream client
func (p PackageHelper) DeputyReqHeader() string {
	return p.deputyReqHeader
}

// DefaultHeaders returns a list of headers that will always be forwarded
// when present in the incoming request.
func (p PackageHelper) DefaultHeaders() []string {
	return p.defaultHeaders
}
