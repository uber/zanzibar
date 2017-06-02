#!/usr/bin/env bash
set -e

PREFIX=examples/example-gateway

bash ./codegen/runner/pre-steps.sh "$PREFIX/build" "$PREFIX"

start=$(cat .TMP_ZANZIBAR_TIMESTAMP_FILE.txt)
go run codegen/runner/runner.go -config="$PREFIX/gateway.json"
end=`date +%s`
runtime=$((end-start))
echo "Generated endpoints/clients : +$runtime"

bash ./codegen/runner/post-steps.sh "$PREFIX/build"

rm .TMP_ZANZIBAR_TIMESTAMP_FILE.txt
