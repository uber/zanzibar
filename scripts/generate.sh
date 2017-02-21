#!/usr/bin/env bash
PREFIX=examples/example-gateway

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

go run codegen/runner/runner.go \
    -thrift_root_dir=examples/example-gateway/idl \
    -gateway_thrift_root_dir=examples/example-gateway/gen-code \
    -type_file_root_dir=examples/example-gateway \
    -target_gen_dir=examples/example-gateway/idl/github.com/uber/zanzibar \
    -client_thrift_dir=examples/example-gateway/idl/github.com/uber/zanzibar/clients \
    -endpoint_thrift_dir=examples/example-gateway/idl/github.com/uber/zanzibar/endpoints

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