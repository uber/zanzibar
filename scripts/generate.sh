#!/usr/bin/env bash

set -e

PREFIX=examples/example-gateway
ANNOPREFIX=${1:-zanzibar}
echo
echo "Generating code for example-gateway"
bash ./codegen/runner/pre-steps.sh "$PREFIX/build" "$PREFIX" "$ANNOPREFIX"
start=$(cat .TMP_ZANZIBAR_TIMESTAMP_FILE.txt)
if [[ -f "$PREFIX/build.yaml" ]]; then
    go run codegen/runner/runner.go -config="$PREFIX/build.yaml"
else
    go run codegen/runner/runner.go -config="$PREFIX/build.json"
fi

PREFIX=examples/selective-gateway
echo
echo "Generating code for selective-gateway"
bash ./codegen/runner/pre-steps.sh "$PREFIX/build" "$PREFIX" "$ANNOPREFIX"
go run codegen/runner/runner.go -config="$PREFIX/build.yaml" -selective

end=`date +%s`
runtime=$((end - start))
echo "Generated build : +$runtime"

rm .TMP_ZANZIBAR_TIMESTAMP_FILE.txt
