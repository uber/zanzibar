# <img src="zanzibar.png" width="352">    [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![Go Report Card][go-report-img]][go-report]


Zanzibar is an extensible framework to build configuration driven web applications. The goal of Zanzibar is to simplify application development into two steps:

1. write configurations for the application and its components;
2. write code to implement and test business logic.

Based on the configurations, Zanzibar generates boilerplates and glue code, wires them up with your business domain code and the runtime components Zanzibar provides to create a deployable binary.

The builtin components of Zanzibar makes it easy to develop microservices and gateway services that proxy or orchestrate microservices. It is also simple to extend Zanzibar with custom plugins to ease the development of applications that suit your specific needs.

Table of Contents
=================

  * [Concepts](#concepts)
     * [Module](#module)
        * [ModuleClass](#moduleclass)
        * [Type](#type)
           * [Client](#client)
           * [Endpoint](#endpoint)
           * [Middleware](#middleware)
           * [Service](#service)
        * [Dependency Injection](#dependency-injection)
           * [Dependency Injection](#dependency-injection-1)
           * [Module Initialization](#module-initialization)
           * [Dependency Rules](#dependency-rules)
     * [Config](#config)
        * [Config Layout](#config-layout)
        * [Module Config](#module-config)
           * [General Layout](#general-layout)
           * [Non-Config Content](#non-config-content)
              * [Custom Client](#custom-client)
              * [Circuit Breaker](#circuit-breaker)
              * [Custom Workflow](#custom-workflow)
           * [Grouping](#grouping)
           * [Config Schema](#config-schema)
        * [Application Config](#application-config)
     * [Code Generation](#code-generation)
        * [Go Structs and (de)serializers](#go-structs-and-deserializers)
        * [Zanzibar-generated Code](#zanzibar-generated-code)
  * [How to Use](#how-to-use)
     * [Install](#install)
     * [Code Gen](#code-gen)
     * [Testing](#testing)
        * [Entry Points](#entry-points)
           * [Service](#service-1)
           * [Endpoint](#endpoint-1)
        * [Fixture](#fixture)
     * [Extend Zanzibar](#extend-zanzibar)
        * [New ModuleClass or Type](#new-moduleclass-or-type)
        * [PostGenHook](#postgenhook)
  * [Development](#development)
     * [Installation](#installation)
     * [Running make generate](#running-make-generate)
     * [Running the tests](#running-the-tests)
     * [Running the benchmarks](#running-the-benchmarks)
     * [Running the end-to-end benchmarks](#running-the-end-to-end-benchmarks)
     * [Running the server](#running-the-server)
     * [Adding new dependencies](#adding-new-dependencies)

## Concepts
Zanzibar is built on three pillars: module, config, code generation.
### Module
Modules are the components that a Zanzibar application is made of. A module belongs to a `ModuleClass`, has a `type` and can have dependencies on other modules.

#### ModuleClass
`ModuleClass` abstracts the functionality of a specific class of components. Zanzibar predefines a few module classes, i.e., `client`, `endpoint`, `middleware` and `service`. Each represents a corresponding abstraction:

ModuleClass | Abstraction
----------- | -----------
client | clients to communicate with downstreams, e.g., database clients and RPC clients
endpoint | application interfaces exposed to upstreams
middleware | common functionality that has less to do with business logic, e.g., rate limiting middleware
service | a collection of endpoints that represents high level application abstraction, e.g., a demo service that prints "Hello World!"

#### Type
The module `type` differentiates module instances of the same `ModuleClass` with further classification. Types are somewhat arbitrary as they are not necessarily abstractions but indications about how Zanzibar should treat the modules.
##### Client
A client module could be of type `http`, `tchannel` or `custom`, where `http` or `tchannel` means Zanzibar will generate a client with given configuration that speaks that protocol while `custom` means the client is fully provided and Zanzibar will use it as is without code generation. In other words, `http` and `tchannel` clients are configuration driven (no user code) whereas `custom` clients are user-defined and can be "smart clients".
##### Endpoint
An `endpoint` module could also be of type `http` or `tchannel`, which determines the protocol that the endpoint will be made available to invoke externally via the Zanzibar router. While `endpoint` modules do not have `custom` type, each method of an `endpoint` module has a `workflowType` that indicates the type of workflow the endpoint method fulfills. The builtin workflow type is `httpClient`, `tchannelClient` and `custom`, where `httpClient` and `tchannelClient` means the endpoint method workflow is to proxy to a client, and `custom` means the workflow is fulfilled by user code, see more in [Custom Workflow](#custom-workflow).

 Note that workflow type is likely to be deprecated in the future so that proxy to a client will be no longer a builtin option.

##### Middleware
The builtin type of middleware module is `default`.

##### Service
The builtin service type is `gateway` (it is likely to change in the future, because `default` is probably a better name).

**Note** Zanzibar has support for user-defined module classes and module types in case the builtin types are not sufficient. The preferred way of extending Zanzibar is through user defined module classes and module types.

#### Dependency Injection
Module dependencies describe the relationships among various modules. The dependency relationship is critical to correctly assemble the modules to a full application.

##### Dependency Injection
A module is expected to define its immediate or direct dependencies. Zanzibar generates a module constructor with dependent modules as parameters, and passes the dependencies to the constructor during initilizaiton.

##### Module Initialization
Zanzibar also constructs a full global dependency graph for a given set of modules. This graph is used to initialize modules in the correct order, e.g. leaf modules are initialized first and then passed to the constructors of parent modules for initialization.

##### Dependency Rules
To establish and enforce abstraction boundaries, dependency rules at `ModuleClass` level are necessary. Zanzibar predefines the following dependency rules for built module classes:

ModuleClass | DependsOn | DependedBy
----------- | --------- | ----------
client | N/A | middleware, endpoint
middleware | client | endpoint
endpoint | client, middleware | service
service | endpoint | N/A

This table exhausts the possible immediate or direct dependency relationships among builtin module classes. Take endpoint module class for example, an endpoint module can depend on client or middleware modules but not endpoint or service modules. The reasoning for such rules aligns with the abstractions the module classes represent.

The `ModuleClass` struct has `DependsOn` and `DependedBy` public fields, which makes it simple to extend the dependency rules with custom module class, e.g., we can define a custom module class `task` that abstracts common business workflow by setting its `DependsOn` field to client and `DependedBy` field to endpoint.

### Config
Configurations are the interface that developers interact with when using the Zanzibar framework, they make up most of Zazibar's API. Various configurarions contain essential meta information of a Zanzibar application and its components. They are source of truth of the application.

#### Config Layout
Because configurations are the core of a Zanzibar application, we create a root directory to host configuration files when starting a Zanzibar application. There are a few typical directories and files under the root directory. Take [example-gateway](https://github.com/uber/zanzibar/tree/master/examples/example-gateway) for example:
```
example-gateway                 # root directory
├── bin                         # directory for generated application binaries
│   └── example-gateway         # generated example-gateway binary
├── build                       # directory for all generated code
│   ├── clients                 # generated mocks and module initializers for clients
│   ├── endpoints               # generated mocks and module initializers for endpoints
│   ├── gen-code                # generated structs and (de)serializers by Thrift compiler
│   ├── middlewares             # generated module initializers for middlewares
│   │   └── default             # generated module initializers for default middlewares
│   └── services                # generated mocks and module intialziers for services
├── build.yaml                  # config file for Zanzibar code generation, see below for details
├── clients                     # config directory for modules of client module class
│   └── bar                     # config directory for a client named 'bar'
├── config                      # config directory for application runtime properties
│   ├── production.yaml         # config file for production environment
│   └── test.yaml               # config file for test environment
├── copyright_header.txt        # optional copyright header for open source application
├── endpoints                   # config directory for modules of endpoint module class
│   └── bar                     # config directory for an endpoint named 'bar'
├── idl                         # idl directory for all thrift files
│   ├── clients                 # idl directory for client thrift files
│   └── endpoints               # idl directory for endpoint thrift files
├── middlewares                 # config directory for modules of middleware module class
│   ├── transform-response      # config directory for a middleware named 'transform-response'
│   ├── default                 # directory for all default middlewares
│   │   └── log-publisher       # config directory for a default middleware named 'log-publisher'
│   └── default.yaml            # config file describing default middlewares and their execution order   
└── services                    # config directory for modules of service module class
    └── example-gateway         # config directory for a service named 'example-gateway'
```

#### Module Config
Each module must have a config file so that it can be recognized by Zanzibar. This section explains how the module config files are organized and what goes into them.

##### General Layout
 Under the application root directory, there should be a corresponding top level config directory for each module class. For Zanzibar builtin module classes, the name of the directory is the plural of the module class name, e.g., a `clients` directory for `client` module class. The directory name is used when registering generator for a module class ([example](https://github.com/uber/zanzibar/blob/master/codegen/module_system.go#L248-L254)). While it is not required,  the same directory naming convention should be followed when defining custom module classes.

 Under a module class directory, there should be a corresponding config directory for each module, e.g., the `clients` directory has a few subdirectories and each of them corresponds to a module.

 Under a module directory, there should be a YAML file that contains the meta information of that module. It is required that the file is named of `{$ModuleClass}-config.yaml`, e.g. the path to the YAML config file of `bar` client module is `clients/bar/client-config.yaml`, similarly the path to the YAML config file of `bar` endpoint module is `endpoints/bar/endpoint-config.yaml`.

##### Non-Config Content
Besides the YAML config file, the module directory also contains other necessary directories/files. For example, the [quux](https://github.com/uber/zanzibar/tree/master/examples/example-gateway/clients/quux) client is a custom (non-generated) client, its module config directory has following layout:
```
quxx                        # client module config directory
├── client-config.yaml      # client module config file
├── fixture                 # directory for fixtures used for testing
│   └── fixure.go           # fixtures that can be used by a generated mock client for testing
└── quux.go                 # custom client implementation, package is imported by generated code
```
For client and endpoint modules of builtin type `custom`, Zanzibar expects user code to be placed in the module directory. This is important because Zaznibar-generated code refers to user code by importing the package of the module directory path. Furthermore, user code of custom client and endpoint modules must also define and implement necessary **public** types and interfaces so that Zanzibar can wire up the modules.

###### Custom Client
For client module of custom type, user code must define a `Client` interface and a `NewClient` constructor that returns the `Client` interface. Below is the example code [snippet](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/clients/quux/quux.go) for the `quux` custom client:
```golang
package quux

import "github.com/uber/zanzibar/examples/example-gateway/build/clients/quux/module"

type Client interface {
	Echo(string) string
}

func NewClient(deps *module.Dependencies) Client {
	return &quux{}
}

type quux struct{}

func (c *quux) Echo(s string) string { return s }
```
Note the type of `deps` parameter passed to `NewClient` constructor function is generated by Zanzibar, as indicated by the import path. Zanzibar takes care of initializing and passing in the acutal `deps` argument, as mentioned in [Dependency Injection](#dependency-injection).

###### Circuit Breaker
For increasing overall system resiliency, zanzibar uses a [Circuit Breaker](https://msdn.microsoft.com/en-us/library/dn589784.aspx) which avoids calling client when there is an increase in failure rate beyond a set
threshold. After a sleepWindowInMilliseconds, client calls are attempted recovery by going in half-open and then close state.

circuitBreakerDisabled: Default false. To disable the circuit-breaker:
```
    "clients.<clientID>.circuitBreakerDisabled" : true
```

maxConcurrentRequests: Default 50. To set how many requests can be run at the same time, beyond which requests are
rejected:
```
   "clients.<clientID>.maxConcurrentRequests": 50
```

errorPercentThreshold: Default 20. To set error percent threshold beyond which to trip the circuit open:
```
    "clients.<clientID>.errorPercentThreshold": 20
```

requestVolumeThreshold: Default 20. To set minimum number of requests that will trip the circuit in a rolling window
 of 10 (For example, if the value is 20, then if only 19 requests are received in the rolling window of 10 seconds the
circuit will not trip open even if all 19 failed):
```
    "clients.<clientID>.requestVolumeThreshold" : true
```

sleepWindowInMilliseconds: Default 5000. To set the amount of time, after tripping the circuit, to reject requests
before allowing attempts again to determine if the circuit should again be closed:
```
    "clients.<clientID>.sleepWindowInMilliseconds" : true
```


###### Custom Workflow
For endpoint module of custom workflow type, user code must define a `New{$endpoint}{$method}Workflow` constructor that returns the Zanzibar-generated `{$endpoint}{$method}Workflow` interface which has a sole `Handle` method. Below is the example code [snippet](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/endpoints/contacts/save_contacts.go) for the `contacts` custom endpoint:
```go
package contacts

import (
	"context"

	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/contacts/module"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/contacts/workflow"
	contacts "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/contacts/contacts"

	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"
)

func NewContactsSaveContactsWorkflow(
    c *module.ClientDependencies,
    l *zap.Logger,
) workflow.ContactsSaveContactsWorkflow { return &saveContacts{ ... } }

type saveContacts struct { ... }

func (w *saveContacts) Handle(
	ctx context.Context,
	headers zanzibar.Header,
	r *contacts.SaveContactsRequest,
) (*contacts.SaveContactsResponse, zanzibar.Header, error) { ... }
```
The idea of the workflow constructor is similar to the client constructor, with a couple of differences:

- the first parameter is specifically `ClientDependencies` and there is an additional logger parameter, this will be changed in the future so that the dependency parameter is generalized;
- the return value is an interface generated by Zanzibar, the parameter and return value of the `Handle` method refers to structs generated by Thrift compiler based on the endpoint thrift file configured in the endpoint-config.yaml, see more in [Config Schema](#config-schema).


##### Grouping
Zanzibar allows nesting module config directories in the sub-directories of a module class config directory. This is useful to group related modules under a sub-directory. For example, the [tchannel](https://github.com/uber/zanzibar/tree/master/examples/example-gateway/endpoints/tchannel) directory groups all [TChannel](https://github.com/uber/tchannel) endpoint modules:
```
endpoints
├── ...
└── tchannel                    # this directory does not correspond to a module, it represents a group
    └── baz                     # module config directory under the 'tchannel' group
        ├── baz_call.go
        ├── baz_call_test.go
        ├── call.yaml
        └── endpoint-config.yaml
```
##### Config Schema
Modules of different `ModuleClass` and `type` are likely to have different fields in their config files. Developers are expected to write module config files according to the [schemas](https://github.com/uber/zanzibar/tree/master/docs).

**Note**: fields are absent in config schemas but present in examples are experimental.

The endpoint module config is different from other module classes as it has multiple YAML files, where each endpoint method corresponds to a YAML file and the endpoint-config.yaml file refers to them.
```
endpoints/multi
├── endpoint-config.yaml    # has a field 'endpoints' that is a list and contains helloA and helloB
├── helloA.yaml             # config file for method helloA
└── helloB.yaml             # config file for method helloB
```
The reason for such layout is to avoid a large endpoint-config.yaml file when an endpoint has many methods.

#### Application Config
Besides the module configs, Zanzibar also expects a YAML file that configures necessary properties to boostrap the code generation process of a Zanzibar application. The schema for application config is defined [here](https://github.com/uber/zanzibar/tree/master/docs/application_config_schema.json).

Unlike the module configs, there is no restriction on how this config file should be named. It can be named `{$appName}.yaml` or `build.yaml` as it is in [example-gateway](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/build.yaml), as long as it is passed correctly as an argument to the code generation [runner](https://github.com/uber/zanzibar/blob/master/codegen/runner/runner.go).

In this config file, you can specify the paths from which to discover modules. You can also specify `default dependencies`.

`Default Dependencies` allow module classes to include instances of other module classes as default dependencies. This means that no explicit configurations are required for certain module instances to be included as a dependency. e.g., we can include `clients/logger` as a default dependency for `endpoint`, and every endpoint will have `clients/logger` as a dependency in its `module/dependencies.go` file, even if the endpoint's `endpoint-config.yaml` file does not list `clients/logger` as a dependency.

Note that these paths support `Glob` patterns.

### Code Generation
Zanzibar provides HTTP and TChannel runtime components for both clients and servers. Once all the configs are properly defined, Zanzibar is able to parse the config files and generate code and wire it up with the runime components to produce a full application. All generated code is placed in the `build` directory.

#### Go Structs and (de)serializers
Zanzibar expects non-custom clients and endpoints to define their interfaces using Thrift ([Zanzibar Thrift file semantics](https://github.com/uber/zanzibar/blob/master/docs/thrift.md)). For example, the `bar` endpoint defines its interfaces using the [bar.thrift](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/idl/endpoints-idl/endpoints/bar/bar.thrift) as specified in [hello.yaml](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/endpoints/bar/hello.yaml#L5). The data types in such thrift files must have their equivalents in Go.

- For tchannel clients/endpoints, network communication is Thrift over TChannel. Zanzibar uses [thriftrw](https://github.com/thriftrw/thriftrw-go) to generate Go structs and thrift (de)serializers;
- For http clients/endpoints, network communication is JSON over HTTP. Zanzibar uses [thriftrw](https://github.com/thriftrw/thriftrw-go) to generate Go structs and then uses [easyjson](https://github.com/mailru/easyjson) to generate JSON (de)serializers.

The [pre-steps.sh](https://github.com/uber/zanzibar/blob/master/codegen/runner/pre-steps.sh) script takes care of this part of the code generation, and places the generated code under `build/gen-code` directory.

#### Zanzibar-generated Code
Everything except `gen-code` under `build` directory is generated by Zanzibar. Zanzibar parses config files for each module to gathers meta information and then executing various [templates](https://github.com/uber/zanzibar/tree/master/codegen/templates) by applying them to the meta data. Here is what is generated for each builtin module class:

- client: dependency type, client interface and constructor if non-custom, mock client constructor
- middleware: dependency type, middleware type and constructor (unstable)
- endpoint: dependency type, endpoint type and constructor, workflow interface, workflow if non-custom, mock workflow constructor if custom
- service: dependency type and initializer, main.go, mock service constructor, service constructor

## How to Use
### Install
Assuming you are using a vendor package management tool like Glide, then the minimal glide.yaml file would look like:
```yaml
- package: go.uber.org/thriftrw
  version: ^1.8.0
- package: github.com/mailru/easyjson
  version: master
- package: github.com/uber/zanzibar
  version: master
```
### Code Gen
After installing the packages, create your module configs and application config in your application root directory. Then you are ready to run the following script to kick off code generation:
```bash
# put this script in application root directory

CONFIG_DIR="."
BUILD_DIR="$CONFIG_DIR/build"
THRIFTRW_SRCS=""

# find all thrift files specified in the config files
config_files=$(find "." -name "*-config.yaml" ! -path "*/build/*" ! -path "*/vendor/*" | sort)
for config_file in ${config_files}; do
	dir=$(dirname "$config_file")
	yaml_files=$(find "$dir" -name "*.yaml")
	for yaml_file in ${yaml_files}; do
		thrift_file=$(yq -r '.. | .idlFile? | select(strings | endswith(".thrift"))' "$yaml_file")
		[[ -z ${thrift_file} ]] && continue
		[[ ${THRIFTRW_SRCS} == *${thrift_file}* ]] && continue
        THRIFTRW_SRCS+=" $CONFIG_DIR/idl/$thrift_file"
	done
done

bash ./vendor/github.com/uber/zanzibar/codegen/runner/pre-steps.sh "$BUILD_DIR" "$CONFIG_DIR" "zanzibar" "$THRIFTRW_SRCS"
go run ./vendor/github.com/uber/zanzibar/codegen/runner/runner.go --config="$CONFIG_DIR/build.yaml"
```
Note the above script will be abstracted for easier usage in the future.
### Testing
Zanzibar comes with builtin integration testing frameworks to help test business logic with ease. Setting [genMock](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/build.yaml#L16) to true will trigger Zanzibar to generate mock client, workflow and service constructors. The mock clients, being the leaf nodes in the dependency graph, are wired with the rest modules to create a testing application, which you can test against by setting expectations of the mock clients. The generated test helpers make writing tests straightforward and concise.

#### Entry Points
Currently Zanzibar provides two entry points to write integration tests: service and endpoint.
##### Service
Service level integration testing treats your application as a black box. Zanzibar starts a local server for your application and you write tests by sending requests to the server and verify the response is expected.
```go
func TestSaveContacts(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	ms.MockClients().Contacts.ExpectSaveContacts().Success()

	endpointReqeust := &endpointContacts.SaveContactsRequest{
		Contacts: []*endpointContacts.Contact{},
	}
	rawBody, _ := endpointReqeust.MarshalJSON()

	res, err := ms.MakeHTTPRequest(
		"POST", "/contacts/foo/contacts", nil, bytes.NewReader(rawBody),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "202 Accepted", res.Status)
}
```
##### Endpoint
Endpoint level integration testing allows focusing on testing the business logic without a full server setup. It is lightweighted and feels more like unit tests.
```go
func TestSaveContactsCallWorkflow(t *testing.T) {
	mh, mc := mockcontactsworkflow.NewContactsSaveContactsWorkflowMock(t)

	mc.Contacts.ExpectSaveContacts().Success()

	endpointReqeust := &endpointContacts.SaveContactsRequest{
		UserUUID: "foo",
		Contacts: []*endpointContacts.Contact{},
	}

	res, resHeaders, err := mh.Handle(context.Background(), nil, endpointReqeust)

	if !assert.NoError(t, err, "got error") {
		return
	}
	assert.Nil(t, resHeaders)
	assert.Equal(t, &endpointContacts.SaveContactsResponse{}, res)
}
```
The above snippets can be found in [save_contacts_test.go](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/endpoints/contacts/save_contacts_test.go).

#### Fixture
Zanzibar uses [gomock](https://github.com/golang/mock) to generate client mocks. To avoid manually setting the same fixture expectations again and again, Zanzibar augments gomock-generated mocks with fixture support. For example, the client-config.yaml file of `contacts` client has a `fixture` [field](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/clients/contacts/client-config.yaml#L11-L18):
```yaml
fixture:
  importPath: github.com/uber/zanzibar/examples/example-gateway/clients/contacts/fixture
  scenarios:
    SaveContacts:
    - success

```
This basically says the `saveContacts` method has a `success` scenario which is defined in the [fixture package](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/clients/contacts/fixture/fixture.go) indicated by the `importPath`. The fixture package is provided by users and here is what it looks like:
```go
package fixture

import (
	mc "github.com/uber/zanzibar/examples/example-gateway/build/clients/contacts/mock-client"
	gen "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/contacts/contacts"
)

var saveContactsFixtures = &mc.SaveContactsScenarios{
	Success: &mc.SaveContactsFixture{
		Arg0Any: true,
		Arg1Any: true,
		Arg2: &gen.SaveContactsRequest{
			UserUUID: "foo",
		},

		Ret0: &gen.SaveContactsResponse{},
	},
}

// Fixture ...
var Fixture = &mc.ClientFixture{
	SaveContacts: saveContactsFixtures,
}
```

With that, in your tests you will be able to write
```go
mc.Contacts.ExpectSaveContacts().Success()
```
rather than
```go
s.mockClient.EXPECT().SaveContacts(arg0, arg1, arg2).Return(ret0, ret1, ret2)
```
Check out [fixture abstraction](https://gist.github.com/ChuntaoLu/9d5dcf37376b855f64e63156f765d183) to see how it works.
### Extend Zanzibar
Once the concepts of module, config and code generation are clear, extending Zanzibar becomes straightforward. There are two ways to extend Zanzibar.
#### New ModuleClass or Type
To extend Zanzibar with new module class or type is simply to extend each of its three pillars. For example, we want to add a new `task` module class to abstract common business workflow, here is what we need to do for each pillar:

- module: understand what meta information is needed for each task module;
- config: add a `tasks` directory under the application root directory, define proper schema for task module class;
- code generation: add templates for task if necessary, create a code generator that implements the [BuildGenerator](https://github.com/uber/zanzibar/blob/master/codegen/module.go#L1041) interface and [register](https://github.com/uber/zanzibar/blob/master/codegen/module_system.go#L248) it onto the module system for the task module class.

The same idea applies for adding new types of an existing module class.
#### PostGenHook
Zanzibar provides post-generation [hooks](https://github.com/uber/zanzibar/blob/master/codegen/module.go#L68) which has access to the meta information of all modules. You can do whatever (mutating the input is probably not a good idea) suits your needs within a post-generation hook. Zanzibar invokes post-generation hooks as the very last step of code generation. In fact, mocks are all generated via post-generation [hooks](https://github.com/uber/zanzibar/blob/master/codegen/module_system.go#L226).

## Development
### Installation

```
mkdir -p $GOPATH/src/github.com/uber
git clone git@github.com:uber/zanzibar $GOPATH/src/github.com/uber/zanzibar
cd $GOPATH/src/github.com/uber/zanzibar
GO111MODULE=off make install
```
### Running make generate
```
make generate
```
### Running the tests

```
make test
```

### Running the benchmarks

```
for i in `seq 5`; do make bench; done
```

### Running the end-to-end benchmarks

First fetch `wrk`

```
git clone https://github.com/wg/wrk ~/wrk
cd ~/wrk
make
sudo ln -s $HOME/wrk/wrk /usr/local/bin/wrk
```

Then you can run the benchmark comparison script

```
# Assume you are on feature branch ABC
./benchmarks/compare_to.sh master
```

### Running the server

First create log dir...

```
sudo mkdir -p /var/log/my-gateway
sudo chown $USER /var/log/my-gateway
chmod 755 /var/log/my-gateway

sudo mkdir -p /var/log/example-gateway
sudo chown $USER /var/log/example-gateway
chmod 755 /var/log/example-gateway
```

```
make run
# Logs are in /var/log/example-gateway/example-gateway.log
```

### Adding new dependencies

We use glide @ 0.12.3 to add dependencies.

Download [glide @ 0.12.3](https://github.com/Masterminds/glide/releases)
and make sure it's available in your path

If we want to add a dependency:

 - Add a new section to the glide.yaml with your package and version
 - run `glide up --quick`
 - check in the `glide.yaml` and `glide.lock`

If you want to update a dependency:

 - Change the `version` field in the `glide.yaml`
 - run `glide up --quick`
 - check in the `glide.yaml` and `glide.lock`

[doc-img]: https://godoc.org/github.com/uber/zanzibar?status.svg
[doc]: https://godoc.org/github.com/uber/zanzibar
[ci-img]: https://github.com/uber/zanzibar/actions/workflows/build.yml/badge.svg
[ci]: https://github.com/uber/zanzibar/actions
[cov-img]: https://coveralls.io/repos/github/uber/zanzibar/badge.svg?branch=master
[cov]: https://coveralls.io/github/uber/zanzibar?branch=master
[go-report-img]: https://goreportcard.com/badge/github.com/uber/zanzibar
[go-report]: https://goreportcard.com/report/github.com/uber/zanzibar
