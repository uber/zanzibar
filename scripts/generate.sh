#!/usr/bin/env bash
PREFIX=examples/example-gateway

go get go.uber.org/thriftrw
echo "Generating Go code from Thrift files"
rm -rf "$PREFIX/gen-code"
mkdir "$PREFIX/gen-code"
for tfile in $(find "$PREFIX/idl/github.com" -name '*.thrift'); do
    thriftrw \
        --out="$PREFIX/gen-code" \
        --thrift-root="$PREFIX/idl/github.com" \
        --no-service-helpers "$tfile"
done

for file in $(find "$PREFIX/gen-code" -name 'versioncheck.go'); do
    rm "$file"
done

echo "Generating JSON Marshal/Unmarshal"
for file in $(find "$PREFIX/gen-code" -name "*.go" | grep -v "versioncheck.go"); do
    ./scripts/easy_json/easy_json "$file"
done


#for file in $(find "$PREFIX/gen-code" -name "*_easyjson.go"); do
    #sed -E "/^package (\w)+$/a\ // @generated/" "$file" > "$file"
    #git clean -qf '*.bak'
#done

echo "Generating JSON Marshal/Unmarshal for rest"
for file in $(find "$PREFIX" -name "*_structs.go"); do
    ./scripts/easy_json/easy_json "$file"
done

#for file in $(find "$PREFIX" -name "*_easyjson.go"); do
    #sed -E "/^package (\w)+$/a\ // @generated/" "$file" > "$file"
    #git clean -qf '*.bak'
#done

echo "Generating Endpoint Handlers"
mkdir "$PREFIX/gen-code/handlers"
for file in $(find "$PREFIX/gen-code/uber/zanzibar/endpoints" -name "*.go" | grep -v "versioncheck.go"); do
    go run lib/gencode/runner/main.go $PREFIX/gen-code/handlers $file
done
