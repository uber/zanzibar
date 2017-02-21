#!/usr/bin/env bash
PREFIX=examples/example-gateway

bash ./codegen/runner/pre-steps.sh "$PREFIX"

go run codegen/runner/runner.go \
    -thrift_root_dir=examples/example-gateway/idl \
    -gateway_thrift_root_dir=examples/example-gateway/gen-code \
    -type_file_root_dir=examples/example-gateway \
    -target_gen_dir=examples/example-gateway/idl/github.com/uber/zanzibar \
    -client_thrift_dir=examples/example-gateway/idl/github.com/uber/zanzibar/clients \
    -endpoint_thrift_dir=examples/example-gateway/idl/github.com/uber/zanzibar/endpoints

bash ./codegen/runner/post-steps.sh "$PREFIX"
