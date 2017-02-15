#!/usr/bin/env bash
PREFIX=examples/example-gateway

go get go.uber.org/thriftrw
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

go run codegen/runner/main.go

echo "Generating JSON Marshal/Unmarshal for rest"
for file in $(find "$PREFIX/gen-code" -name "*_structs.go"); do
    ./scripts/easy_json/easy_json "$file"
done
for file in $(find "$PREFIX/clients" -name "*_structs.go"); do
    ./scripts/easy_json/easy_json "$file"
done
for file in $(find "$PREFIX/endpoints" -name "*_structs.go"); do
    ./scripts/easy_json/easy_json "$file"
done