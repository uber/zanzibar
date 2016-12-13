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

go get github.com/mailru/easyjson/...


echo "Generating JSON Marshal/Unmarshal"
for file in $(find "$PREFIX/gen-code" -name "*.go" | grep -v "versioncheck.go"); do
    easyjson -all "$file"
done
for file in $(find "$PREFIX/gen-code" -name "*_easyjson.go"); do
    sed -i.bak -E "/^package (\w)+$/a\
\\\n// @generated" "$file"
    git clean -qf '*.bak'
done

echo "Generating JSON Marshal/Unmarshal for rest"
go generate $(glide novendor --dir $PREFIX | head -n-1 )
for file in $(find "$PREFIX" -name "*_easyjson.go"); do
    sed -i.bak -E "/^package (\w)+$/a\
\\\n// @generated" "$file"
    git clean -qf '*.bak'
done
