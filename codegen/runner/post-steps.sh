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
EASY_JSON_BINARY="$EASY_JSON_DIR/easy_json"

start=$(cat .TMP_ZANZIBAR_TIMESTAMP_FILE.txt)

echo "Generating JSON Marshal/Unmarshal for rest"

for file in $(find "$PREFIX" -name "*_structs.go" | grep -v "$PREFIX/endpoints"); do
    "$EASY_JSON_BINARY" -all -- "$file"
done

end=`date +%s`
runtime=$((end-start))
echo "Generated easy_json files for clients +$runtime"

for file in $(find "$PREFIX/endpoints" -name "*_structs.go"); do
    "$EASY_JSON_BINARY" -all -- "$file"
done

end=`date +%s`
runtime=$((end-start))
echo "Generated easy_json files for endpoints +$runtime"

