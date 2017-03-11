#!/usr/bin/env bash
set -e

if [ -z $1 ]; then
	echo "build dir argument (\$1) is missing"
	exit 1
fi

if [ -z $2 ]; then
	echo "config dir argument (\$2) is missing"
	exit 1
fi

BUILD_DIR="$1"
CONFIG_DIR="$2"
DIRNAME="$(dirname $0)"
EASY_JSON_RAW_DIR="$DIRNAME/../../scripts/easy_json"
EASY_JSON_DIR="`cd "${EASY_JSON_RAW_DIR}";pwd`"
EASY_JSON_FILE="$EASY_JSON_DIR/easy_json.go"

go get -u go.uber.org/thriftrw
echo "Generating Go code from Thrift files"
rm -rf "$BUILD_DIR/gen-code"
mkdir -p "$BUILD_DIR/gen-code"
for tfile in $(find "$CONFIG_DIR/idl" -name '*.thrift'); do
    thriftrw \
        --out="$BUILD_DIR/gen-code" \
        --thrift-root="$CONFIG_DIR/idl" \
        --no-service-helpers "$tfile"
done
for file in $(find "$BUILD_DIR/gen-code" -name 'versioncheck.go'); do
    rm "$file"
done

echo "Generating JSON Marshal/Unmarshal"
for file in $(find "$BUILD_DIR/gen-code" -name "*.go" | grep -v "versioncheck.go");do
    go run "$EASY_JSON_FILE" -- "$file"
done
