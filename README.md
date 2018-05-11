# <img src="zanzibar.png" width="352">    [![GoDoc][doc-img]][doc] [![Build Status][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![Go Report Card][go-report-img]][go-report]


Zanzibar is an extensible framework to build configuration driven web applications. The goal of Zanizibar is to simplify application development into two steps:

1. write configurations for the application and its components;
2. write code to implement and test business logic.

Based on the configurations, Zanizbar generates boilerplates and glue code, wires them up with your business domain code and the runtime components Zanzibar provides to create a deployable binary.

The builtin components of Zanzibar makes it easy to develop mircoserivces and gateway services that proxy or orchestrate microsevices. It is also simple to extend Zanzibar with custom plugins to ease the development of applications that suit your specific needs.

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
servcie | high level appliation abstraction, e.g., a demo service that prints "Hello World!"

#### Type
The module `type` differentiates module instances of the same `ModuleClass` with further classification.

For example, a client module could be of type `http`, `tchannel` or `custom`, where `http` or `tchannel` means Zanziabr will generate a client with given configuration that speaks that protocol while `custom` means the client is fully provided and Zanzibar will use it as is without code generation.

An `endpoint` module could also be of type `http` or `tchannel`, which represents the endpoint's runtime protocol. While `endpoint` modules do not have `custom` type, each method of an `endpoint` module has a `workflowType` that indicates the type of workflow the endpoint method fulfills. The builtin workflow type is `httpClient`, `tchannelClient` and `custom`, where `httpClient` and `tchannelClient` means the endpoint method workflow is to proxy to a client, and `custom` means the workflow is fulfilled by user code, see more in [Custom Workflow](#custom-workflow). **Note** that workflow type is likely to be deprecated in the future so that proxy to a client will be no longer a builtin option.

The builtin type of middleware module is `default` and the builtin service type is `gateway` (`default` is probably a better name, up to change in the future). As you maybe already noticed, the types are somewhat arbitrary as they are not necessarily abstractions but indications about how Zanzibar should treat the modules.

Note that the module classes and types mentioned above are builtin in the sense that Zanzibar knows how to treat a module with given builtin classes and types. We can always easily extend by defining arbitrary module classs and types as long as we register appropriate module generators to the module system.

#### Dependency
Module dependencies describe the relationships among various modules. The dependency relationship is critical to correctly assemble the modules to a full application.

##### Dependency Injection
A module is expected to define its immediate or direct dependencies. Zanzibar generates a moudle constructor with depedent modules as parameters, and passes the dependencies to the constructor during initilizaiton.

##### Module Initialization
Zanzibar also constructs a full global dependency graph for a given set of modules. This graph is used to initialize modules in the correct order, e.g. leaf modules are initialized first and then passed to the constructors of parent modules for initialization.

##### Dependency Rules
To establish and enforce abstraction boundaries, dependency rules at `ModuleClass` level are necessary. Zanzibar predefines the following dependency rules for built module classes:

ModuleClass | DependsOn | DependedBy
----------- | --------- | ----------
client | N/A | middleware, endpoint
middleware | client | endpoint
endpoint | client, middleware | service
servcie | endpoint | N/A

This table exhausts the possible immediate or direct dependency relationships among builtin module classes. Take endpoint module class for example, an endpoint module can depend on client or middleware modules but not endpoint or servcie modules. The resoning for such rules aligns with the abstractions of the module classes represent.

The `ModuleClass` struct has `DependsOn` and `DependedBy` public fields, which makes it simple to extend the dependency rules with custom module class, e.g., we can define a custom module class `task` that abstracts common business workflow by setting its `DependsOn` field to client and `DependedBy` field to endpoint.

### Config
Configurations are the interface that developers interact with when using the Zanzibar framework, they make up most of Zazibar's API. Various configurarions contain essential meta information of a Zanziabr application and its components. They are source of truth of the application.

**Note**: Currently configurations are in JSON, we plan to migrate to YAML.

#### Config Layout
Because configurations are the core of a Zanzibar application, we create a root directory to host configuration files when starting a Zanziabr application. There are a few typical directories and files under the root directory. Take [example-gateway](https://github.com/uber/zanzibar/tree/master/examples/example-gateway) for example:
```
example-gateway                 # root directory
├── bin                         # directory for generated application binaries
│   └── example-gateway         # generated example-gateway binary
├── build                       # directory for all generated code
│   ├── clients                 # generated mocks and module initializers for clients
│   ├── endpoints               # generated mocks and module initializers for endpoints
│   ├── gen-code                # generated structs and (de)serializers by Thrift compiler
│   ├── middlewares             # generated module initializers for middlewares
│   └── services                # generated mocks and module intialziers for services
├── build.json                  # config file for Zanziabr code generation, see below for details
├── clients                     # config directory for modules of client module class
│   └── bar                     # config directory for a client named 'bar'
├── config                      # config directory for application runtime properties
│   ├── production.json         # config file for production environment
│   └── test.json               # config file for test environment
├── copyright_header.txt        # optional copyright header for open source application
├── endpoints                   # config directory for modules of endpoint module class
│   └── bar                     # config directory for an endpoint named 'bar'
├── idl                         # idl directory for all thrift files
│   ├── clients                 # idl directory for client thrift files
│   └── endpoints               # idl directory for endpoint thrift files
├── middlewares                 # config directory for modules of middleware module class
│   └── transform-response      # config directory for a middleware named 'transform-response'
└── services                    # config directory for modules of servcie module class
    └── example-gateway         # config directory for a servcie named 'example-gateway'
```

#### Module Config
Each module must have a config file so that it can be recognized by Zanzibar. This section explains how the module config files are organized and what goes into them.

##### General Layout
 Under the application root directory, there should be a corresponding top level config directory for each module class. For Zanzibar builtin module classes, the name of the directory is the plural of the module class name, e.g., a `clients` directory for `client` module class. The directory name is used when registering generator for a module class ([example](https://github.com/uber/zanzibar/blob/master/codegen/module_system.go#L248-L254)). While it is not required,  the same directory naming convention should be followed when defining custom module classes.

 Under a module class directory, there should be a corresponding config directory for each module, e.g., the `clients` directory has a few subdirectories and each of them corresponds to a module.

 Under a module directory, there should be a JSON file that contains the meta information of that module. It is required that the file is named of `{$ModuleClass}-config.json`, e.g. the path to the JSON config file of `bar` client module is `clients/bar/client-config.json`, similarly the path to the JSON config file of `bar` endpoint module is `endpoints/bar/endpont-config.json`.

##### Non-Config Content
Besides the JSON config file, the module directory also contains other necessary directories/files. For example, the [quux](https://github.com/uber/zanzibar/tree/master/examples/example-gateway/clients/quux) client is a custom (non-generated) client, its module config directory has following layout:
```
quxx                        # client module config directory
├── client-config.json      # client module config file
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

###### Custom Workflow
For endpoint module of custom workflow type, user code must define a `New{$endpoint}{$method}Workflow` constructor that returns the Zanzibar-generated `{$endpoint}{$method}Workflow` interface which has a sole `Handle` method. Below is the example code [snippet](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/endpoints/contacts/save_contacts.go) for the `contacts` custom endpoint:
```go
package contacts

import (
	"context"

	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/contacts/module"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/contacts/workflow"
	contacts "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/contacts/contacts"

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
- the return value is an interface generated by Zanzibar, the parameter and return value of the `Handle` method refers to structs generated by Thrift compiler based on the endpoint thrift file configured in the endpoint-config.json, see more in [Config Schema](#config-schema).


##### Grouping
Zanzibar allows nesting module config directories in the sub-directories of a module class config directory. This is useful to group related modules under a sub-directory. For example, the [tchannel](https://github.com/uber/zanzibar/tree/master/examples/example-gateway/endpoints/tchannel) directory groups all [TChannel](https://github.com/uber/tchannel) endpoint modules:
```
endpoints
├── ...
└── tchannel                    # this directory does not correspond to a module, it represents a group
    └── baz                     # module config directory under the 'tchannel' group
        ├── baz_call.go
        ├── baz_call_test.go
        ├── call.json
        └── endpoint-config.json
```
##### Config Schema
Modules of different `ModuleClass` and `type` are likely to have different fields in their config files. Developers are expected to write module config files according to the [schemas](https://github.com/uber/zanzibar/tree/master/docs).

**Note**: fields are absent in config schemas but present in examples are experimental.

The endpoint module config is different from other module classes as it has multiple JSON files, where each endpoint method corresponds to a JSON file and the endpoint-config.json file refers to them.
```
endpoints/multi
├── endpoint-config.json    # has a field 'endpoints' that is a list and contains helloA and helloB
├── helloA.json             # config file for method helloA
└── helloB.json             # config file for method helloB
```
The reason for such layout is to avoid a large endpoint-config.json file when an endpoint has many methods.

#### Application Config
Besides the module configs, Zanzibar also expects a JSON file that configures necessary properties to boostrap the code generation process of a Zanzibar application. The schema for application config is defined [here](https://github.com/uber/zanzibar/tree/master/docs/application_cofnig_schema.json).

Unlike the module configs, there is no restriction on how this config file should be named. It can be named `{$appName}.json` or `build.json` as it is in [example-gateway](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/build.json), as long as it is passed correctly as an argument to the code generation [runner](https://github.com/uber/zanzibar/blob/master/codegen/runner/runner.go).

### Code Generation
Zanzibar provides HTTP and TChannel runtime components for both clients and servers. Once all the configs are properly defined, Zanzibar is able to parse the config files and generate code and wire it up with the runime components to produce a full application. All generated code is placed in the `build` directory.

#### Go Structs and (de)serializers
Zanzibar expects non-custom clients and endpoints to define their interfaces using Thrift ([Zanziabr Thrift file semantics](https://github.com/uber/zanzibar/blob/master/docs/thrift.md)). For example, the `bar` endpoint defines its interfaces using the [bar.thrift](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/idl/endpoints/bar/bar.thrift) as specified in [hello.json](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/endpoints/bar/hello.json#L5). The data types in such thrift files must have their equivalents in Go.

- For tchannel clients/endpoints, network communication is Thrift over TChannel. Zanzibar uses [thriftrw](https://github.com/thriftrw/thriftrw-go) to generate Go structs and thrift (de)serializers;
- For http clients/endpoints, network communication is JSON over HTTP. Zanzibar uses [thriftrw](https://github.com/thriftrw/thriftrw-go) to generate Go structs and then uses [easyjson](https://github.com/mailru/easyjson) to generate JSON (de)serializers.

The [pre-steps.sh](https://github.com/uber/zanzibar/blob/master/codegen/runner/pre-steps.sh) script takes care of this part of the code generation, and places the generated code under `build/gen-code` directory.

#### Zanzibar-generated Code
Everything except `gen-code` under `build` directory is generated by Zanzibar. Zanzibar parses config files for each module to gathers meta information and then executing various [templates](https://github.com/uber/zanzibar/tree/master/codegen/templates) by applying them to the meta data. Here is what is generated for each builtin module class:

- client: dependency type, client interface and constructor if non-custom, mock client constructor
- middleware: dependency type, middleware type and constructor (unstable)
- endpoint: dependency type, endpoint type and constructor, workflow interface, workflow if non-custom, mock workflow constructor if custom
- service: dependency type and initializer, main.go, mock service constructor, servcie constructor

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
config_files=$(find "." -name "*-config.json" ! -path "*/build/*" ! -path "*/vendor/*" | sort)
for config_file in ${config_files}; do
	dir=$(dirname "$config_file")
	json_files=$(find "$dir" -name "*.json")
	for json_file in ${json_files}; do
		thrift_file=$(jq -r '.. | .thriftFile? | select(strings | endswith(".thrift"))' "$json_file")
		[[ -z ${thrift_file} ]] && continue
		[[ ${THRIFTRW_SRCS} == *${thrift_file}* ]] && continue
        THRIFTRW_SRCS+=" $CONFIG_DIR/idl/$thrift_file"
	done
done

bash ./vendor/github.com/uber/zanzibar/codegen/runner/pre-steps.sh "$BUILD_DIR" "$CONFIG_DIR" "zanzibar" "$THRIFTRW_SRCS"
go run ./vendor/github.com/uber/zanzibar/codegen/runner/runner.go --config="$CONFIG_DIR/build.json"
```
Note the above script will be abstracted for easier usage in the future.
### Testing
Zanzibar comes with builtin integration testing frameworks to help test business logic with ease. Setting [genMock](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/build.json#L16) to true will trigger Zanzibar to generate mock client, workflow and servcie constructors. The mock clients, being the leaf nodes in the dependency graph, are wired with the rest modules to create a testing application, which you can test against by setting expectations of the mock clients. The generated test helpers make writing tests straightforward and concise.

#### Entry Points
Currently Zanziabr provides two entry points to write integration tests: service and endpoint.
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
Zanzibar uses [gomock](https://github.com/golang/mock) to generate client mocks. To avoid manually setting the same fixture expectations again and again, Zanzibar augments gomock-generated mocks with fixture support. For example, the client-config.json file of `contacts` client has a `fixture` [field](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/clients/contacts/client-config.json#L11-L18):
```json
"fixture": {
    "importPath": "github.com/uber/zanzibar/examples/example-gateway/clients/contacts/fixture",
    "scenarios": {
        "SaveContacts": [
            "success"
        ]
    }
}
```
This basically says the `saveContacts` method has a `success` scenario which is defined in the [fixture package](https://github.com/uber/zanzibar/blob/master/examples/example-gateway/clients/contacts/fixture/fixture.go) indicated by the `importPath`. The fixture package is provided by users and here is what it looks like:
```go
package fixture

import (
	mc "github.com/uber/zanzibar/examples/example-gateway/build/clients/contacts/mock-client"
	gen "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/contacts/contacts"
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
Zanzibar provides post-generation [hooks](https://github.com/uber/zanzibar/blob/master/codegen/module.go#L68) which has access to the meta information of all modules. You can do whatever (mutating the input is probably not a good idea) suits your needs within a post-generation hook. Zanziabr invokes post-generation hooks as the very last step of code generation. In fact, mocks are all generated via post-generation [hooks](https://github.com/uber/zanzibar/blob/master/codegen/module_system.go#L226).

## Development
### Installation

```
mkdir -p $GOPATH/src/github.com/uber
git clone git@github.com:uber/zanzibar $GOPATH/src/github.com/uber/zanzibar
cd $GOPATH/src/github.com/uber/zanzibar
make install
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

### Update golden files

Run the test that compares golden files with `-update` flag, e.g.,
```
go test ./codegen -update
```

[doc-img]: https://godoc.org/github.com/uber/zanzibar?status.svg
[doc]: https://godoc.org/github.com/uber/zanzibar
[ci-img]: https://travis-ci.org/uber/zanzibar.svg?branch=master
[ci]: https://travis-ci.org/uber/zanzibar
[cov-img]: https://coveralls.io/repos/github/uber/zanzibar/badge.svg?branch=master
[cov]: https://coveralls.io/github/uber/zanzibar?branch=master
[go-report-img]: https://goreportcard.com/badge/github.com/uber/zanzibar
[go-report]: https://goreportcard.com/report/github.com/uber/zanzibar
