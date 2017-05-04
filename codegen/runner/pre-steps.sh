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
EASY_JSON_BINARY="$EASY_JSON_DIR/easy_json"


if [ -d "$DIRNAME/../../vendor" ]; then
	THRIFTRW_RAW_DIR="$DIRNAME/../../vendor/go.uber.org/thriftrw"
	THRIFTRW_DIR="`cd "${THRIFTRW_RAW_DIR}";pwd`"
	THRIFTRW_MAIN_FILE="$THRIFTRW_DIR/main.go"
	THRIFTRW_BINARY="$THRIFTRW_DIR/thriftrw"
else
	THRIFTRW_RAW_DIR="$DIRNAME/../../../../../go.uber.org/thriftrw"
	THRIFTRW_DIR="`cd "${THRIFTRW_RAW_DIR}";pwd`"
	THRIFTRW_MAIN_FILE="$THRIFTRW_DIR/main.go"
	THRIFTRW_BINARY="$THRIFTRW_DIR/thriftrw"
fi

start=`date +%s`
echo $start > .TMP_ZANZIBAR_TIMESTAMP_FILE.txt

go build -o $THRIFTRW_BINARY $THRIFTRW_MAIN_FILE
end=`date +%s`
runtime=$((end-start))
echo "Compiled thriftrw : +$runtime"

echo "Generating Go code from Thrift files"
mkdir -p "$BUILD_DIR/gen-code"
for tfile in $(find "$CONFIG_DIR/idl" -name '*.thrift'); do
    "$THRIFTRW_BINARY" --out="$BUILD_DIR/gen-code" \
		--no-embed-idl \
        --thrift-root="$CONFIG_DIR/idl" "$tfile"
done

end=`date +%s`
runtime=$((end-start))
echo "Generated structs : +$runtime"

go build -o $EASY_JSON_BINARY $EASY_JSON_FILE
end=`date +%s`
runtime=$((end-start))
echo "Compiled easyjson : +$runtime"

echo "Generating JSON Marshal/Unmarshal"
for file in $(find "$BUILD_DIR/gen-code" -name "types.go" | grep -v "versioncheck.go");do
    "$EASY_JSON_BINARY" -all -- "$file"
done

end=`date +%s`
runtime=$((end-start))
echo "Generated structs : +$runtime"
