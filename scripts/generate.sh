#!/usr/bin/env bash

set -e

PREFIX=examples/example-gateway
ANNOPREFIX=${1:-zanzibar}

bash ./codegen/runner/pre-steps.sh "$PREFIX/build" "$PREFIX" "$ANNOPREFIX"

start=$(cat .TMP_ZANZIBAR_TIMESTAMP_FILE.txt)
go run codegen/runner/runner.go -config="$PREFIX/build.json"
end=`date +%s`
runtime=$((end - start))
echo "Generated build : +$runtime"

rm .TMP_ZANZIBAR_TIMESTAMP_FILE.txt
