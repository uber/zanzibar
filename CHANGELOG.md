# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 1.0.0 - 2021-08-05
### Changed
- **BREAKING** `gateway.Channel` has been renamed to `gateway.ServerTChannel` to distinguish between client and server TChannels.


- **BREAKING** `gateway.TChannelRouter` has been renamed to `gateway.ServerTChannelRouter`

### Added
- A new boolean flag `dedicated.tchannel.client` is introduced. When set to `true`, each TChannel client creates a dedicated connection instead of registering on the default shared server connection.
Clients need not do anything as the default behaviour is to use a shared connection.([#778](https://github.com/uber/zanzibar/pull/778))

  
- A new field has been added to the gateway to propagate the `TChannelSubLoggerLevel` for clients with dedicated connections.


## 0.6.7 - 2021-02-05
### Changed
- **BREAKING** runtime/client_http_request.go, runtime/grpc_client.go, runtime/http_client.go, runtime/router.go, 
runtime/server_http_request.go, runtime/server_http_response.go, runtime/tchannel_client.go, 
runtime/tchannel_client_raw.go, runtime/tchannel_outbound_call.go, runtime_inbound_call.go, runtime/tchannel_server.go 
now uses contextLogger instead of nornal logger ([#748](https://github.com/uber/zanzibar/pull/748))

## 0.6.5 - 2020-08-10
### Added
- Added support for fetching multiple header values. https://github.com/uber/zanzibar/pull/733.
- Set explicit import alias for github.com/uber/zanzibar/runtime in templates. https://github.com/uber/zanzibar/pull/734.
- Added support for multiple header values in TestGateway utility. https://github.com/uber/zanzibar/pull/737.

## 0.6.2 - 2020-07-17
### Added
- Added support for variadic parameters in augmented mock clients. https://github.com/uber/zanzibar/pull/731.

## 0.6.1 - 2020-07-15
### Added
- Added support for grpc clients that have multiple services defined in proto. https://github.com/uber/zanzibar/pull/730.

## 0.6.0 - 2020-07-13
### Added
- Added support for circuit breaker, logging, and metrics similar to other protocol clients for gRPC clients ([#627](https://github.com/uber/zanzibar/pull/627))
### Fixed
- Fixed some bugs around legacy JSON config file support ([#626](https://github.com/uber/zanzibar/pull/626))
### Changed
- **BREAKING** `NewDefaultModuleSystemWithMockHook` API changed to add option for which hooks to execute. ([#638](https://github.com/uber/zanzibar/pull/638))
- `resolve_thrift` tool will now check if the given file has the `.thrift` extension. ([#634](https://github.com/uber/zanzibar/pull/634))
- **BREAKING** The `thriftRootDir` field type in application config file (build.yaml) is now `idlRootDir` since both thrift and protobuf are suppported (client idl only for now) different idl types. https://github.com/uber/zanzibar/pull/728.
- **BREAKING** The `genCodePackage` field type in application config file (build.yaml) is now `object` with properties `".thrift"` and `".proto"` to support separated gen code dirs for different idl types. https://github.com/uber/zanzibar/pull/728.


## 0.4.0 - 2019-08-21
### Fixed
- Fixed some nil pointer exceptions when dereferencing optional fields. ([#622](https://github.com/uber/zanzibar/pull/622))

## 0.3.1 - 2019-08-20
### Added
- Added support for `MaxTimes` and `MinTimes` in generated mocks. ([#620](https://github.com/uber/zanzibar/pull/620))
- Added support for query parameters for all HTTP methods rather than just `GET`. ([#625](https://github.com/uber/zanzibar/pull/625))

### Changed
- Use a single downstream TCP connection for outbound YARPC requests (when using gRPC). ([#624](https://github.com/uber/zanzibar/pull/624))

## 0.3.0 - 2019-08-09
### Added
- Added support for gRPC downstream clients by wrapping [YARPC](http://go.uber.org/yarpc). ([#592](https://github.com/uber/zanzibar/pull/592))
- Middlewares can be specified as default and therefore mandatory for endpoints. ([#558](https://github.com/uber/zanzibar/pull/558))
- Added getter for HTTP response [`Headers`](https://godoc.org/github.com/uber/zanzibar/runtime#ServerHTTPResponse.Headers). ([#566](https://github.com/uber/zanzibar/pull/566))
- Added support for multiple Thrift exceptions with the same status code. ([#565](https://github.com/uber/zanzibar/pull/565))
- Allow users to specify an incoming header to be designated as a request UUID, and logs using the context logger will include that header as a log field. ([#574](https://github.com/uber/zanzibar/pull/574))
- Endpoint and client request and response headers will be attached to log fields. ([#576](https://github.com/uber/zanzibar/pull/576))
- Added support for nested lists or maps containing structs, for example `list<map<string,FooStruct>>` in type converter. ([#586](https://github.com/uber/zanzibar/pull/586))
- Added support for HTTP query parameters to decoded into a Thrift set, in addition to the list already supported. ([#617](https://github.com/uber/zanzibar/pull/617))
- Alowed customizing the log level of the gateway logger ([#597](https://github.com/uber/zanzibar/pull/597))
- HTTP 404 log messages should include the URL that did not match any route for better debugging. ([#613](https://github.com/uber/zanzibar/pull/613))


### Fixed
- Endpoint mocks using non-standard import paths should generate with the correct import path in the mock. ([#555](https://github.com/uber/zanzibar/pull/555)).
- Fixed a bug with middlewares that have `/` in the name. ([#556](https://github.com/uber/zanzibar/pull/556)). 
- Mock server now uses `Shutdown` (graceful shutdown) rather than `Close`. 
- Test servers should listen on localhost rather than public IP address, preventing firewall warnings. ([#575](https://github.com/uber/zanzibar/pull/575))
- Fixed a nil pointer exception in assigning request UUID to outgoing request when request UUID is not present. ([#580](https://github.com/uber/zanzibar/pull/580/))
- Mock servers fixed to search for configuration under `TEST_SRCDIR` to support running under bazel. ([#587](https://github.com/uber/zanzibar/pull/587))
- When serving HTTP responses, only set `Content-Type` header to `application/json` if it is not already set, rather than unconditionally setting it, possibly overwriting it. ([#604](https://github.com/uber/zanzibar/pull/604))
- Fixed a bug with type converter for int16. ([#610](https://github.com/uber/zanzibar/pull/610))
- Post-gen build hooks should run once per build instead of once per module instance. ([#612](https://github.com/uber/zanzibar/pull/612))

### Changed
- **BREAKING** `NewTChannelClientContext` now requires a `ContextExtractor` ([#608](https://github.com/uber/zanzibar/pull/608))
- **BREAKING** `NewHTTPClientContext` signature changed to support multiple Thrift services. Now requires a `map[string]string` in place of a `[]string` for method names. ([#594](https://github.com/uber/zanzibar/pull/594))
- **BREAKING** `build.yaml` field `moduleSearchPaths` API changed, instead of a list of globbing patterns it is now a `map[string]string` map of class name to globbing pattern. ([#542](https://github.com/uber/zanzibar/pull/542))
- **BREAKING** It is now configurable whether to emit non-runtime metrics with `host` tag or not. Runtime config now expects the boolean field `metrics.m3.includeHost` to be set. Runtime metrics are always emitted with `host` tag. ([#570](https://github.com/uber/zanzibar/pull/570))
- **BREAKING** Circuit breaker metrics are now emitted with tags for the circuit breaker name rather than part of the metric name. ([#595](https://github.com/uber/zanzibar/pull/595))
- **BREAKING** `ClientThriftConfig` struct renamed to `ClientIDLConfig`. Client configuration files key `thriftFile` changed to `idlFile`. ([#618](https://github.com/uber/zanzibar/pull/618))
- HTTP router changed from [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter) to be built-in to avoid limitations of [httprouter#6](https://github.com/julienschmidt/httprouter/issues/6) and [httprouter#175](https://github.com/julienschmidt/httprouter/issues/175). ([#605](https://github.com/uber/zanzibar/pull/605))
- Default values for downstream client headers overwrite request values rather than appending. ([#551](https://github.com/uber/zanzibar/pull/551))
- Unpinned `tchannel-go` and `apache/thrift` in dependencies ([#554](https://github.com/uber/zanzibar/pull/554)). 
- Mandatory header checks will now only check that a header is present rather than checking that it is non-empty ([#588](https://github.com/uber/zanzibar/pull/588))
- When a HTTP client returns a 4xx or 5xx error, changed the log associated with that request to be a warning rather than info level. ([#596](https://github.com/uber/zanzibar/pull/596))
- Panics in type converter should have more information as to what type caused the panic. ([#611](https://github.com/uber/zanzibar/pull/611))


### Removed
- Removed logger metrics since it is barely useful.

## 0.2.0 - 2019-01-17
### Added
- Application configuration (e.g. `config/base.json`) can now be specified in YAML in addition to JSON (#504). Zanzibar will look for the `yaml` file first, and fall back to `json` file if it does not exist. JSON static configuration support may be removed in future major releases.
- Module configuration (`services/<name>/service-config.json`) can now be specified as YAML in addition to JSON (#468).
- Panics in endpoints are now caught (#458). HTTP endpoints return `502` status code with a body `"Unexpected workflow panic, recovered at endpoint."`. TChannel endpoints will return `ErrCodeUnexpected`.
- Transport specific client config structs added (`HTTPClientConfig`, `TChannelClientConfig`, `CustomClientConfig`) that match the JSON serialized objects in `client-config.json` for the supported client transports.
- Client calls are now protected with circuit breaker (https://github.com/uber/zanzibar/pull/539). Circuit breaker is enabled by default for each client, it can be disabled or fine tuned with proper configurations. It also emits appropriate metrics for monitoring/alerting.

### Changed
- **BREAKING** All [`metrics`](https://godoc.org/github.com/uber/zanzibar/runtime#call_metrics.go) counter and timer name has been changed and using RootScope instead of AllHostScope and PerHostScope since all parameter at name (e.g. host, env and etc) is already moved to tags.(e.g. fetch name:$service.$env.per-workers.inbound.calls.recvd is changed to fetch name:endpoint.request env:$env service:$service)
- **BREAKING** Application packages must now export a global variable named `AppOptions` of type [`*zanzibar.Options`](https://godoc.org/github.com/uber/zanzibar/runtime#Options) to be located at package root (the package defined in `build.json`/`build.yaml`).
- **BREAKING** `codegen.NewHTTPClientSpec`, `codegen.NewTChannelClientSpec`, `codegen.NewCustomClientSpec` and `codegen.ClientClassConfig` removed ([#515](https://github.com/uber/zanzibar/pull/515)).
- **BREAKING** HTTP router [`runtime.HTTPRouter`](https://godoc.org/github.com/uber/zanzibar/runtime#HTTPRouter) method `Register` renamed to `Handle` to better unify with the `net/http` standard library.
- **BREAKING** HTTP router type `runtime.HTTPRouter` switched from exposed concrete type to an interface type to allow changing the implementation.
- **BREAKING** `ServerHTTPRequest.Params` type changed from `julienschmidt/httprouter.Params` to `url.Values` from the standard library.
- Application logs should use the context logger in DefaultDeps.
- Added [`ContextExtractor`](https://godoc.org/github.com/uber/zanzibar/runtime#ContextExtractor) interface. It is part of the API for defining "extractors" or functions to pull out dynamic fields like trace ID, request headers, etc. out of the context to be used in log fields and metric tags. These can be used to pull out fields that are application-specific without adding code to zanzibar.
- Zanzibar now requires `yq` to be installed, whereas it previously required `jq` to be installed. `yq` is available over [PyPI](https://pypi.org/project/yq/) (`pip install yq`) and homebrew (`brew install python-yq`).
- Integrated with [Fx](http://go.uber.org/fx) in the main loop of the generated application.

### Deprecated
- JSON static configuration support is now deprecated.
- `JSONFileRaw` and `JSONFileName` fields of [`ModuleInstance`](https://godoc.org/github.com/uber/zanzibar/codegen#ModuleInstance) are now deprecated. When YAML configuration is used, `JSONFileRaw` and `JSONFileName` will be zero-valued.
- Exported types like [`ClientClassConfig`](https://godoc.org/github.com/uber/zanzibar/codegen#ModuleInstance) will have their JSON tags removed in the future.
- `gateway.Logger` is deprecated. Applications should get `ContextLogger` from `DefaultDeps` instead. Internal libraries can use the unexported `gateway.logger`.

### Fixed
- HTTP `DELETE` methods on clients can now send a JSON payload. Previously it was silently discarded.
- Fixed typo in metrics scope tag for `protocol` (was `protocal`).

## 0.1.2 - 2018-08-28
### Added
- **Context logger**: Added [`ContextLogger`](https://godoc.org/github.com/uber/zanzibar/runtime#ContextLogger) interface. This new logger interface automatically has log fields for Zanzibar-defined per-request fields like `requestUUID` to allow correlating all log messages from a given request. [`WithLogFields`](https://godoc.org/github.com/uber/zanzibar/runtime#WithLogFields) method added to `runtime` package to allow users to add their own log fields to be used by subsequent logs that use the context. The context logger is added as a new field to [`DefaultDependencies`](https://godoc.org/github.com/uber/zanzibar/runtime#DefaultDependencies).

### Changed
- Removed support for Go 1.9 and added support for Go 1.11.
- Some request fields like `endpointID` will no longer be present on messages using the default logger.

### Deprecated
- The default logger (`DefaultDependencies.Logger`) is now deprecated along with multiple related functions and public variables. The preferred way to log is now using the context logger.

## 0.1.1 - 2018-08-21
### Added
- Upgraded thriftrw to v1.12.0 from v1.8.0. This adds, among other things, `Ptr()` helper methods and getters for some thriftrw-go defined structs.

## 0.1.0 - 2018-08-17
### Added
- Initial release.
