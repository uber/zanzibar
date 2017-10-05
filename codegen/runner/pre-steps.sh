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
RESOLVE_THRIFT_FILE="$DIRNAME/../../scripts/resolve_thrift/main.go"
RESOLVE_THRIFT_BINARY="$DIRNAME/../../scripts/resolve_thrift/resolve_thrift"

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
gofmt -w "$BUILD_DIR/gen-code/"

end=`date +%s`
runtime=$((end-start))
echo "Generated structs : +$runtime"

go build -o $EASY_JSON_BINARY $EASY_JSON_FILE
end=`date +%s`
runtime=$((end-start))
echo "Compiled easyjson : +$runtime"

go build -o $RESOLVE_THRIFT_BINARY $RESOLVE_THRIFT_FILE

# find the modules that actually need JSON (un)marshallers
target_dirs=""
found_thrifts=""
config_files=$(find ${CONFIG_DIR} -name "*-config.json" | sort)
for config_file in ${config_files}; do
    module_type=$(jq -r .type ${config_file})
    [[ ${module_type} != "http" ]] && continue
    dir=$(dirname ${config_file})
    json_files=$(find ${dir} -name "*.json")
    for json_file in ${json_files}; do
        thrift_file=$(jq -r '.. | .thriftFile? | select(type != "null")' ${json_file})
        [[ -z ${thrift_file} ]] && continue
        [[ ${found_thrifts} == *${thrift_file}* ]] && continue
        found_thrifts+=" $thrift_file"

        thrift_file="$CONFIG_DIR/idl/$thrift_file"
        gen_code_dir=$(
            "$RESOLVE_THRIFT_BINARY" $thrift_file | \
            sed "s|.*idl\/\(.*\)\/.*.thrift|$BUILD_DIR/gen-code/\1|" | \
            sort | uniq | xargs
        )
        target_dirs+=" $gen_code_dir"
    done
done
target_dirs=$(echo ${target_dirs} | tr ' ' '\n' | sort | uniq)

echo "Generating JSON Marshal/Unmarshal"
thriftrw_gofiles=(
	$(find ${target_dirs} -name "*.go" | \
		grep -v "versioncheck.go" | \
		grep -v "easyjson.go" | sort)
)
"$EASY_JSON_BINARY" -all -- "${thriftrw_gofiles[@]}"

end=`date +%s`
runtime=$((end-start))
echo "Generated structs : +$runtime"
