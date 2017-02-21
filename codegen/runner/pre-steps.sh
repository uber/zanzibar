#!/usr/bin/env bash
set -e

if [ -z $1 ]; then
	echo "prefix argument (\$1) is missing"
	exit 1
fi

PREFIX="$1"
DIRNAME="$(dirname $0)"
EASY_JSON_RAW_DIR="$DIRNAME/../../scripts/easy_json"
EASY_JSON_DIR="`cd "${EASY_JSON_RAW_DIR}";pwd`"
EASY_JSON_FILE="$EASY_JSON_DIR/easy_json.go"

go get -u go.uber.org/thriftrw
echo "Generating Go code from Thrift files"
rm -rf "$PREFIX/gen-code"
mkdir "$PREFIX/gen-code"
for tfile in $(find "$PREFIX/idl" -name '*.thrift'); do
    thriftrw \
        --out="$PREFIX/gen-code" \
        --thrift-root="$PREFIX/idl" \
        --no-service-helpers "$tfile"
done
for file in $(find "$PREFIX/gen-code" -name 'versioncheck.go'); do
    rm "$file"
done

echo "Generating JSON Marshal/Unmarshal"
for file in $(find "$PREFIX/gen-code" -name "*.go" | grep -v "versioncheck.go");do
    go run "$EASY_JSON_FILE" -- "$file"
done