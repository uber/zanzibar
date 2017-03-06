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
	"os"
	"path/filepath"
	"runtime"

	"io/ioutil"

	"encoding/json"

	"github.com/pkg/errors"
)

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(file)
}

var templateDir = filepath.Join(getDirName(), "templates", "*.tmpl")
var mandatoryClientFields = []string{
	"clientType",
	"clientId",
	"thriftFile",
	"thriftFileSha",
	"clientName",
	"serviceName",
}
var mandatoryEndpointFields = []string{
	"endpointType",
	"endpointId",
	"handleId",
	"thriftFile",
	"thriftFileSha",
	"thriftMethodName",
	"workflowType",
	"testFixtures",
	"middlewares",
}

// ClientSpec holds information about each client in the
// gateway included its thriftFile and other meta info
type ClientSpec struct {
	// ModuleSpec holds the thrift module information
	ModuleSpec *ModuleSpec
	// ClientType, currently only "http" is supported
	ClientType string
	// GoFileName, the absolute path where the generate client is
	GoFileName string
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
	ThriftServiceName string
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

	clientConfigObj := map[string]string{}
	err = json.Unmarshal(bytes, &clientConfigObj)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not parse json file %s: ", jsonFile,
		)
	}

	if clientConfigObj["clientType"] != "http" {
		return nil, errors.Errorf(
			"Cannot support unknown clientType for client %s", jsonFile,
		)
	}

	for i := 0; i < len(mandatoryClientFields); i++ {
		fieldName := mandatoryClientFields[i]
		if clientConfigObj[fieldName] == "" {
			return nil, errors.Errorf(
				"client config (%s) must have %s field", jsonFile, fieldName,
			)
		}
	}

	thriftFile := filepath.Join(
		h.ThriftIDLPath(), clientConfigObj["thriftFile"],
	)

	mspec, err := NewModuleSpec(thriftFile, h)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not build module spec for thrift %s: ", thriftFile,
		)
	}

	baseName := filepath.Base(jsonFile)
	baseName = baseName[0 : len(baseName)-5]

	goFileName := filepath.Join(
		h.CodeGenTargetPath(),
		"clients",
		baseName,
		baseName+".go",
	)

	goStructsFileName := filepath.Join(
		h.CodeGenTargetPath(),
		"clients",
		baseName,
		baseName+"_structs.go",
	)

	return &ClientSpec{
		ModuleSpec:        mspec,
		ClientType:        clientConfigObj["clientType"],
		GoFileName:        goFileName,
		GoStructsFileName: goStructsFileName,
		ThriftFile:        thriftFile,
		ClientID:          clientConfigObj["clientId"],
		ClientName:        clientConfigObj["clientName"],
		ThriftServiceName: clientConfigObj["serviceName"],
	}, nil
}

// EndpointSpec holds information about each endpoint in the
// gateway including its thriftFile and meta data
type EndpointSpec struct {
	// ModuleSpec holds the thrift module info
	ModuleSpec *ModuleSpec

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
	// TestFixtures, meta data to generate tests, TODO
	TestFixtures []int
	// Middlewares, meta data to add middlewares, TODO
	Middlewares []int

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

// NewEndpointSpec creats an endpoint spec from a json file.
func NewEndpointSpec(
	jsonFile string, h *PackageHelper,
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

	if endpointConfigObj["endpointType"] != "http" {
		return nil, errors.Errorf(
			"Cannot support unknown endpointType for endpoint %s", jsonFile,
		)
	}

	for i := 0; i < len(mandatoryEndpointFields); i++ {
		fieldName := mandatoryEndpointFields[i]
		if endpointConfigObj[fieldName] == "" {
			return nil, errors.Errorf(
				"endpoint config (%s) must have %s field", jsonFile, fieldName,
			)
		}
	}

	thriftFile := filepath.Join(
		h.ThriftIDLPath(), endpointConfigObj["thriftFile"].(string),
	)

	mspec, err := NewModuleSpec(thriftFile, h)
	if err != nil {
		return nil, errors.Wrapf(
			err, "Could not build module spec for thrift %s: ", thriftFile,
		)
	}

	var workflowImportPath string
	var clientName string
	var clientMethod string

	workflowType := endpointConfigObj["workflowType"].(string)
	if workflowType == "httpClient" {
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

	// baseName := filepath.Base(jsonFile)
	// baseName = baseName[0 : len(baseName)-5]

	return &EndpointSpec{
		ModuleSpec:         mspec,
		EndpointType:       endpointConfigObj["endpointType"].(string),
		EndpointID:         endpointConfigObj["endpointId"].(string),
		HandleID:           endpointConfigObj["handleId"].(string),
		ThriftFile:         thriftFile,
		ThriftMethodName:   endpointConfigObj["thriftMethodName"].(string),
		TestFixtures:       endpointConfigObj["testFixtures"].([]int),
		Middlewares:        endpointConfigObj["middlewares"].([]int),
		WorkflowType:       workflowType,
		WorkflowImportPath: workflowImportPath,
		ClientName:         clientName,
		ClientMethod:       clientMethod,
	}, nil
}

// GatewaySpec collects information for the entire gateway
type GatewaySpec struct {
	// package helper for gateway
	PackageHelper *PackageHelper
	// tempalte instance for gateway
	Template *Template

	ClientModules   map[string]*ClientSpec
	EndpointModules map[string]*EndpointSpec

	gatewayName       string
	configDirName     string
	clientConfigDir   string
	endpointThriftDir string
}

// NewGatewaySpec sets up gateway spec
func NewGatewaySpec(
	configDirName string,
	thriftRootDir string,
	typeFileRootDir string,
	targetGenDir string,
	gatewayThriftRootDir string,
	clientConfig string,
	endpointThriftDir string,
	endpointConfig string,
	gatewayName string,
) (*GatewaySpec, error) {
	packageHelper, err := NewPackageHelper(
		thriftRootDir,
		typeFileRootDir,
		targetGenDir,
		gatewayThriftRootDir,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot build package helper")
	}

	tmpl, err := NewTemplate(templateDir)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create template")
	}

	clientJsons, err := filepath.Glob(filepath.Join(
		configDirName,
		clientConfig,
		"*.json",
	))
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load client json files")
	}

	endpointJsons, err := filepath.Glob(filepath.Join(
		configDirName,
		endpointConfig,
		"*.json",
	))
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load endpoint json files")
	}

	spec := &GatewaySpec{
		PackageHelper:   packageHelper,
		Template:        tmpl,
		ClientModules:   map[string]*ClientSpec{},
		EndpointModules: map[string]*EndpointSpec{},

		configDirName:     configDirName,
		clientConfigDir:   clientConfig,
		endpointThriftDir: endpointThriftDir,
		gatewayName:       gatewayName,
	}

	for _, json := range clientJsons {
		cspec, err := NewClientSpec(json, packageHelper)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot parse client json file %s :", json,
			)
		}
		spec.ClientModules[cspec.ThriftFile] = cspec
	}
	for _, json := range endpointJsons {
		espec, err := NewEndpointSpec(json, packageHelper)
		if err != nil {
			return nil, errors.Wrapf(
				err, "Cannot parse endpoint json file %s :", json,
			)
		}
		spec.EndpointModules[espec.ThriftFile] = espec
	}

	return spec, nil
}

// GenerateClients will generate all the clients for the gateway
func (gateway *GatewaySpec) GenerateClients() error {
	for _, module := range gateway.ClientModules {
		_, err := gateway.Template.GenerateClientFile(
			module, gateway.PackageHelper,
		)
		if err != nil {
			return err
		}
	}

	_, err := gateway.Template.GenerateClientsInitFile(
		gateway.ClientModules, gateway.PackageHelper,
	)
	return err
}

// GenerateEndpoints will generate all the endpoints for the gateway
func (gateway *GatewaySpec) GenerateEndpoints() error {
	for _, module := range gateway.EndpointModules {
		_, err := gateway.Template.GenerateEndpointFile(
			module, gateway.PackageHelper,
		)
		if err != nil {
			return err
		}

		_, err = gateway.Template.GenerateEndpointTestFile(
			module, gateway.PackageHelper,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// GenerateMain will generate the main files for the gateway
func (gateway *GatewaySpec) GenerateMain() error {
	_, err := gateway.Template.GenerateMainFile(
		gateway, gateway.PackageHelper,
	)
	return err
}
