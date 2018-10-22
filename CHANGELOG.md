# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased
### Added
- Application configuration (e.g. `config/base.json`) can now be specified in YAML in addition to JSON (). Zanzibar will look for the `yaml` file first, and fall back to `json` file if it does not exist. JSON static configuration support may be removed in future major releases. 
- Module configuration (`services/<name>/service-config.json`) can now be specified as YAML ina ddition to JSON (#468).
- Panics in endpoints are now caught (#458). HTTP endpoints return `502` status code with a body `"Unexpected workflow panic, recovered at endpoint."`. TChannel endpoints will return ErrCodeUnexpected. 

### Changed
- **BREAKING** Use [`ContextMerics`](https://godoc.org/github.com/uber/zanzibar/runtime#ContextMerics) instead of tally.Scope at [`NewRouterEndpoint`](https://godoc.org/github.com/uber/zanzibar/runtime#NewRouterEndpointContext), [`NewHTTPClient`](https://godoc.org/github.com/uber/zanzibar/runtime#NewHTTPClient), [`NewTChannelClient`](https://godoc.org/github.com/uber/zanzibar/runtime#NewTChannelClient)
- **BREAKING** Use [`ContextMerics`](https://godoc.org/github.com/uber/zanzibar/runtime#ContextMerics) to emit metrics instead of metrics from call_metrics.go.
- **BREAKING** All [`metrics`](https://godoc.org/github.com/uber/zanzibar/runtime#call_metrics.go) counter and timer name has been changed and is using RootScope instead of AllHostScope and PerHostScope
- **BREAKING** Application packages must now export a global variable named `AppOptions` of type [`*zanzibar.Options`](https://godoc.org/github.com/uber/zanzibar/runtime#Options) to be located at package root (the package defined in `build.json`/`build.yaml`). 
- Most built-in logs like `Finished an outgoing client HTTP request` now use the context logger. 
- Added [`ContextExtractor`](https://godoc.org/github.com/uber/zanzibar/runtime#ContextExtractor) interface. It is part of the API for dfining "extractors" or functions to pull out dynamic fields like trace ID, request headers, etc. out of the context to be used in log fields and metric tags. These can be used to pull out fields that are application-specific without adding code to zanzibar. 
- Zanzibar now requires `yq` to be installed, whereas it previously required `jq` to be installed. `yq` is available over [PyPI](https://pypi.org/project/yq/) (`pip install yq`) and homebrew (`brew install python-yq`).  

### Deprecated
- JSON static configuration support is now deprecated. 
- `JSONFileRaw` and `JSONFileName` fields of [`ModuleInstance`](https://godoc.org/github.com/uber/zanzibar/codegen#ModuleInstance) are now deprecated. When YAML configuration is used, `JSONFileRaw` and `JSONFileName` will be zero-valued. 
- Exported types like [`ClientClassConfig`](https://godoc.org/github.com/uber/zanzibar/codegen#ModuleInstance) will have their JSON tags removed in the future. 

### Fixed
- HTTP `DELETE` methods on clients can now send a JSON payload. Previously it was silently discarded. 

## 1.0.2 - 2018-08-28 
### Added
- **Context logger**: Added [`ContextLogger`](https://godoc.org/github.com/uber/zanzibar/runtime#ContextLogger) interface. This new logger interface automatically has log fields for Zanzibar-defined per-request fields like `requestUUID` to allow correlating all log messages from a given request. [`WithLogFields`](https://godoc.org/github.com/uber/zanzibar/runtime#WithLogFields) method added to `runtime` package to allow users to add their own log fields to be used by subsequent logs that use the context. The context logger is added as a new field to [`DefaultDependencies`](https://godoc.org/github.com/uber/zanzibar/runtime#DefaultDependencies). 

### Changed
- Removed support for Go 1.9 and added support for Go 1.11. 
- Some request fields like `endpointID` will no longer be present on messages using the default logger. 

### Deprecated
- The default logger (`DefaultDependencies.Logger`) is now deprecated along with multiple related functions and public variables. The preferred way to log is now using the context logger. 

## 1.0.1 - 2018-08-21
### Added
- Upgraded thriftrw to v1.12.0 from v1.8.0. This adds, among other things, `Ptr()` helper methods and getters for some thriftrw-go defined structs. 

## 1.0.0 - 2018-08-17
### Added
- Initial release.
