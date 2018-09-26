# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased
### Changed
- Most built-in logs like `Finished an outgoing client HTTP request` now use the context logger. 

### Fixed
- HTTP `DELETE` methods on clients can now send a JSON payload. Previously it was silently discarded. 

## 1.1.0 - 2018-09-27
### Added
- **Context Extractor**: Added [`ContextExtractor`](https://godoc.org/github.com/uber/zanzibar/runtime#ContextLogger) interface. This new logger interface automatically

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
