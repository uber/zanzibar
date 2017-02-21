#!/usr/bin/env bash
set -e

if [ -z $1 ]; then
	echo "prefix argument (\$1) is missing"
	exit 1
fi

PREFIX="$1"

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
    ./scripts/easy_json/easy_json "$file"
done