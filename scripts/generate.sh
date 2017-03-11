#!/usr/bin/env bash
PREFIX=examples/example-gateway

bash ./codegen/runner/pre-steps.sh "$PREFIX/build" "$PREFIX"
go run codegen/runner/runner.go -config="$PREFIX/gateway.json"
bash ./codegen/runner/post-steps.sh "$PREFIX/build"
