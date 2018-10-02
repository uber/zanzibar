#!/usr/bin/env bash

set -e

function cleanup {
	rm -f $PREFIX/build/zanzibar.tree
}
trap cleanup EXIT

PREFIX=examples/example-gateway
ANNOPREFIX=${1:-zanzibar}

bash ./codegen/runner/pre-steps.sh "$PREFIX/build" "$PREFIX" "$ANNOPREFIX"

start=$(cat .TMP_ZANZIBAR_TIMESTAMP_FILE.txt)
if [ -f "$PREFIX/build.yaml" ]; then
    go run codegen/runner/runner.go -config="$PREFIX/build.yaml"
else
    go run codegen/runner/runner.go -config="$PREFIX/build.json"
fi

end=`date +%s`
runtime=$((end - start))
echo "Generated build : +$runtime"

rm .TMP_ZANZIBAR_TIMESTAMP_FILE.txt
