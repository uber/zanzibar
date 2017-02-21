#!/usr/bin/env bash
set -e

if [ -z $1 ]; then
	echo "prefix argument (\$1) is missing"
	exit 1
fi

PREFIX="$1"

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